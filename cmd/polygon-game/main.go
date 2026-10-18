// polygon-game computes the optimum in Go, then optionally exports a JS scene
// and renders that scene with this repository's existing animation engine.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	animexport "github.com/23dd/animation/export"
	"github.com/23dd/animation/polygon"
	"github.com/23dd/animation/polygon/visual"
	ggrenderer "github.com/23dd/animation/renderer/gg"
	"github.com/23dd/animation/script"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("polygon-game", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "输入文件；- 表示标准输入；省略则使用六边形样例")
	jsPath := flags.String("js", "", "输出独立的 JavaScript 动画文件")
	gifPath := flags.String("gif", "", "导出 GIF；未指定 -js 时自动生成同名 .js")
	pngPath := flags.String("png", "", "导出动画截图；默认是最终结果")
	jsonPath := flags.String("json", "", "输出含 DP 状态与合并步骤的 JSON")
	reportPath := flags.String("report", "", "保存本次运行的文字结果")
	fps := flags.Float64("fps", 12, "GIF 帧率（1..100）")
	scale := flags.Float64("scale", 1, "导出像素倍率；逻辑画布始终为 1280×720")
	at := flags.Float64("time", -1, "PNG 截图的秒数；-1 表示动画末尾")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("不支持额外位置参数，请使用 -input 文件名")
	}
	if math.IsNaN(*fps) || math.IsInf(*fps, 0) || *fps < 1 || *fps > 100 {
		return fmt.Errorf("fps 必须在 1..100 之间")
	}
	if math.IsNaN(*scale) || math.IsInf(*scale, 0) || *scale <= 0 || *scale > 4 {
		return fmt.Errorf("scale 必须大于 0 且不超过 4")
	}
	if math.IsNaN(*at) || math.IsInf(*at, 0) || (*at < 0 && *at != -1) {
		return fmt.Errorf("time 必须为非负秒数或 -1")
	}
	if (*gifPath != "" || *pngPath != "") && *jsPath == "" {
		base := *gifPath
		if base == "" {
			base = *pngPath
		}
		*jsPath = strings.TrimSuffix(base, filepath.Ext(base)) + ".js"
	}
	outputs := []struct{ path, extension string }{{*jsPath, ".js"}, {*gifPath, ".gif"}, {*pngPath, ".png"}, {*jsonPath, ".json"}, {*reportPath, ""}}
	seen := make(map[string]bool)
	inputAbs := ""
	if *input != "" && *input != "-" {
		var err error
		inputAbs, err = filepath.Abs(*input)
		if err != nil {
			return err
		}
	}
	for _, output := range outputs {
		if output.path == "" {
			continue
		}
		if output.extension != "" && !strings.EqualFold(filepath.Ext(output.path), output.extension) {
			return fmt.Errorf("输出文件 %q 的扩展名应为 %s", output.path, output.extension)
		}
		abs, err := filepath.Abs(output.path)
		if err != nil {
			return err
		}
		key := strings.ToLower(filepath.Clean(abs))
		if seen[key] || inputAbs != "" && strings.EqualFold(abs, inputAbs) {
			return fmt.Errorf("输入与各输出文件必须使用不同路径: %s", output.path)
		}
		seen[key] = true
	}

	p := polygon.Demo()
	if *input != "" {
		reader := stdin
		if *input != "-" {
			f, err := os.Open(*input)
			if err != nil {
				return err
			}
			defer f.Close()
			reader = f
		}
		var err error
		p, err = polygon.Parse(reader)
		if err != nil {
			return err
		}
	}
	result, err := polygon.Solve(p)
	if err != nil {
		return err
	}
	var source string
	if *jsPath != "" {
		source, err = visual.Script(result)
		if err != nil {
			return err
		}
	}
	for _, output := range outputs {
		if output.path != "" {
			if err := os.MkdirAll(filepath.Dir(output.path), 0755); err != nil {
				return err
			}
		}
	}
	report := formatReport(result)
	if _, err := io.WriteString(stdout, report); err != nil {
		return err
	}
	if *reportPath != "" {
		if err := os.WriteFile(*reportPath, []byte(report), 0644); err != nil {
			return err
		}
	}
	if *jsonPath != "" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*jsonPath, append(data, '\n'), 0644); err != nil {
			return err
		}
	}
	if *jsPath != "" {
		if err := os.WriteFile(*jsPath, []byte(source), 0644); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "JavaScript 动画:", *jsPath)
	}
	if *gifPath == "" && *pngPath == "" {
		return nil
	}
	runtime, err := script.LoadFile(*jsPath, 1280, 720)
	if err != nil {
		return err
	}
	renderer := ggrenderer.New(1280, 720)
	if err := renderer.Err(); err != nil {
		return err
	}
	width, height := int(math.Round(1280**scale)), int(math.Round(720**scale))
	if width < 1 || height < 1 {
		return fmt.Errorf("scale 产生了空白输出尺寸")
	}
	renderer.SetOutputSize(width, height)
	if *pngPath != "" {
		seconds := *at
		if seconds < 0 {
			seconds = runtime.Timeline.Duration()
		}
		if seconds > runtime.Timeline.Duration() {
			return fmt.Errorf("截图时间 %.2f 超过动画时长 %.2f", seconds, runtime.Timeline.Duration())
		}
		if err := animexport.PNG(*pngPath, runtime.Scene, runtime.Timeline, renderer, seconds); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "PNG 截图:", *pngPath)
	}
	if *gifPath != "" {
		started, last := time.Now(), time.Time{}
		err := animexport.GIFWithProgress(*gifPath, runtime.Scene, runtime.Timeline, renderer, *fps, func(done, total int) {
			if done == total || time.Since(last) >= 2*time.Second {
				fmt.Fprintf(stderr, "\rGIF %d/%d (%d%%) · %s", done, total, done*100/total, time.Since(started).Round(time.Second))
				last = time.Now()
				if done == total {
					fmt.Fprintln(stderr)
				}
			}
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "GIF 动画: %s（%d×%d，%.0f fps，%.2f 秒）\n", *gifPath, width, height, *fps, runtime.Timeline.Duration())
	}
	return nil
}

func formatReport(r *polygon.Result) string {
	var out strings.Builder
	fmt.Fprintln(&out, "多边形游戏：区间动态规划")
	fmt.Fprintln(&out, "顶点值:", strings.Join(r.Problem.Values, " "))
	fmt.Fprintln(&out, "边运算:", strings.Join(r.Problem.Operators, " "))
	fmt.Fprintln(&out, "约定: e_i 连接 v_i → v_(i+1)，e_n 连接 v_n → v_1。")
	fmt.Fprintln(&out, "\n每条首删边对应的最优得分:")
	for _, cut := range r.Cuts {
		fmt.Fprintf(&out, "  删除 e%d，从 v%d 开始: %s\n", cut.Edge, cut.StartVertex, cut.Score)
	}
	fmt.Fprintln(&out, "\n最大得分:", r.MaxScore)
	fmt.Fprintln(&out, "全部最优首删边:", r.BestCuts)
	fmt.Fprintf(&out, "演示方案: 删除 e%d，从 v%d 开始\n", r.ChosenCut, r.StartVertex)
	fmt.Fprintln(&out, "最优表达式:", r.Expression)
	fmt.Fprintln(&out, "\n最优合并过程:")
	for i, step := range r.Steps {
		fmt.Fprintf(&out, "  %d. 沿 e%d: %s %s %s = %s\n", i+1, step.Edge, step.Left, step.Operator, step.Right, step.Value)
	}
	fmt.Fprintf(&out, "\n验证: %d 次合法相邻合并，最终结果 %s，与 DP 最优值一致。\n", len(r.Steps), r.MaxScore)
	fmt.Fprintln(&out, "复杂度: O(n³) 次整数运算，O(n²) 个 DP 状态；整数使用 math/big.Int 防溢出。")
	return out.String()
}
