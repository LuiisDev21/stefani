import { memo } from "react";
import { cn } from "../../lib/cn";
import { wordProgress } from "../../lib/timeline";
import type { Line } from "../../lib/types";

interface LyricLineProps {
  line: Line;
  active: boolean;
  past: boolean;
  /** The line right before or after the active one, which also shows its gloss. */
  near?: boolean;
  /** Playback time, passed only to the active line so idle lines stay static. */
  timeMs?: number;
  showIPA: boolean;
  showTranslation: boolean;
  immersive?: boolean;
  onSeek: (ms: number) => void;
}

/**
 * One lyric line.
 *
 * When the document carries word timings each word is filled by a hard two-stop
 * gradient clipped to the text: that is the only way to show progress inside a
 * single text node, so selection and screen readers still see the lyric once.
 */
export const LyricLine = memo(function LyricLine({
  line,
  active,
  past,
  near,
  timeMs = 0,
  showIPA,
  showTranslation,
  immersive,
  onSeek,
}: LyricLineProps) {
  const words = line.words ?? [];
  const hasWords = words.length > 0;
  const annotation = line.annotation;
  const sizeClass = immersive
    ? "text-[clamp(20px,2.05vw,29px)] font-semibold leading-[1.34] tracking-[-0.3px]"
    : active
      ? "text-[21px] leading-[1.35] font-semibold tracking-[-0.3px]"
      : "text-[16px] leading-[1.45]";

  return (
    <li
      className={cn(
        "group relative scroll-mt-24 py-2.5 transition-[opacity,transform] duration-[420ms] ease-out",
        active ? "translate-y-0 opacity-100" : past ? "-translate-y-[1px] opacity-55" : "translate-y-[2px] opacity-80",
      )}
      data-line-id={line.id}
    >
      <button
        type="button"
        onClick={() => onSeek(line.startMs)}
        aria-label={`Ir a: ${line.text}`}
        className="block w-full cursor-pointer text-left"
      >
        <span className={cn("block transition-colors duration-300", sizeClass)}>
          {hasWords ? (
            <span className="flex flex-wrap items-baseline gap-x-[0.26em] gap-y-1">
              {words.map((word, index) => {
                const progress = active ? wordProgress(word, timeMs) : past ? 1 : 0;
                /* One span per word: a hard gradient fills it left to right
                   without duplicating text nodes. */
                // Immersive follows the reference: the sung line is amber and
                // the upcoming words stay grey inside the same sentence.
                const base = immersive ? "#ffc000" : "#1d1d1f";
                const rest = immersive ? "#8a8a8e" : "#a1a1a6";
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
                    style={fillStyle}
                    className={cn(
                      "relative inline-block",
                      active ? (immersive ? "text-[#ffc000]" : "text-ink") : immersive ? "text-white/65" : "text-body",
                    )}
                  >
                    {word.text}
                  </span>
                );
              })}
            </span>
          ) : (
            <span className={active ? (immersive ? "text-[#ffc000]" : "text-ink") : immersive ? "text-white/65" : "text-body"}>
              {line.text || "\u00a0"}
            </span>
          )}
        </span>
      </button>

      {annotation && (active || near) ? (
        <div
          className={cn(
            "mt-2 space-y-1 transition-opacity duration-300",
            active ? "opacity-100" : "opacity-60",
          )}
        >
          {showIPA && annotation.spanishApprox ? (
            <p className={immersive ? "text-[14px] text-white/80" : "text-[14px] text-ink/75"}>
              {annotation.spanishApprox}
            </p>
          ) : null}
          {showTranslation ? (
            <p className={immersive ? "text-[13px] leading-[1.45] text-white/45" : "text-[14px] leading-[1.43] text-muted"}>
              {annotation.translation}
            </p>
          ) : null}
          {(annotation.notes ?? []).length > 0 ? (
            <ul className="space-y-1 pt-0.5">
              {(annotation.notes ?? []).map((note) => (
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
