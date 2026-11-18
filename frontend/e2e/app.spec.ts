import { expect, test, type Page } from "@playwright/test";

/**
 * End-to-end coverage of the MVP's core flows against the real Go backend.
 * Audio decoding is not asserted: headless Chromium has no audio device, so the
 * tests assert the states the app derives from the catalog and the database.
 */

async function waitForApp(page: Page) {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "Inicio" })).toBeVisible();
  await expect(page.getByText("Accesos rápidos")).toBeVisible();
}

test.describe("STEFANI desktop shell", () => {
  test("loads the library from the Go service", async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on("console", (message) => {
      if (message.type() === "error") consoleErrors.push(message.text());
    });

    await waitForApp(page);

    // The real counters live in the sidebar; the header no longer repeats them.
    await expect(page.locator("aside").getByText(/^\d+ pistas$/)).toBeVisible();
    await expect(page.getByRole("button", { name: /Reproducir Midnight Signals/ })).toBeVisible();
    await expect(page.getByRole("button", { name: /Reproducir Paper Boats/ })).toBeVisible();
    await expect(page.getByRole("button", { name: /Reproducir Slow Morning/ })).toBeVisible();
    expect(consoleErrors).toEqual([]);
  });

  test("navigates every section and keeps aria-current on the active one", async ({ page }) => {
    await waitForApp(page);

    for (const [label, heading] of [
      ["Biblioteca", "Biblioteca"],
      ["Listas", "Listas"],
      ["Favoritos", "Favoritos"],
      ["Ajustes", "Ajustes"],
      ["Inicio", "Inicio"],
    ] as const) {
      await page.getByRole("button", { name: label, exact: true }).click();
      await expect(page.getByRole("heading", { name: heading, level: 1 })).toBeVisible();
      await expect(page.getByRole("button", { name: label, exact: true })).toHaveAttribute(
        "aria-current",
        "page",
      );
    }
  });

  test("shows word-synced lyrics with IPA for a demo track", async ({ page }) => {
    await waitForApp(page);
    await page.getByRole("button", { name: /Reproducir Paper Boats/ }).click();

    const panel = page.locator("aside[data-lyrics-state]");
    await expect(panel.getByText("palabra por palabra")).toBeVisible();
    await expect(panel.getByText("/wiː ˈfoʊldɪd wɜːrdz ˈɪntuː ə boʊt/")).toBeVisible();
    await expect(panel.getByText("uii fóuldid uérds íntu a bóut")).toBeVisible();
    await expect(panel.getByText("Doblamos palabras en un barco")).toBeVisible();
  });

  test("creates a playlist that survives a reload", async ({ page }) => {
    const name = `Rutina ${Date.now()}`;
    await waitForApp(page);
    await page.getByRole("button", { name: "Listas", exact: true }).click();
    await page.getByRole("button", { name: "Nueva lista" }).click();
    await page.getByLabel("Nombre de la lista").fill(name);
    await page.getByRole("button", { name: "Crear", exact: true }).click();

    await expect(page.getByRole("heading", { name, level: 1 })).toBeVisible();

    await page.reload();
    await page.getByRole("button", { name: "Listas", exact: true }).click();
    await expect(page.getByText(name).first()).toBeVisible();
  });

  test("asks for AI setup instead of failing when no key is stored", async ({ page }) => {
    await waitForApp(page);
    await page.getByRole("button", { name: /Reproducir Midnight Signals/ }).click();
    const panel = page.locator("aside[data-lyrics-state]");
    await panel.getByRole("button", { name: "Generar traducción" }).click();

    await expect(page.getByRole("status").filter({ hasText: /Configura|clave/i })).toBeVisible();
    await expect(page.getByRole("heading", { name: "Ajustes", level: 1 })).toBeVisible();
  });

  test("rejects an insecure AI endpoint before saving", async ({ page }) => {
    await waitForApp(page);
    await page.getByRole("button", { name: "Ajustes", exact: true }).click();

    await page.locator('input[placeholder="https://api.openai.com/v1"]').fill("http://example.com/v1");
    await page.getByRole("button", { name: "Guardar cambios" }).click();

    await expect(page.getByRole("alert")).toContainText("https");
  });

  test("searches from the bar and opens the results under it", async ({ page }) => {
    await waitForApp(page);
    await page.getByLabel("Buscar canciones").fill("paper boats");

    const results = page.getByRole("listbox", { name: "Resultados de búsqueda" });
    await expect(results).toBeVisible();
    await expect(results.getByRole("option", { name: /Paper Boats/ }).first()).toBeVisible();
    await expect(results.getByRole("option", { name: /Slow Morning/ })).toHaveCount(0);

    // Picking one plays it and folds the list away.
    await results.getByRole("option", { name: /Paper Boats/ }).first().click();
    await expect(results).toBeHidden();
    await expect(page.getByLabel("Buscar canciones")).toHaveValue("");
  });
});
