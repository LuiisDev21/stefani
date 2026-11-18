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
  const openCollection = useApp((state) => state.openCollection);

  return (
    <section className="mt-10" aria-label={shelf.title}>
      <SectionTitle title={shelf.title} />
      <div className="scrollbar-hidden -mx-1 flex snap-x gap-4 overflow-x-auto px-1 pb-1">
        {shelf.items.map((item, index) => (
          <div key={`${item.videoId || item.listId}-${index}`} className="group w-[152px] shrink-0 snap-start">
            <button
              type="button"
              onClick={() => void playShelfItem(item)}
              className="block w-full text-left"
              title={
                item.listId ? `Abrir ${item.title}` : `Reproducir ${item.title}`
              }
              aria-label={item.listId ? `Abrir ${item.title}` : `Reproducir ${item.title}`}
            >
              <div className="overflow-hidden rounded-[10px] border border-line bg-card">
                <Cover title={item.title} src={item.cover} size={152} rounded="rounded-[0px]" />
              </div>
              <p className="mt-2 line-clamp-2 text-[12.5px] leading-[1.35] text-body transition-colors group-hover:text-ink">
                {item.title}
              </p>
            </button>
            {item.artistId ? (
              <button
                type="button"
                onClick={() => void openCollection(item.artistId)}
                className="mt-0.5 block max-w-full truncate text-left text-[11.5px] text-muted transition-colors hover:text-ink"
                title={`Abrir ${item.subtitle}`}
              >
                {item.subtitle}
              </button>
            ) : (
              <p className="mt-0.5 truncate text-[11.5px] text-muted">{item.subtitle}</p>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
