/**
 * Types mirroring the JSON payloads produced by the Go service (internal/app/dto.go).
 * They are the contract with the backend: when a DTO changes, this file and the
 * Go test that asserts JSON keys must change with it.
 */

export type SyncLevel = "UNSYNCED" | "LINE" | "WORD" | "SYLLABLE";

export interface OnlineStatus {
  enabled: boolean;
  available: boolean;
  version: string;
  detail: string;
}

export interface AppUpdate {
  current: string;
  latest: string;
  available: boolean;
  ready: boolean;
  url: string;
  detail: string;
}

export interface Info {
  desktop: boolean;
  name: string;
  tagline: string;
  version: string;
  identifier: string;
  dbPath: string;
  schema: number;
  online: OnlineStatus;
}

export interface Track {
  id: string;
  provider: string;
  title: string;
  artist: string;
  album: string;
  durationMs: number;
  artworkUrl: string;
  externalUrl: string;
  audioUrl: string;
  playable: boolean;
  inLibrary: boolean;
  favorite: boolean;
  unavailable: boolean;
  note: string;
}

export interface ShelfItem {
  title: string;
  subtitle: string;
  cover: string;
  videoId: string;
  listId: string;
}

export interface Shelf {
  title: string;
  items: ShelfItem[];
}

export interface Playlist {
  id: string;
  name: string;
  description: string;
  artwork: string;
  origin: string;
  externalId: string;
  trackCount: number;
  updatedAt: number;
}

export interface WordTiming {
  text: string;
  startMs: number;
  endMs: number;
}

export interface Note {
  type: string;
  text: string;
}

export interface Annotation {
  lineId: string;
  accent: string;
  ipa: string;
  spanishApprox: string;
  translation: string;
  notes: Note[];
  model: string;
  /** "ai" for a generated annotation, "local" for the bundled dictionary. */
  source: "ai" | "local";
}

export interface Line {
  id: string;
  index: number;
  text: string;
  startMs: number;
  endMs: number;
  words: WordTiming[];
  annotation: Annotation | null;
}

export interface Lyrics {
  trackId: string;
  available: boolean;
  provider: string;
  sync: SyncLevel;
  source: string;
  license: string;
  hasWords: boolean;
  lines: Line[];
  notice: string;
}

export interface HistoryEntry {
  id: number;
  trackId: string;
  title: string;
  artist: string;
  playedAt: number;
  listenedMs: number;
}

export interface LLMSettings {
  provider: string;
  baseUrl: string;
  model: string;
  temperature: number;
  timeoutSeconds: number;
  maxRetries: number;
  jsonMode: boolean;
}

export interface Features {
  youtubeMusic: boolean;
}

export interface Settings {
  llm: LLMSettings;
  features: Features;
  accent: string;
  hasKey: boolean;
  keyStoreAvailable: boolean;
  configPath: string;
  keyStoreError: string;
}

export interface ProviderState {
  name: string;
  state: string;
  detail: string;
}

export interface StoredCounts {
  tracks: number;
  libraryTracks: number;
  favorites: number;
  playlists: number;
}

export interface Bootstrap {
  app: Info;
  recentlyPlayed: Track[];
  library: Track[];
  favorites: Track[];
  playlists: Playlist[];
  settings: Settings;
  providers: ProviderState[];
  counts: StoredCounts;
}

export interface SearchResult {
  query: string;
  items: Track[];
  providers: ProviderState[];
  notice: string;
}

export interface ImportResult {
  imported: Track[];
  skipped: string[];
  notice: string;
  playlist: Playlist | null;
}

export interface AnnotateOutcome {
  generated: number;
  cached: number;
  failed: string[];
  needsSetup: boolean;
  message: string;
}

export interface TestOutcome {
  ok: boolean;
  message: string;
  latencyMs: number;
  models?: string[];
  modelsNotice?: string;
}

export interface SettingsInput {
  llm: LLMSettings;
  features: Features;
  accent: string;
  apiKey: string;
}

export interface Account {
  connected: boolean;
  name: string;
  email: string;
  avatar: string;
  playlists: number;
  locked: boolean;
}

export interface AccountPlaylist {
  id: string;
  title: string;
  subtitle: string;
  cover: string;
}
