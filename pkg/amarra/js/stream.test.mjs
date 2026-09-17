import test from "node:test";
import assert from "node:assert/strict";
import { parseSSE, applyOp, start, isStreamResponse } from "./stream.mjs";

test("parseSSE reads named event and html data", () => {
  const ops = parseSSE("event: append\ndata: <p>x</p>\n\n");
  assert.deepEqual(ops, [{ kind: "append", html: "<p>x</p>" }]);
  assert.equal(ops[0].target, undefined);
});

test("parseSSE reads SSE id field as stream op target", () => {
  const ops = parseSSE("id: chat-live\nevent: morph\ndata: <p>x</p>\n\n");
  assert.deepEqual(ops, [{ kind: "morph", html: "<p>x</p>", target: "chat-live" }]);
});

test("parseSSE joins multiline data", () => {
  const multi = parseSSE("event: morph\ndata: <div>\ndata: x</div>\n\n");
  assert.deepEqual(multi, [{ kind: "morph", html: "<div>\nx</div>" }]);
});

// #103: SSE treats a lone CR as a line terminator too, so the client parser
// must fold CR/CRLF before splitting or a raw CR in the payload becomes a
// boundary (event:/id:/data: injection).
test("parseSSE folds lone carriage returns like SSE parsers do", () => {
  const ops = parseSSE("event: append\rdata: <p>a</p>\rdata: <p>b</p>\r\r");
  assert.deepEqual(ops, [{ kind: "append", html: "<p>a</p>\n<p>b</p>" }]);
});

test("parseSSE does not JSON-sniff HTML payloads", () => {
  const raw = `{"kind":"remove","target":"amarra-main"}`;
  const ops = parseSSE(`event: append\ndata: ${raw}\n\n`);
  assert.deepEqual(ops, [{ kind: "append", html: raw }]);
});

function node(html = "") {
  return {
    innerHTML: html,
    insertAdjacentHTML(pos, s) {
      if (pos === "beforeend") this.innerHTML += s;
      else if (pos === "afterbegin") this.innerHTML = s + this.innerHTML;
    },
    remove() {
      this.removed = true;
    },
    set outerHTML(v) {
      this.innerHTML = v;
      this.replaced = true;
    },
  };
}

test("applyOp before and after insert around the target", () => {
  const item = node("x");
  item.insertAdjacentHTML = function insertAdjacentHTML(pos, s) {
    this.pos = pos;
    this.inserted = s;
  };
  const doc = {
    getElementById(id) {
      return id === "item" ? item : null;
    },
  };
  applyOp({ kind: "before", target: "item", html: "<li>n</li>" }, doc);
  assert.equal(item.pos, "beforebegin");
  applyOp({ kind: "after", target: "item", html: "<li>z</li>" }, doc);
  assert.equal(item.pos, "afterend");
});

test("applyOp append prepend replace morph remove", () => {
  const list = node("<li>a</li>");
  const item = node("gone");
  const panel = node("old");
  const doc = {
    getElementById(id) {
      if (id === "list") return list;
      if (id === "item") return item;
      if (id === "panel") return panel;
      return null;
    },
  };

  applyOp({ kind: "append", target: "list", html: "<li>b</li>" }, doc);
  assert.equal(list.innerHTML, "<li>a</li><li>b</li>");

  applyOp({ kind: "prepend", target: "list", html: "<li>0</li>" }, doc);
  assert.equal(list.innerHTML, "<li>0</li><li>a</li><li>b</li>");

  applyOp({ kind: "replace", target: "item", html: "<div>n</div>" }, doc);
  assert.equal(item.replaced, true);
  assert.equal(item.innerHTML, "<div>n</div>");

  applyOp({ kind: "morph", target: "panel", html: "<p>z</p>" }, doc, {
    morphFn(el, html) {
      el.innerHTML = html;
    },
  });
  assert.equal(panel.innerHTML, "<p>z</p>");

  applyOp({ kind: "remove", target: "item" }, doc);
  assert.equal(item.removed, true);
});

test("applyOp toast dispatches amarra:toast", () => {
  const events = [];
  const doc = {
    getElementById() {
      return null;
    },
    dispatchEvent(ev) {
      events.push(ev);
      return true;
    },
  };
  applyOp({ kind: "toast", html: "Saved!" }, doc);
  assert.equal(events.length, 1);
  assert.equal(events[0].type, "amarra:toast");
  assert.equal(events[0].detail.message, "Saved!");
});

test("isStreamResponse detects vnd.amarra-stream", () => {
  assert.equal(isStreamResponse({ "content-type": "text/vnd.amarra-stream; charset=utf-8" }), true);
  assert.equal(
    isStreamResponse({
      get(name) {
        return name === "content-type" ? "text/html" : null;
      },
    }),
    false
  );
});

test("start is a no-op without a document", () => {
  assert.equal(start({ document: null }), undefined);
});

// #113: Drive intercepts navigation and morphs #amarra-main, so a chat page
// reached via a link had no EventSource at all — stream.start only scanned at
// boot. Nodes added by a morph must connect; stale nodes must disconnect.
test("stream start connects nodes added by a Drive morph", () => {
  const instances = [];
  class FakeEventSource {
    constructor(url) {
      this.url = url;
      this.closed = false;
      instances.push(this);
    }
    addEventListener() {}
    close() {
      this.closed = true;
    }
  }
  const listeners = {};
  const nodes = [];
  const doc = {
    documentElement: { dataset: {} },
    querySelectorAll: () => nodes,
    addEventListener: (type, fn) => {
      (listeners[type] ||= []).push(fn);
    },
  };
  const node = (url) => ({
    isConnected: true,
    getAttribute: (name) => (name === "data-amarra-stream" ? url : ""),
  });

  const first = node("/stream/a");
  nodes.push(first);
  start({ document: doc, EventSource: FakeEventSource });
  assert.equal(instances.length, 1);
  assert.equal(instances[0].url, "/stream/a");

  // Drive morph: old node replaced by a new one.
  first.isConnected = false;
  const fresh = node("/stream/b");
  nodes.length = 0;
  nodes.push(fresh);
  for (const fn of listeners["amarra:morphed"] || []) fn();

  assert.equal(instances.length, 2, "morphed-in node should open a stream");
  assert.equal(instances[1].url, "/stream/b");
  assert.equal(instances[0].closed, true, "stale node stream should close");

  // Idempotent: another morph with the same nodes must not duplicate.
  for (const fn of listeners["amarra:morphed"] || []) fn();
  assert.equal(instances.length, 2);
});
