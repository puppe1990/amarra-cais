import { test, expect } from "@playwright/test";
import { signIn } from "./helpers.mjs";

test("login swaps the auth shell for the app shell", async ({ page }) => {
  await page.goto("/login");
  await expect(page.locator("body")).toHaveAttribute("data-amarra-shell", "auth");
  await signIn(page);
  await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
});
