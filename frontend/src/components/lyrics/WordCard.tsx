import { Plus, Volume2 } from "lucide-react";
import { cn } from "../../lib/cn";
import type { Word } from "../../lib/types";
import { Button } from "../ui/Button";

interface WordCardProps {
  entry: Word | null;
  word: string;
  loading: boolean;
  error: string;
  alreadySaved: boolean;
  onSave: (entry: Word) => void;
  onClose: () => void;
  onSpeak: (text: string) => void;
}

/**
 * Dictionary card for a selected word. AI output is labelled as generated data,
 * never presented as a verified dictionary entry.
 */
export function WordCard({ entry, word, loading, error, alreadySaved, onSave, onClose, onSpeak }: WordCardProps) {
  return (
    <div
      className={cn(
        "mx-3 mb-3 rounded-[10px] border border-line bg-surface/95 p-3",
        "shadow-[0_14px_30px_rgba(0,0,0,0.45)]",
      )}
      role="region"
      aria-label={`Consulta de ${word}`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[14px] font-medium text-ink">{word}</p>
          {entry?.partOfSpeech ? (
            <p className="text-[11px] text-faint">{entry.partOfSpeech}</p>
          ) : null}
        </div>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => onSpeak(entry?.word || word)}
            aria-label={`Escuchar ${word}`}
            className="rounded p-1 text-faint transition-colors hover:bg-elevated hover:text-ink"
          >
            <Volume2 className="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            onClick={onClose}
            className="rounded px-1.5 py-0.5 text-[11px] text-faint transition-colors hover:text-ink"
          >
            Cerrar
          </button>
        </div>
      </div>

      {loading ? (
        <div className="mt-2 space-y-1.5" aria-busy="true">
          <div className="h-3 w-2/3 animate-pulse rounded bg-line" />
          <div className="h-3 w-1/2 animate-pulse rounded bg-line" />
          <div className="h-3 w-4/5 animate-pulse rounded bg-line" />
        </div>
      ) : error ? (
        <p className="mt-2 text-[12px] leading-relaxed text-danger">{error}</p>
      ) : entry ? (
        <div className="mt-2 space-y-1.5">
          {entry.ipa ? <p className="text-[12.5px] text-violet/90">{entry.ipa}</p> : null}
          <p className="text-[13px] text-ink">{entry.translation || entry.meaning}</p>
          {entry.meaning && entry.meaning !== entry.translation ? (
            <p className="text-[12px] leading-relaxed text-muted">{entry.meaning}</p>
          ) : null}
          {entry.example ? (
            <p className="border-l border-line pl-2 text-[11.5px] italic leading-relaxed text-faint">
              {entry.example}
            </p>
          ) : null}
          <div className="flex items-center justify-between pt-1">
            <span className="text-[10px] text-faint">Generado por IA · verifica antes de memorizar</span>
            <Button
              size="sm"
              variant={alreadySaved ? "subtle" : "primary"}
              disabled={alreadySaved}
              icon={<Plus className="h-3.5 w-3.5" />}
              onClick={() => onSave(entry)}
            >
              {alreadySaved ? "Guardada" : "Guardar"}
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
