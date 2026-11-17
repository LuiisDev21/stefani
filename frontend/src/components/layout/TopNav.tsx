import { Upload } from "lucide-react";
import { cn } from "../../lib/cn";
import { useApp, type View } from "../../stores/appStore";

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
        <p className="hidden text-[11px] text-white/45 tnum md:block">
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
