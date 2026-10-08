import test from "node:test";
import assert from "node:assert/strict";
import { register, scan, dispatchLivePush, reset } from "./hook_registry.mjs";

function el(id, name, children = []) {
  return {
    id,
    isConnected: true,
    getAttribute(n) {
      return n === "amarra-hook" ? name : null;
    },
    hasAttribute(n) {
      return n === "amarra-hook";
    },
    contains(node) {
      return children.includes(node);
    },
  };
}

function docWith(nodes) {
  return {
    querySelectorAll(sel) {
      return sel === "[amarra-hook]" ? nodes : [];
    },
    contains(node) {
      return nodes.includes(node);
    },
  };
}

test("scan connects matching amarra-hook, updates on rescan, disconnects when gone", () => {
  reset();
  const events = [];
  register("clip", {
    connect(node) {
      events.push("connect:" + node.id);
    },
    updated(node) {
      events.push("updated:" + node.id);
    },
    disconnect(node) {
      events.push("disconnect:" + node.id);
    },
  });
  const btn = el("btn", "clip");
  const nodes = [btn];
  const doc = docWith(nodes);
  scan(doc);
  assert.deepEqual(events, ["connect:btn"]);
  scan(doc);
  assert.deepEqual(events, ["connect:btn", "updated:btn"]);
  nodes.pop();
  btn.isConnected = false; // removed from the DOM, not just from the scan
  scan(doc);
  assert.deepEqual(events, ["connect:btn", "updated:btn", "disconnect:btn"]);
});

// #329: scan on a subtree must only prune orphans of THAT subtree; a mounted
// hook outside it (another select, the theme toggle, …) stays connected.
test("scan on a subtree keeps hooks outside it mounted", () => {
  reset();
  const events = [];
  register("clip", {
    connect(node) {
      events.push("connect:" + node.id);
    },
    updated(node) {
      events.push("updated:" + node.id);
    },
    disconnect(node) {
      events.push("disconnect:" + node.id);
    },
  });
  const uf = el("uf", "clip");
  const city = el("city", "clip");
  scan(docWith([uf, city]));
  assert.deepEqual(events, ["connect:uf", "connect:city"]);

  scan(city);
  assert.deepEqual(events, ["connect:uf", "connect:city", "updated:city"]);
});

// #329: a node that stays connected inside the scanned root but is no longer
// collected (hook attribute dropped in place) is still an orphan of the scan.
test("scan disconnects a node that stops being collected inside the root", () => {
  reset();
  const events = [];
  register("clip", {
    connect(node) {
      events.push("connect:" + node.id);
    },
    disconnect(node) {
      events.push("disconnect:" + node.id);
    },
  });
  const btn = el("btn", "clip");
  const doc = docWith([btn]);
  scan(doc);

  btn.getAttribute = () => null;
  btn.hasAttribute = () => false;
  scan(doc);
  assert.deepEqual(events, ["connect:btn", "disconnect:btn"]);
});

// #329: the old prune ran on every mounted node missing from the scan, so a
// document scan must still disconnect nodes that left the DOM.
test("scan still disconnects a hook removed from the document", () => {
  reset();
  const events = [];
  register("clip", {
    connect(node) {
      events.push("connect:" + node.id);
    },
    updated(node) {
      events.push("updated:" + node.id);
    },
    disconnect(node) {
      events.push("disconnect:" + node.id);
    },
  });
  const uf = el("uf", "clip");
  const city = el("city", "clip");
  scan(docWith([uf, city]));

  city.isConnected = false;
  scan(docWith([uf]));
  assert.deepEqual(events, ["connect:uf", "connect:city", "updated:uf", "disconnect:city"]);
});

test("scan skips unknown hook names", () => {
  reset();
  const doc = docWith([el("x", "missing")]);
  scan(doc);
});

test("dispatchLivePush delivers handleEvent to mounted hooks", () => {
  reset();
  const seen = [];
  register("chart", {
    connect() {},
    handleEvent(name, payload, node) {
      seen.push({ name, payload, id: node.id });
    },
  });
  const chart = el("c1", "chart");
  scan(docWith([chart]));
  dispatchLivePush("highlight", { id: 7 });
  assert.deepEqual(seen, [{ name: "highlight", payload: { id: 7 }, id: "c1" }]);
});
