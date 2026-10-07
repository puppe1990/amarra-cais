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

// #327: a dependent select fills its options after mount; the listbox must
// rebuild instead of freezing on the first snapshot.
test("select hook rebuilds the list when the native options change", async ({ page }) => {
  await page.goto("/picker");
  await page.evaluate(() => {
    const select = document.querySelector("select[name=distribuidora]");
    const option = document.createElement("option");
    option.value = "neoenergia";
    option.textContent = "Neoenergia";
    select.add(option);
  });

  await page.locator(".cais-select-search-trigger").click();
  await page.locator(".cais-select-search-input").fill("neo");
  const rows = page.locator(".cais-select-search-option:not(.is-hidden)");
  await expect(rows).toHaveCount(1);
  await expect(rows.first()).toContainText("Neoenergia");
});
