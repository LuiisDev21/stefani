import { GraduationCap, Heart, House, Library, ListMusic, Search, Upload } from "lucide-react";
import { cn } from "../../lib/cn";
import { useApp, type View } from "../../stores/appStore";
import { Button } from "../ui/Button";

const groups: Array<{ label: string; items: Array<{ view: View; label: string; icon: typeof House }> }> = [
  {
    label: "Escuchar",
    items: [
      { view: "home", label: "Inicio", icon: House },
      { view: "search", label: "Buscar", icon: Search },
      { view: "library", label: "Biblioteca", icon: Library },
      { view: "playlists", label: "Listas", icon: ListMusic },
      { view: "favorites", label: "Favoritos", icon: Heart },
    ],
  },
  {
    label: "Estudiar",
    items: [{ view: "learning", label: "Aprendizaje", icon: GraduationCap }],
  },
];

interface SidebarProps {
  /** Mobile drawer state; on desktop the sidebar is always visible. */
  open: boolean;
  onNavigate: () => void;
}

/**
 * Persistent destination list. It carries the sections and the two pieces of
 * chrome that are not navigation — the real counters and the import action —
 * so the top bar can stay down to a logo, a search field and settings.
 */
export function Sidebar({ open, onNavigate }: SidebarProps) {
  const view = useApp((state) => state.view);
  const go = useApp((state) => state.go);
  const counts = useApp((state) => state.counts);
  const importAudio = useApp((state) => state.importAudio);

  const navigate = (target: View) => {
    go(target);
    onNavigate();
  };

  return (
    <>
      {open ? (
        <button
          type="button"
          aria-label="Cerrar navegación"
          onClick={onNavigate}
          className="fixed inset-0 z-30 bg-black/60 lg:hidden"
        />
      ) : null}
      <aside
        aria-label="Navegación principal"
        className={cn(
          "fixed inset-y-0 left-0 z-40 flex w-[232px] shrink-0 flex-col border-r border-line bg-bg px-3 py-4",
          "transition-transform duration-200 lg:static lg:z-auto lg:translate-x-0",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <nav className="flex flex-col gap-6">
          {groups.map((group) => (
            <div key={group.label}>
              <p className="px-2.5 pb-1.5 text-[11px] text-ash">{group.label}</p>
              <ul className="space-y-0.5">
                {group.items.map(({ view: target, label, icon: Icon }) => {
                  const active = view === target;
                  return (
                    <li key={target}>
                      <button
                        type="button"
                        onClick={() => navigate(target)}
                        aria-current={active ? "page" : undefined}
                        className={cn(
                          "flex h-8 w-full items-center gap-2.5 rounded-[6px] px-2.5 text-[13px] transition-colors",
                          active ? "bg-card text-ink" : "text-muted hover:bg-surface hover:text-ink",
                        )}
                      >
                        <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
                        <span className="truncate">{label}</span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>

        <div className="mt-auto space-y-3 border-t border-line pt-3">
          <p className="px-2.5 text-[11px] leading-[1.5] text-ash tnum">
            {counts ? `${counts.tracks} pistas · ${counts.annotated} anotaciones` : "Cargando"}
          </p>
          <Button
            variant="subtle"
            size="sm"
            icon={<Upload className="h-3.5 w-3.5" />}
            onClick={() => void importAudio()}
            className="w-full"
          >
            Importar
          </Button>
        </div>
      </aside>
    </>
  );
}
