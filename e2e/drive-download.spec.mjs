import { readFile } from "node:fs/promises";
import { test, expect } from "@playwright/test";

// #366: a POST export returns the file in the response body. Drive used to
// assign location, which discarded the bytes and made the browser repeat the
// request as GET. The intercept counts requests, so a second GET fails here.
test("Drive saves the response of a POST export without a second GET", async ({ page }) => {
  const requests = [];
  await page.route("**/export-probe", async (route) => {
    const request = route.request();
    requests.push({ method: request.method(), postData: request.postData() });
    await route.fulfill({
      status: 200,
      headers: {
        "content-type": "application/pdf",
        "content-disposition": 'attachment; filename="filtrados.pdf"',
      },
      body: "PDF-BYTES",
    });
  });

  await page.goto("/");
  await page.evaluate(() => {
    const form = document.createElement("form");
    form.method = "post";
    form.action = "/export-probe";
    form.innerHTML = '<input name="filter" value="paid"><button type="submit">Export</button>';
    document.querySelector("#amarra-main").append(form);
  });

  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("button", { name: "Export" }).click();
  const download = await downloadPromise;

  expect(download.suggestedFilename()).toBe("filtrados.pdf");
  const path = await download.path();
  expect(await readFile(path, "utf8")).toBe("PDF-BYTES");
  expect(requests).toEqual([{ method: "POST", postData: "filter=paid" }]);
  await expect(page).toHaveURL(/\/$/);
});
