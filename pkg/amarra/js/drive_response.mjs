import { morph } from "./morph.mjs";
import { applyHead, extractBodyAttr, extractHTMLAttr } from "./drive_head.mjs";
import { focusFirstInvalid, restoreScroll } from "./drive_restore.mjs";
import { extractMainHTML, extractMainTagName } from "./drive_html.mjs";
import { applyOutlets } from "./drive_outlets.mjs";

export const DRIVE_ERROR_MESSAGE = "Não foi possível carregar a página.";

export function emitDriveError(doc, detail) {
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(
      new CustomEvent("amarra:drive-error", { bubbles: true, detail: detail ?? {} })
    );
  }
}

function defaultDriveWarn(msg) {
  try {
    globalThis.console?.warn?.(msg);
  } catch {
    /* noop: console unavailable */
  }
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
  applyOutlets(doc, html);
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

export function hashOf(url) {
  const str = String(url ?? "");
  const at = str.indexOf("#");
  return at === -1 ? "" : str.slice(at);
}

export function assignLocation(location, url) {
  if (url && typeof location?.assign === "function") {
    location.assign(url);
    return;
  }
  if (url && location) location.href = url;
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
