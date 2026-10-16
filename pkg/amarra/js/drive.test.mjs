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

test("applyDriveResponse updates title from the response", () => {
  const main = { innerHTML: "old" };
  const doc = { title: "old", querySelector: () => null };
  applyDriveResponse({
    status: 200,
    html: `<html><head><title>Items</title></head><body><main id="amarra-main"><p>n</p></main></body></html>`,
    url: "http://a/x",
    main,
    document: doc,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  assert.equal(doc.title, "Items");
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

test("applyDriveResponse does full visit on layout mismatch", () => {
  const assigned = [];
  const main = { tagName: "MAIN", innerHTML: "old" };
  const result = applyDriveResponse({
    status: 200,
    html: `<html data-amarra-layout="app"><body><main id="amarra-main"><p>new</p></main></body></html>`,
    url: "http://a/dashboard",
    main,
    document: { documentElement: { dataset: { amarraLayout: "public" } } },
    morphFn() {
      throw new Error("layout mismatch should not morph");
    },
    location: {
      assign(url) {
        assigned.push(url);
      },
    },
    history: {
      pushState() {
        throw new Error("layout mismatch should not push history");
      },
    },
  });

  assert.equal(result.action, "assign");
  assert.deepEqual(assigned, ["http://a/dashboard"]);
  assert.equal(main.innerHTML, "old");
});

// #249: POST /login swaps the layout shell in the same request (anonymous →
// signed in). Morphing #amarra-main then leaves the authenticated page inside
// the anonymous shell until a manual reload, so a shell marker change must be a
// full navigation.
test("applyDriveResponse does full visit when the shell marker changes", () => {
  const assigned = [];
  const main = { tagName: "MAIN", innerHTML: "old" };
  const result = applyDriveResponse({
    status: 200,
    html: `<html data-amarra-layout="app"><body data-amarra-shell="app"><main id="amarra-main"><p>new</p></main></body></html>`,
    url: "http://a/dashboard",
    main,
    document: {
      documentElement: { dataset: { amarraLayout: "app" } },
      body: { dataset: { amarraShell: "auth" } },
    },
    morphFn() {
      throw new Error("shell mismatch should not morph");
    },
    location: {
      assign(url) {
        assigned.push(url);
      },
    },
  });

  assert.equal(result.action, "assign");
  assert.deepEqual(assigned, ["http://a/dashboard"]);
  assert.equal(main.innerHTML, "old");
});

test("applyDriveResponse morphs when the shell marker matches", () => {
  const main = { innerHTML: "old" };
  const result = applyDriveResponse({
    status: 200,
    html: `<html data-amarra-layout="app"><body data-amarra-shell="app"><main id="amarra-main"><p>new</p></main></body></html>`,
    url: "http://a/items",
    main,
    document: {
      documentElement: { dataset: { amarraLayout: "app" } },
      body: { dataset: { amarraShell: "app" } },
    },
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  assert.equal(result.action, "morph");
  assert.equal(main.innerHTML.trim(), "<p>new</p>");
});

test("applyDriveResponse morphs when the page carries no shell marker", () => {
  const main = { innerHTML: "old" };
  const result = applyDriveResponse({
    status: 200,
    html: `<main id="amarra-main"><p>new</p></main>`,
    url: "http://a/x",
    main,
    document: { documentElement: { dataset: {} }, body: { dataset: {} } },
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  assert.equal(result.action, "morph");
});

// #249: the reported flow — the document still shows the anonymous shell when
// the POST /login response (after the 303) carries the signed-in one.
test("visit reloads after a login POST that swaps the shell", async () => {
  const assigned = [];
  const main = { tagName: "MAIN", innerHTML: "anon" };
  const doc = {
    querySelector: (sel) => (sel === "#amarra-main" ? main : null),
    getElementById: () => null,
    body: { dataset: { amarraShell: "auth" } },
    documentElement: { dataset: { amarraLayout: "app" } },
    dispatchEvent() {},
  };
  await visit("http://a/login", {
    method: "POST",
    body: "email=demo%40example.com&password=password",
    fetchFn: async () => ({
      status: 200,
      url: "http://a/dashboard",
      headers: { get: () => null },
      text: async () =>
        `<html data-amarra-layout="app"><body data-amarra-shell="app"><main id="amarra-main"><p>signed in</p></main></body></html>`,
    }),
    document: doc,
    location: { href: "http://a/login", assign: (url) => assigned.push(url) },
    history: { pushState() {}, replaceState() {} },
    morphFn() {
      throw new Error("a shell swap must not morph the anonymous shell");
    },
  });

  assert.deepEqual(assigned, ["http://a/dashboard"]);
  assert.equal(main.innerHTML, "anon");
});

test("applyDriveResponse does full visit when #amarra-main tag changes", () => {
  const assigned = [];
  const main = { tagName: "MAIN", innerHTML: "old" };
  const result = applyDriveResponse({
    status: 200,
    html: `<html data-amarra-layout="app"><body><div id="amarra-main"><p>new</p></div></body></html>`,
    url: "http://a/dashboard",
    main,
    document: { documentElement: { dataset: { amarraLayout: "app" } } },
    morphFn() {
      throw new Error("tag mismatch should not morph");
    },
    location: {
      assign(url) {
        assigned.push(url);
      },
    },
  });

  assert.equal(result.action, "assign");
  assert.deepEqual(assigned, ["http://a/dashboard"]);
  assert.equal(main.innerHTML, "old");
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

// #332: #amarra-main lives in pages, not in files. A PDF/DXF/SVG response was
// read as text, failed extractMainHTML and ended as "ignore" — with the click
// already preventDefault'ed, the file never downloaded.
test("visit hands a non-HTML response back to the browser", async () => {
  const assigned = [];
  const main = { innerHTML: "old" };
  const warnings = [];
  const originalWarn = console.warn;
  console.warn = (msg) => warnings.push(String(msg));
  let textReads = 0;
  try {
    const result = await visit("http://a/projetos/1/pdf/memorial", {
      fetchFn: async () => ({
        status: 200,
        url: "http://a/projetos/1/pdf/memorial",
        headers: { get: (n) => (n === "content-type" ? "application/pdf" : null) },
        text: async () => {
          textReads += 1;
          return "%PDF-1.4 binary";
        },
      }),
      document: { querySelector: (sel) => (sel === "#amarra-main" ? main : null) },
      location: { href: "http://a/projetos/1", assign: (url) => assigned.push(url) },
      morphFn() {
        throw new Error("a file response must not morph");
      },
      history: { pushState() {}, replaceState() {} },
    });
    assert.equal(result.action, "assign");
  } finally {
    console.warn = originalWarn;
  }
  assert.deepEqual(assigned, ["http://a/projetos/1/pdf/memorial"]);
  assert.equal(main.innerHTML, "old");
  assert.equal(textReads, 0);
  assert.deepEqual(warnings, []);
});

test("visit assigns when the response is an attachment even if typed as HTML", async () => {
  const assigned = [];
  await visit("http://a/projetos/1/export", {
    fetchFn: async () => ({
      status: 200,
      url: "http://a/projetos/1/export",
      headers: {
        get: (n) =>
          n === "content-type"
            ? "text/html"
            : n === "content-disposition"
              ? 'attachment; filename="dossie.html"'
              : null,
      },
      text: async () => `<main id="amarra-main"><p>nope</p></main>`,
    }),
    document: { querySelector: () => null },
    location: { href: "http://a/projetos/1", assign: (url) => assigned.push(url) },
    morphFn() {
      throw new Error("an attachment must not morph");
    },
    history: { pushState() {}, replaceState() {} },
  });
  assert.deepEqual(assigned, ["http://a/projetos/1/export"]);
});

test("visit still morphs a text/html response with #amarra-main", async () => {
  const main = { innerHTML: "old" };
  const result = await visit("http://a/x", {
    fetchFn: async () => ({
      status: 200,
      url: "http://a/x",
      headers: { get: (n) => (n === "content-type" ? "text/html; charset=utf-8" : null) },
      text: async () => `<main id="amarra-main"><p>new</p></main>`,
    }),
    document: { querySelector: (sel) => (sel === "#amarra-main" ? main : null) },
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {}, replaceState() {} },
  });
  assert.equal(result.action, "morph");
  assert.equal(main.innerHTML.trim(), "<p>new</p>");
});

// #341: fetch never sends the fragment and `Response.url` comes back without it,
// so a Drive visit from `/resumos/x#planos` used to pushState the bare route and
// jump to the top — the native cold load scrolls to the id, the morph must too.
test("applyDriveResponse scrolls to the fragment target instead of the top", () => {
  const main = { innerHTML: "old" };
  const scrolled = [];
  const toTop = [];
  const target = {
    scrollIntoView() {
      scrolled.push("planos");
    },
  };
  const result = applyDriveResponse({
    status: 200,
    html: `<main id="amarra-main"><h2 id="planos">planos</h2></main>`,
    url: "http://a/resumos/x#planos",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    document: {
      getElementById: (id) => (id === "planos" ? target : null),
      dispatchEvent() {},
    },
    history: { pushState() {} },
    location: { href: "http://a/" },
    window: {
      scrollTo(x, y) {
        toTop.push([x, y]);
      },
    },
  });

  assert.equal(result.action, "morph");
  assert.deepEqual(scrolled, ["planos"]);
  assert.deepEqual(toTop, []);
});

test("applyDriveResponse falls back to the top when the fragment id is missing", () => {
  const main = { innerHTML: "old" };
  const toTop = [];
  applyDriveResponse({
    status: 200,
    html: `<main id="amarra-main"><p>n</p></main>`,
    url: "http://a/resumos/x#sumiu",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    document: { getElementById: () => null, dispatchEvent() {} },
    history: { pushState() {} },
    location: { href: "http://a/" },
    window: {
      scrollTo(x, y) {
        toTop.push([x, y]);
      },
    },
  });

  assert.deepEqual(toTop, [[0, 0]]);
});

test("applyDriveResponse decodes the fragment before looking the id up", () => {
  const main = { innerHTML: "old" };
  const ids = [];
  applyDriveResponse({
    status: 200,
    html: `<main id="amarra-main"><h2 id="05-terceiros">t</h2></main>`,
    url: "http://a/x#05-terceiros",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    document: {
      getElementById(id) {
        ids.push(id);
        return null;
      },
      dispatchEvent() {},
    },
    history: { pushState() {} },
    location: { href: "http://a/" },
    window: { scrollTo() {} },
  });

  assert.deepEqual(ids, ["05-terceiros"]);
});

test("visit keeps the requested fragment on the pushed URL", async () => {
  const main = { innerHTML: "old" };
  const pushed = [];
  await visit("http://a/resumos/x#planos", {
    fetchFn: async () => ({
      status: 200,
      url: "http://a/resumos/x",
      headers: { get: () => null },
      text: async () => `<main id="amarra-main"><h2 id="planos">planos</h2></main>`,
    }),
    document: {
      querySelector: (sel) => (sel === "#amarra-main" ? main : null),
      getElementById: (id) => (id === "planos" ? { scrollIntoView() {} } : null),
      dispatchEvent() {},
    },
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: {
      pushState(_s, _t, url) {
        pushed.push(url);
      },
      replaceState() {},
    },
    location: { href: "http://a/" },
  });

  assert.deepEqual(pushed, ["http://a/resumos/x#planos"]);
});

test("start is a no-op without a document", () => {
  assert.equal(start({ document: null }), undefined);
});

test("extractMainHTML uses matching depth not lastIndexOf", () => {
  const html = `<main id="amarra-main"><section><main class="inner">x</main></section><p>y</p></main><main>other</main>`;
  assert.equal(extractMainHTML(html), `<section><main class="inner">x</main></section><p>y</p>`);
});

test("extractMainHTML returns null when marker is missing", () => {
  assert.equal(extractMainHTML("<html><body><p>foreign</p></body></html>"), null);
});

test("applyDriveResponse does not morph when #amarra-main is missing", () => {
  const main = { innerHTML: "old" };
  let pushed = false;
  const result = applyDriveResponse({
    status: 200,
    html: "<p>foreign</p>",
    url: "http://a/x",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: {
      pushState() {
        pushed = true;
      },
    },
  });
  assert.equal(result.action, "ignore");
  assert.equal(main.innerHTML, "old");
  assert.equal(pushed, false);
});

test("applyDriveResponse warns when #amarra-main is unbalanced", () => {
  const main = { innerHTML: "old" };
  const warnings = [];
  const result = applyDriveResponse({
    status: 200,
    html: `<div id="amarra-main"><div><p>unbalanced</p></div>`,
    url: "http://a/dashboard",
    main,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    warn: (msg) => warnings.push(msg),
  });
  assert.equal(result.action, "ignore");
  assert.equal(main.innerHTML, "old");
  assert.equal(warnings.length, 1);
  assert.match(warnings[0], /dashboard/);
});

test("visit applies amarra-stream HTTP bodies instead of morphing main", async () => {
  const list = {
    innerHTML: "<li>a</li>",
    insertAdjacentHTML(pos, s) {
      if (pos === "beforeend") this.innerHTML += s;
    },
  };
  const doc = {
    querySelector: () => ({ innerHTML: "old" }),
    getElementById: (id) => (id === "list" ? list : null),
    dispatchEvent() {
      return true;
    },
  };
  await visit("http://a/items", {
    fetchFn: async () => ({
      status: 200,
      url: "http://a/items",
      headers: { get: (n) => (n === "content-type" ? "text/vnd.amarra-stream" : null) },
      text: async () => "event: append\nid: list\ndata: <li>b</li>\n\n",
    }),
    document: doc,
    morphFn() {
      throw new Error("should not morph");
    },
    history: { pushState() {}, replaceState() {} },
  });
  assert.equal(list.innerHTML, "<li>a</li><li>b</li>");
});

test("visit emits amarra:drive-error on 404 and 500", async () => {
  const events = [];
  const main = { innerHTML: "old" };
  const doc = {
    querySelector: (sel) => (sel === "#amarra-main" ? main : null),
    dispatchEvent(ev) {
      events.push(ev);
      return true;
    },
  };
  const fetchStatus = (status) => async () => ({
    status,
    url: "http://a/x",
    text: async () => "nope",
  });
  await visit("http://a/x", {
    fetchFn: fetchStatus(500),
    document: doc,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  await visit("http://a/x", {
    fetchFn: fetchStatus(404),
    document: doc,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  assert.equal(main.innerHTML, "old");
  const errors = events.filter((ev) => ev.type === "amarra:drive-error");
  assert.equal(errors.length, 2);
  for (const ev of errors) {
    assert.match(String(ev.detail?.message ?? ""), /Não foi possível carregar a página/);
  }
});

// #360: a 404 HTML page with #amarra-main used to emit drive-error and leave
// the previous screen up after preventDefault, so broken links looked like
// no-ops. Morph the error page and keep the URL in sync.
test("visit morphs a 404 page with #amarra-main and pushes the URL", async () => {
  const main = { innerHTML: "old" };
  const pushed = [];
  const events = [];
  const result = await visit("http://a/missing", {
    fetchFn: async () => ({
      status: 404,
      url: "http://a/missing",
      headers: { get: () => null },
      text: async () => `<main id="amarra-main">Página não encontrada</main>`,
    }),
    document: {
      querySelector: (sel) => (sel === "#amarra-main" ? main : null),
      dispatchEvent(ev) {
        events.push(ev.type);
        return true;
      },
    },
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
  assert.match(main.innerHTML, /Página não encontrada/);
  assert.deepEqual(pushed, ["http://a/missing"]);
  assert.equal(
    events.includes("amarra:drive-error"),
    false,
    "a rendered 404 is a navigation, not a silent error"
  );
});

test("visit morphs a 500 page with #amarra-main without pushing or retrying", async () => {
  const main = { innerHTML: "old" };
  const pushed = [];
  let fetches = 0;
  const result = await visit("http://a/broken", {
    method: "POST",
    body: "title=x",
    fetchFn: async () => {
      fetches += 1;
      return {
        status: 500,
        url: "http://a/broken",
        headers: { get: () => null },
        text: async () => `<main id="amarra-main">Erro interno</main>`,
      };
    },
    document: {
      querySelector: (sel) => (sel === "#amarra-main" ? main : null),
      dispatchEvent() {
        return true;
      },
    },
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
  assert.match(main.innerHTML, /Erro interno/);
  assert.deepEqual(pushed, []);
  assert.equal(fetches, 1);
});

test("visit reports a network failure once without retrying", async () => {
  const events = [];
  const main = { innerHTML: "old" };
  let fetches = 0;
  const result = await visit("http://a/x", {
    fetchFn: async () => {
      fetches += 1;
      throw new Error("offline");
    },
    document: {
      querySelector: (sel) => (sel === "#amarra-main" ? main : null),
      dispatchEvent(ev) {
        events.push(ev);
        return true;
      },
    },
    morphFn() {
      throw new Error("network failure must not morph");
    },
    history: {
      pushState() {
        throw new Error("network failure must not push");
      },
    },
  });
  assert.equal(result.action, "error");
  assert.equal(main.innerHTML, "old");
  assert.equal(fetches, 1);
  const err = events.find((ev) => ev.type === "amarra:drive-error");
  assert.ok(err);
  assert.match(String(err.detail?.message ?? ""), /Não foi possível carregar a página/);
});

// #125: two fast clicks left two fetches in flight; the older response could
// arrive last and morph #amarra-main + pushState its own URL (content A under
// URL B). Responses that are no longer the newest visit are dropped.
test("visit drops a superseded response", async () => {
  const pushed = [];
  const morphed = [];
  const main = { innerHTML: "old" };
  const doc = {
    querySelector: () => main,
    getElementById: () => null,
    createElement: () => ({ style: {}, setAttribute() {} }),
    body: { appendChild() {} },
    documentElement: { dataset: {} },
    dispatchEvent() {},
    title: "",
  };
  let resolveA;
  let resolveB;
  const fetchFn = (url) =>
    new Promise((resolve) => {
      const make = (html) => ({
        status: 200,
        url,
        headers: { get: () => null },
        text: async () => html,
      });
      if (url === "/a") {
        resolveA = () => resolve(make(`<main id="amarra-main"><p>A</p></main>`));
      } else {
        resolveB = () => resolve(make(`<main id="amarra-main"><p>B</p></main>`));
      }
    });
  const opts = {
    fetchFn,
    document: doc,
    morphFn: (_el, html) => morphed.push(html),
    history: {
      pushState(_s, _t, url) {
        pushed.push(url);
      },
      replaceState() {},
    },
    location: { href: "http://a/" },
  };

  const first = visit("/a", opts);
  const second = visit("/b", opts);
  resolveB();
  await second;
  resolveA();
  const result = await first;

  assert.equal(result.action, "superseded");
  assert.deepEqual(pushed, ["/b"]);
  assert.equal(morphed.length, 1);
  assert.match(morphed[0], /<p>B<\/p>/);
});

for (const stream of [false, true]) {
  test(`visit drops superseded delayed ${stream ? "stream" : "HTML"} body`, async () => {
    const morphed = [];
    const pushed = [];
    const replacements = [];
    const main = {};
    const target = {
      remove() {
        throw new Error("stale stream applied");
      },
    };
    const doc = {
      querySelector: () => main,
      getElementById: (id) => (id === "victim" ? target : null),
      createElement: () => ({ style: {}, setAttribute() {} }),
      body: { appendChild() {} },
      documentElement: { dataset: {} },
      dispatchEvent() {},
    };
    let resolveBody;
    let bodyStarted;
    const started = new Promise((resolve) => {
      bodyStarted = resolve;
    });
    const opts = {
      document: doc,
      morphFn: (_el, html) => morphed.push(html),
      location: { href: "http://a/" },
      history: {
        pushState: (_s, _t, url) => pushed.push(url),
        replaceState: () => replacements.push("replace"),
      },
      fetchFn: async (url) => ({
        status: 200,
        url,
        headers: {
          get: (name) =>
            name === "content-type" && stream && url === "/old"
              ? "text/vnd.amarra-stream"
              : "text/html",
        },
        text: () =>
          url === "/old"
            ? new Promise((resolve) => {
                resolveBody = resolve;
                bodyStarted();
              })
            : Promise.resolve('<main id="amarra-main">NEW</main>'),
      }),
    };
    const old = visit("/old", opts);
    await started;
    await visit("/new", opts);
    const historyCount = replacements.length;
    resolveBody(
      stream ? "event: remove\nid: victim\ndata: \n\n" : '<main id="amarra-main">OLD</main>'
    );
    assert.equal((await old).action, "superseded");
    assert.deepEqual(morphed, ["NEW"]);
    assert.deepEqual(pushed, ["/new"]);
    assert.equal(replacements.length, historyCount);
  });
}
