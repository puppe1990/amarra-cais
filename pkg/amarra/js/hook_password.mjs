const CLICK = "_amarraPasswordClick";

export function makePassword(findInput) {
  const resolve = findInput ?? defaultFind;
  return {
    connect(el) {
      if (!el || typeof el.addEventListener !== "function") return;
      const fn = (ev) => {
        ev?.preventDefault?.();
        const sel = el.getAttribute?.("data-amarra-password-for") ?? "";
        const input = resolve(sel, el);
        if (!input) return;
        const show = input.type === "password";
        input.type = show ? "text" : "password";
        el.setAttribute?.("aria-pressed", show ? "true" : "false");
        const showIcon = el.querySelector?.('[data-cais-password-icon="show"]');
        const hideIcon = el.querySelector?.('[data-cais-password-icon="hide"]');
        showIcon?.classList?.toggle?.("hidden", show);
        hideIcon?.classList?.toggle?.("hidden", !show);
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

function defaultFind(sel, el) {
  if (!sel) return null;
  const root = el?.ownerDocument ?? globalThis.document;
  try {
    return root?.querySelector?.(sel) ?? null;
  } catch {
    return null;
  }
}

export const password = makePassword();
