import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description: string;
  actions?: ReactNode;
  className?: string;
  compact?: boolean;
}

/** Empty state: says what is missing and what the user can do next. */
export function EmptyState({ icon, title, description, actions, className, compact }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-start gap-3 rounded-[12px] border border-dashed border-line/80 bg-elevated/40",
        compact ? "px-4 py-4" : "px-6 py-8",
        className,
      )}
    >
      {icon ? <div className="text-faint">{icon}</div> : null}
      <div className="space-y-1">
        <p className={cn("font-medium text-ink", compact ? "text-[13px]" : "text-[15px]")}>{title}</p>
        <p className="max-w-[52ch] text-[12.5px] leading-relaxed text-muted">{description}</p>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2 pt-1">{actions}</div> : null}
    </div>
  );
}
