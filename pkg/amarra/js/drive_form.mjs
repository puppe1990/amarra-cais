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

// Native GET replaces the action's query with the form fields. Concatenating
// (`?q=antigo` + `&q=novo`) left Go's Query().Get("q") reading the old value,
// and a `#fragment` swallowed the serialized fields (#361).
export function withQuery(action, fd) {
  if (!fd || typeof fd.entries !== "function") return action;
  const params = new URLSearchParams();
  for (const [key, value] of fd.entries()) {
    if (typeof value === "string") params.append(key, value);
  }
  const raw = String(action ?? "");
  try {
    const u = new URL(raw, "http://amarra.invalid");
    u.search = params.toString();
    if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(raw)) {
      return u.toString();
    }
    return u.pathname + u.search + u.hash;
  } catch {
    const q = params.toString();
    if (!q) return raw;
    const hashAt = raw.indexOf("#");
    const before = hashAt === -1 ? raw : raw.slice(0, hashAt);
    const hash = hashAt === -1 ? "" : raw.slice(hashAt);
    const base = before.split("?")[0];
    return `${base}?${q}${hash}`;
  }
}

export function driveFormBody(
  formData,
  URLSearchParamsCtor = URLSearchParams,
  FileCtor = typeof File !== "undefined" ? File : null
) {
  if (!formData) return null;
  for (const [, value] of formData.entries()) {
    if (FileCtor && value instanceof FileCtor) return formData;
  }
  const body = new URLSearchParamsCtor();
  for (const [key, value] of formData.entries()) body.append(key, value);
  return body;
}
