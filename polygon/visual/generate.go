// Package visual exports the solver's exact trace into a JavaScript scene.
// All drawing, layout, motion and teaching stages live in template.js.
package visual

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/23dd/animation/polygon"
)

//go:embed template.js
var template string

// Script creates a self-contained, editable JS scene for the existing engine.
// Integer values remain JSON strings to avoid JavaScript's 53-bit limitation.
func Script(r *polygon.Result) (string, error) {
	if r == nil {
		return "", fmt.Errorf("动画需要有效的求解结果")
	}
	n := len(r.Problem.Values)
	if n < 3 || n > 8 {
		return "", fmt.Errorf("动画支持 3..8 个顶点；求解器支持 3..%d 个", polygon.MaxVertices)
	}
	if len(r.States) != n*n || len(r.Steps) != n-1 || len(r.Cuts) != n {
		return "", fmt.Errorf("动画结果缺少 DP 状态、删边分数或合并步骤")
	}
	if got, err := polygon.Replay(r.Problem, r.ChosenCut, r.Steps); err != nil || got != r.MaxScore {
		return "", fmt.Errorf("动画结果的合并回放校验失败")
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化动画数据: %w", err)
	}
	return strings.Replace(template, "/*__POLYGON_DATA__*/null", string(data), 1), nil
}
