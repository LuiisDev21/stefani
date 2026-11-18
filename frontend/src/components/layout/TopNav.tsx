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
 * Global navigation: the 44px black bar from the design system, with 12px
 * links and no icons. Every destination is a real view; the library count on
 * the right is real data, and import sits here because it is the one action
 * that turns an empty app into a usable one.
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
    <header className="flex h-11 shrink-0 items-center gap-6 bg-black px-4 text-white sm:px-6">
      <button
        type="button"
        onClick={() => go("home")}
        className="shrink-0 font-display text-[15px] font-semibold tracking-[-0.03em] text-white transition-opacity hover:opacity-80"
        aria-label="STEFANI, ir al inicio"
      >
        STEFANI
      </button>

      <nav aria-label="Navegacion principal" className="flex min-w-0 flex-1 items-center gap-0.5 overflow-x-auto">
        {items.map(({ view: target, label }) => {
          const active = view === target;
          return (
            <button
              key={target}
              type="button"
              onClick={() => go(target)}
              aria-current={active ? "page" : undefined}
              className={cn(
                "shrink-0 rounded-full px-3 py-1.5 text-[12px] tracking-[-0.12px] transition-colors",
                active ? "text-white" : "text-white/70 hover:text-white",
              )}
            >
              {label}
            </button>
          );
        })}
      </nav>

      <div className="flex shrink-0 items-center gap-4">
        {/* Compact now-playing block: what Windows' media widget shows, but in
            the app's own bar, with the two controls that matter while reading. */}
        {current ? (
          <div className="flex min-w-0 items-center gap-2 border-r border-white/15 pr-3">
            <div className="h-7 w-7 shrink-0 overflow-hidden rounded-[6px] bg-white/10">
              {current.artworkUrl ? (
                <img src={current.artworkUrl} alt="" className="h-full w-full object-cover" draggable={false} />
              ) : (
                <span className="flex h-full items-center justify-center text-[11px] font-semibold text-white/40">
                  {current.title.slice(0, 1).toUpperCase()}
                </span>
              )}
            </div>
            <p className="hidden min-w-0 max-w-[210px] truncate text-[12px] text-white/80 lg:block" title={`${current.artist} — ${current.title}`}>
              {current.artist} — {current.title}
            </p>
            <button
              type="button"
              onClick={toggle}
              aria-label={playing ? "Pausar" : "Reproducir"}
              className="flex h-7 w-7 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            >
              {playing ? <Pause className="h-3.5 w-3.5 fill-current" /> : <Play className="ml-0.5 h-3.5 w-3.5 fill-current" />}
            </button>
            <button
              type="button"
              onClick={next}
              aria-label="Siguiente"
              className="flex h-7 w-7 items-center justify-center rounded-full text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            >
              <SkipForward className="h-3.5 w-3.5 fill-current" />
            </button>
          </div>
        ) : null}
        <p className={cn("text-[11px] text-white/45 tnum", current ? "hidden" : "hidden md:block")}>
          {counts ? `${counts.tracks} pistas · ${counts.annotated} anotaciones` : "Cargando"}
        </p>
        <button
          type="button"
          onClick={() => void importAudio()}
          className="flex h-7 items-center gap-1.5 rounded-[8px] bg-white px-3 text-[12px] font-normal text-ink transition-opacity hover:opacity-85 active:scale-[0.97]"
        >
          <Upload className="h-3 w-3" aria-hidden="true" />
          Importar
        </button>
      </div>
    </header>
  );
}
