import {
  GraduationCap,
  Heart,
  Home,
  Library,
  ListMusic,
  Search,
  Settings,
  Upload,
} from "lucide-react";
import { cn } from "../../lib/cn";
import { useApp, type View } from "../../stores/appStore";
import { Button } from "../ui/Button";

const items: Array<{ view: View; label: string; icon: typeof Home }> = [
  { view: "home", label: "Inicio", icon: Home },
  { view: "search", label: "Buscar", icon: Search },
  { view: "library", label: "Biblioteca", icon: Library },
  { view: "playlists", label: "Listas", icon: ListMusic },
  { view: "favorites", label: "Favoritos", icon: Heart },
  { view: "learning", label: "Aprendizaje", icon: GraduationCap },
  { view: "settings", label: "Ajustes", icon: Settings },
];

/**
 * Left rail. Seven destinations, each backed by a real view; the storage line at
 * the bottom reports actual counts, and import lives here because it is the one
 * action that turns an empty app into a usable one.
 */
export function Sidebar() {
  const view = useApp((state) => state.view);
  const go = useApp((state) => state.go);
  const counts = useApp((state) => state.counts);
  const importAudio = useApp((state) => state.importAudio);

  return (
    <aside className="flex w-[208px] shrink-0 flex-col border-r border-line/70 px-3 pb-3 pt-4">
      <div className="flex items-start gap-2.5 px-2">
        <SignalMark />
        <div className="min-w-0">
          <p className="text-[15px] font-semibold leading-none tracking-[-0.01em] text-ink">Stefani</p>
          <p className="mt-1.5 text-[11px] leading-tight text-muted">Musica para aprender ingles</p>
        </div>
      </div>

      <div className="signal-line my-4" />

      <nav aria-label="Navegacion principal" className="space-y-0.5">
        {items.map(({ view: target, label, icon: Icon }) => {
          const active = view === target;
          return (
            <button
              key={target}
              type="button"
              onClick={() => go(target)}
              aria-current={active ? "page" : undefined}
              className={cn(
                "flex w-full items-center gap-2.5 rounded-[8px] px-2.5 py-[7px] text-left text-[13px] transition-colors",
                active
                  ? "bg-surface text-ink"
                  : "text-muted hover:bg-surface/50 hover:text-ink",
              )}
            >
              <Icon className={cn("h-4 w-4", active ? "text-accent" : "text-faint")} aria-hidden="true" />
              <span className="truncate">{label}</span>
            </button>
          );
        })}
      </nav>

      <div className="mt-auto space-y-2.5 pt-4">
        <Button
          size="sm"
          variant="subtle"
          className="w-full"
          icon={<Upload className="h-3.5 w-3.5" />}
          onClick={() => void importAudio()}
        >
          Importar audio
        </Button>
        <p className="px-1 text-[11px] leading-relaxed text-faint tnum">
          {counts ? `${counts.tracks} pistas · ${counts.annotated} anotaciones` : "Cargando datos"}
        </p>
      </div>
    </aside>
  );
}

/** Three bars of unequal height: a signal, not a logo in a box. */
function SignalMark() {
  return (
    <span aria-hidden="true" className="mt-[3px] flex h-4 w-4 items-end gap-[2px]">
      <span className="h-[7px] w-[2px] rounded-full bg-accent" />
      <span className="h-4 w-[2px] rounded-full bg-accent/80" />
      <span className="h-[10px] w-[2px] rounded-full bg-accent/50" />
    </span>
  );
}
