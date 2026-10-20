import { test, expect } from "@playwright/test";

// #369: Idiomorph reuses a node whose id exists in both responses. The morph
// resets value/attributes from the new HTML and drops JS-only data-* attributes,
// but the node — with its listeners and expando properties — stays alive.
test("Drive morph reuses same-id nodes and resets JS-only attributes", async ({ page }) => {
  await page.goto("/morphone");
  await expect(page.locator("#morph-field")).toHaveValue("Morphone");

  await page.evaluate(() => {
    const field = document.querySelector("#morph-field");
    const select = document.querySelector("#morph-select");
    window.__morphNode = field;
    window.__morphSelectNode = select;
    window.__morphMarker = "kept";
    window.__morphInputs = 0;
    field._morphState = "js-property";
    field.setAttribute("data-js-only", "yes");
    field.value = "typed by user";
    select.value = "RJ";
    field.addEventListener("input", () => {
      window.__morphInputs += 1;
    });
  });

  await page.locator("#to-other").click();
  await expect(page).toHaveURL(/\/morptwo$/);
  await expect(page.locator("#morph-field")).toHaveValue("Morptwo");

  const result = await page.evaluate(() => {
    const field = document.querySelector("#morph-field");
    const select = document.querySelector("#morph-select");
    field.dispatchEvent(new Event("input", { bubbles: true }));
    return {
      sameNode: field === window.__morphNode,
      sameSelectNode: select === window.__morphSelectNode,
      marker: window.__morphMarker,
      value: field.value,
      server: field.getAttribute("data-server"),
      jsOnly: field.hasAttribute("data-js-only"),
      property: field._morphState,
      selectValue: select.value,
      inputs: window.__morphInputs,
    };
  });

  expect(result.sameNode).toBe(true);
  expect(result.sameSelectNode).toBe(true);
  expect(result.marker).toBe("kept");
  expect(result.value).toBe("Morptwo");
  expect(result.server).toBe("Morptwo");
  expect(result.jsOnly).toBe(false);
  expect(result.property).toBe("js-property");
  expect(result.selectValue).toBe("SP");
  expect(result.inputs).toBe(1);
});
