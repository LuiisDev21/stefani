import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "../../lib/cn";

interface ModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  width?: string;
}

/**
 * Modal built on Radix: focus trap, Escape to close and scroll lock come from the
 * primitive instead of being reimplemented per screen.
 */
export function Modal({ open, onOpenChange, title, description, children, footer, width = "max-w-[460px]" }: ModalProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-black/70" />
        <Dialog.Content
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-[calc(100vw-32px)] -translate-x-1/2 -translate-y-1/2",
            "rounded-[16px] border border-line bg-card p-6",
            width,
          )}
        >
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-[16px] font-medium text-ink">{title}</Dialog.Title>
              {description ? (
                <Dialog.Description className="mt-1 text-[12.5px] text-muted">{description}</Dialog.Description>
              ) : null}
            </div>
            <Dialog.Close
              aria-label="Cerrar"
              className="rounded p-1 text-faint transition-colors hover:bg-surface hover:text-ink"
            >
              <X className="h-4 w-4" />
            </Dialog.Close>
          </div>
          <div className="mt-4">{children}</div>
          {footer ? <div className="mt-5 flex justify-end gap-2">{footer}</div> : null}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
