# Animation JavaScript API Reference

本文档描述通过 Goja 提供给场景脚本的全局 JavaScript API。编辑器类型定义位于 [`animation-api.js`](animation-api.js)，使用方式如下：

```js
/// <reference path="./animation-api.js" />
```

## 1. 基本概念

场景脚本只负责声明对象和动画，不需要编写渲染循环：

```text
JavaScript Scene
      ↓
Scene + Timeline
      ↓
Go Animation Engine
      ↓
Renderer / Preview / Export
```

坐标系原点位于画布中心：

- X 轴向右为正。
- Y 轴向下为正。
- 角度使用弧度。
- 时间和动画持续时间使用秒。
- 颜色使用 `#RGB`、`#RRGGBB` 或 `#RRGGBBAA`。

## 2. 公共类型

### `Vec2`

二维坐标、偏移或缩放值：

```js
/** @type {Vec2} */
const point = [100, -50];
```

### `MobjectOptions`

所有图形共享的配置：

| 属性 | 类型 | 默认值 | 说明 |
|---|---|---:|---|
| `position` | `Vec2` | `[0, 0]` | 相对画布中心的位置 |
| `rotation` | `number` | `0` | 旋转弧度 |
| `scale` | `number \| Vec2` | `1` | 整体或分轴缩放 |
| `opacity` | `number` | `1` | 透明度，范围为 0～1 |
| `fill` | `Color` | `#ffffff` | 填充颜色 |
| `stroke` | `Color` | `#ffffff` | 描边颜色 |
| `strokeWidth` | `number` | `2` | 描边宽度 |

创建后的对象向 JavaScript 暴露以下动态属性：

| 属性 | 类型 | 说明 |
|---|---|---|
| `x` | `number` | 水平位置，可读写 |
| `y` | `number` | 垂直位置，可读写 |
| `rotation` | `number` | 旋转弧度，可读写 |
| `scaleX` | `number` | 水平缩放，可读写，负数表示水平翻转 |
| `scaleY` | `number` | 垂直缩放，可读写，负数表示垂直翻转 |
| `opacity` | `number` | 透明度，可读写 |
| `update(callback)` | `function` | 添加逐帧 Updater |

## 3. 创建图形

### `circle(options?)`

创建圆形。

```js
const ball = circle({
  radius: 50,
  position: [-200, 0],
  fill: "#3b82f6",
  stroke: "#bfdbfe",
  strokeWidth: 4,
});
```

额外参数：

| 属性 | 类型 | 默认值 |
|---|---|---:|
| `radius` | `number` | `50` |

### `rect(options?)` / `rectangle(options?)`

创建矩形或圆角矩形。两个函数等价。

```js
const card = rect({
  width: 240,
  height: 140,
  radius: 16,
  fill: "#f97316",
});
```

| 属性 | 类型 | 默认值 | 说明 |
|---|---|---:|---|
| `width` | `number` | `100` | 宽度 |
| `height` | `number` | `100` | 高度 |
| `radius` | `number` | `0` | 圆角半径 |

### `line(options?)`

创建线段。

```js
const axis = line({
  start: [-300, 0],
  end: [300, 0],
  stroke: "#94a3b8",
  strokeWidth: 3,
});
```

| 属性 | 类型 | 默认值 |
|---|---|---:|
| `start` | `Vec2` | `[-50, 0]` |
| `end` | `Vec2` | `[50, 0]` |

### `polygon(options)`

创建多边形。

```js
const triangle = polygon({
  points: [[0, -80], [-70, 50], [70, 50]],
  fill: "#22c55e",
});
```

`points` 是必需的 `Vec2[]`。

### `path(options)`

创建与渲染器无关的矢量路径。

```js
const curve = path({
  commands: [
    ["M", 0, 0],
    ["C", 100, -180, 220, 180, 320, 0],
  ],
  fill: "#00000000",
  stroke: "#38bdf8",
  strokeWidth: 4,
});
```

支持的路径命令：

| 命令 | 格式 | 说明 |
|---|---|---|
| Move | `["M", x, y]` | 移动画笔 |
| Line | `["L", x, y]` | 直线 |
| Quadratic | `["Q", cx, cy, x, y]` | 二次贝塞尔曲线 |
| Cubic | `["C", c1x, c1y, c2x, c2y, x, y]` | 三次贝塞尔曲线 |
| Close | `["Z"]` | 闭合路径 |

### `text(content, options?)`

创建文本。

```js
const title = text("Hello Animation", {
  position: [0, -220],
  size: 42,
  font: "./assets/fonts/NotoSansSC-Regular.ttf",
  fill: "#f8fafc",
});
```

| 属性 | 类型 | 默认值 | 说明 |
|---|---|---:|---|
| `size` | `number` | `24` | 字号 |
| `font` | `string` | JetBrains Mono + 思源黑体回退 | `.ttf` 或 `.otf` 字体路径 |

相对字体路径以当前场景脚本所在目录为基准，而不是运行命令时的工作目录。字体会按“文件版本 + 字号”缓存；修改字体文件后，Hot Reload 会重新构建场景并使用新字体。如果字体文件不存在、不是普通文件、内容损坏或扩展名不受支持，场景加载会显示脚本错误并继续保留上一次有效场景。

默认字体链优先使用 JetBrains Mono；当字符缺失时逐字回退到思源黑体 SC，因此中英文可以在同一段文字中自然混排。也可以指定其他字体，例如：

```js
const chineseTitle = text("实时动画引擎", {
  font: "./assets/fonts/NotoSansSC-Regular.ttf",
  size: 48,
  fill: "#f8fafc",
});
```

### `image(source, options?)`

将本地图片加载为可以参与动画的对象，默认以图片中心定位。支持 PNG（含透明通道）、JPEG、静态 WebP 和 GIF 首帧；不播放 GIF 内部动画，暂不支持动画 WebP、SVG、网址或 data URL。

```js
const photo = image("./assets/photo.png", {
  width: 360,
  height: 240,
  fit: "cover",
  position: [-200, 0],
  opacity: 0,
});
add(photo);
play(fadeIn(photo), { duration: 0.8 });
play(parallel(
  moveTo(photo, [200, 0]),
  scaleTo(photo, 1.2),
  rotateTo(photo, Math.PI / 12),
), { duration: 2, ease: "smooth" });
```

| 参数/属性 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `source` | `string` | 必填 | 本地文件路径，相对路径基于场景 JS 所在目录 |
| `width` | `number` | 原始像素宽度 | 逻辑显示框宽度，必须为正的有限数 |
| `height` | `number` | 原始像素高度 | 逻辑显示框高度，必须为正的有限数 |
| `fit` | `"contain" \| "cover" \| "stretch"` | `"contain"` | 图片在显示框内的适配方式 |

未指定宽高时使用原始像素尺寸；只指定其中一项时，另一项按原始比例计算。两项都指定时：

- `contain`：保持比例、完整显示，空余部分透明，不自动填色。
- `cover`：保持比例铺满显示框，居中裁掉超出的部分。
- `stretch`：分别缩放宽高，可能改变比例。

图片支持公共变换、透明度、层级组及 Updater。`fill`、`stroke`、`strokeWidth` 不会给图片染色或添加边框；如需边框，在图片后面放置一个 `rect()`。`create()` / `uncreate()` 沿显示框从左向右揭示 / 从右向左隐藏图片，不是矢量描边。负的 `scaleX` / `scaleY` 可以翻转图片。

返回对象还有以下只读属性；调整显示尺寸时使用 `scaleTo()` / `scaleBy()`：

| 属性 | 说明 |
|---|---|
| `width`, `height` | 逻辑显示框宽高，不包含缩放变换 |
| `naturalWidth`, `naturalHeight` | 图片文件原始像素宽高 |
| `source` | 解析后的本地文件路径 |
| `fit` | 创建时的适配方式 |

文件在脚本加载时解码，同一个运行时复用已解码数据，不会在每帧读盘；新运行时和热重载会重新读取。放在场景目录及子目录中的 `.png`、`.jpg`、`.jpeg`、`.webp`、`.gif` 修改后会触发热重载。目录外的图片可以加载，但修改后需手动重载场景。

文件不存在、格式损坏、无效宽高或不支持的 `fit` 会抛出可捕获的 JavaScript `TypeError`，不会使用 Go `panic`。每张源图片限制为 64 MiB 编码文件、6400 万像素，解码前检查尺寸。

完整演示见 `examples/image/animation.js`。

### `group(...children)`

将多个对象组成一个层级组。移动、缩放、旋转或改变组的透明度时会影响全部子对象。

```js
const pair = group(circle(), rect());
add(pair);
```

## 4. 场景 API

### `add(...objects)`

将对象添加到场景：

```js
add(title, ball, card);
```

### `remove(...objects)`

从场景移除对象：

```js
remove(ball);
```

## 5. 时间轴 API

### `play(...animations, options?)`

把动画添加到当前时间轴位置。传入多个动画时，它们会并行执行。

```js
play(
  moveTo(ball, [300, 0]),
  rotateBy(card, Math.PI),
  {
    duration: 2,
    ease: "smooth",
  },
);
```

配置对象必须位于最后：

| 属性 | 类型 | 说明 |
|---|---|---|
| `duration` | `number` | 覆盖动画时长 |
| `ease` | `EasingName` | 设置缓动；可以单独提供 |

```js
play(moveTo(ball, [0, 0]), { ease: "bounce" });
```

每次 `play` 完成后，时间轴游标会自动移动到该动画末尾。

### `wait(seconds?)`

在时间轴中插入等待时间，默认等待 1 秒：

```js
wait(0.5);
```

## 6. 基础动画

所有动画构造函数都返回 `AnimationHandle`，必须交给 `play` 或动画组合函数。

### 移动

```js
moveTo(target, [x, y]);
moveBy(target, [dx, dy]);
```

### 缩放

```js
scaleTo(target, 2);
scaleTo(target, [2, 0.5]);
scaleBy(target, 1.5);
```

### 旋转

```js
rotateTo(target, Math.PI);
rotateBy(target, Math.PI / 2);
```

### 透明度与颜色

```js
fadeIn(target);
fadeOut(target);
opacityTo(target, 0.4);
fadeTo(target, 0.6);   // 与 opacityTo 等价，名字更直观
colorTo(target, "#ef4444");
```

### 创建与消除

```js
create(target);
uncreate(target);
```

`create` 和 `uncreate` 控制对象的绘制进度以及可见状态。

文字使用 `write` 和 `unwrite`，按 Unicode 字符逐个显示或擦除。中文会按一个汉字一个步骤处理，每个字符出现时会淡入，并且末尾有一个闪烁的打字机光标跟随：

```js
const title = text("你好，动画！", { size: 42 });
add(title);
play(write(title), { duration: 1.5, ease: "smooth" });
wait(0.5);
play(unwrite(title), { duration: 0.8 });
```

光标颜色默认与文字填充色一致，也可以用第二个参数指定，或传 `null` 关闭。第三个参数直接设置对象的透明度（等同 `title.opacity = 0.5`）：

```js
play(write(title, "#f97316")); // 橙色光标
play(write(title, null));      // 无光标
play(write(title, null, 0.5)); // 无光标，同时把文字透明度设为 0.5
```

### `drawLine(line, tip?)`

线段从起点逐步绘制到终点。可选传入一个 `tip` 对象（通常是圆点），它会跟随画笔尖端移动，绘制完成后自动隐藏：

```js
const axis = line({
  start: [-300, 0],
  end: [300, 0],
  stroke: "#3b82f6",
  strokeWidth: 3,
});
const nib = circle({ radius: 8, fill: "#f97316", stroke: "#00000000" });

add(axis, nib);
play(drawLine(axis, nib), { duration: 1.5, ease: "smooth" });
```

不带 `tip` 时，`drawLine(line)` 与 `create(line)` 等价，但语义更明确。

### `drawLineFromCenter(line)`

线段从中点同时向两端展开：

```js
const axis = line({ start: [-300, 0], end: [300, 0], stroke: "#22c55e", strokeWidth: 3 });
add(axis);
play(drawLineFromCenter(axis), { duration: 1.2, ease: "smooth" });
```

### 矩形逐步描边

`create(rect)` 现在会让普通矩形（`radius: 0`）沿周长逐步描边，从左上角顺时针画满一周：

```js
const card = rect({ width: 240, height: 140, fill: "#f97316" });
add(card);
play(create(card), { duration: 1.5 });
```

圆角矩形（`radius > 0`）仍为整体显示。

### `transform(source, destination)`

插值两个对象的公共状态，包括位置、缩放、旋转、透明度和填充颜色。当前版本不执行不同路径之间的几何 Morph。

```js
play(transform(circleA, circleB));
```

### `replacementTransform(source, destination)`

执行 Transform，结束后隐藏源对象并显示目标对象。

### `moveAlongPath(target, route)`

沿路径移动对象：

```js
const route = path({
  commands: [["M", 0, 0], ["C", 100, -100, 200, 100, 300, 0]],
  fill: "#00000000",
});

play(moveAlongPath(ball, route), { duration: 2 });
```

路径坐标作为对象初始位置的相对偏移使用。通常应以 `["M", 0, 0]` 开始，避免动画开始时发生位置跳变。

### 生长与收缩

```js
growFromCenter(target);      // 从中心放大并淡入
shrinkToCenter(target);      // 收缩到中心并淡出
growFromPoint(target, [0, 0]); // 从指定点放大并移回原位
growFromEdge(target, "left");  // 从边缘方向滑入，可传 distance 调整起始偏移
```

`growFromEdge` 的 `edge` 取值 `"left"`、`"right"`、`"up"`、`"down"`，默认 `"up"`；`distance` 默认 `150` 像素。

### 旋转进入与退出

```js
spinIn(target);   // 旋转一圈进入，同时放大并淡入
spinOut(target);  // 旋转一圈退出，同时缩小并淡出
```

### 强调与闪烁

```js
flash(target);               // 颜色短暂变亮再恢复
indicate(target, 1.3);       // 短暂放大并着色强调
blink(target, 2);            // 透明度闪烁 2 次
```

- `flash(target, color?)`：向 `color`（默认 `#ffffff`）高亮再恢复。
- `indicate(target, scale?, color?)`：峰值缩放 `scale`（默认 `1.3`），同时向 `color` 着色。
- `blink(target, times?)`：`times` 为闪烁次数，默认 `1`。

### 摆动与波浪

```js
wiggle(target, 0.3, 2); // 旋转来回摆动
wave(group, 20);        // 子对象依次上下起伏，形成波浪
```

- `wiggle(target, amplitude?, cycles?)`：`amplitude` 为摆动幅度（弧度，默认 `0.3`），`cycles` 为周期数（默认 `2`）。
- `wave(target, amplitude?)`：对 group 的直接子对象施加相位偏移的上下移动；`amplitude` 默认 `20` 像素。

### 聚焦与缩放

```js
zoomIn(target, 2.5);   // 从放大状态缩小归位并淡入
zoomOut(target, 2.5);  // 放大退出并淡出
focusOn(target, 1.25); // 短暂放大再恢复，用于强调
```

## 7. 动画组合

### `sequence(...animations)`

按顺序执行动画：

```js
play(sequence(
  moveTo(ball, [0, 0]),
  scaleTo(ball, 2),
  fadeOut(ball),
));
```

### `parallel(...animations)`

同时执行动画：

```js
play(parallel(
  moveTo(ball, [300, 0]),
  rotateBy(ball, Math.PI),
));
```

### `stagger(delay, ...animations)`

让每个动画比前一个晚指定秒数开始：

```js
play(stagger(0.15,
  fadeIn(item1),
  fadeIn(item2),
  fadeIn(item3),
));
```

### `repeat(animation, count)`

重复执行指定次数。

```js
play(repeat(rotateBy(icon, Math.PI), 4));
```

### `loop(animation, count?)`

当前实现是有限次数的重复，默认执行 2 次，并非无限循环。

```js
play(loop(fadeIn(item), 3));
```

### `yoyo(animation, count)`

交替正向、反向执行动画。

```js
play(yoyo(moveBy(ball, [200, 0]), 4));
```

## 8. 缓动函数

可在 `play` 的 `ease` 属性中使用：

| 名称 | 说明 |
|---|---|
| `linear` | 线性 |
| `easeIn`, `easeInQuad` | 二次缓入 |
| `easeOut`, `easeOutQuad` | 二次缓出 |
| `easeInOut`, `easeInOutQuad` | 二次缓入缓出 |
| `cubic`, `easeInOutCubic` | 三次缓入缓出 |
| `sine` | 正弦缓动 |
| `expo` | 指数缓出 |
| `circ` | 圆形缓出 |
| `back` | 回退缓出 |
| `bounce` | 弹跳缓出 |
| `elastic` | 弹性缓出 |
| `smooth` | Smoothstep |
| `smoother` | Smootherstep |

未知名称会回退到 `linear`。

## 9. Updater

Updater 在场景更新时执行，适合对象跟随、数学函数和程序化运动：

```js
const dot = circle({ radius: 12 });

dot.update((self, time, dt) => {
  self.x = Math.cos(time * 2) * 200;
  self.y = Math.sin(time * 2) * 100;
});

add(dot);
```

参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `self` | `Mobject` | 当前对象 |
| `time` | `number` | 当前时间轴时间 |
| `dt` | `number` | 距离上一次更新的时间 |

为了保证 Seek 和离线导出的结果一致，优先根据绝对 `time` 或 ValueTracker 设置状态，不要仅依赖逐帧累加。

## 10. ValueTracker

### `value(initial?)`

创建数值追踪器：

```js
const angle = value(0);
```

通过 `tracker.value` 读写当前数值。

### `animateValue(tracker, destination)`

创建数值动画：

```js
const angle = value(0);
const dot = circle({ radius: 12 });

dot.update((self) => {
  self.x = Math.cos(angle.value) * 200;
  self.y = Math.sin(angle.value) * 200;
});

add(dot);
play(animateValue(angle, Math.PI * 2), {
  duration: 3,
  ease: "linear",
});
```

## 11. 完整示例

```js
/// <reference path="./animation-api.js" />

const title = text("Animation API", {
  position: [0, -220],
  size: 42,
  fill: "#f8fafc",
});

const ball = circle({
  radius: 48,
  position: [-350, 20],
  fill: "#3b82f6",
  stroke: "#bfdbfe",
  strokeWidth: 4,
});

const target = rect({
  width: 150,
  height: 100,
  position: [300, 20],
  radius: 16,
  fill: "#f97316",
});

add(title, ball, target);

play(create(title), { duration: 0.8, ease: "smooth" });

play(parallel(
  moveTo(ball, [0, 20]),
  scaleTo(ball, 1.4),
  rotateBy(target, Math.PI),
), {
  duration: 2,
  ease: "smooth",
});

wait(0.3);

play(moveTo(ball, [300, 20]), {
  duration: 1,
  ease: "bounce",
});

play(parallel(fadeOut(ball), fadeOut(target)), {
  duration: 0.8,
});
```

运行预览：

```powershell
go run . preview index.js
```

## 12. 编辑器与运行时

- `animation-api.js` 只提供 JSDoc 类型和代码补全，不会被动画引擎执行。
- `jsconfig.json` 开启 JavaScript 类型检查。
- 场景脚本由 Goja 执行。
- 保存场景脚本后，预览窗口会进行事务式热重载。
- 脚本加载失败时继续显示上一次有效场景。
## 13. 3D API

3D API 使用透视相机、三角形光栅化、逐像素深度缓冲、透视正确材质插值、平滑法线、方向光高光、雾化和屏幕空间发光。
原有 2D 对象保持兼容，并可作为 3D 场景的 UI 覆盖层。

```js
const camera = camera3D({
  position: [0, 120, 900],
  target: [0, 0, 0],
  fov: 48,
  lightDirection: [-0.4, 0.8, 1],
  ambient: 0.18,
  diffuse: 0.9,
  fogColor: "#dce7e2",
  fogNear: 700,
  fogFar: 1500,
  fogStrength: 0.35,
});

const planet = sphere3D({
  radius: 150,
  segments: 32,
  rings: 18,
  fill: "#2563eb",
  smoothShading: true,
  specular: 0.35,
  shininess: 64,
  stroke: "#67e8f9",
  glow: 10,
  glowColor: "#22d3ee",
});

add(planet);
play(rotate3DBy(planet, [0.4, Math.PI * 2, 0.2]), { duration: 6 });
play(moveCamera3DTo(camera, [0, 40, 620]), { duration: 2, ease: "smooth" });
```

### 创建对象

- `mesh3D({ vertices, faces, normals?, ...material })`；`normals` 可为每个顶点提供一个自定义法线。
- `box3D({ width, height, depth, ...material })`
- `sphere3D({ radius, segments, rings, ...material })`
- `torus3D({ majorRadius, minorRadius, segments, sides, ...material })`
- `camera3D({ position, target, up, fov, near, lightDirection, ambient, diffuse, fogColor, fogNear, fogFar, fogStrength })`

材质参数：

- `smoothShading`：插值顶点法线，让曲面连续；球体和圆环默认开启。
- `specular`：高光强度，范围 `0..2`，默认 `0.12`。
- `shininess`：高光锐度，范围 `1..256`，默认 `32`。
- `emissive`：不受光照影响的基础亮度，范围 `0..1`。
- `faceColors`：按面的颜色数组。

相机会在 `fogNear` 到 `fogFar` 之间渐进混合 `fogColor`；`fogStrength` 为 `0` 时关闭雾化。

### 3D 动画

- `move3DTo(mesh, [x, y, z])`
- `move3DBy(mesh, [dx, dy, dz])`
- `rotate3DTo(mesh, [rx, ry, rz])`
- `rotate3DBy(mesh, [drx, dry, drz])`
- `scale3DTo(mesh, scale)`
- `scale3DBy(mesh, factor)`
- `moveCamera3DTo(camera, [x, y, z])`
- `lookAt3D(camera, [x, y, z])`
- `cameraFOVTo(camera, degrees)`

`Mesh3DMobject` 还向 JavaScript Updater 暴露 `x3D`、`y3D`、`z3D`、
`rotationX`、`rotationY` 和 `rotationZ`。

## 14. PPTX 导出

无需更改动画 JavaScript，在项目根目录运行：

```powershell
go run . pptx -time 60 examples/rose-curve/animation.js rose.pptx
go run . pptx -times 4,11,26,33,42,50,60 examples/rose-curve/animation.js storyboard.pptx
go run . pptx -interval 8 -scale 2 examples/rose-curve/animation.js sampled.pptx
```

`ppt` 为 `pptx` 的别名。`-time`、`-times`、`-interval` 互斥，选项须放在 JS 路径之前。默认导出第 0 秒、图片分辨率为逻辑画布的 2 倍，页面保持原画布比例。每次最多 500 页，时间不能越界；按间隔采样时包括动画末尾。

默认导出静态图片分镜，一页对应一个时间点，并在备注中记录时间。

原生动画模式：

```powershell
go run . pptx -native examples/pptx-native/animation.js native-animation.pptx
```

`-native` 将整段时间轴转换为一页可自动放映的 PowerPoint 原生动画，支持可编辑文字、圆、矩形、直线、多边形和路径，以及位置、正数缩放、旋转与完整淡入淡出。`-fps` 控制关键帧采样精度，默认 60；缓动近似转换。字体按名称引用、不嵌入；暂不支持渐进绘制、动态几何/颜色、图片和 3D。与 `-time`、`-times`、`-interval` 互斥。

两种导出均由 Go 完成，无需授权密钥、Office 或额外运行时。支持范围和限制见 [PPTX_EXPORT.md](PPTX_EXPORT.md)。
