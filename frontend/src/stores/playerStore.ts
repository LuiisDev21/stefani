import { create } from "zustand";
import { api } from "../lib/api";
import type { Track } from "../lib/types";
import { clock } from "../player/clock";

/**
 * Player store and engine.
 *
 * The HTML audio element lives at module scope: React never owns it, so remounts
 * and route changes cannot interrupt playback. Every side effect (loading,
 * playback state) is triggered from an action here, and the UI only
 * reads state.
 */

export type RepeatMode = "off" | "all" | "one";

interface PlayerState {
  queue: Track[];
  index: number;
  current: Track | null;
  playing: boolean;
  /** True while the current track is being resolved and buffered. */
  loading: boolean;
  volume: number;
  muted: boolean;
  rate: number;
  repeat: RepeatMode;
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
  setRepeat: (mode: RepeatMode) => void;
  clear: () => void;
}

const RATES = [0.75, 1, 1.25];

let audio: HTMLAudioElement | null = null;
let frame = 0;

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
    usePlayer.setState({ playing: false, loading: false, error: message });
    clock.set({ playing: false });
  });
  el.addEventListener("play", () => {
    usePlayer.setState({ playing: true, error: null });
    clock.set({ playing: true });
    startFrames();
  });
  // Upstream audio is resolved on demand, so the first seconds are silence:
  // the transport shows a spinner instead of looking dead.
  el.addEventListener("waiting", () => {
    if (!el.paused) usePlayer.setState({ loading: true });
  });
  el.addEventListener("canplay", () => {
    usePlayer.setState({ loading: false });
  });
  el.addEventListener("playing", () => {
    usePlayer.setState({ loading: false });
  });
  el.addEventListener("pause", () => {
    usePlayer.setState({ playing: false });
    clock.set({ playing: false });
  });
  return el;
}

function startFrames() {
  if (frame) return;
  const tick = () => {
    frame = requestAnimationFrame(tick);
    const el = audio;
    if (!el) return;
    const timeMs = Math.round(el.currentTime * 1000);
    let reported = timeMs;

    clock.set({ timeMs: reported });
  };
  frame = requestAnimationFrame(tick);
}

function stopFrames() {
  if (frame) {
    cancelAnimationFrame(frame);
    frame = 0;
  }
}


function handleEnded() {
  const state = usePlayer.getState();
  if (state.repeat === "one" && state.current) {
    ensureAudio().currentTime = 0;
    void ensureAudio().play();
    return;
  }
  const hasNext = playableIndex(state.queue, state.index + 1, 1) !== -1;
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
      loading: false,
      error: track.note || "Esta pista no se puede reproducir dentro de la app.",
    });
    clock.reset(track.durationMs || 0);
    return;
  }
  usePlayer.setState({ loading: true, error: null });
  el.src = track.audioUrl;
  el.playbackRate = usePlayer.getState().rate;
  el.volume = usePlayer.getState().muted ? 0 : usePlayer.getState().volume;
  clock.reset(track.durationMs || 0);
  if (autoplay) {
    void el.play().catch((error) => {
      usePlayer.setState({ loading: false, error: describePlayError(error) });
    });
  }
}

/** Skips rows YouTube will not serve: they show in the list but cannot play. */
function playableIndex(queue: Track[], from: number, direction: 1 | -1): number {
  for (let i = from; i >= 0 && i < queue.length; i += direction) {
    if (!queue[i]?.unavailable) return i;
  }
  return -1;
}

export const usePlayer = create<PlayerState>((set, get) => ({
  queue: [],
  index: -1,
  current: null,
  playing: false,
  loading: false,
  volume: 0.8,
  muted: false,
  rate: 1,
  repeat: "off",
  error: null,

  loadQueue: (tracks, startIndex = 0, autoplay = false) => {
    if (tracks.length === 0) return;
    const index = Math.min(Math.max(0, startIndex), tracks.length - 1);
    const track = tracks[index];
    set({ queue: tracks, index, current: track, error: null });
    loadAndPlay(track, autoplay);
  },

  playTrack: (track, queue) => {
    const existingIndex = queue?.findIndex((item) => item.id === track.id) ?? -1;
    if (queue && existingIndex >= 0) {
      set({ queue, index: existingIndex, current: track, error: null });
    } else {
      set({ queue: [track], index: 0, current: track, error: null });
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
    const nextIndex = playableIndex(queue, index + 1, 1);
    if (nextIndex === -1) {
      if (repeat === "all") {
        const first = playableIndex(queue, 0, 1);
        if (first === -1) return;
        set({ index: first, current: queue[first] });
        loadAndPlay(queue[first], true);
      }
      return;
    }
    set({ index: nextIndex, current: queue[nextIndex] });
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
    const prevIndex = Math.max(0, playableIndex(queue, index - 1, -1));
    set({ index: prevIndex, current: queue[prevIndex] });
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

  setRepeat: (mode) => {
    set({ repeat: mode });
  },

  clear: () => {
    ensureAudio().pause();
    ensureAudio().removeAttribute("src");
    stopFrames();
    clock.reset();
    set({ queue: [], index: -1, current: null, playing: false, loading: false, error: null });
  },
}));

/** Exported for tests: the rate ladder shown in the player. */
export const PLAYBACK_RATES = RATES;

/** The single audio element, created lazily. */
export function getAudioElement(): HTMLAudioElement | null {
  return audio;
}
