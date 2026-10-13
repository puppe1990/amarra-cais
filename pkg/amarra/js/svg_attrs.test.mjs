import test from "node:test";
import assert from "node:assert/strict";
import { restoreSvgAttributeCase } from "./svg_attrs.mjs";

const SVG_NS = "http://www.w3.org/2000/svg";
const HTML_NS = "http://www.w3.org/1999/xhtml";

function fakeElement({ namespaceURI = HTML_NS, attrs = {}, children = [] } = {}) {
  const map = new Map(Object.entries(attrs));
  return {
    namespaceURI,
    children,
    get attributes() {
      return [...map].map(([name, value]) => ({ name, value }));
    },
    getAttribute(name) {
      return map.has(name) ? map.get(name) : null;
    },
    setAttribute(name, value) {
      map.set(name, String(value));
    },
    hasAttribute(name) {
      return map.has(name);
    },
    removeAttribute(name) {
      map.delete(name);
    },
    querySelectorAll() {
      const out = [];
      const walk = (node) => {
        for (const child of node.children ?? []) {
          out.push(child);
          walk(child);
        }
      };
      walk(this);
      return out;
    },
  };
}

// #330: the HTML parser lowercases attribute names (viewBox -> viewbox) while the
// live SVG node keeps the camelCase ones; a morph would then drop viewBox and the
// drawing stops painting.
test("restoreSvgAttributeCase restores the camelCase name the parse lowercased", () => {
  const svg = fakeElement({
    namespaceURI: SVG_NS,
    attrs: { viewbox: "0 0 830 720", width: "100%" },
  });
  const main = fakeElement({ children: [svg] });

  assert.equal(restoreSvgAttributeCase(main), 1);
  assert.equal(svg.getAttribute("viewBox"), "0 0 830 720");
  assert.equal(svg.getAttribute("viewbox"), null);
  assert.equal(svg.getAttribute("width"), "100%");
});

test("restoreSvgAttributeCase reaches nested SVG nodes (defs/marker)", () => {
  const marker = fakeElement({
    namespaceURI: SVG_NS,
    attrs: { viewbox: "0 0 10 10", refx: "9", refy: "5", markerwidth: "6" },
  });
  const defs = fakeElement({ namespaceURI: SVG_NS, children: [marker] });
  const svg = fakeElement({
    namespaceURI: SVG_NS,
    attrs: { viewbox: "0 0 830 720", preserveaspectratio: "xMidYMid meet" },
    children: [defs],
  });

  assert.equal(restoreSvgAttributeCase(svg), 6);
  assert.equal(svg.getAttribute("viewBox"), "0 0 830 720");
  assert.equal(svg.getAttribute("preserveAspectRatio"), "xMidYMid meet");
  assert.equal(marker.getAttribute("viewBox"), "0 0 10 10");
  assert.equal(marker.getAttribute("refX"), "9");
  assert.equal(marker.getAttribute("refY"), "5");
  assert.equal(marker.getAttribute("markerWidth"), "6");
  assert.equal(marker.getAttribute("refx"), null);
});

test("restoreSvgAttributeCase leaves HTML alone and keeps a correct camelCase value", () => {
  const div = fakeElement({ attrs: { viewbox: "not svg", "data-x": "1" } });
  const svg = fakeElement({ namespaceURI: SVG_NS, attrs: { viewBox: "0 0 4 4" } });

  assert.equal(restoreSvgAttributeCase(div), 0);
  assert.equal(div.getAttribute("viewbox"), "not svg");
  assert.equal(restoreSvgAttributeCase(svg), 0);
  assert.equal(svg.getAttribute("viewBox"), "0 0 4 4");
});

test("restoreSvgAttributeCase prefers the freshly morphed value over a stale camelCase one", () => {
  const svg = fakeElement({
    namespaceURI: SVG_NS,
    attrs: { viewBox: "0 0 1 1", viewbox: "0 0 830 720" },
  });

  assert.equal(restoreSvgAttributeCase(svg), 1);
  assert.equal(svg.getAttribute("viewBox"), "0 0 830 720");
});

test("restoreSvgAttributeCase is a no-op without a queryable root", () => {
  assert.equal(restoreSvgAttributeCase(null), 0);
  assert.equal(restoreSvgAttributeCase({}), 0);
});
