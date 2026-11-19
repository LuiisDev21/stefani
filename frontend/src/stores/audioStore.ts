/**
 * The player's sound: the equaliser and the spatial engine, kept in one place
 * so the panel in Settings only describes intent and this store decides what
 * that means for the engine and for disk.
 *
 * Every change is applied to the audio graph immediately and written to disk
 * behind a debounce — dragging a slider must not queue forty writes.
 */

import { create } from "zustand";
import { audioEngine } from "../audio/engine";
import {
  AudioSettings,
  CUSTOM_PRESET,
  EQ_BANDS,
  PRESETS,
  SpatialMode,
  SpatialSettings,
  clampAudioSettings,
  defaultAudioSettings,
  presetMatchesBands,
} from "../audio/settings";
import { api } from "../lib/api";
import { ensureAudioElement } from "./playerStore";

const SAVE_DELAY = 400;

interface AudioState {
  settings: AudioSettings;
  loaded: boolean;
  /** True when this machine cannot run Web Audio at all. */
  unsupported: string | null;
  hrtf: boolean;

  load: () => Promise<void>;
  setEqEnabled: (enabled: boolean) => void;
  setPreset: (presetId: string) => void;
  setBand: (index: number, gain: number) => void;
  setPreamp: (preamp: number) => void;
  resetEq: () => void;

  setSpatialEnabled: (enabled: boolean) => void;
  setSpatialMode: (mode: SpatialMode) => void;
  setSpatial: <K extends keyof SpatialSettings>(key: K, value: SpatialSettings[K]) => void;
  resetSpatial: () => void;
}

let saveTimer: ReturnType<typeof setTimeout> | undefined;
let pending: AudioSettings | null = null;
let lastSaved = "";

function flush(): void {
  if (!pending) return;
  const payload = JSON.stringify(pending);
  pending = null;
  if (payload === lastSaved) return;
  lastSaved = payload;
  void api.saveAudioSettings(payload).catch(() => {
    // A failed write is retried by the next change; the sound never waits.
    lastSaved = "";
  });
}

function persist(settings: AudioSettings): void {
  pending = settings;
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    saveTimer = undefined;
    flush();
  }, SAVE_DELAY);
}

// Hiding the window is the last moment to write: a close right after a slider
// move must not cost the user that move.
if (typeof document !== "undefined") {
  const onHide = () => {
    if (document.hidden) {
      if (saveTimer) clearTimeout(saveTimer);
      saveTimer = undefined;
      flush();
    }
  };
  document.addEventListener("visibilitychange", onHide);
  window.addEventListener("pagehide", () => {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = undefined;
    flush();
  });
}

function apply(settings: AudioSettings): void {
  audioEngine.apply(settings);
}

export const useAudio = create<AudioState>((set, get) => {
  const commit = (next: AudioSettings) => {
    set({ settings: next });
    apply(next);
    persist(next);
  };

  return {
    settings: defaultAudioSettings(),
    loaded: false,
    unsupported: null,
    hrtf: false,

    load: async () => {
      // The element and the graph are built here, once, before the panel is
      // ever opened: reading the status earlier would always say "unsupported".
      const status = audioEngine.attach(ensureAudioElement());
      set({
        unsupported: status.supported ? null : status.reason ?? "sin Web Audio",
        hrtf: status.hrtf,
      });
      let settings = defaultAudioSettings();
      try {
        const stored = await api.audioSettings();
        if (stored?.json) settings = clampAudioSettings(JSON.parse(stored.json));
      } catch {
        // Corrupt or missing: defaults are the safe answer.
      }
      lastSaved = JSON.stringify(settings);
      set({ settings, loaded: true });
      apply(settings);
    },

    setEqEnabled: (enabled) => {
      const { eq } = get().settings;
      commit({ ...get().settings, eq: { ...eq, enabled } });
    },

    setPreset: (presetId) => {
      const preset = PRESETS.find((item) => item.id === presetId);
      const { eq } = get().settings;
      if (!preset) return;
      commit({
        ...get().settings,
        eq: { ...eq, enabled: true, preset: preset.id, bands: [...preset.bands], preamp: preset.preamp },
      });
    },

    setBand: (index, gain) => {
      if (index < 0 || index >= EQ_BANDS) return;
      const { eq } = get().settings;
      const bands = [...eq.bands];
      bands[index] = gain;
      const stillPreset = PRESETS.find((preset) => preset.id === eq.preset);
      const preset = stillPreset && presetMatchesBands(stillPreset, bands) ? stillPreset.id : CUSTOM_PRESET;
      commit({ ...get().settings, eq: { ...eq, bands, preset } });
    },

    setPreamp: (preamp) => {
      const { eq } = get().settings;
      commit({ ...get().settings, eq: { ...eq, preamp } });
    },

    resetEq: () => {
      const flat = PRESETS[0];
      commit({
        ...get().settings,
        eq: { enabled: true, preset: flat.id, bands: [...flat.bands], preamp: flat.preamp },
      });
    },

    setSpatialEnabled: (enabled) => {
      const { spatial } = get().settings;
      const mode = enabled && spatial.mode === "off" ? "orbit" : spatial.mode;
      commit({ ...get().settings, spatial: { ...spatial, enabled, mode } });
    },

    setSpatialMode: (mode) => {
      const { spatial } = get().settings;
      commit({
        ...get().settings,
        spatial: { ...spatial, mode, enabled: mode !== "off" },
      });
    },

    setSpatial: (key, value) => {
      const { spatial } = get().settings;
      commit({ ...get().settings, spatial: { ...spatial, [key]: value } });
    },

    resetSpatial: () => {
      const base = defaultAudioSettings().spatial;
      const { spatial } = get().settings;
      commit({ ...get().settings, spatial: { ...base, enabled: spatial.enabled, mode: spatial.mode } });
    },
  };
});
