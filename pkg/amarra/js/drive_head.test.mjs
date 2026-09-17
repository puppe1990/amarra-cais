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

function veilDoc() {
  const created = [];
  const byId = {};
  const doc = {
    getElementById(id) {
      return byId[id] || null;
    },
    querySelector() {
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
        querySelector() {
          return null;
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
  };
  return doc;
}

test("showProgress shows a veil with a system-color spinner", () => {
  const doc = veilDoc();
  showProgress(doc);
  const veil = doc.getElementById("amarra-veil");
  assert.ok(veil, "veil created");
  assert.equal(veil.hidden, false);
  assert.equal(veil.querySelector("img"), null, "veil should not use the favicon");
  const spinner = doc._created.find(
    (el) => el.style?.cssText && el.style.cssText.includes("amarra-spin")
  );
  assert.ok(spinner, "spinner created");
  assert.match(spinner.style.cssText, /#c9893a/);
  hideProgress(doc);
  assert.equal(veil.hidden, true);
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
