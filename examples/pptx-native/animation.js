/// <reference path="../../animation-api.js" />

// Preview the same timeline, or export it as editable PowerPoint animation:
// go run . preview examples/pptx-native/animation.js
// go run . pptx -native examples/pptx-native/animation.js native-animation.pptx

const title = text("让图形，在演示文稿里动起来。", {
  position: [0, -225], size: 42, fill: "#353b47",
});
const subtitle = text("淡入  ·  移动  ·  缩放  ·  旋转  ·  淡出", {
  position: [0, -155], size: 23, fill: "#74808b",
});
const baseline = line({start: [-420, 100], end: [420, 100],
  stroke: "#c8ced3", strokeWidth: 1,
});
const ball = circle({
  radius: 45, position: [-330, 0], fill: "#c86e8b",
  strokeWidth: 0, opacity: 0,
});
const tile = rect({
  width: 118, height: 118, radius: 18, position: [330, 0],
  fill: "#91a28e", strokeWidth: 0, opacity: 0,
});
const caption = text("同一段 JS 时间轴，导出可编辑对象与原生动画。", {
  position: [0, 200], size: 25, fill: "#596671",
});
// Use a real full-canvas shape so the background is shared by preview and PPTX.
const background = rect({width: 1280, height: 720, fill: "#f7f5f1", strokeWidth: 0});
add(background, title, subtitle, baseline, ball, tile, caption);

wait(0.5);
play(parallel(fadeIn(ball), fadeIn(tile)), {duration: 0.8});
play(parallel(
  moveTo(ball, [0, 0]),
  scaleTo(ball, 1.25),
  rotateBy(tile, Math.PI / 2),
), {duration: 2, ease: "smooth"});
wait(0.4);
play(parallel(
  moveTo(ball, [330, 0]),
  rotateBy(tile, Math.PI),
), {duration: 2, ease: "smooth"});
wait(0.4);
play(parallel(fadeOut(ball), fadeOut(tile)), {duration: 0.8});
wait(0.8);
