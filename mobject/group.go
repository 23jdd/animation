package mobject

type Group struct {
	BaseMobject
	Children []Mobject
}

func NewGroup(children ...Mobject) *Group {
	return &Group{BaseMobject: NewBase(), Children: append([]Mobject(nil), children...)}
}
func (g *Group) Add(children ...Mobject) { g.Children = append(g.Children, children...) }
func (g *Group) Update(t, dt float64) error {
	if err := Update(g, t, dt); err != nil {
		return err
	}
	for _, child := range g.Children {
		if err := child.Update(t, dt); err != nil {
			return err
		}
	}
	return nil
}
