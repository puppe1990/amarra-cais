import test from "node:test";
import assert from "node:assert/strict";
import { makeTheme } from "./hook_theme.mjs";

function classList(initial = []) {
  const set = new Set(initial);
  return {
    contains: (c) => set.has(c),
    add: (c) => set.add(c),
    remove: (c) => set.delete(c),
  };
}

function button() {
  const listeners = {};
  const attrs = {};
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

test("theme hook toggles html.light, persists localStorage, updates theme-color, unbinds on disconnect", () => {
  const html = { classList: classList() };
  const store = new Map();
  const meta = {
    content: "#111111",
    setAttribute(n, v) {
      if (n === "content") this.content = v;
    },
  };
  const hook = makeTheme({
    html: () => html,
    storage: {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, v),
    },
    themeColorMeta: () => meta,
    lightColor: "#f5f5f4",
    darkColor: "#111111",
  });
  const el = button();
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(html.classList.contains("light"), true);
  assert.equal(store.get("amarra-theme"), "light");
  assert.equal(meta.content, "#f5f5f4");
  el.listeners.click[0]();
  assert.equal(html.classList.contains("light"), false);
  assert.equal(store.get("amarra-theme"), "");
  assert.equal(meta.content, "#111111");
  hook.disconnect(el);
  assert.equal((el.listeners.click || []).length, 0);
});

test("theme hook reads theme-color values from data attributes", () => {
  const html = { classList: classList() };
  const store = new Map();
  const meta = {
    content: "#c9893a",
    setAttribute(n, v) {
      if (n === "content") this.content = v;
    },
  };
  const hook = makeTheme({
    html: () => html,
    storage: {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, v),
    },
    themeColorMeta: () => meta,
  });
  const el = button();
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-color") return "#f5f5f4";
    if (n === "data-amarra-theme-color-off") return "#0f172a";
    return origGet(n);
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(meta.content, "#f5f5f4");
  el.listeners.click[0]();
  assert.equal(meta.content, "#0f172a");
});

test("theme hook restores html.light from localStorage on connect", () => {
  const html = { classList: classList() };
  const store = new Map([["amarra-theme", "light"]]);
  const hook = makeTheme({
    html: () => html,
    storage: {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, v),
    },
  });
  hook.connect(button());
  assert.equal(html.classList.contains("light"), true);
});
