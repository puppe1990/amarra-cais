import test from "node:test";
import assert from "node:assert/strict";
import { makeSelectSearch } from "./hook_select.mjs";

// Minimal DOM stub for the progressive-enhancement combobox: the hook builds a
// trigger + search panel and keeps the native <select> as the source of truth.
function makeNode(tag) {
  const classes = new Set();
  const node = {
    tagName: String(tag).toUpperCase(),
    children: [],
    parentNode: null,
    attrs: {},
    style: {},
    listeners: {},
    textContent: "",
    value: "",
    hidden: false,
    disabled: false,
    focused: false,
    multiple: false,
    options: [],
    appendChild(child) {
      child.parentNode = this;
      this.children.push(child);
      return child;
    },
    insertBefore(child, ref) {
      child.parentNode = this;
      if (ref == null) {
        this.children.push(child);
        return child;
      }
      const i = this.children.indexOf(ref);
      this.children.splice(i < 0 ? this.children.length : i, 0, child);
      return child;
    },
    removeChild(child) {
      this.children = this.children.filter((c) => c !== child);
      child.parentNode = null;
      return child;
    },
    setAttribute(n, v) {
      this.attrs[n] = String(v);
    },
    getAttribute(n) {
      return Object.hasOwn(this.attrs, n) ? this.attrs[n] : null;
    },
    removeAttribute(n) {
      delete this.attrs[n];
    },
    hasAttribute(n) {
      return Object.hasOwn(this.attrs, n);
    },
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      this.listeners[type] = (this.listeners[type] || []).filter((f) => f !== fn);
    },
    fire(type, ev) {
      (this.listeners[type] || []).forEach((fn) => fn(ev || { preventDefault() {} }));
    },
    dispatchEvent(ev) {
      this.fire(ev?.type, ev);
      return true;
    },
    click() {
      this.fire("click", { preventDefault() {}, stopPropagation() {} });
    },
    focus() {
      this.focused = true;
    },
    querySelectorAll(sel) {
      return descendants(this).filter((n) => matches(n, sel));
    },
    querySelector(sel) {
      return this.querySelectorAll(sel)[0] ?? null;
    },
    classList: {
      add: (c) => classes.add(c),
      remove: (c) => classes.delete(c),
      contains: (c) => classes.has(c),
      toggle: (c, force) => {
        const on = force === undefined ? !classes.has(c) : force;
        on ? classes.add(c) : classes.delete(c);
        return on;
      },
    },
  };
  Object.defineProperty(node, "nextSibling", {
    get() {
      if (!this.parentNode) return null;
      const kids = this.parentNode.children;
      const i = kids.indexOf(this);
      return i >= 0 ? (kids[i + 1] ?? null) : null;
    },
  });
  return node;
}

function descendants(root, out = []) {
  for (const child of root.children) {
    out.push(child);
    descendants(child, out);
  }
  return out;
}

function matches(node, sel) {
  if (sel.startsWith(".")) return node.classList?.contains(sel.slice(1));
  const attr = /^\[([^\]=]+)\]$/.exec(sel);
  if (attr) return node.attrs[attr[1]] != null;
  return node.tagName === sel.toUpperCase();
}

function fakeDoc() {
  const doc = makeNode("body");
  doc.body = doc;
  doc.createElement = (tag) => makeNode(tag);
  doc.documentElement = { lang: "pt" };
  const win = {
    matchMedia: () => ({ matches: false }),
  };
  doc.defaultView = win;
  return { doc, win };
}

function option(value, text) {
  const o = makeNode("option");
  o.value = value;
  o.textContent = text;
  return o;
}

function fakeSelect(
  opts = [
    ["", "Selecione..."],
    ["cpfl", "Grupo CPFL"],
    ["enel", "Enel"],
  ]
) {
  const { doc } = fakeDoc();
  const form = makeNode("form");
  const select = makeNode("select");
  select.ownerDocument = doc;
  select.options = opts.map(([v, t]) => option(v, t));
  select.value = "";
  form.appendChild(select);
  doc.body.appendChild(form);
  return { doc, select, form };
}

const trigger = (select) => select.nextSibling.querySelector(".cais-select-search-trigger");
const label = (select) => trigger(select).querySelector(".cais-select-search-label").textContent;
const panel = (select) => select.nextSibling.querySelector(".cais-select-search-panel");
const input = (select) => select.nextSibling.querySelector(".cais-select-search-input");
const options = (select) => select.nextSibling.querySelectorAll(".cais-select-search-option");

test("select hook hides the native control and injects a trigger + search panel", () => {
  const { select } = fakeSelect();
  makeSelectSearch({ isCoarse: () => false }).connect(select);

  assert.equal(select.classList.contains("cais-select-search-native"), true);
  const wrapper = select.nextSibling;
  assert.equal(wrapper.classList.contains("cais-select-search"), true);
  assert.ok(trigger(select), "trigger injected");
  assert.equal(panel(select).hidden, true, "panel starts closed");
  assert.ok(input(select), "search input injected");
  assert.equal(label(select), "Selecione...", "trigger mirrors the empty option");
});

test("select hook opens the panel and filters options case-insensitively", () => {
  const { select } = fakeSelect();
  makeSelectSearch({ isCoarse: () => false }).connect(select);

  trigger(select).click();
  assert.equal(panel(select).hidden, false);
  assert.equal(input(select).focused, true);

  input(select).value = "ENEL";
  input(select).fire("input");
  const rows = options(select);
  assert.equal(rows[2].classList.contains("is-hidden"), false, "Enel stays visible");
  assert.equal(rows[1].classList.contains("is-hidden"), true, "Grupo CPFL filtered out");
});

test("select hook writes the choice back to the native select and fires change", () => {
  const { select } = fakeSelect();
  let changes = 0;
  select.addEventListener("change", () => changes++);
  makeSelectSearch({ isCoarse: () => false }).connect(select);

  trigger(select).click();
  options(select)[2].click();

  assert.equal(select.value, "enel");
  assert.equal(changes, 1, "change bubbles for reveal/amarra-change hooks");
  assert.equal(label(select), "Enel");
  assert.equal(panel(select).hidden, true, "panel closes after choosing");
});

test("select hook drives with ArrowDown/Enter and closes on Escape", () => {
  const { select } = fakeSelect();
  makeSelectSearch({ isCoarse: () => false }).connect(select);

  trigger(select).click();
  const box = input(select);
  box.fire("keydown", { key: "ArrowDown", preventDefault() {} });
  box.fire("keydown", { key: "ArrowDown", preventDefault() {} });
  box.fire("keydown", { key: "Enter", preventDefault() {} });
  assert.equal(select.value, "enel", "second option after the empty one is Enel");

  trigger(select).click();
  box.fire("keydown", { key: "Escape", preventDefault() {} });
  assert.equal(panel(select).hidden, true);
});

test("select hook opts out with data-amarra-select-search=false", () => {
  const { select } = fakeSelect();
  select.setAttribute("data-amarra-select-search", "false");
  makeSelectSearch({ isCoarse: () => false }).connect(select);
  assert.equal(select.nextSibling, null, "no wrapper injected on opt-out");
  assert.equal(select.classList.contains("cais-select-search-native"), false);
});

test("select hook leaves coarse pointers on the native picker", () => {
  const { select } = fakeSelect();
  makeSelectSearch({ isCoarse: () => true }).connect(select);
  assert.equal(select.nextSibling, null);
});

test("select hook disconnect removes the injected UI and restores the select", () => {
  const { select } = fakeSelect();
  const hook = makeSelectSearch({ isCoarse: () => false });
  hook.connect(select);
  assert.ok(select.nextSibling);

  hook.disconnect(select);
  assert.equal(select.nextSibling, null);
  assert.equal(select.classList.contains("cais-select-search-native"), false);
});

// #327: a dependent select (estado → cidade) fills its options after the hook
// mounted; the listbox must rebuild instead of freezing on the first snapshot.
test("select hook rebuilds rows when the native options change", () => {
  const { select } = fakeSelect();
  let observer = null;
  class MO {
    constructor(cb) {
      this.cb = cb;
      observer = this;
    }
    observe() {}
    disconnect() {
      this.disconnected = true;
    }
    trigger() {
      this.cb();
    }
  }
  const hook = makeSelectSearch({ isCoarse: () => false, MutationObserver: MO });
  hook.connect(select);
  assert.ok(observer, "an observer watches the select");

  select.options = [option("", "Selecione..."), option("sp", "São Paulo"), option("rj", "Rio")];
  observer.trigger();

  const rows = options(select);
  assert.equal(rows.length, 3, "rows follow the new options");
  assert.equal(rows[1].textContent, "São Paulo");

  trigger(select).click();
  input(select).value = "rio";
  input(select).fire("input");
  assert.equal(rows[2].classList.contains("is-hidden"), false);
  assert.equal(rows[1].classList.contains("is-hidden"), true);

  hook.disconnect(select);
  assert.equal(observer.disconnected, true, "observer is disconnected on unmount");
});
