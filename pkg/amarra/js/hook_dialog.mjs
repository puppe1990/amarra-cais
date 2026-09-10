const OPEN = "_amarraDialogOpen";
const CLOSE = "_amarraDialogClose";
const ONCLOSE = "_amarraDialogOnClose";

// #21: native <dialog> gives Esc, focus trap, and ::backdrop for free — the
// hook only wires open/close buttons and returns focus to the opener.
// Markup: container[amarra-hook=dialog] > button[data-amarra-dialog-open]
// + dialog[data-amarra-dialog-target] (+ optional [data-amarra-dialog-close]).
// form method="dialog" needs no JS.
export function makeDialog() {
  return {
    connect(el) {
      if (!el?.querySelector) return;
      const dlg = el.querySelector("[data-amarra-dialog-target]");
      if (!dlg || typeof dlg.showModal !== "function") return;
      if (!dlg.getAttribute?.("aria-modal")) dlg.setAttribute?.("aria-modal", "true");

      let opener = null;
      for (const btn of el.querySelectorAll?.("[data-amarra-dialog-open]") ?? []) {
        const fn = (ev) => {
          ev?.preventDefault?.();
          opener = btn;
          dlg.showModal?.();
        };
        btn[OPEN] = fn;
        btn.addEventListener?.("click", fn);
      }
      for (const btn of el.querySelectorAll?.("[data-amarra-dialog-close]") ?? []) {
        const fn = (ev) => {
          ev?.preventDefault?.();
          dlg.close?.();
        };
        btn[CLOSE] = fn;
        btn.addEventListener?.("click", fn);
      }
      const onClose = () => opener?.focus?.();
      dlg[ONCLOSE] = onClose;
      dlg.addEventListener?.("close", onClose);
    },
    disconnect(el) {
      if (!el?.querySelector) return;
      for (const btn of el.querySelectorAll?.("[data-amarra-dialog-open]") ?? []) {
        const fn = btn?.[OPEN];
        if (!fn) continue;
        btn.removeEventListener?.("click", fn);
        delete btn[OPEN];
      }
      for (const btn of el.querySelectorAll?.("[data-amarra-dialog-close]") ?? []) {
        const fn = btn?.[CLOSE];
        if (!fn) continue;
        btn.removeEventListener?.("click", fn);
        delete btn[CLOSE];
      }
      const dlg = el.querySelector("[data-amarra-dialog-target]");
      if (!dlg) return;
      const onClose = dlg?.[ONCLOSE];
      if (onClose) {
        dlg.removeEventListener?.("close", onClose);
        delete dlg[ONCLOSE];
      }
    },
  };
}

export const dialog = makeDialog();
