import { cn } from "../../lib/cn";

interface CoverProps {
  title: string;
  artist?: string;
  src?: string;
  size?: number;
  className?: string;
  rounded?: string;
}

/**
 * Artwork. Real artwork is used when the provider supplies it; otherwise a
 * deterministic gradient is derived from the track title. It is clearly a
 * generated placeholder (no typography pretending to be an album cover), but it
 * keeps the layout stable and gives every row a stable identity.
 */
export function Cover({ title, src, size = 40, className, rounded = "rounded-[6px]" }: CoverProps) {
  const hue = hashHue(title);
  const style = {
    width: size,
    height: size,
    background: src
      ? undefined
      : `radial-gradient(120% 120% at 20% 15%, hsl(${hue} 55% 42%) 0%, hsl(${(hue + 40) % 360} 48% 22%) 55%, #0f1420 100%)`,
  };
  return (
    <div
      aria-hidden="true"
      className={cn("shrink-0 overflow-hidden border border-white/5", rounded, className)}
      style={style}
    >
      {src ? (
        <img src={src} alt="" loading="lazy" className="h-full w-full object-cover" />
      ) : (
        <div className="flex h-full w-full items-end justify-end p-1">
          <span
            className="text-[9px] font-medium uppercase tracking-[0.18em] text-white/55"
            style={{ fontSize: Math.max(7, size / 6) }}
          >
            {initials(title)}
          </span>
        </div>
      )}
    </div>
  );
}

function hashHue(text: string): number {
  let hash = 0;
  for (let i = 0; i < text.length; i += 1) {
    hash = (hash * 31 + text.charCodeAt(i)) % 360;
  }
  return hash;
}

function initials(title: string): string {
  const words = title.trim().split(/\s+/).slice(0, 2);
  return words.map((word) => word[0]?.toUpperCase() ?? "").join("");
}
