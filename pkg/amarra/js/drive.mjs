import { morph } from "./morph.mjs";
import { bumpSequence, isCurrentSequence } from "./sequence.mjs";
import { csrfTokenFromMeta } from "./hook.mjs";
import { visitIntoFrame } from "./frame.mjs";
import { applyOp, isStreamResponse, parseSSE } from "./stream.mjs";
import {
  applyHead,
  extractBodyAttr,
  extractHTMLAttr,
  hideProgress,
  showProgress,
} from "./drive_head.mjs";
import {
  confirmOk,
  disableSubmit,
  driveFormBody,
  requestMethod,
  restoreSubmit,
  withQuery,
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
  const open = extractMainOpen(str);
  if (!open) return null;
  const start = open.index + open[0].length;
  return sliceMatchingClose(str, start, open[1]);
}

export function extractMainTagName(html) {
  return extractMainOpen(String(html ?? ""))?.[1]?.toUpperCase() ?? null;
}

function defaultDriveWarn(msg) {
  try {
    globalThis.console?.warn?.(msg);
  } catch {
    /* noop: console unavailable */
  }
}

const DRIVE_ERROR_MESSAGE = "Não foi possível carregar a página.";

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
  warn = defaultDriveWarn,
} = {}) {
  if (status === 401 || status === 403) {
    location?.reload?.();
    return { action: "reload" };
  }

  const fragment = extractMainHTML(html);
  if (fragment == null) {
    // #360: 404/500 without a Drive page used to stay silent after
    // preventDefault. Tell the user; keep 200/422 on the missing-main warning.
    if (status !== 200 && status !== 422) {
      emitDriveError(doc, { status, message: DRIVE_ERROR_MESSAGE });
      return { action: "ignore" };
    }
    // Silent ignores cost hours of debugging (#83): the target HTML has no
    // (or unbalanced) #amarra-main, so there is nothing to morph.
    warn(
      `amarra drive: ignored ${status} response for ${url ?? "(unknown url)"} — #amarra-main missing or unbalanced HTML`
    );
    return { action: "ignore" };
  }
  if (needsFullVisit({ html, main, document: doc })) {
    assignLocation(location, url);
    return { action: "assign" };
  }
  applyHead(doc, html);
  applyLayoutMarker(doc, html);
  applyShellMarker(doc, html);
  if (main) (morphFn ?? morph)(main, fragment);
  // 404 is a navigation to a missing resource; keep the URL in sync (#360).
  // 500/422 stay on the current URL so a refresh does not re-POST.
  if ((status === 200 || status === 404) && push && url && history?.pushState) {
    if (!location?.href || url !== location.href) {
      history.pushState({ amarra: true, scrollY: 0 }, "", url);
    }
    scrollToFragment(doc, win, url);
  }
  if (status === 200 && !push) restoreScroll(win, history?.state);
  if (status === 422) focusFirstInvalid(doc);
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:morphed", { bubbles: true }));
  }
  return { action: "morph" };
}

// #249: a login/logout POST swaps the layout shell in the same request (rail +
// topbar vs centered auth card), but Drive only morphs #amarra-main — the
// authenticated page would then sit inside the anonymous shell until a manual
// reload. The layout marks the shell on <body>, so a change means full visit.
function needsFullVisit({ html, main, document: doc }) {
  if (
    markerChanged(
      doc?.documentElement?.dataset?.amarraLayout,
      extractHTMLAttr(html, "data-amarra-layout")
    )
  ) {
    return true;
  }
  if (markerChanged(doc?.body?.dataset?.amarraShell, extractBodyAttr(html, "data-amarra-shell"))) {
    return true;
  }
  const currentTag = main?.tagName?.toUpperCase?.();
  const nextTag = extractMainTagName(html);
  return !!(currentTag && nextTag && currentTag !== nextTag);
}

function markerChanged(current, next) {
  return !!(current && next && current !== next);
}

// #341: with a fragment the reader asked for a section, and a cold load scrolls
// there natively — after a morph Drive has to do it instead of jumping to the top.
// getElementById (not querySelector) because an HTML5 id may start with a digit.
function scrollToFragment(doc, win, url) {
  const target = fragmentTarget(doc, url);
  if (target && typeof target.scrollIntoView === "function") {
    target.scrollIntoView();
    return;
  }
  win?.scrollTo?.(0, 0);
}

function fragmentTarget(doc, url) {
  const id = decodeFragment(hashOf(url));
  if (!id || typeof doc?.getElementById !== "function") return null;
  return doc.getElementById(id);
}

function decodeFragment(hash) {
  const raw = hash.startsWith("#") ? hash.slice(1) : hash;
  if (!raw) return "";
  try {
    return decodeURIComponent(raw);
  } catch {
    return raw;
  }
}

function hashOf(url) {
  const str = String(url ?? "");
  const at = str.indexOf("#");
  return at === -1 ? "" : str.slice(at);
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

function assignLocation(location, url) {
  if (url && typeof location?.assign === "function") {
    location.assign(url);
    return;
  }
  if (url && location) location.href = url;
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

function applyLayoutMarker(doc, html) {
  const layout = extractHTMLAttr(html, "data-amarra-layout");
  if (layout != null && doc?.documentElement?.dataset) {
    doc.documentElement.dataset.amarraLayout = layout;
  }
}

function applyShellMarker(doc, html) {
  const shell = extractBodyAttr(html, "data-amarra-shell");
  if (shell != null && doc?.body?.dataset) {
    doc.body.dataset.amarraShell = shell;
  }
}

export async function visit(url, opts = {}) {
  const fetchFn = opts.fetchFn ?? opts.fetch ?? fetch;
  const doc = opts.document;
  const seq = bumpSequence(doc);
  showProgress(doc);
  try {
    const res = await fetchFn(url, {
      method: opts.method ?? "GET",
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
      // and the file never downloaded — hand the navigation to the browser.
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

function extractMainOpen(str) {
  return str.match(/<([a-zA-Z][\w:-]*)(?=[^>]*\sid\s*=\s*["']amarra-main["'])[^>]*>/i);
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

function emitDriveError(doc, detail) {
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(
      new CustomEvent("amarra:drive-error", { bubbles: true, detail: detail ?? {} })
    );
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
