import test from "node:test";
import assert from "node:assert/strict";
import { extractTitle, applyHead, showProgress, hideProgress } from "./drive_head.mjs";

test("extractTitle reads the document title", () => {
  assert.equal(extractTitle("<html><head><title>Hi</title></head></html>"), "Hi");
  assert.equal(extractTitle("<p>nope</p>"), null);
});

test("applyHead updates title and csrf meta from response HTML", () => {
  const meta = {
    content: "old",
    setAttribute(name, value) {
      if (name === "content") this.content = value;
    },
  };
  const doc = {
    title: "old",
    querySelector(sel) {
      return sel === 'meta[name="csrf-token"]' ? meta : null;
    },
  };
  applyHead(
    doc,
    `<html><head><title>New</title><meta name="csrf-token" content="tok2"></head></html>`
  );
  assert.equal(doc.title, "New");
  assert.equal(meta.content, "tok2");
});

test("applyHead updates html lang from response HTML", () => {
  const doc = {
    documentElement: {
      lang: "en",
    },
    querySelector() {
      return null;
    },
  };

  applyHead(doc, `<html lang="pt-BR"><head><title>Novo</title></head></html>`);

  assert.equal(doc.documentElement.lang, "pt-BR");
});

function veilDoc(iconHref = "/static/icons/app.png") {
  const created = [];
  const img = {
    tagName: "img",
    setAttribute(name, value) {
      this[name] = value;
    },
    style: {},
  };
  const veil = {
    tagName: "div",
    id: "",
    hidden: false,
    style: {},
    setAttribute(name, value) {
      this[name] = value;
    },
    appendChild(el) {
      created.push(el);
    },
    querySelector(sel) {
      return sel === "img" ? img : null;
    },
  };
  const byId = {};
  const doc = {
    getElementById(id) {
      return byId[id] || null;
    },
    querySelector(sel) {
      if (sel === 'link[rel="icon"]') {
        return { getAttribute: () => iconHref };
      }
      return null;
    },
    createElement(tag) {
      const el = {
        tagName: tag,
        id: "",
        hidden: false,
        style: {},
        textContent: "",
        setAttribute(name, value) {
          this[name] = value;
        },
        appendChild(child) {
          created.push(child);
        },
        querySelector(s) {
          return s === "img" ? img : null;
        },
      };
      return el;
    },
    head: {
      appendChild(el) {
        created.push(el);
      },
    },
    body: {
      appendChild(el) {
        created.push(el);
        if (el.id) byId[el.id] = el;
      },
    },
    _created: created,
    _veil: veil,
  };
  return doc;
}

test("showProgress shows a veil with the current favicon", () => {
  const doc = veilDoc("/static/icons/app.png");
  showProgress(doc);
  const veil = doc.getElementById("amarra-veil");
  assert.ok(veil, "veil created");
  assert.equal(veil.hidden, false);
  const img = veil.querySelector("img");
  assert.equal(img.src, "/static/icons/app.png");
  hideProgress(doc);
  assert.equal(veil.hidden, true);
});

test("showProgress falls back to the default icon without link[rel=icon]", () => {
  const doc = veilDoc(null);
  doc.querySelector = () => null;
  showProgress(doc);
  const veil = doc.getElementById("amarra-veil");
  const img = veil.querySelector("img");
  assert.equal(img.src, "/static/icons/icon.png");
});

test("showProgress creates a bar and hideProgress hides it", () => {
  const created = [];
  const doc = {
    getElementById(id) {
      return created.find((el) => el.id === id) || null;
    },
    createElement(tag) {
      return {
        tagName: tag,
        id: "",
        hidden: false,
        style: {},
        setAttribute() {},
      };
    },
    body: {
      appendChild(el) {
        created.push(el);
      },
    },
  };
  showProgress(doc);
  const bar = created.find((el) => el.id === "amarra-progress");
  assert.ok(bar, "progress bar created");
  assert.equal(bar.hidden, false);
  hideProgress(doc);
  assert.equal(bar.hidden, true);
});
