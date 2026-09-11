const CLICK = "_amarraThemeClick";
const DEFAULT_KEY = "amarra-theme";
const DEFAULT_CLASS = "light";

export function makeTheme(opts = {}) {
  const getHtml = opts.html ?? (() => globalThis.document?.documentElement);
  const getStorage = () => opts.storage ?? globalThis.localStorage;
  const getMeta =
    opts.themeColorMeta ?? (() => globalThis.document?.querySelector?.('meta[name="theme-color"]'));
  // #30: key/class are per-element so apps can rebrand the hook from HTML
  // (data-amarra-theme-key / data-amarra-theme-class) without re-registering.
  // The FOUC snippet in the layout head must use the same key.
  const keyFor = (el) => el?.getAttribute?.("data-amarra-theme-key") || opts.key || DEFAULT_KEY;
  const classFor = (el) =>
    el?.getAttribute?.("data-amarra-theme-class") || opts.className || DEFAULT_CLASS;
  function apply(on, el) {
    const className = classFor(el);
    const key = keyFor(el);
    const html = getHtml();
    if (html?.classList) {
      if (on) html.classList.add(className);
      else html.classList.remove(className);
    }
    try {
      getStorage()?.setItem?.(key, on ? className : "");
    } catch {
      // private mode / blocked storage
    }
    const meta = getMeta?.();
    const lightColor = el?.getAttribute?.("data-amarra-theme-color") || opts.lightColor;
    const darkColor = el?.getAttribute?.("data-amarra-theme-color-off") || opts.darkColor;
    const color = on ? lightColor : darkColor;
    if (meta && color) meta.setAttribute?.("content", color);
    // #30: swap the visible label (on = theme applied) and keep aria-pressed.
    const onLabel = el?.getAttribute?.("data-amarra-theme-on-label") || opts.onLabel;
    const offLabel = el?.getAttribute?.("data-amarra-theme-off-label") || opts.offLabel;
    const label = on ? onLabel : offLabel;
    if (label && el) swapThemeLabel(el, label);
    el?.setAttribute?.("aria-pressed", on ? "true" : "false");
  }

  return {
    connect(el) {
      if (!el || typeof el.addEventListener !== "function") return;
      let stored = "";
      try {
        stored = getStorage()?.getItem?.(keyFor(el)) ?? "";
      } catch {
        stored = "";
      }
      if (stored === classFor(el)) apply(true, el);
      const fn = (ev) => {
        ev?.preventDefault?.();
        const html = getHtml();
        const on = !html?.classList?.contains?.(classFor(el));
        apply(on, el);
      };
      el[CLICK] = fn;
      el.addEventListener("click", fn);
    },
    disconnect(el) {
      const fn = el?.[CLICK];
      if (!fn || typeof el.removeEventListener !== "function") return;
      el.removeEventListener("click", fn);
      delete el[CLICK];
    },
  };
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
