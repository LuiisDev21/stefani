import { ChevronDown, Pause, Play, SkipBack, SkipForward, Repeat1 } from "lucide-react";
import { useMemo } from "react";
import { activeLineIndex, loopFromLines } from "../../lib/timeline";
import { formatTime } from "../../lib/format";
import { useDuration, usePlaybackTime } from "../../player/clock";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { IconButton } from "../../components/ui/IconButton";
import { LyricsView } from "../../components/lyrics/LyricsView";
import { useLocalFlag } from "../../lib/useLocalFlag";

/**
 * Immersive study mode: the library and the list disappear, artwork becomes the
 * background and the lyric column owns the screen. Nothing is duplicated from the
 * side panel except the lyric renderer itself.
 */
export function ImmersiveView() {
  const setImmersive = useApp((state) => state.setImmersive);
  const lyrics = useApp((state) => state.lyrics);
  const loading = useApp((state) => state.lyricsLoading);
  const lookupWord = useApp((state) => state.lookupWord);
  const importLyrics = useApp((state) => state.importLyrics);

  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const toggle = usePlayer((state) => state.toggle);
  const next = usePlayer((state) => state.next);
  const previous = usePlayer((state) => state.previous);
  const seekMs = usePlayer((state) => state.seekMs);
  const setLoop = usePlayer((state) => state.setLoop);

  const timeMs = usePlaybackTime();
  const durationMs = useDuration();
  const totalMs = durationMs || current?.durationMs || 0;

  const [showIPA, toggleIPA] = useLocalFlag("showIPA", true);
  const [showTranslation, toggleTranslation] = useLocalFlag("showTranslation", true);

  const activeIndex = useMemo(
    () => (lyrics?.lines ? activeLineIndex(lyrics.lines, timeMs) : -1),
    [lyrics, timeMs],
  );

  const hue = useMemo(() => {
    const text = current?.title ?? "stefani";
    let hash = 0;
    for (let i = 0; i < text.length; i += 1) hash = (hash * 31 + text.charCodeAt(i)) % 360;
    return hash;
  }, [current?.title]);

  const repeatLine = () => {
    if (!lyrics?.lines || activeIndex < 0) return;
    const bounds = loopFromLines(lyrics.lines, activeIndex, activeIndex);
    if (!bounds) return;
    setLoop(bounds);
    seekMs(bounds.startMs);
  };

  return (
    <div className="fixed inset-0 z-30 flex flex-col bg-bg" role="dialog" aria-modal="true" aria-label="Modo inmersivo">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-0 opacity-70"
        style={{
          background: `radial-gradient(120% 90% at 15% 0%, hsl(${hue} 45% 26%) 0%, transparent 55%),
                       radial-gradient(120% 90% at 85% 10%, hsl(${(hue + 60) % 360} 38% 20%) 0%, transparent 60%),
                       linear-gradient(180deg, #0b0e16 0%, #0b0e16 100%)`,
        }}
      />

      <header className="relative z-10 flex items-start justify-between gap-4 px-6 pt-4">
        <div className="min-w-0">
          <p className="text-[11px] font-medium text-accent/90">Modo inmersivo</p>
          <h2 className="mt-1 truncate text-[20px] font-medium tracking-[-0.02em] text-ink">
            {current?.title ?? "Sin cancion"}
          </h2>
          <p className="truncate text-[12px] text-muted">{current?.artist ?? "Elige una cancion"}</p>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={toggleIPA}
            aria-pressed={showIPA}
            className={`rounded-[5px] border px-2 py-[3px] text-[10.5px] transition-colors ${
              showIPA ? "border-accent/40 bg-accent/10 text-accent" : "border-line text-faint hover:text-muted"
            }`}
          >
            IPA + ES
          </button>
          <button
            type="button"
            onClick={toggleTranslation}
            aria-pressed={showTranslation}
            className={`rounded-[5px] border px-2 py-[3px] text-[10.5px] transition-colors ${
              showTranslation ? "border-accent/40 bg-accent/10 text-accent" : "border-line text-faint hover:text-muted"
            }`}
          >
            Traduccion
          </button>
          <IconButton label="Salir del modo inmersivo" onClick={() => setImmersive(false)}>
            <ChevronDown className="h-4 w-4" />
          </IconButton>
        </div>
      </header>

      <div className="relative z-10 mx-auto min-h-0 w-full max-w-[900px] flex-1 px-6">
        <LyricsView
          lyrics={lyrics}
          loading={loading}
          offsetMs={0}
          showIPA={showIPA}
          showTranslation={showTranslation}
          immersive
          onSeek={seekMs}
          onWord={(word, line) => void lookupWord(word, line.id, line.text)}
          onImport={() => void importLyrics()}
        />
      </div>

      <footer className="relative z-10 flex items-center gap-4 border-t border-line/70 bg-bg/70 px-6 py-3 backdrop-blur-sm">
        <div className="flex items-center gap-1">
          <IconButton label="Anterior" onClick={previous} disabled={!current}>
            <SkipBack className="h-4 w-4" />
          </IconButton>
          <button
            type="button"
            onClick={toggle}
            disabled={!current}
            aria-label={playing ? "Pausar" : "Reproducir"}
            className="mx-1 flex h-10 w-10 items-center justify-center rounded-full bg-ink text-[#0b0e16] transition-colors hover:bg-white disabled:opacity-40"
          >
            {playing ? <Pause className="h-4 w-4 fill-current" /> : <Play className="ml-0.5 h-4 w-4 fill-current" />}
          </button>
          <IconButton label="Siguiente" onClick={next} disabled={!current}>
            <SkipForward className="h-4 w-4" />
          </IconButton>
          <IconButton label="Repetir el verso actual" onClick={repeatLine} disabled={activeIndex < 0 || !current?.playable}>
            <Repeat1 className="h-4 w-4" />
          </IconButton>
        </div>

        <span className="w-10 text-right text-[10.5px] text-faint tnum">{formatTime(timeMs)}</span>
        <input
          type="range"
          min={0}
          max={Math.max(totalMs, 1)}
          value={Math.min(timeMs, totalMs)}
          onChange={(event) => seekMs(Number(event.target.value))}
          aria-label="Progreso de la cancion"
          className="stefani-slider flex-1"
          style={{ ["--fill" as string]: `${totalMs > 0 ? (timeMs / totalMs) * 100 : 0}%` }}
          disabled={!current?.playable}
        />
        <span className="w-10 text-[10.5px] text-faint tnum">{formatTime(totalMs)}</span>
      </footer>
    </div>
  );
}
