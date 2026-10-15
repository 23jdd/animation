# Animation

A compact real-time animation engine in Go, driven by JavaScript. Its layers follow [`Architecture.md`](Architecture.md): JavaScript describes a scene, a deterministic Go timeline changes mobject state, the renderer rasterizes that state with `gg`, and Ebitengine presents the result.

[中文](README_ZN.md)


## Preview
<img src="animation.gif" align="center"/>

## Run it

The included `index.js` is a complete example.
```powershell
go run . preview index.js
go run . render -width 1280 -height 720 -time 2 index.js frame.png
go run . gif -width 1280 -height 720 -fps 30 index.js animation.gif
go run . mp4 -width 1280 -height 720 -fps 60 index.js animation.mp4
```

MP4 export requires `ffmpeg` on `PATH`. PNG and GIF export have no external runtime dependency. Install the command with `go install ./cmd/anim`, then use `anim` in place of `go run .`.

Use `-scale 2` with `render`, `gif`, or `mp4` to double the output pixel dimensions without changing scene layout. For example: `go run . gif -scale 2 -fps 30 index.js animation-2x.gif` renders the default 1280×720 scene at 2560×1440.

GIF export streams frames to disk, reports progress, and uses an accelerated color lookup with adaptive dominant colors and error diffusion. Memory usage does not grow with the number of frames. GIF supports up to 100 fps; frame delays are distributed to preserve the requested average rate (including 30/60 fps).

The preview includes a responsive EbitenUI animation player below the canvas. Resizing the window reflows the controls while the animation canvas is centered and scaled without changing its aspect ratio. The player provides a draggable timeline, play/pause, restart, frame seeking, playback-speed buttons, current/duration time, playback state, and script-error display. It also watches the script and nearby `.js`, `.json`, `.svg`, font, and image assets. Reloads are transactional: an invalid edit leaves the last good scene visible and shows the script error. The current timeline position is preserved after a successful reload.

Preview controls:

| Key | Action |
|---|---|
| Space | Play / pause |
| R | Restart |
| Left / Right | Seek by 0.25 seconds |
| Home | Jump to start |
| 1 / 2 / 3 | 0.5x / 1x / 2x speed |

## JavaScript API

完整接口说明见 [`API_REFERENCE.md`](API_REFERENCE.md)。

Create mobjects with `circle`, `rect`, `line`, `polygon`, `path`, `text`, and `group`, then add them with `add(...)`.

```js
const dot = circle({
  radius: 40,
  position: [-300, 0],
  fill: "#3b82f6",
  stroke: "#bfdbfe",
  strokeWidth: 3,
});

add(dot);
play(create(dot), { duration: 0.5 });
play(moveTo(dot, [300, 0]), { duration: 2, ease: "smooth" });
```

Animations include `moveTo`, `moveBy`, `scaleTo`, `scaleBy`, `rotateTo`, `rotateBy`, `fadeIn`, `fadeOut`, `opacityTo`, `colorTo`, `create`, `uncreate`, `write`, `unwrite`, `drawLine`, `drawLineFromCenter`, `transform`, `replacementTransform`, `moveAlongPath`, `growFromCenter`, `shrinkToCenter`, `growFromPoint`, `growFromEdge`, `spinIn`, `spinOut`, `flash`, `indicate`, `blink`, `wiggle`, `wave`, `zoomIn`, `zoomOut`, and `focusOn`. Compose them with `sequence`, `parallel`, `stagger`, `repeat`, `loop`, and `yoyo`.

Custom TrueType/OpenType fonts are supported by `text()` and are resolved relative to the scene script:

```js
text("实时动画", {
  font: "./assets/fonts/NotoSansSC-Regular.ttf",
  size: 48,
});
```

When `font` is omitted, scene text and the EbitenUI player use an embedded font stack: JetBrains Mono Regular is preferred for Latin text and Source Han Sans SC Regular is selected automatically for missing Chinese/CJK glyphs. The application does not depend on fonts installed on the operating system. Both SIL Open Font License 1.1 texts are included in [`fontasset/assets`](fontasset/assets).

Supported easing names include `linear`, `easeIn`, `easeOut`, `easeInOut`, `cubic`, `sine`, `expo`, `circ`, `back`, `bounce`, `elastic`, `smooth`, and `smoother`.

Paths use renderer-neutral commands:

```js
const route = path({
  commands: [
    ["M", -300, 0],
    ["C", -100, -200, 100, 200, 300, 0],
  ],
  stroke: "#22c55e",
  fill: "#00000000",
});
```

Use `value(initial)` with `animateValue`, or attach an updater for procedural motion:

```js
const angle = value(0);
dot.update((self) => {
  self.x = Math.cos(angle.value) * 200;
  self.y = Math.sin(angle.value) * 200;
});
play(animateValue(angle, Math.PI * 2), { duration: 3 });
```

All offline frames are produced by seeking the same timeline used in preview, so output is independent of preview frame rate.

## Editor completion with JSDoc

[`animation-api.js`](animation-api.js) documents every JavaScript global, option object, path command, easing name, updater, and animation handle. [`jsconfig.json`](jsconfig.json) enables JavaScript type checking and completion in editors that understand JSDoc.

Add this as the first line of a scene script:

```js
/// <reference path="./animation-api.js" />
```

The included [`index.js`](index.js) already contains the reference. The declarations are editor-only and are never executed by the animation runtime.
