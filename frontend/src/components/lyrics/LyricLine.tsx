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
 * When the document carries word timings each word is filled by a hard two-stop
 * gradient clipped to the text: that is the only way to show progress inside a
 * single text node, so selection and screen readers still see the lyric once.
 * Words are buttons so they can be reached with Tab and opened in the
 * dictionary.
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
    ? "text-[clamp(24px,3vw,40px)] leading-[1.3]"
    : active
      ? "text-[21px] leading-[1.35] font-semibold tracking-[-0.3px]"
      : "text-[16px] leading-[1.45]";

  return (
    <li
      className={cn(
        "group relative scroll-mt-24 py-2.5 transition-opacity duration-200",
        active ? "opacity-100" : past ? "opacity-55" : "opacity-80",
      )}
      data-line-id={line.id}
    >
      <button
        type="button"
        onClick={() => onSeek(line.startMs)}
        aria-label={`Ir a: ${line.text}`}
        className="block w-full cursor-pointer text-left"
      >
        <span className={cn("block transition-colors", sizeClass)}>
          {hasWords ? (
            <span className="flex flex-wrap items-baseline gap-x-[0.26em] gap-y-1">
              {line.words.map((word, index) => {
                const progress = active ? wordProgress(word, timeMs) : past ? 1 : 0;
                /* One span per word: a hard gradient fills it left to right
                   without duplicating text nodes. */
                const base = immersive ? "#ffffff" : "#1d1d1f";
                const rest = immersive ? "#6e6e73" : "#a1a1a6";
                const fillStyle = active
                  ? {
                      backgroundImage: `linear-gradient(to right, ${base} ${progress * 100}%, ${rest} ${progress * 100}%)`,
                      WebkitBackgroundClip: "text",
                      backgroundClip: "text",
                      WebkitTextFillColor: "transparent",
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
                      "relative inline-block cursor-pointer rounded-[4px] outline-offset-4 transition-colors",
                      immersive ? "hover:bg-white/10" : "hover:bg-accent/8",
                      active ? (immersive ? "text-white" : "text-ink") : immersive ? "text-white/55" : "text-muted",
                    )}
                  >
                    {word.text}
                  </span>
                );
              })}
            </span>
          ) : (
            <span className={active ? (immersive ? "text-white" : "text-ink") : immersive ? "text-white/55" : "text-muted"}>
              {line.text || "\u00a0"}
            </span>
          )}
        </span>
      </button>

      {active && annotation ? (
        <div className="mt-2 space-y-1">
          {showIPA ? (
            <p className={immersive ? "text-[14px] text-white/55" : "text-[14px] text-muted"} lang="en-Latn-fonipa">
              {annotation.ipa}
            </p>
          ) : null}
          {showIPA ? (
            <p className={immersive ? "text-[14px] text-white/80" : "text-[14px] text-ink/75"}>
              {annotation.spanishApprox}
            </p>
          ) : null}
          {showTranslation ? (
            <p className={immersive ? "text-[14px] leading-[1.43] text-white/55" : "text-[14px] leading-[1.43] text-muted"}>
              {annotation.translation}
            </p>
          ) : null}
          {annotation.notes.length > 0 ? (
            <ul className="space-y-1 pt-0.5">
              {annotation.notes.map((note) => (
                <li
                  key={note.text}
                  className={immersive ? "text-[12.5px] leading-[1.45] text-white/40" : "text-[12.5px] leading-[1.45] text-faint"}
                >
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
