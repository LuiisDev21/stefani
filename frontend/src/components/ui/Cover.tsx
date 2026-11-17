import { cn } from "../../lib/cn";

interface CoverProps {
  title: string;
  artist?: string;
  src?: string;
  size?: number;
  className?: string;
  rounded?: string;
  /** The design system reserves its one drop shadow for artwork on a surface. */
  shadow?: boolean;
}

/**
 * Artwork. Real artwork is used when the provider supplies it. Without it, the
 * cover becomes a typographic sleeve instead of a fake album: the title set in
 * the display face on parchment, with the artist underneath. At row sizes the
 * same idea reduces to the initials, so the list keeps its rhythm.
 */
export function Cover({ title, artist, src, size = 40, className, rounded = "rounded-[8px]", shadow }: CoverProps) {
  const large = size >= 120;
  return (
    <div
      aria-hidden="true"
      className={cn(
        "shrink-0 overflow-hidden",
        src ? "bg-elevated" : "border border-line-soft bg-elevated",
        rounded,
        shadow && "product-shadow",
        className,
      )}
      style={{ width: size, height: size }}
    >
      {src ? (
        <img src={src} alt="" loading="lazy" className="h-full w-full object-cover" />
      ) : large ? (
        <div className="flex h-full w-full flex-col justify-between p-5">
          <span className="text-[11px] text-faint">STEFANI</span>
          <span className="font-display text-[24px] font-semibold leading-[1.15] tracking-[-0.4px] text-ink">
            {title}
          </span>
          <span className="text-[12px] text-muted">{artist || "Sesión de estudio"}</span>
        </div>
      ) : (
        <div className="flex h-full w-full items-center justify-center">
          <span
            className="font-semibold leading-none tracking-[-0.02em] text-faint"
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
