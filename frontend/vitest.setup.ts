import "@testing-library/jest-dom/vitest";

// jsdom has no media playback; components under test only need the element to
// exist and to expose the properties the store touches.
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
