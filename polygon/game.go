// Package polygon solves the classical polygon game using interval DP.
// It has no dependency on the animation engine or a graphics backend.
package polygon

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"strings"
)

const MaxVertices = 100

// Problem numbers vertices and edges from 1. Edge i joins vertex i to
// vertex i+1 (edge n joins n to 1); Operators[i-1] labels that outgoing edge.
type Problem struct {
	Values    []string `json:"values"`
	Operators []string `json:"operators"`
}

// State describes a circular interval. Start is 1-based and Length is its
// number of vertices. Split is the number of vertices in the left subinterval
// of an optimal maximum; Min and Max are decimal strings, not float64 values.
type State struct {
	Start     int    `json:"start"`
	Length    int    `json:"length"`
	Min       string `json:"min"`
	Max       string `json:"max"`
	Split     int    `json:"split,omitempty"`
	Operator  string `json:"operator,omitempty"`
	Left      string `json:"left,omitempty"`
	Right     string `json:"right,omitempty"`
	LeftMode  string `json:"left_mode,omitempty"`
	RightMode string `json:"right_mode,omitempty"`
}

type Cut struct {
	Edge        int    `json:"edge"`
	StartVertex int    `json:"start_vertex"`
	Score       string `json:"score"`
}

// Step is one legal contraction. Vertices record the original vertices in
// each operand, so the solution can be replayed without ambiguous labels.
type Step struct {
	Edge          int    `json:"edge"`
	LeftVertices  []int  `json:"left_vertices"`
	RightVertices []int  `json:"right_vertices"`
	Left          string `json:"left"`
	Operator      string `json:"operator"`
	Right         string `json:"right"`
	Value         string `json:"value"`
}

type Result struct {
	Problem     Problem `json:"problem"`
	MaxScore    string  `json:"max_score"`
	BestCuts    []int   `json:"best_cuts"`
	ChosenCut   int     `json:"chosen_cut"`
	StartVertex int     `json:"start_vertex"`
	Expression  string  `json:"expression"`
	Steps       []Step  `json:"steps"`
	Cuts        []Cut   `json:"cuts"`
	States      []State `json:"states"`
}

// Demo includes negative operands: keeping only interval maxima is incorrect.
func Demo() Problem {
	return Problem{
		Values:    []string{"-3", "2", "-4", "5", "-2", "3"},
		Operators: []string{"+", "*", "+", "*", "+", "*"},
	}
}

// Parse reads: n, n integer vertex values, then n outgoing edge operators.
// Tokens may be separated by any whitespace. x and × are aliases of *.
func Parse(r io.Reader) (Problem, error) {
	in := bufio.NewReader(r)
	// Accept UTF-8 BOMs from Windows editors without treating them as part of n.
	if prefix, _ := in.Peek(3); len(prefix) == 3 && prefix[0] == 0xef && prefix[1] == 0xbb && prefix[2] == 0xbf {
		_, _ = in.Discard(3)
	}
	var n int
	if _, err := fmt.Fscan(in, &n); err != nil {
		return Problem{}, fmt.Errorf("读取顶点个数: %w", err)
	}
	if n < 3 || n > MaxVertices {
		return Problem{}, fmt.Errorf("顶点个数必须在 3..%d 之间", MaxVertices)
	}
	p := Problem{Values: make([]string, n), Operators: make([]string, n)}
	for i := range p.Values {
		if _, err := fmt.Fscan(in, &p.Values[i]); err != nil {
			return Problem{}, fmt.Errorf("读取顶点 v%d: %w", i+1, err)
		}
	}
	for i := range p.Operators {
		if _, err := fmt.Fscan(in, &p.Operators[i]); err != nil {
			return Problem{}, fmt.Errorf("读取边 e%d: %w", i+1, err)
		}
	}
	var extra string
	if _, err := fmt.Fscan(in, &extra); err == nil {
		return Problem{}, fmt.Errorf("输入末尾有多余内容 %q", extra)
	} else if err != io.EOF {
		return Problem{}, fmt.Errorf("读取输入末尾: %w", err)
	}
	p, _, err := normalize(p)
	return p, err
}

func normalize(p Problem) (Problem, []*big.Int, error) {
	n := len(p.Values)
	if n < 3 || n > MaxVertices || len(p.Operators) != n {
		return Problem{}, nil, fmt.Errorf("需要 3..%d 个顶点，以及相同数量的边运算符", MaxVertices)
	}
	q := Problem{Values: make([]string, n), Operators: make([]string, n)}
	values := make([]*big.Int, n)
	for i, s := range p.Values {
		value, ok := new(big.Int).SetString(strings.TrimSpace(s), 10)
		if !ok {
			return Problem{}, nil, fmt.Errorf("v%d 的值 %q 不是十进制整数", i+1, s)
		}
		values[i], q.Values[i] = value, value.String()
		op := strings.TrimSpace(p.Operators[i])
		if op == "x" || op == "×" {
			op = "*"
		}
		if op != "+" && op != "*" {
			return Problem{}, nil, fmt.Errorf("e%d 的运算符 %q 无效，仅支持 + 和 *", i+1, op)
		}
		q.Operators[i] = op
	}
	return q, values, nil
}

type choice struct {
	split             int
	leftMin, rightMin bool
}

type cell struct {
	min, max         *big.Int
	minPick, maxPick choice
}

func (c *cell) value(minimum bool) *big.Int {
	if minimum {
		return c.min
	}
	return c.max
}

func (c *cell) pick(minimum bool) choice {
	if minimum {
		return c.minPick
	}
	return c.maxPick
}

func calculate(a, b *big.Int, operator string) *big.Int {
	if operator == "+" {
		return new(big.Int).Add(a, b)
	}
	return new(big.Int).Mul(a, b)
}

// Solve uses doubled circular input and O(n^3) arithmetic operations with
// O(n^2) DP cells. big.Int prevents overflow; integer bit costs are additional.
func Solve(problem Problem) (*Result, error) {
	p, numbers, err := normalize(problem)
	if err != nil {
		return nil, err
	}
	n := len(numbers)
	dp := make([][]cell, 2*n)
	for i := range dp {
		dp[i] = make([]cell, n+1)
		dp[i][1] = cell{min: new(big.Int).Set(numbers[i%n]), max: new(big.Int).Set(numbers[i%n])}
	}
	for length := 2; length <= n; length++ {
		for start := 0; start+length <= 2*n; start++ {
			current := &dp[start][length]
			for leftLength := 1; leftLength < length; leftLength++ {
				left, right := &dp[start][leftLength], &dp[start+leftLength][length-leftLength]
				op := p.Operators[(start+leftLength-1)%n]
				// Negative multiplication needs all four endpoint combinations.
				for _, leftMin := range []bool{false, true} {
					for _, rightMin := range []bool{false, true} {
						candidate := calculate(left.value(leftMin), right.value(rightMin), op)
						pick := choice{leftLength, leftMin, rightMin}
						if current.min == nil || candidate.Cmp(current.min) < 0 {
							current.min, current.minPick = new(big.Int).Set(candidate), pick
						}
						if current.max == nil || candidate.Cmp(current.max) > 0 {
							current.max, current.maxPick = new(big.Int).Set(candidate), pick
						}
					}
				}
			}
		}
	}

	r := &Result{Problem: p}
	var best *big.Int
	// Iterate by edge number, not by start, so all tied optimal cuts are sorted.
	for edge := 1; edge <= n; edge++ {
		start := edge % n
		score := dp[start][n].max
		r.Cuts = append(r.Cuts, Cut{edge, start + 1, score.String()})
		if best == nil || score.Cmp(best) > 0 {
			best = new(big.Int).Set(score)
			r.BestCuts = []int{edge}
		} else if score.Cmp(best) == 0 {
			r.BestCuts = append(r.BestCuts, edge)
		}
	}
	r.MaxScore, r.ChosenCut = best.String(), r.BestCuts[0]
	start := r.ChosenCut % n
	r.StartVertex = start + 1
	for length := 1; length <= n; length++ {
		for begin := 0; begin < n; begin++ {
			c := &dp[begin][length]
			s := State{Start: begin + 1, Length: length, Min: c.min.String(), Max: c.max.String()}
			if length > 1 {
				pick := c.maxPick
				s.Split, s.Operator = pick.split, p.Operators[(begin+pick.split-1)%n]
				s.Left = dp[begin][pick.split].value(pick.leftMin).String()
				s.Right = dp[begin+pick.split][length-pick.split].value(pick.rightMin).String()
				s.LeftMode, s.RightMode = mode(pick.leftMin), mode(pick.rightMin)
			}
			r.States = append(r.States, s)
		}
	}

	var recover func(int, int, bool) (string, []int, *big.Int)
	recover = func(begin, length int, minimum bool) (string, []int, *big.Int) {
		if length == 1 {
			value := p.Values[begin%n]
			if numbers[begin%n].Sign() < 0 {
				value = "(" + value + ")"
			}
			return value, []int{begin%n + 1}, numbers[begin%n]
		}
		pick := dp[begin][length].pick(minimum)
		leftExpr, leftVertices, leftValue := recover(begin, pick.split, pick.leftMin)
		rightExpr, rightVertices, rightValue := recover(begin+pick.split, length-pick.split, pick.rightMin)
		edge := (begin+pick.split-1)%n + 1
		op := p.Operators[edge-1]
		value := calculate(leftValue, rightValue, op)
		r.Steps = append(r.Steps, Step{edge, leftVertices, rightVertices, leftValue.String(), op, rightValue.String(), value.String()})
		vertices := append(append([]int(nil), leftVertices...), rightVertices...)
		return "(" + leftExpr + " " + op + " " + rightExpr + ")", vertices, value
	}
	r.Expression, _, _ = recover(start, n, false)
	if score, err := Replay(p, r.ChosenCut, r.Steps); err != nil {
		return nil, fmt.Errorf("最优方案回放校验失败: %w", err)
	} else if score != r.MaxScore {
		return nil, fmt.Errorf("最优方案回放结果 %s 与 DP 结果 %s 不一致", score, r.MaxScore)
	}
	return r, nil
}

func mode(minimum bool) string {
	if minimum {
		return "min"
	}
	return "max"
}

// Replay checks every merge for adjacency, original edge, operand values and
// result. It returns the remaining vertex value, or an error for an illegal move.
func Replay(problem Problem, cut int, steps []Step) (string, error) {
	p, numbers, err := normalize(problem)
	if err != nil {
		return "", err
	}
	n := len(numbers)
	if cut < 1 || cut > n || len(steps) != n-1 {
		return "", fmt.Errorf("首删边应在 1..%d，合并步骤应为 %d 次", n, n-1)
	}
	type token struct {
		vertices []int
		value    *big.Int
	}
	tokens := make([]token, n)
	for i := range tokens {
		vertex := (cut + i) % n
		tokens[i] = token{[]int{vertex + 1}, numbers[vertex]}
	}
	for index, step := range steps {
		found := -1
		for i := 0; i+1 < len(tokens); i++ {
			if equalVertices(tokens[i].vertices, step.LeftVertices) && equalVertices(tokens[i+1].vertices, step.RightVertices) {
				found = i
				break
			}
		}
		if found < 0 {
			return "", fmt.Errorf("第 %d 步的两个操作数不是当前相邻顶点", index+1)
		}
		left, right := tokens[found], tokens[found+1]
		edge := left.vertices[len(left.vertices)-1]
		op := p.Operators[edge-1]
		value := calculate(left.value, right.value, op)
		if edge != step.Edge || op != step.Operator || left.value.String() != step.Left || right.value.String() != step.Right || value.String() != step.Value {
			return "", fmt.Errorf("第 %d 步的边、运算符、操作数或结果不正确", index+1)
		}
		merged := token{append(append([]int(nil), left.vertices...), right.vertices...), value}
		tokens[found] = merged
		tokens = append(tokens[:found+1], tokens[found+2:]...)
	}
	return tokens[0].value.String(), nil
}

func equalVertices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
