import { Minus, Square, X } from "lucide-react";
import { api } from "../../lib/api";

const native = typeof window !== "undefined" && "_wails" in window;

/**
 * Minimise, maximise and close for the frameless window. Only rendered inside
 * the desktop app; the browser build keeps the system chrome.
 */
export function WindowControls() {
  if (!native) return null;
  return (
    <div
      className="ml-1 flex items-center"
      style={{ ["--wails-draggable" as string]: "no-drag" }}
    >
      <button
        type="button"
        aria-label="Minimizar"
        onClick={() => void api.windowMinimise()}
        className="flex h-8 w-[42px] items-center justify-center rounded-[6px] text-muted transition-colors hover:bg-surface hover:text-ink"
      >
        <Minus className="h-4 w-4" />
      </button>
      <button
        type="button"
        aria-label="Maximizar"
        onClick={() => void api.windowToggleMaximise()}
        className="flex h-8 w-[42px] items-center justify-center rounded-[6px] text-muted transition-colors hover:bg-surface hover:text-ink"
      >
        <Square className="h-3 w-3" />
      </button>
      <button
        type="button"
        aria-label="Cerrar"
        onClick={() => void api.windowClose()}
        className="flex h-8 w-[42px] items-center justify-center rounded-[6px] text-muted transition-colors hover:bg-[#c42b1c] hover:text-white"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
