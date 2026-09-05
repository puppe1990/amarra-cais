export function confirmOk(el, confirmFn) {
  const msg = el?.getAttribute?.("data-amarra-confirm");
  if (!msg) return true;
  const fn =
    confirmFn ??
    (typeof globalThis.confirm === "function" ? globalThis.confirm.bind(globalThis) : () => true);
  return !!fn(msg);
}

export function requestMethod(el, fallback = "GET") {
  const attr = el?.getAttribute?.("data-amarra-method");
  if (attr) return String(attr).toUpperCase();
  const hidden = el?.querySelector?.('input[name="_method"]');
  if (hidden?.value) return String(hidden.value).toUpperCase();
  return String(fallback || "GET").toUpperCase();
}

export function disableSubmit(el) {
  if (!el) return null;
  const prev = { el, disabled: !!el.disabled, text: el.textContent };
  el.disabled = true;
  const withText = el.getAttribute?.("data-amarra-disable-with");
  if (withText) el.textContent = withText;
  return prev;
}

export function restoreSubmit(prev) {
  if (!prev?.el) return;
  prev.el.disabled = prev.disabled;
  if (prev.text != null) prev.el.textContent = prev.text;
}
