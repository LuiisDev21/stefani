import { useCallback, useMemo, useRef } from "react";
import { cn } from "../../lib/cn";

interface WaveProgressProps {
  valueMs: number;
  totalMs: number;
  onSeek: (ms: number) => void;
  disabled?: boolean;
  seed?: string;
  tone?: "app" | "dark";
  className?: string;
}

/** Deterministic bar heights: same song, same silhouette, no audio analysis. */
function silhouette(seed: string, count: number): number[] {
  let hash = 2166136261;
  for (let i = 0; i < seed.length; i += 1) {
    hash ^= seed.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  const out: number[] = [];
  for (let i = 0; i < count; i += 1) {
    hash ^= i * 2654435761;
    hash = Math.imul(hash, 16777619);
    const unit = ((hash >>> 0) % 1000) / 1000;
    out.push(0.28 + unit * 0.72);
  }
  return out;
}

/**
 * Waveform progress bar. It is a slider underneath: pointer, keyboard and
 * screen readers keep working exactly as with the old range input.
 */
export function WaveProgress({
  valueMs,
  totalMs,
  onSeek,
  disabled,
  seed = "",
  tone = "app",
  className,
}: WaveProgressProps) {
  const ref = useRef<HTMLDivElement>(null);
  const bars = useMemo(() => silhouette(`${seed}|${Math.round(totalMs / 1000)}`, 96), [seed, totalMs]);
  const ratio = totalMs > 0 ? Math.min(1, Math.max(0, valueMs / totalMs)) : 0;

  const seekAt = useCallback(
    (clientX: number) => {
      const node = ref.current;
      if (!node || disabled || totalMs <= 0) return;
      const rect = node.getBoundingClientRect();
      const pct = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
      onSeek(Math.round(pct * totalMs));
    },
    [disabled, onSeek, totalMs],
  );

  return (
    <div
      ref={ref}
      role="slider"
      tabIndex={disabled ? -1 : 0}
      aria-label="Progreso de la canción"
      aria-valuemin={0}
      aria-valuemax={Math.round(totalMs)}
      aria-valuenow={Math.round(valueMs)}
      aria-disabled={disabled || undefined}
      className={cn(
        "group/wave flex h-8 items-end gap-[2px] rounded-[3px] outline-none",
        disabled ? "cursor-default opacity-50" : "cursor-pointer",
        "focus-visible:ring-2 focus-visible:ring-accent/70",
        className,
      )}
      onPointerDown={(event) => {
        if (disabled) return;
        event.currentTarget.setPointerCapture(event.pointerId);
        seekAt(event.clientX);
      }}
      onPointerMove={(event) => {
        if (disabled || !event.currentTarget.hasPointerCapture(event.pointerId)) return;
        seekAt(event.clientX);
      }}
      onKeyDown={(event) => {
        if (disabled) return;
        const step = 5000;
        if (event.key === "ArrowLeft") {
          event.preventDefault();
          onSeek(Math.max(0, valueMs - step));
        } else if (event.key === "ArrowRight") {
          event.preventDefault();
          onSeek(Math.min(totalMs, valueMs + step));
        } else if (event.key === "Home") {
          event.preventDefault();
          onSeek(0);
        } else if (event.key === "End") {
          event.preventDefault();
          onSeek(totalMs);
        }
      }}
    >
      {bars.map((height, index) => {
        const played = (index + 0.5) / bars.length <= ratio;
        return (
          <span
            key={index}
            aria-hidden="true"
            className={cn(
              "flex-1 rounded-[1px] transition-colors duration-150",
              tone === "dark"
                ? played
                  ? "bg-white/90"
                  : "bg-white/20 group-hover/wave:bg-white/30"
                : played
                  ? "bg-ink"
                  : "bg-line group-hover/wave:bg-line-strong",
            )}
            style={{ height: `${Math.round(height * 100)}%` }}
          />
        );
      })}
    </div>
  );
}
