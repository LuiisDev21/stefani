import { CheckCircle2, Info, X, XCircle } from "lucide-react";
import { useApp } from "../../stores/appStore";
import { cn } from "../../lib/cn";

/** Bottom-right status message. Announced politely; never steals focus. */
export function Toast() {
  const toast = useApp((state) => state.toast);
  const dismiss = useApp((state) => state.dismissToast);
  if (!toast) return null;

  const Icon = toast.kind === "ok" ? CheckCircle2 : toast.kind === "error" ? XCircle : Info;
  const accent =
    toast.kind === "ok" ? "text-ink" : toast.kind === "error" ? "text-danger" : "text-ink";

  return (
    <div
      role="status"
      aria-live="polite"
      className="fixed bottom-[104px] left-1/2 z-50 w-[min(520px,calc(100vw-32px))] -translate-x-1/2"
    >
      <div
        className={cn(
          "flex items-start gap-3 rounded-[10px] border border-line bg-card px-4 py-3",
        )}
      >
        <Icon className={cn("mt-0.5 h-4 w-4 shrink-0", accent)} aria-hidden="true" />
        <p className="flex-1 text-[13px] leading-snug text-ink">{toast.message}</p>
        <button
          type="button"
          onClick={dismiss}
          aria-label="Cerrar aviso"
          className="rounded p-0.5 text-faint transition-colors hover:text-ink"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  );
}
