import test from "node:test";
import assert from "node:assert/strict";
import { makeDropdown } from "./hook_dropdown.mjs";

function button() {
  const listeners = {};
  return {
    listeners,
    attrs: { "aria-expanded": "false" },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
    setAttribute(n, v) {
      this.attrs[n] = String(v);
    },
    click() {
      (listeners.click || []).forEach((fn) => fn({ preventDefault() {} }));
    },
  };
}

function menu() {
  const listeners = {};
  return {
    listeners,
    hidden: true,
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
    click() {
      (listeners.click || []).forEach((fn) => fn({ preventDefault() {} }));
    },
  };
}

function documentStub() {
  return {
    listeners: {},
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      this.listeners[type] = (this.listeners[type] || []).filter((f) => f !== fn);
    },
    fire(type, ev) {
      (this.listeners[type] || []).forEach((fn) => fn(ev));
    },
  };
}

function container(btn, menuEl, doc) {
  return {
    contains(t) {
      return t === btn || t === menuEl;
    },
    querySelector(sel) {
      if (sel === "[data-amarra-dropdown-button]") return btn;
      if (sel === "[data-amarra-dropdown-menu]") return menuEl;
      return null;
    },
    ownerDocument: doc,
  };
}

function setup() {
  const btn = button();
  const menuEl = menu();
  const doc = documentStub();
  const el = container(btn, menuEl, doc);
  makeDropdown().connect(el);
  return { btn, menuEl, doc, el };
}

test("dropdown hook toggles menu and aria-expanded on button click", () => {
  const { btn, menuEl } = setup();
  assert.equal(menuEl.hidden, true);

  btn.click();
  assert.equal(menuEl.hidden, false);
  assert.equal(btn.attrs["aria-expanded"], "true");

  btn.click();
  assert.equal(menuEl.hidden, true);
  assert.equal(btn.attrs["aria-expanded"], "false");
});

test("dropdown hook closes on Escape and on clicks outside the container", () => {
  const { btn, menuEl, doc } = setup();
  btn.click();
  assert.equal(menuEl.hidden, false);

  doc.fire("keydown", { key: "Escape", preventDefault() {} });
  assert.equal(menuEl.hidden, true);
  assert.equal(btn.attrs["aria-expanded"], "false");

  btn.click();
  assert.equal(menuEl.hidden, false);
  doc.fire("click", { target: { inside: false }, preventDefault() {} });
  assert.equal(menuEl.hidden, true);
});

test("dropdown hook keeps the menu open for clicks inside the container", () => {
  const { btn, menuEl, doc } = setup();
  btn.click();
  doc.fire("click", { target: menuEl, preventDefault() {} });
  assert.equal(menuEl.hidden, false, "outside-close must ignore clicks inside the hook container");
});

test("dropdown hook closes when a menu item is clicked and unbinds on disconnect", () => {
  const { btn, menuEl, doc, el } = setup();
  btn.click();
  menuEl.click();
  assert.equal(menuEl.hidden, true);

  btn.click();
  makeDropdown().disconnect(el);
  assert.equal((btn.listeners.click || []).length, 0);
  assert.equal((menuEl.listeners.click || []).length, 0);
  assert.equal((doc.listeners.click || []).length, 0);
  assert.equal((doc.listeners.keydown || []).length, 0);
});
