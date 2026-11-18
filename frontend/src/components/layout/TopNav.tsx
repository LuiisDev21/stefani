import { LoaderCircle, Menu, Search, Settings2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { cn } from "../../lib/cn";
import { useApp } from "../../stores/appStore";

interface TopNavProps {
  onToggleNav: () => void;
}

/**
 * Application chrome: the mark, one global search field and settings. Every
 * other destination lives in the sidebar, so the bar stays a command line
 * rather than a second menu.
 */
export function TopNav({ onToggleNav }: TopNavProps) {
  const view = useApp((state) => state.view);
  const go = useApp((state) => state.go);
  const query = useApp((state) => state.searchQuery);
  const searching = useApp((state) => state.searching);
  const search = useApp((state) => state.search);

  const [value, setValue] = useState(query);
  const inputRef = useRef<HTMLInputElement>(null);
  const timer = useRef<number | null>(null);

  // Keep the field in sync when the query is cleared from elsewhere.
  useEffect(() => {
    setValue(query);
  }, [query]);

  useEffect(() => {
    return () => {
      if (timer.current) window.clearTimeout(timer.current);
    };
  }, []);

  // Ctrl/Cmd+K focuses the field from anywhere, the way a command palette does.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        inputRef.current?.focus();
        inputRef.current?.select();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const run = (next: string) => {
    if (timer.current) window.clearTimeout(timer.current);
    void search(next);
  };

  const schedule = (next: string) => {
    setValue(next);
    if (view !== "search") go("search");
    if (timer.current) window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => void search(next), 350);
  };

  return (
    <header className="grid h-[52px] shrink-0 grid-cols-[1fr_minmax(0,520px)_1fr] items-center gap-4 border-b border-line bg-bg px-4 sm:px-6">
      <div className="flex min-w-0 items-center gap-2">
        <button
          type="button"
          onClick={onToggleNav}
          aria-label="Abrir navegación"
          className="flex h-8 w-8 items-center justify-center rounded-[6px] text-muted transition-colors hover:bg-surface hover:text-ink lg:hidden"
        >
          <Menu className="h-4 w-4" />
        </button>
        <button
          type="button"
          onClick={() => go("home")}
          className="shrink-0 font-display text-[15px] font-semibold tracking-[-0.03em] text-ink transition-opacity hover:opacity-80"
          aria-label="STEFANI, ir al inicio"
        >
          STEFANI
        </button>
      </div>

      <form
        role="search"
        onSubmit={(event) => {
          event.preventDefault();
          run(value);
        }}
        className="relative min-w-0"
      >
        <Search
          className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint"
          aria-hidden="true"
        />
        <input
          ref={inputRef}
          type="search"
          value={value}
          onChange={(event) => schedule(event.target.value)}
          placeholder="Buscar canciones, artistas o álbumes"
          aria-label="Buscar canciones"
          className="h-8 w-full rounded-[6px] border border-line bg-elevated pl-9 pr-16 text-[13px] text-ink placeholder:text-faint focus:border-line-strong"
        />
        <span className="pointer-events-none absolute right-2.5 top-1/2 hidden -translate-y-1/2 items-center gap-1 sm:flex">
          {searching ? (
            <LoaderCircle className="h-3.5 w-3.5 animate-spin text-faint" aria-hidden="true" />
          ) : (
            <kbd className="rounded-[3px] border border-line px-1 py-[1px] text-[10px] text-ash">Ctrl K</kbd>
          )}
        </span>
      </form>

      <div className="flex items-center justify-end gap-1">
        <button
          type="button"
          onClick={() => go("settings")}
          aria-label="Ajustes"
          aria-current={view === "settings" ? "page" : undefined}
          className={cn(
            "flex h-8 w-8 items-center justify-center rounded-[6px] transition-colors",
            view === "settings" ? "bg-card text-ink" : "text-muted hover:bg-surface hover:text-ink",
          )}
        >
          <Settings2 className="h-4 w-4" />
        </button>
      </div>
    </header>
  );
}
