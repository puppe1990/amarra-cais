import test from "node:test";
import assert from "node:assert/strict";
import { morph } from "./morph.mjs";

test("morph uses injected morphFn", () => {
  const el = { innerHTML: "old" };
  let called = false;
  morph(el, "<p>n</p>", (node, html) => {
    called = true;
    node.innerHTML = html;
  });
  assert.equal(called, true);
  assert.equal(el.innerHTML, "<p>n</p>");
});

test("morph falls back to innerHTML when Idiomorph is absent", () => {
  const el = { innerHTML: "old" };
  morph(el, "<b>x</b>");
  assert.equal(el.innerHTML, "<b>x</b>");
});
