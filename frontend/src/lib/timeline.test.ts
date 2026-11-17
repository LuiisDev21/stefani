import { describe, expect, it } from "vitest";
import type { Line, Lyrics } from "./types";
import {
  activeLineIndex,
  activeWordIndex,
  applyOffset,
  loopFromLines,
  supportsKaraoke,
  wordProgress,
} from "./timeline";

function line(index: number, start: number, end: number, words = false): Line {
  return {
    id: `l${index}`,
    index,
    text: `line ${index}`,
    startMs: start,
    endMs: end,
    words: words
      ? [
          { text: "one", startMs: start, endMs: start + (end - start) / 2 },
          { text: "two", startMs: start + (end - start) / 2, endMs: end },
        ]
      : [],
    annotation: null,
  };
}

const lines = [line(0, 5000, 8000), line(1, 8000, 11000), line(2, 11000, 14000)];

describe("activeLineIndex", () => {
  it("returns -1 for an empty document", () => {
    expect(activeLineIndex([], 1000)).toBe(-1);
  });

  it("highlights the first line before it starts, so the intro is not blank", () => {
    expect(activeLineIndex(lines, 0)).toBe(0);
    expect(activeLineIndex(lines, 4999)).toBe(0);
  });

  it("follows the clock across lines and gaps", () => {
    expect(activeLineIndex(lines, 5000)).toBe(0);
    expect(activeLineIndex(lines, 7999)).toBe(0);
    expect(activeLineIndex(lines, 8000)).toBe(1);
    expect(activeLineIndex(lines, 12000)).toBe(2);
  });

  it("keeps the last line active after the song ends", () => {
    expect(activeLineIndex(lines, 999999)).toBe(2);
  });
});

describe("wordProgress", () => {
  const word = { text: "slowly", startMs: 1000, endMs: 2000 };

  it("is 0 before the word, 1 after it and proportional inside", () => {
    expect(wordProgress(word, 0)).toBe(0);
    expect(wordProgress(word, 1500)).toBeCloseTo(0.5);
    expect(wordProgress(word, 5000)).toBe(1);
  });

  it("never divides by zero for a zero-length word", () => {
    expect(wordProgress({ text: "x", startMs: 100, endMs: 100 }, 100)).toBe(1);
    expect(wordProgress({ text: "x", startMs: 100, endMs: 100 }, 50)).toBe(0);
  });
});

describe("activeWordIndex", () => {
  const words = [
    { text: "a", startMs: 0, endMs: 500 },
    { text: "b", startMs: 500, endMs: 1000 },
    { text: "c", startMs: 1000, endMs: 1500 },
  ];

  it("returns -1 without words", () => {
    expect(activeWordIndex([], 100)).toBe(-1);
  });

  it("points at the word being sung", () => {
    expect(activeWordIndex(words, 0)).toBe(0);
    expect(activeWordIndex(words, 750)).toBe(1);
    expect(activeWordIndex(words, 1450)).toBe(2);
  });

  it("holds the last word after the line ends", () => {
    expect(activeWordIndex(words, 99999)).toBe(2);
  });
});

describe("loopFromLines", () => {
  it("builds inclusive bounds with a small pad", () => {
    const bounds = loopFromLines(lines, 1, 1);
    expect(bounds).not.toBeNull();
    expect(bounds!.startMs).toBe(7880);
    expect(bounds!.endMs).toBe(11120);
    expect(bounds!.label).toBe("line 1");
  });

  it("orders the selection when the user picks backwards", () => {
    const bounds = loopFromLines(lines, 2, 0);
    expect(bounds!.startMs).toBe(4880);
    expect(bounds!.endMs).toBe(14120);
  });

  it("returns null without lines", () => {
    expect(loopFromLines([], 0, 1)).toBeNull();
  });
});

describe("supportsKaraoke", () => {
  const base: Lyrics = {
    trackId: "t",
    available: true,
    provider: "demo",
    sync: "WORD",
    source: "s",
    license: "l",
    hasWords: true,
    lines: [],
    notice: "",
  };

  it("requires an available document with word timings", () => {
    expect(supportsKaraoke(base)).toBe(true);
    expect(supportsKaraoke({ ...base, hasWords: false })).toBe(false);
    expect(supportsKaraoke({ ...base, available: false })).toBe(false);
    expect(supportsKaraoke(null)).toBe(false);
  });
});

describe("applyOffset", () => {
  it("clamps at zero and applies the shift", () => {
    expect(applyOffset(1000, 200)).toBe(1200);
    expect(applyOffset(100, -300)).toBe(0);
  });
});
