const OPEN_ATTR = "data-amarra-sidebar-open";
const STATE = "_amarraSidebar";

// #212: below lg the rail is an off-canvas drawer. The old toggle was a 10px
// label with no state; this hook gives the button aria-expanded, Esc/outside
// close and a single source of truth for the panel's `data-amarra-sidebar-open`.
export function makeSidebar(deps = {}) {
  const isDesktop = deps.isDesktop ?? (() => false);

  return {
    connect(el) {
      if (!el?.getAttribute) return;
      const doc = el.ownerDocument ?? deps.document ?? null;
      const sel = el.getAttribute("data-amarra-sidebar-target") || "";
      const panel = sel && doc?.querySelector ? doc.querySelector(sel) : null;
      if (!panel) return;

      const state = {};
      const apply = (next) => {
        state.open = next;
        if (next) panel.setAttribute?.(OPEN_ATTR, "");
        else panel.removeAttribute?.(OPEN_ATTR);
        el.setAttribute?.("aria-expanded", next ? "true" : "false");
      };
      const onToggle = () => apply(!state.open);
      const onKey = (ev) => {
        if (ev?.key !== "Escape" || !state.open) return;
        apply(false);
        el.focus?.();
      };
      const onDocClick = (ev) => {
        if (!state.open || isDesktop()) return;
        const target = ev?.target;
        if (el.contains?.(target) || panel.contains?.(target)) return;
        apply(false);
      };
      state.onToggle = onToggle;
      state.onKey = onKey;
      state.onDocClick = onDocClick;
      el.addEventListener?.("click", onToggle);
      doc?.addEventListener?.("keydown", onKey);
      doc?.addEventListener?.("click", onDocClick, true);
      apply(false);
      el[STATE] = state;
    },
    disconnect(el) {
      const state = el?.[STATE];
      if (!state) return;
      el.removeEventListener?.("click", state.onToggle);
      const doc = el.ownerDocument ?? deps.document ?? null;
      doc?.removeEventListener?.("keydown", state.onKey);
      doc?.removeEventListener?.("click", state.onDocClick, true);
      delete el[STATE];
    },
  };
}

export const sidebar = makeSidebar();
