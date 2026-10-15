/// <reference path="./animation-api.js" />

// NEON SIGNAL — a complete animation-engine showcase.
// Run: go run . preview index.js

const C = {
  white: "#f8fafc",
  soft: "#b8c7e0",
  muted: "#cbd5e1",
  cyan: "#22d3ee",
  blue: "#3b82f6",
  violet: "#8b5cf6",
  pink: "#ec4899",
  green: "#34d399",
  panel: "#111827dd",
  border: "#334155aa",
  grid: "#25314a55",
  clear: "#00000000",
};

// Deterministic pseudo-random values keep offline export reproducible.
let seed = 0x51a7;
function random() {
  seed = (seed * 1664525 + 1013904223) >>> 0;
  return seed / 4294967296;
}

// Deep-space star field with subtle procedural shimmer.
const stars = [];
while (stars.length < 64) {
  const x = -620 + random() * 1240;
  const y = -335 + random() * 670;
  // Keep stars away from typography-heavy header and telemetry areas.
  if ((y < -175 && Math.abs(x) < 570) || (y > 215 && Math.abs(x) < 455)) {
    continue;
  }
  const i = stars.length;
  const radius = 0.7 + random() * 1.7;
  const phase = random() * Math.PI * 2;
  const base = 0.18 + random() * 0.35;
  const star = circle({
    radius,
    position: [x, y],
    fill: i % 9 === 0 ? C.cyan : C.white,
    stroke: C.clear,
    opacity: base,
  });
  star.update((self, time) => {
    self.opacity = base + (0.12 + radius * 0.04) * Math.sin(time * 1.8 + phase);
  });
  stars.push(star);
}
const starField = group(...stars);

// Perspective grid converging toward the core.
const gridLines = [];
for (let x = -600; x <= 600; x += 100) {
  gridLines.push(line({
    start: [x, 330],
    end: [x * 0.18, -18],
    stroke: C.grid,
    strokeWidth: 1,
  }));
}
for (let y = 40; y <= 330; y += 48) {
  const halfWidth = 170 + (y / 330) * 450;
  gridLines.push(line({
    start: [-halfWidth, y],
    end: [halfWidth, y],
    stroke: C.grid,
    strokeWidth: 1,
  }));
}
const perspectiveGrid = group(...gridLines);
perspectiveGrid.opacity = 0;

// Header and live-status chip.
const eyebrow = text("SYSTEM 07  /  VISUAL COMPUTE", {
  position: [0, -292], size: 15, fill: C.cyan, opacity: 0,
});
const title = text("NEON SIGNAL", {
  position: [0, -245], size: 45, fill: C.white, opacity: 0,
});
const subtitle = text("REAL-TIME MOTION  ×  实时动态图形", {
  position: [0, -202], size: 17, fill: C.soft, opacity: 0,
});
const statusPill = rect({
  position: [470, -288], width: 178, height: 34, radius: 17,
  fill: "#0f2c27dd", stroke: "#34d39988", strokeWidth: 1, opacity: 0,
});
const statusDot = circle({
  position: [414, -288], radius: 5, fill: C.green, stroke: C.clear, opacity: 0,
});
const statusText = text("CORE ONLINE", {
  position: [486, -288], size: 13, fill: C.green, opacity: 0,
});

// Central reactor: layered rings, glow and rotating geometry.
const outerRing = circle({
  radius: 130, fill: C.clear, stroke: "#22d3ee55", strokeWidth: 2, opacity: 0,
});
const orbitRing = circle({
  radius: 96, scale: [1.45, 0.48], rotation: -0.22,
  fill: C.clear, stroke: "#8b5cf699", strokeWidth: 3, opacity: 0,
});
const innerRing = circle({
  radius: 72, fill: C.clear, stroke: "#3b82f6aa", strokeWidth: 2, opacity: 0,
});
const coreGlow = circle({
  radius: 52, fill: "#2563eb44", stroke: "#22d3ee66", strokeWidth: 4, opacity: 0,
});
const core = polygon({
  points: [[0, -39], [34, -19], [34, 19], [0, 39], [-34, 19], [-34, -19]],
  fill: "#0ea5e9dd", stroke: C.white, strokeWidth: 2, scale: 0.25, opacity: 0,
});
const coreLabel = text("07", {size: 27, fill: C.white, opacity: 0});
const coreCaption = text("QUANTUM CORE", {
  position: [0, 112], size: 14, fill: C.cyan, opacity: 0,
});

outerRing.update((self, time) => { self.rotation = time * 0.13; });
orbitRing.update((self, time) => { self.rotation = -0.22 - time * 0.2; });
innerRing.update((self, time) => { self.rotation = time * 0.28; });

// Orbit route and a bright signal packet.
const orbitPath = path({
  commands: [
    ["M", -140, 0],
    ["C", -140, -75, 140, -75, 140, 0],
    ["C", 140, 75, -140, 75, -140, 0],
  ],
  fill: C.clear, stroke: "#22d3ee33", strokeWidth: 2, opacity: 0,
});
const packetHalo = circle({
  radius: 15, fill: "#22d3ee22", stroke: C.clear, opacity: 0,
});
const packet = circle({
  radius: 6, fill: C.white, stroke: C.cyan, strokeWidth: 3, opacity: 0,
});

// Network nodes and links surrounding the reactor.
const links = [
  line({start: [-430, -70], end: [-130, 0], stroke: "#3b82f666", strokeWidth: 2, opacity: 0}),
  line({start: [-420, 125], end: [-115, 55], stroke: "#8b5cf666", strokeWidth: 2, opacity: 0}),
  line({start: [430, -70], end: [130, 0], stroke: "#22d3ee66", strokeWidth: 2, opacity: 0}),
  line({start: [420, 125], end: [115, 55], stroke: "#ec489966", strokeWidth: 2, opacity: 0}),
];

/**
 * @param {number} x
 * @param {number} y
 * @param {string} code
 * @param {string} label
 * @param {Color} accent
 * @returns {Group}
 */
function networkNode(x, y, code, label, accent) {
  const halo = circle({radius: 28, fill: `${accent}22`, stroke: `${accent}55`, strokeWidth: 1});
  const dot = circle({radius: 10, fill: accent, stroke: C.white, strokeWidth: 2});
  const codeText = text(code, {position: [0, 47], size: 14, fill: accent});
  const labelText = text(label, {position: [0, 68], size: 14, fill: C.muted});
  const item = group(halo, dot, codeText, labelText);
  item.x = x;
  item.y = y;
  item.opacity = 0;
  halo.update((self, time) => {
    self.opacity = 0.45 + 0.25 * Math.sin(time * 2.4 + x * 0.01);
  });
  return item;
}

const nodes = [
  networkNode(-430, -70, "NODE A", "INPUT / 输入", C.blue),
  networkNode(-420, 125, "NODE B", "MEMORY / 记忆", C.violet),
  networkNode(430, -70, "NODE C", "RENDER / 渲染", C.cyan),
  networkNode(420, 125, "NODE D", "OUTPUT / 输出", C.pink),
];

// Lower telemetry cards.
/**
 * @param {number} x
 * @param {string} valueText
 * @param {string} labelText
 * @param {Color} accent
 * @returns {Group}
 */
function metric(x, valueText, labelText, accent) {
  const panel = rect({
    width: 188, height: 76, radius: 14,
    fill: C.panel, stroke: C.border, strokeWidth: 1,
  });
  const stripe = rect({
    position: [-88, 0], width: 4, height: 46, radius: 2,
    fill: accent, stroke: C.clear,
  });
  const valueLabel = text(valueText, {position: [8, -11], size: 24, fill: C.white});
  const nameLabel = text(labelText, {position: [8, 19], size: 14, fill: C.muted});
  const card = group(panel, stripe, valueLabel, nameLabel);
  card.x = x;
  card.y = 268;
  card.opacity = 0;
  return card;
}

const metrics = [
  metric(-310, "60 FPS", "FRAME SYNC / 帧同步", C.cyan),
  metric(-105, "10.25 s", "TIMELINE / 时间轴", C.blue),
  metric(105, "1280×720", "CANVAS / 画布", C.violet),
  metric(310, "LIVE", "HOT RELOAD / 热更新", C.pink),
];

// A scanning beam crosses the network during the final phase.
const scanPosition = value(-610);
const scanBeam = line({
  start: [0, -165], end: [0, 218],
  stroke: "#22d3ee77", strokeWidth: 2, opacity: 0,
});
scanBeam.update((self) => { self.x = scanPosition.value; });

const finale = text("EVERY FRAME IS A MOMENT  /  每一帧，皆为此刻", {
  position: [0, 327], size: 14, fill: C.green, opacity: 0,
});

add(
  starField, perspectiveGrid, ...links, orbitPath,
  outerRing, orbitRing, innerRing, coreGlow, core, coreLabel, coreCaption,
  packetHalo, packet, ...nodes, ...metrics, scanBeam,
  eyebrow, title, subtitle, statusPill, statusDot, statusText, finale,
);

// Act I — wake the visual system.
play(parallel(
  fadeIn(perspectiveGrid),
  stagger(0.12, fadeIn(eyebrow), fadeIn(title), fadeIn(subtitle)),
), {duration: 1.25, ease: "smooth"});

play(parallel(
  stagger(0.12,
    parallel(create(links[0]), fadeIn(links[0])),
    parallel(create(links[1]), fadeIn(links[1])),
    parallel(create(links[2]), fadeIn(links[2])),
    parallel(create(links[3]), fadeIn(links[3])),
  ),
  stagger(0.14, fadeIn(nodes[0]), fadeIn(nodes[1]), fadeIn(nodes[2]), fadeIn(nodes[3])),
), {duration: 1.1, ease: "easeOut"});

// Act II — assemble and ignite the central reactor.
play(parallel(
  parallel(create(outerRing), fadeIn(outerRing)),
  parallel(create(innerRing), fadeIn(innerRing)),
  fadeIn(orbitRing), fadeIn(orbitPath),
), {duration: 1.0, ease: "easeInOut"});

play(parallel(
  fadeIn(coreGlow), fadeIn(core), scaleTo(core, 1),
  stagger(0.12, fadeIn(coreLabel), fadeIn(coreCaption)),
), {duration: 0.85, ease: "back"});

// Act III — transmit one packet around the core while telemetry arrives.
play(parallel(
  moveAlongPath(packet, orbitPath),
  moveAlongPath(packetHalo, orbitPath),
  fadeIn(packet), fadeIn(packetHalo),
  yoyo(scaleTo(packetHalo, 1.45), 4),
  rotateBy(core, Math.PI * 2),
  stagger(0.14, fadeIn(metrics[0]), fadeIn(metrics[1]), fadeIn(metrics[2]), fadeIn(metrics[3])),
), {duration: 2.25, ease: "sine"});

// Act IV — scan, pulse and confirm the live system.
play(parallel(
  fadeIn(scanBeam),
  animateValue(scanPosition, 610),
  yoyo(scaleTo(coreGlow, 1.18), 5),
  yoyo(colorTo(core, C.violet), 3),
  stagger(0.12, fadeIn(statusPill), fadeIn(statusDot), fadeIn(statusText)),
), {duration: 1.75, ease: "easeInOut"});

play(parallel(
  fadeOut(scanBeam), fadeOut(packetHalo), fadeOut(packet),
  fadeIn(finale), scaleTo(outerRing, 1.08),
), {duration: 0.85, ease: "smooth"});

wait(1.2);
