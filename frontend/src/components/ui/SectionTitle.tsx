import type { ReactNode } from "react";

interface SectionTitleProps {
  title: string;
  hint?: string;
  action?: ReactNode;
}

/**
 * Section heading at the 20px / 500 step, with a 13px supporting line. It sits
 * below the 40px view title, so every screen has two heading levels instead of
 * a long tail of 12px labels.
 */
export function SectionTitle({ title, hint, action }: SectionTitleProps) {
  return (
    <div className="mb-4 flex items-end justify-between gap-4">
      <div className="min-w-0">
        <h2 className="truncate text-[20px] font-medium leading-[1.25] tracking-[-0.02em] text-ink">{title}</h2>
        {hint ? <p className="mt-1 text-[13px] leading-[1.5] text-muted">{hint}</p> : null}
      </div>
      {action ? <div className="shrink-0 pb-0.5">{action}</div> : null}
    </div>
  );
}
