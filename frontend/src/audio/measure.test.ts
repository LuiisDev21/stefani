// @vitest-environment node
/**
 * These tests render real audio through the real filter chain. They need a Web
 * Audio implementation, which jsdom does not have, so they run in Node with
 * node-web-audio-api providing OfflineAudioContext.
 */
import { OfflineAudioContext } from "node-web-audio-api";
import { beforeAll, describe, expect, it } from "vitest";
import { dbAt, measureEq } from "./measure";
import { PRESETS, clampAudioSettings, defaultAudioSettings } from "./settings";

beforeAll(() => {
  const scope = globalThis as unknown as { OfflineAudioContext?: unknown };
  scope.OfflineAudioContext = OfflineAudioContext;
});

function withEq(presetId: string, overrides: Partial<{ preamp: number; bands: number[] }> = {}) {
  const preset = PRESETS.find((item) => item.id === presetId)!;
  return clampAudioSettings({
    eq: {
      enabled: true,
      preset: presetId,
      bands: overrides.bands ?? [...preset.bands],
      preamp: overrides.preamp ?? preset.preamp,
    },
  });
}

describe("the EQ through real audio", () => {
  it("is transparent when it is switched off", async () => {
    const measurement = await measureEq(defaultAudioSettings());
    measurement.bands.forEach((band) => {
      expect(Math.abs(band.db)).toBeLessThan(0.6);
    });
  });

  // The preamp is a global offset, so the shape of a preset is the difference
  // between its bands, not the absolute level of any one of them.
  it("moves the bass and leaves the treble alone on Bass Boost", async () => {
    const measurement = await measureEq(withEq("bass-boost"));
    expect(dbAt(measurement, 62) - dbAt(measurement, 8000)).toBeGreaterThan(4);
    expect(dbAt(measurement, 62)).toBeGreaterThan(0);
  });

  it("moves the treble and leaves the bass alone on Treble Boost", async () => {
    const measurement = await measureEq(withEq("treble-boost"));
    expect(dbAt(measurement, 8000) - dbAt(measurement, 62)).toBeGreaterThan(4);
    expect(dbAt(measurement, 8000)).toBeGreaterThan(0);
  });

  it("moves a band more than its neighbours two octaves away", async () => {
    const bands = new Array(10).fill(0);
    bands[4] = 9; // 500 Hz
    const measurement = await measureEq(withEq("custom", { bands, preamp: 0 }));
    const centre = Math.abs(dbAt(measurement, 500));
    const below = Math.abs(dbAt(measurement, 125));
    const above = Math.abs(dbAt(measurement, 2000));
    expect(centre).toBeGreaterThan(6);
    expect(centre).toBeGreaterThan(below * 2);
    expect(centre).toBeGreaterThan(above * 2);
  });

  it("cuts as well as boosts", async () => {
    const bands = new Array(10).fill(0);
    bands[0] = -10;
    const measurement = await measureEq(withEq("custom", { bands, preamp: 0 }));
    expect(dbAt(measurement, 31)).toBeLessThan(-5);
  });

  it("keeps the loudest preset inside full scale", async () => {
    const measurement = await measureEq(withEq("deep-bass"));
    expect(measurement.peak).toBeLessThanOrEqual(1);
  });

  it("scales with the preamp", async () => {
    const quiet = await measureEq(withEq("flat", { preamp: -6 }));
    const loud = await measureEq(withEq("flat", { preamp: 0 }));
    const quietLevel = dbAt(quiet, 1000);
    expect(quietLevel).toBeLessThan(-3);
    expect(dbAt(loud, 1000)).toBeGreaterThan(quietLevel + 3);
  });
});
