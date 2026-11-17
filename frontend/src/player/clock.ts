import { useSyncExternalStore } from "react";

/**
 * The playback clock.
 *
 * The audio element is the only authority on time. This store publishes its
 * position at animation-frame rate so lyric highlighting stays smooth, while
 * keeping the update cost low: the snapshot object is replaced only when a value
 * actually changes, and consumers subscribe to the fields they need.
 */
export interface ClockState {
  timeMs: number;
  durationMs: number;
  playing: boolean;
}

class ClockStore {
  private state: ClockState = { timeMs: 0, durationMs: 0, playing: false };
  private listeners = new Set<() => void>();

  subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };

  getSnapshot = (): ClockState => this.state;

  set(patch: Partial<ClockState>): void {
    const next = { ...this.state, ...patch };
    if (
      next.timeMs === this.state.timeMs &&
      next.durationMs === this.state.durationMs &&
      next.playing === this.state.playing
    ) {
      return;
    }
    this.state = next;
    this.listeners.forEach((listener) => listener());
  }

  reset(durationMs = 0): void {
    this.set({ timeMs: 0, durationMs, playing: false });
  }
}

export const clock = new ClockStore();

/** Subscribes to the whole clock. Use sparingly: prefer usePlaybackTime. */
export function useClock(): ClockState {
  return useSyncExternalStore(clock.subscribe, clock.getSnapshot, clock.getSnapshot);
}

/** Subscribes to the position in milliseconds only. */
export function usePlaybackTime(): number {
  return useSyncExternalStore(
    clock.subscribe,
    () => clock.getSnapshot().timeMs,
    () => 0,
  );
}

/** Subscribes to the duration only. */
export function useDuration(): number {
  return useSyncExternalStore(
    clock.subscribe,
    () => clock.getSnapshot().durationMs,
    () => 0,
  );
}

/** Subscribes to the playing flag only. */
export function useIsPlaying(): boolean {
  return useSyncExternalStore(
    clock.subscribe,
    () => clock.getSnapshot().playing,
    () => false,
  );
}
