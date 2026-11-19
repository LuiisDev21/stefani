import "@testing-library/jest-dom/vitest";

// jsdom has no media playback; components under test only need the element to
// exist and to expose the properties the store touches. Node tests (the
// offline audio renders) have no window at all, so this is guarded.
if (typeof window !== "undefined" && typeof window.HTMLMediaElement !== "undefined") {
  Object.defineProperty(window.HTMLMediaElement.prototype, "play", {
    configurable: true,
    value: () => Promise.resolve(),
  });
  Object.defineProperty(window.HTMLMediaElement.prototype, "pause", {
    configurable: true,
    value: () => undefined,
  });
  Object.defineProperty(window.HTMLMediaElement.prototype, "load", {
    configurable: true,
    value: () => undefined,
  });
}
