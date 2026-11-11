const CLICK = "_amarraThemeClick";
const DEFAULT_KEY = "amarra-theme";
const DEFAULT_CLASS = "light";

function themeKey(el, opts) {
  return el?.getAttribute?.("data-amarra-theme-key") || opts.key || DEFAULT_KEY;
}

function themeClass(el, opts) {
  return el?.getAttribute?.("data-amarra-theme-class") || opts.className || DEFAULT_CLASS;
}

function applyTheme(on, el, deps) {
  const className = themeClass(el, deps.opts);
  const key = themeKey(el, deps.opts);
  const html = deps.getHtml();
  if (html?.classList) {
    if (on) html.classList.add(className);
    else html.classList.remove(className);
  }
  try {
    deps.getStorage()?.setItem?.(key, on ? className : "");
  } catch {
    // private mode / blocked storage
  }
  const meta = deps.getMeta?.();
  const lightColor = el?.getAttribute?.("data-amarra-theme-color") || deps.opts.lightColor;
  const darkColor = el?.getAttribute?.("data-amarra-theme-color-off") || deps.opts.darkColor;
  const color = on ? lightColor : darkColor;
  if (meta && color) meta.setAttribute?.("content", color);
  // #30: swap the visible label (on = theme applied) and keep aria-pressed.
  const onLabel = el?.getAttribute?.("data-amarra-theme-on-label") || deps.opts.onLabel;
  const offLabel = el?.getAttribute?.("data-amarra-theme-off-label") || deps.opts.offLabel;
  const label = on ? onLabel : offLabel;
  if (label && el) swapThemeLabel(el, label);
  el?.setAttribute?.("aria-pressed", on ? "true" : "false");
}

export function makeTheme(opts = {}) {
  const deps = {
    opts,
    getHtml: opts.html ?? (() => globalThis.document?.documentElement),
    getStorage: () => opts.storage ?? globalThis.localStorage,
    getMeta:
      opts.themeColorMeta ??
      (() => globalThis.document?.querySelector?.('meta[name="theme-color"]')),
  };
  // #30: key/class are per-element so apps can rebrand the hook from HTML
  // (data-amarra-theme-key / data-amarra-theme-class) without re-registering.
  // The FOUC snippet in the layout head must use the same key.
  return {
    connect(el) {
      connectTheme(el, deps);
    },
    disconnect(el) {
      disconnectTheme(el);
    },
  };
}

function connectTheme(el, deps) {
  if (!el || typeof el.addEventListener !== "function") return;
  let stored = "";
  try {
    stored = deps.getStorage()?.getItem?.(themeKey(el, deps.opts)) ?? "";
  } catch {
    stored = "";
  }
  if (stored === themeClass(el, deps.opts)) applyTheme(true, el, deps);
  const fn = (ev) => {
    ev?.preventDefault?.();
    const html = deps.getHtml();
    const on = !html?.classList?.contains?.(themeClass(el, deps.opts));
    applyTheme(on, el, deps);
  };
  el[CLICK] = fn;
  el.addEventListener("click", fn);
}

function disconnectTheme(el) {
  const fn = el?.[CLICK];
  if (!fn || typeof el.removeEventListener !== "function") return;
  el.removeEventListener("click", fn);
  delete el[CLICK];
}

export const theme = makeTheme();

function swapThemeLabel(el, label) {
  const slot = el.querySelector?.("[data-amarra-theme-label]");
  if (slot) {
    slot.textContent = label;
    return;
  }
  // textContent on the button would drop icon spans/SVGs (#41).
  if (el.children?.length) return;
  el.textContent = label;
}
