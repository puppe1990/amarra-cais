const OPEN = "_amarraDialogOpen";
const CLOSE = "_amarraDialogClose";
const ONCLOSE = "_amarraDialogOnClose";
const STATE = "_amarraDialogState";

// #21: native <dialog> gives Esc, focus trap, and ::backdrop for free — the
// hook wires open/close and returns focus to the opener.
// #392: an opener href / data-amarra-dialog-src is copied onto the inner
// <amarra-frame> (or [data-amarra-dialog-body]); setting src loads the
// fragment via Amarra-Frame. preventDefault runs first so Drive does not
// morph #amarra-main. Static dialogs without href stay as #21.
// Markup: container[amarra-hook=dialog] > a|button[data-amarra-dialog-open]
// + dialog[data-amarra-dialog-target] (+ optional [data-amarra-dialog-close]).
// form method="dialog" needs no JS.
export function makeDialog() {
  return {
    connect(el) {
      if (!el?.querySelector) return;
      const dlg = el.querySelector("[data-amarra-dialog-target]");
      if (!dlg || typeof dlg.showModal !== "function") return;
      if (!dlg.getAttribute?.("aria-modal")) dlg.setAttribute?.("aria-modal", "true");

      const openers = [...(el.querySelectorAll?.("[data-amarra-dialog-open]") ?? [])];
      const closers = [...(el.querySelectorAll?.("[data-amarra-dialog-close]") ?? [])];
      let opener = null;
      for (const btn of openers) {
        const fn = (ev) => {
          ev?.preventDefault?.();
          opener = btn;
          dlg.showModal?.();
          loadDialogSrc(btn, dlg);
        };
        btn[OPEN] = fn;
        btn.addEventListener?.("click", fn);
      }
      for (const btn of closers) {
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

      // Bound nodes live on the container so disconnect can unbind exactly
      // what this connect bound, even after a morph replaced the children.
      el[STATE] = { dlg, openers, closers };
    },
    // #126: Idiomorph keeps the container and swaps children; rebind so the
    // hook tracks the new nodes instead of stale references.
    updated(el) {
      dialog.disconnect(el);
      dialog.connect(el);
    },
    disconnect(el) {
      const st = el?.[STATE];
      if (!st) return;
      for (const btn of st.openers ?? []) {
        const fn = btn?.[OPEN];
        if (!fn) continue;
        btn.removeEventListener?.("click", fn);
        delete btn[OPEN];
      }
      for (const btn of st.closers ?? []) {
        const fn = btn?.[CLOSE];
        if (!fn) continue;
        btn.removeEventListener?.("click", fn);
        delete btn[CLOSE];
      }
      const onClose = st.dlg?.[ONCLOSE];
      if (onClose) {
        st.dlg.removeEventListener?.("close", onClose);
        delete st.dlg[ONCLOSE];
      }
      delete el[STATE];
    },
  };
}

export const dialog = makeDialog();

function loadDialogSrc(btn, dlg) {
  const src =
    btn.getAttribute?.("data-amarra-dialog-src") || btn.getAttribute?.("href") || "";
  if (!src) return;
  const frame =
    dlg.querySelector?.("amarra-frame") || dlg.querySelector?.("[data-amarra-dialog-body]");
  if (frame?.setAttribute) frame.setAttribute("src", src);
}
