import { describe, expect, it } from "vitest";
import {
  CUSTOM_PRESET,
  EQ_BANDS,
  GAIN_MAX,
  GAIN_MIN,
  PRESETS,
  clampAudioSettings,
  defaultAudioSettings,
  equalPowerGains,
  orbitPosition,
  presetMatchesBands,
} from "./settings";

describe("presets", () => {
  it("gives every preset one gain per band", () => {
    PRESETS.forEach((preset) => {
      expect(preset.bands).toHaveLength(EQ_BANDS);
      preset.bands.forEach((gain) => {
        expect(gain).toBeGreaterThanOrEqual(GAIN_MIN);
        expect(gain).toBeLessThanOrEqual(GAIN_MAX);
      });
    });
  });

  it("leaves headroom in the preamp wherever a preset boosts", () => {
    PRESETS.forEach((preset) => {
      const peak = Math.max(...preset.bands);
      if (peak <= 0) return;
      // A boost of N dB has to be compensated, or the limiter does the mixing.
      expect(preset.preamp).toBeLessThanOrEqual(-Math.round(peak / 2));
    });
  });

  it("keeps the flat preset flat", () => {
    const flat = PRESETS.find((preset) => preset.id === "flat");
    expect(flat?.bands.every((gain) => gain === 0)).toBe(true);
    expect(flat?.preamp).toBe(0);
  });

  it("recognises a curve that still matches its preset", () => {
    const bass = PRESETS.find((preset) => preset.id === "bass-boost")!;
    expect(presetMatchesBands(bass, [...bass.bands])).toBe(true);
    const moved = [...bass.bands];
    moved[3] += 3;
    expect(presetMatchesBands(bass, moved)).toBe(false);
  });
});

describe("clampAudioSettings", () => {
  it("falls back to defaults on garbage", () => {
    expect(clampAudioSettings(null)).toEqual(defaultAudioSettings());
    expect(clampAudioSettings("nonsense")).toEqual(defaultAudioSettings());
    expect(clampAudioSettings({ eq: 7 })).toEqual(defaultAudioSettings());
  });

  it("clamps out-of-range gains instead of trusting them", () => {
    const settings = clampAudioSettings({
      eq: { enabled: true, preset: "custom", bands: Array(EQ_BANDS).fill(99), preamp: -400 },
    });
    expect(settings.eq.bands.every((gain) => gain === GAIN_MAX)).toBe(true);
    expect(settings.eq.preamp).toBe(GAIN_MIN);
    expect(settings.eq.enabled).toBe(true);
  });

  it("repairs a band array of the wrong length", () => {
    const settings = clampAudioSettings({ eq: { preset: "bass-boost", bands: [1, 2] } });
    const bass = PRESETS.find((preset) => preset.id === "bass-boost")!;
    expect(settings.eq.bands).toEqual(bass.bands);
  });

  it("accepts the custom preset but rejects an unknown one", () => {
    expect(clampAudioSettings({ eq: { preset: CUSTOM_PRESET } }).eq.preset).toBe(CUSTOM_PRESET);
    expect(clampAudioSettings({ eq: { preset: "nonexistent" } }).eq.preset).toBe("flat");
  });

  it("keeps the direction to one of the two values it can be", () => {
    expect(clampAudioSettings({ spatial: { direction: -1 } }).spatial.direction).toBe(-1);
    expect(clampAudioSettings({ spatial: { direction: 7 } }).spatial.direction).toBe(1);
  });

  it("clamps the spatial controls", () => {
    const settings = clampAudioSettings({
      spatial: { enabled: true, mode: "orbit", speed: 99, distance: -4, mix: 400, width: -1 },
    });
    expect(settings.spatial.mode).toBe("orbit");
    expect(settings.spatial.speed).toBeLessThanOrEqual(0.6);
    expect(settings.spatial.distance).toBe(1);
    expect(settings.spatial.mix).toBe(100);
    expect(settings.spatial.width).toBe(0);
  });

  it("drops an unknown spatial mode back to off", () => {
    expect(clampAudioSettings({ spatial: { mode: "9d" as never } }).spatial.mode).toBe("off");
  });
});

describe("equalPowerGains", () => {
  it("hands the signal over without a gap", () => {
    expect(equalPowerGains(0)).toEqual({ dry: 1, wet: 0 });
    const full = equalPowerGains(100);
    expect(full.dry).toBeCloseTo(0, 6);
    expect(full.wet).toBeCloseTo(1, 6);
  });

  it("keeps the power constant along the sweep", () => {
    for (let mix = 0; mix <= 100; mix += 10) {
      const { dry, wet } = equalPowerGains(mix);
      expect(dry * dry + wet * wet).toBeCloseTo(1, 6);
    }
  });
});

describe("orbitPosition", () => {
  it("stays on a circle of the given radius", () => {
    for (let t = 0; t < 12; t += 0.5) {
      const { x, y, z } = orbitPosition(t, 2, 0.06, 1);
      expect(Math.hypot(x, z)).toBeCloseTo(2, 6);
      expect(y).toBe(0);
    }
  });

  it("completes one revolution per period", () => {
    const radius = 1.5;
    const speed = 0.1;
    const start = orbitPosition(0, radius, speed, 1);
    const lap = orbitPosition(1 / speed, radius, speed, 1);
    expect(lap.x).toBeCloseTo(start.x, 5);
    expect(lap.z).toBeCloseTo(start.z, 5);
  });

  it("runs backwards when the direction flips", () => {
    const forward = orbitPosition(1.3, 1, 0.2, 1);
    const backward = orbitPosition(1.3, 1, 0.2, -1);
    expect(backward.z).toBeCloseTo(-forward.z, 6);
  });

  it("does not move when the speed is zero", () => {
    const a = orbitPosition(0, 2, 0, 1);
    const b = orbitPosition(9.7, 2, 0, 1);
    expect(b).toEqual(a);
  });
});
