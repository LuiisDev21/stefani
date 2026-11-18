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
        "flex flex-col items-start gap-3 rounded-[16px] border border-line bg-elevated",
        compact ? "px-5 py-5" : "px-7 py-9",
        className,
      )}
    >
      {icon ? <div className="text-faint">{icon}</div> : null}
      <div className="space-y-1.5">
        <p className={cn("font-semibold text-ink", compact ? "text-[15px]" : "text-[19px] tracking-[-0.2px]")}>
          {title}
        </p>
        <p className="max-w-[52ch] text-[14px] leading-[1.43] text-muted">{description}</p>
      </div>
      {actions ? <div className="flex flex-wrap items-center gap-2 pt-1">{actions}</div> : null}
    </div>
  );
}
