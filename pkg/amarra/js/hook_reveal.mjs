const SYNC = "_amarraRevealSync";

export function makeReveal(findTarget) {
  const resolve = findTarget ?? defaultFind;
  return {
    connect(el) {
      if (!el || typeof el.addEventListener !== "function") return;
      const fn = () => {
        const match = el.getAttribute?.("data-amarra-reveal-show") ?? "";
        const sel = el.getAttribute?.("data-amarra-reveal-target") ?? "";
        const target = resolve(sel, el);
        if (!target) return;
        target.hidden = el.value !== match;
      };
      el[SYNC] = fn;
      el.addEventListener("change", fn);
      el.addEventListener("click", fn);
      fn();
    },
    disconnect(el) {
      const fn = el?.[SYNC];
      if (!fn || typeof el.removeEventListener !== "function") return;
      el.removeEventListener("change", fn);
      el.removeEventListener("click", fn);
      delete el[SYNC];
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

export const reveal = makeReveal();
