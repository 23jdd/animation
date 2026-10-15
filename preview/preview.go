package preview

import (
	"github.com/23dd/animation/engine"
	"github.com/hajimehoshi/ebiten/v2"
)

func Run(e *engine.Engine, title string) error {
	w, h := e.Renderer.Size()
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowSize(w, h+playerHeight)
	ebiten.SetWindowSizeLimits(720, playerHeight+240, -1, -1)
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(New(e))
}
