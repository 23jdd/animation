// @ts-check

/**
 * A two-dimensional coordinate: `[x, y]`.
 *
 * @typedef {[number, number]} Vec2
 */

/**
 * A three-dimensional coordinate: `[x, y, z]`.
 *
 * @typedef {[number, number, number]} Vec3
 */

/**
 * A CSS-style hexadecimal color such as `#3b82f6` or `#3b82f6ff`.
 *
 * @typedef {string} Color
 */

/**
 * Name of a built-in easing function.
 *
 * @typedef {'linear'|'easeIn'|'easeOut'|'easeInOut'|'easeInQuad'|'easeOutQuad'|'easeInOutQuad'|'cubic'|'easeInOutCubic'|'sine'|'expo'|'circ'|'back'|'bounce'|'elastic'|'smooth'|'smoother'} EasingName
 */

/**
 * Options shared by drawable objects.
 *
 * @typedef {Object} MobjectOptions
 * @property {Vec2} [position] Position relative to the center of the canvas. Defaults to `[0, 0]`.
 * @property {number} [rotation] Rotation in radians. Defaults to `0`.
 * @property {number|Vec2} [scale] Uniform or per-axis scale. Defaults to `1`.
 * @property {number} [opacity] Opacity between `0` and `1`. Defaults to `1`.
 * @property {Color} [fill] Fill color. Defaults to `#ffffff`.
 * @property {Color} [stroke] Stroke color. Defaults to `#ffffff`.
 * @property {number} [strokeWidth] Stroke width in pixels. Defaults to `2`.
 */

/**
 * Options accepted by {@link circle}.
 *
 * @typedef {MobjectOptions & {radius?: number}} CircleOptions
 */

/**
 * Options accepted by {@link rect} and {@link rectangle}.
 *
 * @typedef {MobjectOptions & {width?: number, height?: number, radius?: number}} RectangleOptions
 */

/**
 * Options accepted by {@link line}.
 *
 * @typedef {MobjectOptions & {start?: Vec2, end?: Vec2}} LineOptions
 */

/**
 * Options accepted by {@link polygon}.
 *
 * @typedef {MobjectOptions & {points: Vec2[]}} PolygonOptions
 */

/**
 * One renderer-independent path command.
 *
 * - `['M', x, y]` moves the cursor.
 * - `['L', x, y]` adds a line.
 * - `['Q', cx, cy, x, y]` adds a quadratic curve.
 * - `['C', c1x, c1y, c2x, c2y, x, y]` adds a cubic curve.
 * - `['Z']` closes the path.
 *
 * @typedef {['M', number, number] | ['L', number, number] | ['Q', number, number, number, number] | ['C', number, number, number, number, number, number] | ['Z']} PathCommand
 */

/**
 * Options accepted by {@link path}.
 *
 * @typedef {MobjectOptions & {commands: PathCommand[]}} PathOptions
 */

/**
 * Options accepted by {@link text}.
 *
 * @typedef {MobjectOptions & {size?: number, font?: string}} TextOptions
 */

/** @typedef {'contain'|'cover'|'stretch'} ImageFit */

/**
 * Options accepted by {@link image}. Fill/stroke options do not tint a bitmap.
 * With neither dimension supplied, the image uses its original pixel size.
 * With one dimension supplied, the other is calculated from the aspect ratio.
 * With both supplied, `fit` controls how the bitmap fits this display box.
 *
 * @typedef {MobjectOptions & {width?: number, height?: number, fit?: ImageFit}} ImageOptions
 */

/**
 * Options shared by 3D meshes.
 *
 * @typedef {Object} Mesh3DOptions
 * @property {Vec3} [position] World-space position. Defaults to `[0, 0, 0]`.
 * @property {Vec3} [rotation] Euler rotation `[x, y, z]` in radians.
 * @property {number|Vec3} [scale] Uniform or per-axis 3D scale.
 * @property {number} [opacity] Opacity between `0` and `1`.
 * @property {Color} [fill] Base surface color.
 * @property {Color} [stroke] Edge color.
 * @property {number} [strokeWidth] Edge width in pixels.
 * @property {boolean} [wireframe] Draw edges without filled faces.
 * @property {boolean} [cullBackfaces] Hide faces pointing away from the camera.
 * @property {boolean} [shading] Apply directional lighting.
 * @property {boolean} [smoothShading] Interpolate vertex normals for smooth curved surfaces. Spheres and tori enable it by default.
 * @property {number} [emissive] Unlit color contribution between `0` and `1`.
 * @property {number} [specular] White highlight strength between `0` and `2`. Defaults to `0.12`.
 * @property {number} [shininess] Highlight sharpness between `1` and `256`. Defaults to `32`.
 * @property {number} [glow] Screen-space glow radius in pixels.
 * @property {Color} [glowColor] Glow color.
 * @property {Color[]} [faceColors] Optional color override for each face.
 */

/** @typedef {Mesh3DOptions & {vertices: Vec3[], faces: number[][], normals?: Vec3[]}} CustomMesh3DOptions */
/** @typedef {Mesh3DOptions & {width?: number, height?: number, depth?: number}} Box3DOptions */
/** @typedef {Mesh3DOptions & {radius?: number, segments?: number, rings?: number}} Sphere3DOptions */
/** @typedef {Mesh3DOptions & {majorRadius?: number, minorRadius?: number, segments?: number, sides?: number}} Torus3DOptions */

/**
 * 3D camera and lighting options.
 *
 * @typedef {Object} Camera3DOptions
 * @property {Vec3} [position] Camera position. Defaults to `[0, 0, 800]`.
 * @property {Vec3} [target] Point the camera looks at.
 * @property {Vec3} [up] Camera up direction.
 * @property {number} [fov] Vertical field of view in degrees.
 * @property {number} [near] Near clipping distance.
 * @property {Vec3} [lightDirection] Direction from the surface toward the light.
 * @property {number} [ambient] Ambient light contribution.
 * @property {number} [diffuse] Directional light strength.
 * @property {Color} [fogColor] Atmospheric fog color.
 * @property {number} [fogNear] Distance where fog begins.
 * @property {number} [fogFar] Distance where fog reaches full strength.
 * @property {number} [fogStrength] Fog amount between `0` and `1`. Defaults to `0`.
 */

/**
 * Options accepted by {@link play}.
 *
 * @typedef {Object} PlayOptions
 * @property {number} [duration] Duration in seconds. Defaults to the animation duration.
 * @property {EasingName} [ease] Easing function. Defaults to `linear`.
 */

/**
 * An object that can be animated and rendered.
 *
 * @typedef {Object} Mobject
 * @property {number} x Horizontal position.
 * @property {number} y Vertical position.
 * @property {number} rotation Rotation in radians.
 * @property {number} scaleX Horizontal scale, including negative values for mirroring.
 * @property {number} scaleY Vertical scale, including negative values for mirroring.
 * @property {number} opacity Opacity between `0` and `1`.
 * @property {function(Updater): Mobject} update Attaches a procedural updater and returns this object.
 */

/**
 * A function called whenever the scene updates.
 *
 * @callback Updater
 * @param {Mobject} self Object being updated.
 * @param {number} time Absolute timeline time in seconds.
 * @param {number} dt Time since the previous update in seconds.
 * @returns {void}
 */

/**
 * An animation description accepted by {@link play} and composition functions.
 *
 * @typedef {Object} AnimationHandle
 * @property {boolean} __animationHandle Editor-only type marker.
 */

/**
 * A numeric value controlled by an animation.
 *
 * @typedef {Object} ValueTracker
 * @property {number} value Current tracked value.
 */

/** @typedef {Mobject} Circle */
/** @typedef {Mobject} Rectangle */
/** @typedef {Mobject} Line */
/** @typedef {Mobject} Polygon */
/** @typedef {Mobject} PathMobject */
/** @typedef {Mobject} TextMobject */
/**
 * Image dimensions and metadata are read-only; animate size using scaleTo.
 *
 * @typedef {Mobject & Readonly<{
 *   width: number, height: number, naturalWidth: number, naturalHeight: number,
 *   source: string, fit: ImageFit
 * }>} ImageMobject
 */
/** @typedef {Mobject} Group */
/**
 * @typedef {Mobject & {
 *   x3D: number, y3D: number, z3D: number,
 *   rotationX: number, rotationY: number, rotationZ: number
 * }} Mesh3DMobject
 */
/**
 * @typedef {Object} Camera3DHandle
 * @property {number} x Camera world X.
 * @property {number} y Camera world Y.
 * @property {number} z Camera world Z.
 * @property {number} targetX Look-at target X.
 * @property {number} targetY Look-at target Y.
 * @property {number} targetZ Look-at target Z.
 * @property {number} fov Vertical field of view in degrees.
 * @property {number} ambient Ambient light strength.
 * @property {number} diffuse Directional light strength.
 * @property {number} fogNear Distance where fog begins.
 * @property {number} fogFar Distance where fog reaches full strength.
 * @property {number} fogStrength Fog amount between `0` and `1`.
 */

/**
 * Creates a circle.
 *
 * @param {CircleOptions} [options] Circle options.
 * @returns {Circle} Circle mobject.
 */
function circle(options = {}) {
  return /** @type {Circle} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a rectangle.
 *
 * @param {RectangleOptions} [options] Rectangle options.
 * @returns {Rectangle} Rectangle mobject.
 */
function rect(options = {}) {
  return /** @type {Rectangle} */ (/** @type {unknown} */ ({}));
}

/**
 * Alias of {@link rect}.
 *
 * @param {RectangleOptions} [options] Rectangle options.
 * @returns {Rectangle} Rectangle mobject.
 */
function rectangle(options = {}) {
  return rect(options);
}

/**
 * Creates a line.
 *
 * @param {LineOptions} [options] Line options.
 * @returns {Line} Line mobject.
 */
function line(options = {}) {
  return /** @type {Line} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a polygon.
 *
 * @param {PolygonOptions} options Polygon options.
 * @returns {Polygon} Polygon mobject.
 */
function polygon(options) {
  return /** @type {Polygon} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a vector path.
 *
 * @param {PathOptions} options Path options.
 * @returns {PathMobject} Path mobject.
 */
function path(options) {
  return /** @type {PathMobject} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a text object.
 *
 * @param {string} content Text content.
 * @param {TextOptions} [options] Text options.
 * @returns {TextMobject} Text mobject.
 */
function text(content, options = {}) {
  return /** @type {TextMobject} */ (/** @type {unknown} */ ({}));
}

/**
 * Loads a local PNG, JPEG, static WebP or GIF (first frame) as an animatable object.
 * Relative paths resolve from the scene script's directory, not the shell cwd.
 * Decoded pixels are shared within the runtime. Errors throw a JS TypeError.
 * Supports move/scale/rotate/fade, groups, and a left-to-right create reveal.
 * SVG, network URLs, data URLs and animated-image playback are not supported.
 *
 * @param {string} source Local image file path.
 * @param {ImageOptions} [options] Display box and common transform options.
 * @returns {ImageMobject} Image mobject. Call add() to put it in the scene.
 */
function image(source, options = {}) {
  return /** @type {ImageMobject} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a group of mobjects.
 *
 * @param {...Mobject} children Child objects.
 * @returns {Group} Group mobject.
 */
function group(...children) {
  return /** @type {Group} */ (/** @type {unknown} */ ({}));
}

/**
 * Creates a custom polygon mesh in 3D world space.
 *
 * @param {CustomMesh3DOptions} options Mesh geometry and material.
 * @returns {Mesh3DMobject} 3D mesh.
 */
function mesh3D(options) {
  return /** @type {Mesh3DMobject} */ (/** @type {unknown} */ ({}));
}

/** @param {Box3DOptions} [options] @returns {Mesh3DMobject} */
function box3D(options = {}) {
  return /** @type {Mesh3DMobject} */ (/** @type {unknown} */ ({}));
}

/** @param {Sphere3DOptions} [options] @returns {Mesh3DMobject} */
function sphere3D(options = {}) {
  return /** @type {Mesh3DMobject} */ (/** @type {unknown} */ ({}));
}

/** @param {Torus3DOptions} [options] @returns {Mesh3DMobject} */
function torus3D(options = {}) {
  return /** @type {Mesh3DMobject} */ (/** @type {unknown} */ ({}));
}

/**
 * Configures the scene's perspective camera and directional light.
 *
 * @param {Camera3DOptions} [options]
 * @returns {Camera3DHandle}
 */
function camera3D(options = {}) {
  return /** @type {Camera3DHandle} */ (/** @type {unknown} */ ({}));
}

/**
 * Adds one or more objects to the scene.
 *
 * @param {...Mobject} objects Objects to add.
 * @returns {void}
 */
function add(...objects) {}

/**
 * Removes one or more objects from the scene.
 *
 * @param {...Mobject} objects Objects to remove.
 * @returns {void}
 */
function remove(...objects) {}

/**
 * Schedules animations at the current timeline position. Multiple animations run in parallel.
 * A {@link PlayOptions} object may be supplied as the final argument.
 *
 * @param {...(AnimationHandle|PlayOptions)} animationsAndOptions Animations followed by optional timing options.
 * @returns {void}
 */
function play(...animationsAndOptions) {}

/**
 * Advances the timeline without an animation.
 *
 * @param {number} [seconds] Number of seconds. Defaults to `1`.
 * @returns {void}
 */
function wait(seconds = 1) {}

/**
 * Moves an object to an absolute position.
 *
 * @param {Mobject} target Target object.
 * @param {Vec2} destination Destination position.
 * @returns {AnimationHandle} Animation handle.
 */
function moveTo(target, destination) { return animationStub(); }

/**
 * Moves an object by an offset.
 *
 * @param {Mobject} target Target object.
 * @param {Vec2} offset Position offset.
 * @returns {AnimationHandle} Animation handle.
 */
function moveBy(target, offset) { return animationStub(); }

/**
 * Scales an object to an absolute scale.
 *
 * @param {Mobject} target Target object.
 * @param {number|Vec2} scale Destination scale.
 * @returns {AnimationHandle} Animation handle.
 */
function scaleTo(target, scale) { return animationStub(); }

/**
 * Scales an object by a factor.
 *
 * @param {Mobject} target Target object.
 * @param {number|Vec2} factor Scale factor.
 * @returns {AnimationHandle} Animation handle.
 */
function scaleBy(target, factor) { return animationStub(); }

/**
 * Rotates an object to an absolute angle.
 *
 * @param {Mobject} target Target object.
 * @param {number} radians Destination angle in radians.
 * @returns {AnimationHandle} Animation handle.
 */
function rotateTo(target, radians) { return animationStub(); }

/**
 * Rotates an object by an angle.
 *
 * @param {Mobject} target Target object.
 * @param {number} radians Angle in radians.
 * @returns {AnimationHandle} Animation handle.
 */
function rotateBy(target, radians) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {Vec3} destination @returns {AnimationHandle} */
function move3DTo(target, destination) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {Vec3} offset @returns {AnimationHandle} */
function move3DBy(target, offset) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {Vec3} rotation @returns {AnimationHandle} */
function rotate3DTo(target, rotation) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {Vec3} rotation @returns {AnimationHandle} */
function rotate3DBy(target, rotation) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {number|Vec3} scale @returns {AnimationHandle} */
function scale3DTo(target, scale) { return animationStub(); }

/** @param {Mesh3DMobject} target @param {number|Vec3} factor @returns {AnimationHandle} */
function scale3DBy(target, factor) { return animationStub(); }

/** @param {Camera3DHandle} camera @param {Vec3} position @returns {AnimationHandle} */
function moveCamera3DTo(camera, position) { return animationStub(); }

/** @param {Camera3DHandle} camera @param {Vec3} target @returns {AnimationHandle} */
function lookAt3D(camera, target) { return animationStub(); }

/** @param {Camera3DHandle} camera @param {number} fov @returns {AnimationHandle} */
function cameraFOVTo(camera, fov) { return animationStub(); }

/**
 * Fades an object in.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function fadeIn(target) { return animationStub(); }

/**
 * Fades an object out.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function fadeOut(target) { return animationStub(); }

/**
 * Animates an object's opacity.
 *
 * @param {Mobject} target Target object.
 * @param {number} opacity Destination opacity between `0` and `1`.
 * @returns {AnimationHandle} Animation handle.
 */
function opacityTo(target, opacity) { return animationStub(); }

/**
 * Fades an object's opacity to a specified value. Alias of {@link opacityTo}.
 *
 * @param {Mobject} target Target object.
 * @param {number} opacity Destination opacity between `0` and `1`.
 * @returns {AnimationHandle} Animation handle.
 */
function fadeTo(target, opacity) { return animationStub(); }

/**
 * Animates an object's fill color.
 *
 * @param {Mobject} target Target object.
 * @param {Color} color Destination color.
 * @returns {AnimationHandle} Animation handle.
 */
function colorTo(target, color) { return animationStub(); }

/**
 * Reveals an object by drawing it.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function create(target) { return animationStub(); }

/**
 * Hides an object by reversing its draw progress.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function uncreate(target) { return animationStub(); }

/**
 * Reveals text progressively, one Unicode character at a time, with a blinking
 * typewriter caret following the last character.
 *
 * @param {TextMobject} target Text object to write.
 * @param {Color} [cursor] Caret color. Defaults to the text fill color; pass `null` to hide the caret.
 * @param {number} [opacity] Sets the object's opacity directly (same as `target.opacity = opacity`).
 * @returns {AnimationHandle} Animation handle.
 */
function write(target, cursor, opacity) { return animationStub(); }

/**
 * Erases text progressively, one Unicode character at a time, with a blinking
 * typewriter caret following the last character.
 *
 * @param {TextMobject} target Text object to erase.
 * @param {Color} [cursor] Caret color. Defaults to the text fill color; pass `null` to hide the caret.
 * @returns {AnimationHandle} Animation handle.
 */
function unwrite(target, cursor) { return animationStub(); }

/**
 * Draws a line progressively from its start to its end.
 *
 * @param {Line} target Line object to draw.
 * @param {Mobject} [tip] Optional object that follows the drawing head and is hidden on completion.
 * @returns {AnimationHandle} Animation handle.
 */
function drawLine(target, tip) { return animationStub(); }

/**
 * Reveals a line by growing it outward from its midpoint.
 *
 * @param {Line} target Line object to draw.
 * @returns {AnimationHandle} Animation handle.
 */
function drawLineFromCenter(target) { return animationStub(); }

/**
 * Moves an object along a path.
 *
 * @param {Mobject} target Target object.
 * @param {PathMobject} route Path to follow.
 * @returns {AnimationHandle} Animation handle.
 */
function moveAlongPath(target, route) { return animationStub(); }

/**
 * Transforms one object toward another object's state.
 *
 * @param {Mobject} source Source object.
 * @param {Mobject} destination Destination object.
 * @returns {AnimationHandle} Animation handle.
 */
function transform(source, destination) { return animationStub(); }

/**
 * Transforms the source and replaces it with the destination.
 *
 * @param {Mobject} source Source object.
 * @param {Mobject} destination Destination object.
 * @returns {AnimationHandle} Animation handle.
 */
function replacementTransform(source, destination) { return animationStub(); }

/**
 * Scales an object up from zero at its center while fading it in.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function growFromCenter(target) { return animationStub(); }

/**
 * Scales an object down to zero at its center while fading it out.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function shrinkToCenter(target) { return animationStub(); }

/**
 * Scales an object up from a fixed point, moving it back into place.
 *
 * @param {Mobject} target Target object.
 * @param {Vec2} [point] Origin point. Defaults to `[0, 0]`.
 * @returns {AnimationHandle} Animation handle.
 */
function growFromPoint(target, point) { return animationStub(); }

/**
 * Grows an object in from a direction.
 *
 * @param {Mobject} target Target object.
 * @param {'left'|'right'|'up'|'down'} [edge] Direction to grow from. Defaults to `'up'`.
 * @param {number} [distance] Starting offset in pixels. Defaults to `150`.
 * @returns {AnimationHandle} Animation handle.
 */
function growFromEdge(target, edge, distance) { return animationStub(); }

/**
 * Spins an object in from a full turn while scaling it up and fading it in.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function spinIn(target) { return animationStub(); }

/**
 * Spins an object out with a full turn while scaling it down and fading it out.
 *
 * @param {Mobject} target Target object.
 * @returns {AnimationHandle} Animation handle.
 */
function spinOut(target) { return animationStub(); }

/**
 * Briefly brightens an object's fill and stroke toward a highlight color.
 *
 * @param {Mobject} target Target object.
 * @param {Color} [color] Highlight color. Defaults to `#ffffff`.
 * @returns {AnimationHandle} Animation handle.
 */
function flash(target, color) { return animationStub(); }

/**
 * Draws attention to an object by briefly scaling it up and tinting it.
 *
 * @param {Mobject} target Target object.
 * @param {number} [scale] Peak scale factor. Defaults to `1.3`.
 * @param {Color} [color] Highlight color. Defaults to `#ffffff`.
 * @returns {AnimationHandle} Animation handle.
 */
function indicate(target, scale, color) { return animationStub(); }

/**
 * Dips an object's opacity to zero one or more times.
 *
 * @param {Mobject} target Target object.
 * @param {number} [times] Number of blinks. Defaults to `1`.
 * @returns {AnimationHandle} Animation handle.
 */
function blink(target, times) { return animationStub(); }

/**
 * Wobbles an object's rotation around its resting angle.
 *
 * @param {Mobject} target Target object.
 * @param {number} [amplitude] Wobble amplitude in radians. Defaults to `0.3`.
 * @param {number} [cycles] Number of wobble cycles. Defaults to `2`.
 * @returns {AnimationHandle} Animation handle.
 */
function wiggle(target, amplitude, cycles) { return animationStub(); }

/**
 * Moves a group's children (or a single object) up and down with a phase
 * offset, producing a traveling wave.
 *
 * @param {Mobject} target Target object or group.
 * @param {number} [amplitude] Vertical amplitude in pixels. Defaults to `20`.
 * @returns {AnimationHandle} Animation handle.
 */
function wave(target, amplitude) { return animationStub(); }

/**
 * Scales an object down from an enlarged size into place while fading it in.
 *
 * @param {Mobject} target Target object.
 * @param {number} [fromScale] Starting scale factor. Defaults to `2.5`.
 * @returns {AnimationHandle} Animation handle.
 */
function zoomIn(target, fromScale) { return animationStub(); }

/**
 * Scales an object up to an enlarged size while fading it out.
 *
 * @param {Mobject} target Target object.
 * @param {number} [toScale] Ending scale factor. Defaults to `2.5`.
 * @returns {AnimationHandle} Animation handle.
 */
function zoomOut(target, toScale) { return animationStub(); }

/**
 * Briefly zooms an object in and back out to draw attention to it.
 *
 * @param {Mobject} target Target object.
 * @param {number} [scale] Peak scale factor. Defaults to `1.25`.
 * @returns {AnimationHandle} Animation handle.
 */
function focusOn(target, scale) { return animationStub(); }

/**
 * Runs animations one after another.
 *
 * @param {...AnimationHandle} animations Animations to compose.
 * @returns {AnimationHandle} Composed animation.
 */
function sequence(...animations) { return animationStub(); }

/**
 * Runs animations at the same time.
 *
 * @param {...AnimationHandle} animations Animations to compose.
 * @returns {AnimationHandle} Composed animation.
 */
function parallel(...animations) { return animationStub(); }

/**
 * Starts each animation after a fixed delay.
 *
 * @param {number} delay Delay between animation starts in seconds.
 * @param {...AnimationHandle} animations Animations to compose.
 * @returns {AnimationHandle} Composed animation.
 */
function stagger(delay, ...animations) { return animationStub(); }

/**
 * Repeats an animation.
 *
 * @param {AnimationHandle} animation Animation to repeat.
 * @param {number} count Number of repetitions.
 * @returns {AnimationHandle} Repeated animation.
 */
function repeat(animation, count) { return animationStub(); }

/**
 * Repeats an animation. The default repeat count is `2`.
 *
 * @param {AnimationHandle} animation Animation to loop.
 * @param {number} [count] Number of repetitions.
 * @returns {AnimationHandle} Repeated animation.
 */
function loop(animation, count = 2) { return animationStub(); }

/**
 * Repeats an animation while alternating its direction.
 *
 * @param {AnimationHandle} animation Animation to alternate.
 * @param {number} count Number of repetitions.
 * @returns {AnimationHandle} Alternating animation.
 */
function yoyo(animation, count) { return animationStub(); }

/**
 * Creates an animatable numeric value.
 *
 * @param {number} [initial] Initial value. Defaults to `0`.
 * @returns {ValueTracker} Value tracker.
 */
function value(initial = 0) { return {value: initial}; }

/**
 * Animates a numeric value tracker.
 *
 * @param {ValueTracker} tracker Value tracker.
 * @param {number} destination Destination value.
 * @returns {AnimationHandle} Animation handle.
 */
function animateValue(tracker, destination) { return animationStub(); }

/**
 * Creates an editor-only animation placeholder.
 *
 * @private
 * @returns {AnimationHandle} Animation placeholder.
 */
function animationStub() {
  return /** @type {AnimationHandle} */ (/** @type {unknown} */ ({}));
}
