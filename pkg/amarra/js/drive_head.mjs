import { csrfTokenFromMeta } from "./hook.mjs";

export function extractTitle(html) {
  const m = String(html ?? "").match(/<title\b[^>]*>([\s\S]*?)<\/title>/i);
  return m ? m[1].trim() : null;
}

export function extractHTMLAttr(html, name) {
  const open = String(html ?? "").match(/<html\b[^>]*>/i)?.[0] ?? "";
  const escaped = String(name).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const m = open.match(new RegExp(`\\s${escaped}\\s*=\\s*["']([^"']*)["']`, "i"));
  return m ? m[1] : null;
}

export function applyHead(doc, html) {
  if (!doc) return;
  const title = extractTitle(html);
  if (title != null) doc.title = title;
  const lang = extractHTMLAttr(html, "lang");
  if (lang != null && doc.documentElement) doc.documentElement.lang = lang;
  const token = csrfTokenFromMeta(html);
  if (!token) return;
  const meta = doc.querySelector?.('meta[name="csrf-token"]');
  if (!meta) return;
  if (typeof meta.setAttribute === "function") meta.setAttribute("content", token);
  else meta.content = token;
}

export function progressIconSrc(doc) {
  try {
    const link = doc?.querySelector?.('link[rel="icon"]');
    const href = link?.getAttribute?.("href") ?? link?.href;
    if (href) return href;
  } catch {
    /* noop: fall through to default */
  }
  return "/static/icons/icon.png";
}

function ensureVeilKeyframes(doc) {
  if (!doc?.createElement || doc.getElementById?.("amarra-veil-style")) return;
  const style = doc.createElement("style");
  style.id = "amarra-veil-style";
  style.textContent =
    "@keyframes amarra-icon-pulse{0%,100%{transform:scale(1);opacity:.85}50%{transform:scale(1.12);opacity:1}}" +
    "@media (prefers-reduced-motion:reduce){#amarra-veil img{animation:none!important}}";
  (doc.head ?? doc.body)?.appendChild?.(style);
}

export function showProgress(doc) {
  if (!doc?.createElement || !doc.body) return;
  let bar = doc.getElementById?.("amarra-progress");
  if (!bar) {
    bar = doc.createElement("div");
    bar.id = "amarra-progress";
    bar.setAttribute("role", "progressbar");
    bar.style.cssText =
      "position:fixed;top:0;left:0;right:0;height:2px;background:#c9893a;z-index:9999";
    doc.body.appendChild(bar);
  }
  bar.hidden = false;
  let veil = doc.getElementById?.("amarra-veil");
  if (!veil) {
    veil = doc.createElement("div");
    veil.id = "amarra-veil";
    veil.setAttribute("role", "status");
    veil.setAttribute("aria-label", "Loading");
    veil.style.cssText =
      "position:fixed;inset:0;z-index:9998;display:flex;align-items:center;justify-content:center;" +
      "background:rgba(10,10,12,.45);backdrop-filter:blur(2px);opacity:0;transition:opacity .18s ease;pointer-events:none";
    const img = doc.createElement("img");
    img.setAttribute("alt", "");
    img.style.cssText =
      "width:56px;height:56px;border-radius:14px;animation:amarra-icon-pulse 1.1s ease-in-out infinite";
    veil.appendChild?.(img);
    ensureVeilKeyframes(doc);
    doc.body.appendChild(veil);
  }
  veil.querySelector?.("img")?.setAttribute?.("src", progressIconSrc(doc));
  veil.hidden = false;
  if (veil.style) veil.style.opacity = "1";
}

export function hideProgress(doc) {
  const bar = doc?.getElementById?.("amarra-progress");
  if (bar) bar.hidden = true;
  const veil = doc?.getElementById?.("amarra-veil");
  if (veil) {
    if (veil.style) veil.style.opacity = "0";
    veil.hidden = true;
  }
}
