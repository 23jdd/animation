# 把温柔，写成一朵玫瑰

一支 64 秒的浪漫小短片：拆开一封信、一笔描花、层层绽放、花环盛开，最后汇成一束七朵玫瑰。奶油粉底色、鼠尾草绿枝叶、轻轻漂浮的花瓣，配原创程序合成轻音乐，无配音。

所有图形与动作由 JavaScript 绘制，没有使用位图花朵或外部素材。音乐同样由 JavaScript 生成，不依赖音频采样包。

## 成品

- `rose-curve.mp4`：1280 × 720、30 fps，包含轻音乐。
- `preview.gif`：960 × 540、12 fps，无声预览。
- `cover.png`、`poster.png`：开场与花束定格。
- `music.m4a`：单独的配乐。
- `animation.js`：场景、曲线、绽放和花束动画。
- `music.js`：立体声合成配乐源码。
- `subtitles.srt`：画面中主要文案的字幕副本；视频已绘制文字，无须再叠加。
- `animation_test.go`：数学坐标、多种花瓣数的描线终点、七幕场景、花束分步组装、螺旋状态及反向跳转测试。
- `rose-storyboard.pptx`：七幕静态分镜，选取第 4、11、26、33、42、50、60 秒，以 2 倍分辨率导出；不包含配乐或原生 PPT 动画。

## 丰富版新增

- 信封打开、信纸升起，信上的玫瑰线逐笔描绘。
- 三瓣、五瓣、七瓣曲线依次成形，对应「初见、心动、相伴」。
- 30 片花瓣沿螺旋汇聚，五层花瓣错时缩放、旋转、展开。
- 12 朵小花逐朵组成花环，蝴蝶沿弧线飞行，双翼轻轻扇动。
- 花束由三朵扩为七朵，增加三枚花蕾、满天星枝、折叠包装纸、丝带和「予你」吊牌。
- 花瓣内侧增加细线褶皱与高光；背景有分层漂浮花瓣和稀疏微光。
- 配乐扩成八个慢速小节，花束收拢后温柔落回主和弦。

## 时间线

| 时间 | 内容 |
| --- | --- |
| 0–7 秒 | 把温柔，写成一朵玫瑰，花朵在枝叶间展开 |
| 7–14 秒 | 一封信打开，玫瑰线在信纸上出现 |
| 14–27 秒 | 光点逐瓣描出一条完整的五瓣玫瑰线 |
| 27–36 秒 | 三瓣的初见，五瓣的心动，七瓣的相伴 |
| 36–45 秒 | 飘落的花瓣沿螺旋靠近，层叠花瓣绽放 |
| 45–54 秒 | 十二朵花围成花环：你在，便是花期 |
| 54–64 秒 | 七朵花、花蕾与满天星聚成礼物花束 |

## 曲线怎么画

五瓣玫瑰线使用极坐标方程：

```javascript
const r = a * Math.cos(5 * theta);
const x = r * Math.cos(theta);
const y = r * Math.sin(theta);
```

绘制到屏幕时使用 `[x, -y]`，因为屏幕的纵轴向下。`a` 是半径，也就是图形的大小。

本片从 `theta = -Math.PI / 10` 画到 `9 * Math.PI / 10`，从中心出发，依次描过五瓣，再回到中心。半径为负时仍按公式换算坐标，不能将它截断成零。

对于正整数 `n`，`r = a cos(nθ)` 在 `n` 为奇数时有 `n` 瓣，完整描绘的参数跨度为 `π`；偶数时有 `2n` 瓣，跨度为 `2π`。

14–27 秒展示的是未经变形的数学曲线。27–36 秒依次绘制 `n = 3、5、7` 的完整曲线；不会用非整数参数混变来冒充完整的整数瓣玫瑰线。

其余镜头中的丰满花朵，是对单瓣采样后横向拓宽、旋转复制、分层组合的装饰造型，**不是一条原始极坐标曲线本身**。花瓣内部的细线模拟褶皱，满天星与叶脉则使用独立的矢量路径和小圆点绘制。

## 运行

在项目根目录：

```powershell
go run . preview examples/rose-curve/animation.js
go run . gif -fps 12 -scale 0.75 examples/rose-curve/animation.js examples/rose-curve/preview.gif
go test ./examples/rose-curve -v
go run . pptx -times 4,11,26,33,42,50,60 examples/rose-curve/animation.js examples/rose-curve/rose-storyboard.pptx
```

引擎预览和 GIF 不播放配乐。重新导出带音乐的视频，需要 Go、Node.js 和 PATH 中的 FFmpeg：

```powershell
$roseOutput = Join-Path ([IO.Path]::GetTempPath()) ('rose-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $roseOutput | Out-Null
go run . mp4 -fps 30 examples/rose-curve/animation.js "$roseOutput/silent.mp4"
node examples/rose-curve/music.js "$roseOutput/score.wav"
ffmpeg -y -i "$roseOutput/silent.mp4" -i "$roseOutput/score.wav" -map 0:v:0 -map 1:a:0 -c:v copy -c:a aac -b:a 128k -shortest -movflags +faststart examples/rose-curve/rose-curve.mp4
```

配乐为 48 kHz 立体声 PCM 合成，慢速和弦与稀疏的柔和铃音；淡入、淡出并限制峰值，避免削波。最终 MP4 使用 AAC 音轨。临时目录保留中间文件，需要时可手动清理。

## 调整心意

- 改文案：修改 `animation.js` 中的 `label(...)` 和 `CUES`，同步修改字幕副本。
- 改配色：修改顶部的 `P` 与 `flower()` 的花瓣颜色。
- 改花束大小：修改 `bouquetFlowers` 的七个半径和 `bouquetAnchors`。
- 改花环：修改 `wreathItems` 的数量、椭圆半径和花朵大小。
- 改螺旋：修改 `spiralState()` 中的半径、转数、持续时间与透明度包络。
- 改数学曲线：主描线镜头按五瓣设计；更改 `FILM.n` 时，需要同步修改主 `makeTracer()`、参考线、文案和单瓣装饰造型。`galleryItems` 则分别管理三种整数参数曲线。
- 改音乐：修改 `music.js` 的和弦、音符时间和音量后重新合成、合入视频。

动作全部按绝对时间计算，支持播放器暂停、前后拖动和反复跳转，不依赖上一帧状态。

## 验证

已通过本示例测试、项目全量 `go test ./...` 和本示例的 `go vet`。测试覆盖 `n = 3、4、5、7` 的描线终点与闭合、正向及反向跳转的关键帧一致性，以及七幕显隐、螺旋范围和花束分步组装。

已检查各幕定格及编码后的花束画面，并完整解码检查 MP4 的视频、音频与 GIF：MP4 时长 64.00 秒，GIF 时长约 64.08 秒；成片音轨平均音量约 −29.1 dBFS、峰值约 −15.7 dBFS，无削波。
