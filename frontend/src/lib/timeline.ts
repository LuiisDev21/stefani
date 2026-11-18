import type { Line, Lyrics, WordTiming } from "./types";

/**
 * Pure timeline helpers.
 *
 * One rule governs this file: the player clock is the only source of truth. The
 * lyric view derives everything (active line, active word, fill progress) from
 * a timestamp; it never keeps its own timer.
 */

/**
 * Returns the index of the line that should be highlighted at `timeMs`.
 * A line is active from its start until the next line starts, so gaps between
 * verses do not flicker back to the previous line. Before the first line the
 * first one is highlighted: the panel should show its pronunciation while the
 * intro plays instead of looking empty.
 */
export function activeLineIndex(lines: Line[], timeMs: number): number {
  if (lines.length === 0) return -1;
  if (timeMs < lines[0].startMs) return 0;
  let low = 0;
  let high = lines.length - 1;
  let found = -1;
  while (low <= high) {
    const mid = (low + high) >> 1;
    const line = lines[mid];
    if (line.startMs <= timeMs) {
      found = mid;
      low = mid + 1;
    } else {
      high = mid - 1;
    }
  }
  return found;
}

/** Progress inside a word, clamped to 0..1. */
export function wordProgress(word: WordTiming, timeMs: number): number {
  const span = word.endMs - word.startMs;
  if (span <= 0) return timeMs >= word.startMs ? 1 : 0;
  return Math.min(1, Math.max(0, (timeMs - word.startMs) / span));
}

/** Index of the word being sung, or the last finished word during a gap. */
export function activeWordIndex(words: WordTiming[], timeMs: number): number {
  if (words.length === 0) return -1;
  for (let i = 0; i < words.length; i += 1) {
    if (timeMs < words[i].endMs) {
      return timeMs >= words[i].startMs ? i : Math.max(0, i - 1);
    }
  }
  return words.length - 1;
}


/**
 * Offset (in ms) between the audio clock and the lyric document. Positive values
 * mean the lyrics should be shown earlier; the seek bar and A-B loop always use
 * raw audio time.
 */
export function applyOffset(timeMs: number, offsetMs: number): number {
  return Math.max(0, timeMs + offsetMs);
}

/** Reports whether a lyric document can drive word-level highlighting. */
export function supportsKaraoke(lyrics: Lyrics | null): boolean {
  return Boolean(lyrics && lyrics.available && lyrics.hasWords);
}
