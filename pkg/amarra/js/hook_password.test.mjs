import test from "node:test";
import assert from "node:assert/strict";
import { makePassword } from "./hook_password.mjs";

function button(sel) {
  const listeners = {};
  const attrs = {
    "data-amarra-password-for": sel,
    "aria-pressed": "false",
  };
  return {
    listeners,
    getAttribute(n) {
      return Object.hasOwn(attrs, n) ? attrs[n] : null;
    },
    setAttribute(n, v) {
      attrs[n] = String(v);
    },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
  };
}

test("password hook toggles input type and aria-pressed on click and unbinds on disconnect", () => {
  const input = { type: "password" };
  const hook = makePassword((sel) => (sel === "#password" ? input : null));
  const el = button("#password");
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(input.type, "text");
  assert.equal(el.getAttribute("aria-pressed"), "true");
  el.listeners.click[0]();
  assert.equal(input.type, "password");
  assert.equal(el.getAttribute("aria-pressed"), "false");
  hook.disconnect(el);
  assert.equal((el.listeners.click || []).length, 0);
});

test("password hook toggles show/hide icons when present", () => {
  const input = { type: "password" };
  const showIcon = { classList: classSet() };
  const hideIcon = { classList: classSet(["hidden"]) };
  const hook = makePassword((sel) => (sel === "#password" ? input : null));
  const el = button("#password");
  el.querySelector = (sel) => {
    if (sel === '[data-cais-password-icon="show"]') return showIcon;
    if (sel === '[data-cais-password-icon="hide"]') return hideIcon;
    return null;
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(showIcon.classList.contains("hidden"), true);
  assert.equal(hideIcon.classList.contains("hidden"), false);
  el.listeners.click[0]();
  assert.equal(showIcon.classList.contains("hidden"), false);
  assert.equal(hideIcon.classList.contains("hidden"), true);
});

function classSet(initial = []) {
  const set = new Set(initial);
  return {
    contains: (c) => set.has(c),
    add: (c) => set.add(c),
    remove: (c) => set.delete(c),
    toggle(c, force) {
      if (force === true) set.add(c);
      else if (force === false) set.delete(c);
      else if (set.has(c)) set.delete(c);
      else set.add(c);
    },
  };
}
