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

// Live named targets send the element itself (`<span id="count">1</span>`).
// innerHTML morph inserts that span into #count and Idiomorph throws
// "The new child element contains the parent" (#294).
test("morph outerHTML fallback replaces the element", () => {
  const parent = { innerHTML: "" };
  const el = {
    innerHTML: "0",
    set outerHTML(v) {
      parent.innerHTML = v;
    },
  };
  morph(el, '<span id="count">1</span>', undefined, { morphStyle: "outerHTML" });
  assert.equal(parent.innerHTML, '<span id="count">1</span>');
});
