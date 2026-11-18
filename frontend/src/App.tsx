import { useEffect, useState } from "react";
import { TooltipProvider } from "@radix-ui/react-tooltip";
import { TopNav } from "./components/layout/TopNav";
import { Sidebar } from "./components/layout/Sidebar";
import { PlayerBar } from "./components/layout/PlayerBar";
import { MediaSessionBridge } from "./player/MediaSessionBridge";
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

/** Application shell: global navigation, active view, study panel, transport. */
export default function App() {
  const ready = useApp((state) => state.ready);
  const fatalError = useApp((state) => state.fatalError);
  const view = useApp((state) => state.view);
  const immersive = useApp((state) => state.immersive);
  const pickerTrack = useApp((state) => state.pickerTrack);
  const playlists = useApp((state) => state.playlists);
  const closePicker = useApp((state) => state.closePicker);
  const init = useApp((state) => state.init);
  const [navOpen, setNavOpen] = useState(false);

  useEffect(() => {
    void init();
  }, [init]);

  useKeyboardShortcuts();

  if (!ready) {
    return (
      <div className="flex h-full flex-col bg-bg" aria-busy="true" aria-label="Iniciando STEFANI">
        <div className="h-[52px] border-b border-line bg-bg" />
        <div className="mx-auto w-full max-w-[900px] px-8 pt-16">
          <div className="h-9 w-64 animate-pulse rounded-[8px] bg-surface" />
          <div className="mt-4 h-5 w-80 animate-pulse rounded-[6px] bg-line-soft" />
          <div className="mt-12 grid grid-cols-3 gap-6">
            {[0, 1, 2].map((key) => (
              <div key={key} className="space-y-3">
                <div className="aspect-square animate-pulse rounded-[16px] bg-surface" />
                <div className="h-4 w-3/4 animate-pulse rounded bg-line-soft" />
                <div className="h-3 w-1/2 animate-pulse rounded bg-line-soft" />
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (fatalError) {
    return (
      <div className="flex h-full items-center justify-center bg-bg px-8">
        <div className="max-w-[560px]">
          <p className="text-[12px] font-semibold text-danger">Error al iniciar</p>
          <h1 className="mt-2 text-[34px] leading-[1.1] text-ink">STEFANI no pudo cargar sus datos</h1>
          <p className="mt-3 text-[15px] leading-[1.47] text-muted">{fatalError}</p>
          <p className="mt-3 text-[13px] leading-[1.5] text-faint">
            Revisa los registros en la carpeta de configuración de la app. Abrir la aplicación no modifica la base de
            datos, asi que puedes corregir el problema e intentarlo otra vez.
          </p>
        </div>
      </div>
    );
  }

  return (
    <TooltipProvider delayDuration={350}>
      <div className="flex h-full flex-col bg-bg">
        <TopNav onToggleNav={() => setNavOpen((open) => !open)} />
        <div className="flex min-h-0 flex-1">
          <Sidebar open={navOpen} onNavigate={() => setNavOpen(false)} />
          <main className="scroll-area min-w-0 flex-1" id="main-content">
            {view === "home" && <HomeView />}
            {view === "search" && <SearchView />}
            {view === "library" && <LibraryView />}
            {view === "favorites" && <FavoritesView />}
            {view === "playlists" && <PlaylistsView />}
            {view === "learning" && <LearningView />}
            {view === "settings" && <SettingsView />}
          </main>
          <div className="hidden border-l border-line lg:flex">
            <StudyPanel />
          </div>
        </div>
        <PlayerBar />
        <MediaSessionBridge />
        <Toast />
        <PlaylistPicker track={pickerTrack} playlists={playlists} onClose={closePicker} />
        {immersive ? <ImmersiveView /> : null}
      </div>
    </TooltipProvider>
  );
}
