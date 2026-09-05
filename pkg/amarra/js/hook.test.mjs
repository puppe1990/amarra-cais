import test from "node:test";
import assert from "node:assert/strict";
import {
  csrfTokenFromMeta,
  showToast,
  applyFocus,
  applyOptimistic,
  rollbackOptimistic,
  start,
  register,
  reset,
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

test("start registers password and theme builtins", () => {
  reset();
  const click = {};
  const input = { type: "password" };
  const html = {
    classList: {
      contains: () => false,
      add() {},
      remove() {},
    },
  };
  const passwordBtn = {
    getAttribute(n) {
      if (n === "amarra-hook") return "password";
      if (n === "data-amarra-password-for") return "#password";
      return null;
    },
    hasAttribute(n) {
      return n === "amarra-hook";
    },
    addEventListener(type, fn) {
      (click[type] ??= []).push(fn);
    },
    removeEventListener() {},
    setAttribute() {},
    ownerDocument: {
      querySelector(sel) {
        return sel === "#password" ? input : null;
      },
    },
  };
  const themeBtn = {
    getAttribute(n) {
      return n === "amarra-hook" ? "theme" : null;
    },
    hasAttribute(n) {
      return n === "amarra-hook";
    },
    addEventListener(type, fn) {
      (click["theme-" + type] ??= []).push(fn);
    },
    removeEventListener() {},
  };
  const doc = {
    documentElement: { dataset: {}, classList: html.classList },
    addEventListener() {},
    querySelectorAll(sel) {
      return sel === "[amarra-hook]" ? [passwordBtn, themeBtn] : [];
    },
  };
  start({ document: doc });
  assert.ok(click.click?.length, "password hook should bind click");
  click.click[0]();
  assert.equal(input.type, "text");
  assert.ok(click["theme-click"]?.length, "theme hook should bind click");
});

test("start scans amarra-hook and rescans after amarra:morphed", () => {
  reset();
  const events = [];
  register("clip", {
    connect(el) {
      events.push("connect:" + el.id);
    },
    updated(el) {
      events.push("updated:" + el.id);
    },
  });
  const btn = {
    id: "btn",
    getAttribute(n) {
      return n === "amarra-hook" ? "clip" : null;
    },
    hasAttribute(n) {
      return n === "amarra-hook";
    },
  };
  const listeners = {};
  const doc = {
    listeners,
    documentElement: { dataset: {} },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    dispatchEvent(ev) {
      for (const fn of listeners[ev.type] || []) fn(ev);
      return true;
    },
    querySelectorAll(sel) {
      return sel === "[amarra-hook]" ? [btn] : [];
    },
  };
  start({ document: doc });
  assert.deepEqual(events, ["connect:btn"]);
  doc.dispatchEvent(new CustomEvent("amarra:morphed"));
  assert.deepEqual(events, ["connect:btn", "updated:btn"]);
});

test("amarra:drive-error rolls back optimistic UI from start()", () => {
  const listeners = {};
  const el = {
    classList: classList(["bg-slate-100", "text-slate-600"]),
    querySelector: () => null,
    closest() {
      return this;
    },
    getAttribute() {
      return "toggle";
    },
    setAttribute() {},
  };
  const doc = {
    listeners,
    documentElement: { dataset: {} },
    addEventListener(type, fn, _opts) {
      (listeners[type] ??= []).push(fn);
    },
    dispatchEvent(ev) {
      for (const fn of listeners[ev.type] || []) fn(ev);
      return true;
    },
  };
  start({ document: doc });
  for (const fn of listeners.click) fn({ target: el });
  assert.equal(el.classList.contains("bg-green-50"), true);
  doc.dispatchEvent(new CustomEvent("amarra:drive-error"));
  assert.equal(el.classList.contains("bg-slate-100"), true);
});
