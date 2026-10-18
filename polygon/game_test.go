package polygon

import (
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestKnownExamples(t *testing.T) {
	tests := []struct {
		name string
		p    Problem
		want string
		cuts []int
	}{
		{"classic", Problem{[]string{"-7", "4", "2", "5"}, []string{"+", "*", "*", "+"}}, "33", []int{1, 4}},
		{"addition ties", Problem{[]string{"1", "2", "3"}, []string{"+", "+", "+"}}, "6", []int{1, 2, 3}},
		{"negative result", Problem{[]string{"-2", "-3", "-4"}, []string{"*", "*", "*"}}, "-24", []int{1, 2, 3}},
		{"zero", Problem{[]string{"0", "0", "0"}, []string{"+", "*", "+"}}, "0", []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := Solve(tt.p)
			if err != nil {
				t.Fatal(err)
			}
			if r.MaxScore != tt.want || !reflect.DeepEqual(r.BestCuts, tt.cuts) {
				t.Fatalf("score=%s cuts=%v, want %s %v", r.MaxScore, r.BestCuts, tt.want, tt.cuts)
			}
			got, err := Replay(tt.p, r.ChosenCut, r.Steps)
			if err != nil || got != tt.want {
				t.Fatalf("replay=%q err=%v", got, err)
			}
		})
	}
}

// Exhaustive parenthesization is independent of the min/max recurrence.
func brute(p Problem, start, length int) []int64 {
	if length == 1 {
		v, _ := strconv.ParseInt(p.Values[start%len(p.Values)], 10, 64)
		return []int64{v}
	}
	var results []int64
	for split := 1; split < length; split++ {
		for _, left := range brute(p, start, split) {
			for _, right := range brute(p, start+split, length-split) {
				if p.Operators[(start+split-1)%len(p.Values)] == "+" {
					results = append(results, left+right)
				} else {
					results = append(results, left*right)
				}
			}
		}
	}
	return results
}

func TestAllStatesAgainstExhaustiveSearch(t *testing.T) {
	rng := rand.New(rand.NewSource(20261017))
	for n := 3; n <= 7; n++ {
		for trial := 0; trial < 35; trial++ {
			p := Problem{Values: make([]string, n), Operators: make([]string, n)}
			for i := 0; i < n; i++ {
				p.Values[i] = strconv.Itoa(rng.Intn(9) - 4)
				p.Operators[i] = "+"
				if rng.Intn(2) == 0 {
					p.Operators[i] = "*"
				}
			}
			r, err := Solve(p)
			if err != nil {
				t.Fatal(err)
			}
			for _, state := range r.States {
				all := brute(p, state.Start-1, state.Length)
				lo, hi := all[0], all[0]
				for _, v := range all {
					lo = min(lo, v)
					hi = max(hi, v)
				}
				if state.Min != strconv.FormatInt(lo, 10) || state.Max != strconv.FormatInt(hi, 10) {
					t.Fatalf("p=%+v state=%+v exhaustive=[%d,%d]", p, state, lo, hi)
				}
			}
			if got, err := Replay(p, r.ChosenCut, r.Steps); err != nil || got != r.MaxScore {
				t.Fatalf("replay=%s err=%v", got, err)
			}
		}
	}
}

func TestArbitraryPrecisionAndNoInputMutation(t *testing.T) {
	large := "100000000000000000000"
	p := Problem{[]string{large, large, large, large}, []string{"x", "×", "*", "*"}}
	r, err := Solve(p)
	if err != nil {
		t.Fatal(err)
	}
	want := new(big.Int).Exp(big.NewInt(10), big.NewInt(80), nil).String()
	if r.MaxScore != want {
		t.Fatalf("got %s want %s", r.MaxScore, want)
	}
	if p.Operators[0] != "x" || p.Operators[1] != "×" {
		t.Fatal("Solve mutated caller input")
	}
}

func TestParseAndInvalidInputs(t *testing.T) {
	if _, err := Parse(strings.NewReader("\ufeff3\n1 2 3\n+ * +")); err != nil {
		t.Fatalf("UTF-8 BOM input: %v", err)
	}
	p, err := Parse(strings.NewReader("3\n-2 0 +004\nx × +\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.Values, []string{"-2", "0", "4"}) || !reflect.DeepEqual(p.Operators, []string{"*", "*", "+"}) {
		t.Fatalf("%+v", p)
	}
	for _, input := range []string{"", "two", "2 1 2 + +", "101", "3 1 2 3 + +", "3 1 2 3 + / *", "3 1.2 2 3 + + +", "3 1 2 3 + + + extra"} {
		if _, err := Parse(strings.NewReader(input)); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	if _, err := Solve(Problem{}); err == nil {
		t.Fatal("accepted empty polygon")
	}
	if _, err := Solve(Problem{[]string{"1", "2", "3"}, []string{"+"}}); err == nil {
		t.Fatal("accepted unequal lengths")
	}
}

func TestReplayRejectsIllegalMoves(t *testing.T) {
	r, err := Solve(Demo())
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]Step){
		func(s []Step) { s[0].Value = "999999" },
		func(s []Step) { s[0].Edge = r.ChosenCut },
		func(s []Step) { s[0].LeftVertices = []int{99} },
		func(s []Step) { s[0].Operator = "/" },
	} {
		steps := append([]Step(nil), r.Steps...)
		mutate(steps)
		if _, err := Replay(r.Problem, r.ChosenCut, steps); err == nil {
			t.Fatal("accepted invalid contraction")
		}
	}
	if _, err := Replay(r.Problem, 0, r.Steps); err == nil {
		t.Fatal("accepted invalid cut")
	}
	if _, err := Replay(r.Problem, r.ChosenCut, r.Steps[:1]); err == nil {
		t.Fatal("accepted incomplete plan")
	}
}
