import { expect } from "@playwright/test";

export async function signIn(page) {
  await page.goto("/login");
  await page.locator("#email").fill("demo@example.com");
  await page.locator("#password").fill("password");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/dashboard/);
  await expect(page.locator("body")).toHaveAttribute("data-amarra-shell", "app");
}

// Contact lives only on the home page; the header brand is on every layout.
export async function morphHome(page) {
  await page.locator('header a[href="/"]').click();
  await expect(page).toHaveURL(/\/$/);
}

export async function waitForLiveJoin(page) {
  const ws = await page.waitForEvent("websocket", {
    predicate: (socket) => socket.url().includes("/amarra/live"),
  });
  await ws.waitForEvent("framereceived", {
    predicate: (event) => {
      try {
        const msg = JSON.parse(event.payload);
        return msg.type === "ok" || msg.type === "morph";
      } catch {
        return false;
      }
    },
  });
  return ws;
}

export async function clickInc(page) {
  const now = await page.locator("#count").innerText();
  await page.getByRole("button", { name: "Inc" }).click();
  // A Playwright retry can fire two incs if Idiomorph mutates #count mid-click.
  await expect(page.locator("#count")).not.toHaveText(now);
}
