import { Events } from "@wailsio/runtime";
import { create } from "zustand";
import { api, errorMessage } from "../lib/api";
import type {
  Bootstrap,
  Info,
  Lyrics,
  Playlist,
  Progress,
  ProviderState,
  SearchResult,
  Settings,
  StoredCounts,
  Track,
  Word,
} from "../lib/types";
import { usePlayer } from "./playerStore";

export type View =
  | "home"
  | "search"
  | "library"
  | "playlists"
  | "favorites"
  | "learning"
  | "settings";

export type ToastKind = "ok" | "error" | "info";

export interface Toast {
  id: number;
  kind: ToastKind;
  message: string;
}

export interface WordPanel {
  word: string;
  lineId: string;
  context: string;
  loading: boolean;
  entry: Word | null;
  error: string;
}

interface AppState {
  ready: boolean;
  fatalError: string;

  info: Info | null;
  demo: Track[];
  recently: Track[];
  library: Track[];
  favorites: Track[];
  playlists: Playlist[];
  progress: Progress[];
  vocabulary: Word[];
  providers: ProviderState[];
  settings: Settings | null;
  counts: StoredCounts | null;

  view: View;
  activePlaylist: Playlist | null;
  playlistTracks: Track[];

  lyrics: Lyrics | null;
  lyricsLoading: boolean;

  searchQuery: string;
  searching: boolean;
  searchResult: SearchResult | null;

  annotating: boolean;
  annotateProgress: { done: number; total: number } | null;

  wordPanel: WordPanel | null;
  toast: Toast | null;
  immersive: boolean;
  pickerTrack: Track | null;

  init: () => Promise<void>;
  go: (view: View, playlist?: Playlist) => void;
  openTrack: (track: Track, queue?: Track[]) => Promise<void>;
  loadLyrics: (trackId: string, refresh?: boolean) => Promise<void>;
  annotate: () => Promise<void>;
  toggleLibrary: (track: Track) => Promise<void>;
  toggleFavorite: (track: Track) => Promise<void>;

  search: (query: string) => Promise<void>;
  clearSearch: () => void;

  importAudio: () => Promise<void>;
  importLyrics: () => Promise<void>;
  importPlaylist: (url: string) => Promise<void>;
  openPlaylist: (playlist: Playlist) => Promise<void>;
  createPlaylist: (name: string) => Promise<Playlist | null>;
  deletePlaylist: (id: string) => Promise<void>;
  addToPlaylist: (playlistId: string, track: Track) => Promise<void>;
  removeFromPlaylist: (playlistId: string, position: number) => Promise<void>;
  movePlaylistItem: (playlistId: string, from: number, to: number) => Promise<void>;

  lookupWord: (word: string, lineId: string, context: string) => Promise<void>;
  closeWordPanel: () => void;
  saveWord: (entry: Word) => Promise<void>;
  deleteWord: (id: string) => Promise<void>;

  saveSettings: (input: Parameters<typeof api.saveSettings>[0]) => Promise<boolean>;
  testLLM: () => Promise<void>;
  clearKey: () => Promise<void>;

  refreshLibrary: () => Promise<void>;
  refreshPlaylists: () => Promise<void>;
  refreshProgress: () => Promise<void>;
  refreshVocabulary: () => Promise<void>;

  notify: (kind: ToastKind, message: string) => void;
  dismissToast: () => void;
  setImmersive: (value: boolean) => void;
  openPicker: (track: Track) => void;
  closePicker: () => void;
}

let toastId = 0;

function applyBootstrap(data: Bootstrap) {
  return {
    ready: true,
    info: data.app,
    demo: data.demoTracks ?? [],
    recently: data.recentlyPlayed ?? [],
    library: data.library ?? [],
    favorites: data.favorites ?? [],
    playlists: data.playlists ?? [],
    progress: data.progress ?? [],
    vocabulary: data.vocabulary ?? [],
    providers: data.providers ?? [],
    settings: data.settings,
    counts: data.counts,
  };
}

export const useApp = create<AppState>((set, get) => ({
  ready: false,
  fatalError: "",

  info: null,
  demo: [],
  recently: [],
  library: [],
  favorites: [],
  playlists: [],
  progress: [],
  vocabulary: [],
  providers: [],
  settings: null,
  counts: null,

  view: "home",
  activePlaylist: null,
  playlistTracks: [],

  lyrics: null,
  lyricsLoading: false,

  searchQuery: "",
  searching: false,
  searchResult: null,

  annotating: false,
  annotateProgress: null,

  wordPanel: null,
  toast: null,
  immersive: false,
  pickerTrack: null,

  init: async () => {
    try {
      const data = await api.bootstrap();
      set(applyBootstrap(data));
      listenForProgress();
    } catch (error) {
      set({ ready: true, fatalError: errorMessage(error) });
    }
  },

  go: (view, playlist) => {
    set({ view, activePlaylist: playlist ?? null, immersive: false });
    if (view === "playlists" && playlist) {
      void get().openPlaylist(playlist);
    }
    if (view === "learning") void get().refreshProgress();
    if (view === "favorites") void get().refreshLibrary();
  },

  openTrack: async (track, queue) => {
    usePlayer.getState().playTrack(track, queue);
    set({ lyrics: null, lyricsLoading: true, wordPanel: null });
    await get().loadLyrics(track.id);
  },

  loadLyrics: async (trackId, refresh = false) => {
    set({ lyricsLoading: true });
    try {
      const lyrics = await api.lyrics(trackId, refresh);
      // Ignore a response that arrived after the user changed track.
      if (usePlayer.getState().current?.id !== trackId) return;
      set({ lyrics, lyricsLoading: false });
    } catch (error) {
      set({ lyricsLoading: false });
      get().notify("error", errorMessage(error));
    }
  },

  annotate: async () => {
    const track = usePlayer.getState().current;
    if (!track) return;
    set({ annotating: true, annotateProgress: { done: 0, total: 0 } });
    try {
      const outcome = await api.annotateTrack(track.id);
      set({ annotating: false, annotateProgress: null });
      if (outcome.needsSetup) {
        get().notify("info", outcome.message || "Configura la IA en Ajustes");
        set({ view: "settings" });
        return;
      }
      if (outcome.message) {
        get().notify(outcome.failed.length > 0 ? "error" : "info", outcome.message);
      } else if (outcome.generated > 0) {
        get().notify("ok", `${outcome.generated} verso(s) anotados`);
      } else {
        get().notify("info", "La letra ya tenia todas las anotaciones");
      }
      await get().loadLyrics(track.id);
    } catch (error) {
      set({ annotating: false, annotateProgress: null });
      get().notify("error", errorMessage(error));
    }
  },

  toggleLibrary: async (track) => {
    try {
      const updated = await api.setInLibrary(track.id, !track.inLibrary);
      get().notify("ok", updated.inLibrary ? "Guardada en tu biblioteca" : "Quitada de la biblioteca");
      await get().refreshLibrary();
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  toggleFavorite: async (track) => {
    try {
      await api.setFavorite(track.id, !track.favorite);
      await get().refreshLibrary();
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  search: async (query) => {
    const trimmed = query.trim();
    if (!trimmed) {
      set({ searchQuery: "", searchResult: null, searching: false });
      return;
    }
    set({ searchQuery: query, searching: true });
    try {
      const result = await api.search(trimmed, 30);
      set({ searchResult: result, searching: false });
    } catch (error) {
      set({ searching: false });
      get().notify("error", errorMessage(error));
    }
  },

  clearSearch: () => set({ searchQuery: "", searchResult: null }),

  importAudio: async () => {
    try {
      const result = await api.importLocalTracks();
      if (result.notice) get().notify(result.imported.length > 0 ? "ok" : "info", result.notice);
      await get().refreshLibrary();
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  importLyrics: async () => {
    const track = usePlayer.getState().current;
    if (!track) {
      get().notify("info", "Elige una cancion primero");
      return;
    }
    try {
      const lyrics = await api.importLyricsFile(track.id);
      set({ lyrics });
      get().notify(lyrics.available ? "ok" : "info", lyrics.available ? "Letra importada" : lyrics.notice);
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  importPlaylist: async (url) => {
    try {
      const result = await api.importPlaylist(url);
      if (result.notice) get().notify(result.imported.length > 0 ? "ok" : "info", result.notice);
      await get().refreshPlaylists();
      if (result.playlist) {
        await get().openPlaylist(result.playlist);
        set({ view: "playlists" });
      }
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  openPlaylist: async (playlist) => {
    set({ activePlaylist: playlist, playlistTracks: [] });
    try {
      const tracks = await api.playlistTracks(playlist.id);
      set({ playlistTracks: tracks });
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  createPlaylist: async (name) => {
    try {
      const playlist = await api.createPlaylist(name);
      await get().refreshPlaylists();
      get().notify("ok", `Lista "${playlist.name}" creada`);
      return playlist;
    } catch (error) {
      get().notify("error", errorMessage(error));
      return null;
    }
  },

  deletePlaylist: async (id) => {
    try {
      await api.deletePlaylist(id);
      set({ activePlaylist: null, playlistTracks: [] });
      await get().refreshPlaylists();
      get().notify("ok", "Lista eliminada");
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  addToPlaylist: async (playlistId, track) => {
    try {
      await api.addToPlaylist(playlistId, track.id);
      if (get().activePlaylist?.id === playlistId) {
        await get().openPlaylist(get().activePlaylist as Playlist);
      }
      await get().refreshPlaylists();
      get().notify("ok", `"${track.title}" agregada a la lista`);
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  removeFromPlaylist: async (playlistId, position) => {
    try {
      await api.removeFromPlaylist(playlistId, position);
      const playlist = get().activePlaylist;
      if (playlist) await get().openPlaylist(playlist);
      await get().refreshPlaylists();
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  movePlaylistItem: async (playlistId, from, to) => {
    try {
      await api.movePlaylistItem(playlistId, from, to);
      const playlist = get().activePlaylist;
      if (playlist) await get().openPlaylist(playlist);
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  lookupWord: async (word, lineId, context) => {
    const track = usePlayer.getState().current;
    if (!track) return;
    set({
      wordPanel: { word, lineId, context, loading: true, entry: null, error: "" },
    });
    try {
      const entry = await api.lookupWord({ trackId: track.id, lineId, word, context });
      set({ wordPanel: { word, lineId, context, loading: false, entry, error: "" } });
    } catch (error) {
      set({ wordPanel: { word, lineId, context, loading: false, entry: null, error: errorMessage(error) } });
    }
  },

  closeWordPanel: () => set({ wordPanel: null }),

  saveWord: async (entry) => {
    try {
      await api.saveWord({
        word: entry.word,
        lemma: entry.lemma,
        partOfSpeech: entry.partOfSpeech,
        meaning: entry.meaning,
        ipa: entry.ipa,
        translation: entry.translation,
        example: entry.example,
        trackId: entry.trackId,
      });
      await get().refreshVocabulary();
      get().notify("ok", `"${entry.word}" guardada en vocabulario`);
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  deleteWord: async (id) => {
    try {
      await api.deleteWord(id);
      await get().refreshVocabulary();
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  saveSettings: async (input) => {
    try {
      const settings = await api.saveSettings(input);
      set({ settings });
      if (!settings.keyStoreAvailable || settings.keyStoreError) {
        get().notify("error", settings.keyStoreError || "El almacen de credenciales no esta disponible");
        return false;
      }
      get().notify("ok", "Ajustes guardados");
      return true;
    } catch (error) {
      get().notify("error", errorMessage(error));
      return false;
    }
  },

  testLLM: async () => {
    try {
      const outcome = await api.testLLM();
      get().notify(outcome.ok ? "ok" : "error", `${outcome.message} (${outcome.latencyMs} ms)`);
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  clearKey: async () => {
    try {
      const settings = await api.clearLLMKey();
      set({ settings });
      get().notify("ok", "Clave eliminada del almacen del sistema");
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  refreshLibrary: async () => {
    try {
      const [library, favorites] = await Promise.all([api.library(), api.favorites()]);
      set({ library, favorites });
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  refreshPlaylists: async () => {
    try {
      const playlists = await api.playlists();
      set({ playlists });
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  refreshProgress: async () => {
    try {
      const progress = await api.progress();
      set({ progress });
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  refreshVocabulary: async () => {
    try {
      const vocabulary = await api.vocabulary();
      set({ vocabulary });
    } catch (error) {
      get().notify("error", errorMessage(error));
    }
  },

  notify: (kind, message) => {
    toastId += 1;
    set({ toast: { id: toastId, kind, message } });
    window.setTimeout(() => {
      if (get().toast?.id === toastId) set({ toast: null });
    }, 5200);
  },

  dismissToast: () => set({ toast: null }),

  setImmersive: (value) => set({ immersive: value }),

  openPicker: (track) => set({ pickerTrack: track }),
  closePicker: () => set({ pickerTrack: null }),
}));

/** Subscribes to backend progress events. Safe outside Wails: errors are ignored. */
function listenForProgress() {
  try {
    Events.On("annotate:progress", (event: unknown) => {
      const payload = (event as { data?: { done?: number; total?: number } })?.data ?? {};
      const done = Number(payload.done ?? 0);
      const total = Number(payload.total ?? 0);
      if (!Number.isFinite(done) || !Number.isFinite(total)) return;
      useApp.setState({ annotateProgress: { done, total } });
    });
  } catch {
    // The runtime is unavailable in a plain browser; progress simply stays idle.
  }
}
