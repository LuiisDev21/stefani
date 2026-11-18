import { cn } from "../../lib/cn";

interface CoverProps {
  title: string;
  artist?: string;
  src?: string;
  size?: number;
  className?: string;
  rounded?: string;
  /** Kept for call sites: the system has no drop shadows, so it no longer draws one. */
  shadow?: boolean;
}

/**
 * Artwork. Real artwork is used when the provider supplies it. Without it, the
 * cover becomes a typographic sleeve instead of a fake album: the title set in
 * the display face on the card surface, with the artist underneath. At row
 * sizes the same idea reduces to the initials, so the list keeps its rhythm.
 */
export function Cover({ title, artist, src, size = 40, className, rounded = "rounded-[8px]" }: CoverProps) {
  const large = size >= 120;
  return (
    <div
      aria-hidden="true"
      className={cn("shrink-0 overflow-hidden border border-line bg-card", rounded, className)}
      style={{ width: size, height: size }}
    >
      {src ? (
        <img src={src} alt="" loading="lazy" className="h-full w-full object-cover" />
      ) : large ? (
        <div className="flex h-full w-full flex-col justify-between p-5">
          <span className="text-[10px] tracking-[0.02em] text-ash">STEFANI</span>
          <span className="font-display text-[30px] font-semibold leading-[1.08] tracking-[-0.035em] text-ink">
            {title}
          </span>
          <span className="text-[12px] text-muted">{artist || "Sin artista"}</span>
        </div>
      ) : (
        <div className="flex h-full w-full items-center justify-center">
          <span
            className="font-semibold leading-none tracking-[-0.02em] text-ash"
            style={{ fontSize: Math.max(9, size / 3.4) }}
          >
            {initials(title)}
          </span>
        </div>
      )}
    </div>
  );
}

function initials(title: string): string {
  const words = title.trim().split(/\s+/).slice(0, 2);
  return words.map((word) => word[0]?.toUpperCase() ?? "").join("");
}
