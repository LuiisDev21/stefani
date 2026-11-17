import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "../../lib/cn";

type Variant = "primary" | "subtle" | "ghost" | "danger";
type Size = "sm" | "md";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  icon?: ReactNode;
}

const base =
  "inline-flex items-center justify-center gap-2 rounded-[8px] font-medium transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-45 select-none";

const variants: Record<Variant, string> = {
  // The only saturated surface in the palette: reserved for the one primary
  // action on a screen.
  primary: "bg-accent text-[#0b0e16] hover:bg-[#9dc0ff] active:bg-[#74a1f5]",
  subtle: "bg-surface text-ink hover:bg-[#232b3d] border border-line",
  ghost: "text-muted hover:text-ink hover:bg-surface/70",
  danger: "text-danger hover:bg-danger/10 border border-transparent hover:border-danger/30",
};

const sizes: Record<Size, string> = {
  sm: "h-8 px-3 text-[12.5px]",
  md: "h-9 px-4 text-[13px]",
};

export function Button({ variant = "subtle", size = "md", icon, className, children, ...rest }: ButtonProps) {
  return (
    <button type="button" className={cn(base, variants[variant], sizes[size], className)} {...rest}>
      {icon}
      {children}
    </button>
  );
}
