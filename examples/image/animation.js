/// <reference path="../../animation-api.js" />

// go run . preview examples/image/animation.js
// 本例的 PNG 来自同项目的玫瑰曲线示例，保存在 assets 中，可独立运行。
const clear = "#00000000";
const backdrop = rect({width: 1280, height: 720, fill: "#f8f3ee", stroke: clear});
const title = text("让图片，也成为动画的一部分。", {position: [0, -266], size: 35, fill: "#674d59"});
const subtitle = text("保持比例 · 居中裁剪 · 拉伸适配", {position: [0, -213], size: 18, fill: "#997d89"});

/** @param {ImageFit} fit @param {number} x @param {string} title @param {string} description */
function card(fit, x, title, description) {
  const frame = rect({width: 328, height: 300, radius: 12, fill: "#fffdf9", stroke: "#e1d0d0", strokeWidth: 1});
  const slot = rect({width: 286, height: 204, position: [0, -26], fill: "#e8e2df", stroke: clear});
  const photo = image("./assets/rose.png", {width: 286, height: 204, fit, position: [0, -26]});
  const heading = text(title, {position: [0, 103], size: 24, fill: "#a26b87"});
  const caption = text(description, {position: [0, 132], size: 13, fill: "#997d89"});
  const g = group(frame, slot, photo, heading, caption);
  g.x = x; g.opacity = 0;
  return {g, photo};
}
const cards = [
  card("contain", -356, "contain", "完整显示，留白透明"),
  card("cover", 0, "cover", "保持比例，居中裁剪"),
  card("stretch", 356, "stretch", "独立缩放宽高"),
];
const note = text("普通图片同样可以淡入、移动、旋转、缩放和翻转。", {position: [0, 246], size: 23, fill: "#674d59"});
const credit = text("PNG / JPEG / WebP / GIF 首帧 · 路径相对动画脚本 · 每张图片只解码一次", {position: [0, 300], size: 14, fill: "#a48c96"});
add(backdrop, title, subtitle, ...cards.map(c => c.g), note, credit);

play(stagger(.2, ...cards.map(c => fadeIn(c.g))), {duration: 1.8, ease: "smooth"});
wait(1.2);
play(parallel(
  rotateTo(cards[0].g, -.045), rotateTo(cards[2].g, .045),
  moveTo(cards[1].g, [0, -20]), scaleTo(cards[1].g, 1.05),
), {duration: 1.8, ease: "smooth"});
wait(1);
play(scaleTo(cards[2].photo, [-1, 1]), {duration: 1.2, ease: "smooth"});
wait(1);
play(scaleTo(cards[2].photo, 1), {duration: 1.2, ease: "smooth"});
play(parallel(
  rotateTo(cards[0].g, 0), rotateTo(cards[2].g, 0),
  moveTo(cards[1].g, [0, 0]), scaleTo(cards[1].g, 1),
), {duration: 1.5, ease: "smooth"});
wait(1.3);
play(uncreate(cards[1].photo), {duration: 1.1, ease: "smooth"});
play(create(cards[1].photo), {duration: 1.1, ease: "smooth"});
wait(1.8);
