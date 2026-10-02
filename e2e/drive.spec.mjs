import { test, expect } from "@playwright/test";

test("Drive morphs contact without a full reload and restores history", async ({ page }) => {
  await page.goto("/");
  await page.evaluate(() => {
    window.__amarraDoc = true;
  });
  await page.getByRole("link", { name: "Contact" }).click();
  await expect(page).toHaveURL(/\/contact$/);
  await expect(page.locator("#name")).toBeVisible();
  expect(await page.evaluate(() => window.__amarraDoc)).toBe(true);

  await page.goBack();
  await expect(page).toHaveURL(/\/$/);
  expect(await page.evaluate(() => window.__amarraDoc)).toBe(true);
});

test("422 contact submit focuses the first aria-invalid field", async ({ page }) => {
  await page.goto("/contact");
  await page.getByRole("button", { name: "Send" }).click();
  await expect(page.locator("#name")).toHaveAttribute("aria-invalid", "true");
  await expect(page.locator("#name")).toBeFocused();
});
