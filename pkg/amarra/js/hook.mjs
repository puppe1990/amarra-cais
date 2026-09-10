import { register, scan, dispatchLivePush, reset } from "./hook_registry.mjs";
import { bulk } from "./hook_bulk.mjs";
import { clipboard } from "./hook_clipboard.mjs";
import { dialog } from "./hook_dialog.mjs";
import { dropdown } from "./hook_dropdown.mjs";
import { nav } from "./hook_nav.mjs";
import { password } from "./hook_password.mjs";
import { reveal } from "./hook_reveal.mjs";
import { theme } from "./hook_theme.mjs";

export { register, scan, dispatchLivePush, reset };

register("bulk", bulk);
register("clipboard", clipboard);
register("dialog", dialog);
register("dropdown", dropdown);
register("nav", nav);
register("password", password);
register("reveal", reveal);
register("theme", theme);

const ON_CLASSES = ["bg-green-50", "text-green-700"];
const OFF_CLASSES = ["bg-slate-100", "text-slate-600"];
const TOAST_MS = 2000;

export function csrfTokenFromMeta(htmlOrDoc) {
  if (!htmlOrDoc) return "";
  if (typeof htmlOrDoc === "string") {
    const named = htmlOrDoc.match(/<meta\b[^>]*\bname\s*=\s*["']csrf-token["'][^>]*>/i);
    const tag =
      named?.[0] ||
      htmlOrDoc.match(
        /<meta\b[^>]*\bcontent\s*=\s*["'][^"']*["'][^>]*\bname\s*=\s*["']csrf-token["'][^>]*>/i
      )?.[0];
    if (!tag) return "";
    const content = tag.match(/\bcontent\s*=\s*["']([^"']*)["']/i);
    return content ? content[1] : "";
  }
  const meta = htmlOrDoc.querySelector?.('meta[name="csrf-token"]');
  if (!meta) return "";
  return meta.content || meta.getAttribute?.("content") || "";
}

export function showToast(message, doc, opts = {}) {
  if (!message || !doc) return;
  const host = doc.getElementById?.("amarra-toast-host");
  if (!host) return;
  if (host._amarraToastTimer) {
    clearTimeout(host._amarraToastTimer);
    host._amarraToastTimer = null;
  }
  host.innerHTML =
    '<div class="amarra-toast-enter fixed top-24 left-1/2 -translate-x-1/2 z-50 bg-slate-900 text-white px-5 py-3 rounded-2xl shadow-xl flex items-center gap-2 border border-slate-700/50" role="status">' +
    '<span class="text-xs font-bold"></span></div>';
  const span = host.querySelector?.("span");
  if (span) span.textContent = message;
  const duration = opts.duration ?? TOAST_MS;
  if (duration > 0) {
    host._amarraToastTimer = setTimeout(() => {
      host.innerHTML = "";
      host._amarraToastTimer = null;
    }, duration);
  }
}

export function applyFocus(selector, doc) {
  if (!selector || !doc?.querySelector) return;
  const el = doc.querySelector(selector);
  if (el && typeof el.focus === "function") el.focus();
}

export function afterMorph(doc) {
  if (!doc?.querySelector) return;
  const marked = doc.querySelector("[data-amarra-focus]");
  if (!marked) return;
  const sel = marked.getAttribute?.("data-amarra-focus");
  if (sel && sel !== "true") {
    applyFocus(sel, doc);
    return;
  }
  if (typeof marked.focus === "function") marked.focus();
}

export function applyOptimistic(el, mode) {
  if (!el) return null;
  if (mode === "count") {
    const countEl = el.querySelector?.("[data-amarra-count]") || el;
    const prev = countEl.textContent;
    const n = parseInt(String(prev ?? "").trim(), 10);
    countEl.textContent = String((Number.isNaN(n) ? 0 : n) + 1);
    return { el, mode, countEl, prev };
  }
  if (mode === "remove") {
    const hadOpacity = !!el.classList?.contains("opacity-0");
    el.classList?.add("opacity-0", "transition-opacity", "duration-150");
    return { el, mode, hadOpacity };
  }
  const wasOn = hasClasses(el, ON_CLASSES);
  if (wasOn) setClasses(el, OFF_CLASSES, ON_CLASSES);
  else setClasses(el, ON_CLASSES, OFF_CLASSES);
  return { el, mode: "toggle", wasOn };
}

export function rollbackOptimistic(state) {
  if (!state?.el) return;
  const { el } = state;
  if (state.mode === "count") {
    state.countEl.textContent = state.prev;
    return;
  }
  if (state.mode === "remove") {
    if (!state.hadOpacity) el.classList?.remove("opacity-0", "transition-opacity", "duration-150");
    return;
  }
  if (state.wasOn) setClasses(el, ON_CLASSES, OFF_CLASSES);
  else setClasses(el, OFF_CLASSES, ON_CLASSES);
}

export function start(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc || typeof doc.addEventListener !== "function") return;
  if (doc.documentElement?.dataset?.amarraHook === "true") return;
  if (doc.documentElement?.dataset) doc.documentElement.dataset.amarraHook = "true";

  register("bulk", bulk);
  register("clipboard", clipboard);
  register("dialog", dialog);
  register("dropdown", dropdown);
  register("nav", nav);
  register("password", password);
  register("reveal", reveal);
  register("theme", theme);
  scan(doc);

  let optimistic = null;

  doc.addEventListener("amarra:toast", (ev) => {
    showToast(ev.detail?.message ?? "", doc, opts);
  });

  doc.addEventListener("amarra:morphed", () => {
    optimistic = null;
    afterMorph(doc);
    scan(doc);
  });

  doc.addEventListener("amarra:drive-error", () => {
    rollbackOptimistic(optimistic);
    optimistic = null;
  });

  doc.addEventListener(
    "click",
    (ev) => {
      const target = ev.target?.closest?.("[data-amarra-optimistic]");
      if (!target) return;
      optimistic = applyOptimistic(target, target.getAttribute("data-amarra-optimistic"));
      target.setAttribute?.("aria-busy", "true");
    },
    true
  );
}

function hasClasses(el, classes) {
  return classes.every((c) => el.classList?.contains(c));
}

function setClasses(el, add, remove) {
  remove.forEach((c) => el.classList?.remove(c));
  add.forEach((c) => el.classList?.add(c));
}
