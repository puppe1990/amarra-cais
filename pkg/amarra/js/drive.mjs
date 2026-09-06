import { morph } from "./morph.mjs";
import { csrfTokenFromMeta } from "./hook.mjs";
import { visitIntoFrame } from "./frame.mjs";
import { applyOp, isStreamResponse, parseSSE } from "./stream.mjs";
import { applyHead, hideProgress, showProgress } from "./drive_head.mjs";
import {
  confirmOk,
  disableSubmit,
  driveFormBody,
  requestMethod,
  restoreSubmit,
} from "./drive_form.mjs";
import { captureScroll, focusFirstInvalid, restoreScroll } from "./drive_restore.mjs";

export function shouldInterceptClick({
  href,
  target,
  download,
  origin,
  locationOrigin,
  skip,
  currentHref,
  resolvedHref,
  button,
} = {}) {
  if (skip) return false;
  if (button != null && button !== 0) return false;
  if (download) return false;
  if (target && target !== "_self") return false;
  if (!href) return false;
  if (/^(mailto|javascript|tel):/i.test(href)) return false;
  if (isHashOnlyNavigation(href, resolvedHref, currentHref)) return false;
  if (origin && locationOrigin && origin !== locationOrigin) return false;
  return true;
}

export function shouldInterceptSubmit({ skip, target, method, origin, locationOrigin } = {}) {
  if (skip) return false;
  if (target && target !== "_self") return false;
  if (String(method || "GET").toLowerCase() === "dialog") return false;
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
  const open = str.match(/<([a-zA-Z][\w:-]*)(?=[^>]*\sid\s*=\s*["']amarra-main["'])[^>]*>/i);
  if (!open) return null;
  const start = open.index + open[0].length;
  return sliceMatchingClose(str, start, open[1]);
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
  window: win,
} = {}) {
  if (status === 401 || status === 403) {
    location?.reload?.();
    return { action: "reload" };
  }
  if (status !== 200 && status !== 422) {
    emitDriveError(doc);
    return { action: "ignore" };
  }

  const fragment = extractMainHTML(html);
  if (fragment == null) return { action: "ignore" };
  applyHead(doc, html);
  if (main) (morphFn ?? morph)(main, fragment);
  if (status === 200 && push && url && history?.pushState) {
    if (!location?.href || url !== location.href) {
      history.pushState({ amarra: true, scrollY: 0 }, "", url);
    }
    win?.scrollTo?.(0, 0);
  }
  if (status === 200 && !push) restoreScroll(win, history?.state);
  if (status === 422) focusFirstInvalid(doc);
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:morphed", { bubbles: true }));
  }
  return { action: "morph" };
}

export async function visit(url, opts = {}) {
  const fetchFn = opts.fetchFn ?? opts.fetch ?? fetch;
  const doc = opts.document;
  showProgress(doc);
  try {
    const res = await fetchFn(url, {
      method: opts.method ?? "GET",
      headers: { ...driveHeaders(opts.csrfToken), ...opts.headers },
      body: opts.body,
      redirect: "follow",
      credentials: "same-origin",
    });
    const win = opts.window ?? (typeof window !== "undefined" ? window : null);
    const location = opts.location ?? win?.location ?? null;
    const history = opts.history ?? win?.history ?? null;
    if (opts.push !== false) captureScroll(history, win?.scrollY ?? 0);
    const html = res.status === 401 || res.status === 403 ? "" : await res.text();
    if (isStreamResponse(res.headers)) {
      for (const op of parseSSE(html)) applyOp(op, doc, opts);
      return { action: "stream" };
    }
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
      window: win,
    });
  } finally {
    hideProgress(doc);
  }
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
  const FormDataCtor = opts.FormData ?? (typeof FormData !== "undefined" ? FormData : null);
  const shared = { ...opts, document: doc, location, history, fetchFn, csrfToken };

  doc.addEventListener("click", (event) => {
    if (event.defaultPrevented) return;
    if (event.button != null && event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const a = findAnchor(event.target);
    if (!a) return;
    const href = a.getAttribute?.("href") ?? a.href;
    const resolved = resolveURL(href, location);
    if (
      !shouldInterceptClick({
        href,
        resolvedHref: resolved?.href,
        currentHref: location?.href,
        target: a.getAttribute?.("target") ?? a.target ?? "",
        download: !!(a.hasAttribute?.("download") || a.download),
        origin: resolved?.origin,
        locationOrigin: location?.origin,
        skip: a.hasAttribute?.("data-amarra-skip"),
        button: event.button,
      })
    ) {
      return;
    }
    if (!confirmOk(a, opts.confirm)) return;
    const method = requestMethod(a, "GET");
    event.preventDefault();
    const hrefURL = resolved?.href ?? href;
    void (async () => {
      if (await visitIntoFrame(a, hrefURL, shared)) return;
      await visit(hrefURL, { ...shared, method });
    })().catch(() => emitDriveError(doc));
  });

  doc.addEventListener("submit", (event) => {
    if (event.defaultPrevented) return;
    const form = findForm(event.target);
    if (!form) return;
    const submitter = event.submitter;
    const rawAction =
      submitter?.getAttribute?.("formaction") ||
      form.getAttribute?.("action") ||
      form.action ||
      location?.href ||
      "";
    const method = (
      submitter?.getAttribute?.("formmethod") ||
      form.getAttribute?.("method") ||
      form.method ||
      "GET"
    ).toUpperCase();
    const resolved = resolveURL(rawAction, location);
    if (
      !shouldInterceptSubmit({
        skip: form.hasAttribute?.("data-amarra-skip"),
        target: form.getAttribute?.("target") ?? form.target ?? "",
        method,
        origin: resolved?.origin,
        locationOrigin: location?.origin,
      })
    ) {
      return;
    }
    if (!confirmOk(form, opts.confirm) || !confirmOk(submitter, opts.confirm)) return;
    const verb = requestMethod(form, method);
    event.preventDefault();
    const fd = FormDataCtor ? formDataWithSubmitter(form, submitter, FormDataCtor) : null;
    const url = verb === "GET" ? withQuery(rawAction, fd) : rawAction;
    const body = verb === "GET" ? undefined : driveFormBody(fd);
    const disabled = disableSubmit(submitter);
    void visit(url, {
      ...shared,
      method: verb,
      body,
    })
      .catch(() => emitDriveError(doc))
      .finally(() => restoreSubmit(disabled));
  });

  if (typeof window !== "undefined" && opts.popstate !== false) {
    window.addEventListener("popstate", () => {
      void visit(location?.href ?? window.location.href, { ...shared, push: false });
    });
  }
}

function isHashOnlyNavigation(href, resolvedHref, currentHref) {
  if (href === "#" || (typeof href === "string" && href.startsWith("#"))) return true;
  if (!currentHref) return false;
  try {
    const next = new URL(resolvedHref || href, currentHref);
    const cur = new URL(currentHref);
    if (
      next.origin !== cur.origin ||
      next.pathname !== cur.pathname ||
      next.search !== cur.search
    ) {
      return false;
    }
    return next.hash !== cur.hash;
  } catch {
    return false;
  }
}

function sliceMatchingClose(str, start, tag) {
  const lower = str.toLowerCase();
  const name = tag.toLowerCase();
  const closeToken = `</${name}>`;
  let depth = 1;
  let i = start;
  while (i < str.length && depth > 0) {
    const nextOpen = findOpenTag(lower, i, name);
    const nextClose = lower.indexOf(closeToken, i);
    if (nextClose === -1) return null;
    if (nextOpen !== -1 && nextOpen < nextClose) {
      depth += 1;
      i = nextOpen + name.length + 1;
      continue;
    }
    depth -= 1;
    if (depth === 0) return str.slice(start, nextClose);
    i = nextClose + closeToken.length;
  }
  return null;
}

function findOpenTag(lower, from, name) {
  const token = `<${name}`;
  let i = from;
  while (i < lower.length) {
    const j = lower.indexOf(token, i);
    if (j === -1) return -1;
    const after = lower[j + token.length];
    if (after === ">" || after === "/" || (after && /\s/.test(after))) return j;
    i = j + token.length;
  }
  return -1;
}

function formDataWithSubmitter(form, submitter, Ctor) {
  try {
    return new Ctor(form, submitter);
  } catch {
    const fd = new Ctor();
    if (submitter?.name && typeof fd.append === "function") {
      fd.append(submitter.name, submitter.value ?? "");
    }
    return fd;
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
  if (href == null || href === "") {
    try {
      return location?.href ? new URL(location.href) : null;
    } catch {
      return null;
    }
  }
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
