import { test, expect } from "@playwright/test";

test("select hook searches and writes the choice back to the native select", async ({ page }) => {
  await page.goto("/picker");
  const native = page.locator("select[name=distribuidora]");
  const trigger = page.locator(".cais-select-search-trigger");
  await expect(trigger).toBeVisible();

  await trigger.click();
  const search = page.locator(".cais-select-search-input");
  await search.fill("ene");

  const rows = page.locator(".cais-select-search-option:not(.is-hidden)");
  await expect(rows).toHaveCount(1);
  await rows.first().click();

  await expect(native).toHaveValue("enel");
  await expect(trigger).toContainText("Enel");
  await expect(page.locator(".cais-select-search-panel")).toBeHidden();
});
