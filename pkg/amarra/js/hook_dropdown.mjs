const BTN = "_amarraDropdownToggle";
const MENU = "_amarraDropdownMenuClick";
const STATE = "_amarraDropdownState";

// #22: row actions and overflow menus. Click toggles, Esc and outside clicks
// close, clicks inside the menu close after acting. Lifecycle rides the hook
// registry so Drive morphs rebind instead of leaking document listeners.
export function makeDropdown() {
  return {
    connect(el) {
      if (!el?.querySelector) return;
      const btn = el.querySelector("[data-amarra-dropdown-button]");
      const menu = el.querySelector("[data-amarra-dropdown-menu]");
      if (!btn || !menu) return;
      const doc = el.ownerDocument ?? globalThis.document;

      const close = () => {
        menu.hidden = true;
        btn.setAttribute?.("aria-expanded", "false");
      };
      const toggle = (ev) => {
        ev?.preventDefault?.();
        const open = menu.hidden;
        menu.hidden = !open;
        btn.setAttribute?.("aria-expanded", String(open));
      };
      const onDocClick = (ev) => {
        if (el.contains?.(ev?.target)) return;
        close();
      };
      const onKey = (ev) => {
        if (ev?.key === "Escape") close();
      };

      btn[BTN] = toggle;
      btn.addEventListener?.("click", toggle);
      menu[MENU] = close;
      menu.addEventListener?.("click", close);
      doc?.addEventListener?.("click", onDocClick);
      doc?.addEventListener?.("keydown", onKey);

      // Bound nodes live on the container so disconnect unbinds exactly what
      // this connect bound, even after a morph replaced the children.
      el[STATE] = { btn, menu, doc, onDocClick, onKey };
    },
    // #126: Idiomorph keeps the container and swaps children; rebind so the
    // hook tracks the new nodes and document listeners do not accumulate.
    updated(el) {
      dropdown.disconnect(el);
      dropdown.connect(el);
    },
    disconnect(el) {
      const st = el?.[STATE];
      if (!st) return;
      const toggle = st.btn?.[BTN];
      if (toggle) {
        st.btn.removeEventListener?.("click", toggle);
        delete st.btn[BTN];
      }
      const menuClose = st.menu?.[MENU];
      if (menuClose) {
        st.menu.removeEventListener?.("click", menuClose);
        delete st.menu[MENU];
      }
      if (st.onDocClick) st.doc?.removeEventListener?.("click", st.onDocClick);
      if (st.onKey) st.doc?.removeEventListener?.("keydown", st.onKey);
      delete el[STATE];
    },
  };
}

export const dropdown = makeDropdown();
