const BTN = "_amarraDropdownToggle";
const MENU = "_amarraDropdownMenuClick";
const DOC_CLICK = "_amarraDropdownDocClick";
const DOC_KEY = "_amarraDropdownDocKey";

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
      el[DOC_CLICK] = onDocClick;
      doc?.addEventListener?.("click", onDocClick);
      el[DOC_KEY] = onKey;
      doc?.addEventListener?.("keydown", onKey);
    },
    disconnect(el) {
      if (!el?.querySelector) return;
      const btn = el.querySelector("[data-amarra-dropdown-button]");
      const menu = el.querySelector("[data-amarra-dropdown-menu]");
      if (!btn || !menu) return;
      const doc = el.ownerDocument ?? globalThis.document;
      const toggle = btn[BTN];
      if (toggle) {
        btn.removeEventListener?.("click", toggle);
        delete btn[BTN];
      }
      const menuClose = menu[MENU];
      if (menuClose) {
        menu.removeEventListener?.("click", menuClose);
        delete menu[MENU];
      }
      const onDocClick = el[DOC_CLICK];
      if (onDocClick) {
        doc?.removeEventListener?.("click", onDocClick);
        delete el[DOC_CLICK];
      }
      const onKey = el[DOC_KEY];
      if (onKey) {
        doc?.removeEventListener?.("keydown", onKey);
        delete el[DOC_KEY];
      }
    },
  };
}

export const dropdown = makeDropdown();
