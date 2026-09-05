import test from "node:test";
import assert from "node:assert/strict";
import { parseSSE, applyOp, start } from "./stream.mjs";

test("parseSSE reads named event and html data", () => {
  const ops = parseSSE("event: append\ndata: <p>x</p>\n\n");
  assert.deepEqual(ops, [{ kind: "append", html: "<p>x</p>" }]);
});

test("parseSSE joins multiline data and optional JSON target", () => {
  const multi = parseSSE("event: morph\ndata: <div>\ndata: x</div>\n\n");
  assert.deepEqual(multi, [{ kind: "morph", html: "<div>\nx</div>" }]);

  const json = parseSSE(`event: append\ndata: {"target":"list","html":"<li>1</li>"}\n\n`);
  assert.equal(json[0].kind, "append");
  assert.equal(json[0].target, "list");
  assert.equal(json[0].html, "<li>1</li>");
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

test("start is a no-op without a document", () => {
  assert.equal(start({ document: null }), undefined);
});
