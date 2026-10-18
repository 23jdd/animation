# PPTX 导出：分镜与原生动画

PPTX 导出由 Go 标准库直接实现，不需要商业授权密钥、Node.js、FFmpeg 或安装 Office。

默认模式把选定时间点的画面作为图片分镜；`-native` 模式把整段时间轴转换为一页可编辑对象和 PowerPoint 原生动画。两种模式都不依赖 Office 来生成文件。

## 原生动画

```powershell
go run . pptx -native examples/pptx-native/animation.js native-animation.pptx
```

在 PowerPoint 中启动放映后自动播放。文字仍是文字，圆和矩形仍是形状，可以在动画窗格中查看、修改动画。普通编辑视图展示初始布局，实际效果请在放映中查看。

- 支持圆、矩形、圆角矩形、直线、多边形、贝塞尔路径、文字；组被展平为独立对象并保留变换和透明度。
- 支持位置、正数缩放、旋转，以及完整的淡入淡出。`moveTo` / `moveBy`、`scaleTo` / `scaleBy`、`rotateTo` / `rotateBy` 与组合时间轴无需改写。
- `sequence`、`parallel`、`stagger`、`repeat`、`yoyo` 和确定性的 updater 都通过现有时间轴求值；前提是它们只改变支持的属性。
- `-fps 60` 控制状态采样精度，范围为 1～120。位置用原生属性关键帧，缩放和旋转用原生行为；缓动曲线经过采样和折线简化，属于近似转换，不是逐像素复现。短于采样间隔的效果可能丢失，应提高 `-fps`。
- 淡入淡出使用 PowerPoint 自身的 fade 插值，保留采样得到的开始/结束时间，不保证 JS 的自定义透明度缓动一致；不支持任意局部透明度变化。
- `-native` 与 `-time`、`-times`、`-interval` 互斥。原生对象没有图片分辨率，`-scale` 不改变它们。
- 字体按名称引用，不嵌入 PPTX；使用自定义字体时，播放电脑也需要安装该字体，否则演示软件可能替换。默认文字使用 Source Han Sans SC。
- 当前不支持图片、3D、`create` / `write` 等渐进绘制、动画中修改文本/路径/颜色/描边、反射、零/负缩放、剪切变换。检测到不支持的状态会报错，不会静默截图。
- 最多 300 秒、500 个叶对象、18000 个常规采样点、100 万对象采样。较长或较复杂的动画可降低 `-fps`，或使用图片分镜。

Go 调用：

```go
err := export.PPTXNative(output, scene, timeline, 60)
```

## 图片分镜

每个选定时间点对应一页幻灯片，以无损 PNG 保留当前渲染器的画面。文字、矢量图形、透明图片和 3D 场景均使用相同的渲染结果。每页备注记录对应的动画时间。

### 使用

在项目根目录运行，所有选项位于 JS 路径之前。

### 单帧

```powershell
go run . pptx -time 60 examples/rose-curve/animation.js rose.pptx
```

默认不传选择参数时，仅导出第 0 秒。若开场尚未淡入，应该选择稍后的时间点。

### 多页分镜

```powershell
go run . pptx -times 4,11,26,33,42,50,60 examples/rose-curve/animation.js rose-storyboard.pptx
```

保留时间列表中的顺序，也允许重复时间点。每个时间必须是 `0` 到动画总时长之间的有限数。

### 固定间隔

```powershell
go run . pptx -interval 8 examples/rose-curve/animation.js sampled.pptx
```

包括第 0 秒、各间隔时间和动画末尾。`-time`、`-times`、`-interval` 互斥。`-fps` 不控制 PPTX 的页数。

### 分辨率与比例

```powershell
go run . pptx -time 60 -scale 3 examples/rose-curve/animation.js high-resolution.pptx
```

- 默认 `-scale 2`，1280 × 720 的逻辑画布得到 2560 × 1440 的页面图片。
- `-scale` 只改变图片分辨率，不改变排版或幻灯片的物理尺寸。
- 页面保持逻辑画布比例，支持横向、纵向和自定义比例。常规画布长边为 13⅓ 英寸；16:9 时页面尺寸为 13⅓ × 7.5 英寸。
- `ppt` 是 `pptx` 的命令别名，输出扩展名仍必须是 `.pptx`，不支持旧版二进制 `.ppt`。
- 每次最多 500 页，每帧最多 6400 万像素。

## Go API

```go
err := export.PPTX(output, scene, timeline, renderer, []float64{4, 11, 26})
// 或 export.PPTXWithProgress(..., times, func(done, total int) { ... })
```

空时间切片默认选择第 0 秒。导出结束或失败后恢复原时间轴位置和场景状态。逐帧写入 ZIP，不同时保留所有页面图片；写入失败不会覆盖原有输出文件。

## 文件内容与限制

使用 `archive/zip`、`encoding/xml`、`image/png` 生成 OOXML 包，包含演示文稿、母版、空白布局、主题、每页图片、备注、内容类型和关系文件。导出过程离线运行，不上传画面。

不加 `-native` 时导出的是静态分镜文件：不会播放 JS 时间轴或音乐，画面中的文字和图形不会转换成可单独编辑的 PowerPoint 对象。可以在演示软件里移动、替换整页图片，或新增文本、批注和页面切换效果。原生模式也不嵌入音乐。

导出样例位于 `examples/rose-curve/rose-storyboard.pptx`。测试覆盖 ZIP/XML 结构、全部关系引用、页面比例、PNG 像素、时间顺序、状态恢复、参数校验以及失败时保留原文件。
