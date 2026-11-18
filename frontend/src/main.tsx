import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./styles/index.css";

declare global {
  interface Window {
    __stefaniErrors?: string[];
  }
}

window.__stefaniErrors = [];
const record = (value: unknown) => {
  const text = String((value as Error)?.stack ?? value).slice(0, 300);
  window.__stefaniErrors?.push(text);
  if (window.__stefaniErrors && window.__stefaniErrors.length > 10) {
    window.__stefaniErrors.shift();
  }
};
window.addEventListener("error", (event) => record(event.error ?? event.message));
window.addEventListener("unhandledrejection", (event) => record(event.reason));

class Boundary extends React.Component<{ children: React.ReactNode }, { error: string | null }> {
  state = { error: null as string | null };

  static getDerivedStateFromError(error: unknown) {
    return { error: String((error as Error)?.message ?? error) };
  }

  componentDidCatch(error: unknown) {
    record(`boundary: ${String((error as Error)?.stack ?? error)}`);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex h-screen flex-col items-center justify-center gap-4 bg-[#07080a] px-8 text-center">
          <p className="text-[14px] text-ink">La interfaz falló</p>
          <p className="max-w-[520px] break-words text-[12px] text-muted">{this.state.error}</p>
          <button
            type="button"
            className="rounded-[8px] border border-line px-4 py-2 text-[13px] text-ink"
            onClick={() => location.reload()}
          >
            Reintentar
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <Boundary>
      <App />
    </Boundary>
  </React.StrictMode>,
);
