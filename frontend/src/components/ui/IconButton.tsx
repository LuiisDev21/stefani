import * as Tooltip from "@radix-ui/react-tooltip";
import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "../../lib/cn";

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
  children: ReactNode;
  active?: boolean;
  size?: "sm" | "md";
}

/**
 * Icon-only control. The accessible name is mandatory, and the tooltip repeats
 * it on hover so the icon never has to be guessed.
 */
export function IconButton({ label, children, active, size = "md", className, ...rest }: IconButtonProps) {
  return (
    <Tooltip.Root delayDuration={350}>
      <Tooltip.Trigger asChild>
        <button
          type="button"
          aria-label={label}
          aria-pressed={active}
          className={cn(
            "inline-flex items-center justify-center rounded-full text-ink transition-colors",
            size === "sm" ? "h-8 w-8" : "h-9 w-9",
            "hover:bg-surface",
            active && "text-accent hover:text-accent",
            className,
          )}
          {...rest}
        >
          {children}
        </button>
      </Tooltip.Trigger>
      <Tooltip.Portal>
        <Tooltip.Content
          sideOffset={6}
          className="z-50 rounded-[6px] border border-line bg-card px-2.5 py-1.5 text-[12px] text-ink"
        >
          {label}
          <Tooltip.Arrow className="fill-card" />
        </Tooltip.Content>
      </Tooltip.Portal>
    </Tooltip.Root>
  );
}
