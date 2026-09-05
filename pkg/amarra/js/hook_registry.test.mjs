import test from "node:test";
import assert from "node:assert/strict";
import { register, scan, dispatchLivePush, reset } from "./hook_registry.mjs";

function el(id, name) {
  return {
    id,
    getAttribute(n) {
      return n === "amarra-hook" ? name : null;
    },
    hasAttribute(n) {
      return n === "amarra-hook";
    },
  };
}

function docWith(nodes) {
  return {
    querySelectorAll(sel) {
      return sel === "[amarra-hook]" ? nodes : [];
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
  scan(doc);
  assert.deepEqual(events, ["connect:btn", "updated:btn", "disconnect:btn"]);
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
