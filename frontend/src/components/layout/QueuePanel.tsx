import { Mic } from "lucide-react";
import { formatTime } from "../../lib/format";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Cover } from "../ui/Cover";

/**
 * The right panel showing what plays next. Rows jump straight to a song, and
 * the mic returns to the lyrics of whatever is sounding.
 */
export function QueuePanel() {
  const queue = usePlayer((state) => state.queue);
  const index = usePlayer((state) => state.index);
  const current = usePlayer((state) => state.current);
  const openTrack = useApp((state) => state.openTrack);
  const setSidePanel = useApp((state) => state.setSidePanel);

  const upcoming = queue.slice(index + 1);

  return (
    <aside className="flex h-full w-[340px] shrink-0 flex-col border-l border-line" aria-label="Cola">
      <header className="flex items-center justify-between px-6 pt-5 pb-2">
        <p className="text-[17px] font-semibold tracking-[-0.02em] text-ink">Cola</p>
        <button
          type="button"
          onClick={() => setSidePanel("lyrics")}
          className="inline-flex h-8 w-8 items-center justify-center rounded-full text-muted transition-colors hover:text-ink"
          aria-label="Volver a la letra"
          title="Volver a la letra"
        >
          <Mic size={15} strokeWidth={2} />
        </button>
      </header>

      <div className="scroll-area min-h-0 flex-1 px-6 pb-8">
        {current ? (
          <>
            <p className="mt-2 text-[11px] text-muted">Sonando ahora</p>
            <div className="mt-2 flex items-center gap-3">
              <Cover title={current.title} src={current.artworkUrl} size={40} rounded="rounded-[6px]" />
              <div className="min-w-0">
                <p className="truncate text-[13px] font-medium text-ink">{current.title}</p>
                <p className="truncate text-[11.5px] text-muted">{current.artist || "Artista desconocido"}</p>
              </div>
            </div>

            <p className="mt-6 text-[11px] text-muted">A continuación</p>
            {upcoming.length === 0 ? (
              <p className="mt-2 text-[12.5px] text-faint">No queda nada en la cola.</p>
            ) : (
              <ul className="mt-2 space-y-0.5">
                {upcoming.map((track, offset) => (
                  <li key={`${track.id}-${index + 1 + offset}`}>
                    <button
                      type="button"
                      onClick={() => void openTrack(track, queue)}
                      className="flex w-full items-center gap-3 rounded-[8px] px-1.5 py-1.5 text-left transition-colors hover:bg-surface"
                      title={`Reproducir ${track.title}`}
                    >
                      <Cover title={track.title} src={track.artworkUrl} size={32} rounded="rounded-[5px]" />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-[12.5px] text-body">{track.title}</p>
                        <p className="truncate text-[11px] text-muted">{track.artist || "Artista desconocido"}</p>
                      </div>
                      {track.durationMs > 0 ? (
                        <span className="text-[11px] text-faint tnum">{formatTime(track.durationMs)}</span>
                      ) : null}
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </>
        ) : (
          <p className="mt-2 text-[12.5px] text-faint">Reproduce una canción para ver la cola.</p>
        )}
      </div>
    </aside>
  );
}
