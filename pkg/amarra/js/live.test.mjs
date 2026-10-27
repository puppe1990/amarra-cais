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
  reconnectDelay,
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

// g live counter Rendered.HTML is the #count element; innerHTML morph throws in Idiomorph (#294).
test("applyLiveMessage morphs a named target as outerHTML", () => {
  const calls = [];
  const target = { id: "count" };
  const root = {
    querySelector(sel) {
      return sel === "#count" ? target : null;
    },
  };
  applyLiveMessage(
    { type: "morph", html: '<span id="count">1</span>', target: "count" },
    root,
    (el, html, _fn, opts) => {
      calls.push([el, html, opts]);
    }
  );
  assert.deepEqual(calls, [[target, '<span id="count">1</span>', { morphStyle: "outerHTML" }]]);
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

// #375: a fixed 1000 ms reconnect synchronizes every client after an outage
// and retries permanent rejections forever. Backoff is bounded, jittered, and
// resets once a connection opens.
test("reconnectDelay grows, caps, and jitters with a deterministic random", () => {
  const fixed = { random: () => 0.5 };
  assert.equal(reconnectDelay(1, fixed), 1000);
  assert.equal(reconnectDelay(2, fixed), 2000);
  assert.equal(reconnectDelay(3, fixed), 4000);
  assert.equal(reconnectDelay(10, fixed), 30000);

  const low = reconnectDelay(2, { random: () => 0 });
  const high = reconnectDelay(2, { random: () => 1 });
  assert.equal(low, 1600);
  assert.equal(high, 2400);
});

function liveFixture(view = "counter") {
  const sockets = [];
  const timers = [];
  const events = [];
  class FakeWebSocket {
    constructor(url) {
      this.url = url;
      this.readyState = 0;
      this.handlers = {};
      this.closed = false;
      sockets.push(this);
    }
    addEventListener(type, fn) {
      (this.handlers[type] ||= []).push(fn);
    }
    send() {}
    close() {
      this.closed = true;
      this.readyState = 3;
    }
    emit(type, ev = {}) {
      for (const fn of this.handlers[type] || []) fn(ev);
    }
    open() {
      this.readyState = 1;
      this.emit("open", {});
    }
    drop(code = 1006) {
      this.readyState = 3;
      this.emit("close", { code, reason: "", wasClean: code === 1000 });
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
  const root = {
    isConnected: true,
    attrs: {},
    getAttribute(name) {
      return name === "amarra-live" ? view : "";
    },
    setAttribute(name, value) {
      this.attrs[name] = value;
    },
    removeAttribute(name) {
      delete this.attrs[name];
    },
    dispatchEvent(ev) {
      events.push(ev);
    },
  };
  roots.push(root);
  start({
    document: doc,
    WebSocket: FakeWebSocket,
    location: { protocol: "http:", host: "localhost:8080" },
    setTimeout: (fn, ms) => {
      timers.push({ fn, ms });
      return timers.length;
    },
    clearTimeout: () => {},
    random: () => 0.5,
  });
  return { sockets, timers, events, root, doc, listeners };
}

test("live reconnects with progressive backoff and resets after open", () => {
  const { sockets, timers, root } = liveFixture();
  assert.equal(sockets.length, 1);
  sockets[0].drop(1006);
  assert.equal(timers.length, 1);
  assert.equal(timers[0].ms, 1000);
  assert.equal(root.attrs["data-amarra-live-state"], "reconnecting");

  timers[0].fn();
  assert.equal(sockets.length, 2);
  sockets[1].open();
  assert.equal(root.attrs["data-amarra-live-state"], "open");

  sockets[1].drop(1006);
  assert.equal(timers.length, 2);
  assert.equal(timers[1].ms, 1000, "a successful open resets the attempt count");

  // A second failure before reopening doubles the wait.
  timers[1].fn();
  sockets[2].drop(1006);
  assert.equal(timers.length, 3);
  assert.equal(timers[2].ms, 2000);
});

test("live treats a policy violation as permanent and stops retrying", () => {
  const { sockets, timers, root, events } = liveFixture();
  sockets[0].open();
  sockets[0].drop(1008);
  assert.equal(timers.length, 0, "CSRF rejection must not schedule another attempt");
  assert.equal(root.attrs["data-amarra-live-state"], "failed");
  const state = events.at(-1);
  assert.equal(state.type, "amarra:live-state");
  assert.equal(state.detail.state, "failed");
  assert.equal(state.detail.permanent, true);
  assert.equal(state.detail.code, 1008);
});

test("live cancels the reconnect when the root leaves the document", () => {
  const { sockets, timers, root } = liveFixture();
  root.isConnected = false;
  sockets[0].drop(1006);
  assert.equal(timers.length, 0, "removed roots must not schedule a reconnect");
  assert.equal(root.attrs["data-amarra-live-state"], undefined);
});

// #375: a stale timer or a Drive morph must not open a second socket for the
// same root while a reconnect is pending.
test("live does not open duplicate sockets for one root", () => {
  const { sockets, timers, listeners } = liveFixture();
  sockets[0].drop(1006);
  for (const fn of listeners["amarra:morphed"] || []) fn();
  assert.equal(sockets.length, 1, "pending reconnect must not create a second socket");
  assert.equal(timers.length, 1, "one pending attempt per root");

  timers[0].fn();
  assert.equal(sockets.length, 2);
});
