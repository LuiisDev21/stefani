import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "../../lib/cn";

type Variant = "primary" | "subtle" | "ghost" | "danger";
type Size = "sm" | "md" | "lg";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  icon?: ReactNode;
}

/*
  Two grammars, exactly as the design system defines them: the white pill with
  black text for the one primary action, and quiet surfaces for everything
  else. Labels sit at weight 500; the rest of the chrome stays at 400.
*/
const base =
  "inline-flex select-none items-center justify-center gap-2 whitespace-nowrap font-medium transition-[background-color,color,border-color,transform] duration-150 active:scale-[0.98] disabled:pointer-events-none disabled:opacity-40";

const variants: Record<Variant, string> = {
  primary: "rounded-[8px] bg-accent text-on-accent hover:bg-accent-focus",
  subtle: "rounded-[8px] border border-line bg-transparent text-ink hover:border-line-strong hover:bg-surface",
  ghost: "rounded-[8px] text-body hover:bg-surface hover:text-ink",
  danger: "rounded-[8px] border border-line bg-transparent text-danger hover:border-danger/40 hover:bg-danger/10",
};

const sizes: Record<Size, string> = {
  sm: "h-8 px-3.5 text-[13px]",
  md: "h-9 px-4 text-[14px]",
  lg: "h-10 px-5 text-[14px]",
};

export function Button({ variant = "subtle", size = "md", icon, className, children, ...rest }: ButtonProps) {
  return (
    <button type="button" className={cn(base, variants[variant], sizes[size], className)} {...rest}>
      {icon}
      {children}
    </button>
  );
}
