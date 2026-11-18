import { describe, expect, it } from "vitest";
import { clamp, formatRelativeDate, formatTime } from "./format";

describe("formatTime", () => {
  it("formats under an hour as m:ss", () => {
    expect(formatTime(0)).toBe("0:00");
    expect(formatTime(41_000)).toBe("0:41");
    expect(formatTime(125_600)).toBe("2:05");
  });

  it("adds hours when needed", () => {
    expect(formatTime(3_600_000)).toBe("1:00:00");
    expect(formatTime(3_725_000)).toBe("1:02:05");
  });

  it("treats invalid input as zero", () => {
    expect(formatTime(-5)).toBe("0:00");
    expect(formatTime(Number.NaN)).toBe("0:00");
  });
});


describe("formatRelativeDate", () => {
  const now = Date.parse("2026-11-17T12:00:00Z");

  it("describes recent activity", () => {
    expect(formatRelativeDate(now - 30_000, now)).toBe("hace un momento");
    expect(formatRelativeDate(now - 5 * 60_000, now)).toBe("hace 5 min");
    expect(formatRelativeDate(now - 3 * 3_600_000, now)).toBe("hace 3 h");
    expect(formatRelativeDate(now - 2 * 86_400_000, now)).toBe("hace 2 dias");
  });

  it("handles missing timestamps", () => {
    expect(formatRelativeDate(0, now)).toBe("sin actividad");
  });
});

describe("clamp", () => {
  it("bounds values", () => {
    expect(clamp(5, 0, 10)).toBe(5);
    expect(clamp(-1, 0, 10)).toBe(0);
    expect(clamp(11, 0, 10)).toBe(10);
  });
});
