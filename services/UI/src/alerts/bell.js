const PARTIALS = [
  { ratio: 1, gain: 1, decay: 2.4 },
  { ratio: 2.76, gain: 0.55, decay: 1.4 },
  { ratio: 5.4, gain: 0.32, decay: 0.8 },
  { ratio: 8.93, gain: 0.18, decay: 0.5 },
];
const BASE_FREQUENCY = 880;
const STRIKES = [0, 0.55, 1.1];
const VOLUME = 0.9;

let context = null;
const listeners = new Set();

function audio() {
  if (!context) {
    const Context = window.AudioContext || window.webkitAudioContext;

    context = Context ? new Context() : null;
  }

  return context;
}

function notify() {
  listeners.forEach((listener) => listener(soundReady()));
}

export function soundReady() {
  return Boolean(context && context.state === "running");
}

export function onSoundChange(listener) {
  listeners.add(listener);

  return () => listeners.delete(listener);
}

export function unlockSound() {
  const current = audio();

  if (current && current.state !== "running") {
    current.resume().then(notify, notify);
  }
}

function strike(current, start) {
  const master = current.createGain();

  master.gain.value = VOLUME;
  master.connect(current.destination);

  for (const partial of PARTIALS) {
    const oscillator = current.createOscillator();
    const gain = current.createGain();

    oscillator.type = "sine";
    oscillator.frequency.value = BASE_FREQUENCY * partial.ratio;
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(partial.gain, start + 0.005);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + partial.decay);
    oscillator.connect(gain).connect(master);
    oscillator.start(start);
    oscillator.stop(start + partial.decay + 0.05);
  }
}

export function ring() {
  const current = audio();

  if (!current || current.state !== "running") {
    return false;
  }

  STRIKES.forEach((offset) => strike(current, current.currentTime + offset));

  return true;
}

if (typeof window !== "undefined") {
  ["pointerdown", "keydown"].forEach((event) => window.addEventListener(event, unlockSound, { passive: true }));
}
