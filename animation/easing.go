package animation

import "math"

type Easing func(float64) float64

func Linear(t float64) float64      { return t }
func EaseInQuad(t float64) float64  { return t * t }
func EaseOutQuad(t float64) float64 { return 1 - (1-t)*(1-t) }
func EaseInOutQuad(t float64) float64 {
	if t < .5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}
func EaseInCubic(t float64) float64  { return t * t * t }
func EaseOutCubic(t float64) float64 { return 1 - math.Pow(1-t, 3) }
func EaseInOutCubic(t float64) float64 {
	if t < .5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}
func Smooth(t float64) float64        { return t * t * (3 - 2*t) }
func Smoother(t float64) float64      { return t * t * t * (t*(t*6-15) + 10) }
func EaseInSine(t float64) float64    { return 1 - math.Cos(t*math.Pi/2) }
func EaseOutSine(t float64) float64   { return math.Sin(t * math.Pi / 2) }
func EaseInOutSine(t float64) float64 { return -(math.Cos(math.Pi*t) - 1) / 2 }
func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*t-10)
}
func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}
func EaseInCirc(t float64) float64  { return 1 - math.Sqrt(1-t*t) }
func EaseOutCirc(t float64) float64 { return math.Sqrt(1 - math.Pow(t-1, 2)) }
func EaseOutBack(t float64) float64 {
	const c1 = 1.70158
	c3 := c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}
func EaseOutBounce(t float64) float64 {
	const n1, d1 = 7.5625, 2.75
	switch {
	case t < 1/d1:
		return n1 * t * t
	case t < 2/d1:
		t -= 1.5 / d1
		return n1*t*t + .75
	case t < 2.5/d1:
		t -= 2.25 / d1
		return n1*t*t + .9375
	default:
		t -= 2.625 / d1
		return n1*t*t + .984375
	}
}
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t*10-.75)*(2*math.Pi/3)) + 1
}

func EasingByName(name string) Easing {
	switch name {
	case "easeIn", "easeInQuad":
		return EaseInQuad
	case "easeOut", "easeOutQuad":
		return EaseOutQuad
	case "easeInOut", "easeInOutQuad":
		return EaseInOutQuad
	case "cubic", "easeInOutCubic":
		return EaseInOutCubic
	case "sine":
		return EaseInOutSine
	case "expo":
		return EaseOutExpo
	case "circ":
		return EaseOutCirc
	case "back":
		return EaseOutBack
	case "bounce":
		return EaseOutBounce
	case "elastic":
		return EaseOutElastic
	case "smooth":
		return Smooth
	case "smoother":
		return Smoother
	default:
		return Linear
	}
}
