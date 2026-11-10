import test from "node:test";
import assert from "node:assert/strict";
import {
  extractTitle,
  extractHTMLAttr,
  extractBodyAttr,
  applyHead,
  showProgress,
  hideProgress,
} from "./drive_head.mjs";

// #249: the layout shell marker sits on <body>, and extractHTMLAttr only reads
// the <html> tag.
test("extractBodyAttr reads a marker off the body tag", () => {
  const html = `<html lang="pt-BR"><body class="min-h-screen" data-amarra-shell="auth"><main id="amarra-main"></main></body></html>`;
  assert.equal(extractBodyAttr(html, "data-amarra-shell"), "auth");
  assert.equal(extractHTMLAttr(html, "data-amarra-shell"), null);
  assert.equal(extractBodyAttr(`<body data-amarra-shell="app">`, "data-amarra-shell"), "app");
  assert.equal(extractBodyAttr(`<html><body></body></html>`, "data-amarra-shell"), null);
});

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
  assert.match(spinner.style.cssText, /#1f8ba3/);
  assert.match(spinner.style.cssText, /rgba\(31,139,163,\.25\)/);
  hideProgress(doc);
  assert.equal(veil.hidden, true);
});

function articleHTML({ slug, title, desc, jsonld, extra = "" }) {
  return `<html lang="pt-BR"><head>
    <title>${title}</title>
    <link rel="canonical" href="https://sintaxe.example/${slug}">
    <link rel="alternate" hreflang="en" href="https://sintaxe.example/en/${slug}">
    <meta name="description" content="${desc}">
    <meta property="og:url" content="https://sintaxe.example/${slug}">
    <meta property="og:title" content="${title}">
    <meta name="twitter:card" content="summary">
    <script type="application/ld+json">${jsonld}</script>
    ${extra}
    <link rel="stylesheet" href="/static/css/styles.css">
    <meta name="csrf-token" content="tok">
  </head><body><main id="amarra-main">${title}</main></body></html>`;
}

function makeHeadDoc() {
  const nodes = [];
  const head = {
    childNodes: nodes,
    appendChild(el) {
      el.parentNode = this;
      nodes.push(el);
      return el;
    },
    querySelectorAll(sel) {
      const tag = String(sel).toUpperCase();
      return nodes.filter((el) => el.tagName === tag);
    },
  };
  return {
    title: "",
    head,
    documentElement: { lang: "en" },
    createElement(tag) {
      const attrs = {};
      const el = {
        tagName: tag.toUpperCase(),
        getAttribute(name) {
          return Object.prototype.hasOwnProperty.call(attrs, name) ? attrs[name] : null;
        },
        setAttribute(name, value) {
          // Mirror Chromium: void-tag slashes must not land here as "/".
          if (!/^[A-Za-z_:][\w:.-]*$/.test(String(name))) {
            throw new Error(
              `InvalidCharacterError: Failed to execute 'setAttribute' on 'Element': '${name}' is not a valid attribute name.`
            );
          }
          attrs[name] = String(value);
        },
        textContent: "",
        parentNode: null,
        remove() {
          if (!this.parentNode) return;
          const list = this.parentNode.childNodes;
          const i = list.indexOf(this);
          if (i >= 0) list.splice(i, 1);
          this.parentNode = null;
        },
      };
      return el;
    },
    querySelector() {
      return null;
    },
  };
}

function headAttr(doc, tag, attr, value) {
  return [...doc.head.querySelectorAll(tag)].find((el) => el.getAttribute(attr) === value) ?? null;
}

test("applyHead copies canonical, og, description and JSON-LD from article A to B", () => {
  const doc = makeHeadDoc();
  applyHead(
    doc,
    articleHTML({
      slug: "a",
      title: "Post A",
      desc: "Desc A",
      jsonld: '{"@id":"a"}',
      extra: `<meta property="og:image" content="https://sintaxe.example/a.png">`,
    })
  );

  applyHead(
    doc,
    articleHTML({
      slug: "b",
      title: "Post B",
      desc: "Desc B",
      jsonld: '{"@id":"b"}',
    })
  );

  assert.equal(doc.title, "Post B");
  assert.equal(
    headAttr(doc, "link", "rel", "canonical").getAttribute("href"),
    "https://sintaxe.example/b"
  );
  assert.equal(
    headAttr(doc, "link", "hreflang", "en").getAttribute("href"),
    "https://sintaxe.example/en/b"
  );
  assert.equal(
    headAttr(doc, "meta", "property", "og:url").getAttribute("content"),
    "https://sintaxe.example/b"
  );
  assert.equal(headAttr(doc, "meta", "property", "og:title").getAttribute("content"), "Post B");
  assert.equal(headAttr(doc, "meta", "name", "description").getAttribute("content"), "Desc B");
  const jsonld = [...doc.head.querySelectorAll("script")].filter(
    (el) => el.getAttribute("type") === "application/ld+json"
  );
  assert.equal(jsonld.length, 1);
  assert.equal(jsonld[0].textContent, '{"@id":"b"}');
  assert.equal(headAttr(doc, "meta", "property", "og:image"), null);
  assert.equal(headAttr(doc, "link", "rel", "stylesheet"), null);
});

test("applyHead copies self-closing canonical and og tags without treating slash as an attribute", () => {
  const doc = makeHeadDoc();
  applyHead(
    doc,
    `<html><head>
      <title>Sign in</title>
      <meta charset="UTF-8" />
      <meta name="description" content="Login" />
      <meta property="og:title" content="Sign in" />
      <link rel="canonical" href="https://example.test/login" />
      <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    </head><body><main id="amarra-main">x</main></body></html>`
  );

  assert.equal(doc.title, "Sign in");
  assert.equal(
    headAttr(doc, "link", "rel", "canonical").getAttribute("href"),
    "https://example.test/login"
  );
  assert.equal(headAttr(doc, "meta", "name", "description").getAttribute("content"), "Login");
  assert.equal(headAttr(doc, "meta", "property", "og:title").getAttribute("content"), "Sign in");
  assert.equal(headAttr(doc, "link", "rel", "canonical").getAttribute("/"), null);
});

test("applyHead does not invent document meta when the response has none", () => {
  const doc = makeHeadDoc();
  applyHead(doc, `<html><head><title>Home</title></head><body></body></html>`);
  assert.equal(doc.title, "Home");
  assert.equal(doc.head.childNodes.length, 0);
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
