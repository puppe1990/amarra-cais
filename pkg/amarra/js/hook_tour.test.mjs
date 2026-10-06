import test from "node:test";
import assert from "node:assert/strict";
import { makeTour } from "./hook_tour.mjs";

// Minimal DOM stub: the tour hook builds its UI with createElement, so the test
// models nodes, attributes, listeners and layout rects without jsdom.
function makeNode(tag) {
  const node = {
    tagName: String(tag).toUpperCase(),
    children: [],
    parentNode: null,
    attrs: {},
    style: {},
    listeners: {},
    textContent: "",
    disabled: false,
    focused: false,
    rect: { top: 0, left: 0, width: 100, height: 40, right: 100, bottom: 40 },
    appendChild(child) {
      child.parentNode = this;
      this.children.push(child);
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
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      this.listeners[type] = (this.listeners[type] || []).filter((f) => f !== fn);
    },
    fire(type, ev) {
      (this.listeners[type] || []).forEach((fn) => fn(ev || { preventDefault() {} }));
    },
    click() {
      this.fire("click", { preventDefault() {} });
    },
    focus() {
      this.focused = true;
    },
    getBoundingClientRect() {
      return this.rect;
    },
    scrollIntoView() {
      this.scrolled = true;
    },
    querySelectorAll(sel) {
      const want = parseSel(sel);
      return descendants(this).filter((n) => n.attrs[want] != null);
    },
    querySelector(sel) {
      return this.querySelectorAll(sel)[0] ?? null;
    },
  };
  return node;
}

function descendants(root, out = []) {
  for (const child of root.children) {
    out.push(child);
    descendants(child, out);
  }
  return out;
}

function parseSel(sel) {
  const m = /^\[(.+)\]$/.exec(sel);
  return m ? m[1] : sel;
}

function channel() {
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

function fakeDoc() {
  const doc = makeNode("body");
  doc.body = doc;
  doc.createElement = (tag) => makeNode(tag);
  const win = channel();
  win.innerWidth = 1024;
  win.innerHeight = 768;
  doc.defaultView = win;
  return { doc, win };
}

function step(opts = {}) {
  const node = makeNode("section");
  if (opts.title != null) node.setAttribute("data-amarra-tour-title", opts.title);
  if (opts.text != null) node.setAttribute("data-amarra-tour-text", opts.text);
  node.rect = {
    top: opts.top ?? 0,
    left: opts.left ?? 0,
    width: opts.width ?? 100,
    height: opts.height ?? 40,
    right: (opts.left ?? 0) + (opts.width ?? 100),
    bottom: (opts.top ?? 0) + (opts.height ?? 40),
  };
  return node;
}

function container(steps, starts, doc) {
  return {
    ownerDocument: doc,
    steps,
    starts,
    querySelectorAll(sel) {
      if (sel === "[data-amarra-tour-start]") return this.starts;
      if (sel === "[data-amarra-tour-step]") return this.steps;
      return [];
    },
  };
}

function fakeTour() {
  const { doc, win } = fakeDoc();
  const s1 = step({
    title: "Resumo",
    text: "Visão geral",
    top: 100,
    left: 40,
    width: 200,
    height: 60,
  });
  const s2 = step({
    title: "Filtros",
    text: "Refine a lista",
    top: 300,
    left: 20,
    width: 120,
    height: 40,
  });
  const start = makeNode("button");
  const el = container([s1, s2], [start], doc);
  const hook = makeTour({ requestAnimationFrame: (fn) => fn() });
  hook.connect(el);
  return { doc, win, el, hook, s1, s2, start };
}

const ui = (doc, attr) => doc.body.querySelector(`[${attr}]`);

test("tour hook builds an overlay and spotlights the first step on start", () => {
  const { doc, start, s1 } = fakeTour();
  assert.equal(doc.body.children.length, 0);

  start.click();

  assert.equal(doc.body.children.length, 3, "block + hole + tip are appended to <body>");
  assert.equal(s1.scrolled, true, "the active step is scrolled into view");

  const hole = ui(doc, "data-amarra-tour-hole");
  assert.equal(hole.style.top, "94px");
  assert.equal(hole.style.left, "34px");
  assert.equal(hole.style.width, "212px");
  assert.equal(hole.style.height, "72px");

  assert.equal(ui(doc, "data-amarra-tour-heading").textContent, "Resumo");
  assert.equal(ui(doc, "data-amarra-tour-body").textContent, "Visão geral");
  assert.equal(ui(doc, "data-amarra-tour-count").textContent, "1 / 2");
  assert.equal(ui(doc, "data-amarra-tour-prev").disabled, true);
  assert.equal(ui(doc, "data-amarra-tour-next").textContent, "Próximo →");
});

test("tour hook advances with next, completes on the last step, and returns focus", () => {
  const { doc, start } = fakeTour();
  start.click();

  ui(doc, "data-amarra-tour-next").click();
  assert.equal(ui(doc, "data-amarra-tour-heading").textContent, "Filtros");
  assert.equal(ui(doc, "data-amarra-tour-count").textContent, "2 / 2");
  assert.equal(ui(doc, "data-amarra-tour-prev").disabled, false);
  assert.equal(ui(doc, "data-amarra-tour-next").textContent, "Concluir");

  ui(doc, "data-amarra-tour-next").click();
  assert.equal(doc.body.children.length, 0, "the tour UI is removed on completion");
  assert.equal(start.focused, true, "focus returns to the start button");
});

test("tour hook ignores steps that are hidden or have no layout box", () => {
  const { doc, el, start } = fakeTour();
  el.steps = [
    step({ title: "Invisível", top: 10, width: 0, height: 0 }),
    step({ title: "Visível", top: 50 }),
  ];
  start.click();
  assert.equal(ui(doc, "data-amarra-tour-count").textContent, "1 / 1");
  assert.equal(ui(doc, "data-amarra-tour-heading").textContent, "Visível");
});

test("tour hook drives with ArrowRight, ArrowLeft and Escape", () => {
  const { doc, start } = fakeTour();
  start.click();

  doc.fire("keydown", { key: "ArrowRight", preventDefault() {} });
  assert.equal(ui(doc, "data-amarra-tour-count").textContent, "2 / 2");

  doc.fire("keydown", { key: "ArrowLeft", preventDefault() {} });
  assert.equal(ui(doc, "data-amarra-tour-count").textContent, "1 / 2");

  doc.fire("keydown", { key: "Escape", preventDefault() {} });
  assert.equal(doc.body.children.length, 0);
});

test("tour hook repositions on scroll via rAF and unbinds everything on disconnect", () => {
  const { doc, win, el, s1, start, hook } = fakeTour();
  start.click();
  const before = ui(doc, "data-amarra-tour-hole").style.top;

  s1.rect = { top: 200, left: 0, width: 100, height: 40, right: 100, bottom: 240 };
  win.fire("scroll", {});
  assert.notEqual(ui(doc, "data-amarra-tour-hole").style.top, before);

  hook.disconnect(el);
  assert.equal(doc.body.children.length, 0);
  assert.equal(win.listeners.scroll.length, 0);
  start.click();
  assert.equal(doc.body.children.length, 0, "disconnected start buttons must not reopen the tour");
});

test("tour hook updated rebinds the start button swapped by a morph", () => {
  const { doc, el, hook } = fakeTour();
  const stale = el.starts[0];

  const fresh = makeNode("button");
  el.starts = [fresh];
  hook.updated(el);
  assert.equal(stale.listeners.click.length, 0, "stale start button is unbound");

  fresh.click();
  assert.equal(doc.body.children.length, 3, "fresh start button opens the tour");
});

test("tour hook does nothing when no step is visible", () => {
  const { doc, el, start } = fakeTour();
  el.steps = [step({ width: 0, height: 0 })];
  start.click();
  assert.equal(doc.body.children.length, 0);
  assert.equal(start.focused, false);
});
