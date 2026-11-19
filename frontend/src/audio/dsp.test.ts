import { describe, expect, it } from "vitest";
import { logBands, spectrumBars, sumResponseDb, toDb } from "./dsp";

describe("sumResponseDb", () => {
  it("adds the filters the way a chain does", () => {
    const a = new Float32Array([0, 3, 6]);
    const b = new Float32Array([0, -1, 2]);
    expect(Array.from(sumResponseDb([a, b], 3))).toEqual([0, 2, 8]);
  });

  it("is identical to one filter when only one is given", () => {
    const only = new Float32Array([1, 2, 3]);
    expect(Array.from(sumResponseDb([only], 3))).toEqual([1, 2, 3]);
  });

  it("treats a missing sample as silence", () => {
    const short = new Float32Array([4]);
    expect(Array.from(sumResponseDb([short], 3))).toEqual([4, 0, 0]);
  });
});

describe("toDb", () => {
  it("maps unity to 0 dB and halves to -6", () => {
    expect(toDb(1)).toBeCloseTo(0, 6);
    expect(toDb(0.5)).toBeCloseTo(-6.02, 2);
    expect(toDb(0)).toBeLessThan(-100);
  });
});

describe("logBands", () => {
  const bands = logBands(24, 1024, 44100, 2048);

  it("returns one band per bar, in order", () => {
    expect(bands).toHaveLength(24);
    for (let i = 1; i < bands.length; i += 1) {
      expect(bands[i].from).toBeGreaterThanOrEqual(bands[i - 1].from);
      expect(bands[i].hz).toBeGreaterThan(bands[i - 1].hz);
    }
  });

  it("covers every band with at least one bin", () => {
    bands.forEach((band) => {
      expect(band.to).toBeGreaterThan(band.from);
      expect(band.to).toBeLessThanOrEqual(1024);
    });
  });

  it("starts low and stops below Nyquist", () => {
    expect(bands[0].hz).toBeLessThan(60);
    expect(bands[bands.length - 1].hz).toBeLessThan(18000);
  });
});

describe("spectrumBars", () => {
  const bands = logBands(8, 256, 44100, 512);

  it("returns a value per band", () => {
    const data = new Uint8Array(256).fill(128);
    expect(spectrumBars(data, bands)).toHaveLength(8);
  });

  it("reads silence as zero and a loud bin as high", () => {
    const silent = new Uint8Array(256);
    expect(Array.from(spectrumBars(silent, bands)).every((value) => value === 0)).toBe(true);

    const loud = new Uint8Array(256);
    loud.fill(255);
    expect(Math.max(...spectrumBars(loud, bands))).toBeGreaterThan(0.9);
  });

  it("keeps a lone peak visible instead of averaging it away", () => {
    const data = new Uint8Array(256);
    const target = bands[3];
    data[target.from] = 240;
    const bars = spectrumBars(data, bands);
    expect(bars[3]).toBeGreaterThan(0.5);
  });

  it("never leaves the 0..1 range", () => {
    const data = new Uint8Array(256);
    for (let i = 0; i < data.length; i += 1) data[i] = i % 256;
    spectrumBars(data, bands).forEach((value) => {
      expect(value).toBeGreaterThanOrEqual(0);
      expect(value).toBeLessThanOrEqual(1);
    });
  });
});
