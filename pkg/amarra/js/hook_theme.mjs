const CLICK = "_amarraThemeClick";
const DEFAULT_KEY = "amarra-theme";
const DEFAULT_CLASS = "light";

export function makeTheme(opts = {}) {
  const className = opts.className ?? DEFAULT_CLASS;
  const key = opts.key ?? DEFAULT_KEY;
  const getHtml = opts.html ?? (() => globalThis.document?.documentElement);
  const getStorage = () => opts.storage ?? globalThis.localStorage;
  const getMeta =
    opts.themeColorMeta ?? (() => globalThis.document?.querySelector?.('meta[name="theme-color"]'));
  function apply(on, el) {
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
  }

  return {
    connect(el) {
      if (!el || typeof el.addEventListener !== "function") return;
      let stored = "";
      try {
        stored = getStorage()?.getItem?.(key) ?? "";
      } catch {
        stored = "";
      }
      if (stored === className) apply(true, el);
      const fn = (ev) => {
        ev?.preventDefault?.();
        const html = getHtml();
        const on = !html?.classList?.contains?.(className);
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
