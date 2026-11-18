import { Heart } from "lucide-react";
import { useApp } from "../../stores/appStore";
import { EmptyState } from "../../components/ui/EmptyState";
import { plural } from "../../lib/format";
import { TopBar } from "../../components/layout/TopBar";
import { TrackList } from "../../components/track/TrackList";

/** Favourite tracks. */
export function FavoritesView() {
  const favorites = useApp((state) => state.favorites);
  const openPicker = useApp((state) => state.openPicker);

  return (
    <div className="min-h-0">
      <TopBar title="Favoritos" subtitle={favorites.length > 0 ? plural(favorites.length, "canción", "canciones") : undefined} />
      <div className="px-7 pb-10">
        {favorites.length === 0 ? (
          <EmptyState
            icon={<Heart className="h-5 w-5" />}
            title="Todavía no marcaste favoritos"
            description="El corazon de cada fila guarda la canción en favoritos, sin moverla de tu biblioteca."
          />
        ) : (
          <TrackList tracks={favorites} onAdd={openPicker} />
        )}
      </div>
    </div>
  );
}
