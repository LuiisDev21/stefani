import type { Track } from "../lib/types";

/**
 * What the current source can actually do.
 *
 * The UI must never offer a control the provider cannot honour: a catalog-only
 * entry (YouTube Music metadata) has no audio element behind it, so seeking,
 * rate and seeking are reported as unavailable instead of silently failing.
 */
export interface PlaybackCapabilities {
  play: boolean;
  seek: boolean;
  rate: boolean;
  volume: boolean;
  /** True when seeking lands on an exact millisecond (local files, demo). */
  exactSeek: boolean;
}

export function capabilitiesFor(track: Track | null | undefined): PlaybackCapabilities {
  const playable = Boolean(track?.playable);
  return {
    play: playable,
    seek: playable,
    rate: playable,
    volume: true,
    exactSeek: playable,
  };
}
