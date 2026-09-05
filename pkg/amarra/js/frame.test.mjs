import test from "node:test";
import assert from "node:assert/strict";
import { frameHeaders, loadFrame, define } from "./frame.mjs";

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

test("define returns false without customElements", () => {
  assert.equal(define({ customElements: null, HTMLElement: null }), false);
});
