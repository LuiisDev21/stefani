/**
 * Typed facade over the generated Wails bindings.
 *
 * Everything the UI knows about the backend goes through this module: it is the
 * only place that imports the generated code, so a binding change breaks one
 * file instead of the whole app. Results are cast to the interfaces in
 * ./types.ts, which mirror internal/app/dto.go.
 */
import * as App from "../../bindings/github.com/LuiisDev21/stefani/internal/app/app.js";
import type {
  Account,
  AppUpdate,
  AccountPlaylist,
  AnnotateOutcome,
  Bootstrap,
  HistoryEntry,
  ImportResult,
  Info,
  Lyrics,
  OnlineStatus,
  Playlist,
  SearchResult,
  Settings,
  SettingsInput,
  TestOutcome,
  Track,
} from "./types";

const as = <T,>(value: unknown): Promise<T> => value as Promise<T>;

export const api = {
  info: () => as<Info>(App.Info()),
  updateYtDlp: () => as<OnlineStatus>(App.UpdateYtDlp()),
  checkAppUpdate: () => as<AppUpdate>(App.CheckAppUpdate()),
  installAppUpdate: () => as<AppUpdate>(App.InstallAppUpdate()),

  accountStatus: () => as<Account>(App.AccountStatus()),
  connectAccount: () => as<void>(App.ConnectAccount()),
  signOutAccount: () => as<void>(App.SignOutAccount()),
  finishAccount: () => as<Account>(App.FinishAccount()),
  refreshAccount: () => as<Account>(App.RefreshAccount()),
  accountPlaylists: () => as<AccountPlaylist[]>(App.AccountPlaylists()),
  importAccountPlaylist: (id: string, name: string) =>
    as<ImportResult>(App.ImportAccountPlaylist(id, name)),
  bootstrap: () => as<Bootstrap>(App.Bootstrap()),

  search: (query: string, limit = 25) => as<SearchResult>(App.Search(query, limit)),
  track: (id: string) => as<Track>(App.Track(id)),

  importLocalTracks: () => as<ImportResult>(App.ImportLocalTracks()),
  importDroppedFiles: (paths: string[]) => as<ImportResult>(App.ImportDroppedFiles(paths)),
  importPlaylist: (rawUrl: string) => as<ImportResult>(App.ImportPlaylist(rawUrl)),

  library: () => as<Track[]>(App.Library()),
  favorites: () => as<Track[]>(App.Favorites()),
  setInLibrary: (id: string, save: boolean) => as<Track>(App.SetInLibrary(id, save)),
  setFavorite: (id: string, favorite: boolean) => as<Track>(App.SetFavorite(id, favorite)),

  playlists: () => as<Playlist[]>(App.Playlists()),
  playlist: (id: string) => as<Playlist>(App.Playlist(id)),
  playlistTracks: (id: string) => as<Track[]>(App.PlaylistTracks(id)),
  createPlaylist: (name: string) => as<Playlist>(App.CreatePlaylist(name)),
  updatePlaylist: (id: string, name: string, description: string) =>
    as<void>(App.UpdatePlaylist(id, name, description)),
  deletePlaylist: (id: string) => as<void>(App.DeletePlaylist(id)),
  addToPlaylist: (playlistId: string, trackId: string) =>
    as<void>(App.AddToPlaylist(playlistId, trackId)),
  removeFromPlaylist: (playlistId: string, position: number) =>
    as<void>(App.RemoveFromPlaylist(playlistId, position)),
  movePlaylistItem: (playlistId: string, from: number, to: number) =>
    as<void>(App.MovePlaylistItem(playlistId, from, to)),

  lyrics: (trackId: string, refresh = false) => as<Lyrics>(App.Lyrics(trackId, refresh)),
  importLyricsFile: (trackId: string) => as<Lyrics>(App.ImportLyricsFile(trackId)),
  annotateTrack: (trackId: string) => as<AnnotateOutcome>(App.AnnotateTrack(trackId)),

  history: (limit = 30) => as<HistoryEntry[]>(App.History(limit)),

  settings: () => as<Settings>(App.Settings()),
  saveSettings: (input: SettingsInput) => as<Settings>(App.SaveSettings(input)),
  testLLM: () => as<TestOutcome>(App.TestLLM()),
  clearLLMKey: () => as<Settings>(App.ClearLLMKey()),

  setTrackDuration: (trackId: string, durationMs: number) =>
    as<void>(App.SetTrackDuration(trackId, durationMs)),
  reportPlayback: (trackId: string, listenedMs: number) =>
    as<void>(App.ReportPlayback(trackId, listenedMs)),
  openExternal: (url: string) => as<void>(App.OpenExternal(url)),
};

/** Extracts a readable message from a rejected binding call. */
export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  if (error && typeof error === "object" && "message" in error) {
    return String((error as { message: unknown }).message);
  }
  return "Ocurrio un error inesperado";
}
