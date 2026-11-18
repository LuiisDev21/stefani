import type { ReactNode } from "react";

interface QuickCardProps {
  icon: ReactNode;
  label: string;
  detail: string;
  onClick: () => void;
}

/**
 * One-line shortcut into a section, with its real count underneath. The icon
 * sits in its own tile so a row of them reads as a launcher, not as a menu.
 */
export function QuickCard({ icon, label, detail, onClick }: QuickCardProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="group flex w-full items-center gap-3 rounded-[10px] border border-line bg-elevated p-3 text-left transition-colors hover:bg-card"
    >
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-[8px] bg-card text-body transition-colors group-hover:text-ink">
        {icon}
      </span>
      <span className="min-w-0">
        <span className="block truncate text-[14px] text-ink">{label}</span>
        <span className="block truncate text-[12px] text-ash tnum">{detail}</span>
      </span>
    </button>
  );
}
