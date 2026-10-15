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

光标颜色默认与文字填充色一致，也可以用第二个参数指定，或传 `null` 关闭。第三个参数可设置文字的整体透明度（默认 `1`）：

```js
play(write(title, "#f97316"));      // 橙色光标
play(write(title, null));           // 无光标
play(write(title, null, 0.5));      // 无光标，文字半透明
play(write(title, "#f97316", 0.6)); // 橙色光标，文字 60% 透明
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
