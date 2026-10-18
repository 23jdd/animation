# PowerPoint 原生动画示例

同一份 JS 可实时预览，也可导出为可编辑的 PowerPoint 对象和动画。

```powershell
go run . preview examples/pptx-native/animation.js
go run . pptx -native examples/pptx-native/animation.js examples/pptx-native/native-animation.pptx
```

动画长 7.7 秒：等待 0.5 秒后淡入，并行移动、缩放、旋转，中途停顿，再移动和旋转，最后淡出。请在 PowerPoint 中放映，编辑视图不会播放效果。文字及图形保持可编辑，字体按名称引用而非嵌入。

导出器只使用 Go，不需要 Office 或商业密钥。完整支持范围、采样精度与限制见项目根目录 `PPTX_EXPORT.md`。
