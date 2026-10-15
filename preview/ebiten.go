package preview

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/ebitenui/ebitenui"
	uiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/23dd/animation/engine"
	"github.com/23dd/animation/fontasset"
	"github.com/23dd/animation/renderer"
)

const (
	playerHeight  = 126
	timelineSteps = 10_000
)

type Game struct {
	Engine *engine.Engine
	frame  *ebiten.Image
	width  int
	height int

	ui             *ebitenui.UI
	playButton     *widget.Button
	timelineSlider *widget.Slider
	timeLabel      *widget.Label
	statusLabel    *widget.Label
	expectedSlider int
}

func New(e *engine.Engine) *Game {
	w, h := e.Renderer.Size()
	g := &Game{Engine: e, frame: ebiten.NewImage(w, h), width: w, height: h}
	g.buildUI()
	return g
}

func (g *Game) Update() error {
	g.ui.Update()
	g.handleKeyboard()
	g.Engine.Update()
	g.syncPlayerState()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 12, G: 14, B: 20, A: 255})
	windowWidth, windowHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	x, y, renderWidth, renderHeight := fitCanvasPixels(g.width, g.height, windowWidth, windowHeight-playerHeight)
	directOutput := false
	if sizer, ok := g.Engine.Renderer.(renderer.OutputSizer); ok {
		sizer.SetOutputSize(renderWidth, renderHeight)
		directOutput = true
	}
	src := g.Engine.Render()
	rgba, ok := src.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(src.Bounds())
		draw.Draw(rgba, rgba.Bounds(), src, src.Bounds().Min, draw.Src)
	}
	frameWidth, frameHeight := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	if bounds := g.frame.Bounds(); bounds.Dx() != frameWidth || bounds.Dy() != frameHeight {
		g.frame = ebiten.NewImage(frameWidth, frameHeight)
	}
	g.frame.WritePixels(rgba.Pix)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest}
	if directOutput {
		op.GeoM.Translate(float64(x), float64(y))
	} else {
		x, y, scale := fitCanvas(frameWidth, frameHeight, windowWidth, windowHeight-playerHeight)
		op.Filter = ebiten.FilterLinear
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
	}
	screen.DrawImage(g.frame, op)
	g.ui.Draw(screen)
}

func fitCanvasPixels(sourceWidth, sourceHeight, availableWidth, availableHeight int) (x, y, width, height int) {
	if sourceWidth <= 0 || sourceHeight <= 0 || availableWidth <= 0 || availableHeight <= 0 {
		return 0, 0, max(1, sourceWidth), max(1, sourceHeight)
	}
	if float64(availableWidth)/float64(availableHeight) > float64(sourceWidth)/float64(sourceHeight) {
		height = availableHeight
		width = int(math.Round(float64(height) * float64(sourceWidth) / float64(sourceHeight)))
	} else {
		width = availableWidth
		height = int(math.Round(float64(width) * float64(sourceHeight) / float64(sourceWidth)))
	}
	x = (availableWidth - width) / 2
	y = (availableHeight - height) / 2
	return
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth < 1 {
		outsideWidth = g.width
	}
	if outsideHeight < 1 {
		outsideHeight = g.height + playerHeight
	}
	return outsideWidth, outsideHeight
}

func fitCanvas(sourceWidth, sourceHeight, availableWidth, availableHeight int) (x, y, scale float64) {
	if sourceWidth <= 0 || sourceHeight <= 0 || availableWidth <= 0 || availableHeight <= 0 {
		return 0, 0, 1
	}
	scale = math.Min(float64(availableWidth)/float64(sourceWidth), float64(availableHeight)/float64(sourceHeight))
	x = (float64(availableWidth) - float64(sourceWidth)*scale) / 2
	y = (float64(availableHeight) - float64(sourceHeight)*scale) / 2
	return x, y, scale
}

func (g *Game) handleKeyboard() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.togglePlayback()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.Engine.Restart()
		g.Engine.Play()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		g.Engine.Seek(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		g.seekBy(-.25)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		g.seekBy(.25)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit1) {
		g.Engine.SetSpeed(.5)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit2) {
		g.Engine.SetSpeed(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit3) {
		g.Engine.SetSpeed(2)
	}
}

func (g *Game) togglePlayback() {
	state := g.Engine.State()
	if state.Playing {
		g.Engine.Pause()
		return
	}
	if state.Duration > 0 && state.Time >= state.Duration {
		g.Engine.Restart()
	}
	g.Engine.Play()
}

func (g *Game) seekBy(delta float64) { g.Engine.Seek(g.Engine.State().Time + delta) }

func (g *Game) syncPlayerState() {
	state := g.Engine.State()
	if state.Playing {
		g.playButton.SetText("Pause")
	} else {
		g.playButton.SetText("Play")
	}
	g.timeLabel.Label = fmt.Sprintf("%s / %s", formatTime(state.Time), formatTime(state.Duration))
	if state.Error != nil {
		g.statusLabel.Label = "Script error: " + state.Error.Error()
	} else {
		mode := "Paused"
		if state.Playing {
			mode = "Playing"
		}
		g.statusLabel.Label = fmt.Sprintf("%s  |  %.1fx  |  Space: play/pause  R: restart  Arrows: seek", mode, state.Speed)
	}
	progress := 0
	if state.Duration > 0 {
		progress = int(math.Round(state.Time / state.Duration * timelineSteps))
	}
	if progress < 0 {
		progress = 0
	}
	if progress > timelineSteps {
		progress = timelineSteps
	}
	g.expectedSlider = progress
	g.timelineSlider.Current = progress
}

func (g *Game) buildUI() {
	face := loadFace(16)
	textColor := &widget.LabelColor{Idle: color.NRGBA{R: 226, G: 232, B: 240, A: 255}}
	root := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewAnchorLayout()))
	panel := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(uiimage.NewNineSliceColor(color.NRGBA{R: 20, G: 25, B: 36, A: 248})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionVertical), widget.RowLayoutOpts.Padding(&widget.Insets{Left: 16, Right: 16, Top: 10, Bottom: 8}), widget.RowLayoutOpts.Spacing(7))),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{HorizontalPosition: widget.AnchorLayoutPositionCenter, VerticalPosition: widget.AnchorLayoutPositionEnd, StretchHorizontal: true}), widget.WidgetOpts.MinSize(0, playerHeight)),
	)

	g.timelineSlider = widget.NewSlider(
		widget.SliderOpts.MinMax(0, timelineSteps),
		widget.SliderOpts.InitialCurrent(0),
		widget.SliderOpts.Images(sliderTrack(), sliderHandle()),
		widget.SliderOpts.FixedHandleSize(16),
		widget.SliderOpts.TrackOffset(0),
		widget.SliderOpts.PageSizeFunc(func() int { return 100 }),
		widget.SliderOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}), widget.WidgetOpts.MinSize(0, 18)),
		widget.SliderOpts.ChangedHandler(func(args *widget.SliderChangedEventArgs) {
			if !args.Dragging && args.Current == g.expectedSlider {
				return
			}
			duration := g.Engine.State().Duration
			g.Engine.Seek(duration * float64(args.Current) / timelineSteps)
		}),
	)

	controls := widget.NewContainer(widget.ContainerOpts.Layout(widget.NewRowLayout(widget.RowLayoutOpts.Direction(widget.DirectionHorizontal), widget.RowLayoutOpts.Spacing(8))))
	restart := g.button("Restart", face, func() { g.Engine.Restart(); g.Engine.Play() })
	back := g.button("-0.25s", face, func() { g.seekBy(-.25) })
	g.playButton = g.button("Pause", face, g.togglePlayback)
	forward := g.button("+0.25s", face, func() { g.seekBy(.25) })
	slow := g.button("0.5x", face, func() { g.Engine.SetSpeed(.5) })
	normal := g.button("1x", face, func() { g.Engine.SetSpeed(1) })
	fast := g.button("2x", face, func() { g.Engine.SetSpeed(2) })
	g.timeLabel = widget.NewLabel(widget.LabelOpts.Text("00:00.000 / 00:00.000", &face, textColor), widget.LabelOpts.TextOpts(widget.TextOpts.Position(widget.TextPositionEnd, widget.TextPositionCenter), widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Position: widget.RowLayoutPositionCenter, Stretch: true}))))
	controls.AddChild(restart, back, g.playButton, forward, slow, normal, fast, g.timeLabel)

	g.statusLabel = widget.NewLabel(widget.LabelOpts.Text("Ready", &face, &widget.LabelColor{Idle: color.NRGBA{R: 148, G: 163, B: 184, A: 255}}), widget.LabelOpts.TextOpts(widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}))))
	panel.AddChild(g.timelineSlider, controls, g.statusLabel)
	root.AddChild(panel)
	g.ui = &ebitenui.UI{Container: root}
}

func (g *Game) button(label string, face text.Face, clicked func()) *widget.Button {
	return widget.NewButton(
		widget.ButtonOpts.Image(buttonImage()),
		widget.ButtonOpts.Text(label, &face, &widget.ButtonTextColor{Idle: color.NRGBA{R: 241, G: 245, B: 249, A: 255}, Hover: color.White, Pressed: color.White}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 12, Right: 12, Top: 5, Bottom: 5}),
		widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(64, 30)),
		widget.ButtonOpts.DisableDefaultKeys(),
		widget.ButtonOpts.ClickedHandler(func(*widget.ButtonClickedEventArgs) { clicked() }),
	)
}

func buttonImage() *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle:    uiimage.NewBorderedNineSliceColor(color.NRGBA{R: 43, G: 52, B: 68, A: 255}, color.NRGBA{R: 71, G: 85, B: 105, A: 255}, 1),
		Hover:   uiimage.NewBorderedNineSliceColor(color.NRGBA{R: 55, G: 67, B: 87, A: 255}, color.NRGBA{R: 96, G: 165, B: 250, A: 255}, 1),
		Pressed: uiimage.NewBorderedNineSliceColor(color.NRGBA{R: 30, G: 64, B: 110, A: 255}, color.NRGBA{R: 59, G: 130, B: 246, A: 255}, 1),
	}
}

func sliderTrack() *widget.SliderTrackImage {
	return &widget.SliderTrackImage{Idle: uiimage.NewNineSliceColor(color.NRGBA{R: 51, G: 65, B: 85, A: 255}), Hover: uiimage.NewNineSliceColor(color.NRGBA{R: 71, G: 85, B: 105, A: 255})}
}

func sliderHandle() *widget.ButtonImage {
	return &widget.ButtonImage{Idle: uiimage.NewNineSliceColor(color.NRGBA{R: 59, G: 130, B: 246, A: 255}), Hover: uiimage.NewNineSliceColor(color.NRGBA{R: 96, G: 165, B: 250, A: 255}), Pressed: uiimage.NewNineSliceColor(color.NRGBA{R: 37, G: 99, B: 235, A: 255})}
}

func loadFace(size float64) text.Face {
	primarySource, err := text.NewGoTextFaceSource(bytes.NewReader(fontasset.JetBrainsMonoRegular()))
	if err != nil {
		panic(fmt.Errorf("load player primary font: %w", err))
	}
	fallbackSource, err := text.NewGoTextFaceSource(bytes.NewReader(fontasset.SourceHanSansSCRegular()))
	if err != nil {
		panic(fmt.Errorf("load player fallback font: %w", err))
	}
	face, err := text.NewMultiFace(
		&text.GoTextFace{Source: primarySource, Size: size},
		&text.GoTextFace{Source: fallbackSource, Size: size},
	)
	if err != nil {
		panic(fmt.Errorf("create player font fallback: %w", err))
	}
	return face
}

func formatTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	minutes := int(seconds) / 60
	secs := int(seconds) % 60
	millis := int(math.Round((seconds - math.Floor(seconds)) * 1000))
	if millis == 1000 {
		secs++
		millis = 0
		if secs == 60 {
			minutes++
			secs = 0
		}
	}
	return fmt.Sprintf("%02d:%02d.%03d", minutes, secs, millis)
}
