import { useEffect } from "react";
import { TooltipProvider } from "@radix-ui/react-tooltip";
import { Sidebar } from "./components/layout/Sidebar";
import { PlayerBar } from "./components/layout/PlayerBar";
import { StudyPanel } from "./components/layout/StudyPanel";
import { Toast } from "./components/ui/Toast";
import { PlaylistPicker } from "./components/track/PlaylistPicker";
import { HomeView } from "./features/home/HomeView";
import { SearchView } from "./features/search/SearchView";
import { LibraryView } from "./features/library/LibraryView";
import { FavoritesView } from "./features/favorites/FavoritesView";
import { PlaylistsView } from "./features/playlists/PlaylistsView";
import { LearningView } from "./features/learning/LearningView";
import { SettingsView } from "./features/settings/SettingsView";
import { ImmersiveView } from "./features/immersive/ImmersiveView";
import { useApp } from "./stores/appStore";
import { useKeyboardShortcuts } from "./lib/useKeyboardShortcuts";

/** Application shell: navigation rail, active view, study panel, transport. */
export default function App() {
  const ready = useApp((state) => state.ready);
  const fatalError = useApp((state) => state.fatalError);
  const view = useApp((state) => state.view);
  const immersive = useApp((state) => state.immersive);
  const pickerTrack = useApp((state) => state.pickerTrack);
  const playlists = useApp((state) => state.playlists);
  const closePicker = useApp((state) => state.closePicker);
  const init = useApp((state) => state.init);

  useEffect(() => {
    void init();
  }, [init]);

  useKeyboardShortcuts();

  if (!ready) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="w-[280px] space-y-3" aria-busy="true" aria-label="Iniciando STEFANI">
          <p className="text-[13px] font-semibold text-muted">Stefani</p>
          <div className="h-[2px] w-full overflow-hidden rounded bg-line">
            <div className="h-full w-1/3 animate-pulse bg-accent" />
          </div>
          <p className="text-[11.5px] text-faint">Preparando la biblioteca…</p>
        </div>
      </div>
    );
  }

  if (fatalError) {
    return (
      <div className="flex h-full items-center justify-center px-8">
        <div className="max-w-[560px]">
          <p className="text-[12px] font-semibold text-danger">Error al iniciar</p>
          <h1 className="mt-2 text-[20px] font-medium text-ink">STEFANI no pudo cargar sus datos</h1>
          <p className="mt-2 text-[12.5px] leading-relaxed text-muted">{fatalError}</p>
          <p className="mt-3 text-[11.5px] leading-relaxed text-faint">
            Revisa los registros en la carpeta de configuracion de la app. Abrir la aplicacion no modifica la base de
            datos, asi que puedes corregir el problema e intentarlo otra vez.
          </p>
        </div>
      </div>
    );
  }

  return (
    <TooltipProvider delayDuration={350}>
      <div className="flex h-full flex-col">
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          <main className="scroll-area min-w-0 flex-1" id="main-content">
            {view === "home" && <HomeView />}
            {view === "search" && <SearchView />}
            {view === "library" && <LibraryView />}
            {view === "favorites" && <FavoritesView />}
            {view === "playlists" && <PlaylistsView />}
            {view === "learning" && <LearningView />}
            {view === "settings" && <SettingsView />}
          </main>
          <div className="hidden lg:flex">
            <StudyPanel />
          </div>
        </div>
        <PlayerBar />
        <Toast />
        <PlaylistPicker track={pickerTrack} playlists={playlists} onClose={closePicker} />
        {immersive ? <ImmersiveView /> : null}
      </div>
    </TooltipProvider>
  );
}
