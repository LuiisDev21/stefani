import { FileText, PenLine, RefreshCw, Repeat1, X } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { cn } from "../../lib/cn";
import { clamp } from "../../lib/format";
import { activeLineIndex, loopFromLines } from "../../lib/timeline";
import { usePlaybackTime } from "../../player/clock";
import { capabilitiesFor } from "../../player/capabilities";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { api } from "../../lib/api";
import { Button } from "../ui/Button";
import { IconButton } from "../ui/IconButton";
import { LyricsView } from "../lyrics/LyricsView";
import { WordCard } from "../lyrics/WordCard";
import { useLocalFlag } from "../../lib/useLocalFlag";

/**
 * Right panel: lyrics, pronunciation and the controls that belong to study.
 *
 * It is contextual, never a dashboard: the header reports the document's real
 * sync level, and every control maps to something the backend can actually do
 * for this track.
 */
export function StudyPanel() {
  const lyrics = useApp((state) => state.lyrics);
  const loading = useApp((state) => state.lyricsLoading);
  const annotating = useApp((state) => state.annotating);
  const progress = useApp((state) => state.annotateProgress);
  const wordPanel = useApp((state) => state.wordPanel);
  const vocabulary = useApp((state) => state.vocabulary);
  const loadLyrics = useApp((state) => state.loadLyrics);
  const importLyrics = useApp((state) => state.importLyrics);
  const annotate = useApp((state) => state.annotate);
  const lookupWord = useApp((state) => state.lookupWord);
  const closeWordPanel = useApp((state) => state.closeWordPanel);
  const saveWord = useApp((state) => state.saveWord);
  const notify = useApp((state) => state.notify);

  const current = usePlayer((state) => state.current);
  const hasGeneratedAnnotations = useMemo(
    () => Boolean(lyrics?.lines.some((line) => line.annotation?.source === "ai")),
    [lyrics],
  );
  // A track can start playing from anywhere (search, queue, immersive view):
  // the panel is the single owner of loading its lyric, once per track.
  const requested = useRef<string | null>(null);
  useEffect(() => {
    if (!current?.id) return;
    if (lyrics?.trackId === current.id) {
      requested.current = current.id;
      return;
    }
    if (requested.current === current.id) return;
    requested.current = current.id;
    void loadLyrics(current.id);
  }, [current?.id, lyrics?.trackId, loadLyrics]);
  const loop = usePlayer((state) => state.loop);
  const setLoop = usePlayer((state) => state.setLoop);
  const seekMs = usePlayer((state) => state.seekMs);
  const caps = capabilitiesFor(current);

  const [showIPA, toggleIPA] = useLocalFlag("showIPA", true);
  const [showTranslation, toggleTranslation] = useLocalFlag("showTranslation", true);
  const [offsetMs, setOffsetMs] = useState(0);

  const timeMs = usePlaybackTime();
  const activeIndex = useMemo(
    () => (lyrics?.lines ? activeLineIndex(lyrics.lines, Math.max(0, timeMs - offsetMs)) : -1),
    [lyrics, timeMs, offsetMs],
  );

  const savedWords = useMemo(
    () => new Set(vocabulary.map((item) => item.word.toLowerCase())),
    [vocabulary],
  );

  const repeatLine = () => {
    if (!lyrics?.lines || activeIndex < 0) {
      notify("info", "Reproduce la canción para saber que verso estas escuchando");
      return;
    }
    const bounds = loopFromLines(lyrics.lines, activeIndex, activeIndex);
    if (!bounds) return;
    setLoop(bounds);
    seekMs(bounds.startMs);
    if (current) {
      void api.recordStudy({ trackId: current.id, seconds: 0, lines: 1 }).catch(() => undefined);
    }
  };

  const speak = (text: string) => {
    try {
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.lang = "en-US";
      utterance.rate = 0.9;
      window.speechSynthesis.cancel();
      window.speechSynthesis.speak(utterance);
    } catch {
      notify("info", "La sintesis de voz no esta disponible en este sistema");
    }
  };

  const syncLabel = lyrics?.available
    ? lyrics.hasWords
      ? "palabra por palabra"
      : lyrics.sync === "UNSYNCED"
        ? "sin sincronizar"
        : "por verso"
    : "";

  return (
    <aside
      className="flex h-full w-[400px] shrink-0 flex-col bg-bg"
      data-lyrics-state={lyrics ? (lyrics.available ? "ready" : "empty") : "none"}
      data-lyrics-lines={lyrics?.lines.length ?? 0}
      data-lyrics-notice={lyrics?.notice ?? ""}
    >
      <div className="px-6 pt-5">
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 items-center gap-3">
            <p className="text-[17px] font-semibold tracking-[-0.02em] text-ink">Letra</p>
            {syncLabel ? <span className="truncate text-[11px] text-muted">{syncLabel}</span> : null}
            {!lyrics?.available && loading && current ? (
              <span className="truncate text-[11px] text-muted">Buscando…</span>
            ) : null}
          </div>
          <div className="flex items-center gap-0.5">
            <IconButton
              label="Volver a buscar la letra"
              size="sm"
              disabled={!current}
              onClick={() => current && void loadLyrics(current.id, true)}
            >
              <RefreshCw className="h-3.5 w-3.5" />
            </IconButton>
            <IconButton label="Importar archivo .lrc" size="sm" disabled={!current} onClick={() => void importLyrics()}>
              <FileText className="h-3.5 w-3.5" />
            </IconButton>
          </div>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-1.5">
          <Chip active={showIPA} onClick={toggleIPA} label="IPA + ES" />
          <Chip active={showTranslation} onClick={toggleTranslation} label="Traducción" />
          {lyrics?.available ? (
            <Button
              size="sm"
              variant={hasGeneratedAnnotations ? "subtle" : "primary"}
              className="ml-auto h-7 px-3 text-[12px]"
              icon={<PenLine className="h-3.5 w-3.5" />}
              aria-label="Anotar con IA"
              disabled={annotating}
              title="Traducir y anotar los versos con tu proveedor de IA"
              onClick={() => void annotate()}
            >
              {annotating ? "Anotando…" : "Anotar"}
            </Button>
          ) : null}
        </div>

        {annotating && progress && progress.total > 0 ? (
          <div className="mt-2" aria-live="polite">
            <div className="h-[2px] w-full overflow-hidden rounded bg-line">
              <div
                className="h-full bg-accent transition-[width] duration-300"
                style={{ width: `${clamp((progress.done / progress.total) * 100, 0, 100)}%` }}
              />
            </div>
            <p className="mt-1 text-[10.5px] text-faint tnum">
              {progress.done} de {progress.total} versos
            </p>
          </div>
        ) : null}
      </div>

      <div className="mt-4 min-h-0 flex-1 px-6">
        <LyricsView
          lyrics={lyrics}
          loading={loading}
          offsetMs={offsetMs}
          showIPA={showIPA}
          showTranslation={showTranslation}
          onSeek={seekMs}
          onWord={(word, line) => void lookupWord(word, line.id, line.text)}
          onImport={() => void importLyrics()}
        />
      </div>

      {wordPanel ? (
        <WordCard
          word={wordPanel.word}
          entry={wordPanel.entry}
          loading={wordPanel.loading}
          error={wordPanel.error}
          alreadySaved={savedWords.has((wordPanel.entry?.word || wordPanel.word).toLowerCase())}
          onSave={(entry) => void saveWord(entry)}
          onClose={closeWordPanel}
          onSpeak={speak}
        />
      ) : (
        <div className="flex items-center justify-between gap-2 border-t border-line/70 px-3 py-2">
          <div className="flex min-w-0 items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              icon={<Repeat1 className="h-3.5 w-3.5" />}
              onClick={repeatLine}
              disabled={!lyrics?.available || !caps.abLoop}
            >
              Repetir verso
            </Button>
            {loop ? (
              <button
                type="button"
                onClick={() => setLoop(null)}
                className="flex items-center gap-1 truncate rounded-[5px] border border-accent/35 px-1.5 py-[3px] text-[10.5px] text-accent"
                aria-label="Dejar de repetir el fragmento"
              >
                <span className="truncate">Repitiendo</span>
                <X className="h-3 w-3 shrink-0" />
              </button>
            ) : null}
          </div>
          <div className="flex items-center gap-1">
            <span className="text-[10.5px] text-faint">Desfase</span>
            <button
              type="button"
              aria-label="Adelantar la letra 100 milisegundos"
              className="rounded px-1 text-[12px] text-muted hover:text-ink"
              onClick={() => setOffsetMs((value) => clamp(value + 100, -3000, 3000))}
            >
              −
            </button>
            <span className="w-12 text-center text-[12px] text-muted tnum">{offsetMs} ms</span>
            <button
              type="button"
              aria-label="Atrasar la letra 100 milisegundos"
              className="rounded px-1 text-[12px] text-muted hover:text-ink"
              onClick={() => setOffsetMs((value) => clamp(value - 100, -3000, 3000))}
            >
              +
            </button>
            {offsetMs !== 0 ? (
              <button
                type="button"
                aria-label="Restablecer desfase"
                className="text-faint hover:text-ink"
                onClick={() => setOffsetMs(0)}
              >
                <X className="h-3 w-3" />
              </button>
            ) : null}
          </div>
        </div>
      )}
    </aside>
  );
}

function Chip({ active, onClick, label }: { active: boolean; onClick: () => void; label: string }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "rounded-full border px-2.5 py-[3px] text-[11.5px] transition-colors",
        active
          ? "border-accent bg-accent/6 text-accent"
          : "border-line bg-transparent text-muted hover:bg-elevated hover:text-ink",
      )}
    >
      {label}
    </button>
  );
}
