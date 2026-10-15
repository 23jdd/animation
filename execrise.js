/// <reference path="./animation-api.js" />
/**
 *
 * @param {string} content
 * @param {TextOptions} options
 */

const clear="#00000000"
const background = rect({width: 1280, height: 720, fill: "#125637", stroke: clear});
add(background)
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

add(title,ball, target);

play(write(title), { duration: 2, ease: "smooth" });

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
