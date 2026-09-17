import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import vm from "node:vm";

const SW_SOURCE = readFileSync(new URL("../pwa/assets/sw.js", import.meta.url), "utf8");

function response(body, headers = {}) {
  return {
    ok: true,
    type: "basic",
    body,
    headers: { get: (name) => headers[String(name).toLowerCase()] ?? null },
    clone() {
      return response(body, headers);
    },
  };
}

function request(url, { method = "GET", mode = "cors", accept = "text/html" } = {}) {
  return {
    url,
    method,
    mode,
    headers: {
      get: (name) => (String(name).toLowerCase() === "accept" ? accept : null),
    },
  };
}

function createServiceWorker({ fetchImpl } = {}) {
  const handlers = {};
  const cacheStore = new Map();
  const deleted = [];
  const cache = {
    put(req, res) {
      cacheStore.set(req.url, res);
      return Promise.resolve();
    },
    match(req) {
      return Promise.resolve(cacheStore.get(typeof req === "string" ? req : req.url));
    },
    addAll(urls) {
      for (const url of urls) cacheStore.set(url, response(url));
      return Promise.resolve();
    },
  };
  const sandbox = {
    self: {
      addEventListener(type, fn) {
        (handlers[type] ||= []).push(fn);
      },
      skipWaiting() {
        return Promise.resolve();
      },
      clients: {
        claim() {
          return Promise.resolve();
        },
      },
    },
    caches: {
      open() {
        return Promise.resolve(cache);
      },
      keys() {
        return Promise.resolve([]);
      },
      delete(name) {
        deleted.push(name);
        return Promise.resolve(true);
      },
      match(req) {
        const key = typeof req === "string" ? new URL(req, "https://app.test").href : req.url;
        return Promise.resolve(cacheStore.get(key));
      },
    },
    fetch: fetchImpl ?? (() => Promise.resolve(response("ok"))),
    Response: { error: () => ({ type: "error" }) },
    URL,
  };
  vm.createContext(sandbox);
  vm.runInContext(SW_SOURCE, sandbox);
  return { handlers, cacheStore, deleted };
}

async function fire(handlers, req) {
  let responded;
  for (const fn of handlers.fetch || []) {
    fn({
      request: req,
      respondWith(promise) {
        responded = promise;
      },
    });
  }
  const res = await responded;
  await new Promise((resolve) => setTimeout(resolve, 0));
  return res;
}

// #114: view.Write sends Cache-Control: no-store on authenticated pages and
// Drive fragments are per-user; the SW cached any ok response and served it in
// the offline fallback — leaking the previous user's dashboard on shared devices.
test("service worker never caches no-store responses", async () => {
  const { handlers, cacheStore } = createServiceWorker({
    fetchImpl: () => Promise.resolve(response("dashboard", { "cache-control": "no-store" })),
  });
  const res = await fire(handlers, request("https://app.test/dashboard", { mode: "navigate" }));
  assert.equal(res.body, "dashboard");
  assert.equal(cacheStore.size, 0, "no-store HTML must not enter the Cache Storage");
});

test("service worker never caches private responses", async () => {
  const { handlers, cacheStore } = createServiceWorker({
    fetchImpl: () =>
      Promise.resolve(response("dashboard", { "cache-control": "private, max-age=0" })),
  });
  await fire(handlers, request("https://app.test/dashboard", { mode: "navigate" }));
  assert.equal(cacheStore.size, 0, "private HTML must not enter the Cache Storage");
});

test("service worker still caches cacheable static assets", async () => {
  const { handlers, cacheStore } = createServiceWorker({
    fetchImpl: () => Promise.resolve(response("js", { "cache-control": "max-age=3600" })),
  });
  const res = await fire(
    handlers,
    request("https://app.test/static/js/amarra.js", { accept: "*/*" })
  );
  assert.equal(res.body, "js");
  assert.equal(cacheStore.get("https://app.test/static/js/amarra.js").body, "js");
});

test("navigation fallback serves offline.html instead of a cached page", async () => {
  const { handlers, cacheStore } = createServiceWorker({
    fetchImpl: () => Promise.reject(new Error("offline")),
  });
  cacheStore.set("https://app.test/static/offline.html", response("offline-shell"));
  cacheStore.set("https://app.test/dashboard", response("previous-user-dashboard"));

  const res = await fire(handlers, request("https://app.test/dashboard", { mode: "navigate" }));
  assert.equal(res.body, "offline-shell");
});

test("Drive fragment failure never serves a cached page", async () => {
  const { handlers, cacheStore } = createServiceWorker({
    fetchImpl: () => Promise.reject(new Error("offline")),
  });
  cacheStore.set("https://app.test/dashboard", response("previous-user-dashboard"));

  const res = await fire(handlers, request("https://app.test/dashboard"));
  assert.equal(res.type, "error");
});

test("POST /logout clears cached entries", async () => {
  const { handlers, cacheStore, deleted } = createServiceWorker({
    fetchImpl: () => Promise.resolve(response("", { location: "/login" })),
  });
  cacheStore.set("https://app.test/dashboard", response("secret"));

  const res = await fire(handlers, request("https://app.test/logout", { method: "POST" }));
  assert.equal(res.ok, true);
  assert.ok(deleted.length > 0, "logout must clear the cache");
});

test("amarra:clear-cache message wipes the cache", async () => {
  const { handlers, deleted } = createServiceWorker();
  const waits = [];
  for (const fn of handlers.message || []) {
    fn({ data: { type: "amarra:clear-cache" }, waitUntil: (promise) => waits.push(promise) });
  }
  await Promise.all(waits);
  assert.ok(deleted.length > 0, "message must clear the cache");
});
