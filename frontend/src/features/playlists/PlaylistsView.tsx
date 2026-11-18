import { ArrowDown, ArrowLeft, ArrowUp, ListMusic, Play, Plus, Trash2, UploadCloud } from "lucide-react";
import { useState } from "react";
import { api } from "../../lib/api";
import { formatRelativeDate, formatTime, plural } from "../../lib/format";
import type { Playlist, Track } from "../../lib/types";
import { useApp } from "../../stores/appStore";
import { usePlayer } from "../../stores/playerStore";
import { Button } from "../../components/ui/Button";
import { Cover } from "../../components/ui/Cover";
import { EmptyState } from "../../components/ui/EmptyState";
import { IconButton } from "../../components/ui/IconButton";
import { Modal } from "../../components/ui/Modal";
import { TopBar } from "../../components/layout/TopBar";

/**
 * Playlists: list and detail in one column, with an explicit back action.
 * Reordering uses buttons rather than drag-only, so it also works with a
 * keyboard and a screen reader.
 */
export function PlaylistsView() {
  const playlists = useApp((state) => state.playlists);
  const active = useApp((state) => state.activePlaylist);
  const tracks = useApp((state) => state.playlistTracks);
  const go = useApp((state) => state.go);
  const openPlaylist = useApp((state) => state.openPlaylist);
  const createPlaylist = useApp((state) => state.createPlaylist);
  const deletePlaylist = useApp((state) => state.deletePlaylist);
  const rename = useApp((state) => state.refreshPlaylists);
  const importPlaylist = useApp((state) => state.importPlaylist);
  const account = useApp((state) => state.account);
  const accountPlaylists = useApp((state) => state.accountPlaylists);
  const importAccountPlaylist = useApp((state) => state.importAccountPlaylist);
  const loadAccount = useApp((state) => state.loadAccount);
  const removeFromPlaylist = useApp((state) => state.removeFromPlaylist);
  const movePlaylistItem = useApp((state) => state.movePlaylistItem);
  const openTrack = useApp((state) => state.openTrack);
  const notify = useApp((state) => state.notify);

  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const toggle = usePlayer((state) => state.toggle);

  const [creating, setCreating] = useState(false);
  const [importing, setImporting] = useState(false);
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

  const openRename = (playlist: Playlist) => {
    setEditName(playlist.name);
    setEditDescription(playlist.description);
    setEditing(true);
  };

  const saveRename = async () => {
    if (!active) return;
    try {
      await api.updatePlaylist(active.id, editName, editDescription);
      await rename();
      await openPlaylist({ ...active, name: editName, description: editDescription });
      setEditing(false);
      notify("ok", "Lista actualizada");
    } catch (error) {
      notify("error", error instanceof Error ? error.message : String(error));
    }
  };

  if (active) {
    const totalMs = tracks.reduce((sum, track) => sum + track.durationMs, 0);
    return (
      <div className="min-h-0">
        <TopBar
          title={active.name}
          subtitle={`${plural(tracks.length, "canción", "canciones")} · ${formatTime(totalMs)} · ${active.origin === "ytmusic" ? "importada de YouTube Music" : "lista local"}`}
          actions={
            <>
              <Button size="sm" variant="ghost" icon={<ArrowLeft className="h-3.5 w-3.5" />} onClick={() => go("playlists")}>
                Listas
              </Button>
              <Button size="sm" variant="subtle" onClick={() => openRename(active)}>
                Editar
              </Button>
              <Button
                size="sm"
                variant="primary"
                icon={<Play className="h-3.5 w-3.5 fill-current" />}
                disabled={tracks.length === 0}
                onClick={() => tracks[0] && void openTrack(tracks[0], tracks)}
              >
                Reproducir
              </Button>
            </>
          }
        />

        <div className="px-7 pb-10">
          {active.description ? (
            <p className="mb-4 max-w-[70ch] text-[12.5px] leading-relaxed text-muted">{active.description}</p>
          ) : null}

          {tracks.length === 0 ? (
            <EmptyState
              icon={<ListMusic className="h-5 w-5" />}
              title="Esta lista esta vacia"
              description="Agrega canciones desde la biblioteca, la busqueda o la pantalla de inicio con el boton de lista."
            />
          ) : (
            <ol className="space-y-0.5">
              {tracks.map((track, index) => (
                <PlaylistRow
                  key={`${track.id}-${index}`}
                  track={track}
                  index={index}
                  total={tracks.length}
                  active={current?.id === track.id}
                  playing={playing}
                  onPlay={() => (current?.id === track.id ? toggle() : void openTrack(track, tracks))}
                  onUp={() => void movePlaylistItem(active.id, index, index - 1)}
                  onDown={() => void movePlaylistItem(active.id, index, index + 1)}
                  onRemove={() => void removeFromPlaylist(active.id, index)}
                />
              ))}
            </ol>
          )}
        </div>

        <Modal
          open={editing}
          onOpenChange={setEditing}
          title="Editar lista"
          footer={
            <>
              <Button variant="danger" size="sm" icon={<Trash2 className="h-3.5 w-3.5" />} onClick={() => setConfirmDelete(true)}>
                Eliminar
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setEditing(false)}>
                Cancelar
              </Button>
              <Button variant="primary" size="sm" onClick={() => void saveRename()} disabled={!editName.trim()}>
                Guardar
              </Button>
            </>
          }
        >
          <div className="space-y-3">
            <label className="block">
              <span className="mb-1 block text-[11px] text-faint">Nombre</span>
              <input
                value={editName}
                onChange={(event) => setEditName(event.target.value)}
                className="h-9 w-full rounded-[7px] border border-line bg-surface px-3 text-[13px] text-ink focus:border-line-strong"
              />
            </label>
            <label className="block">
              <span className="mb-1 block text-[11px] text-faint">Descripción</span>
              <textarea
                value={editDescription}
                onChange={(event) => setEditDescription(event.target.value)}
                rows={3}
                className="w-full rounded-[7px] border border-line bg-surface px-3 py-2 text-[13px] text-ink focus:border-line-strong"
              />
            </label>
          </div>
        </Modal>

        <Modal
          open={confirmDelete}
          onOpenChange={setConfirmDelete}
          title="Eliminar lista"
          description={`Se eliminara "${active.name}" con sus canciones guardadas en ella. Los archivos de audio no se tocan.`}
          footer={
            <>
              <Button variant="ghost" size="sm" onClick={() => setConfirmDelete(false)}>
                Cancelar
              </Button>
              <Button
                variant="danger"
                size="sm"
                onClick={() => {
                  setConfirmDelete(false);
                  setEditing(false);
                  void deletePlaylist(active.id);
                }}
              >
                Eliminar definitivamente
              </Button>
            </>
          }
        >
          <p className="text-[12.5px] text-muted">Esta accion no se puede deshacer.</p>
        </Modal>
      </div>
    );
  }

  return (
    <div className="min-h-0">
      <TopBar
        title="Listas"
        subtitle={playlists.length > 0 ? plural(playlists.length, "lista", "listas") : undefined}
        actions={
          <>
            <Button size="sm" variant="subtle" icon={<UploadCloud className="h-3.5 w-3.5" />} onClick={() => setImporting(true)}>
              Importar URL
            </Button>
            <Button size="sm" variant="primary" icon={<Plus className="h-3.5 w-3.5" />} onClick={() => setCreating(true)}>
              Nueva lista
            </Button>
          </>
        }
      />

      <div className="px-7 pb-10">
        {account?.connected && accountPlaylists.length > 0 ? (
          <section className="mb-8">
            <div className="flex items-center justify-between">
              <h2 className="px-2 text-[12px] font-semibold text-ash">De tu cuenta</h2>
              <button
                type="button"
                onClick={() => void loadAccount()}
                className="px-2 text-[11.5px] text-muted transition-colors hover:text-ink"
              >
                Actualizar
              </button>
            </div>
            <ul className="mt-2 space-y-1">
              {accountPlaylists.map((item) => (
                <li key={item.id}>
                  <div className="flex items-center gap-3 rounded-[9px] px-2 py-2">
                    <Cover title={item.title} src={item.cover} size={38} />
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-[13.5px] text-ink">{item.title}</span>
                      <span className="block truncate text-[11.5px] text-muted">{item.subtitle}</span>
                    </span>
                    <Button
                      size="sm"
                      variant="subtle"
                      onClick={() => void importAccountPlaylist(item.id, item.title)}
                    >
                      Importar
                    </Button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        ) : null}

        {playlists.length === 0 ? (
          <EmptyState
            icon={<ListMusic className="h-5 w-5" />}
            title="Sin listas todavía"
            description="Crea una lista propia o importa una playlist pública de YouTube Music. El audio se reproduce dentro de la app cuando el reproductor en linea esta disponible."
            actions={
              <>
                <Button size="sm" variant="primary" onClick={() => setCreating(true)}>
                  Crear lista
                </Button>
                <Button size="sm" variant="subtle" onClick={() => setImporting(true)}>
                  Importar de YouTube Music
                </Button>
              </>
            }
          />
        ) : (
          <ul className="space-y-1">
            {playlists.map((playlist) => (
              <li key={playlist.id}>
                <button
                  type="button"
                  onClick={() => void openPlaylist(playlist)}
                  className="flex w-full items-center gap-3 rounded-[9px] px-2 py-2 text-left transition-colors hover:bg-surface/50"
                >
                  <Cover title={playlist.name} size={38} />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-[13px] text-ink">{playlist.name}</span>
                    <span className="block truncate text-[11.5px] text-faint tnum">
                      {playlist.trackCount} canciones · {formatRelativeDate(playlist.updatedAt)}
                    </span>
                  </span>
                  <span className="text-[11px] text-faint">
                    {playlist.origin === "ytmusic" ? "YouTube" : "local"}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      <Modal
        open={creating}
        onOpenChange={setCreating}
        title="Nueva lista"
        description="Ponle un nombre; después agrega canciones desde cualquier lista de pistas."
        footer={
          <>
            <Button variant="ghost" size="sm" onClick={() => setCreating(false)}>
              Cancelar
            </Button>
            <Button
              variant="primary"
              size="sm"
              disabled={!name.trim()}
              onClick={() => {
                void createPlaylist(name).then((playlist) => {
                  setName("");
                  setCreating(false);
                  if (playlist) void openPlaylist(playlist);
                });
              }}
            >
              Crear
            </Button>
          </>
        }
      >
        <input
          autoFocus
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Nombre de la lista"
          aria-label="Nombre de la lista"
          className="h-9 w-full rounded-[7px] border border-line bg-surface px-3 text-[13px] text-ink placeholder:text-faint focus:border-line-strong"
        />
      </Modal>

      <Modal
        open={importing}
        onOpenChange={setImporting}
        title="Importar playlist"
        description="Pega la URL de una playlist pública de YouTube Music. Se importan los metadatos; la reproducción se hace en la pagina oficial."
        footer={
          <>
            <Button variant="ghost" size="sm" onClick={() => setImporting(false)}>
              Cancelar
            </Button>
            <Button
              variant="primary"
              size="sm"
              disabled={!url.trim()}
              onClick={() => {
                void importPlaylist(url).then(() => {
                  setUrl("");
                  setImporting(false);
                });
              }}
            >
              Importar
            </Button>
          </>
        }
      >
        <input
          autoFocus
          value={url}
          onChange={(event) => setUrl(event.target.value)}
          placeholder="https://music.youtube.com/playlist?list=..."
          aria-label="URL de la playlist"
          className="h-9 w-full rounded-[7px] border border-line bg-surface px-3 text-[12.5px] text-ink placeholder:text-faint focus:border-line-strong"
        />
      </Modal>
    </div>
  );
}

interface PlaylistRowProps {
  track: Track;
  index: number;
  total: number;
  active: boolean;
  playing: boolean;
  onPlay: () => void;
  onUp: () => void;
  onDown: () => void;
  onRemove: () => void;
}

function PlaylistRow({ track, index, total, active, playing, onPlay, onUp, onDown, onRemove }: PlaylistRowProps) {
  return (
    <li className="group flex items-center gap-3 rounded-[7px] px-2 py-1.5 transition-colors hover:bg-surface">
      <button
        type="button"
        onClick={onPlay}
        aria-label={`Reproducir ${track.title}`}
        className="flex h-7 w-7 items-center justify-center rounded text-[11px] text-faint tnum hover:text-ink"
      >
        {active && playing ? <span className="text-accent">II</span> : index + 1}
      </button>
      <button type="button" onClick={onPlay} className="flex min-w-0 flex-1 items-center gap-3 text-left">
        <Cover title={track.title} src={track.artworkUrl} size={32} />
        <span className="min-w-0">
          <span className={`block truncate text-[12.5px] ${active ? "text-accent" : "text-ink"}`}>{track.title}</span>
          <span className="block truncate text-[11px] text-faint">{track.artist || "Artista desconocido"}</span>
        </span>
      </button>
      <span className="w-9 text-right text-[10.5px] text-faint tnum">
        {track.durationMs > 0 ? formatTime(track.durationMs) : "—"}
      </span>
      <div className="flex items-center opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
        <IconButton label="Subir" size="sm" onClick={onUp} disabled={index === 0}>
          <ArrowUp className="h-3.5 w-3.5" />
        </IconButton>
        <IconButton label="Bajar" size="sm" onClick={onDown} disabled={index === total - 1}>
          <ArrowDown className="h-3.5 w-3.5" />
        </IconButton>
        <IconButton label="Quitar de la lista" size="sm" onClick={onRemove}>
          <Trash2 className="h-3.5 w-3.5" />
        </IconButton>
      </div>
    </li>
  );
}
