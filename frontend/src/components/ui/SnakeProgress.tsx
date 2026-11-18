import { useCallback, useMemo, useRef } from "react";
import { cn } from "../../lib/cn";

interface SnakeProgressProps {
  valueMs: number;
  totalMs: number;
  onSeek: (ms: number) => void;
  disabled?: boolean;
  seed?: string;
  tone?: "app" | "dark";
  className?: string;
}

const WIDTH = 1000;
const HEIGHT = 36;
const SAMPLES = 110;

/**
 * One deterministic silhouette per track: a line that snakes across the bar.
 * It reads as a hand-drawn seismograph instead of an equaliser, and it costs a
 * single path, so the progress only moves `stroke-dashoffset`.
 */
function buildSnake(seed: string): { d: string; ys: number[] } {
  let state = 2166136261;
  for (let i = 0; i < seed.length; i += 1) {
    state ^= seed.charCodeAt(i);
    state = Math.imul(state, 16777619);
  }
  const random = () => {
    state ^= state << 13;
    state ^= state >>> 17;
    state ^= state << 5;
    return (state >>> 0) / 4294967296;
  };

  const phase = random() * Math.PI * 2;
  const drift = 0.42 + random() * 0.22;
  const ys: number[] = [];
  for (let i = 0; i <= SAMPLES; i += 1) {
    const tooth = i % 2 === 0 ? 1 : -1;
    const energy = 0.35 + random() * 0.65;
    const wobble = Math.sin(i * drift + phase) * (HEIGHT * 0.16);
    const y = HEIGHT / 2 + wobble + tooth * energy * (HEIGHT * 0.26);
    ys.push(Math.round(y * 10) / 10);
  }
  const step = WIDTH / SAMPLES;
  const d = ys
    .map((y, i) => `${i === 0 ? "M" : "L"} ${Math.round(i * step * 10) / 10} ${y}`)
    .join(" ");
  return { d, ys };
}

/**
 * Snake progress bar. Under the drawing it is still a slider: pointer, keyboard
 * and screen readers behave exactly like with a range input.
 */
export function SnakeProgress({
  valueMs,
  totalMs,
  onSeek,
  disabled,
  seed = "",
  tone = "app",
  className,
}: SnakeProgressProps) {
  const ref = useRef<SVGSVGElement>(null);
  const snake = useMemo(() => buildSnake(seed || "stefani"), [seed]);
  const ratio = totalMs > 0 ? Math.min(1, Math.max(0, valueMs / totalMs)) : 0;
  const headIndex = Math.round(ratio * SAMPLES);
  const headX = (headIndex / SAMPLES) * WIDTH;
  const headY = snake.ys[headIndex] ?? HEIGHT / 2;

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

  const track = tone === "dark" ? "rgba(255,255,255,0.28)" : "#30343a";
  const fill = tone === "dark" ? "rgba(255,255,255,0.92)" : "#f4f4f6";

  return (
    <svg
      ref={ref}
      viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
      preserveAspectRatio="none"
      role="slider"
      tabIndex={disabled ? -1 : 0}
      aria-label="Progreso de la canción"
      aria-valuemin={0}
      aria-valuemax={Math.round(totalMs)}
      aria-valuenow={Math.round(valueMs)}
      aria-disabled={disabled || undefined}
      className={cn(
        "h-8 w-full touch-none rounded-[4px] outline-none",
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
      <path
        d={snake.d}
        fill="none"
        stroke={track}
        strokeWidth={2.4}
        strokeLinecap="round"
        strokeLinejoin="round"
        vectorEffect="non-scaling-stroke"
      />
      <path
        d={snake.d}
        fill="none"
        stroke={fill}
        strokeWidth={2.4}
        strokeLinecap="round"
        strokeLinejoin="round"
        vectorEffect="non-scaling-stroke"
        pathLength={1}
        strokeDasharray="1"
        strokeDashoffset={1 - ratio}
      />
      <circle cx={headX} cy={headY} r={3.2} fill={fill} />
    </svg>
  );
}
