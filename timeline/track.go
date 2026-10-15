package timeline

import "github.com/23dd/animation/animation"

type Track struct {
	Start     float64
	Animation animation.Animation
}

func (t Track) End() float64 { return t.Start + t.Animation.Duration() }
