import { useApp } from "../../stores/appStore";
import { Cover } from "../ui/Cover";
import { SectionTitle } from "../ui/SectionTitle";
import type { Shelf } from "../../lib/types";

/**
 * One personalised row of the home feed. It scrolls sideways and keeps the
 * card sizes of the rest of the app: square cover, two lines of title, one of
 * subtitle.
 */
export function ShelfRow({ shelf }: { shelf: Shelf }) {
  const playShelfItem = useApp((state) => state.playShelfItem);

  return (
    <section className="mt-10" aria-label={shelf.title}>
      <SectionTitle title={shelf.title} />
      <div className="scrollbar-hidden -mx-1 flex snap-x gap-4 overflow-x-auto px-1 pb-1">
        {shelf.items.map((item, index) => (
          <button
            key={`${item.videoId || item.listId}-${index}`}
            type="button"
            onClick={() => void playShelfItem(item)}
            className="group w-[152px] shrink-0 snap-start text-left"
            title={`Reproducir ${item.title}`}
            aria-label={`Reproducir ${item.title}`}
          >
            <div className="overflow-hidden rounded-[10px] border border-line bg-card">
              <Cover title={item.title} src={item.cover} size={152} rounded="rounded-[0px]" />
            </div>
            <p className="mt-2 line-clamp-2 text-[12.5px] leading-[1.35] text-body transition-colors group-hover:text-ink">
              {item.title}
            </p>
            <p className="mt-0.5 truncate text-[11.5px] text-muted">{item.subtitle}</p>
          </button>
        ))}
      </div>
    </section>
  );
}
