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

test("password hook ignores invalid selectors instead of throwing", () => {
  const hook = makePassword();
  const el = button("#");
  el.ownerDocument = {
    querySelector() {
      throw new Error("not a valid selector");
    },
  };
  hook.connect(el);
  assert.doesNotThrow(() => el.listeners.click[0]());
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

// #132: the fallback used parentElement.querySelector("input") — the first
// input of the wrapper, which may be the email field. Prefer the password
// field in the same form/wrapper.
test("password hook fallback picks the password input, not the first input", () => {
  const email = { type: "email" };
  const passwordInput = { type: "password" };
  const hook = makePassword();
  const el = button("");
  el.parentElement = {
    querySelectorAll(sel) {
      return sel === 'input[type="password"]' ? [passwordInput] : [];
    },
    querySelector() {
      return email; // what the old fallback would have toggled
    },
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(passwordInput.type, "text");
  assert.equal(email.type, "email", "email field must not be toggled");
  el.listeners.click[0]();
  assert.equal(passwordInput.type, "password");
});

test("password hook fallback never toggles a form without password fields", () => {
  const email = { type: "email" };
  const hook = makePassword();
  const el = button("");
  el.parentElement = {
    querySelectorAll() {
      return [];
    },
    querySelector() {
      return email;
    },
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(email.type, "email");
  assert.equal(el.getAttribute("aria-pressed"), "false");
});

test("password hook picks the closest preceding password field in a form", () => {
  const pw = { type: "password" };
  const confirm = { type: "password" };
  const hook = makePassword();
  const el = button("");
  el.closest = (sel) => (sel === "form" ? form : null);
  const form = {
    querySelectorAll() {
      return [pw, confirm];
    },
  };
  el.compareDocumentPosition = (node) => (node === pw ? 2 : 4); // pw precedes, confirm follows
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(pw.type, "text");
  assert.equal(confirm.type, "password");
});

test("password hook swaps aria-label from data-amarra-label-show/hide (#28)", () => {
  // Labels follow the visible affordance (the next action), like the icons:
  // hidden input shows the "show" label, visible input the "hide" label.
  const input = { type: "password" };
  const hook = makePassword((sel) => (sel === "#password" ? input : null));
  const el = button("#password");
  el.setAttribute("data-amarra-label-show", "Show password");
  el.setAttribute("data-amarra-label-hide", "Hide password");
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(el.getAttribute("aria-label"), "Hide password");
  el.listeners.click[0]();
  assert.equal(el.getAttribute("aria-label"), "Show password");
});

test("password hook accepts data-amarra-password-icon aliases (#28)", () => {
  const input = { type: "password" };
  const showIcon = { classList: classSet() };
  const hideIcon = { classList: classSet(["hidden"]) };
  const hook = makePassword((sel) => (sel === "#password" ? input : null));
  const el = button("#password");
  el.querySelector = (sel) => {
    if (sel === '[data-amarra-password-icon="show"]') return showIcon;
    if (sel === '[data-amarra-password-icon="hide"]') return hideIcon;
    return null;
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(showIcon.classList.contains("hidden"), true);
  assert.equal(hideIcon.classList.contains("hidden"), false);
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
