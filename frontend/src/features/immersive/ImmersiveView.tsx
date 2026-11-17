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

  const repeatLine = () => {
    if (!lyrics?.lines || activeIndex < 0) return;
    const bounds = loopFromLines(lyrics.lines, activeIndex, activeIndex);
    if (!bounds) return;
    setLoop(bounds);
    seekMs(bounds.startMs);
  };

  return (
    <div className="fixed inset-0 z-30 flex flex-col bg-black text-white" role="dialog" aria-modal="true" aria-label="Modo inmersivo">

      <header className="relative z-10 flex items-start justify-between gap-4 px-6 pt-4">
        <div className="min-w-0">
          <p className="text-[12px] font-medium text-white/45">Modo inmersivo</p>
          <h2 className="mt-1 truncate text-[21px] font-semibold tracking-[-0.3px] text-white">
            {current?.title ?? "Sin canción"}
          </h2>
          <p className="truncate text-[13px] text-white/55">{current?.artist ?? "Elige una canción"}</p>
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={toggleIPA}
            aria-pressed={showIPA}
            className={`rounded-full border px-3 py-1 text-[12px] transition-colors ${
              showIPA ? "border-white bg-white text-black" : "border-white/20 text-white/60 hover:text-white"
            }`}
          >
            IPA + ES
          </button>
          <button
            type="button"
            onClick={toggleTranslation}
            aria-pressed={showTranslation}
            className={`rounded-full border px-3 py-1 text-[12px] transition-colors ${
              showTranslation ? "border-white bg-white text-black" : "border-white/20 text-white/60 hover:text-white"
            }`}
          >
            Traducción
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

      <footer className="relative z-10 flex items-center gap-4 border-t border-white/15 bg-black/70 px-6 py-3">
        <div className="flex items-center gap-1">
          <IconButton label="Anterior" onClick={previous} disabled={!current}>
            <SkipBack className="h-4 w-4" />
          </IconButton>
          <button
            type="button"
            onClick={toggle}
            disabled={!current}
            aria-label={playing ? "Pausar" : "Reproducir"}
            className="mx-1 flex h-10 w-10 items-center justify-center rounded-full bg-white text-black transition-colors hover:bg-white/85 disabled:opacity-40"
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

        <span className="w-10 text-right text-[11px] text-white/45 tnum">{formatTime(timeMs)}</span>
        <input
          type="range"
          min={0}
          max={Math.max(totalMs, 1)}
          value={Math.min(timeMs, totalMs)}
          onChange={(event) => seekMs(Number(event.target.value))}
          aria-label="Progreso de la canción"
          className="stefani-slider on-dark flex-1"
          style={{ ["--fill" as string]: `${totalMs > 0 ? (timeMs / totalMs) * 100 : 0}%` }}
          disabled={!current?.playable}
        />
        <span className="w-10 text-[11px] text-white/45 tnum">{formatTime(totalMs)}</span>
      </footer>
    </div>
  );
}
