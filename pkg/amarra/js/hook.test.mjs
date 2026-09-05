import test from "node:test";
import assert from "node:assert/strict";
import {
  csrfTokenFromMeta,
  showToast,
  applyFocus,
  applyOptimistic,
  rollbackOptimistic,
  start,
} from "./hook.mjs";

test("csrfTokenFromMeta reads meta content", () => {
  assert.equal(csrfTokenFromMeta('<meta name="csrf-token" content="abc">'), "abc");
});

test("csrfTokenFromMeta returns empty when missing", () => {
  assert.equal(csrfTokenFromMeta("<html></html>"), "");
  assert.equal(csrfTokenFromMeta(""), "");
});

function classList(initial = []) {
  const set = new Set(initial);
  return {
    contains: (c) => set.has(c),
    add: (...cs) => cs.forEach((c) => set.add(c)),
    remove: (...cs) => cs.forEach((c) => set.delete(c)),
    toArray: () => [...set],
  };
}

test("applyOptimistic toggle count remove and rollback", () => {
  const toggle = {
    classList: classList(["bg-slate-100", "text-slate-600"]),
    querySelector: () => null,
    textContent: "",
  };
  let toggleState = applyOptimistic(toggle, "toggle");
  assert.equal(toggle.classList.contains("bg-green-50"), true);
  rollbackOptimistic(toggleState);
  assert.equal(toggle.classList.contains("bg-slate-100"), true);

  const countEl = { textContent: "2" };
  const count = {
    classList: classList(),
    querySelector: (sel) => (sel === "[data-amarra-count]" ? countEl : null),
    textContent: "2",
  };
  const countState = applyOptimistic(count, "count");
  assert.equal(countEl.textContent, "3");
  rollbackOptimistic(countState);
  assert.equal(countEl.textContent, "2");

  const remove = { classList: classList() };
  const removeState = applyOptimistic(remove, "remove");
  assert.equal(remove.classList.contains("opacity-0"), true);
  rollbackOptimistic(removeState);
  assert.equal(remove.classList.contains("opacity-0"), false);
});

test("showToast writes into #amarra-toast-host", () => {
  const host = {
    innerHTML: "",
    querySelector(sel) {
      if (sel === "span") return this._span;
      return null;
    },
  };
  host._span = { textContent: "" };
  const doc = {
    getElementById(id) {
      return id === "amarra-toast-host" ? host : null;
    },
  };
  showToast("Hello", doc, { duration: 0 });
  assert.match(host.innerHTML, /role="status"/);
  assert.equal(host._span.textContent, "Hello");
});

test("applyFocus focuses matching selector", () => {
  let focused = false;
  const el = {
    focus() {
      focused = true;
    },
  };
  const doc = {
    querySelector(sel) {
      return sel === "#email" ? el : null;
    },
  };
  applyFocus("#email", doc);
  assert.equal(focused, true);
});

test("start is a no-op without a document", () => {
  assert.equal(start({ document: null }), undefined);
});
