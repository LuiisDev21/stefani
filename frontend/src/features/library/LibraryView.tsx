import { LibraryBig, Upload } from "lucide-react";
import { useApp } from "../../stores/appStore";
import { Button } from "../../components/ui/Button";
import { EmptyState } from "../../components/ui/EmptyState";
import { TopBar } from "../../components/layout/TopBar";
import { TrackList } from "../../components/track/TrackList";

/** Saved tracks. */
export function LibraryView() {
  const library = useApp((state) => state.library);
  const counts = useApp((state) => state.counts);
  const importAudio = useApp((state) => state.importAudio);
  const openPicker = useApp((state) => state.openPicker);
  const focusSearch = useApp((state) => state.focusSearch);

  return (
    <div className="min-h-0">
      <TopBar
        title="Biblioteca"
        subtitle={
          counts
            ? `${counts.libraryTracks} guardadas · ${counts.favorites} favoritas · ${counts.tracks} totales`
            : undefined
        }
        actions={
          <>
            <Button size="sm" variant="subtle" icon={<Upload className="h-3.5 w-3.5" />} onClick={() => void importAudio()}>
              Importar audio
            </Button>
            <Button size="sm" variant="ghost" onClick={() => focusSearch()}>
              Buscar mas música
            </Button>
          </>
        }
      />

      <div className="px-7 pb-10">
        {library.length === 0 ? (
          <EmptyState
            icon={<LibraryBig className="h-5 w-5" />}
            title="Tu biblioteca esta vacia"
            description="Guarda una canción con el corazon, o importa archivos de audio desde tu equipo. Los archivos se referencian en su ubicacion original, no se copian."
            actions={
              <>
                <Button size="sm" variant="primary" onClick={() => focusSearch()}>
                  Buscar canciones
                </Button>
                <Button size="sm" variant="subtle" onClick={() => void importAudio()}>
                  Importar audio
                </Button>
              </>
            }
          />
        ) : (
          <TrackList tracks={library} onAdd={openPicker} />
        )}
      </div>
    </div>
  );
}
