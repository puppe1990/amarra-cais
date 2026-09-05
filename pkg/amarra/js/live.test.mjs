import test from "node:test";
import assert from "node:assert/strict";
import { wsURL, liveRoot, eventName, formPayload, applyLiveMessage } from "./live.mjs";

test("wsURL uses ws and query", () => {
  assert.equal(
    wsURL(
      { protocol: "http:", host: "localhost:8080" },
      { view: "counter", topic: "counter:home" }
    ),
    "ws://localhost:8080/amarra/live?view=counter&topic=counter%3Ahome"
  );
  assert.equal(
    wsURL({ protocol: "https:", host: "ex.com" }, { view: "chat" }),
    "wss://ex.com/amarra/live?view=chat"
  );
});

test("liveRoot finds amarra-live ancestor", () => {
  const root = {
    getAttribute: () => "counter",
    closest(sel) {
      return sel === "[amarra-live]" ? root : null;
    },
  };
  const child = {
    closest(sel) {
      return sel === "[amarra-live]" ? root : null;
    },
  };
  assert.equal(liveRoot(child), root);
  assert.equal(liveRoot(null), null);
});

test("eventName reads live bindings", () => {
  const el = {
    getAttribute(name) {
      if (name === "amarra-click") return "inc";
      if (name === "amarra-submit") return "save";
      return "";
    },
  };
  assert.equal(eventName(el, "click"), "inc");
  assert.equal(eventName(el, "submit"), "save");
  assert.equal(eventName(el, "change"), "");
});

test("formPayload flattens FormData", () => {
  class FD {
    constructor() {
      this._ = [
        ["title", "hi"],
        ["csrf_token", "x"],
      ];
    }
    *[Symbol.iterator]() {
      yield* this._;
    }
  }
  assert.deepEqual(formPayload({}, FD), { title: "hi", csrf_token: "x" });
});

test("applyLiveMessage morphs target id", () => {
  const calls = [];
  const target = { id: "count" };
  const root = {
    querySelector(sel) {
      return sel === "#count" ? target : null;
    },
  };
  applyLiveMessage({ type: "morph", html: "<b>1</b>", target: "count" }, root, (el, html) => {
    calls.push([el, html]);
  });
  assert.deepEqual(calls, [[target, "<b>1</b>"]]);
});
