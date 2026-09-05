import test from "node:test";
import assert from "node:assert/strict";
import { shouldInterceptSubmit, start } from "./drive.mjs";

function fakeDocument() {
  const listeners = {};
  return {
    listeners,
    documentElement: { dataset: {} },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    dispatchEvent(ev) {
      for (const fn of listeners[ev.type] || []) fn(ev);
      return true;
    },
    querySelector() {
      return null;
    },
  };
}

function fire(doc, event) {
  for (const fn of doc.listeners[event.type] || []) fn(event);
}

function recordingFetch() {
  const fetches = [];
  const fetchFn = async (url, opts) => {
    fetches.push({ url, method: opts?.method, body: opts?.body, headers: opts?.headers });
    return {
      status: 200,
      url,
      text: async () => `<main id="amarra-main"><p>ok</p></main>`,
    };
  };
  return { fetches, fetchFn };
}

function clickEvent(anchor, extras = {}) {
  return {
    type: "click",
    defaultPrevented: false,
    button: extras.button ?? 0,
    metaKey: false,
    ctrlKey: false,
    shiftKey: false,
    altKey: false,
    target: anchor,
    preventDefault() {
      this.defaultPrevented = true;
    },
  };
}

function fakeAnchor({
  href,
  resolvedHref,
  target = "",
  download = false,
  skip = false,
  confirm = null,
  method = null,
} = {}) {
  return {
    tagName: "A",
    href: resolvedHref ?? href,
    target,
    download: download ? "" : undefined,
    getAttribute(name) {
      if (name === "href") return href;
      if (name === "target") return target;
      if (name === "data-amarra-confirm") return confirm;
      if (name === "data-amarra-method") return method;
      return null;
    },
    hasAttribute(name) {
      if (name === "download") return download;
      if (name === "data-amarra-skip") return skip;
      if (name === "data-amarra-confirm") return !!confirm;
      if (name === "data-amarra-method") return !!method;
      return false;
    },
    closest(sel) {
      return sel === "a[href]" ? this : null;
    },
  };
}

test("start does not Drive-fetch hash-only or middle clicks", async () => {
  const doc = fakeDocument();
  const location = {
    href: "http://a/page",
    origin: "http://a",
    pathname: "/page",
    search: "",
    hash: "",
  };
  const { fetches, fetchFn } = recordingFetch();
  start({ document: doc, location, fetchFn, popstate: false, morphFn() {} });

  const hashAttr = fakeAnchor({ href: "#section", resolvedHref: "http://a/page#section" });
  const hashClick = clickEvent(hashAttr);
  fire(doc, hashClick);
  assert.equal(hashClick.defaultPrevented, false);

  const bareHash = fakeAnchor({ href: "#", resolvedHref: "http://a/page#" });
  const bareClick = clickEvent(bareHash);
  fire(doc, bareClick);
  assert.equal(bareClick.defaultPrevented, false);

  const fullHash = fakeAnchor({
    href: "http://a/page#section",
    resolvedHref: "http://a/page#section",
  });
  fire(doc, clickEvent(fullHash));

  const middle = fakeAnchor({ href: "/x", resolvedHref: "http://a/x" });
  const middleClick = clickEvent(middle, { button: 1 });
  fire(doc, middleClick);
  assert.equal(middleClick.defaultPrevented, false);

  await Promise.resolve();
  assert.equal(fetches.length, 0);
});

test("start Drive-fetches same-origin path changes", async () => {
  const doc = fakeDocument();
  const location = { href: "http://a/page", origin: "http://a" };
  const { fetches, fetchFn } = recordingFetch();
  start({ document: doc, location, fetchFn, popstate: false, morphFn() {} });
  const a = fakeAnchor({ href: "/x", resolvedHref: "http://a/x" });
  const ev = clickEvent(a);
  fire(doc, ev);
  assert.equal(ev.defaultPrevented, true);
  await Promise.resolve();
  assert.equal(fetches.length, 1);
  assert.equal(fetches[0].url, "http://a/x");
});

test("shouldInterceptSubmit rejects skip, blank target, dialog, and cross-origin", () => {
  assert.equal(
    shouldInterceptSubmit({
      skip: true,
      action: "/x",
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptSubmit({
      target: "_blank",
      action: "/x",
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptSubmit({
      method: "dialog",
      action: "/x",
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptSubmit({
      action: "http://b/x",
      origin: "http://b",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptSubmit({
      method: "POST",
      action: "/x",
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    true
  );
});

function fakeForm({ action, method = "post", target = "", skip = false }) {
  return {
    tagName: "FORM",
    method,
    action,
    target,
    getAttribute(name) {
      if (name === "action") return action;
      if (name === "method") return method;
      if (name === "target") return target;
      return null;
    },
    hasAttribute(name) {
      return name === "data-amarra-skip" && skip;
    },
  };
}

function submitEvent(form, extras = {}) {
  return {
    type: "submit",
    defaultPrevented: false,
    target: form,
    submitter: extras.submitter ?? null,
    preventDefault() {
      this.defaultPrevented = true;
    },
  };
}

function FakeFormData(form, submitter) {
  this.form = form;
  this.submitter = submitter;
  this.pairs = [];
  if (submitter?.name) this.pairs.push([submitter.name, submitter.value ?? ""]);
}
FakeFormData.prototype.entries = function* entries() {
  yield* this.pairs;
};
FakeFormData.prototype.append = function append(key, value) {
  this.pairs.push([key, value]);
};

test("start does not intercept skip, blank, dialog, or cross-origin forms", async () => {
  const doc = fakeDocument();
  const location = { href: "http://a/page", origin: "http://a" };
  const { fetches, fetchFn } = recordingFetch();
  start({
    document: doc,
    location,
    fetchFn,
    popstate: false,
    morphFn() {},
    FormData: FakeFormData,
  });

  const skipped = submitEvent(fakeForm({ action: "/save", skip: true }));
  fire(doc, skipped);
  assert.equal(skipped.defaultPrevented, false);

  const blank = submitEvent(fakeForm({ action: "/save", target: "_blank" }));
  fire(doc, blank);
  assert.equal(blank.defaultPrevented, false);

  const dialog = submitEvent(fakeForm({ action: "/save", method: "dialog" }));
  fire(doc, dialog);
  assert.equal(dialog.defaultPrevented, false);

  const cross = submitEvent(fakeForm({ action: "http://b/save" }));
  fire(doc, cross);
  assert.equal(cross.defaultPrevented, false);

  await Promise.resolve();
  assert.equal(fetches.length, 0);
});

test("start skips navigation when confirm is cancelled", async () => {
  const doc = fakeDocument();
  const location = { href: "http://a/page", origin: "http://a" };
  const { fetches, fetchFn } = recordingFetch();
  start({
    document: doc,
    location,
    fetchFn,
    popstate: false,
    morphFn() {},
    confirm: () => false,
  });
  const ev = clickEvent(fakeAnchor({ href: "/x", resolvedHref: "http://a/x", confirm: "Delete?" }));
  fire(doc, ev);
  await Promise.resolve();
  assert.equal(fetches.length, 0);
  assert.equal(ev.defaultPrevented, false);
});

test("start uses data-amarra-method on links", async () => {
  const doc = fakeDocument();
  const location = { href: "http://a/page", origin: "http://a" };
  const { fetches, fetchFn } = recordingFetch();
  start({ document: doc, location, fetchFn, popstate: false, morphFn() {}, confirm: () => true });
  const ev = clickEvent(
    fakeAnchor({ href: "/items/1", resolvedHref: "http://a/items/1", method: "delete" })
  );
  fire(doc, ev);
  await Promise.resolve();
  assert.equal(fetches.length, 1);
  assert.equal(fetches[0].method, "DELETE");
});

test("start includes submitter and formaction in the Drive POST", async () => {
  const doc = fakeDocument();
  const location = { href: "http://a/page", origin: "http://a" };
  const { fetches, fetchFn } = recordingFetch();
  start({
    document: doc,
    location,
    fetchFn,
    popstate: false,
    morphFn() {},
    FormData: FakeFormData,
  });
  const form = fakeForm({ action: "/save", method: "post" });
  const submitter = {
    name: "op",
    value: "create",
    getAttribute(name) {
      if (name === "formaction") return "/alt";
      return null;
    },
  };
  const ev = submitEvent(form, { submitter });
  fire(doc, ev);
  assert.equal(ev.defaultPrevented, true);
  await Promise.resolve();
  assert.equal(fetches.length, 1);
  assert.equal(fetches[0].url, "/alt");
  assert.equal(fetches[0].method, "POST");
  assert.equal(fetches[0].body.submitter, submitter);
  assert.deepEqual([...fetches[0].body.entries()], [["op", "create"]]);
});
