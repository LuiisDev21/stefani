import { useCallback, useEffect, useRef, useState } from "react";
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
 * (centre to peak), 40dp wavelength. The container here is 28px tall instead of
 * 10dp, so the wave keeps the same proportions with a stroke of 3.5px and an
 * amplitude of 5px. The wave travels one wavelength per second, which is what
 * `waveSpeed = wavelength` means in the spec.
 *
 * The playhead is a dot that stays on the centre line, as in the reference: the
 * wave travels underneath and tapers into the dot, so the stroke never ends
 * away from it and the dot never bobs up and down.
 */
const HEIGHT = 28;
const STROKE = 3.5;
const TRACK_STROKE = 2;
const AMPLITUDE = 5;
const WAVELENGTH = 44;
const PLAYHEAD = 3.4;
const STOP = 3.2;

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
  const waveRef = useRef<SVGPathElement>(null);
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

  // Read inside the loop without restarting it on every progress tick.
  const live = useRef({ activeWidth, amplitude, width, disabled });
  live.current = { activeWidth, amplitude, width, disabled };

  useEffect(() => {
    const speed = WAVELENGTH / 1100; // one wavelength per second
    const started = performance.now();
    let frame = 0;
    let painted = 0;
    const tick = (now: number) => {
      frame = requestAnimationFrame(tick);
      // ~24 fps is plenty for a wave travelling 44 px per second: rebuilding
      // the path at 60 costs more than the motion is worth.
      if (now - painted < 41) return;
      painted = now;
      const state = live.current;
      const offset = state.disabled ? 0 : ((now - started) * speed) % WAVELENGTH;
      const edge = state.activeWidth;
      const scale = state.amplitude;
      const center = HEIGHT / 2;
      const step = 4;
      const parts: string[] = [];
      for (let x = 0; x <= Math.max(edge, 0.1); x += step) {
        const taper = Math.min(1, Math.max(0, (edge - x) / 16));
        const y = center + scale * AMPLITUDE * taper * Math.sin((2 * Math.PI * (x - offset)) / WAVELENGTH);
        parts.push(`${parts.length === 0 ? "M" : "L"}${x.toFixed(1)} ${y.toFixed(2)}`);
      }
      parts.push(`L${Math.max(edge, 0.1).toFixed(1)} ${center}`);
      waveRef.current?.setAttribute("d", parts.join(" "));
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, []);

  const track = tone === "dark" ? "rgba(255,255,255,0.26)" : "#2c2f33";
  const fill = tone === "dark" ? "rgba(255,255,255,0.95)" : "#f4f4f6";
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
      <line
        x1={activeWidth}
        y1={HEIGHT / 2}
        x2={Math.max(activeWidth, stopX - 6)}
        y2={HEIGHT / 2}
        stroke={track}
        strokeWidth={TRACK_STROKE}
        strokeLinecap="round"
      />

      <path
        ref={waveRef}
        fill="none"
        stroke={fill}
        strokeWidth={STROKE}
        strokeLinecap="round"
        strokeLinejoin="round"
      />

      <circle cx={stopX} cy={HEIGHT / 2} r={STOP} fill={track} />
      {ratio > 0 ? <circle cx={activeWidth} cy={HEIGHT / 2} r={PLAYHEAD} fill={fill} /> : null}
    </svg>
  );
}
