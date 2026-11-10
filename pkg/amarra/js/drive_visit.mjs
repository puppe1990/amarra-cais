import { bumpSequence, isCurrentSequence } from "./sequence.mjs";
import { csrfTokenFromMeta } from "./hook.mjs";
import { visitIntoFrame } from "./frame.mjs";
import { applyOp, isStreamResponse, parseSSE } from "./stream.mjs";
import { hideProgress, showProgress } from "./drive_head.mjs";
import {
  confirmOk,
  disableSubmit,
  driveFormBody,
  requestMethod,
  restoreSubmit,
  withQuery,
} from "./drive_form.mjs";
import { captureScroll } from "./drive_restore.mjs";
import { downloadBlobResponse, isMutationMethod } from "./drive_download.mjs";
import {
  applyDriveResponse,
  assignLocation,
  DRIVE_ERROR_MESSAGE,
  emitDriveError,
  hashOf,
} from "./drive_response.mjs";

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

export async function visit(url, opts = {}) {
  const fetchFn = opts.fetchFn ?? opts.fetch ?? fetch;
  const doc = opts.document;
  const method = opts.method ?? "GET";
  const seq = bumpSequence(doc);
  showProgress(doc);
  try {
    const res = await fetchFn(url, {
      method,
      headers: { ...driveHeaders(opts.csrfToken), ...opts.headers },
      body: opts.body,
      redirect: "follow",
      credentials: "same-origin",
    });
    if (!isCurrentSequence(doc, seq)) {
      return { action: "superseded" };
    }
    const win = opts.window ?? (typeof window !== "undefined" ? window : null);
    const location = opts.location ?? win?.location ?? null;
    const history = opts.history ?? win?.history ?? null;
    if (opts.push !== false) captureScroll(history, win?.scrollY ?? 0);
    const authFailure = res.status === 401 || res.status === 403;
    const stream = isStreamResponse(res.headers);
    if (!authFailure && !stream && !isHtmlResponse(res.headers)) {
      // #332: a PDF/DXF/SVG/attachment response has no #amarra-main to morph.
      // Drive used to read it as text, end as "ignore" after preventDefault,
      // and the file never downloaded. A GET can be handed to the browser;
      // #366: a mutation cannot, because assign would discard the response and
      // replay the POST as GET (losing body/filters). Save those bytes instead.
      if (isMutationMethod(method)) {
        const file = await downloadBlobResponse(res, {
          document: doc,
          url: res.url || url,
          createObjectURL: opts.createObjectURL,
          revokeObjectURL: opts.revokeObjectURL,
          schedule: opts.schedule,
          isCurrent: () => isCurrentSequence(doc, seq),
        });
        if (!file) return { action: "superseded" };
        return { action: "download", filename: file.filename };
      }
      assignLocation(location, res.url || url);
      return { action: "assign" };
    }
    const html = authFailure ? "" : await res.text();
    // Reading the body can outlive a newer visit even after fetch resolved (#348).
    if (!isCurrentSequence(doc, seq)) return { action: "superseded" };
    if (stream) {
      for (const op of parseSSE(html)) applyOp(op, doc, opts);
      return { action: "stream" };
    }
    return applyDriveResponse({
      status: res.status,
      html,
      url: withRequestHash(res.url || url, url),
      main: doc?.querySelector?.("#amarra-main") ?? opts.main ?? null,
      morphFn: opts.morphFn,
      location,
      history,
      document: doc,
      push: opts.push !== false,
      window: win,
    });
  } catch {
    emitDriveError(doc, { message: DRIVE_ERROR_MESSAGE });
    return { action: "error" };
  } finally {
    if (isCurrentSequence(doc, seq)) hideProgress(doc);
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
    event.preventDefault();
    if (!confirmOk(a, opts.confirm)) return;
    const method = requestMethod(a, "GET");
    const hrefURL = resolved?.href ?? href;
    void (async () => {
      if (await visitIntoFrame(a, hrefURL, shared)) return;
      await visit(hrefURL, { ...shared, method });
    })().catch(() => emitDriveError(doc, { message: DRIVE_ERROR_MESSAGE }));
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
    event.preventDefault();
    if (!confirmOk(form, opts.confirm) || !confirmOk(submitter, opts.confirm)) return;
    const verb = requestMethod(form, method);
    const fd = FormDataCtor ? formDataWithSubmitter(form, submitter, FormDataCtor) : null;
    const url = verb === "GET" ? withQuery(rawAction, fd) : rawAction;
    const body = verb === "GET" ? undefined : driveFormBody(fd);
    const disabled = disableSubmit(submitter);
    void visit(url, {
      ...shared,
      method: verb,
      body,
    })
      .catch(() => emitDriveError(doc, { message: DRIVE_ERROR_MESSAGE }))
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

// fetch strips the fragment from both the request and `Response.url`, so the
// visited URL loses `#planos` unless the requested fragment is carried over.
function withRequestHash(responseURL, requestURL) {
  const hash = hashOf(requestURL);
  if (!hash) return responseURL;
  const base = String(responseURL ?? "");
  const at = base.indexOf("#");
  return (at === -1 ? base : base.slice(0, at)) + hash;
}

// #332: pages are HTML; files (PDF/DXF/SVG downloads, attachment exports) are
// not. An empty type stays HTML — minimal responses and test doubles may not
// carry headers, and those are still drive-navigable pages.
function isHtmlResponse(headers) {
  const type = String(headers?.get?.("content-type") ?? "")
    .split(";")[0]
    .trim()
    .toLowerCase();
  if (/attachment/i.test(String(headers?.get?.("content-disposition") ?? ""))) return false;
  return type === "" || type === "text/html" || type === "application/xhtml+xml";
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
