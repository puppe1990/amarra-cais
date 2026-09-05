import { morph } from "./morph.mjs";
import { csrfTokenFromMeta } from "./hook.mjs";

export function shouldInterceptClick({
  href,
  target,
  download,
  origin,
  locationOrigin,
  skip,
} = {}) {
  if (skip) return false;
  if (download) return false;
  if (target && target !== "_self") return false;
  if (!href || href === "#") return false;
  if (/^(mailto|javascript|tel):/i.test(href)) return false;
  if (origin && locationOrigin && origin !== locationOrigin) return false;
  return true;
}

export function driveHeaders(csrfToken) {
  const headers = {
    "Amarra-Drive": "true",
    Accept: "text/html",
  };
  if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
  return headers;
}

export function extractMainHTML(html) {
  const str = String(html ?? "");
  const open = str.match(/<([a-zA-Z][\w:-]*)([^>]*\sid\s*=\s*["']amarra-main["'][^>]*)>/i);
  if (!open) return str;
  const tag = open[1];
  const start = open.index + open[0].length;
  const close = `</${tag}>`;
  const end = str.toLowerCase().lastIndexOf(close.toLowerCase());
  if (end === -1 || end < start) return str.slice(start);
  return str.slice(start, end);
}

export function applyDriveResponse({
  status,
  html,
  url,
  main,
  morphFn,
  location,
  history,
  document: doc,
  push = true,
} = {}) {
  if (status === 401 || status === 403) {
    location?.reload?.();
    return { action: "reload" };
  }
  if (status !== 200 && status !== 422) return { action: "ignore" };

  const fragment = extractMainHTML(html);
  if (main) (morphFn ?? morph)(main, fragment);
  if (status === 200 && push && url && history?.pushState) {
    if (!location?.href || url !== location.href) history.pushState({ amarra: true }, "", url);
  }
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:morphed", { bubbles: true }));
  }
  return { action: "morph" };
}

export async function visit(url, opts = {}) {
  const fetchFn = opts.fetchFn ?? opts.fetch ?? fetch;
  const res = await fetchFn(url, {
    method: opts.method ?? "GET",
    headers: { ...driveHeaders(opts.csrfToken), ...opts.headers },
    body: opts.body,
    redirect: "follow",
    credentials: "same-origin",
  });
  const location = opts.location ?? (typeof window !== "undefined" ? window.location : null);
  const history = opts.history ?? (typeof window !== "undefined" ? window.history : null);
  const doc = opts.document;
  const html = res.status === 401 || res.status === 403 ? "" : await res.text();
  return applyDriveResponse({
    status: res.status,
    html,
    url: res.url || url,
    main: doc?.querySelector?.("#amarra-main") ?? opts.main ?? null,
    morphFn: opts.morphFn,
    location,
    history,
    document: doc,
    push: opts.push !== false,
  });
}

export function start(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc || typeof doc.addEventListener !== "function") return;
  if (doc.documentElement?.dataset?.amarraDrive === "true") return;
  if (doc.documentElement?.dataset) doc.documentElement.dataset.amarraDrive = "true";

  const location = opts.location ?? (typeof window !== "undefined" ? window.location : null);
  const history = opts.history ?? (typeof window !== "undefined" ? window.history : null);
  const fetchFn = opts.fetchFn ?? opts.fetch ?? (typeof fetch !== "undefined" ? fetch : null);
  const csrfToken = opts.csrfToken ?? csrfTokenFromMeta(doc);
  const shared = { ...opts, document: doc, location, history, fetchFn, csrfToken };

  doc.addEventListener("click", (event) => {
    if (event.defaultPrevented) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const a = findAnchor(event.target);
    if (!a) return;
    const href = a.getAttribute?.("href") ?? a.href;
    const resolved = resolveURL(href, location);
    if (
      !shouldInterceptClick({
        href: resolved?.href ?? href,
        target: a.getAttribute?.("target") ?? a.target ?? "",
        download: !!(a.hasAttribute?.("download") || a.download),
        origin: resolved?.origin,
        locationOrigin: location?.origin,
        skip: a.hasAttribute?.("data-amarra-skip"),
      })
    ) {
      return;
    }
    event.preventDefault();
    void visit(resolved?.href ?? href, shared).catch(() => emitDriveError(doc));
  });

  doc.addEventListener("submit", (event) => {
    if (event.defaultPrevented) return;
    const form = findForm(event.target);
    if (!form || form.hasAttribute?.("data-amarra-skip")) return;
    event.preventDefault();
    const method = (form.getAttribute?.("method") || form.method || "GET").toUpperCase();
    const action = form.getAttribute?.("action") || form.action || location?.href || "";
    const fd = typeof FormData === "function" ? new FormData(form) : null;
    const url = method === "GET" ? withQuery(action, fd) : action;
    void visit(url, {
      ...shared,
      method,
      body: method === "GET" ? undefined : fd,
    }).catch(() => emitDriveError(doc));
  });

  if (typeof window !== "undefined" && opts.popstate !== false) {
    window.addEventListener("popstate", () => {
      void visit(location?.href ?? window.location.href, { ...shared, push: false });
    });
  }
}

function emitDriveError(doc) {
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:drive-error", { bubbles: true }));
  }
}

function findAnchor(target) {
  if (!target) return null;
  if (typeof target.closest === "function") return target.closest("a[href]");
  let node = target;
  while (node) {
    const tag = node.tagName;
    if ((tag === "A" || tag === "a") && (node.href || node.getAttribute?.("href"))) return node;
    node = node.parentElement || node.parentNode;
  }
  return null;
}

function findForm(target) {
  if (!target) return null;
  if (target.tagName === "FORM" || target.tagName === "form") return target;
  if (typeof target.closest === "function") return target.closest("form");
  return null;
}

function resolveURL(href, location) {
  if (!href) return null;
  try {
    return new URL(href, location?.href ?? location?.origin ?? "http://localhost");
  } catch {
    return null;
  }
}

function withQuery(action, fd) {
  if (!fd || typeof fd.entries !== "function") return action;
  const params = new URLSearchParams();
  for (const [key, value] of fd.entries()) {
    if (typeof value === "string") params.append(key, value);
  }
  const q = params.toString();
  if (!q) return action;
  return action.includes("?") ? `${action}&${q}` : `${action}?${q}`;
}
