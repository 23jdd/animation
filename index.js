/// <reference path="./animation-api.js" />

// ORBIT / 轨迹 — an editorial motion poster, designed at 1280 × 720.
// go run . preview index.js
// go run . render -time 6 index.js orbit.png
const ink = "#20231f", paper = "#eeeade", orange = "#ed592e", clear = "#00000000";
const background = rect({width: 1280, height: 720, fill: paper, stroke: clear});
const topRule = line({start: [-580, -282], end: [580, -282], stroke: ink, strokeWidth: 1});
const bottomRule = line({start: [-580, 264], end: [580, 264], stroke: ink, strokeWidth: 1});
const brand = text("O / MOTION STUDIES", {position: [-470, -313], size: 18, fill: ink});
const edition = text("VOL. 01     /     2026", {position: [452, -313], size: 16, fill: ink});
const footer = text("用代码，雕刻时间。", {position: [-443, 310], size: 20, fill: ink});
const signature = text("FORM  /  SPACE  /  TIME", {position: [440, 310], size: 16, fill: ink});
const kicker = text("001 — THE ART OF REPETITION", {position: [-411, -209], size: 16, fill: ink, opacity: 0});
const title = text("ORBIT", {position: [-405, -105], size: 86, fill: ink, opacity: 0});
const subtitle = text("秩序之外，自有轨迹。", {position: [-398, -20], size: 22, fill: ink, opacity: 0});
const note = text("A STUDY IN CONTINUOUS MOTION", {position: [-412, 38], size: 14, fill: "#60645b", opacity: 0});
const badge = group(
  circle({radius: 23, fill: orange, stroke: clear}),
  line({start: [-9, 0], end: [9, 0], stroke: paper, strokeWidth: 2}),
  line({start: [0, -9], end: [0, 9], stroke: paper, strokeWidth: 2}),
);
badge.x = -553; badge.y = 179; badge.opacity = 0;
const caption = text("LESS NOISE. MORE MOTION.", {position: [-352, 179], size: 16, fill: ink, opacity: 0});
const spread = value(0);
const tilt = value(-0.48);
const phase = value(0.8);

// All procedural motion uses absolute time: preview, seek and export agree.
const shell = circle({radius: 207, fill: orange, stroke: clear});
const meridians = [];
for (let i = 0; i < 22; i++) {
  const ring = circle({radius: 207, fill: clear, stroke: "#542c20", strokeWidth: 1.2,
    scale: [Math.max(0.015, Math.abs(Math.cos(i * Math.PI / 22))), 1]});
  ring.update((self, time) => {
    self.rotation = 0.16 * Math.sin(time * 0.7 + i * Math.PI / 22) + spread.value * (i - 10.5) * 0.045;
    self.x = spread.value * (i - 10.5) * 5;
    self.y = spread.value * Math.sin(i * 0.7) * 18;
  });
  meridians.push(ring);
}
const latitudes = [];
for (let i = -4; i <= 4; i++) {
  const y = i * 40;
  latitudes.push(circle({position: [0, y], radius: Math.sqrt(207 * 207 - y * y),
    scale: [1, 0.16], fill: clear, stroke: "#542c20", strokeWidth: 1}));
}
const globe = group(shell, ...meridians, ...latitudes);
globe.rotation = -0.38;
const sculpture = group(globe);
sculpture.x = 284; sculpture.y = -5; sculpture.opacity = 0;
const orbit = circle({position: [284, -5], radius: 265, scale: [1, 0.40], rotation: -0.48,
  fill: clear, stroke: ink, strokeWidth: 1.5, opacity: 0});
orbit.update((self) => { self.rotation = tilt.value; });
const satellite = circle({radius: 10, fill: ink, stroke: paper, strokeWidth: 3, opacity: 0});
function orbitalPosition(self, lag) {
  const a = phase.value - lag, x = 265 * Math.cos(a), y = 106 * Math.sin(a);
  self.x = 284 + x * Math.cos(tilt.value) - y * Math.sin(tilt.value);
  self.y = -5 + x * Math.sin(tilt.value) + y * Math.cos(tilt.value);
}
satellite.update((self) => { orbitalPosition(self, 0); });
const trailDots = [];
for (let i = 12; i >= 1; i--) {
  const dot = circle({radius: 2 + (12 - i) * 0.4, fill: ink, stroke: clear, opacity: (13 - i) / 20});
  dot.update((self) => { orbitalPosition(self, i * 0.065); });
  trailDots.push(dot);
}
const trail = group(...trailDots);
trail.opacity = 0;
const pulses = [0, 1, 2].map(() => circle({position: [284, -5], radius: 207,
  fill: clear, stroke: orange, strokeWidth: 1.5, opacity: 0}));
const rays = [];
for (let i = 0; i < 24; i++) {
  const a = i * Math.PI / 12;
  rays.push(line({start: [Math.cos(a) * 220, Math.sin(a) * 220],
    end: [Math.cos(a) * (i % 3 === 0 ? 241 : 229), Math.sin(a) * (i % 3 === 0 ? 241 : 229)],
    stroke: ink, strokeWidth: 1}));
}
const dial = group(...rays);
dial.x = 284; dial.y = -5; dial.opacity = 0;
const objectLabel = text("FIG. 01   /   ORBITAL FORM", {position: [286, 235], size: 14, fill: ink, opacity: 0});
const rail = line({start: [-580, 264], end: [580, 264], stroke: orange, strokeWidth: 3, opacity: 0});
add(background, topRule, bottomRule, brand, edition, footer, signature,
  kicker, title, subtitle, note, badge, caption, ...pulses, sculpture, dial, orbit, trail, satellite, objectLabel, rail);

// 01 / Typesetting: typography rises into place while the sphere unfolds.
title.y = -70;
subtitle.y = 5;
play(parallel(
  stagger(0.12, fadeIn(kicker), fadeIn(title), fadeIn(subtitle), fadeIn(note)),
  moveTo(title, [-405, -105]), moveTo(subtitle, [-398, -20]),
), {duration: 1.2, ease: "smooth"});
play(parallel(fadeIn(sculpture), sequence(scaleTo(sculpture, [0.03, 0.03]), scaleTo(sculpture, 1)),
  rotateBy(globe, Math.PI * 2)), {duration: 1.8, ease: "smooth"});

// 02 / Launch: draw the orbit and accelerate the satellite, leaving a dotted trail.
play(parallel(create(orbit), fadeIn(orbit), fadeIn(satellite), fadeIn(trail),
  fadeIn(badge), fadeIn(caption), fadeIn(objectLabel), fadeIn(dial),
  animateValue(phase, Math.PI * 3)), {duration: 2, ease: "smooth"});

// 03 / Expansion: the meridians separate into a floating sculpture.
play(parallel(animateValue(spread, 1), opacityTo(shell, 0.18),
  animateValue(tilt, 0.45), animateValue(phase, Math.PI * 7),
  rotateBy(dial, Math.PI / 3), rotateBy(badge, Math.PI),
  stagger(0.18, ...pulses.map(p => sequence(
    opacityTo(p, 0.6), parallel(scaleTo(p, 1.16), fadeOut(p)),
  )))), {duration: 3, ease: "smooth"});

// 04 / Reassembly: an opposing tilt and a full turn bring the form together.
play(parallel(animateValue(spread, 0), opacityTo(shell, 1),
  animateValue(tilt, -0.8), animateValue(phase, Math.PI * 11),
  rotateBy(globe, Math.PI * 2), rotateBy(dial, -Math.PI / 3),
  create(rail), fadeIn(rail)), {duration: 3, ease: "smooth"});

// 05 / Settle: slow down to a composed final poster.
play(parallel(animateValue(tilt, -0.48), animateValue(phase, Math.PI * 12.5),
  fadeOut(trail), fadeOut(dial), rotateBy(badge, Math.PI / 2)), {duration: 2, ease: "smooth"});
wait(1);
