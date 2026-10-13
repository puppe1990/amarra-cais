import { test, expect } from "@playwright/test";

test("Drive morphs an anonymous page without a full reload and restores history", async ({
  page,
}) => {
  await page.goto("/");
  await page.evaluate(() => {
    window.__amarraDoc = true;
  });
  await page.getByRole("link", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/login$/);
  await expect(page.locator("#password")).toBeVisible();
  expect(await page.evaluate(() => window.__amarraDoc)).toBe(true);

  await page.goBack();
  await expect(page).toHaveURL(/\/$/);
  expect(await page.evaluate(() => window.__amarraDoc)).toBe(true);
});

test("422 signup submit focuses the first aria-invalid field", async ({ page }) => {
  await page.goto("/signup");
  await page.getByRole("button", { name: "Sign up" }).click();
  await expect(page.locator("#email")).toHaveAttribute("aria-invalid", "true");
  await expect(page.locator("#email")).toBeFocused();
});
