import { cn } from "../../lib/cn";

/** Small pill toggle shared by the study panel and the immersive view. */
export function Chip({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "rounded-full border px-2.5 py-[3px] text-[11.5px] transition-colors",
        active
          ? "border-accent bg-accent/6 text-accent"
          : "border-line bg-transparent text-muted hover:bg-elevated hover:text-ink",
      )}
    >
      {label}
    </button>
  );
}
