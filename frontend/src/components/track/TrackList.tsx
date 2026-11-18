import { useMemo } from "react";
import { api } from "../../lib/api";
import type { Track } from "../../lib/types";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { TrackRow } from "./TrackRow";

interface TrackListProps {
  tracks: Track[];
  onAdd?: (track: Track) => void;
  onExternal?: (track: Track) => void;
}

/** Ordered track list wired to the player and the app store. */
export function TrackList({ tracks, onAdd, onExternal }: TrackListProps) {
  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const loading = usePlayer((state) => state.loading);
  const toggleFavorite = useApp((state) => state.toggleFavorite);
  const openTrack = useApp((state) => state.openTrack);

  const ids = useMemo(() => tracks.map((track) => track.id).join("|"), [tracks]);

  return (
    <ol className="space-y-0.5" aria-label="Canciones" data-ids={ids}>
      {tracks.map((track, index) => (
        <TrackRow
          key={`${track.id}-${index}`}
          track={track}
          index={index}
          active={current?.id === track.id}
          playing={playing}
          loading={loading && current?.id === track.id}
          onPlay={() => {
            if (!track.unavailable) void openTrack(track, tracks);
          }}
          onFavorite={() => void toggleFavorite(track)}
          onAdd={onAdd ? () => onAdd(track) : undefined}
          onExternal={
            onExternal
              ? () => onExternal(track)
              : track.externalUrl
                ? () => void api.openExternal(track.externalUrl).catch(() => undefined)
                : undefined
          }
        />
      ))}
    </ol>
  );
}
