/// <reference path="./animation-api.js" />

// FUNCTION LAB · 六条有趣的函数与参数曲线。
// 预览：go run . preview function-gallery.js
// 静帧：go run . render -time 10 -scale 2 function-gallery.js functions.png
// 动画：go run . gif -fps 12 -scale 0.9 function-gallery.js functions.gif
//
// 怎样画函数：
// 1. point(t) 返回数学坐标 [x, y]。
// 2. 在 range 范围内取 samples + 1 个点，连接成 path。
// 3. 画布的 Y 向下，数学的 Y 向上，所以绘制时把 Y 取负。
// 4. create() 描线，animateValue() 控制亮点，两者共用线性进度。
// 修改下方 definitions 中的公式、区间、颜色，就可以探索新的图案。

const CLEAR = "#00000000";
const PI = Math.PI;
const TAU = PI * 2;
const P = {
  background: "#0a111e",
  panel: "#101d2e",
  border: "#26394d",
  grid: "#1c2c40",
  axis: "#3b536a",
  ink: "#edf5ff",
  muted: "#9aafc7",
};

/**
 * point 是数学公式；scale 只决定它在卡片中画多大。
 * @typedef {Object} CurveDefinition
 * @property {string} title
 * @property {string} tag
 * @property {string} color
 * @property {string[]} formulas
 * @property {Vec2} range
 * @property {Vec2} scale
 * @property {number} samples
 * @property {(t: number) => Vec2} point
 */

/** @type {CurveDefinition[]} */
const definitions = [
  {
    title: "衰减的波", tag: "Y = F(X)", color: "#38bdf8",
    formulas: ["y = sin(3x) exp(-0.15x^2)", "-4 <= x <= 4"],
    range: [-4, 4], scale: [36, 55], samples: 720,
    point: (x) => [x, Math.sin(3 * x) * Math.exp(-0.15 * x * x)],
  },
  {
    title: "心形曲线", tag: "PARAMETRIC", color: "#fb7185",
    formulas: ["x = 16 sin(t)^3", "y = 13cos(t)-5cos(2t)-2cos(3t)-cos(4t)"],
    range: [0, TAU], scale: [4, 4], samples: 1200,
    point: (t) => [
      16 * Math.pow(Math.sin(t), 3),
      13 * Math.cos(t) - 5 * Math.cos(2 * t)
        - 2 * Math.cos(3 * t) - Math.cos(4 * t),
    ],
  },
  {
    title: "五瓣玫瑰", tag: "POLAR", color: "#2dd4bf",
    formulas: ["r = cos(5t)", "x = r cos(t);  y = r sin(t)"],
    // 奇数瓣玫瑰在 [0, PI] 内就能画完整，避免同一轨迹重复描两遍。
    range: [0, PI], scale: [64, 64], samples: 1400,
    point: (t) => {
      const r = Math.cos(5 * t);
      return [r * Math.cos(t), r * Math.sin(t)];
    },
  },
  {
    title: "李萨如曲线", tag: "PARAMETRIC", color: "#a78bfa",
    formulas: ["x = 1.5 sin(3t + PI/2)", "y = sin(2t)"],
    range: [0, TAU], scale: [66, 66], samples: 1600,
    point: (t) => [1.5 * Math.sin(3 * t + PI / 2), Math.sin(2 * t)],
  },
  {
    title: "生长的螺旋", tag: "POLAR", color: "#fbbf24",
    formulas: ["r = exp(0.14t)", "x = r cos(t);  y = r sin(t)"],
    range: [0, 6 * PI], scale: [4.55, 4.55], samples: 1800,
    point: (t) => {
      const r = Math.exp(0.14 * t);
      return [r * Math.cos(t), r * Math.sin(t)];
    },
  },
  {
    title: "蝴蝶曲线", tag: "PARAMETRIC", color: "#f472b6",
    formulas: ["r = exp(cos(t))-2cos(4t)-sin(t/12)^5", "x = r sin(t);  y = r cos(t)"],
    range: [0, 12 * PI], scale: [21.5, 21.5], samples: 3600,
    point: (t) => {
      const r = Math.exp(Math.cos(t)) - 2 * Math.cos(4 * t)
        - Math.pow(Math.sin(t / 12), 5);
      return [r * Math.sin(t), r * Math.cos(t)];
    },
  },
];

/** @param {string} content @param {number} left @param {number} y
 * @param {number} size @param {string} fill */
function leftText(content, left, y, size, fill) {
  let width = 0;
  for (const ch of content) width += (ch.codePointAt(0) ?? 0) > 127 ? size : size * 0.6;
  return text(content, { position: [left + width / 2, y], size, fill });
}

add(
  rect({ width: 1280, height: 720, fill: P.background, stroke: CLEAR }),
  circle({ radius: 280, position: [640, -360], fill: "#38bdf808", stroke: CLEAR }),
);
const heading = group(
  leftText("FUNCTION LAB / MATH IN MOTION", -588, -327, 11, "#38bdf8"),
  leftText("把函数画成图案", -588, -287, 32, P.ink),
  leftText("让一个点沿着公式移动，看看它会留下什么。", -588, -249, 15, P.muted),
  text("6 CURVES / 2D", { position: [492, -285], size: 13, fill: P.muted }),
);
heading.opacity = 0;
add(heading);

/**
 * @param {CurveDefinition} definition
 * @param {number} index
 */
function makePlot(definition, index) {
  const { point, range, samples, scale, color } = definition;
  /** @type {PathCommand[]} */
  const commands = [];
  for (let i = 0; i <= samples; i += 1) {
    const t = range[0] + (range[1] - range[0]) * i / samples;
    const [x, y] = point(t);
    commands.push([i === 0 ? "M" : "L", x * scale[0], -y * scale[1]]);
  }

  const frame = rect({
    width: 376, height: 240, radius: 16,
    fill: P.panel, stroke: P.border, strokeWidth: 1,
  });
  /** @type {Mobject[]} */
  const grid = [];
  for (let x = -144; x <= 144; x += 48) {
    grid.push(line({ start: [x, -67], end: [x, 67], stroke: P.grid, strokeWidth: 0.75 }));
  }
  for (let y = -48; y <= 48; y += 24) {
    grid.push(line({ start: [-157, y], end: [157, y], stroke: P.grid, strokeWidth: 0.75 }));
  }
  grid.push(
    line({ start: [-157, 0], end: [157, 0], stroke: P.axis, strokeWidth: 1 }),
    line({ start: [0, -67], end: [0, 67], stroke: P.axis, strokeWidth: 1 }),
  );

  // 两条完全相同的折线路径：宽而淡的外缘，细而亮的主线。
  const glow = path({ commands, fill: CLEAR, stroke: color + "20", strokeWidth: 9, opacity: 0 });
  const trace = path({ commands, fill: CLEAR, stroke: color, strokeWidth: 2.3, opacity: 0 });
  const halo = circle({ radius: 9, fill: color + "35", stroke: CLEAR, opacity: 0 });
  const head = circle({ radius: 3.7, fill: "#ffffff", stroke: color, strokeWidth: 1.5, opacity: 0 });
  const progress = value(0);
  const [startX, startY] = point(range[0]);
  head.x = halo.x = startX * scale[0];
  head.y = halo.y = -startY * scale[1];

  // 每帧只更新两个小圆的位置。曲线几何在加载时已经生成。
  head.update((self) => {
    const t = range[0] + (range[1] - range[0]) * progress.value;
    const [x, y] = point(t);
    self.x = x * scale[0];
    self.y = -y * scale[1];
    halo.x = self.x;
    halo.y = self.y;
  });

  const root = group(
    frame,
    leftText(String(index + 1).padStart(2, "0"), -166, -94, 12, color),
    leftText(definition.title, -140, -94, 18, P.ink),
    text(definition.tag, { position: [123, -94], size: 9, fill: P.muted }),
    ...grid,
    glow, trace, halo, head,
    text(definition.formulas[0], { position: [0, 84], size: 11.5, fill: color }),
    text(definition.formulas[1], { position: [0, 104], size: 10.5, fill: P.muted }),
  );
  root.x = -400 + index % 3 * 400;
  root.y = index < 3 ? -105 : 154;
  root.opacity = 0;
  add(root);
  return { root, trace, glow, head, halo, progress, samples };
}

const plots = definitions.map(makePlot);
const footer = group(
  text("亮点的每一个位置，都由公式计算。", {
    position: [0, 308], size: 15, fill: P.muted,
  }),
  text("Y = F(X) / POLAR / PARAMETRIC     ·     空格暂停，R 重播", {
    position: [0, 337], size: 11, fill: "#637f9b",
  }),
);
footer.opacity = 0;
add(footer);

// 时间线：展示画布 → 同时描线 → 沿完整曲线来回追踪。
play(fadeIn(heading), { duration: 0.55, ease: "smooth" });
play(parallel(
  stagger(0.1, ...plots.map((plot) => fadeIn(plot.root))),
  fadeIn(footer),
), { duration: 0.85, ease: "smooth" });
wait(0.35);

// 在描线开始的同一时刻设置可见度，避免完整曲线提前出现。
play(parallel(...plots.flatMap((plot) => [
  opacityTo(plot.trace, 1), opacityTo(plot.glow, 1),
  opacityTo(plot.head, 1), opacityTo(plot.halo, 1),
])), { duration: 0 });

play(parallel(...plots.flatMap((plot) => [
  create(plot.trace),
  create(plot.glow),
  animateValue(plot.progress, 1),
])), { duration: 7.8, ease: "linear" });
wait(0.9);

// 反向追踪不擦除曲线，让你观察点的运动与图形之间的关系。
play(parallel(...plots.map((plot) => animateValue(plot.progress, 0))), {
  duration: 6.2, ease: "linear",
});
play(parallel(...plots.map((plot) => animateValue(plot.progress, 1))), {
  duration: 6.2, ease: "linear",
});
wait(1.5);
