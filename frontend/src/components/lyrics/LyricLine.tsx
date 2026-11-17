import { memo } from "react";
import { cn } from "../../lib/cn";
import { wordProgress } from "../../lib/timeline";
import type { Line } from "../../lib/types";

interface LyricLineProps {
  line: Line;
  active: boolean;
  past: boolean;
  /** Playback time, passed only to the active line so idle lines stay static. */
  timeMs?: number;
  showIPA: boolean;
  showTranslation: boolean;
  immersive?: boolean;
  onSeek: (ms: number) => void;
  onWord: (word: string, line: Line) => void;
}

/**
 * One lyric line.
 *
 * When the document carries word timings each word is a fill gradient whose
 * progress is derived from the clock; otherwise the whole line is highlighted
 * for its span. Words are buttons so they can be reached with Tab and read by a
 * screen reader, and clicking one opens the dictionary.
 */
export const LyricLine = memo(function LyricLine({
  line,
  active,
  past,
  timeMs = 0,
  showIPA,
  showTranslation,
  immersive,
  onSeek,
  onWord,
}: LyricLineProps) {
  const hasWords = line.words.length > 0;
  const annotation = line.annotation;
  const sizeClass = immersive
    ? "text-[clamp(22px,2.6vw,34px)] leading-[1.35]"
    : active
      ? "text-[19px] leading-[1.4]"
      : "text-[15px] leading-[1.45]";

  return (
    <li
      className={cn(
        "group relative scroll-mt-24 py-2 transition-colors duration-200",
        active ? "opacity-100" : past ? "opacity-[0.55]" : "opacity-[0.75]",
      )}
      data-line-id={line.id}
    >
      <button
        type="button"
        onClick={() => onSeek(line.startMs)}
        aria-label={`Ir a: ${line.text}`}
        className="block w-full cursor-pointer text-left"
      >
        <span className={cn("block font-medium tracking-[-0.01em] transition-colors", sizeClass)}>
          {hasWords ? (
            <span className="flex flex-wrap items-baseline gap-x-[0.26em] gap-y-1">
              {line.words.map((word, index) => {
                const progress = active ? wordProgress(word, timeMs) : past ? 1 : 0;
                /* One span per word, filled with a clipped gradient: no duplicate
                   text nodes, so selection, screen readers and tests see the lyric
                   once. */
                const fillStyle = active
                  ? {
                      backgroundImage: `linear-gradient(to right, #f4f6fa ${progress * 100}%, #a7b3c4 ${progress * 100}%)`,
                      WebkitBackgroundClip: "text",
                      backgroundClip: "text",
                      color: "transparent",
                    }
                  : undefined;
                return (
                  <span
                    key={`${word.text}-${index}`}
                    role="button"
                    tabIndex={active ? 0 : -1}
                    aria-label={`Consultar ${cleanWord(word.text)}`}
                    onClick={(event) => {
                      event.stopPropagation();
                      onWord(cleanWord(word.text), line);
                    }}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        event.stopPropagation();
                        onWord(cleanWord(word.text), line);
                      }
                    }}
                    style={fillStyle}
                    className={cn(
                      "relative inline-block cursor-pointer rounded-[3px] outline-offset-4",
                      "hover:bg-white/5",
                      active ? "text-ink" : "text-muted",
                    )}
                  >
                    {word.text}
                  </span>
                );
              })}
            </span>
          ) : (
            <span className={active ? "text-ink" : "text-muted"}>{line.text || "\u00a0"}</span>
          )}
        </span>
      </button>

      {active && annotation ? (
        <div className="mt-1.5 space-y-1 pl-0.5">
          {showIPA ? (
            <p className="text-[13px] font-normal text-violet/90" lang="en-Latn-fonipa">
              {annotation.ipa}
            </p>
          ) : null}
          {showIPA ? (
            <p className="text-[13px] font-normal text-accent/85">{annotation.spanishApprox}</p>
          ) : null}
          {showTranslation ? (
            <p className="text-[12.5px] text-muted">{annotation.translation}</p>
          ) : null}
          {annotation.notes.length > 0 ? (
            <ul className="space-y-1 pt-0.5">
              {annotation.notes.map((note) => (
                <li key={note.text} className="text-[11.5px] leading-relaxed text-faint">
                  {note.text}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}
    </li>
  );
});

/** Strips punctuation so a clicked token is a dictionary word. */
export function cleanWord(raw: string): string {
  return raw.replace(/^[^\p{L}']+|[^\p{L}']+$/gu, "");
}
