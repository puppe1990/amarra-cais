import test from "node:test";
import assert from "node:assert/strict";
import { morph } from "./morph.mjs";

test("morph uses injected morphFn", () => {
  const el = { innerHTML: "old" };
  let called = false;
  morph(el, "<p>n</p>", (node, html) => {
    called = true;
    node.innerHTML = html;
  });
  assert.equal(called, true);
  assert.equal(el.innerHTML, "<p>n</p>");
});

test("morph falls back to innerHTML when Idiomorph is absent", () => {
  const el = { innerHTML: "old" };
  morph(el, "<b>x</b>");
  assert.equal(el.innerHTML, "<b>x</b>");
});

// Live named targets send the element itself (`<span id="count">1</span>`).
// innerHTML morph inserts that span into #count and Idiomorph throws
// "The new child element contains the parent" (#294).
test("morph outerHTML fallback replaces the element", () => {
  const parent = { innerHTML: "" };
  const el = {
    innerHTML: "0",
    set outerHTML(v) {
      parent.innerHTML = v;
    },
  };
  morph(el, '<span id="count">1</span>', undefined, { morphStyle: "outerHTML" });
  assert.equal(parent.innerHTML, '<span id="count">1</span>');
});

// #330: a parser that hands Idiomorph a lowercased `viewbox` makes the morph drop
// the camelCase attribute of the live SVG; the wrapper normalizes it back.
test("morph restores camelCase SVG attributes after the morph", () => {
  const SVG_NS = "http://www.w3.org/2000/svg";
  const attrs = new Map([["viewbox", "0 0 830 720"]]);
  const svg = {
    namespaceURI: SVG_NS,
    children: [],
    get attributes() {
      return [...attrs].map(([name, value]) => ({ name, value }));
    },
    getAttribute: (name) => (attrs.has(name) ? attrs.get(name) : null),
    setAttribute(name, value) {
      attrs.set(name, String(value));
    },
    removeAttribute(name) {
      attrs.delete(name);
    },
    querySelectorAll: () => [],
  };
  const main = {
    children: [svg],
    querySelectorAll: () => [svg],
  };

  morph(main, '<svg viewbox="0 0 830 720"></svg>', (el) => {
    el.morphed = true;
  });

  assert.equal(main.morphed, true);
  assert.equal(svg.getAttribute("viewBox"), "0 0 830 720");
  assert.equal(svg.getAttribute("viewbox"), null);
});

test("morph normalizes the parent when an outerHTML morph detached the target", () => {
  const SVG_NS = "http://www.w3.org/2000/svg";
  const attrs = new Map([["viewbox", "0 0 10 10"]]);
  const replaced = {
    namespaceURI: SVG_NS,
    children: [],
    get attributes() {
      return [...attrs].map(([name, value]) => ({ name, value }));
    },
    getAttribute: (name) => (attrs.has(name) ? attrs.get(name) : null),
    setAttribute(name, value) {
      attrs.set(name, String(value));
    },
    removeAttribute(name) {
      attrs.delete(name);
    },
    querySelectorAll: () => [],
  };
  const parent = {
    children: [replaced],
    querySelectorAll: () => [replaced],
  };
  const el = { isConnected: false, parentNode: parent, innerHTML: "0" };

  morph(el, "<svg></svg>", () => {});

  assert.equal(replaced.getAttribute("viewBox"), "0 0 10 10");
});
