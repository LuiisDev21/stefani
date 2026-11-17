import { ArrowDown, LoaderCircle, Music4 } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { usePlaybackTime } from "../../player/clock";
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
  onWord: (word: string, line: Line) => void;
  onImport: () => void;
}

/**
 * Scrollable lyric column.
 *
 * Autoscroll follows the active line, but the moment a wheel, touch or keyboard
 * scroll happens the view stops moving and offers a "follow again" control.
 * Nothing here owns a timer: position comes from the shared clock.
 */
export function LyricsView({
  lyrics,
  loading,
  offsetMs,
  showIPA,
  showTranslation,
  immersive,
  onSeek,
  onWord,
  onImport,
}: LyricsViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const lineRefs = useRef(new Map<string, HTMLLIElement>());
  const [manual, setManual] = useState(false);
  const manualTimer = useRef<number | null>(null);
  const timeMs = usePlaybackTime();

  const lines = lyrics?.lines ?? [];
  const effectiveTime = Math.max(0, timeMs - offsetMs);
  const activeIndex = useMemo(() => activeLineIndex(lines, effectiveTime), [lines, effectiveTime]);

  const pauseAutoscroll = useCallback(() => {
    setManual(true);
    if (manualTimer.current) window.clearTimeout(manualTimer.current);
    manualTimer.current = window.setTimeout(() => setManual(false), 6000);
  }, []);

  useEffect(() => {
    return () => {
      if (manualTimer.current) window.clearTimeout(manualTimer.current);
    };
  }, []);

  useEffect(() => {
    if (manual || activeIndex < 0) return;
    const line = lines[activeIndex];
    const node = line ? lineRefs.current.get(line.id) : null;
    const container = containerRef.current;
    if (!node || !container) return;
    const target = node.offsetTop - container.clientHeight * (immersive ? 0.38 : 0.32);
    container.scrollTo({ top: Math.max(0, target), behavior: "smooth" });
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
        title="Todavia no hay letra"
        description={lyrics?.notice || "Elige una cancion para cargar su letra, o importa un archivo .lrc."}
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
            className="pointer-events-auto shadow-[0_8px_20px_rgba(0,0,0,0.45)]"
            onClick={() => setManual(false)}
          >
            Seguir la letra
          </Button>
        </div>
      ) : null}
      <div
        ref={containerRef}
        className="scroll-area h-full pr-2"
        tabIndex={0}
        onKeyDown={handleScrollKey}
        onWheel={pauseAutoscroll}
        onTouchMove={pauseAutoscroll}
        aria-label="Letra sincronizada"
      >
        <ol className={immersive ? "space-y-3 pb-40 pt-24" : "space-y-1 pb-24 pt-2"}>
          {lines.map((line, index) => {
            const isActive = index === activeIndex;
            return (
              <div
                key={line.id}
                ref={(node) => {
                  if (node) lineRefs.current.set(line.id, node as unknown as HTMLLIElement);
                  else lineRefs.current.delete(line.id);
                }}
              >
                <LyricLine
                  line={line}
                  active={isActive}
                  past={index < activeIndex}
                  timeMs={isActive ? effectiveTime : undefined}
                  showIPA={showIPA}
                  showTranslation={showTranslation}
                  immersive={immersive}
                  onSeek={onSeek}
                  onWord={onWord}
                />
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
