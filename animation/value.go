package animation

type ValueTracker struct{ Value float64 }

func NewValueTracker(v float64) *ValueTracker { return &ValueTracker{Value: v} }
func AnimateValue(v *ValueTracker, to float64) Animation {
	var from float64
	return New(1, Linear, func(float64) { from = v.Value }, func(t float64) { v.Value = from + (to-from)*t }, nil, func(float64) { v.Value = from })
}
