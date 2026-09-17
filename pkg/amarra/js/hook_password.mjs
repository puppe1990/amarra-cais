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
        swapAriaLabel(el, show);
        toggleIcons(el, show);
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
  if (sel) {
    const root = el?.ownerDocument ?? globalThis.document;
    try {
      const found = root?.querySelector?.(sel);
      if (found) return found;
    } catch {
      return null;
    }
  }
  // #28/#132: auth pages wrap <input> + toggle in one relative element. The
  // old fallback took the first input of the parent — with several fields it
  // toggled the wrong one (e.g. the email field). Prefer a password field in
  // the same form (or wrapper) and never guess a non-password input.
  const scope = el?.closest?.("form") ?? el?.parentElement ?? null;
  const inputs = scope?.querySelectorAll?.('input[type="password"]') ?? [];
  if (inputs.length === 0) return null;
  if (inputs.length === 1) return inputs[0];
  // Several password fields (password + confirmation): pick the closest one
  // preceding the toggle when the DOM API is available.
  let best = inputs[0];
  for (const input of inputs) {
    if (el?.compareDocumentPosition?.(input) & 2) best = input;
  }
  return best;
}

// #28: aria-label follows the visible affordance (the next action), like the
// icon spans: hidden input shows the "show" label, visible input the "hide".
function swapAriaLabel(el, show) {
  const showLabel = el.getAttribute?.("data-amarra-label-show");
  const hideLabel = el.getAttribute?.("data-amarra-label-hide");
  if (!showLabel && !hideLabel) return;
  el.setAttribute?.("aria-label", show ? hideLabel || showLabel : showLabel || hideLabel);
}

function toggleIcons(el, show) {
  // #28: data-amarra-password-icon is the amarra name; data-cais-password-icon
  // stays as the legacy alias so existing markup keeps working.
  const showIcon =
    el.querySelector?.('[data-amarra-password-icon="show"]') ??
    el.querySelector?.('[data-cais-password-icon="show"]');
  const hideIcon =
    el.querySelector?.('[data-amarra-password-icon="hide"]') ??
    el.querySelector?.('[data-cais-password-icon="hide"]');
  showIcon?.classList?.toggle?.("hidden", show);
  hideIcon?.classList?.toggle?.("hidden", !show);
}

export const password = makePassword();
