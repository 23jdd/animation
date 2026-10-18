# 图片对象动画

演示本地图片的三种适配方式，以及淡入、移动、旋转、缩放、翻转和横向揭示。示例素材为本项目玫瑰曲线示例生成的 PNG，已放在 `assets/rose.png`。

在项目根目录运行：

```powershell
go run . preview examples/image/animation.js
go run . render -time 4 -scale 2 examples/image/animation.js examples/image/poster.png
go run . gif -fps 12 -scale 0.75 examples/image/animation.js examples/image/preview.gif
go run . mp4 -fps 30 examples/image/animation.js examples/image/image-demo.mp4
```

MP4 导出需要 PATH 中存在 FFmpeg。

最小示例：

```javascript
const photo = image("./assets/rose.png", {width: 360, opacity: 0});
add(photo);
play(fadeIn(photo), {duration: 1});
play(moveTo(photo, [200, 0]), {duration: 2, ease: "smooth"});
```

路径基于当前场景 JS，而不是运行命令时的目录。仅设置 `width` 会自动按比例计算高度。设置宽高后，`fit` 可以选择 `contain`（完整显示）、`cover`（居中裁剪）或 `stretch`（拉伸）。

支持透明 PNG、JPEG、静态 WebP 和 GIF 首帧，不支持图片内部的动画、动画 WebP、网络地址或 SVG。图片尺寸及源文件信息为只读属性；用 `scaleTo()` 改变显示大小。素材在脚本加载时解码并缓存，热重载后重新读取。

已导出 16 秒演示视频与 GIF，完整解码检查无错误，并检查过 2 倍分辨率定格。自动化测试覆盖透明度、裁剪、镜像、旋转、相机、导出缩放、文件格式和错误处理；本例额外验证关键帧前后跳转以及 `create` / `uncreate` 的揭示进度。
