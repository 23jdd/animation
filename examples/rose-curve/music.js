// Original, quietly romantic accompaniment. No samples or third-party packages.
// Usage: node examples/rose-curve/music.js path/to/rose-score.wav
const fs = require("node:fs");
const path = require("node:path");

const RATE = 48000;
const DURATION = 64;
const TAU = 2 * Math.PI;
const clamp = (x, low, high) => Math.max(low, Math.min(high, x));
const smooth = x => { x = clamp(x, 0, 1); return x * x * (3 - 2 * x); };
const frequency = midi => 440 * Math.pow(2, (midi - 69) / 12);

function synthesize() {
  const frames = RATE * DURATION;
  const left = new Float32Array(frames), right = new Float32Array(frames);

  function note(start, duration, midi, level, pan, kind) {
    const hz = frequency(midi);
    const begin = Math.max(0, Math.floor(start * RATE));
    const end = Math.min(frames, Math.ceil((start + duration) * RATE));
    const gainL = Math.cos((pan + 1) * Math.PI / 4);
    const gainR = Math.sin((pan + 1) * Math.PI / 4);
    for (let i = begin; i < end; i++) {
      const t = i / RATE - start, phase = TAU * hz * t;
      let envelope, tone;
      if (kind === "pad") {
        envelope = smooth(t / 1.7) * smooth((duration - t) / 2.7);
        tone = (.72 * Math.sin(phase) + .2 * Math.sin(phase * 1.0017)
          + .08 * Math.sin(phase * 2)) * (.96 + .04 * Math.sin(t * 1.3));
      } else {
        envelope = (1 - Math.exp(-t / .025)) * Math.exp(-t / 1.55)
          * smooth((duration - t) / .8);
        tone = .88 * Math.sin(phase) + .09 * Math.sin(phase * 2) * Math.exp(-t * 2)
          + .03 * Math.sin(phase * 3) * Math.exp(-t * 3);
      }
      const sample = level * envelope * tone;
      left[i] += sample * gainL;
      right[i] += sample * gainR;
    }
  }

  // Eight slow bars, with a soft C-major resolution as the bouquet gathers.
  const chords = [
    [48, 55, 59, 64], [45, 52, 55, 60], [41, 48, 52, 57],
    [43, 50, 52, 59], [52, 55, 59, 62], [41, 48, 52, 57],
    [50, 57, 60, 65], [48, 55, 59, 64],
  ];
  chords.forEach((chord, bar) => {
    chord.forEach((midi, j) => note(bar * 8, 10.5, midi, .017, (j - 1.5) * .27, "pad"));
    [1.15, 3.15, 5.4, 6.95].forEach((offset, j) => {
      if (bar * 8 + offset > DURATION - 4) return;
      note(bar * 8 + offset, 4.2, chord[[1, 3, 2, 3][j]] + 12, .041,
        j % 2 ? .28 : -.28, "bell");
    });
  });
  // A sparse original upper line; the final C is allowed to dissolve into silence.
  [[2.3, 76], [6.2, 74], [10.2, 72], [14.6, 76], [18.5, 77], [22.3, 76],
    [26.2, 74], [29.1, 79], [32.7, 76], [36.2, 74], [39.5, 72],
    [44.7, 76], [48.5, 77], [51.8, 76], [55.2, 74], [59.6, 72]]
    .forEach(([start, midi], i) => note(start, 4.4, midi, .025, .13 * Math.sin(i), "bell"));

  // Short, non-feedback echoes, copied from the dry signal to avoid runaway tails.
  const dryL = left.slice(), dryR = right.slice();
  for (const [seconds, gain] of [[.223, .12], [.419, .075], [.731, .04]]) {
    const delay = Math.round(seconds * RATE);
    for (let i = delay; i < frames; i++) {
      left[i] += dryR[i - delay] * gain;
      right[i] += dryL[i - delay] * gain;
    }
  }

  let peak = 0, energy = 0;
  for (let i = 0; i < frames; i++) {
    const fade = smooth(i / RATE / 1.2) * smooth((DURATION - i / RATE) / 4);
    left[i] *= fade; right[i] *= fade;
    peak = Math.max(peak, Math.abs(left[i]), Math.abs(right[i]));
    energy += left[i] ** 2 + right[i] ** 2;
  }
  const rms = Math.sqrt(energy / (2 * frames));
  const gain = Math.min(.035 / (rms || 1), .42 / (peak || 1));
  return {left, right, gain, peak: peak * gain, rms: rms * gain};
}

function wave(score) {
  const {left, right, gain} = score;
  const dataBytes = left.length * 4;
  const output = Buffer.alloc(44 + dataBytes);
  output.write("RIFF", 0); output.writeUInt32LE(36 + dataBytes, 4);
  output.write("WAVEfmt ", 8); output.writeUInt32LE(16, 16);
  output.writeUInt16LE(1, 20); output.writeUInt16LE(2, 22);
  output.writeUInt32LE(RATE, 24); output.writeUInt32LE(RATE * 4, 28);
  output.writeUInt16LE(4, 32); output.writeUInt16LE(16, 34);
  output.write("data", 36); output.writeUInt32LE(dataBytes, 40);
  for (let i = 0; i < left.length; i++) {
    for (const [channel, sample] of [[0, left[i]], [1, right[i]]]) {
      if (!Number.isFinite(sample)) throw new Error(`Non-finite audio sample at ${i}`);
      output.writeInt16LE(Math.round(clamp(sample * gain, -1, 1) * 32767), 44 + i * 4 + channel * 2);
    }
  }
  return output;
}

if (require.main === module) {
  const destination = path.resolve(process.argv[2] || path.join(__dirname, "rose-score.wav"));
  const score = synthesize();
  fs.mkdirSync(path.dirname(destination), {recursive: true});
  fs.writeFileSync(destination, wave(score));
  console.log(JSON.stringify({file: destination, duration: DURATION, sampleRate: RATE,
    channels: 2, peak: score.peak, rms: score.rms}));
}

module.exports = {synthesize, wave, RATE, DURATION};
