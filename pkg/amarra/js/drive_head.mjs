import { csrfTokenFromMeta } from "./hook.mjs";

export function extractTitle(html) {
  const m = String(html ?? "").match(/<title\b[^>]*>([\s\S]*?)<\/title>/i);
  return m ? m[1].trim() : null;
}

export function applyHead(doc, html) {
  if (!doc) return;
  const title = extractTitle(html);
  if (title != null) doc.title = title;
  const token = csrfTokenFromMeta(html);
  if (!token) return;
  const meta = doc.querySelector?.('meta[name="csrf-token"]');
  if (!meta) return;
  if (typeof meta.setAttribute === "function") meta.setAttribute("content", token);
  else meta.content = token;
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
}

export function hideProgress(doc) {
  const bar = doc?.getElementById?.("amarra-progress");
  if (bar) bar.hidden = true;
}
