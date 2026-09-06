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
  assert.equal(created.length, 1);
  assert.equal(created[0].id, "amarra-progress");
  assert.equal(created[0].hidden, false);
  hideProgress(doc);
  assert.equal(created[0].hidden, true);
});
