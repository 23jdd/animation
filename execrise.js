/// <reference path="./animation-api.js" />

// 动画入门课：画面中直接显示中文讲解、代码高亮和同步演示。
// 预览：go run . preview execrise.js
// 导出：go run . gif -fps 12 -scale 0.8 execrise.js tutorial.gif
// 默认画布 1280 × 720。右侧演示区是一个独立 group，使用局部坐标。
// 第一次学习可以直接观看；修改动画请从文件下半部的「课程时间线」开始。
// 各章节的 code 数组是屏幕字幕；真正执行的动画是章节下方的 play()。

const CLEAR = "#00000000";
const P = {
  background: "#0b1220",
  panel: "#111d2e",
  code: "#0b1423",
  border: "#26374b",
  ink: "#edf4fc",
  muted: "#9aaec3",
  quiet: "#647d95",
  blue: "#38bdf8",
  green: "#2dd4bf",
  orange: "#fb923c",
};

// ── 教学画面的固定布局 ──────────────────────────────────────────────

function label(content, x, y, size = 16, fill = P.ink) {
  return text(content, { position: [x, y], size, fill });
}

// 文字默认居中绘制。这里估算文字宽度，让代码按同一左边缘排版。
function leftText(content, left, y, size = 16, fill = P.ink) {
  let width = 0;
  for (const character of content) {
    width += character.codePointAt(0) > 127 ? size : size * 0.6;
  }
  return label(content, left + width / 2, y, size, fill);
}

function stroke(start, end, color = P.border, width = 1) {
  return line({ start, end, stroke: color, strokeWidth: width });
}

const background = rect({
  width: 1280, height: 720, fill: P.background, stroke: CLEAR,
});
const leftPanel = rect({
  width: 570, height: 442, radius: 20,
  position: [-305, 13], fill: P.panel, stroke: P.border, strokeWidth: 1,
});
const rightPanel = rect({
  width: 570, height: 442, radius: 20,
  position: [305, 13], fill: P.panel, stroke: P.border, strokeWidth: 1,
});
add(
  background,
  leftText("ANIMATION / LESSON 01", -590, -324, 12, P.blue),
  leftText("把代码变成动画", -590, -281, 34),
  leftText("从一个圆开始 · 跟着画面一步步学", -590, -240, 16, P.muted),
  label("9 STEPS  /  2D", 504, -286, 14, P.muted),
  leftPanel,
  rightPanel,
  leftText("实时演示", 48, -176, 19),
  label("LOCAL COORDINATES", 465, -176, 11, P.quiet),
  stroke([46, -150], [564, -150]),
  stroke([46, 180], [564, 180]),
  label("演示区中心是 [0, 0]；向右 X 增大，向下 Y 增大", 305, 158, 13, P.muted),
  label("空格：暂停 / 继续     ← →：查看过程     R：从头学习", 0, 338, 13, P.muted),
);

// ── 演示区：所有教学对象都使用同一个局部坐标系 ────────────────────────

const gridParts = [];
for (let x = -200; x <= 200; x += 40) {
  gridParts.push(stroke([x, -105], [x, 105], "#1d2d40", 0.8));
}
for (let y = -80; y <= 80; y += 40) {
  gridParts.push(stroke([-210, y], [210, y], "#1d2d40", 0.8));
}
gridParts.push(
  stroke([-214, 0], [214, 0], "#3e566e", 1.3),
  stroke([0, -110], [0, 110], "#3e566e", 1.3),
  polygon({ points: [[214, 0], [205, -4], [205, 4]], fill: P.quiet, stroke: CLEAR }),
  polygon({ points: [[0, 110], [-4, 102], [4, 102]], fill: P.quiet, stroke: CLEAR }),
  circle({ radius: 3, fill: P.muted, stroke: CLEAR }),
  label("x", 225, 0, 12, P.quiet),
  label("y", 14, 107, 12, P.quiet),
  label("[0, 0]", 34, 18, 12, P.quiet),
);
const grid = group(...gridParts);

const ball = circle({
  radius: 26, position: [-120, 0],
  fill: P.blue, stroke: "#bae6fd", strokeWidth: 2, opacity: 0,
});
const square = rect({
  width: 58, height: 58, radius: 8, position: [-110, -25],
  fill: P.orange, stroke: "#fed7aa", strokeWidth: 2, opacity: 0,
});
const message = text("你好，动画！", {
  position: [0, 0], size: 30, fill: P.ink, opacity: 0,
});
const stage = group(grid, ball, square, message);
stage.x = 305;
stage.y = 27;
add(stage);

// 底部课程导航。亮点随章节推进，已经学过的步骤会保留颜色。
const stepNames = ["创建", "淡入", "移动", "偏移", "缩放", "旋转", "并行", "书写", "练习"];
const stepDots = [];
for (let i = 0; i < stepNames.length; i += 1) {
  const x = -520 + i * 130;
  if (i > 0) add(stroke([x - 113, 280], [x - 17, 280]));
  const dot = circle({
    radius: 11, position: [x, 280], fill: P.border, stroke: CLEAR,
  });
  stepDots.push(dot);
  add(dot, label(String(i + 1), x, 280, 12), label(stepNames[i], x, 307, 12, P.muted));
}
const stepCursor = circle({
  radius: 17, position: [-520, 280], fill: CLEAR, stroke: P.blue, strokeWidth: 1.5,
});
add(stepCursor);

// ── 教学字幕与代码卡片 ────────────────────────────────────────────────

let previousLesson = null;
const lessonWindows = [];

function annotations(...objects) {
  const result = group(...objects);
  result.x = 305;
  result.y = 27;
  result.opacity = 0;
  add(result);
  return result;
}

function ghost(x, y, radius = 26) {
  return circle({
    radius, position: [x, y], fill: CLEAR, stroke: "#38bdf870", strokeWidth: 1.5,
  });
}

function beginLesson(index, options) {
  if (previousLesson) {
    play(parallel(
      fadeOut(previousLesson.panel),
      fadeOut(previousLesson.annotations),
    ), { duration: 0.25, ease: "smooth" });
  }

  const codeTop = -61;
  const codeSpacing = 24;
  const highlight = group(
    rect({ width: 484, height: 23, radius: 4, fill: "#38bdf81c", stroke: CLEAR }),
    rect({ width: 3, height: 19, position: [-240, 0], fill: P.blue, stroke: CLEAR }),
  );
  highlight.x = -305;
  highlight.y = codeTop;
  const parts = [
    circle({ radius: 15, position: [-551, -176], fill: "#164763", stroke: CLEAR }),
    label(String(index), -551, -176, 15, P.blue),
    leftText(options.title, -520, -176, 23),
    leftText(options.explanation, -565, -132, 16, P.muted),
    rect({
      width: 526, height: 228, radius: 12, position: [-305, 10],
      fill: P.code, stroke: "#203044", strokeWidth: 1,
    }),
    leftText("lesson.js", -548, -89, 11, P.quiet),
    label("JAVASCRIPT", -105, -89, 10, P.quiet),
    highlight,
  ];
  options.code.forEach((content, row) => {
    parts.push(
      label(String(row + 1).padStart(2, "0"), -554, codeTop + row * codeSpacing, 12, P.quiet),
      leftText(content, -531, codeTop + row * codeSpacing, 15, "#d4e8fb"),
    );
  });
  parts.push(
    rect({
      width: 526, height: 66, radius: 10, position: [-305, 180],
      fill: "#142639", stroke: CLEAR,
    }),
    leftText("记住这一点", -550, 162, 12, P.blue),
    leftText(options.note, -550, 188, 15, P.ink),
    label(options.output, 305, 209, 14, P.muted),
  );
  const panel = group(...parts);
  panel.opacity = 0;
  add(panel);
  const markup = options.annotations ?? annotations();
  const lesson = { panel, annotations: markup, highlight, codeTop, codeSpacing };
  lessonWindows.push(lesson);
  play(parallel(
    fadeIn(panel),
    fadeIn(markup),
    colorTo(stepDots[index - 1], P.blue),
    moveTo(stepCursor, [-520 + (index - 1) * 130, 280]),
  ), { duration: 0.45, ease: "smooth" });
  previousLesson = lesson;
  return lesson;
}

function focus(lesson, row) {
  play(moveTo(lesson.highlight, [-305, lesson.codeTop + row * lesson.codeSpacing]), {
    duration: 0.25, ease: "smooth",
  });
}

// ── 课程时间线：讲解 → 阅读 → 高亮代码 → 同步演示 → 停留 ───────────────

// 01. 对象可以已经创建并加入场景，但仍然完全透明。
const lesson1 = beginLesson(1, {
  title: "创建一个圆",
  explanation: "circle() 创建对象，add() 把它加入场景。",
  code: [
    "const ball = circle({",
    "  radius: 26,",
    "  position: [-120, 0],",
    '  fill: "#38bdf8",',
    "  opacity: 0,",
    "});",
    "add(ball);",
  ],
  note: "opacity 为 0 时，对象存在，但暂时看不见。",
  output: "当前：对象已创建，透明度为 0",
  annotations: annotations(
    ghost(-120, 0),
    label("圆会出现在这里", -120, -52, 15, P.blue),
    label("[-120, 0]", -120, 48, 14, P.muted),
  ),
});
wait(2.1);
focus(lesson1, 4);
wait(1.7);
focus(lesson1, 6);
wait(1.4);

// 02. 淡入，并把 duration 和 ease 与实际运动对应起来。
const lesson2 = beginLesson(2, {
  title: "让圆显示出来",
  explanation: "fadeIn() 改变透明度，play() 安排动画。",
  code: [
    "play(fadeIn(ball), {",
    "  duration: 1.2,",
    '  ease: "smooth",',
    "});",
  ],
  note: "duration 的单位是秒；smooth 让变化平滑。",
  output: "观察：透明度从 0 变为 1",
});
wait(1.7);
focus(lesson2, 1);
play(fadeIn(ball), { duration: 1.2, ease: "smooth" });
wait(1.1);

// 03. 目标坐标和当前位置不同：moveTo 表示移动到哪里。
const lesson3 = beginLesson(3, {
  title: "移动到目标坐标",
  explanation: "moveTo() 指定终点，不是移动距离。",
  code: [
    "play(moveTo(ball, [120, 0]), {",
    "  duration: 2,",
    '  ease: "smooth",',
    "});",
  ],
  note: "X 从 -120 变到 120；Y 为 0，保持水平。",
  output: "起点 [-120, 0]  →  终点 [120, 0]",
  annotations: annotations(
    ghost(120, 0),
    stroke([-78, -46], [82, -46], P.blue, 1.5),
    polygon({ points: [[82, -46], [73, -50], [73, -42]], fill: P.blue, stroke: CLEAR }),
    label("向右移动", 0, -68, 15, P.blue),
    label("目标 [120, 0]", 120, 49, 14, P.muted),
  ),
});
wait(1.8);
focus(lesson3, 0);
play(moveTo(ball, [120, 0]), { duration: 2, ease: "smooth" });
wait(1.1);

// 04. 相对偏移与绝对位置放在相邻章节比较。
const lesson4 = beginLesson(4, {
  title: "在当前位置上偏移",
  explanation: "moveBy() 表示在原位置上再移动多少。",
  code: [
    "// 当前坐标：[120, 0]",
    "play(moveBy(ball, [-80, 70]), {",
    "  duration: 1.5,",
    '  ease: "smooth",',
    "});",
    "// 结果坐标：[40, 70]",
  ],
  note: "120 - 80 = 40；0 + 70 = 70。正 Y 向下。",
  output: "[120, 0] + [-80, 70] = [40, 70]",
  annotations: annotations(
    ghost(40, 70),
    label("向左 80，向下 70", -80, -70, 15, P.blue),
    label("目标 [40, 70]", 40, 110, 13, P.muted),
  ),
});
wait(2);
focus(lesson4, 1);
play(moveBy(ball, [-80, 70]), { duration: 1.5, ease: "smooth" });
focus(lesson4, 5);
wait(1.2);

// 05. 先放大，再恢复，看到 scaleTo 的绝对比例。
const lesson5 = beginLesson(5, {
  title: "放大与恢复",
  explanation: "scaleTo() 的数值相对于对象原始大小。",
  code: [
    "play(scaleTo(ball, 1.6), {",
    "  duration: 1.2,",
    "});",
    "",
    "play(scaleTo(ball, 1), {",
    "  duration: 0.8,",
    "});",
  ],
  note: "1 是原始大小，1.6 是 1.6 倍，0.5 是一半。",
  output: "观察：1 倍 → 1.6 倍 → 1 倍",
  annotations: annotations(
    ghost(40, 70, 26),
    ghost(40, 70, 26 * 1.6),
    label("位置不变，大小改变", 0, -70, 15, P.blue),
  ),
});
wait(1.8);
focus(lesson5, 0);
play(scaleTo(ball, 1.6), { duration: 1.2, ease: "smooth" });
wait(0.9);
focus(lesson5, 4);
play(scaleTo(ball, 1), { duration: 0.8, ease: "smooth" });
wait(0.8);

// 06. 方块边缘让旋转角度看得见。
const lesson6 = beginLesson(6, {
  title: "让方块旋转",
  explanation: "rotateTo() 设置目标角度，单位是弧度。",
  code: [
    "const square = rect({",
    "  width: 58, height: 58, opacity: 0,",
    "  position: [-110, -25],",
    "});",
    "add(square);",
    "play(fadeIn(square), { duration: 0.6 });",
    "play(rotateTo(square, Math.PI / 4),",
    "  { duration: 1.6 });",
  ],
  note: "Math.PI / 4 = 45°；Math.PI / 2 = 90°。",
  output: "观察方块：0° → 45°",
  annotations: annotations(
    label("45°", -110, -83, 20, P.orange),
    label("Math.PI / 4", 90, -70, 17, P.muted),
  ),
});
wait(1.8);
focus(lesson6, 5);
play(fadeIn(square), { duration: 0.6, ease: "smooth" });
focus(lesson6, 6);
play(rotateTo(square, Math.PI / 4), { duration: 1.6, ease: "smooth" });
wait(1.2);

// 07. 多个 play 是顺序；parallel 内的动作是同时。
const lesson7 = beginLesson(7, {
  title: "同时播放多个动画",
  explanation: "parallel() 让几个动作一起开始。",
  code: [
    "play(parallel(",
    "  moveTo(ball, [115, -30]),",
    '  colorTo(ball, "#2dd4bf"),',
    "  rotateBy(square, Math.PI / 4),",
    "), {",
    "  duration: 2.2,",
    "});",
  ],
  note: "rotateBy 是再转 45°；方块最后到达 90°。",
  output: "圆：移动 + 变色     方块：继续旋转",
  annotations: annotations(
    label("旋转", -110, -95, 15, P.orange),
    label("移动 + 变色", 115, -95, 15, P.green),
    ghost(115, -30),
  ),
});
wait(2.2);
focus(lesson7, 0);
play(parallel(
  moveTo(ball, [115, -30]),
  colorTo(ball, P.green),
  rotateBy(square, Math.PI / 4),
), { duration: 2.2, ease: "smooth" });
wait(1.3);

// 08. 场景转换后真正演示逐字出现的文字。
const lesson8 = beginLesson(8, {
  title: "逐字显示文字",
  explanation: "write() 让文字一个字一个字地出现。",
  code: [
    'const message = text("你好，动画！", {',
    "  opacity: 0,",
    "});",
    "add(message);",
    "play(write(message, null, 1), {",
    "  duration: 2.4,",
    "});",
  ],
  note: "null 关闭光标；1 是书写时的透明度。",
  output: "write 是逐字显示，不是汉字笔画描摹",
  annotations: annotations(
    label("先隐藏，书写开始时再显示", 0, 75, 15, P.muted),
  ),
});
play(parallel(fadeOut(ball), fadeOut(square), fadeOut(grid)), {
  duration: 0.4, ease: "smooth",
});
wait(1.8);
focus(lesson8, 4);
play(write(message, null, 1), { duration: 2.4, ease: "linear" });
wait(1.4);

// 09. 最后一屏是可动手修改的练习，而不是立即清空画面。
const lesson9 = beginLesson(9, {
  title: "轮到你来试一试",
  explanation: "一次只改一个数值，观察画面的变化。",
  code: [
    "// 练习 1：把 radius: 26 改成 40",
    "// 练习 2：把 duration: 2 改成 4",
    "// 练习 3：把 [120, 0] 改成 [0, -60]",
    "// 练习 4：换一个你喜欢的 fill 颜色",
    "",
    "// 改实际参数，不是这段展示文字。",
  ],
  note: "圆的大小改 ball 定义；动作改章节下方的 play。",
  output: "第一课完成 · 按 R 重播，或修改脚本继续练习",
});
play(fadeOut(message), { duration: 0.3 });

const completed = group(
  path({
    commands: [["M", -19, -55], ["L", -5, -41], ["L", 21, -71]],
    fill: CLEAR, stroke: P.green, strokeWidth: 5,
  }),
  label("第一课完成", 0, -9, 29),
  label("坐标决定位置，duration 决定时长", 0, 35, 16, P.muted),
  label("多个 play 按顺序，parallel 同时播放", 0, 66, 15, P.muted),
  label("现在，从改一个数值开始。", 0, 103, 16, P.green),
);
completed.x = 305;
completed.y = 27;
completed.opacity = 0;
add(completed);
play(fadeIn(completed), { duration: 0.7, ease: "smooth" });
wait(5);


