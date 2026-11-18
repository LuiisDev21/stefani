import type { ReactNode } from "react";

interface TopBarProps {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}

/**
 * View header: the display headline at 40px with the system's tight tracking,
 * one supporting line at 14px, actions on the right.
 */
export function TopBar({ title, subtitle, actions }: TopBarProps) {
  return (
    <header className="flex items-end justify-between gap-6 px-8 pb-6 pt-9">
      <div className="min-w-0">
        <h1 className="truncate text-[40px] font-semibold leading-[1.05] tracking-[-0.035em] text-ink">{title}</h1>
        {subtitle ? <p className="mt-2 truncate text-[14px] leading-[1.5] text-muted">{subtitle}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
    </header>
  );
}
