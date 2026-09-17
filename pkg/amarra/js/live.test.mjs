import test from "node:test";
import assert from "node:assert/strict";
import {
  wsURL,
  liveRoot,
  eventName,
  formPayload,
  applyLiveMessage,
  debounceWait,
  setLoading,
  start,
} from "./live.mjs";

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

test("applyLiveMessage applies ops, patch, and pushes", () => {
  const ops = [];
  const pushes = [];
  let patched = "";
  const root = {
    querySelector() {
      return null;
    },
  };
  applyLiveMessage(
    {
      type: "morph",
      html: "",
      ops: [{ kind: "append", target: "log", html: "<li>1</li>" }],
      patch: "/counter?n=1",
      pushes: [{ event: "tick", payload: { n: 1 } }],
    },
    root,
    () => {},
    {
      applyOp(op) {
        ops.push(op);
      },
      dispatchPush(event, payload) {
        pushes.push([event, payload]);
      },
      history: {
        pushState(_s, _t, url) {
          patched = url;
        },
      },
    }
  );
  assert.deepEqual(ops, [{ kind: "append", target: "log", html: "<li>1</li>" }]);
  assert.equal(patched, "/counter?n=1");
  assert.deepEqual(pushes, [["tick", { n: 1 }]]);
});

test("debounceWait reads amarra-debounce ms", () => {
  assert.equal(
    debounceWait({
      getAttribute(n) {
        return n === "amarra-debounce" ? "300" : null;
      },
    }),
    300
  );
  assert.equal(debounceWait({ getAttribute: () => null }), 0);
});

test("setLoading toggles amarra-click-loading and amarra-loading", () => {
  const classes = (initial = []) => {
    const set = new Set(initial);
    return {
      add: (c) => set.add(c),
      remove: (c) => set.delete(c),
      contains: (c) => set.has(c),
    };
  };
  const el = { classList: classes() };
  const root = { classList: classes() };
  setLoading(el, root, true);
  assert.equal(el.classList.contains("amarra-click-loading"), true);
  assert.equal(root.classList.contains("amarra-loading"), true);
  setLoading(el, root, false);
  assert.equal(el.classList.contains("amarra-click-loading"), false);
  assert.equal(root.classList.contains("amarra-loading"), false);
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

// #113: a live page reached through Drive navigation never joined — start only
// scanned [amarra-live] at boot. A morph must connect new roots and drop the
// sockets of replaced ones.
test("live start joins roots added by a Drive morph", () => {
  const sockets = [];
  class FakeWebSocket {
    constructor(url) {
      this.url = url;
      this.closed = false;
      this.readyState = 0;
      sockets.push(this);
    }
    addEventListener() {}
    close() {
      this.closed = true;
    }
  }
  const listeners = {};
  const roots = [];
  const doc = {
    documentElement: { dataset: {} },
    querySelectorAll: () => roots,
    addEventListener: (type, fn) => {
      (listeners[type] ||= []).push(fn);
    },
  };
  const root = (view) => ({
    isConnected: true,
    getAttribute: (name) => (name === "amarra-live" ? view : ""),
  });

  const first = root("chat");
  roots.push(first);
  start({
    document: doc,
    WebSocket: FakeWebSocket,
    location: { protocol: "http:", host: "localhost:8080" },
  });
  assert.equal(sockets.length, 1);
  assert.equal(sockets[0].url, "ws://localhost:8080/amarra/live?view=chat&topic=chat");

  first.isConnected = false;
  const fresh = root("counter");
  roots.length = 0;
  roots.push(fresh);
  for (const fn of listeners["amarra:morphed"] || []) fn();

  assert.equal(sockets.length, 2, "morphed-in root should open a socket");
  assert.equal(sockets[1].url, "ws://localhost:8080/amarra/live?view=counter&topic=counter");
  assert.equal(sockets[0].closed, true, "replaced root socket should close");

  for (const fn of listeners["amarra:morphed"] || []) fn();
  assert.equal(sockets.length, 2, "sync must be idempotent");
});
