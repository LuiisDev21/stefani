import type { ReactNode } from "react";

interface SectionTitleProps {
  title: string;
  hint?: string;
  action?: ReactNode;
}

/** Section heading: one label, optional hint, optional action. */
export function SectionTitle({ title, hint, action }: SectionTitleProps) {
  return (
    <div className="mb-2 flex items-end justify-between gap-4">
      <div>
        <h2 className="text-[12.5px] font-semibold tracking-[-0.005em] text-ink/90">{title}</h2>
        {hint ? <p className="mt-1 text-[11.5px] text-muted">{hint}</p> : null}
      </div>
      {action}
    </div>
  );
}
