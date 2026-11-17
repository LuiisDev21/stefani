import type { ReactNode } from "react";

interface TopBarProps {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}

/** Section header: one title, one supporting line, actions on the right. */
export function TopBar({ title, subtitle, actions }: TopBarProps) {
  return (
    <header className="flex items-end justify-between gap-6 px-7 pb-3 pt-5">
      <div className="min-w-0">
        <h1 className="truncate text-[21px] font-medium leading-tight tracking-[-0.015em] text-ink">
          {title}
        </h1>
        {subtitle ? <p className="mt-1 truncate text-[12px] text-muted">{subtitle}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 items-center gap-2">{actions}</div> : null}
    </header>
  );
}
