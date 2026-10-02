import { test, expect } from "@playwright/test";
import { morphHome, signIn } from "./helpers.mjs";

test("Stream EventSource reconnects after a Drive morph", async ({ page }) => {
  await signIn(page);
  await page.goto("/chat");
  const opened = page.waitForRequest(/\/chat\/\d+\/stream/);
  await page.getByRole("button", { name: "New chat" }).click();
  await expect(page).toHaveURL(/\/chat\/\d+$/);
  // Unstyled e2e CSS leaves this flex child at 0 height; attached is the contract.
  await expect(page.locator("#chat-messages[data-amarra-stream]")).toBeAttached();
  await opened;

  await morphHome(page);

  await page.goBack();
  await expect(page).toHaveURL(/\/chat\/\d+$/);
  await expect(page.locator("#chat-messages[data-amarra-stream]")).toBeAttached();
  await page.locator("textarea[name=content]").fill("hello from e2e");
  await page.getByRole("button", { name: "Send" }).click();
  await expect(page.locator("#chat-messages")).toContainText("hello from e2e");
});
