import { ChevronLeft, Play, Shuffle } from "lucide-react";
import { useApp } from "../../stores/appStore";
import { Button } from "../../components/ui/Button";
import { Cover } from "../../components/ui/Cover";
import { EmptyState } from "../../components/ui/EmptyState";
import { ShelfRow } from "../../components/home/ShelfRow";
import { TrackList } from "../../components/track/TrackList";
import { TopBar } from "../../components/layout/TopBar";

/**
 * One page of the catalogue: an album, a playlist or an artist. Albums and
 * lists show their songs; an artist shows its popular songs and the rows of
 * records and playlists that lead further in.
 */
export function CollectionView() {
  const collection = useApp((state) => state.collection);
  const loading = useApp((state) => state.collectionLoading);
  const go = useApp((state) => state.go);
  const openTrack = useApp((state) => state.openTrack);
  const openCollection = useApp((state) => state.openCollection);

  const back = (
    <button
      type="button"
      onClick={() => go("home")}
      className="mb-4 inline-flex h-8 items-center gap-1.5 rounded-full pr-3 pl-2 text-[12.5px] text-muted transition-colors hover:text-ink"
      aria-label="Volver al inicio"
    >
      <ChevronLeft size={16} strokeWidth={2} />
      Inicio
    </button>
  );

  if (loading) {
    return (
      <>
        <TopBar title="Abriendo" />
        <div className="px-8 pb-28">
          {back}
          <div className="flex animate-pulse items-end gap-6">
            <div className="h-[168px] w-[168px] rounded-[10px] bg-surface" />
            <div className="space-y-3 pb-2">
              <div className="h-3 w-16 rounded bg-surface" />
              <div className="h-8 w-72 rounded bg-surface" />
              <div className="h-3 w-48 rounded bg-surface" />
            </div>
          </div>
        </div>
      </>
    );
  }

  if (!collection) {
    return (
      <>
        <TopBar title="Colección" />
        <div className="px-8 pb-28">
          {back}
          <EmptyState
            title="No se pudo abrir"
            description="La colección no está disponible en esta cuenta."
          />
        </div>
      </>
    );
  }

  const playAll = (shuffle: boolean) => {
    if (collection.tracks.length === 0) return;
    const queue = shuffle
      ? [...collection.tracks].sort(() => Math.random() - 0.5)
      : collection.tracks;
    void openTrack(queue[0], queue);
  };

  const kindLabel =
    collection.kind === "artist" ? "Artista" : collection.kind === "album" ? "Álbum" : "Lista";
  // The subtitle repeats the kind ("Álbum • 2026"), which the label already says.
  const detail = collection.subtitle.replace(/^(Álbum|EP|Sencillo|Single|Lista|Playlist)\s*•\s*/i, "");

  return (
    <>
      <TopBar title={collection.title} />
      <div className="px-8 pb-28">
        {back}

        <header className="flex items-end gap-6">
          <Cover
            title={collection.title}
            src={collection.cover}
            size={168}
            rounded={collection.kind === "artist" ? "rounded-full" : "rounded-[10px]"}
            bordered={collection.kind !== "artist"}
          />
          <div className="min-w-0 pb-1">
            <p className="text-[11px] font-medium text-muted">{kindLabel}</p>
            <h2 className="mt-1 truncate text-[34px] leading-[1.1] font-semibold tracking-[-0.03em] text-ink">
              {collection.title}
            </h2>
            <p className="mt-2 text-[12.5px] text-muted">
              {[collection.artist, detail, collection.meta]
                .filter((part) => part && part !== kindLabel)
                .join(" · ")}
            </p>
          </div>
        </header>

        <div className="mt-6 flex items-center gap-2">
          <Button onClick={() => playAll(false)} disabled={collection.tracks.length === 0}>
            <Play size={15} strokeWidth={2.4} className="fill-current" />
            Reproducir
          </Button>
          <button
            type="button"
            onClick={() => playAll(true)}
            disabled={collection.tracks.length === 0}
            className="inline-flex h-9 items-center gap-2 rounded-full border border-line px-4 text-[12.5px] text-body transition-colors hover:text-ink disabled:opacity-40"
          >
            <Shuffle size={15} strokeWidth={2} />
            Aleatorio
          </button>
        </div>

        {collection.tracks.length > 0 && (
          <section className="mt-9" aria-label="Canciones">
            <TrackList tracks={collection.tracks} />
          </section>
        )}

        {collection.tracks.length === 0 && collection.shelves.length === 0 && (
          <div className="mt-10">
            <EmptyState
              title="Sin canciones"
              description="Esta colección no expone sus pistas en la cuenta."
            />
          </div>
        )}

        {collection.shelves.map((shelf) => (
          <ShelfRow key={shelf.title} shelf={shelf} />
        ))}

        {collection.kind !== "artist" && collection.artistId !== "" && (
          <button
            type="button"
            onClick={() => void openCollection(collection.artistId)}
            className="mt-10 inline-flex items-center gap-2 text-[12.5px] text-muted transition-colors hover:text-ink"
          >
            Ver más de {collection.artist}
          </button>
        )}
      </div>
    </>
  );
}
