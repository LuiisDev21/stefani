import type { ReactNode } from "react";

interface SectionTitleProps {
  title: string;
  hint?: string;
  action?: ReactNode;
}

/**
 * Section heading at the tagline step of the type ladder (21px / 600), with a
 * 14px supporting line. It sits below the 34px view title, so every screen has
 * two heading levels instead of a long tail of 12px labels.
 */
export function SectionTitle({ title, hint, action }: SectionTitleProps) {
  return (
    <div className="mb-4 flex items-end justify-between gap-4">
      <div className="min-w-0">
        <h2 className="truncate text-[21px] font-semibold leading-[1.19] tracking-[-0.231px] text-ink">{title}</h2>
        {hint ? <p className="mt-1 text-[14px] leading-[1.43] text-muted">{hint}</p> : null}
      </div>
      {action ? <div className="shrink-0 pb-0.5">{action}</div> : null}
    </div>
  );
}
