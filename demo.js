/// <reference path="./animation-api.js" />

// EFFECTS DEMO — 动画效果演示
// 预览: go run . preview demo.js
// GIF:  go run . gif -fps 30 demo.js effects.gif
// 视频: go run . mp4 -fps 60 demo.js effects.mp4

const bg = "#0b1020", ink = "#e6eaf2", dim = "#7c8698";
const blue = "#4cc9f0", orange = "#ff9e5e", green = "#5eead4",
      pink = "#f472b6", yellow = "#facc15", purple = "#a78bfa";

const background = rect({ width: 1280, height: 720, fill: bg, stroke: "#00000000" });
const title = text("EFFECTS DEMO", { position: [0, -295], size: 42, fill: ink });
const rule = line({ start: [-570, -252], end: [570, -252], stroke: dim, strokeWidth: 1 });
const subtitle = text("grow · spin · emphasis · wave · zoom", { position: [0, -218], size: 20, fill: dim });
add(background, title, rule, subtitle);


play(growFromCenter(title), { duration: 0.7, ease: "smooth" });
play(write(subtitle), { duration: 0.8 });
play(create(rule), { duration: 0.6 });


const ball = circle({ radius: 46, position: [-450, -60], fill: blue, stroke: "#dbeafe" });
const box = rect({ width: 100, height: 100, position: [-180, -60], radius: 18, fill: orange, stroke: "#ffedd5" });
const coin = circle({ radius: 34, position: [90, -60], fill: yellow, stroke: "#fef9c3" });
const diamond = polygon({
  points: [[0, -44], [44, 0], [0, 44], [-44, 0]],
  position: [360, -60], fill: green, stroke: "#d1fae5",
});
add(ball, box, coin, diamond);

play(growFromCenter(ball), { duration: 0.7, ease: "smooth" });
play(growFromEdge(box, "right", 160), { duration: 0.7, ease: "smooth" });
play(growFromPoint(coin, [0, 0]), { duration: 0.7, ease: "smooth" });
play(growFromCenter(diamond), { duration: 0.7, ease: "smooth" });
wait(0.4);


const triangle = polygon({
  points: [[0, -46], [-42, 40], [42, 40]],
  position: [-270, 60], fill: pink, stroke: "#fce7f3",
});
const hexagon = polygon({
  points: [[0, -42], [36, -21], [36, 21], [0, 42], [-36, 21], [-36, -21]],
  position: [90, 60], fill: purple, stroke: "#ede9fe",
});
add(triangle, hexagon);

play(spinIn(triangle), { duration: 0.8 });
play(spinIn(hexagon), { duration: 0.8 });
wait(0.4);


play(flash(ball, "#ffffff"), { duration: 0.5 });
play(indicate(box, 1.35, "#ffffff"), { duration: 0.6 });
play(blink(coin, 3), { duration: 0.7 });
play(wiggle(diamond, 0.5, 3), { duration: 0.8 });
play(focusOn(triangle, 1.3), { duration: 0.7 });
wait(0.3);


const beads = [];
for (let i = 0; i < 10; i++) {
  beads.push(circle({
    radius: 12,
    position: [-450 + i * 100, 150],
    fill: i % 2 === 0 ? blue : green,
    stroke: "#00000000",
  }));
}
const waveGroup = group(...beads);
add(waveGroup);

play(stagger(0.05, ...beads.map(b => growFromCenter(b))), { duration: 0.5 });
play(wave(waveGroup, 26), { duration: 1.6 });
wait(0.2);


const ring = circle({
  radius: 56, position: [-450, 250],
  fill: "#00000000", stroke: yellow, strokeWidth: 4,
});
const zoomBox = rect({
  width: 110, height: 110, position: [-180, 250], radius: 20,
  fill: purple, stroke: "#ede9fe",
});
const focusDot = circle({ radius: 30, position: [90, 250], fill: orange, stroke: "#ffedd5" });
const outBall = circle({ radius: 46, position: [360, 250], fill: green, stroke: "#d1fae5" });
add(ring, zoomBox, focusDot, outBall);

play(zoomIn(ring, 3), { duration: 0.7, ease: "smooth" });
play(zoomIn(zoomBox, 2.5), { duration: 0.7, ease: "smooth" });
play(focusOn(focusDot, 1.35), { duration: 0.8 });
play(growFromCenter(outBall), { duration: 0.7, ease: "smooth" });
wait(0.3);


play(parallel(
  shrinkToCenter(ball),
  spinOut(box),
  spinOut(triangle),
  shrinkToCenter(diamond),
  shrinkToCenter(hexagon),
), { duration: 0.8 });

play(parallel(
  shrinkToCenter(coin),
  ...beads.map(b => shrinkToCenter(b)),
  zoomOut(ring, 2.5),
  zoomOut(zoomBox, 2.5),
  shrinkToCenter(focusDot),
  spinOut(outBall),
), { duration: 0.8 });

wait(0.5);
