import type { ReactNode } from "react";

interface TopBarProps {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}

/**
 * Section header: the display headline at 34px with the system's negative
 * tracking, one supporting line at 14px, actions on the right as pills.
 */
export function TopBar({ title, subtitle, actions }: TopBarProps) {
  return (
    <header className="flex items-end justify-between gap-6 px-8 pb-5 pt-8">
      <div className="min-w-0">
        <h1 className="truncate text-[34px] font-semibold leading-[1.1] tracking-[-0.374px] text-ink">{title}</h1>
        {subtitle ? <p className="mt-1.5 truncate text-[14px] leading-[1.43] text-muted">{subtitle}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
    </header>
  );
}
