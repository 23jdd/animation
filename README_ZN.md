# Animation — 实时动画引擎

一个使用 **Go 语言开发、由 JavaScript 驱动的轻量级实时动画引擎**。

引擎的分层设计遵循 [`Architecture.md`](Architecture.md) 中的架构：

- **JavaScript**：描述场景及动画行为。
- **Go 时间轴（Timeline）**：以确定性的方式更新动画对象（Mobject）的状态。
- **gg 渲染器**：将动画对象的状态光栅化为图像。
- **Ebitengine**：负责实时显示渲染结果。

[English](README.md)
## 预览
<img src="animation.gif" align="center"/>

## 运行方式

项目附带的 `index.js` 是一个完整的使用示例。

```powershell
# 实时预览动画
go run . preview index.js

# 将指定时间点的画面导出为 PNG
go run . render -width 1280 -height 720 -time 2 index.js frame.png

# 导出 GIF 动画
go run . gif -width 1280 -height 720 -fps 30 index.js animation.gif

# 导出 MP4 视频
go run . mp4 -width 1280 -height 720 -fps 60 index.js animation.mp4
```

导出 MP4 视频需要安装 `ffmpeg`，并确保它位于系统的 `PATH` 环境变量中。

PNG 和 GIF 导出不依赖额外的外部运行时工具。

也可以将程序安装为命令行工具：

```powershell
go install ./cmd/anim
```

安装完成后，可以直接使用 `anim` 命令代替 `go run .`。

### 实时预览播放器

预览窗口在动画画布下方集成了一个基于 **EbitenUI** 实现的自适应动画播放器。

调整窗口大小时，播放器控件会自动重新布局，动画画布则始终保持居中，并在不改变宽高比的情况下进行缩放。

播放器提供以下功能：

- **时间轴控制**：支持拖动时间轴，快速跳转到指定动画时间。
- **播放与暂停**：随时控制动画播放状态。
- **重新播放**：将动画重置到起始位置。
- **逐帧定位**：支持跳转到指定帧。
- **播放速度调整**：支持不同倍率的播放速度。
- **时间显示**：显示当前播放时间及动画总时长。
- **播放状态显示**：实时展示播放器状态。
- **脚本错误提示**：在界面中显示 JavaScript 脚本错误。
- **热重载（Hot Reload）**：监控脚本及相关资源文件的变化，并自动重新加载。

热重载支持监控当前脚本以及附近的 `.js`、`.json`、`.svg`、字体和图像资源文件。

热重载采用**事务式更新机制**：

当修改后的脚本无效时，播放器不会直接替换当前场景，而是继续显示上一次成功加载的场景，并展示相应的脚本错误。

当重新加载成功时，播放器会保留当前时间轴位置，避免每次修改代码后都需要从头播放。

### 预览快捷键

| 快捷键 | 功能 |
|---|---|
| `Space` | 播放 / 暂停 |
| `R` | 重新播放 |
| `←` / `→` | 向前 / 向后跳转 0.25 秒 |
| `Home` | 跳转到动画起点 |
| `1` | 0.5 倍速播放 |
| `2` | 1 倍速播放 |
| `3` | 2 倍速播放 |

## JavaScript API

完整接口文档请参阅 [`API_REFERENCE.md`](API_REFERENCE.md)。

### 创建动画对象

使用以下函数创建动画对象（Mobject）：

| 函数 | 功能 |
|---|---|
| `circle()` | 创建圆形 |
| `rect()` | 创建矩形 |
| `line()` | 创建线段 |
| `polygon()` | 创建多边形 |
| `path()` | 创建路径 |
| `text()` | 创建文本 |
| `group()` | 创建对象组 |

创建对象后，通过 `add(...)` 将其添加到场景中。

例如：

```js
const dot = circle({
  radius: 40,
  position: [-300, 0],
  fill: "#3b82f6",
  stroke: "#bfdbfe",
  strokeWidth: 3,
});

add(dot);

// 播放创建动画
play(create(dot), { duration: 0.5 });

// 将圆形移动到指定位置
play(moveTo(dot, [300, 0]), {
  duration: 2,
  ease: "smooth",
});
```

上述代码首先创建一个圆形对象，然后播放对象的创建动画，最后使用平滑缓动效果将圆形移动到指定位置。

### 动画操作

引擎支持以下动画操作：

| API | 功能 |
|---|---|
| `moveTo` | 移动到指定位置 |
| `moveBy` | 按相对位移移动 |
| `scaleTo` | 缩放到指定比例 |
| `scaleBy` | 按相对比例缩放 |
| `rotateTo` | 旋转到指定角度 |
| `rotateBy` | 按相对角度旋转 |
| `fadeIn` | 淡入 |
| `fadeOut` | 淡出 |
| `opacityTo` | 将透明度变化到指定值 |
| `colorTo` | 将颜色过渡到指定颜色 |
| `create` | 播放对象创建动画 |
| `uncreate` | 播放对象消除动画 |
| `write` | 逐字显示文本 |
| `unwrite` | 逐字擦除文本 |
| `drawLine` | 从起点画线到终点，可选笔尖跟随 |
| `drawLineFromCenter` | 线段从中点向两端展开 |
| `transform` | 从一个对象变换为另一个对象 |
| `replacementTransform` | 使用变换动画替换对象 |
| `moveAlongPath` | 沿指定路径移动 |
| `growFromCenter` | 从中心放大进入 |
| `shrinkToCenter` | 收缩到中心退出 |
| `growFromPoint` | 从指定点放大进入 |
| `growFromEdge` | 从边缘方向滑入 |
| `spinIn` | 旋转一圈进入 |
| `spinOut` | 旋转一圈退出 |
| `flash` | 颜色短暂高亮 |
| `indicate` | 放大并着色强调 |
| `blink` | 透明度闪烁 |
| `wiggle` | 旋转摆动 |
| `wave` | 子对象波浪起伏 |
| `zoomIn` | 从放大状态缩小归位 |
| `zoomOut` | 放大退出 |
| `focusOn` | 短暂放大强调 |

### 动画组合

通过以下函数，可以将多个动画组合成复杂的动画序列：

| API | 功能 |
|---|---|
| `sequence` | 顺序执行多个动画 |
| `parallel` | 并行执行多个动画 |
| `stagger` | 按一定时间间隔错开执行动画 |
| `repeat` | 重复执行动画 |
| `loop` | 循环播放动画 |
| `yoyo` | 正向与反向交替播放动画 |

### 自定义字体

`text()` 支持加载自定义 TrueType 和 OpenType 字体。

字体文件路径相对于当前场景脚本所在目录进行解析。

```js
text("实时动画", {
  font: "./assets/fonts/NotoSansSC-Regular.ttf",
  size: 48,
});
```

如果未指定 `font`，场景文本和 EbitenUI 播放器将使用内嵌字体组合：

- **JetBrains Mono Regular**：优先用于拉丁文字。
- **Source Han Sans SC Regular**：自动用于补充缺失的中文及其他 CJK 字形。

因此，应用程序不依赖操作系统中安装的字体，也能够正确显示中文等文字。

两款字体对应的 SIL Open Font License 1.1 许可证文本均包含在 [`fontasset/assets`](fontasset/assets) 目录中。

### 缓动函数（Easing）

引擎提供多种缓动函数，用于控制动画随时间变化的速度。

| 缓动名称 | 效果 |
|---|---|
| `linear` | 线性变化 |
| `easeIn` | 缓入 |
| `easeOut` | 缓出 |
| `easeInOut` | 缓入缓出 |
| `cubic` | 三次曲线缓动 |
| `sine` | 正弦曲线缓动 |
| `expo` | 指数曲线缓动 |
| `circ` | 圆形曲线缓动 |
| `back` | 回退缓动 |
| `bounce` | 弹跳缓动 |
| `elastic` | 弹性缓动 |
| `smooth` | 平滑缓动 |
| `smoother` | 更平滑的缓动 |

### 路径绘制

路径使用与具体渲染器无关的命令表示，因此路径定义不依赖 `gg` 的内部实现。

例如：

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

上述代码创建了一条贝塞尔曲线路径。

其中：

- `M`：移动路径起点。
- `C`：绘制三次贝塞尔曲线。
- `stroke`：设置描边颜色。
- `fill`：设置填充颜色，这里使用完全透明的颜色。

### 动画值与动态更新

可以通过 `value(initial)` 创建可动画化的数值，并使用 `animateValue` 控制其变化。

也可以为动画对象注册 `updater`（更新函数），实现程序化运动。

例如：

```js
const angle = value(0);

dot.update((self) => {
  self.x = Math.cos(angle.value) * 200;
  self.y = Math.sin(angle.value) * 200;
});

play(animateValue(angle, Math.PI * 2), {
  duration: 3,
});
```

这个示例通过不断改变 `angle` 的值，使圆形对象沿半径为 200 的圆周运动。

与直接修改坐标的动画不同，更新函数可以根据当前动画状态动态计算对象的位置、大小或其他属性。

这适合实现数学可视化、参数化动画和程序化运动。

### 确定性时间轴

所有离线导出帧都通过定位到与实时预览相同的时间轴来生成。

因此，**离线渲染结果不会受到实时预览帧率的影响**。

即使实时预览出现掉帧，导出的动画仍然可以按照指定时间精确计算每一帧的状态。

## 使用 JSDoc 实现编辑器自动补全

项目提供 [`animation-api.js`](animation-api.js)，用于描述 JavaScript 动画 API 的类型信息。

该文件包含：

- 所有 JavaScript 全局函数。
- API 参数及配置对象。
- 路径命令。
- 缓动函数名称。
- 动画对象更新函数（Updater）。
- 动画句柄（Animation Handle）。

同时，项目中的 [`jsconfig.json`](jsconfig.json) 用于启用 JavaScript 类型检查及编辑器自动补全功能。

支持 JSDoc 的编辑器可以根据这些声明提供类型提示、参数补全和错误检查。

### 启用自动补全

在场景脚本的第一行添加：

```js
/// <reference path="./animation-api.js" />
```

项目附带的 [`index.js`](index.js) 已经包含该引用。

这些声明**仅用于编辑器的静态分析与自动补全**，不会由动画引擎执行，也不会影响 JavaScript 脚本的实际运行。