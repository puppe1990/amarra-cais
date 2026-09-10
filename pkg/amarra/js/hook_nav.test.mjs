import test from "node:test";
import assert from "node:assert/strict";
import { makeNav } from "./hook_nav.mjs";

function link(href) {
  const attrs = {};
  if (href != null) attrs.href = href;
  const classes = new Set();
  return {
    attrs,
    classes,
    getAttribute(n) {
      return Object.hasOwn(attrs, n) ? attrs[n] : null;
    },
    setAttribute(n, v) {
      attrs[n] = String(v);
    },
    removeAttribute(n) {
      delete attrs[n];
    },
    classList: {
      contains: (c) => classes.has(c),
      add: (c) => classes.add(c),
      remove: (c) => classes.delete(c),
    },
  };
}

function container(links, attrs = {}) {
  return {
    attrs,
    links,
    getAttribute(n) {
      return Object.hasOwn(attrs, n) ? attrs[n] : null;
    },
    querySelectorAll(sel) {
      return sel === "a[href]" ? [...links] : [];
    },
  };
}

function location(pathname) {
  return { href: `http://localhost${pathname}`, pathname };
}

function window() {
  return {
    listeners: {},
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      this.listeners[type] = (this.listeners[type] || []).filter((f) => f !== fn);
    },
  };
}

test("nav hook marks the link matching location.pathname (#27)", () => {
  const ledger = link("/dashboard");
  const iam = link("/settings");
  const el = container([ledger, iam], {
    "data-amarra-nav-on": "text-copper",
    "data-amarra-nav-off": "text-foam/50",
  });
  ledger.classes.add("text-copper"); // SSR first paint says dashboard is active
  makeNav({ location: () => location("/settings"), window: window() }).connect(el);

  // We are on /settings now: iam gets the on-classes and aria-current,
  // ledger loses them and gains the off-classes.
  assert.equal(iam.classes.has("text-copper"), true);
  assert.equal(iam.attrs["aria-current"], "page");
  assert.equal(iam.classes.has("text-foam/50"), false);
  assert.equal(ledger.classes.has("text-copper"), false);
  assert.equal(ledger.classes.has("text-foam/50"), true);
  assert.equal("aria-current" in ledger.attrs, false);
});

test("nav hook ignores query strings when matching (#27)", () => {
  const filtered = link("/settings?tab=keys");
  const other = link("/dashboard");
  const el = container([filtered, other], {
    "data-amarra-nav-on": "on",
    "data-amarra-nav-off": "off",
  });
  makeNav({ location: () => location("/settings"), window: window() }).connect(el);
  assert.equal(filtered.classes.has("on"), true);
  assert.equal(other.classes.has("on"), false);
});

test("nav hook re-syncs on popstate and disconnects the listener (#27)", () => {
  let path = "/dashboard";
  const win = window();
  const hook = makeNav({
    location: () => location(path),
    window: win,
  });
  const ledger = link("/dashboard");
  const iam = link("/settings");
  const el = container([ledger, iam], {
    "data-amarra-nav-on": "on",
    "data-amarra-nav-off": "off",
  });
  hook.connect(el);
  assert.equal(ledger.classes.has("on"), true);

  path = "/settings";
  win.listeners.popstate[0]();
  assert.equal(iam.classes.has("on"), true);
  assert.equal(ledger.classes.has("on"), false);

  hook.disconnect(el);
  assert.equal((win.listeners.popstate || []).length, 0);
});

test("nav hook works without off-classes and with relative hrefs (#27)", () => {
  const rel = link("settings");
  const el = container([rel], { "data-amarra-nav-on": "on" });
  makeNav({ location: () => location("/settings"), window: window() }).connect(el);
  assert.equal(rel.classes.has("on"), true);
  assert.equal(rel.attrs["aria-current"], "page");
});
