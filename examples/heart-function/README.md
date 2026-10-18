# 把一个心形，写成函数

第二支简洁数学科普短片：暖白底、玫瑰色曲线、轻微的呼吸感与缓慢转场。时长约 44 秒，1280 × 720、30 fps，中文字幕，无配音、无音乐。

动画全部用 JavaScript 编写，使用当前 Go 动画引擎绘制，不加载图片或外部 JS 库。这是一个数学心形，不是人体心脏的解剖图，也不是心电图。

## 文件

- [heart-function.mp4](heart-function.mp4)：视频成片，中文字幕已绘入画面。
- [preview.gif](preview.gif)：960 × 540、12 fps 的循环预览。
- [animation.js](animation.js)：完整动画源文件，公式、采样、描线和镜头都在这里。
- [subtitles.srt](subtitles.srt)：对应字幕与配音参考；片头标题承担第一条字幕，不重复叠字。
- [cover.png](cover.png)：片头封面。
- [poster.png](poster.png)：心形完成时的公式画面。
- [animation_test.go](animation_test.go)：坐标、对称性、描线终点和回放测试。

## 公式

```text
x(t) = 16 sin³(t)
y(t) = 13 cos(t) − 5 cos(2t) − 2 cos(3t) − cos(4t)
0 ≤ t ≤ 2π
```

同一个参数 `t` 同时决定 `x` 和 `y`，得到点 `P(t) = (x(t), y(t))`。让参数连续变化，点的轨迹就是心形。

严格地说，这是一条**参数曲线**，不是单个 `y = f(x)`：例如 `x = 0` 时，曲线上既有 `(0, 5)`，又有 `(0, -17)`。

数学坐标中 y 向上，引擎画布中 y 向下，所以绘图时使用 `[scale*x, -scale*y]`，不改变数学公式本身。

## 分镜

| 时间 | 内容 |
| --- | --- |
| 0–5 秒 | 一个轻轻呼吸的心形，引出主题 |
| 5–12 秒 | 两个公式、一个参数，以及点的横纵坐标投影 |
| 12–28 秒 | 从顶部凹口出发，逐点描出完整心形；关键坐标稍作停留 |
| 28–36 秒 | 一对镜像点同步运动，说明为什么左右对称 |
| 36–44 秒 | 闭合曲线、温柔的呼吸与结束字幕 |

描线会在以下坐标停留，便于观察：

| 参数 | 点的位置 |
| --- | --- |
| 0 | (0, 5)，顶部凹口 |
| π/2 | (16, 4)，右端 |
| π | (0, -17)，下方尖点 |
| 3π/2 | (-16, 4)，左端 |
| 2π | (0, 5)，回到起点 |

## 为什么左右对称

把 `t` 换成 `2π − t`：

```text
sin(2π − t) = −sin(t)
cos(n(2π − t)) = cos(nt)       对这里的整数 n = 1, 2, 3, 4
```

因此：

```text
x(2π − t) = −x(t)
y(2π − t) = y(t)
```

横坐标互为相反数，纵坐标相同，所以关于 y 轴对称。

## 如何计算 scale

源码采样 1600 个区间（包含两个端点，共 1601 个点），计算实际包围盒。横坐标范围是 `[-16,16]`，纵坐标约为 `[-17,11.923]`。

如果目标显示区宽高是 `W,H`，保持比例的统一倍率为：

```text
scale = min(W / (maxX − minX), H / (maxY − minY))
```

横纵方向使用同一个倍率，避免心形被拉宽或拉长。标题、公式和字幕不参与形状缩放，静止时保持原生字号。

## 描线为什么能跟准那个点

动画按照参数 `t` 揭示曲线，而不是简单地按曲线弧长播放描线。两者并不等价：这个心形在相同的参数增量下，移动距离并不固定。

源码把完整曲线拆为 50 个区块；完成的区块各用一条 path 绘制，当前区块只使用 32 条可复用线段。最后一小段的终点直接取 `heart(t)`，因此移动点与描线头一致，也不必每帧绘制 1600 个独立对象。

所有图形加载时创建一次，每帧只按绝对时间修改现有对象，支持任意时间跳转与重复导出。

## 预览和导出

在项目根目录执行：

```powershell
go run . preview examples/heart-function/animation.js

go run . mp4 -fps 30 examples/heart-function/animation.js examples/heart-function/heart-function.mp4

go run . gif -fps 12 -scale 0.75 examples/heart-function/animation.js examples/heart-function/preview.gif

go run . render -time 27.4 examples/heart-function/animation.js examples/heart-function/poster.png
```

重新导出 MP4 需要 `ffmpeg` 位于当前终端的 `PATH`。播放已生成的成片不需要开发环境，GIF、PNG 和实时预览也不需要 FFmpeg。

## 验证

```powershell
go test ./examples/heart-function -v
go test ./...
```

测试检查五个关键坐标、65 个参数位置的镜像关系、包围盒等比例适配、部分线段的实际终点是否等于 `heart(t)`，以及跳转到片尾后再返回同一时刻的画面像素哈希是否一致。

成品已完整解码检查：MP4 为 1280 × 720、30 fps、1321 帧、44.03 秒；GIF 为 960 × 540、529 帧、44.08 秒。导出器包含末尾一帧，因此媒体时长略长于场景的 44 秒。另已检查编码后的视频抽帧。

![片头封面](cover.png)
