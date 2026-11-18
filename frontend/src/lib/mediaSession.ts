/**
 * Operating-system media integration.
 *
 * The Media Session API is what Windows reads to fill its media widget: cover,
 * title and artist in the notification panel, plus the hardware media keys.
 * Everything here is best-effort — a browser without the API must not break.
 */
import type { Track } from "./types";
import { hiResArtwork } from "./artwork";

export interface MediaSessionActions {
  play: () => void;
  pause: () => void;
  next: () => void;
  previous: () => void;
  seek: (ms: number) => void;
}

export function mediaSessionAvailable(): boolean {
  return typeof navigator !== "undefined" && "mediaSession" in navigator;
}

/** Artwork needs an absolute URL: relative ones are resolved against the app. */
function absolute(src: string): string {
  try {
    return new URL(src, window.location.href).href;
  } catch {
    return src;
  }
}

export function syncMetadata(track: Track, playing: boolean, actions: MediaSessionActions): void {
  if (!mediaSessionAvailable()) return;
  const session = navigator.mediaSession;

  // The widget renders the cover large, so ask for a 544px copy instead of the
  // 120px thumbnail the search row carries.
  const artwork = track.artworkUrl
    ? [{ src: absolute(hiResArtwork(track.artworkUrl)), sizes: "544x544", type: "image/jpeg" }]
    : [];
  session.metadata = new MediaMetadata({
    title: track.title,
    artist: track.artist,
    album: track.album ?? "",
    artwork,
  });
  session.playbackState = playing ? "playing" : "paused";

  setHandler(session, "play", () => actions.play());
  setHandler(session, "pause", () => actions.pause());
  setHandler(session, "nexttrack", () => actions.next());
  setHandler(session, "previoustrack", () => actions.previous());
  setHandler(session, "seekto", (details) => {
    if (typeof details.seekTime === "number") actions.seek(details.seekTime * 1000);
  });
}

export function syncPosition(positionMs: number, durationMs: number): void {
  if (!mediaSessionAvailable() || durationMs <= 0) return;
  try {
    navigator.mediaSession.setPositionState({
      duration: durationMs / 1000,
      position: Math.min(Math.max(positionMs, 0), durationMs) / 1000,
      playbackRate: 1,
    });
  } catch {
    // Some engines reject states while the element is loading; the next tick
    // sends a valid one.
  }
}

export function clearSession(): void {
  if (!mediaSessionAvailable()) return;
  navigator.mediaSession.metadata = null;
  navigator.mediaSession.playbackState = "none";
}

type Handler = Parameters<NonNullable<MediaSession["setActionHandler"]>>[1];

function setHandler(session: MediaSession, action: MediaSessionAction, handler: Handler): void {
  try {
    session.setActionHandler(action, handler);
  } catch {
    // Unsupported action in this engine: ignore it instead of failing playback.
  }
}
