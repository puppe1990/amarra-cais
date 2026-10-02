import { test, expect } from "@playwright/test";
import { clickInc, morphHome, waitForLiveJoin } from "./helpers.mjs";

test("Live counter works after a Drive morph away and back", async ({ page }) => {
  const joined = waitForLiveJoin(page);
  await page.goto("/live/counter");
  await joined;
  await expect(page.locator("#count")).toHaveText("0");
  await clickInc(page);

  await morphHome(page);

  await page.goBack();
  await expect(page).toHaveURL(/\/live\/counter$/);
  await expect(page.locator("#count")).toBeVisible();
  await clickInc(page);
});
