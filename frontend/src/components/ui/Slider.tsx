import { useCallback, useRef } from "react";
import { cn } from "../../lib/cn";

interface SliderProps {
  value: number;
  max: number;
  onChange: (value: number) => void;
  label: string;
  className?: string;
  /** Formats the value for the accessible aria-valuetext. */
  format?: (value: number) => string;
  disabled?: boolean;
}

/**
 * Range input styled as the signal line: a 2px rail that thickens on hover.
 * Native input semantics keep keyboard support (arrows, Home/End) for free.
 */
export function Slider({ value, max, onChange, label, className, format, disabled }: SliderProps) {
  const ref = useRef<HTMLInputElement>(null);
  const safeMax = max > 0 ? max : 1;
  const percent = Math.min(100, Math.max(0, (value / safeMax) * 100));

  const handleKey = useCallback(
    (event: React.KeyboardEvent<HTMLInputElement>) => {
      if (event.key !== "ArrowRight" && event.key !== "ArrowLeft") return;
      event.preventDefault();
      const step = event.shiftKey ? safeMax / 20 : safeMax / 200;
      onChange(Math.min(safeMax, Math.max(0, value + (event.key === "ArrowRight" ? step : -step))));
    },
    [onChange, safeMax, value],
  );

  return (
    <input
      ref={ref}
      type="range"
      min={0}
      max={safeMax}
      step={1}
      value={value}
      disabled={disabled}
      aria-label={label}
      aria-valuetext={format ? format(value) : undefined}
      onKeyDown={handleKey}
      onChange={(event) => onChange(Number(event.target.value))}
      className={cn("stefani-slider", className)}
      style={{ ["--fill" as string]: `${percent}%` }}
    />
  );
}
