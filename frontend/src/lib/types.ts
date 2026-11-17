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

export interface Info {
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
  note: string;
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

export interface Word {
  id: string;
  word: string;
  lemma: string;
  partOfSpeech: string;
  meaning: string;
  ipa: string;
  translation: string;
  example: string;
  source: string;
  trackId: string;
  createdAt: number;
}

export interface Progress {
  trackId: string;
  title: string;
  artist: string;
  linesPracticed: number;
  studySeconds: number;
  lastStudiedAt: number;
  status: string;
  wordsSaved: number;
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
  vocabulary: number;
  annotated: number;
}

export interface Bootstrap {
  app: Info;
  demoTracks: Track[];
  recentlyPlayed: Track[];
  library: Track[];
  favorites: Track[];
  playlists: Playlist[];
  progress: Progress[];
  vocabulary: Word[];
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
}

export interface WordRequest {
  trackId: string;
  lineId: string;
  word: string;
  context: string;
}

export interface WordInput {
  word: string;
  lemma: string;
  partOfSpeech: string;
  meaning: string;
  ipa: string;
  translation: string;
  example: string;
  trackId: string;
}

export interface StudyInput {
  trackId: string;
  seconds: number;
  lines: number;
}

export interface SettingsInput {
  llm: LLMSettings;
  features: Features;
  accent: string;
  apiKey: string;
}
