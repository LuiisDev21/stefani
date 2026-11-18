import { Pause, Play, SkipForward, Upload } from "lucide-react";
import { cn } from "../../lib/cn";
import { useApp, type View } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";

const items: Array<{ view: View; label: string }> = [
  { view: "home", label: "Inicio" },
  { view: "search", label: "Buscar" },
  { view: "library", label: "Biblioteca" },
  { view: "playlists", label: "Listas" },
  { view: "favorites", label: "Favoritos" },
  { view: "learning", label: "Aprendizaje" },
  { view: "settings", label: "Ajustes" },
];

/**
 * Global navigation: a 52px canvas bar closed by a single hairline, with the
 * active destination on the card surface instead of a colored underline. Every
 * destination is a real view; the count on the right is real data, and import
 * sits here because it is the one action that turns an empty app into a usable
 * one.
 */
export function TopNav() {
  const view = useApp((state) => state.view);
  const go = useApp((state) => state.go);
  const counts = useApp((state) => state.counts);
  const importAudio = useApp((state) => state.importAudio);
  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const toggle = usePlayer((state) => state.toggle);
  const next = usePlayer((state) => state.next);

  return (
    <header className="flex h-[52px] shrink-0 items-center gap-6 border-b border-line bg-bg px-4 sm:px-6">
      <button
        type="button"
        onClick={() => go("home")}
        className="shrink-0 font-display text-[15px] font-semibold tracking-[-0.03em] text-ink transition-opacity hover:opacity-80"
        aria-label="STEFANI, ir al inicio"
      >
        STEFANI
      </button>

      <nav aria-label="Navegacion principal" className="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto">
        {items.map(({ view: target, label }) => {
          const active = view === target;
          return (
            <button
              key={target}
              type="button"
              onClick={() => go(target)}
              aria-current={active ? "page" : undefined}
              className={cn(
                "shrink-0 rounded-[6px] px-3 py-1.5 text-[13px] transition-colors",
                active ? "bg-card text-ink" : "text-muted hover:bg-surface hover:text-ink",
              )}
            >
              {label}
            </button>
          );
        })}
      </nav>

      <div className="flex shrink-0 items-center gap-4">
        {current ? (
          <div className="flex min-w-0 items-center gap-2 border-r border-line pr-3">
            <div className="h-7 w-7 shrink-0 overflow-hidden rounded-[6px] border border-line bg-card">
              {current.artworkUrl ? (
                <img src={current.artworkUrl} alt="" className="h-full w-full object-cover" draggable={false} />
              ) : (
                <span className="flex h-full items-center justify-center text-[11px] font-semibold text-ash">
                  {current.title.slice(0, 1).toUpperCase()}
                </span>
              )}
            </div>
            <p className="hidden min-w-0 max-w-[210px] truncate text-[12px] text-body lg:block" title={`${current.artist} — ${current.title}`}>
              {current.artist} — {current.title}
            </p>
            <button
              type="button"
              onClick={toggle}
              aria-label={playing ? "Pausar" : "Reproducir"}
              className="flex h-7 w-7 items-center justify-center rounded-[6px] text-body transition-colors hover:bg-surface hover:text-ink"
            >
              {playing ? <Pause className="h-3.5 w-3.5 fill-current" /> : <Play className="ml-0.5 h-3.5 w-3.5 fill-current" />}
            </button>
            <button
              type="button"
              onClick={next}
              aria-label="Siguiente"
              className="flex h-7 w-7 items-center justify-center rounded-[6px] text-body transition-colors hover:bg-surface hover:text-ink"
            >
              <SkipForward className="h-3.5 w-3.5 fill-current" />
            </button>
          </div>
        ) : null}
        <p className={cn("text-[11px] text-ash tnum", current ? "hidden" : "hidden md:block")}>
          {counts ? `${counts.tracks} pistas · ${counts.annotated} anotaciones` : "Cargando"}
        </p>
        <button
          type="button"
          onClick={() => void importAudio()}
          className="flex h-8 items-center gap-1.5 rounded-[6px] bg-accent px-3 text-[13px] font-medium text-on-accent transition-colors hover:bg-accent-focus active:scale-[0.98]"
        >
          <Upload className="h-3.5 w-3.5" aria-hidden="true" />
          Importar
        </button>
      </div>
    </header>
  );
}
