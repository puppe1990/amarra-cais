import test from "node:test";
import assert from "node:assert/strict";
import { makeDialog } from "./hook_dialog.mjs";

function openerButton(label = "Open") {
  const listeners = {};
  return {
    label,
    listeners,
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
    click() {
      (listeners.click || []).forEach((fn) => fn({ preventDefault() {} }));
    },
    focus() {
      this.focused = true;
    },
  };
}

function dialogEl() {
  const listeners = {};
  const attrs = {};
  return {
    listeners,
    attrs,
    open: false,
    showModal() {
      this.open = true;
    },
    close() {
      if (!this.open) return;
      this.open = false;
      (listeners.close || []).forEach((fn) => fn({ preventDefault() {} }));
    },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
    focus() {
      this.focused = true;
    },
    getAttribute(n) {
      return Object.hasOwn(attrs, n) ? attrs[n] : null;
    },
    setAttribute(n, v) {
      attrs[n] = String(v);
    },
  };
}

function container(openers, dialogs, closers = []) {
  return {
    querySelector(sel) {
      return sel === "[data-amarra-dialog-target]" ? (dialogs[0] ?? null) : null;
    },
    querySelectorAll(sel) {
      if (sel === "[data-amarra-dialog-open]") return [...openers];
      if (sel === "[data-amarra-dialog-close]") return [...closers];
      return [];
    },
  };
}

test("dialog hook opens the native dialog and returns focus to the opener on close", () => {
  const open = openerButton();
  const dlg = dialogEl();
  const hook = makeDialog();
  const el = container([open], [dlg]);
  hook.connect(el);
  assert.equal(dlg.attrs["aria-modal"], "true");

  open.click();
  assert.equal(dlg.open, true);

  dlg.close(); // native Esc / form method=dialog / programmatic all land here
  assert.equal(open.focused, true);

  hook.disconnect(el);
  assert.equal((open.listeners.click || []).length, 0);
  assert.equal((dlg.listeners.close || []).length, 0);
});

test("dialog hook closes via data-amarra-dialog-close buttons", () => {
  const open = openerButton();
  const dlg = dialogEl();
  const close = openerButton("Yes");
  const hook = makeDialog();
  hook.connect(container([open], [dlg], [close]));

  open.click();
  assert.equal(dlg.open, true);
  close.click();
  assert.equal(dlg.open, false);
});

test("dialog hook ignores containers without a dialog target", () => {
  const hook = makeDialog();
  assert.doesNotThrow(() => hook.connect(container([openerButton()], [])));
  assert.doesNotThrow(() => hook.connect(container([openerButton()], [{ showModal: undefined }])));
});

test("dialog hook keeps an existing aria-modal value", () => {
  const dlg = dialogEl();
  dlg.attrs["aria-modal"] = "true";
  makeDialog().connect(container([], [dlg]));
  assert.equal(dlg.attrs["aria-modal"], "true");
});
