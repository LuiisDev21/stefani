import { Check, LoaderCircle } from "lucide-react";
import { useState } from "react";
import type { Playlist, Track } from "../../lib/types";
import { useApp } from "../../stores/appStore";
import { Button } from "../ui/Button";
import { Modal } from "../ui/Modal";

interface PlaylistPickerProps {
  track: Track | null;
  playlists: Playlist[];
  onClose: () => void;
}

/** Adds a track to an existing playlist or to a new one created on the spot. */
export function PlaylistPicker({ track, playlists, onClose }: PlaylistPickerProps) {
  const addToPlaylist = useApp((state) => state.addToPlaylist);
  const createPlaylist = useApp((state) => state.createPlaylist);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);
  const [addedTo, setAddedTo] = useState("");

  const handleAdd = async (playlist: Playlist) => {
    if (!track) return;
    setBusy(true);
    await addToPlaylist(playlist.id, track);
    setBusy(false);
    setAddedTo(playlist.id);
    onClose();
  };

  const handleCreate = async () => {
    if (!name.trim() || !track) return;
    setBusy(true);
    const playlist = await createPlaylist(name.trim());
    if (playlist) {
      await addToPlaylist(playlist.id, track);
      setName("");
      onClose();
    }
    setBusy(false);
  };

  return (
    <Modal
      open={Boolean(track)}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title="Agregar a una lista"
      description={track ? `"${track.title}" se agregara a la lista que elijas.` : undefined}
      footer={
        <>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Cerrar
          </Button>
          <Button variant="primary" size="sm" onClick={() => void handleCreate()} disabled={!name.trim() || busy}>
            Crear y agregar
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        {playlists.length === 0 ? (
          <p className="text-[12.5px] text-muted">Todavía no tienes listas. Crea la primera aquí abajo.</p>
        ) : (
          <ul className="max-h-[220px] space-y-1 overflow-y-auto pr-1">
            {playlists.map((playlist) => (
              <li key={playlist.id}>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => void handleAdd(playlist)}
                  className="flex w-full items-center justify-between gap-3 rounded-[7px] px-2.5 py-2 text-left transition-colors hover:bg-surface"
                >
                  <span className="min-w-0">
                    <span className="block truncate text-[13px] text-ink">{playlist.name}</span>
                    <span className="block text-[11px] text-faint tnum">{playlist.trackCount} canciones</span>
                  </span>
                  {addedTo === playlist.id ? (
                    <Check className="h-4 w-4 text-accent" />
                  ) : busy ? (
                    <LoaderCircle className="h-4 w-4 animate-spin text-faint" />
                  ) : null}
                </button>
              </li>
            ))}
          </ul>
        )}

        <div className="space-y-1.5 border-t border-line pt-3">
          <label htmlFor="new-playlist" className="block text-[11px] text-faint">
            Nueva lista
          </label>
          <input
            id="new-playlist"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Nombre de la lista"
            className="h-9 w-full rounded-[7px] border border-line bg-bg px-3 text-[13px] text-ink placeholder:text-faint focus:border-accent/60"
          />
        </div>
      </div>
    </Modal>
  );
}
