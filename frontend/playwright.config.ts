import { defineConfig } from "@playwright/test";
import os from "node:os";
import path from "node:path";

/**
 * E2E runs against the real Go service in server mode (`-tags server`), not a
 * mocked frontend: bindings, SQLite, the demo library and the media server are
 * all the production code paths, served over HTTP to Chromium.
 *
 * Each run gets its own STEFANI_HOME so tests never touch a real profile.
 */
const e2eHome = process.env.STEFANI_E2E_HOME || path.join(os.tmpdir(), "stefani-e2e-playwright");

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [["github"], ["list"]] : [["list"]],
  use: {
    baseURL: "http://127.0.0.1:8080",
    viewport: { width: 1440, height: 900 },
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  webServer: {
    command: "pnpm run build && cd .. && go run -tags server,production .",
    url: "http://127.0.0.1:8080/",
    reuseExistingServer: !process.env.CI,
    timeout: 240_000,
    env: {
      ...process.env,
      STEFANI_HOME: e2eHome,
      // Tests must never read or write the developer's real stored API key.
      STEFANI_KEYRING_SERVICE: "stefani-e2e",
      // Deterministic and offline: the suite does not hit YouTube.
      STEFANI_DISABLE_ONLINE: "1",
      // Never call a real provider from the suite.
      STEFANI_DISABLE_AI: "1",
      // The generated catalog is a fixture now: tests ask for it explicitly.
      STEFANI_SEED_DEMO: "1",
    } as Record<string, string>,
    stdout: "pipe",
    stderr: "pipe",
  },
});
