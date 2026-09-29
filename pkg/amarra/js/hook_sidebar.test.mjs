import test from "node:test";
import assert from "node:assert/strict";
import { makeSidebar } from "./hook_sidebar.mjs";

function setup() {
  const listeners = {};
  const attrs = { "data-amarra-sidebar-target": "#amarra-nav", "aria-expanded": "false" };
  const panel = {
    open: false,
    setAttribute(name) {
      if (name === "data-amarra-sidebar-open") this.open = true;
    },
    removeAttribute(name) {
      if (name === "data-amarra-sidebar-open") this.open = false;
    },
    contains: () => false,
  };
  const doc = {
    querySelector: (sel) => (sel === "#amarra-nav" ? panel : null),
    addEventListener(type, fn) {
      (listeners["doc:" + type] ??= []).push(fn);
    },
    removeEventListener() {},
  };
  const button = {
    focused: false,
    getAttribute: (name) => attrs[name] ?? null,
    setAttribute(name, value) {
      attrs[name] = value;
    },
    addEventListener(type, fn) {
      (listeners["btn:" + type] ??= []).push(fn);
    },
    removeEventListener() {},
    contains: () => false,
    focus() {
      this.focused = true;
    },
    ownerDocument: doc,
  };
  const fire = (key, ev) => (listeners[key] || []).forEach((fn) => fn(ev));
  return { button, panel, listeners, attrs, fire };
}

test("sidebar hook toggles panel state and aria-expanded", () => {
  const { button, panel, attrs, fire } = setup();
  const hook = makeSidebar();
  hook.connect(button);

  assert.equal(attrs["aria-expanded"], "false");
  assert.equal(panel.open, false);

  fire("btn:click");
  assert.equal(attrs["aria-expanded"], "true");
  assert.equal(panel.open, true);

  fire("btn:click");
  assert.equal(attrs["aria-expanded"], "false");
  assert.equal(panel.open, false);
});

test("sidebar hook closes on Escape and returns focus", () => {
  const { button, panel, attrs, fire } = setup();
  const hook = makeSidebar();
  hook.connect(button);
  fire("btn:click");
  assert.equal(panel.open, true);

  fire("doc:keydown", { key: "Escape" });
  assert.equal(panel.open, false);
  assert.equal(attrs["aria-expanded"], "false");
  assert.equal(button.focused, true);
});

test("sidebar hook closes on outside click", () => {
  const { button, panel, fire } = setup();
  const hook = makeSidebar();
  hook.connect(button);
  fire("btn:click");
  assert.equal(panel.open, true);

  fire("doc:click", { target: {} });
  assert.equal(panel.open, false);
});

test("sidebar hook is a no-op without a target panel", () => {
  const { button } = setup();
  const hook = makeSidebar();
  hook.connect({ ...button, getAttribute: () => null });
  assert.equal(button.focused, false);
});

test("sidebar hook disconnect unbinds listeners", () => {
  const { button, listeners } = setup();
  const hook = makeSidebar();
  hook.connect(button);
  const before = (listeners["btn:click"] || []).length;
  hook.disconnect(button);
  assert.equal(button._amarraSidebar, undefined);
  assert.equal(before, 1);
});
