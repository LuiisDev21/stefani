import { ExternalLink, Heart, ListPlus, Pause, Play } from "lucide-react";
import { cn } from "../../lib/cn";
import { formatTime } from "../../lib/format";
import type { Track } from "../../lib/types";
import { Cover } from "../ui/Cover";
import { IconButton } from "../ui/IconButton";

interface TrackRowProps {
  track: Track;
  index: number;
  active: boolean;
  playing: boolean;
  onPlay: () => void;
  onFavorite: () => void;
  onAdd?: () => void;
  onExternal?: () => void;
}

/**
 * One track row. Compact by design: the library is a list to scan, not a grid of
 * cards. The play affordance replaces the index on hover/focus so the row keeps a
 * single focal point.
 */
export function TrackRow({
  track,
  index,
  active,
  playing,
  onPlay,
  onFavorite,
  onAdd,
  onExternal,
}: TrackRowProps) {
  return (
    <li
      className={cn(
        "group grid grid-cols-[28px_minmax(0,1fr)_auto] items-center gap-3 rounded-[7px] px-2 py-[6px]",
        "transition-colors hover:bg-surface/50",
        active && "bg-surface/40",
      )}
    >
      <button
        type="button"
        onClick={onPlay}
        aria-label={`Reproducir ${track.title}`}
        className="relative flex h-7 w-7 items-center justify-center rounded text-[11px] text-faint tnum"
      >
        <span className={cn("group-hover:opacity-0", active && "opacity-0")}>{index + 1}</span>
        <span
          className={cn(
            "absolute inset-0 flex items-center justify-center opacity-0",
            "group-hover:opacity-100 group-focus-visible:opacity-100",
            active && "opacity-100",
          )}
        >
          {active && playing ? (
            <Pause className="h-3.5 w-3.5 fill-current text-accent" />
          ) : (
            <Play className="h-3.5 w-3.5 fill-current text-ink" />
          )}
        </span>
      </button>

      <button type="button" onClick={onPlay} className="grid min-w-0 grid-cols-[36px_minmax(0,1fr)] items-center gap-3 text-left">
        <Cover title={track.title} src={track.artworkUrl} size={36} />
        <span className="min-w-0">
          <span className="flex items-center gap-2">
            <span className={cn("truncate text-[13px]", active ? "font-medium text-accent" : "text-ink")}>
              {track.title}
            </span>
            {!track.playable ? (
              <span className="shrink-0 rounded-[3px] border border-line px-1.5 py-[1px] text-[9.5px] text-faint">
                catalogo
              </span>
            ) : null}
          </span>
          <span className="mt-[1px] block truncate text-[11.5px] text-muted">
            {track.artist || "Artista desconocido"}
            {track.album ? ` · ${track.album}` : ""}
          </span>
        </span>
      </button>

      <div className="flex items-center gap-0.5">
        <span className="mr-1 w-9 text-right text-[10.5px] text-faint tnum">
          {track.durationMs > 0 ? formatTime(track.durationMs) : "—"}
        </span>
        {onAdd ? (
          <IconButton label="Agregar a una lista" size="sm" onClick={onAdd} className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100">
            <ListPlus className="h-3.5 w-3.5" />
          </IconButton>
        ) : null}
        {onExternal && track.externalUrl ? (
          <IconButton label="Abrir en la app oficial" size="sm" onClick={onExternal} className="opacity-0 group-hover:opacity-100 focus-visible:opacity-100">
            <ExternalLink className="h-3.5 w-3.5" />
          </IconButton>
        ) : null}
        <IconButton
          label={track.favorite ? "Quitar de favoritos" : "Guardar en favoritos"}
          size="sm"
          onClick={onFavorite}
          className={cn(track.favorite && "text-accent hover:text-accent")}
        >
          <Heart className={cn("h-3.5 w-3.5", track.favorite && "fill-current")} />
        </IconButton>
      </div>
    </li>
  );
}
