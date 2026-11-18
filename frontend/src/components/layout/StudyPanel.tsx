import { FileText, Languages, RefreshCw, Speech } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import { clamp } from "../../lib/format";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Button } from "../ui/Button";
import { IconButton } from "../ui/IconButton";
import { Chip } from "../ui/Chip";
import { LyricsView } from "../lyrics/LyricsView";
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
  const loadLyrics = useApp((state) => state.loadLyrics);
  const importLyrics = useApp((state) => state.importLyrics);
  const annotate = useApp((state) => state.annotate);

  const current = usePlayer((state) => state.current);
  const lines = lyrics?.lines ?? [];
  const hasGeneratedAnnotations = useMemo(
    () => lines.some((line) => line.annotation?.source === "ai"),
    [lines],
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
  const seekMs = usePlayer((state) => state.seekMs);

  const [showIPA, toggleIPA] = useLocalFlag("showIPA", true);
  const [showTranslation, toggleTranslation] = useLocalFlag("showTranslation", true);
  const [offsetMs, setOffsetMs] = useState(0);

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
      data-lyrics-lines={lines.length}
      data-lyrics-notice={lyrics?.notice ?? ""}
    >
      <div className="px-6 pt-5">
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 items-center gap-2">
            <p className="text-[17px] font-semibold tracking-[-0.02em] text-ink" title={syncLabel || undefined}>
              Letra
            </p>
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
            {lyrics?.available ? (
              <Button
                size="sm"
                variant={hasGeneratedAnnotations ? "subtle" : "primary"}
                className="ml-1 h-7 px-3 text-[12px]"
                icon={<Languages className="h-3.5 w-3.5" />}
                aria-label="Generar traducción"
                disabled={annotating}
                title="Generar la traducción de los versos con tu proveedor de IA"
                onClick={() => void annotate()}
              >
                {annotating ? "Generando…" : "Generar"}
              </Button>
            ) : null}
          </div>
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-1.5">
          <Chip
            active={showIPA}
            onClick={toggleIPA}
            label="Pronunciación"
            icon={<Speech className="h-[13px] w-[13px]" />}
          />
          <Chip
            active={showTranslation}
            onClick={toggleTranslation}
            label="Traducción"
            icon={<Languages className="h-[13px] w-[13px]" />}
          />
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
          onImport={() => void importLyrics()}
        />
      </div>

      <div className="flex items-center justify-end gap-2 border-t border-line/70 px-3 py-2">
        <span className="text-[10.5px] text-faint">Desfase</span>
        <button
          type="button"
          aria-label="Atrasar la letra 100 milisegundos"
          className="rounded px-1 text-[12px] text-muted hover:text-ink"
          onClick={() => setOffsetMs((value) => clamp(value - 100, -3000, 3000))}
        >
          −
        </button>
        <span className="w-12 text-center text-[12px] text-muted tnum">{offsetMs} ms</span>
        <button
          type="button"
          aria-label="Adelantar la letra 100 milisegundos"
          className="rounded px-1 text-[12px] text-muted hover:text-ink"
          onClick={() => setOffsetMs((value) => clamp(value + 100, -3000, 3000))}
        >
          +
        </button>
        {offsetMs !== 0 ? (
          <button
            type="button"
            aria-label="Restablecer desfase"
            className="text-[11px] text-faint hover:text-ink"
            onClick={() => setOffsetMs(0)}
          >
            Restablecer
          </button>
        ) : null}
      </div>
    </aside>
  );
}

