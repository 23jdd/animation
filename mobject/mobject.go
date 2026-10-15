package mobject

import (
	"image/color"
	"sync/atomic"

	"github.com/23dd/animation/geometry"
)

type Updater func(Mobject, float64, float64)

type Mobject interface {
	ID() uint64
	Base() *BaseMobject
	Update(time, dt float64)
}

var nextID atomic.Uint64

type BaseMobject struct {
	ObjectID     uint64
	Transform    geometry.Transform
	Opacity      float64
	IsVisible    bool
	DrawProgress float64
	Fill         color.RGBA
	Stroke       color.RGBA
	StrokeWidth  float64
	updaters     []Updater
}

func NewBase() BaseMobject {
	return BaseMobject{ObjectID: nextID.Add(1), Transform: geometry.NewTransform(), Opacity: 1, IsVisible: true, DrawProgress: 1, Fill: color.RGBA{255, 255, 255, 255}, Stroke: color.RGBA{255, 255, 255, 255}, StrokeWidth: 2}
}
func (b *BaseMobject) ID() uint64                  { return b.ObjectID }
func (b *BaseMobject) Base() *BaseMobject          { return b }
func (b *BaseMobject) Position() geometry.Vec2     { return b.Transform.Position }
func (b *BaseMobject) SetPosition(v geometry.Vec2) { b.Transform.Position = v }
func (b *BaseMobject) AddUpdater(u Updater)        { b.updaters = append(b.updaters, u) }
func (b *BaseMobject) ClearUpdaters()              { b.updaters = nil }
func (b *BaseMobject) RunUpdaters(self Mobject, time, dt float64) {
	for _, u := range b.updaters {
		u(self, time, dt)
	}
}

func Update(m Mobject, time, dt float64) { m.Base().RunUpdaters(m, time, dt) }
