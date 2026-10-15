import { test, expect } from "@playwright/test";
import { morphHome, signIn } from "./helpers.mjs";

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

// #330: inline SVG has to keep its camelCase attributes through a Drive morph —
// the illustration is drawn from `viewBox`, and losing it leaves the box empty.
test("Drive morph keeps camelCase SVG attributes of inline SVG", async ({ page }) => {
  await signIn(page);
  await morphHome(page);
  const sheet = page.locator('#amarra-main svg[viewBox="0 0 200 200"]');
  await expect(sheet).toHaveCount(1);
  expect(await sheet.evaluate((el) => el.getAttribute("viewBox"))).toBe("0 0 200 200");
});

// The password fields are `required`, so native validation would block an empty
// submit: the passwords are filled and the (optional per HTML5) email is left
// empty, which is what the server rejects with 422.
test("422 signup submit focuses the first aria-invalid field", async ({ page }) => {
  await page.goto("/signup");
  await page.locator("#password").fill("password123");
  await page.locator("#password_confirmation").fill("password123");
  await page.getByRole("button", { name: "Sign up" }).click();
  await expect(page.locator("#email")).toHaveAttribute("aria-invalid", "true");
  await expect(page.locator("#email")).toBeFocused();
});

for (const confirmedElement of ["form", "submitter", "link"]) {
  test(`cancel ${confirmedElement} confirmation blocks native action`, async ({ page }) => {
    await page.goto("/login");
    const requests = [];
    page.on("request", (request) => {
      if (request.url().includes("/cancel-probe")) requests.push(request.url());
    });
    page.on("dialog", (dialog) => dialog.dismiss());
    await page.evaluate((element) => {
      const confirm = 'data-amarra-confirm="Delete?"';
      document.querySelector("#amarra-main").innerHTML =
        element === "link"
          ? `<a href="/cancel-probe" ${confirm}>Cancel probe</a>`
          : `<form action="/cancel-probe" method="post" ${element === "form" ? confirm : ""}>
            <button ${element === "submitter" ? confirm : ""}>Cancel probe</button></form>`;
      document.addEventListener(
        element === "link" ? "click" : "submit",
        (event) => {
          window.__cancelPrevented = event.defaultPrevented;
        },
        { once: true }
      );
    }, confirmedElement);
    await page.getByText("Cancel probe", { exact: true }).click();
    expect(await page.evaluate(() => window.__cancelPrevented)).toBe(true);
    await expect(page).toHaveURL(/\/login$/);
    expect(requests).toEqual([]);
  });
}

test("Drive removes flash without reusing it as the next page wrapper", async ({ page }) => {
  await page.goto("/login");
  // Native login consumes the one-shot flash on the destination, without the
  // shell-change Drive fallback making a second GET before this probe.
  await page
    .locator("form:has(#email)")
    .evaluate((form) => form.setAttribute("data-amarra-skip", ""));
  await page.locator("#email").fill("demo@example.com");
  await page.locator("#password").fill("password");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  const flash = page.locator("#amarra-flash");
  await expect(flash).toBeVisible();
  await page.evaluate(() => {
    window.__oldFlash = document.getElementById("amarra-flash");
    const link = document.createElement("a");
    link.href = "/dashboard";
    link.textContent = "Flash navigation probe";
    document.querySelector("#amarra-main").append(link);
  });
  await page.getByText("Flash navigation probe", { exact: true }).click();
  await expect(flash).toHaveCount(0);
  expect(await page.evaluate(() => window.__oldFlash.isConnected)).toBe(false);
  await page.evaluate(() => window.__oldFlash.remove());
  await expect(page.locator("#amarra-main")).not.toBeEmpty();
});
