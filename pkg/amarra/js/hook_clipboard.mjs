const CLICK = "_amarraClipboardClick";

export function makeClipboard(writeText) {
  return {
    connect(el) {
      if (!el || typeof el.addEventListener !== "function") return;
      const fn = (ev) => {
        ev?.preventDefault?.();
        const text = el.getAttribute?.("data-amarra-copy") ?? "";
        writeText?.(text);
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

export const clipboard = makeClipboard((text) => {
  const write = globalThis.navigator?.clipboard?.writeText;
  if (typeof write === "function") return write.call(globalThis.navigator.clipboard, text);
});
