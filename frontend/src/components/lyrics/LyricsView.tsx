import { ArrowDown, LoaderCircle, Music4 } from "lucide-react";
import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import { clock, usePlaybackTime } from "../../player/clock";
import { activeLineIndex } from "../../lib/timeline";
import type { Line, Lyrics } from "../../lib/types";
import { EmptyState } from "../ui/EmptyState";
import { Button } from "../ui/Button";
import { LyricLine } from "./LyricLine";

interface LyricsViewProps {
  lyrics: Lyrics | null;
  loading: boolean;
  offsetMs: number;
  showIPA: boolean;
  showTranslation: boolean;
  immersive?: boolean;
  onSeek: (ms: number) => void;
  onImport: () => void;
}

/**
 * The active line index, without subscribing the whole column to the clock.
 * The clock publishes at animation-frame rate; here only a change of index
 * re-renders the list, and the active line subscribes on its own.
 */
function useActiveLineIndex(lines: Line[], offsetMs: number): number {
  return useSyncExternalStore(
    clock.subscribe,
    () => (lines.length === 0 ? -1 : activeLineIndex(lines, Math.max(0, clock.getSnapshot().timeMs - offsetMs))),
    () => -1,
  );
}

/** The only line that follows the clock: the one being sung. */
function ActiveLine(props: Omit<React.ComponentProps<typeof LyricLine>, "timeMs"> & { offsetMs: number }) {
  const { offsetMs, ...rest } = props;
  // 30 fps: karaoke text is not readable faster, and half the paints is real
  // headroom on the WebView2 renderer.
  const raw = usePlaybackTime();
  const timeMs = Math.floor(raw / 33) * 33;
  return <LyricLine {...rest} timeMs={Math.max(0, timeMs - offsetMs)} />;
}

/**
 * Scrollable lyric column.
 *
 * Autoscroll follows the active line with an eased animation; a wheel, touch,
 * keyboard or scrollbar movement pauses it and offers a "follow again" control.
 */
export function LyricsView({
  lyrics,
  loading,
  offsetMs,
  showIPA,
  showTranslation,
  immersive,
  onSeek,
  onImport,
}: LyricsViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const lineRefs = useRef(new Map<string, HTMLLIElement>());
  const [manual, setManual] = useState(false);
  const manualTimer = useRef<number | null>(null);
  const scrollTimer = useRef<number | null>(null);
  const programmatic = useRef(false);

  const lines = lyrics?.lines ?? [];
  const activeIndex = useActiveLineIndex(lines, offsetMs);

  const pauseAutoscroll = useCallback(() => {
    setManual(true);
    if (manualTimer.current) window.clearTimeout(manualTimer.current);
    manualTimer.current = window.setTimeout(() => setManual(false), 6000);
  }, []);

  useEffect(() => {
    return () => {
      if (manualTimer.current) window.clearTimeout(manualTimer.current);
      if (scrollTimer.current) window.clearTimeout(scrollTimer.current);
    };
  }, []);

  useEffect(() => {
    if (manual || activeIndex < 0) return;
    const line = lines[activeIndex];
    const node = line ? lineRefs.current.get(line.id) : null;
    const container = containerRef.current;
    if (!node || !container) return;

    const target = Math.max(0, node.offsetTop - container.clientHeight * (immersive ? 0.42 : 0.32));
    // The browser animates this on the compositor: writing scrollTop per frame
    // forces a layout on the main thread and is what made the view feel stuck.
    programmatic.current = true;
    container.scrollTo({ top: target, behavior: "smooth" });
    if (scrollTimer.current) window.clearTimeout(scrollTimer.current);
    scrollTimer.current = window.setTimeout(() => {
      programmatic.current = false;
    }, 900);
  }, [activeIndex, manual, lines, immersive]);

  const handleScrollKey = useCallback(
    (event: React.KeyboardEvent<HTMLDivElement>) => {
      if (["ArrowDown", "ArrowUp", "PageDown", "PageUp", "Home", "End"].includes(event.key)) {
        pauseAutoscroll();
      }
    },
    [pauseAutoscroll],
  );

  if (loading) {
    return (
      <div className="space-y-3 px-1 py-2" aria-busy="true" aria-label="Cargando letra">
        {[88, 72, 94, 64, 80, 68].map((width, index) => (
          <div
            key={index}
            className="h-6 animate-pulse rounded bg-surface/70"
            style={{ width: `${width}%` }}
          />
        ))}
      </div>
    );
  }

  if (!lyrics || !lyrics.available || lines.length === 0) {
    return (
      <EmptyState
        icon={<Music4 className="h-5 w-5" aria-hidden="true" />}
        title="Todavía no hay letra"
        description={lyrics?.notice || "Elige una canción para cargar su letra, o importa un archivo .lrc."}
        actions={
          <Button size="sm" variant="subtle" onClick={onImport}>
            Importar .lrc
          </Button>
        }
      />
    );
  }

  return (
    <div className="relative h-full">
      {manual ? (
        <div className="pointer-events-none absolute inset-x-0 top-0 z-20 flex justify-center pt-1">
          <Button
            size="sm"
            variant="subtle"
            icon={<ArrowDown className="h-3.5 w-3.5" />}
            className="pointer-events-auto border border-line bg-card"
            onClick={() => setManual(false)}
          >
            Seguir la letra
          </Button>
        </div>
      ) : null}
      <div
        ref={containerRef}
        className={immersive ? "scroll-area scrollbar-hidden h-full pr-2" : "scroll-area h-full pr-2"}
        tabIndex={0}
        onKeyDown={handleScrollKey}
        onWheel={pauseAutoscroll}
        onTouchMove={pauseAutoscroll}
        onScroll={() => {
          if (!programmatic.current) pauseAutoscroll();
        }}
        aria-label="Letra sincronizada"
      >
        <ol className={immersive ? "space-y-5 pb-40 pt-24" : "space-y-1 pb-24 pt-2"}>
          {lines.map((line, index) => {
            const isActive = index === activeIndex;
            const past = index < activeIndex;
            return (
              <div
                key={line.id}
                ref={(node) => {
                  if (node) lineRefs.current.set(line.id, node as unknown as HTMLLIElement);
                  else lineRefs.current.delete(line.id);
                }}
              >
                {isActive ? (
                  <ActiveLine
                    line={line}
                    active
                    past={false}
                    offsetMs={offsetMs}
                    showIPA={showIPA}
                    showTranslation={showTranslation}
                    immersive={immersive}
                    onSeek={onSeek}
                  />
                ) : (
                  <LyricLine
                    line={line}
                    active={false}
                    past={past}
                    near={Math.abs(index - activeIndex) === 1}
                    showIPA={showIPA}
                    showTranslation={showTranslation}
                    immersive={immersive}
                    onSeek={onSeek}
                  />
                )}
              </div>
            );
          })}
        </ol>
      </div>
      {loading ? (
        <div className="absolute inset-x-0 top-1 flex justify-center">
          <LoaderCircle className="h-4 w-4 animate-spin text-faint" aria-hidden="true" />
        </div>
      ) : null}
    </div>
  );
}
