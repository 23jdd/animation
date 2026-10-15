/// <reference path="./animation-api.js" />

// TYPEWRITER DEMO — 打字机效果演示
// 预览: go run . preview demo.js
// GIF:  go run . gif -fps 30 demo.js typewriter.gif
// 视频: go run . mp4 -fps 60 demo.js typewriter.mp4

const bg = "#0b1020", ink = "#e6eaf2", dim = "#7c8698",
      orange = "#ff9e5e", green = "#5eead4";

const background = rect({ width: 1280, height: 720, fill: bg, stroke: "#00000000" });
add(background);

// 标题 — 橙色光标
const title = text("TYPEWRITER", { position: [0, -200], size: 40, fill: orange });
title.opacity=0.5
add(title);

play(write(title, orange,1.0), { duration: 1.2, ease: "linear" });
// 中文一行 — 光标默认与文字同色
const line1 = text("用代码，雕刻时间。", { position: [0, -110], size: 30, fill: ink });
add(line1);

play(write(line1), { duration: 1.8, ease: "linear" });

// 中文二行 — 绿色光标
const line2 = text("每一帧，都是确定的。", { position: [0, -40], size: 30, fill: green });
add(line2);
play(write(line2, green), { duration: 1.8, ease: "linear" });

// 英文一行 — 默认光标
const line3 = text("A study in continuous motion.", { position: [0, 30], size: 26, fill: dim });
add(line3);
play(write(line3), { duration: 2, ease: "linear" });

wait(0.6);

// 依次擦除（unwrite 也带光标）
play(unwrite(line3), { duration: 0.9, ease: "linear" });
play(unwrite(line2), { duration: 0.9, ease: "linear" });
play(unwrite(line1), { duration: 0.9, ease: "linear" });
play(unwrite(title), { duration: 0.7, ease: "linear" });

wait(0.5);
