package mobject

type Group struct {
	BaseMobject
	Children []Mobject
}

func NewGroup(children ...Mobject) *Group {
	return &Group{BaseMobject: NewBase(), Children: append([]Mobject(nil), children...)}
}
func (g *Group) Add(children ...Mobject) { g.Children = append(g.Children, children...) }
func (g *Group) Update(t, dt float64) {
	Update(g, t, dt)
	for _, child := range g.Children {
		child.Update(t, dt)
	}
}
