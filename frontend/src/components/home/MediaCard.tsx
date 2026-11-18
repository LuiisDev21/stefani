import { Play } from "lucide-react";
import type { Track } from "../../lib/types";
import { cn } from "../../lib/cn";

interface MediaCardProps {
  track: Track;
  onPlay: () => void;
  className?: string;
}

/**
 * Cover-first card for a shelf of songs. The artwork is the content, so the
 * play control only appears on hover or focus; the labels stay below it in the
 * same two-line rhythm as a record sleeve.
 */
export function MediaCard({ track, onPlay, className }: MediaCardProps) {
  const initials = track.title
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase() ?? "")
    .join("");

  return (
    <button type="button" onClick={onPlay} className={cn("group w-full text-left", className)}>
      <div className="relative aspect-square w-full overflow-hidden rounded-[10px] border border-line bg-card">
        {track.artworkUrl ? (
          <img src={track.artworkUrl} alt="" loading="lazy" className="h-full w-full object-cover" draggable={false} />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <span className="text-[28px] font-semibold tracking-[-0.02em] text-ash">{initials}</span>
          </div>
        )}
        <span
          aria-hidden="true"
          className="absolute inset-0 flex items-center justify-center bg-black/45 opacity-0 transition-opacity duration-150 group-hover:opacity-100 group-focus-visible:opacity-100"
        >
          <span className="flex h-10 w-10 items-center justify-center rounded-full bg-accent text-on-accent">
            <Play className="ml-0.5 h-4 w-4 fill-current" />
          </span>
        </span>
      </div>
      <p className="mt-3 truncate text-[13px] leading-[1.4] text-ink">{track.title}</p>
      <p className="truncate text-[12px] leading-[1.4] text-ash">{track.artist || "Artista desconocido"}</p>
    </button>
  );
}
