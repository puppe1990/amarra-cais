import { test, expect } from "@playwright/test";

test("dialog hook opens a native dialog, Esc closes, focus returns", async ({ page }) => {
  await page.goto("/dialog");
  const opener = page.getByRole("button", { name: "Open dialog" });
  await opener.click();
  const dialog = page.locator("dialog");
  await expect(dialog).toBeVisible();
  await expect(dialog).toHaveJSProperty("open", true);

  await page.keyboard.press("Escape");
  await expect(dialog).toBeHidden();
  await expect(opener).toBeFocused();
});
