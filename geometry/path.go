package geometry

type PathCommandType uint8

const (
	MoveCommand PathCommandType = iota
	LineCommand
	QuadraticCommand
	CubicCommand
	CloseCommand
)

type PathCommand struct {
	Type   PathCommandType
	Points []Vec2
}

type Path struct{ Commands []PathCommand }

func (p *Path) MoveTo(v Vec2) *Path {
	p.Commands = append(p.Commands, PathCommand{MoveCommand, []Vec2{v}})
	return p
}
func (p *Path) LineTo(v Vec2) *Path {
	p.Commands = append(p.Commands, PathCommand{LineCommand, []Vec2{v}})
	return p
}
func (p *Path) QuadraticTo(c, end Vec2) *Path {
	p.Commands = append(p.Commands, PathCommand{QuadraticCommand, []Vec2{c, end}})
	return p
}
func (p *Path) CubicTo(c1, c2, end Vec2) *Path {
	p.Commands = append(p.Commands, PathCommand{CubicCommand, []Vec2{c1, c2, end}})
	return p
}
func (p *Path) Close() *Path {
	p.Commands = append(p.Commands, PathCommand{Type: CloseCommand})
	return p
}
