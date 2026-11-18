import { ChevronDown, Pause, Play, Repeat1, SkipBack, SkipForward } from "lucide-react";
import { useMemo } from "react";
import { activeLineIndex, loopFromLines } from "../../lib/timeline";
import { formatTime } from "../../lib/format";
import { useDuration, usePlaybackTime } from "../../player/clock";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { IconButton } from "../../components/ui/IconButton";
import { LyricsView } from "../../components/lyrics/LyricsView";
import { useLocalFlag } from "../../lib/useLocalFlag";
import { hiResArtwork } from "../../lib/artwork";

/** Two initials for the artwork placeholder. */
function initials(title: string): string {
  return title
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase() ?? "")
    .join("");
}

/**
 * Immersive study mode.
 *
 * Layout follows the reference player: the artwork and the song credits take
 * the left column, the lyric owns the right one, and the record itself tints
 * the background through a blurred copy of the cover. The active line is amber;
 * everything else stays quiet grey so the eye only has one place to land.
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
    <div
      className="fixed inset-0 z-30 flex flex-col overflow-hidden bg-[#0b0b0c] text-white"
      role="dialog"
      aria-modal="true"
      aria-label="Modo inmersivo"
      onKeyDown={(event) => {
        if (event.key === "Escape") setImmersive(false);
      }}
    >
      {/* The cover itself lights the room: a blurred, oversized copy behind
          everything, dimmed so text keeps its contrast. */}
      {current?.artworkUrl ? (
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 scale-125 opacity-40 blur-[110px]"
          style={{ backgroundImage: `url("${hiResArtwork(current.artworkUrl)}")`, backgroundSize: "cover", backgroundPosition: "center" }}
        />
      ) : null}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0 bg-[#0b0b0c]/72" />

      <header className="relative z-10 flex items-center justify-end gap-4 px-5 pt-4 sm:px-8">
        <div className="flex items-center gap-1.5">
          <button
            type="button"
            onClick={toggleIPA}
            aria-pressed={showIPA}
            className={`rounded-full border px-3 py-1 text-[11px] transition-colors ${
              showIPA ? "border-white/90 text-white" : "border-white/15 text-white/45 hover:text-white"
            }`}
          >
            IPA + ES
          </button>
          <button
            type="button"
            onClick={toggleTranslation}
            aria-pressed={showTranslation}
            className={`rounded-full border px-3 py-1 text-[11px] transition-colors ${
              showTranslation ? "border-white/90 text-white" : "border-white/15 text-white/45 hover:text-white"
            }`}
          >
            Traducción
          </button>
        </div>
      </header>

      <main className="relative z-10 flex min-h-0 flex-1 items-center gap-[clamp(32px,6vw,96px)] px-[clamp(20px,6vw,110px)] py-6">
        <section className="hidden w-[clamp(200px,26vw,460px)] shrink-0 self-center sm:block">
          <div className="aspect-square w-full overflow-hidden rounded-[12px] bg-white/[0.06] shadow-[var(--shadow-product)]">
            {current?.artworkUrl ? (
              <img src={hiResArtwork(current.artworkUrl)} alt="" className="h-full w-full object-cover" draggable={false} />
            ) : (
              <div className="flex h-full items-center justify-center text-[clamp(48px,6vw,96px)] font-semibold text-white/20">
                {current ? initials(current.title) : ""}
              </div>
            )}
          </div>
          <div className="mt-5 min-w-0">
            <h2 className="truncate text-[clamp(18px,1.7vw,23px)] font-semibold tracking-[-0.4px] text-white">
              {current?.title ?? "Sin canción"}
            </h2>
            <p className="mt-0.5 truncate text-[clamp(13px,1.1vw,15px)] text-white/55">
              {current?.artist ?? "Elige una canción"}
            </p>
          </div>
        </section>

        <section className="relative min-h-0 flex-1 self-stretch" aria-label="Letra">
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
        </section>
      </main>

      <footer className="relative z-10 flex items-center gap-3 px-[clamp(20px,6vw,110px)] pb-5">
        <div className="flex items-center gap-1">
          <IconButton label="Anterior" onClick={previous} disabled={!current}>
            <SkipBack className="h-4 w-4" />
          </IconButton>
          <button
            type="button"
            onClick={toggle}
            disabled={!current}
            aria-label={playing ? "Pausar" : "Reproducir"}
            className="mx-1 flex h-12 w-12 items-center justify-center rounded-full bg-white text-black transition-transform active:scale-[0.96] disabled:opacity-40"
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

        <span className="w-10 text-right text-[11px] text-white/40 tnum">{formatTime(timeMs)}</span>
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
        <span className="w-10 text-[11px] text-white/40 tnum">{formatTime(totalMs)}</span>
        <button
          type="button"
          onClick={() => setImmersive(false)}
          aria-label="Salir del modo inmersivo"
          title="Salir del modo inmersivo"
          className="ml-1 flex h-9 w-9 items-center justify-center rounded-full text-white/50 transition-colors hover:bg-white/10 hover:text-white"
        >
          <ChevronDown className="h-4 w-4" aria-hidden="true" />
        </button>
      </footer>
    </div>
  );
}
