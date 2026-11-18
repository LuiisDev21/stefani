import { useEffect } from "react";
import { useDuration, usePlaybackTime } from "../player/clock";
import { usePlayer } from "../stores/playerStore";
import { clearSession, mediaSessionAvailable, syncMetadata, syncPosition } from "../lib/mediaSession";

/**
 * Keeps the OS media session in sync with the player.
 *
 * It renders nothing: its only job is to publish the current track to Windows
 * (cover, title, artist, progress) and to route the hardware keys back into the
 * store. Metadata changes rarely, so it depends on the track identity; position
 * is updated on the clock tick.
 */
export function MediaSessionBridge() {
  const current = usePlayer((state) => state.current);
  const playing = usePlayer((state) => state.playing);
  const timeMs = usePlaybackTime();
  const durationMs = useDuration();
  const totalMs = durationMs || current?.durationMs || 0;

  useEffect(() => {
    if (!mediaSessionAvailable()) return;
    if (!current) {
      clearSession();
      document.title = "STEFANI";
      return;
    }
    syncMetadata(current, playing, {
      play: () => {
        if (!usePlayer.getState().playing) void usePlayer.getState().toggle();
      },
      pause: () => {
        if (usePlayer.getState().playing) void usePlayer.getState().toggle();
      },
      next: () => void usePlayer.getState().next(),
      previous: () => void usePlayer.getState().previous(),
      seek: (ms) => usePlayer.getState().seekMs(ms),
    });
    document.title = `${current.artist} — ${current.title}`;
  }, [current, playing]);

  useEffect(() => {
    if (!mediaSessionAvailable() || !current) return;
    syncPosition(timeMs, totalMs);
  }, [timeMs, totalMs, current]);

  return null;
}
