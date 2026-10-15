package preview

import (
	"image"
	"testing"

	"github.com/23dd/animation/engine"
	"github.com/23dd/animation/scene"
	"github.com/23dd/animation/timeline"
	"github.com/ebitenui/ebitenui/widget"
)

type testRenderer struct{ width, height int }

func (r testRenderer) Render(*scene.Scene) image.Image {
	return image.NewRGBA(image.Rect(0, 0, r.width, r.height))
}
func (r testRenderer) Size() (int, int) { return r.width, r.height }

func TestPlayButtonPauseResumeAndReplay(t *testing.T) {
	tl := timeline.New()
	tl.Wait(2)
	e := engine.New(scene.New(320, 180), tl, testRenderer{320, 180})
	g := New(e)
	g.playButton.Validate()

	g.playButton.ClickedEvent.Fire(&widget.ButtonClickedEventArgs{Button: g.playButton})
	g.ui.Update()
	if e.State().Playing {
		t.Fatal("play button did not pause")
	}
	g.playButton.ClickedEvent.Fire(&widget.ButtonClickedEventArgs{Button: g.playButton})
	g.ui.Update()
	if !e.State().Playing {
		t.Fatal("play button did not resume")
	}

	e.Pause()
	e.Seek(2)
	g.playButton.ClickedEvent.Fire(&widget.ButtonClickedEventArgs{Button: g.playButton})
	g.ui.Update()
	state := e.State()
	if !state.Playing || state.Time != 0 {
		t.Fatalf("replay state = %+v", state)
	}
}
