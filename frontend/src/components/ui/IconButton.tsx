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
            "inline-flex items-center justify-center rounded-[8px] text-muted transition-colors",
            size === "sm" ? "h-7 w-7" : "h-9 w-9",
            "hover:bg-surface hover:text-ink",
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
          className="z-50 rounded-[6px] border border-line bg-surface px-2 py-1 text-[11.5px] text-ink shadow-[0_8px_24px_rgba(0,0,0,0.45)]"
        >
          {label}
          <Tooltip.Arrow className="fill-surface" />
        </Tooltip.Content>
      </Tooltip.Portal>
    </Tooltip.Root>
  );
}
