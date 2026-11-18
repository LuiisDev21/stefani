import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { cn } from "../../lib/cn";

interface WavyProgressProps {
  valueMs: number;
  totalMs: number;
  onSeek: (ms: number) => void;
  disabled?: boolean;
  seed?: string;
  tone?: "app" | "dark";
  className?: string;
}

/*
 * Material 3 wavy linear progress indicator, rebuilt on SVG.
 *
 * Tokens taken from LinearProgressIndicatorTokens: 4dp stroke, 3dp amplitude
 * (centre to peak), 40dp wavelength, 4dp stop indicator, 4dp gap. The container
 * here is 26px tall instead of 10dp, so the wave keeps the same proportions with
 * a stroke of 3.5px and an amplitude of 5px. The wave travels one wavelength per
 * second, which is what `waveSpeed = wavelength` means in the spec.
 */
const HEIGHT = 28;
const STROKE = 3.5;
const AMPLITUDE = 5;
const WAVELENGTH = 44;
const GAP = 7;
const STOP = 3.4;

/** Sine samples across the full bar, extended by one wavelength on each side. */
function wavePath(width: number): string {
  const center = HEIGHT / 2;
  const start = -WAVELENGTH;
  const end = width + WAVELENGTH;
  const step = 1.6;
  const points: string[] = [];
  for (let x = start; x <= end; x += step) {
    const y = center + AMPLITUDE * Math.sin((2 * Math.PI * x) / WAVELENGTH);
    points.push(`${points.length === 0 ? "M" : "L"} ${x.toFixed(1)} ${y.toFixed(2)}`);
  }
  return points.join(" ");
}

/**
 * Wavy progress bar. Under the drawing it is still a slider: pointer, keyboard
 * and screen readers behave exactly like with a range input.
 */
export function WavyProgress({
  valueMs,
  totalMs,
  onSeek,
  disabled,
  tone = "app",
  className,
}: WavyProgressProps) {
  const ref = useRef<SVGSVGElement>(null);
  // Two instances live at once (bar and immersive view): each needs its own clip.
  const clipId = useRef(`wavy-${Math.random().toString(36).slice(2, 9)}`).current;
  const [width, setWidth] = useState(0);
  const ratio = totalMs > 0 ? Math.min(1, Math.max(0, valueMs / totalMs)) : 0;

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry) setWidth(Math.round(entry.contentRect.width));
    });
    observer.observe(node);
    setWidth(Math.round(node.getBoundingClientRect().width));
    return () => observer.disconnect();
  }, []);

  const path = useMemo(() => wavePath(Math.max(width, 1)), [width]);
  // The wave flattens below 10 % and above 95 %, as in WavyProgressIndicatorDefaults.
  const amplitude = ratio <= 0.1 ? ratio / 0.1 : ratio >= 0.95 ? (1 - ratio) / 0.05 : 1;
  const activeWidth = ratio * width;

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

  const track = tone === "dark" ? "rgba(255,255,255,0.26)" : "#2c2f33";
  const fill = tone === "dark" ? "rgba(255,255,255,0.95)" : "#f4f4f6";
  const trackStart = activeWidth + GAP;
  const stopX = Math.max(STOP, width - STOP);

  return (
    <svg
      ref={ref}
      viewBox={`0 0 ${Math.max(width, 1)} ${HEIGHT}`}
      preserveAspectRatio="none"
      role="slider"
      tabIndex={disabled ? -1 : 0}
      aria-label="Progreso de la canción"
      aria-valuemin={0}
      aria-valuemax={Math.round(totalMs)}
      aria-valuenow={Math.round(valueMs)}
      aria-disabled={disabled || undefined}
      className={cn(
        "h-7 w-full touch-none rounded-[4px] outline-none",
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
      <defs>
        <clipPath id={clipId}>
          <rect x={0} y={0} width={Math.max(activeWidth, 0)} height={HEIGHT} />
        </clipPath>
      </defs>

      <line
        x1={trackStart}
        y1={HEIGHT / 2}
        x2={Math.max(trackStart, stopX - GAP)}
        y2={HEIGHT / 2}
        stroke={track}
        strokeWidth={STROKE}
        strokeLinecap="round"
      />

      <g clipPath={`url(#${clipId})`}>
        <g
          className={disabled ? undefined : "wavy-flow"}
          style={{ animationPlayState: disabled ? "paused" : "running" }}
        >
          <g transform={`translate(0 ${HEIGHT / 2}) scale(1 ${amplitude}) translate(0 ${-HEIGHT / 2})`}>
            <path
              d={path}
              fill="none"
              stroke={fill}
              strokeWidth={STROKE}
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </g>
        </g>
      </g>

      <circle cx={stopX} cy={HEIGHT / 2} r={STOP} fill={track} />
      {ratio > 0 ? <circle cx={activeWidth} cy={HEIGHT / 2} r={STOP} fill={fill} /> : null}
    </svg>
  );
}
