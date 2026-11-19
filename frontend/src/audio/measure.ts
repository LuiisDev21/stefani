/**
 * Offline measurement of the EQ: renders the very same filter chain the player
 * uses and compares it against a flat render, so the curve on screen can be
 * checked against the audio it claims to describe.
 *
 * Nothing here touches the speakers or the live graph.
 */

import { buildEqChain } from "./engine";
import { AudioSettings, defaultAudioSettings } from "./settings";

export const MEASURE_FREQUENCIES = [31, 62, 125, 250, 500, 1000, 2000, 4000, 8000, 16000];

export interface ChainMeasurement {
  /** Absolute peak of the rendered signal, 0..1. Anything above 1 clips. */
  peak: number;
  /**
   * Difference in dB per measured frequency against a flat render at unity
   * gain, so the preamp shows up in every band as the same offset. The shape
   * of a preset is the difference between its bands, not this level.
   */
  bands: { hz: number; db: number }[];
}

function offlineCtor(): typeof OfflineAudioContext | undefined {
  const scope = globalThis as unknown as {
    OfflineAudioContext?: typeof OfflineAudioContext;
    webkitOfflineAudioContext?: typeof OfflineAudioContext;
  };
  return scope.OfflineAudioContext ?? scope.webkitOfflineAudioContext;
}

/** Deterministic white noise: the same input on every run and every machine. */
function noiseBuffer(ctx: BaseAudioContext, seconds: number, amplitude: number): AudioBuffer {
  const rate = ctx.sampleRate;
  const length = Math.floor(rate * seconds);
  const buffer = ctx.createBuffer(1, length, rate);
  const data = buffer.getChannelData(0);
  let seed = 0x2f6e2b1;
  for (let i = 0; i < length; i += 1) {
    seed = (seed * 1103515245 + 12345) & 0x7fffffff;
    data[i] = (seed / 0x3fffffff - 1) * amplitude;
  }
  return buffer;
}

async function render(settings: AudioSettings, seconds = 1): Promise<Float32Array> {
  const Ctor = offlineCtor();
  if (!Ctor) throw new Error("OfflineAudioContext no está disponible en este entorno");
  const rate = 44100;
  const ctx = new Ctor(1, Math.floor(rate * seconds), rate);
  const source = ctx.createBufferSource();
  source.buffer = noiseBuffer(ctx, seconds, 0.2);
  const chain = buildEqChain(ctx, settings);
  source.connect(chain.input);
  chain.output.connect(ctx.destination);
  source.start();
  const rendered = await ctx.startRendering();
  return rendered.getChannelData(0).slice();
}

/**
 * Goertzel: the energy at one frequency, without an FFT. Enough to ask "did
 * the 62 Hz band really move", and cheap enough to run on every band.
 */
function energyAt(data: Float32Array, rate: number, frequency: number): number {
  const n = data.length;
  const k = Math.round((n * frequency) / rate);
  const omega = (2 * Math.PI * k) / n;
  const coeff = 2 * Math.cos(omega);
  let s1 = 0;
  let s2 = 0;
  for (let i = 0; i < n; i += 1) {
    const s0 = data[i] + coeff * s1 - s2;
    s2 = s1;
    s1 = s0;
  }
  const magnitude = Math.sqrt(s1 * s1 + s2 * s2 - coeff * s1 * s2);
  return magnitude / (n / 2);
}

function peakOf(data: Float32Array): number {
  let peak = 0;
  for (let i = 0; i < data.length; i += 1) {
    const value = Math.abs(data[i]);
    if (value > peak) peak = value;
  }
  return peak;
}

/**
 * Renders the chain twice — flat and with the given settings — and reports how
 * much each measured frequency moved, plus the peak of the processed render.
 */
export async function measureEq(settings: AudioSettings): Promise<ChainMeasurement> {
  const flat = defaultAudioSettings();
  flat.eq.enabled = false;
  const [reference, processed] = await Promise.all([render(flat), render(settings)]);
  const rate = 44100;

  const bands = MEASURE_FREQUENCIES.map((hz) => {
    const before = energyAt(reference, rate, hz);
    const after = energyAt(processed, rate, hz);
    const db = 20 * Math.log10(Math.max(after, 1e-9) / Math.max(before, 1e-9));
    return { hz, db };
  });

  return { peak: peakOf(processed), bands };
}

export function dbAt(measurement: ChainMeasurement, hz: number): number {
  return measurement.bands.find((band) => band.hz === hz)?.db ?? 0;
}
