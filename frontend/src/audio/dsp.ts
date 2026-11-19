/**
 * Small DSP helpers that are plain maths: combining filter responses and
 * folding an FFT into log-spaced bars. Both are pure so they can be checked
 * without an audio device, and both are used by the panel.
 */

/** Adds per-filter responses (already in dB) into one curve. */
export function sumResponseDb(responses: Float32Array[], length: number): Float32Array {
  const total = new Float32Array(length);
  responses.forEach((response) => {
    for (let i = 0; i < length; i += 1) {
      total[i] += response[i] ?? 0;
    }
  });
  return total;
}

/** Magnitudes (0..1) to decibels, floored so silence does not become -Infinity. */
export function toDb(magnitude: number): number {
  return 20 * Math.log10(Math.max(magnitude, 1e-6));
}

export interface LogBand {
  /** First FFT bin included. */
  from: number;
  /** Last bin included, exclusive. */
  to: number;
  /** Centre frequency, for labels. */
  hz: number;
}

/**
 * Splits the spectrum into bars that are evenly spaced on a log axis, which is
 * how hearing works and how every analyser worth looking at draws it.
 */
export function logBands(barCount: number, binCount: number, sampleRate: number, fftSize: number): LogBand[] {
  const nyquist = sampleRate / 2;
  const minHz = 30;
  const maxHz = Math.min(nyquist, 18000);
  const bands: LogBand[] = [];
  const ratio = Math.pow(maxHz / minHz, 1 / barCount);
  for (let bar = 0; bar < barCount; bar += 1) {
    const low = minHz * Math.pow(ratio, bar);
    const high = minHz * Math.pow(ratio, bar + 1);
    const from = Math.max(0, Math.floor((low / nyquist) * binCount));
    const to = Math.max(from + 1, Math.ceil((high / nyquist) * binCount));
    bands.push({ from, to, hz: Math.sqrt(low * high) });
  }
  void fftSize;
  return bands;
}

export interface SpectrumScale {
  /** The analyser's minDecibels. */
  minDb: number;
  /** The analyser's maxDecibels. */
  maxDb: number;
  /** Anything below this reads as zero. Defaults to the analyser floor. */
  floorDb?: number;
}

/**
 * Peak per log band, scaled to 0..1.
 *
 * getByteFrequencyData already hands back decibels mapped onto 0..255 between
 * minDecibels and maxDecibels — taking log10 of that byte again is what turns
 * every bar into a full-height block. Peak rather than mean keeps one loud
 * partial from disappearing into its neighbours.
 */
export function spectrumBars(
  data: Uint8Array,
  bands: LogBand[],
  scale: SpectrumScale = { minDb: -90, maxDb: -10 },
): Float32Array {
  const range = Math.max(1, scale.maxDb - scale.minDb);
  const floor = scale.floorDb ?? scale.minDb;
  const out = new Float32Array(bands.length);
  bands.forEach((band, index) => {
    let peak = 0;
    for (let bin = band.from; bin < band.to && bin < data.length; bin += 1) {
      if (data[bin] > peak) peak = data[bin];
    }
    const db = scale.minDb + (peak / 255) * range;
    out[index] = Math.min(1, Math.max(0, (db - floor) / Math.max(1, scale.maxDb - floor)));
  });
  return out;
}
