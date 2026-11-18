import { LoaderCircle, Menu, Search, Settings } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { cn } from "../../lib/cn";
import { formatTime } from "../../lib/format";
import type { Track } from "../../lib/types";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Cover } from "../ui/Cover";

interface TopNavProps {
  onToggleNav: () => void;
}

const MAX_RESULTS = 12;

/**
 * Application chrome: the mark, one global search field and settings. Every
 * other destination lives in the sidebar, so the bar stays a command line
 * rather than a second menu. Results open under the field, so searching never
 * costs a page.
 */
export function TopNav({ onToggleNav }: TopNavProps) {
  const view = useApp((state) => state.view);
  const go = useApp((state) => state.go);
  const query = useApp((state) => state.searchQuery);
  const searching = useApp((state) => state.searching);
  const result = useApp((state) => state.searchResult);
  const search = useApp((state) => state.search);
  const clearSearch = useApp((state) => state.clearSearch);
  const focusToken = useApp((state) => state.searchFocus);
  const playTrack = usePlayer((state) => state.playTrack);

  const [value, setValue] = useState(query);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(0);

  const inputRef = useRef<HTMLInputElement>(null);
  const boxRef = useRef<HTMLDivElement>(null);
  const timer = useRef<number | null>(null);

  // Searching the catalogue returns the same song once per past lookup; a
  // dropdown needs one row per song.
  const items = useMemo(() => {
    const seen = new Set<string>();
    return (result?.items ?? [])
      .filter((track) => {
        const key = `${track.title.toLowerCase()}|${track.artist.toLowerCase()}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      })
      .slice(0, MAX_RESULTS);
  }, [result]);
  const panelOpen = open && value.trim().length > 0;

  // The query only changes from the panel itself; keep the field honest.
  useEffect(() => {
    setValue(query);
  }, [query]);

  // Any "search" affordance in the app bumps this token.
  useEffect(() => {
    if (focusToken === 0) return;
    inputRef.current?.focus();
    setOpen(true);
  }, [focusToken]);

  useEffect(() => {
    setActive(0);
  }, [result]);

  useEffect(() => {
    return () => {
      if (timer.current) window.clearTimeout(timer.current);
    };
  }, []);

  // A click anywhere else closes the list without erasing the text.
  useEffect(() => {
    const onPointerDown = (event: PointerEvent) => {
      if (!boxRef.current?.contains(event.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, []);

  // Ctrl/Cmd+K focuses the field from anywhere, the way a command palette does.
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        inputRef.current?.focus();
        inputRef.current?.select();
        setOpen(true);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  const choose = (track: Track) => {
    playTrack(track);
    clearSearch();
    setValue("");
    setOpen(false);
    inputRef.current?.blur();
  };

  const schedule = (next: string) => {
    setValue(next);
    setOpen(true);
    if (timer.current) window.clearTimeout(timer.current);
    if (!next.trim()) {
      clearSearch();
      return;
    }
    timer.current = window.setTimeout(() => void search(next), 350);
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
      inputRef.current?.blur();
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setOpen(true);
      setActive((index) => Math.min(index + 1, Math.max(items.length - 1, 0)));
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive((index) => Math.max(index - 1, 0));
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      const target = items[active] ?? items[0];
      if (target) choose(target);
    }
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

      <div ref={boxRef} className="relative min-w-0">
        <form
          role="search"
          onSubmit={(event) => {
            event.preventDefault();
            const target = items[active] ?? items[0];
            if (target) choose(target);
          }}
          className="relative"
        >
          <Search
            className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-faint"
            aria-hidden="true"
          />
          <input
            ref={inputRef}
            type="search"
            value={value}
            role="combobox"
            aria-label="Buscar canciones"
            aria-expanded={panelOpen}
            aria-controls="global-search-results"
            aria-autocomplete="list"
            aria-activedescendant={
              panelOpen && items[active] ? `search-option-${items[active].id}` : undefined
            }
            onChange={(event) => schedule(event.target.value)}
            onFocus={() => setOpen(true)}
            onKeyDown={onKeyDown}
            placeholder="Buscar canciones, artistas o álbumes"
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

        {panelOpen ? (
          <div
            id="global-search-results"
            role="listbox"
            aria-label="Resultados de búsqueda"
            className="absolute left-0 right-0 top-[calc(100%+12px)] z-50 max-h-[420px] overflow-y-auto rounded-[12px] border border-line bg-elevated p-1"
          >
            {searching && items.length === 0 ? (
              <div className="space-y-1 p-1" aria-busy="true">
                {[0, 1, 2].map((index) => (
                  <div key={index} className="h-10 animate-pulse rounded-[8px] bg-surface/60" />
                ))}
              </div>
            ) : items.length > 0 ? (
              items.map((track, index) => (
                <button
                  key={track.id}
                  id={`search-option-${track.id}`}
                  type="button"
                  role="option"
                  aria-selected={index === active}
                  onMouseEnter={() => setActive(index)}
                  onClick={() => choose(track)}
                  className={cn(
                    "flex w-full items-center gap-2.5 rounded-[8px] px-2 py-1.5 text-left transition-colors",
                    index === active ? "bg-card" : "hover:bg-card/70",
                  )}
                >
                  <Cover title={track.title} artist={track.artist} src={track.artworkUrl} size={30} />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] text-ink">{track.title}</span>
                    <span className="block truncate text-[11.5px] text-muted">
                      {track.artist || "Sin artista"}
                    </span>
                  </span>
                  {track.durationMs > 0 ? (
                    <span className="shrink-0 text-[11px] text-faint tnum">
                      {formatTime(track.durationMs)}
                    </span>
                  ) : null}
                </button>
              ))
            ) : (
              <p className="px-3 py-4 text-[13px] text-muted">Sin resultados</p>
            )}
          </div>
        ) : null}
      </div>

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
          <Settings className="h-4 w-4" />
        </button>
      </div>
    </header>
  );
}
