import { test, expect } from "@playwright/test";

test.describe("service worker", { tag: "@sw" }, () => {
  test("registers, serves amarra.js offline, and skips no-store HTML", async ({
    page,
    context,
  }) => {
    await page.goto("/login");
    const registration = await page.evaluate(async () => {
      const ready = await navigator.serviceWorker.ready;
      return Boolean(ready.active);
    });
    expect(registration).toBe(true);

    const login = await page.request.get("/login");
    expect(login.headers()["cache-control"] || "").toMatch(/no-store/i);

    await page.evaluate(async () => {
      const cache = await caches.open((await caches.keys())[0]);
      await cache.add("/static/js/amarra.js");
    });

    await context.setOffline(true);
    const js = await page.evaluate(async () => {
      const res = await fetch("/static/js/amarra.js");
      return { ok: res.ok, status: res.status };
    });
    expect(js.ok).toBe(true);

    await page.goto("/login", { waitUntil: "domcontentloaded" }).catch(() => {});
    await expect(page.getByRole("heading", { name: "You are offline" })).toBeVisible();
  });

  test("logout message clears the static cache", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(async () => {
      await navigator.serviceWorker.ready;
      const keys = await caches.keys();
      if (keys[0]) {
        const cache = await caches.open(keys[0]);
        await cache.put("/static/js/amarra.js", new Response("cached"));
      }
    });

    await page.evaluate(async () => {
      const reg = await navigator.serviceWorker.ready;
      reg.active.postMessage({ type: "amarra:clear-cache" });
    });

    await expect
      .poll(async () =>
        page.evaluate(async () => {
          const keys = await caches.keys();
          return keys.length;
        })
      )
      .toBe(0);
  });
});
