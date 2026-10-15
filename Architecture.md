# Architecture

## 1. 项目定位

本项目是一个使用 Go 编写的实时动画引擎，整体设计受到 Manim 启发。

主要目标：

- 使用 JavaScript 描述动画
- 使用 `goja` 执行 JavaScript
- 支持 Hot Reload
- 支持实时动画预览
- 使用 Go 实现 Scene / Animation / Timeline
- 使用 `gg` 进行 2D Vector Rasterization
- 使用 Ebitengine 提供实时窗口和画面显示
- 动画系统与 Renderer 解耦
- 支持 PNG、GIF、MP4 等输出

整体流程：

```text
scene.js
   ↓
 goja
   ↓
Scene / Animation / Timeline
   ↓
Mobject
   ↓
Renderer
   ↓
gg
   ↓
image.RGBA
   ↓
Ebitengine
   ↓
Realtime Preview
```

---

# 2. Overall Architecture

```text
┌─────────────────────────────────────────────┐
│                User Script                  │
│                  scene.js                   │
│                                             │
│ circle()                                    │
│ add()                                       │
│ play()                                      │
│ moveTo()                                    │
│ create()                                    │
│ parallel()                                  │
└────────────────────┬────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────┐
│                 Script Layer                │
│                                             │
│                  goja VM                    │
│                     │                       │
│                JS API Binding               │
│                     │                       │
│               Script Loader                 │
└────────────────────┬────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────┐
│                Animation Core               │
│                                             │
│   Scene                                     │
│   Timeline                                  │
│   Animation                                 │
│   Mobject                                   │
│   Updater                                   │
│   ValueTracker                              │
│   Easing                                    │
└────────────────────┬────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────┐
│                 Renderer                    │
│                                             │
│       Mobject → Vector / Geometry           │
│                     │                       │
│                     ▼                       │
│                     gg                      │
│                     │                       │
│                image.RGBA                   │
└────────────────────┬────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────┐
│                 Preview                     │
│                                             │
│               Ebitengine                    │
│                     │                       │
│                  Window                     │
│                     │                       │
│               Realtime Play                 │
└─────────────────────────────────────────────┘
```

---

# 3. Core Design Principle

核心原则：

> Animation 负责对象如何变化，Renderer 负责对象如何被绘制。

Animation 不依赖：

```text
gg
Ebitengine
OpenGL
GPU API
```

Renderer 不需要知道：

```text
MoveTo
FadeIn
Sequence
Timeline
```

两者通过 `Mobject` 的状态连接：

```text
MoveTo
   ↓
修改 Circle.Position
   ↓
Renderer 读取 Circle.Position
   ↓
绘制 Circle
```

---

# 4. Project Structure

```text
animation/
│
├── cmd/
│   └── anim/
│       └── main.go
│
├── engine/
│   ├── engine.go
│   ├── clock.go
│   └── state.go
│
├── scene/
│   ├── scene.go
│   └── camera.go
│
├── mobject/
│   ├── mobject.go
│   ├── group.go
│   ├── circle.go
│   ├── rectangle.go
│   ├── line.go
│   ├── polygon.go
│   ├── path.go
│   └── text.go
│
├── animation/
│   ├── animation.go
│   ├── move.go
│   ├── scale.go
│   ├── rotate.go
│   ├── fade.go
│   ├── create.go
│   ├── transform.go
│   ├── sequence.go
│   ├── parallel.go
│   └── easing.go
│
├── timeline/
│   ├── timeline.go
│   └── track.go
│
├── geometry/
│   ├── vec2.go
│   ├── rect.go
│   ├── transform.go
│   ├── matrix.go
│   ├── bezier.go
│   └── path.go
│
├── renderer/
│   ├── renderer.go
│   └── gg/
│       └── renderer.go
│
├── preview/
│   ├── preview.go
│   └── ebiten.go
│
├── script/
│   ├── runtime.go
│   ├── loader.go
│   ├── api.go
│   └── bindings.go
│
├── hotreload/
│   ├── watcher.go
│   └── reload.go
│
└── export/
    ├── image.go
    ├── gif.go
    └── video.go
```

---

# 5. Engine

`Engine` 是整个运行时的入口。

负责管理：

```text
Scene
Timeline
Renderer
Script Runtime
Preview
Hot Reload
Clock
```

示例结构：

```go
type Engine struct {
    Scene    *Scene
    Timeline *Timeline

    Runtime  *ScriptRuntime
    Renderer Renderer

    Playing bool
    Time    float64
    Speed   float64
}
```

运行流程：

```text
Load Script
    ↓
Build Scene
    ↓
Run Preview
    ↓
Update Timeline
    ↓
Update Scene
    ↓
Render
    ↓
Present
```

---

# 6. Scene

`Scene` 表示一个完整动画场景。

```go
type Scene struct {
    Objects []Mobject
    Camera  *Camera
}
```

负责：

```text
Add Mobject
Remove Mobject
Update Mobject
Manage Camera
```

例如：

```go
scene.Add(circle)
scene.Remove(circle)
```

Scene 保存对象。

Timeline 管理这些对象随时间产生的变化。

---

# 7. Mobject

`Mobject` 是所有可动画、可绘制对象的基础抽象。

包括：

```text
Circle
Rectangle
Line
Polygon
Path
Text
Group
```

可以定义：

```go
type Mobject interface {
    Update(dt float64)

    Position() Vec2
    SetPosition(Vec2)

    Transform() *Transform

    Visible() bool
}
```

公共状态：

```go
type BaseMobject struct {
    Position Vec2

    Rotation float64

    Scale Vec2

    Opacity float64

    Visible bool
}
```

动画修改这些状态。

Renderer 根据这些状态进行绘制。

---

# 8. Geometry

Geometry 描述对象的几何信息。

例如：

```text
Circle
Rectangle
Polygon
Bezier
Path
```

Mobject 包含：

```text
Geometry
+
Transform
+
Style
+
State
```

例如：

```go
type Circle struct {
    BaseMobject

    Radius float64

    Fill   Color
    Stroke Color

    StrokeWidth float64
}
```

---

# 9. Path

`Path` 是矢量动画系统的重要基础。

支持：

```text
MoveTo
LineTo
QuadraticTo
CubicTo
Close
```

例如：

```go
type Path struct {
    Commands []PathCommand
}
```

Path 用于实现：

```text
Create
TrimPath
Morph
Transform
MoveAlongPath
```

`geometry.Path` 不直接依赖 `gg.Path`。

而是：

```text
geometry.Path
      ↓
Renderer
      ↓
gg.MoveTo()
gg.LineTo()
gg.QuadraticTo()
gg.CubicTo()
```

这样 Geometry 与具体渲染实现保持解耦。

---

# 10. Animation

Animation 描述对象随时间产生的变化。

```go
type Animation interface {
    Duration() float64

    Start()

    Update(progress float64)

    Finish()
}
```

其中：

```text
progress

0.0 -------------------------- 1.0
start                           end
```

例如 `MoveTo`：

```go
type MoveTo struct {
    Target Mobject

    From Vec2
    To   Vec2
}
```

更新：

```go
func (a *MoveTo) Update(t float64) {
    p := LerpVec2(
        a.From,
        a.To,
        t,
    )

    a.Target.SetPosition(p)
}
```

---

# 11. Animation Types

基础动画：

```text
MoveTo
MoveBy

ScaleTo
ScaleBy

RotateTo
RotateBy

FadeIn
FadeOut

OpacityTo
ColorTo

Create
Uncreate

Transform
ReplacementTransform

MoveAlongPath
```

动画组合：

```text
Sequence
Parallel
Stagger
Repeat
Loop
Yoyo
```

---

# 12. Easing

Easing 将线性时间：

```text
0 → 1
```

映射为动画进度：

```text
0 → Ease(t) → 1
```

接口：

```go
type Easing func(t float64) float64
```

支持：

```text
Linear

EaseIn
EaseOut
EaseInOut

Quad
Cubic
Quart
Quint

Sine
Expo
Circ

Back
Bounce
Elastic

Smooth
```

---

# 13. Timeline

Timeline 是动画调度系统。

Timeline 保存：

```text
Animation
Start Time
Duration
```

例如：

```text
Time

0s          1s          2s          3s

Create
████████████

            Move
            █████████████████████

            Fade
            ███████████
```

结构：

```go
type Timeline struct {
    Tracks []Track

    Time     float64
    Duration float64
}
```

Track：

```go
type Track struct {
    Start float64

    Animation Animation
}
```

Timeline 可以根据任意时间点重新计算动画状态：

```go
timeline.Seek(currentTime)
```

这样可以支持：

```text
Play
Pause
Resume
Restart
Seek
Playback Speed
Offline Render
Hot Reload Time Restore
```

---

# 14. Deterministic Animation

动画状态必须由时间决定，而不是由帧数决定。

错误：

```go
circle.X += 2
```

因为：

```text
60 FPS
144 FPS
```

会产生不同结果。

正确：

```go
progress :=
    (currentTime - startTime) /
    duration
```

然后：

```go
circle.X = Lerp(
    startX,
    endX,
    progress,
)
```

因此：

```text
30 FPS
60 FPS
144 FPS
```

在相同时间点应该得到相同的 Scene 状态。

这也是以下能力的基础：

```text
Seek
Replay
Offline Rendering
Hot Reload
Timeline Editor
```

---

# 15. Renderer

Renderer 使用抽象接口。

```go
type Renderer interface {
    BeginFrame()

    DrawCircle(*Circle)
    DrawRectangle(*Rectangle)
    DrawLine(*Line)
    DrawPath(*Path)
    DrawText(*Text)

    EndFrame()

    Image() image.Image
}
```

Renderer 只负责：

> 将 Scene / Mobject 转换成最终图像。

---

# 16. gg Renderer

2D Vector Rasterization 使用：

```text
github.com/fogleman/gg
```

负责：

```text
Circle
Rectangle
Line
Path
Bezier
Fill
Stroke
Text
Transform
Clip
```

流程：

```text
Mobject
   ↓
GGRenderer
   ↓
gg.Context
   ↓
image.RGBA
```

Renderer 持有并复用：

```go
*image.RGBA
*gg.Context
```

避免每帧创建新的 framebuffer 和 Context。

---

# 17. Ebitengine Preview

Ebitengine 负责：

```text
Window
Event Loop
Realtime Preview
Input
GPU Presentation
```

渲染流程：

```text
gg
 ↓
image.RGBA
 ↓
ebiten.Image.WritePixels()
 ↓
screen.DrawImage()
 ↓
Window
```

运行循环：

```text
Update()
   ↓
Clock
   ↓
Timeline
   ↓
Scene

Draw()
   ↓
Renderer
   ↓
RGBA Frame
   ↓
Ebitengine
   ↓
Present
```

---

# 18. Script Runtime

用户使用 JavaScript 描述 Scene。

例如：

```js
const c = circle({
    radius: 50,
    position: [100, 300],
    fill: "#3b82f6",
})

add(c)

play(
    create(c),
)

play(
    moveTo(c, [800, 300]),
    {
        duration: 2,
        ease: "smooth",
    },
)
```

JavaScript 使用：

```text
goja
```

执行。

---

# 19. JavaScript API

## Mobject

```js
circle()
rect()
line()
polygon()
path()
text()
group()
```

## Scene

```js
add()
remove()
```

## Animation

```js
play()
wait()

moveTo()
moveBy()

scaleTo()
scaleBy()

rotateTo()
rotateBy()

fadeIn()
fadeOut()

create()
uncreate()

transform()
moveAlongPath()
```

## Composition

```js
sequence()
parallel()
stagger()
repeat()
```

---

# 20. JavaScript Responsibilities

JavaScript 负责：

> 描述 Scene 和 Animation。

JavaScript 不负责：

```text
Window Loop
Renderer Loop
GPU
Framebuffer
Frame Scheduling
```

不应该：

```js
while (true) {
    update()
    draw()
}
```

而应该：

```text
JavaScript
    ↓
Scene Description
    ↓
Animation Description
    ↓
Go Timeline
    ↓
Go Engine
```

真正的运行循环由 Go / Ebitengine 管理。

---

# 21. Updater

除了 Timeline Animation，也允许对象拥有 Updater。

JavaScript：

```js
circle.update((self, time, dt) => {
    self.y = Math.sin(time) * 100
})
```

Go：

```go
type Updater func(
    obj Mobject,
    time float64,
    dt float64,
)
```

Updater 适合：

```text
物体跟随
数学函数
程序化动画
粒子
动态图形
```

普通动画尽量使用 Go Animation 实现：

```text
MoveTo
Fade
Scale
Rotate
Transform
```

避免大量对象每帧频繁调用 JavaScript VM。

---

# 22. ValueTracker

`ValueTracker` 表示一个可以被 Animation 控制的数值。

例如：

```js
const angle = value(0)

circle.update(() => {
    circle.x = Math.cos(angle.value) * 200
    circle.y = Math.sin(angle.value) * 200
})

play(
    animateValue(angle, Math.PI * 2),
    {
        duration: 3,
    },
)
```

适用于：

```text
数学动画
参数动画
图表动画
程序化动画
```

---

# 23. Hot Reload

Hot Reload 使用：

```text
fsnotify
+
goja
```

流程：

```text
scene.js
   ↓
File Changed
   ↓
Debounce
   ↓
Read Script
   ↓
Create New goja VM
   ↓
Register API
   ↓
Execute Script
   ↓
Build New Scene
   ↓
Swap Scene
```

这样 JavaScript 修改后不需要重新编译整个 Go 应用。

---

# 24. Transactional Hot Reload

Reload 不立即销毁当前有效 Scene。

流程：

```text
Current Scene
     │
     └───────────────┐
                     │
              Create New VM
                     ↓
               Run Script
                     │
          ┌──────────┴──────────┐
          │                     │
        Error                Success
          │                     │
Keep Current Scene        Swap Scene
```

新脚本执行失败：

```text
Current Scene continues
+
Show Script Error
```

新脚本执行成功：

```text
Old Scene
   ↓
New Scene
```

从而避免代码输入错误导致 Preview 直接消失。

---

# 25. VM Lifecycle

每次 Reload 创建独立的新：

```go
vm := goja.New()
```

然后：

```text
New VM
 ↓
Register API
 ↓
Execute Script
 ↓
Build Scene
 ↓
Replace Runtime
```

这样可以避免旧 VM 中：

```text
Global Variables
Callbacks
Closures
Timers
Temporary State
```

污染新的 Scene。

---

# 26. Hot Reload Timeline State

Reload 可以保存当前时间：

```text
Current Time
    ↓
5.42s
    ↓
Reload Script
    ↓
Build Scene
    ↓
Timeline.Seek(5.42)
    ↓
Render
```

用户修改：

```js
moveTo(circle, [600, 300])
```

保存后，可以直接看到当前时间位置对应的新结果，而不是必须从头播放。

也可以支持：

```text
Restart On Reload
Preserve Time On Reload
```

两种行为。

---

# 27. File Watcher

文件监听使用：

```text
fsnotify
```

监听：

```text
.js
.json
.svg
fonts
images
assets
```

由于编辑器保存可能产生：

```text
WRITE
WRITE
RENAME
CREATE
WRITE
```

需要增加 debounce：

```text
File Changed
    ↓
100~200ms Debounce
    ↓
Reload
```

---

# 28. Error Handling

Script Error 不应该导致 Preview 退出。

例如：

```text
ReferenceError:
circle2 is not defined

scene.js:18
```

Preview 可以继续显示上一次有效 Scene：

```text
┌──────────────────────────────────────────┐
│                                          │
│           Previous Valid Scene           │
│                                          │
├──────────────────────────────────────────┤
│ Script Error                             │
│                                          │
│ ReferenceError: circle2 is not defined   │
│ scene.js:18                              │
└──────────────────────────────────────────┘
```

修复代码并保存：

```text
Reload
 ↓
Success
 ↓
Error disappears
```

---

# 29. Preview Controls

Preview 支持：

```text
Space       Play / Pause

R           Restart

← / →       Seek

Home        Jump To Start

0.5x
1x
2x          Playback Speed
```

内部 API：

```go
Play()

Pause()

Resume()

Restart()

Seek(time)

SetSpeed(speed)
```

---

# 30. Clock

Clock 是 Timeline 的时间来源。

```go
type Clock interface {
    Time() float64
}
```

Realtime Preview 使用：

```text
Realtime Clock
```

Export 使用：

```text
Fixed Step Clock
```

这样 Animation Core 不需要知道当前是在：

```text
Preview
```

还是：

```text
Export
```

---

# 31. Export

Realtime Preview 与 Export 共用：

```text
Scene
Timeline
Animation
Mobject
Renderer
```

区别在于 Clock。

Realtime：

```text
Wall Clock
 ↓
Timeline
```

Export：

```text
Fixed Time Step
 ↓
Timeline
```

例如 60 FPS：

```text
0.000000
0.016667
0.033333
0.050000
...
```

每个时间点都生成确定性的画面。

---

# 32. Video Export

视频输出流程：

```text
Scene
 ↓
Timeline
 ↓
Renderer
 ↓
RGBA Frame
 ↓
FFmpeg stdin
 ↓
MP4
```

可以直接通过 pipe 将 raw frame 输入 FFmpeg：

```text
Go Renderer
     ↓
Raw RGBA
     ↓
FFmpeg
     ↓
Video
```

不需要依赖大量中间 PNG 文件。

---

# 33. Renderer Backends

Renderer 是可替换的。

整体设计：

```text
                     Renderer
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
    GG Renderer    GPU Renderer    SVG Renderer
        │
        ▼
   image.RGBA
```

Scene、Animation、Timeline、Mobject 和 JavaScript API 不依赖具体 Renderer。

因此切换渲染实现不会影响动画核心。

---

# 34. Realtime Rendering Pipeline

完整 realtime pipeline：

```text
                  scene.js
                     │
                     ▼
                   goja
                     │
                     ▼
                   Scene
                     │
        ┌────────────┴────────────┐
        │                         │
        ▼                         ▼
     Mobject                   Timeline
                                  │
                                  ▼
                              Animation
        │                         │
        └────────────┬────────────┘
                     │
                     ▼
                 GG Renderer
                     │
                     ▼
                 image.RGBA
                     │
                     ▼
              Ebitengine Image
                     │
                     ▼
                   Window
```

每一帧：

```text
Clock
  ↓
Timeline.Seek(time)
  ↓
Update Mobject State
  ↓
Render Scene
  ↓
gg rasterize
  ↓
RGBA
  ↓
Ebitengine Present
```

---

# 35. Hot Reload Pipeline

```text
scene.js
   │
   ▼
fsnotify
   │
   ▼
debounce
   │
   ▼
New goja VM
   │
   ▼
Register JS API
   │
   ▼
Run Script
   │
   ▼
Build Scene + Timeline
   │
   ├──── Error ────→ Keep Current Scene
   │
   ▼
Success
   │
   ▼
Swap Runtime
   │
   ▼
Seek Previous Time
   │
   ▼
Realtime Preview
```

---

# 36. Example

```js
const title = text("Hello Animation", {
    size: 48,
})

const circle = circle({
    radius: 50,
    position: [100, 300],
    fill: "#3b82f6",
})

add(title)
add(circle)

play(
    create(title),
    {
        duration: 1,
    },
)

play(
    parallel(
        moveTo(circle, [500, 300]),
        scaleTo(circle, 2),
        rotateBy(circle, Math.PI),
    ),
    {
        duration: 2,
        ease: "smooth",
    },
)

wait(1)

play(
    fadeOut(circle),
)
```

运行：

```bash
anim preview scene.js
```

输出：

```text
Watching scene.js...

✓ Script loaded
✓ Scene created
✓ Preview started

scene.js changed

✓ Reloaded
```

开发流程：

```text
Write JavaScript
      ↓
Save
      ↓
Hot Reload
      ↓
Realtime Preview
      ↓
Modify
      ↓
Save
      ↓
Realtime Preview
```

---

# 37. Architecture Summary

```text
                      User
                       │
                       ▼
                   scene.js
                       │
                       ▼
                ┌────────────┐
                │   goja VM  │
                └──────┬─────┘
                       │
                       ▼
                 JavaScript API
                       │
                       ▼
        ┌──────────────────────────┐
        │       Animation Core     │
        │                          │
        │ Scene                    │
        │ Timeline                 │
        │ Animation                │
        │ Mobject                  │
        │ Updater                  │
        │ ValueTracker             │
        └────────────┬─────────────┘
                     │
                     ▼
                  Geometry
                     │
                     ▼
                 Renderer
                     │
                     ▼
                     gg
                     │
                     ▼
                image.RGBA
                     │
                     ▼
                 Ebitengine
                     │
                     ▼
              Realtime Preview


scene.js
   ▲
   │
fsnotify
   │
   └──────── Hot Reload ───────→ goja VM
```

核心职责：

```text
JavaScript
→ 描述动画

goja
→ 执行脚本

Scene
→ 管理动画对象

Mobject
→ 保存对象状态

Animation
→ 描述状态变化

Timeline
→ 管理时间

Geometry
→ 描述矢量图形

gg
→ Rasterization

Ebitengine
→ Realtime Window

fsnotify
→ Hot Reload

FFmpeg
→ Video Export
```

整个架构保持：

```text
Script
Animation Core
Renderer
Preview
Export
```

相互解耦，使动画系统、脚本系统与渲染后端能够独立扩展。