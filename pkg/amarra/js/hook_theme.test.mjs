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
  const el = {
    listeners,
    children: [],
    _text: undefined,
    get textContent() {
      return this._text;
    },
    set textContent(v) {
      this._text = v;
      this.children = [];
    },
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
  return el;
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

test("theme hook reads storage key and class from data attributes (#30)", () => {
  const html = { classList: classList() };
  const store = new Map();
  const hook = makeTheme({
    html: () => html,
    storage: {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, v),
    },
  });
  const el = button();
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-key") return "app-theme";
    if (n === "data-amarra-theme-class") return "theme-on";
    return origGet(n);
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(html.classList.contains("theme-on"), true);
  assert.equal(html.classList.contains("light"), false);
  assert.equal(store.get("app-theme"), "theme-on");
  assert.equal(store.has("amarra-theme"), false);
});

test("theme hook restores using the data-attr storage key on connect (#30)", () => {
  const html = { classList: classList() };
  const store = new Map([["app-theme", "dark-mode"]]);
  const hook = makeTheme({
    html: () => html,
    storage: {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, v),
    },
  });
  const el = button();
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-key") return "app-theme";
    if (n === "data-amarra-theme-class") return "dark-mode";
    return origGet(n);
  };
  hook.connect(el);
  assert.equal(html.classList.contains("dark-mode"), true);
});

test("theme hook does not wipe svg children when swapping labels (#41)", () => {
  const html = { classList: classList() };
  const hook = makeTheme({ html: () => html, storage: newStorage() });
  const el = button();
  const svg = { nodeName: "SVG" };
  el.children = [svg];
  el.querySelector = () => null;
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-on-label") return "Dark mode";
    if (n === "data-amarra-theme-off-label") return "Light mode";
    return origGet(n);
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(el.children[0], svg);
  assert.equal(el.textContent, undefined);
  assert.equal(el.getAttribute("aria-pressed"), "true");
});

test("theme hook swaps [data-amarra-theme-label] instead of the whole button (#41)", () => {
  const html = { classList: classList() };
  const hook = makeTheme({ html: () => html, storage: newStorage() });
  const el = button();
  const slot = { textContent: "Light mode" };
  const svg = { nodeName: "SVG" };
  el.children = [svg, slot];
  el.querySelector = (sel) => (sel === "[data-amarra-theme-label]" ? slot : null);
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-on-label") return "Dark mode";
    if (n === "data-amarra-theme-off-label") return "Light mode";
    return origGet(n);
  };
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(slot.textContent, "Dark mode");
  assert.equal(el.children[0], svg);
  assert.equal(el.textContent, undefined);
});

test("theme hook swaps labels and keeps aria-pressed (#30)", () => {
  const html = { classList: classList() };
  const hook = makeTheme({ html: () => html, storage: newStorage() });
  const el = button();
  const origGet = el.getAttribute.bind(el);
  el.getAttribute = (n) => {
    if (n === "data-amarra-theme-on-label") return "Dark mode";
    if (n === "data-amarra-theme-off-label") return "Light mode";
    return origGet(n);
  };
  el.textContent = "Light mode";
  hook.connect(el);
  el.listeners.click[0]();
  assert.equal(el.textContent, "Dark mode");
  assert.equal(el.getAttribute("aria-pressed"), "true");
  el.listeners.click[0]();
  assert.equal(el.textContent, "Light mode");
  assert.equal(el.getAttribute("aria-pressed"), "false");
});

function newStorage() {
  const store = new Map();
  return {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, v),
  };
}
