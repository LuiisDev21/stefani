/** Formats milliseconds as m:ss, or h:mm:ss when the track is long. */
export function formatTime(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) ms = 0;
  const totalSeconds = Math.floor(ms / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }
  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}

/** Formats a study duration as "12 min" or "1 h 05 min". */

/** Relative date used in the library and history lists. */
export function formatRelativeDate(ms: number, now = Date.now()): string {
  if (!ms) return "sin actividad";
  const diff = now - ms;
  const day = 24 * 60 * 60 * 1000;
  if (diff < 60_000) return "hace un momento";
  if (diff < 3_600_000) return `hace ${Math.floor(diff / 60_000)} min`;
  if (diff < day) return `hace ${Math.floor(diff / 3_600_000)} h`;
  if (diff < 7 * day) return `hace ${Math.floor(diff / day)} dias`;
  return new Date(ms).toLocaleDateString("es", { day: "numeric", month: "short" });
}

/** Clamps a number to a range. */
export function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

/** "1 canción" / "3 canciones": a count and its noun never disagree. */
export function plural(count: number, one: string, many: string): string {
  return `${count} ${count === 1 ? one : many}`;
}
