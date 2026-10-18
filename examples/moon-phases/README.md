# 月亮，为什么会变？

一个约 44 秒的简洁科普短片，使用当前 animation 引擎制作。

深蓝背景、月白图形、少量暖金、清晰中文字幕与缓慢转场。二维矢量图形全部用 JavaScript 绘制，不依赖图片贴图、网页或外部 JS 库。此版为**无配音、无音乐的字幕版**。

## 成品与源码

- [moon-phases.mp4](moon-phases.mp4)：1280 × 720、30 fps、H.264 视频，中文字幕已绘入画面。
- [preview.gif](preview.gif)：960 × 540、12 fps 的循环预览。
- [animation.js](animation.js)：完整 JS 场景，含月相形状、镜头编排和动画。
- [subtitles.srt](subtitles.srt)：对应字幕与配音文稿；标题画面承担第一条字幕的内容，未重复叠字。
- [cover.png](cover.png)：片头封面。
- [poster.png](poster.png)：月相原理镜头。
- [animation_test.go](animation_test.go)：亮面面积、月相名称和任意时间跳转的验证。

## 分镜

| 时间 | 内容 |
| --- | --- |
| 0–5 秒 | 用一轮缓慢变化的月亮提出问题 |
| 5–12 秒 | 太阳光照向月球，说明月光来自反射 |
| 12–28 秒 | 左侧显示日、地、月关系，右侧同步显示从地球看到的月相 |
| 28–35 秒 | 对比月相与月食，澄清地球影子的误解 |
| 35–44 秒 | 八种常见月相与约 29.5 天的循环 |

关键月相会稍作停留，便于看清新月、上弦月、满月和下弦月。实际时间已压缩；运动速度和尺寸不是天文数值模拟。

## 科普内容与示意范围

- 月亮不自己发光，我们看到的月光来自太阳光的反射。
- 在通常情况下，太阳照亮月球约一半；我们看到其中多大的部分，随月球与太阳、地球的相对位置改变。
- 日常月相不是地球的影子造成的；月球进入地球影子时才会发生月食。
- 从一次新月到下一次新月，平均约 29.5 天。这是月相周期，不是月球相对恒星公转一圈约 27.3 天的周期。
- 轨道画面是简化的二维投影，省略了轨道倾角；月食一幕说明倾角使月食并非每次满月都出现。
- 月相亮面的左右朝向采用北半球常见示意，实际朝向会随观测位置和天空方向改变。
- 画面里的星体大小、距离、月面纹理和月食颜色均为美术示意，不是照片或精确星历。

事实核对来源：

- [NASA：Moon Phases](https://science.nasa.gov/moon/moon-phases/)
- [NASA：Top Moon Questions](https://science.nasa.gov/moon/top-moon-questions/)
- [NASA：Eclipses and the Moon](https://science.nasa.gov/moon/eclipses/)

## 如何绘制连续变化的月相

用 `angle` 表示从新月起算、相对太阳方向的演示角度。月面亮部是半圆与一个投影椭圆的组合，椭圆的水平半径为：

```text
|cos(angle)| × 月面半径
```

相应的可见亮面面积比例为：

```text
(1 − cos(angle)) / 2
```

这样可以连续绘出新月、弯月、半月、凸月和满月；不是用另一颗圆球挡住月亮，也不是用地球影子模拟月相。

所有图形只创建一次。每帧按绝对时间修改变换和透明度，拖动进度或重新导出不会累积误差。字号保持原生大小，只有月面的几何椭圆发生拉伸，文字不参与缩放。

## 预览与导出

在项目根目录执行：

```powershell
go run . preview examples/moon-phases/animation.js

go run . mp4 -fps 30 examples/moon-phases/animation.js examples/moon-phases/moon-phases.mp4

go run . gif -fps 12 -scale 0.75 examples/moon-phases/animation.js examples/moon-phases/preview.gif

go run . render -time 20.5 examples/moon-phases/animation.js examples/moon-phases/poster.png
```

视频导出使用项目已有的 FFmpeg 接口，重新导出 MP4 时需要 `ffmpeg` 在当前终端的 `PATH` 中。GIF、PNG 和实时预览不需要 FFmpeg。成品 MP4 可直接播放，无需安装这些开发依赖。

## 测试

```powershell
go test ./examples/moon-phases -v
go test ./...
```

测试对八种代表月相及回到新月的实际渲染亮面比例进行数值比较；核对月相名称与短暂停留位置；检查各镜头正常更新，以及跳转到片尾后再返回同一时刻时画面像素哈希一致。

成品已完整解码验证：MP4 为 1321 帧、44.03 秒；GIF 为 529 帧、44.08 秒。导出器包含末尾一帧，所以媒体时长略长于场景设定的 44 秒。另已抽取编码后的视频帧，检查文字、月相和构图。

![片头封面](cover.png)
