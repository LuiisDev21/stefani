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
  Two grammars, exactly as the design system defines them: the pill for
  actions (Action Blue, full radius) and quiet surfaces for everything else.
  Weight stays 400; the ladder has no 500.
*/
const base =
  "inline-flex select-none items-center justify-center gap-2 whitespace-nowrap font-normal transition-[background-color,color,transform,border-color] duration-150 active:scale-[0.97] disabled:pointer-events-none disabled:opacity-40";

const variants: Record<Variant, string> = {
  primary: "rounded-full bg-accent text-on-accent hover:bg-accent-focus",
  subtle: "rounded-full border border-line bg-bg text-ink hover:bg-elevated",
  ghost: "rounded-full text-accent hover:bg-accent/8",
  danger: "rounded-full border border-line bg-bg text-danger hover:bg-danger/6",
};

const sizes: Record<Size, string> = {
  sm: "h-8 px-3.5 text-[13px]",
  md: "h-10 px-5 text-[14px]",
  lg: "h-11 px-6 text-[15px]",
};

export function Button({ variant = "subtle", size = "md", icon, className, children, ...rest }: ButtonProps) {
  return (
    <button type="button" className={cn(base, variants[variant], sizes[size], className)} {...rest}>
      {icon}
      {children}
    </button>
  );
}
