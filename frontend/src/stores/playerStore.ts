import { create } from "zustand";
import { api } from "../lib/api";
import type { LoopBounds } from "../lib/timeline";
import type { Track } from "../lib/types";
import { clock } from "../player/clock";

/**
 * Player store and engine.
 *
 * The HTML audio element lives at module scope: React never owns it, so remounts
 * and route changes cannot interrupt playback. Every side effect (loading,
 * looping, study accounting) is triggered from an action here, and the UI only
 * reads state.
 */

export type RepeatMode = "off" | "all" | "one";

interface PlayerState {
  queue: Track[];
  index: number;
  current: Track | null;
  playing: boolean;
  volume: number;
  muted: boolean;
  rate: number;
  repeat: RepeatMode;
  loop: LoopBounds | null;
  error: string | null;

  loadQueue: (tracks: Track[], startIndex?: number, autoplay?: boolean) => void;
  playTrack: (track: Track, queue?: Track[]) => void;
  toggle: () => void;
  pause: () => void;
  next: () => void;
  previous: () => void;
  seekMs: (ms: number) => void;
  setVolume: (value: number) => void;
  toggleMute: () => void;
  setRate: (value: number) => void;
  setLoop: (bounds: LoopBounds | null) => void;
  setRepeat: (mode: RepeatMode) => void;
  clear: () => void;
}

const RATES = [0.75, 1, 1.25];

let audio: HTMLAudioElement | null = null;
let frame = 0;
let studySeconds = 0;
let studyTrackId = "";

function ensureAudio(): HTMLAudioElement {
  if (audio) return audio;
  const el = new Audio();
  el.preload = "metadata";
  el.setAttribute("data-stefani-player", "true");
  el.style.display = "none";
  // Attached to the document so the element is inspectable and so media
  // sessions behave like any other page audio.
  document.body.appendChild(el);
  audio = el;

  el.addEventListener("loadedmetadata", () => {
    const durationMs = Number.isFinite(el.duration) ? Math.round(el.duration * 1000) : 0;
    clock.set({ durationMs });
    const track = usePlayer.getState().current;
    if (track && track.durationMs === 0 && durationMs > 0) {
      void api.setTrackDuration(track.id, durationMs).catch(() => undefined);
    }
  });
  el.addEventListener("ended", () => handleEnded());
  el.addEventListener("error", () => {
    const track = usePlayer.getState().current;
    const message = track
      ? `No pude reproducir "${track.title}". Revisa que el archivo siga disponible.`
      : "No pude reproducir el audio.";
    usePlayer.setState({ playing: false, error: message });
    clock.set({ playing: false });
  });
  el.addEventListener("play", () => {
    usePlayer.setState({ playing: true, error: null });
    clock.set({ playing: true });
    startFrames();
  });
  el.addEventListener("pause", () => {
    usePlayer.setState({ playing: false });
    clock.set({ playing: false });
    flushStudy();
  });
  return el;
}

function startFrames() {
  if (frame) return;
  const tick = () => {
    frame = requestAnimationFrame(tick);
    const el = audio;
    if (!el) return;
    const state = usePlayer.getState();
    const timeMs = Math.round(el.currentTime * 1000);
    let reported = timeMs;

    if (state.loop && !el.paused) {
      if (timeMs >= state.loop.endMs || timeMs < state.loop.startMs - 400) {
        el.currentTime = state.loop.startMs / 1000;
        reported = state.loop.startMs;
      }
    }
    clock.set({ timeMs: reported });

    // Study accounting: count only real listening time on the current track.
    if (!el.paused && state.current) {
      studySeconds += 1 / 60;
      if (state.current.id !== studyTrackId) {
        studyTrackId = state.current.id;
        studySeconds = 0;
      }
      if (studySeconds >= 30) {
        const seconds = Math.round(studySeconds);
        studySeconds = 0;
        void api.recordStudy({ trackId: state.current.id, seconds, lines: 0 }).catch(() => undefined);
      }
    }
  };
  frame = requestAnimationFrame(tick);
}

function stopFrames() {
  if (frame) {
    cancelAnimationFrame(frame);
    frame = 0;
  }
}

function flushStudy() {
  const state = usePlayer.getState();
  if (state.current && studySeconds >= 5) {
    const seconds = Math.round(studySeconds);
    studySeconds = 0;
    void api.recordStudy({ trackId: state.current.id, seconds, lines: 0 }).catch(() => undefined);
  }
}

function handleEnded() {
  const state = usePlayer.getState();
  if (state.repeat === "one" && state.current) {
    ensureAudio().currentTime = 0;
    void ensureAudio().play();
    return;
  }
  const hasNext = state.index + 1 < state.queue.length;
  if (hasNext || state.repeat === "all") {
    state.next();
    return;
  }
  usePlayer.setState({ playing: false });
  clock.set({ playing: false });
  stopFrames();
}

function describePlayError(error: unknown): string {
  const name = (error as { name?: string } | null)?.name;
  if (name === "NotAllowedError") {
    return "Pulsa reproducir para iniciar el audio.";
  }
  if (error instanceof Error) return error.message;
  return "No pude reproducir el audio.";
}

function loadAndPlay(track: Track, autoplay: boolean) {
  const el = ensureAudio();
  if (!track.playable || !track.audioUrl) {
    usePlayer.setState({
      current: track,
      playing: false,
      error: track.note || "Esta pista no se puede reproducir dentro de la app.",
    });
    clock.reset(track.durationMs || 0);
    return;
  }
  el.src = track.audioUrl;
  el.playbackRate = usePlayer.getState().rate;
  el.volume = usePlayer.getState().muted ? 0 : usePlayer.getState().volume;
  clock.reset(track.durationMs || 0);
  studyTrackId = track.id;
  studySeconds = 0;
  if (autoplay) {
    void el.play().catch((error) => {
      usePlayer.setState({ error: describePlayError(error) });
    });
  }
}

export const usePlayer = create<PlayerState>((set, get) => ({
  queue: [],
  index: -1,
  current: null,
  playing: false,
  volume: 0.8,
  muted: false,
  rate: 1,
  repeat: "off",
  loop: null,
  error: null,

  loadQueue: (tracks, startIndex = 0, autoplay = false) => {
    if (tracks.length === 0) return;
    const index = Math.min(Math.max(0, startIndex), tracks.length - 1);
    const track = tracks[index];
    set({ queue: tracks, index, current: track, error: null, loop: null });
    loadAndPlay(track, autoplay);
  },

  playTrack: (track, queue) => {
    const existingIndex = queue?.findIndex((item) => item.id === track.id) ?? -1;
    if (queue && existingIndex >= 0) {
      set({ queue, index: existingIndex, current: track, error: null, loop: null });
    } else {
      set({ queue: [track], index: 0, current: track, error: null, loop: null });
    }
    loadAndPlay(track, true);
  },

  toggle: () => {
    const el = ensureAudio();
    const state = get();
    if (!state.current) return;
    if (el.paused) {
      if (!el.src) {
        loadAndPlay(state.current, true);
        return;
      }
      void el.play().catch((error) => set({ error: describePlayError(error) }));
    } else {
      el.pause();
    }
  },

  pause: () => {
    ensureAudio().pause();
  },

  next: () => {
    const { queue, index, repeat } = get();
    if (queue.length === 0) return;
    const nextIndex = index + 1;
    if (nextIndex >= queue.length) {
      if (repeat === "all") {
        set({ index: 0, current: queue[0], loop: null });
        loadAndPlay(queue[0], true);
      }
      return;
    }
    set({ index: nextIndex, current: queue[nextIndex], loop: null });
    loadAndPlay(queue[nextIndex], true);
  },

  previous: () => {
    const { queue, index } = get();
    if (queue.length === 0) return;
    // Restart the current track first: that is what listeners expect.
    if (ensureAudio().currentTime > 3) {
      ensureAudio().currentTime = 0;
      clock.set({ timeMs: 0 });
      return;
    }
    const prevIndex = Math.max(0, index - 1);
    set({ index: prevIndex, current: queue[prevIndex], loop: null });
    loadAndPlay(queue[prevIndex], true);
  },

  seekMs: (ms) => {
    const el = ensureAudio();
    const state = get();
    const clamped = Math.max(0, Math.min(ms, Number.isFinite(el.duration) ? el.duration * 1000 : ms));
    if (Number.isFinite(el.duration)) {
      el.currentTime = clamped / 1000;
    }
    clock.set({ timeMs: clamped, durationMs: state.current?.durationMs || clock.getSnapshot().durationMs });
  },

  setVolume: (value) => {
    const volume = Math.min(1, Math.max(0, value));
    ensureAudio().volume = volume;
    set({ volume, muted: false });
  },

  toggleMute: () => {
    const muted = !get().muted;
    ensureAudio().volume = muted ? 0 : get().volume;
    set({ muted });
  },

  setRate: (value) => {
    const rate = RATES.includes(value) ? value : 1;
    ensureAudio().playbackRate = rate;
    set({ rate });
  },

  setLoop: (bounds) => {
    set({ loop: bounds });
  },

  setRepeat: (mode) => {
    set({ repeat: mode });
  },

  clear: () => {
    ensureAudio().pause();
    ensureAudio().removeAttribute("src");
    stopFrames();
    clock.reset();
    set({ queue: [], index: -1, current: null, playing: false, loop: null, error: null });
  },
}));

/** Exported for tests: the rate ladder shown in the player. */
export const PLAYBACK_RATES = RATES;

/** The single audio element, created lazily. */
export function getAudioElement(): HTMLAudioElement | null {
  return audio;
}
