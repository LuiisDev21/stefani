import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

/**
 * Small pill toggle shared by the study panel and the immersive view. With an
 * icon it collapses to a round button; the label still names it for tools and
 * screen readers.
 */
export function Chip({
  active,
  onClick,
  label,
  icon,
}: {
  active: boolean;
  onClick: () => void;
  label: string;
  icon?: ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      aria-label={icon ? label : undefined}
      title={icon ? label : undefined}
      className={cn(
        "inline-flex items-center justify-center gap-1.5 rounded-full border transition-colors",
        icon ? "h-6 w-6" : "px-2.5 py-[3px] text-[11.5px]",
        active
          ? "border-accent bg-accent/6 text-accent"
          : "border-line bg-transparent text-muted hover:bg-elevated hover:text-ink",
      )}
    >
      {icon ?? label}
    </button>
  );
}
