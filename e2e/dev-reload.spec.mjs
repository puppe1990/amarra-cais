import fs from "node:fs";
import path from "node:path";
import { test, expect } from "@playwright/test";

// #340: `amarra-cais dev` re-parses templates in-process and pings the tab over
// /_dev/reload, so saving an HTML file refreshes the page without an air rebuild.
// scripts/e2e-scaffold.sh exports the generated app so the spec can edit it; the
// production pass (--grep @sw) skips this file because it only runs in dev.
test("a template edit reloads the open tab without a rebuild", async ({ page }) => {
  const appDir = process.env.E2E_APP_DIR;
  test.skip(!appDir, "E2E_APP_DIR is exported by scripts/e2e-scaffold.sh");

  await page.goto("/");
  await expect(page.locator("#amarra-main")).toBeVisible();

  const layout = path.join(appDir, "web/templates/layouts/app.html");
  const original = fs.readFileSync(layout, "utf8");
  const marker = `dev-reload-${Date.now()}`;
  fs.writeFileSync(
    layout,
    original.replace('<main id="amarra-main"', `<main id="amarra-main" data-dev-probe="${marker}"`)
  );

  try {
    await expect(page.locator(`#amarra-main[data-dev-probe="${marker}"]`)).toBeVisible({
      timeout: 20_000,
    });
  } finally {
    fs.writeFileSync(layout, original);
  }
});
