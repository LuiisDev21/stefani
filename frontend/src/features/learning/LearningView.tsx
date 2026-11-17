import { GraduationCap, Play, Trash2, Volume2 } from "lucide-react";
import { formatRelativeDate, formatStudyTime } from "../../lib/format";
import { api } from "../../lib/api";
import { useApp } from "../../stores/appStore";
import { Button } from "../../components/ui/Button";
import { EmptyState } from "../../components/ui/EmptyState";
import { IconButton } from "../../components/ui/IconButton";
import { SectionTitle } from "../../components/ui/SectionTitle";
import { TopBar } from "../../components/layout/TopBar";

/** Study progress and saved vocabulary, with the actions that belong to them. */
export function LearningView() {
  const progress = useApp((state) => state.progress);
  const vocabulary = useApp((state) => state.vocabulary);
  const deleteWord = useApp((state) => state.deleteWord);
  const openTrack = useApp((state) => state.openTrack);
  const go = useApp((state) => state.go);
  const notify = useApp((state) => state.notify);

  const totalSeconds = progress.reduce((sum, item) => sum + item.studySeconds, 0);
  const totalLines = progress.reduce((sum, item) => sum + item.linesPracticed, 0);

  const study = async (trackId: string) => {
    try {
      const track = await api.track(trackId);
      await openTrack(track);
      go("home");
    } catch {
      notify("error", "No pude abrir esa cancion");
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
      notify("info", "La sintesis de voz no esta disponible");
    }
  };

  return (
    <div className="min-h-0">
      <TopBar title="Aprendizaje" subtitle="Tiempo de practica, versos trabajados y vocabulario guardado" />

      <div className="px-7 pb-10">
        <dl className="flex flex-wrap items-baseline gap-x-10 gap-y-3">
          <Stat label="Tiempo de practica" value={formatStudyTime(totalSeconds)} />
          <Stat label="Versos trabajados" value={String(totalLines)} />
          <Stat label="Canciones con progreso" value={String(progress.length)} />
          <Stat label="Palabras guardadas" value={String(vocabulary.length)} />
        </dl>

        <div className="signal-line my-6" />

        <section>
          <SectionTitle title="Progreso por cancion" />
          {progress.length === 0 ? (
            <EmptyState
              icon={<GraduationCap className="h-5 w-5" />}
              title="Sin sesiones registradas"
              description="Escucha una cancion con la letra abierta: el tiempo se acumula automaticamente mientras suena, y cada verso que repites suma practica."
              actions={
                <Button size="sm" variant="primary" onClick={() => go("home")}>
                  Empezar con la demo
                </Button>
              }
            />
          ) : (
            <ul className="divide-y divide-line/60">
              {progress.map((item) => (
                <li key={item.trackId} className="flex items-center gap-3 py-2.5">
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] text-ink">{item.title || item.trackId}</p>
                    <p className="truncate text-[11.5px] text-faint tnum">
                      {formatStudyTime(item.studySeconds)} · {item.linesPracticed} versos · {item.wordsSaved} palabras
                      {" · "}
                      {formatRelativeDate(item.lastStudiedAt)}
                    </p>
                  </div>
                  <span className="hidden text-[11px] text-faint sm:block">
                    {item.status === "known" ? "conocida" : item.status === "new" ? "nueva" : "en curso"}
                  </span>
                  <Button
                    size="sm"
                    variant="ghost"
                    icon={<Play className="h-3.5 w-3.5 fill-current" />}
                    onClick={() => void study(item.trackId)}
                  >
                    Estudiar
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </section>

        <div className="signal-line my-6" />

        <section>
          <SectionTitle title="Vocabulario" hint="Palabras que guardaste desde la letra" />
          {vocabulary.length === 0 ? (
            <p className="text-[12.5px] leading-relaxed text-muted">
              Toca una palabra en la letra para consultarla; si te sirve, guardala y aparecera aqui.
            </p>
          ) : (
            <ul className="grid gap-2 md:grid-cols-2">
              {vocabulary.map((word) => (
                <li key={word.id} className="rounded-[9px] border border-line/70 px-3 py-2.5">
                  <div className="flex items-start justify-between gap-2">
                    <div className="min-w-0">
                      <p className="truncate text-[13px] font-medium text-ink">{word.word}</p>
                      <p className="truncate text-[11.5px] text-muted">{word.translation || word.meaning}</p>
                      {word.ipa ? <p className="mt-0.5 text-[11.5px] text-violet/85">{word.ipa}</p> : null}
                    </div>
                    <div className="flex shrink-0 items-center">
                      <IconButton label={`Escuchar ${word.word}`} size="sm" onClick={() => speak(word.word)}>
                        <Volume2 className="h-3.5 w-3.5" />
                      </IconButton>
                      <IconButton label={`Eliminar ${word.word}`} size="sm" onClick={() => void deleteWord(word.id)}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </IconButton>
                    </div>
                  </div>
                  {word.example ? (
                    <p className="mt-1.5 border-l border-line pl-2 text-[11px] italic leading-relaxed text-faint">
                      {word.example}
                    </p>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-[11px] text-faint">{label}</dt>
      <dd className="mt-0.5 text-[19px] font-medium text-ink tnum">{value}</dd>
    </div>
  );
}
