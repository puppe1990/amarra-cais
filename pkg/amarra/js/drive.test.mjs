import test from "node:test";
import assert from "node:assert/strict";
import {
  shouldInterceptClick,
  driveHeaders,
  extractMainHTML,
  applyDriveResponse,
  visit,
  start,
} from "./drive.mjs";

test("shouldInterceptClick ignores new tab and download", () => {
  assert.equal(
    shouldInterceptClick({
      href: "/x",
      target: "_blank",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptClick({
      href: "/x",
      target: "",
      download: true,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
});

test("shouldInterceptClick allows same-origin GET link", () => {
  assert.equal(
    shouldInterceptClick({
      href: "http://a/x",
      target: "",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    true
  );
});

test("shouldInterceptClick ignores skip, cross-origin, and empty href", () => {
  assert.equal(
    shouldInterceptClick({
      href: "http://a/x",
      target: "",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
      skip: true,
    }),
    false
  );
  assert.equal(
    shouldInterceptClick({
      href: "http://b/x",
      target: "",
      download: false,
      origin: "http://b",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptClick({
      href: "",
      target: "",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
});

test("driveHeaders sets Amarra-Drive and CSRF", () => {
  const h = driveHeaders("tok");
  assert.equal(h["Amarra-Drive"], "true");
  assert.equal(h["X-CSRF-Token"], "tok");
  assert.equal(h["Accept"], "text/html");
});

test("extractMainHTML pulls #amarra-main inner HTML", () => {
  const html = `<html><body><nav>n</nav><main id="amarra-main"><p>hi</p></main></body></html>`;
  assert.equal(extractMainHTML(html).trim(), "<p>hi</p>");
});

test("applyDriveResponse morphs on 200 and pushState", () => {
  const main = { innerHTML: "old" };
  const pushed = [];
  const result = applyDriveResponse({
    status: 200,
    html: `<main id="amarra-main"><p>new</p></main>`,
    url: "http://a/x",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: {
      pushState(_s, _t, url) {
        pushed.push(url);
      },
    },
  });
  assert.equal(result.action, "morph");
  assert.equal(main.innerHTML.trim(), "<p>new</p>");
  assert.deepEqual(pushed, ["http://a/x"]);
});

test("applyDriveResponse morphs 422 without pushState", () => {
  const main = { innerHTML: "old" };
  const pushed = [];
  applyDriveResponse({
    status: 422,
    html: `<main id="amarra-main"><p>err</p></main>`,
    url: "http://a/x",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: {
      pushState(_s, _t, url) {
        pushed.push(url);
      },
    },
  });
  assert.equal(main.innerHTML.trim(), "<p>err</p>");
  assert.deepEqual(pushed, []);
});

test("applyDriveResponse reloads on 401 and 403", () => {
  let reloads = 0;
  const location = {
    reload() {
      reloads += 1;
    },
  };
  assert.equal(applyDriveResponse({ status: 401, location }).action, "reload");
  assert.equal(applyDriveResponse({ status: 403, location }).action, "reload");
  assert.equal(reloads, 2);
});

test("visit sends drive headers and follows redirects", async () => {
  let init;
  const main = { innerHTML: "old" };
  const fetchFn = async (_url, opts) => {
    init = opts;
    return {
      status: 200,
      url: "http://a/y",
      text: async () => `<main id="amarra-main"><p>ok</p></main>`,
    };
  };
  const pushed = [];
  await visit("http://a/x", {
    fetchFn,
    csrfToken: "tok",
    document: { querySelector: (sel) => (sel === "#amarra-main" ? main : null) },
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: {
      pushState(_s, _t, url) {
        pushed.push(url);
      },
    },
  });
  assert.equal(init.headers["Amarra-Drive"], "true");
  assert.equal(init.headers["X-CSRF-Token"], "tok");
  assert.equal(init.redirect, "follow");
  assert.equal(main.innerHTML.trim(), "<p>ok</p>");
  assert.deepEqual(pushed, ["http://a/y"]);
});

test("start is a no-op without a document", () => {
  assert.equal(start({ document: null }), undefined);
});
