import test from "node:test";
import assert from "node:assert/strict";
import { frameHeaders, loadFrame, define, visitIntoFrame, observeLazy } from "./frame.mjs";

test("frameHeaders sets Amarra-Frame and CSRF", () => {
  const h = frameHeaders("cart", "tok");
  assert.equal(h["Amarra-Frame"], "cart");
  assert.equal(h["X-CSRF-Token"], "tok");
  assert.equal(h["Accept"], "text/html");
});

test("loadFrame fetches with Amarra-Frame and morphs", async () => {
  let req;
  const el = {
    id: "cart",
    innerHTML: "",
    getAttribute(name) {
      if (name === "src") return "/cart";
      if (name === "id") return "cart";
      return null;
    },
  };
  const fetchFn = async (url, opts) => {
    req = { url, opts };
    return { status: 200, text: async () => "<p>cart</p>" };
  };
  let morphed;
  await loadFrame(el, {
    fetchFn,
    morphFn: (node, html) => {
      morphed = { node, html };
    },
    csrfToken: "tok",
  });
  assert.equal(req.url, "/cart");
  assert.equal(req.opts.headers["Amarra-Frame"], "cart");
  assert.equal(req.opts.headers["X-CSRF-Token"], "tok");
  assert.equal(morphed.node, el);
  assert.equal(morphed.html, "<p>cart</p>");
});

test("loadFrame dispatches amarra:morphed", async () => {
  const events = [];
  const el = {
    getAttribute(name) {
      if (name === "src") return "/x";
      if (name === "id") return "f";
      return null;
    },
  };
  const doc = {
    dispatchEvent(ev) {
      events.push(ev.type);
      return true;
    },
  };
  await loadFrame(el, {
    document: doc,
    fetchFn: async () => ({ status: 200, text: async () => "<p>x</p>" }),
    morphFn() {},
  });
  assert.equal(events.includes("amarra:morphed"), true);
});

test("define returns false without customElements", () => {
  assert.equal(define({ customElements: null, HTMLElement: null }), false);
});

test("visitIntoFrame morphs the named frame instead of Drive", async () => {
  let req;
  const frameEl = {
    id: "cart",
    getAttribute(name) {
      if (name === "id") return "cart";
      if (name === "src") return this._src;
      return null;
    },
    setAttribute(name, value) {
      if (name === "src") this._src = value;
    },
  };
  const doc = {
    getElementById(id) {
      return id === "cart" ? frameEl : null;
    },
  };
  const fetchFn = async (url, opts) => {
    req = { url, opts };
    return { status: 200, text: async () => "<p>cart</p>" };
  };
  let morphed;
  const a = {
    getAttribute(name) {
      return name === "data-amarra-frame" ? "cart" : null;
    },
  };
  const ok = await visitIntoFrame(a, "/cart", {
    document: doc,
    fetchFn,
    morphFn: (node, html) => {
      morphed = { node, html };
    },
    csrfToken: "tok",
  });
  assert.equal(ok, true);
  assert.equal(req.url, "/cart");
  assert.equal(req.opts.headers["Amarra-Frame"], "cart");
  assert.equal(morphed.node, frameEl);
  assert.equal(morphed.html, "<p>cart</p>");
});

test("visitIntoFrame returns false for _top or missing frame", async () => {
  assert.equal(
    await visitIntoFrame(
      {
        getAttribute() {
          return "_top";
        },
      },
      "/x",
      { document: { getElementById: () => null } }
    ),
    false
  );
  assert.equal(
    await visitIntoFrame(
      {
        getAttribute() {
          return "missing";
        },
      },
      "/x",
      { document: { getElementById: () => null } }
    ),
    false
  );
});

test("observeLazy waits for intersection before fetching", async () => {
  let observed;
  let fetchCount = 0;
  const el = {
    getAttribute(name) {
      if (name === "src") return "/lazy";
      if (name === "id") return "panel";
      return null;
    },
  };
  const IO = class {
    constructor(cb) {
      this.cb = cb;
    }
    observe(node) {
      observed = node;
    }
    disconnect() {
      this.disconnected = true;
    }
  };
  observeLazy(el, {
    IntersectionObserver: IO,
    fetchFn: async () => {
      fetchCount += 1;
      return { status: 200, text: async () => "<p>x</p>" };
    },
    morphFn() {},
  });
  assert.equal(observed, el);
  assert.equal(fetchCount, 0);
});
