import test from "node:test";
import assert from "node:assert/strict";
import { makeDialog } from "./hook_dialog.mjs";

function openerButton(label = "Open", attrs = {}) {
  const listeners = {};
  return {
    label,
    listeners,
    attrs: { ...attrs },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
    getAttribute(n) {
      return Object.hasOwn(this.attrs, n) ? this.attrs[n] : null;
    },
    click(ev) {
      const event = ev ?? { preventDefault() {} };
      (listeners.click || []).forEach((fn) => fn(event));
    },
    focus() {
      this.focused = true;
    },
  };
}

function frameEl() {
  const attrs = {};
  return {
    attrs,
    tagName: "AMARRA-FRAME",
    getAttribute(n) {
      return Object.hasOwn(attrs, n) ? attrs[n] : null;
    },
    setAttribute(n, v) {
      attrs[n] = String(v);
    },
  };
}

function dialogEl(opts = {}) {
  const listeners = {};
  const attrs = {};
  const frame = opts.frame ?? null;
  const body = opts.body ?? null;
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
    querySelector(sel) {
      if (sel === "amarra-frame") return frame;
      if (sel === "[data-amarra-dialog-body]") return body;
      return null;
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

// #126: after a partial morph the old buttons stayed bound and the new ones
// never opened the dialog.
test("dialog hook updated rebinds buttons replaced by a morph", () => {
  const hook = makeDialog();
  const staleOpen = openerButton("stale");
  const staleDlg = dialogEl();
  const el = {
    openers: [staleOpen],
    dialogs: [staleDlg],
    querySelector(sel) {
      return sel === "[data-amarra-dialog-target]" ? (this.dialogs[0] ?? null) : null;
    },
    querySelectorAll(sel) {
      return sel === "[data-amarra-dialog-open]" ? [...this.openers] : [];
    },
  };
  hook.connect(el);

  const freshOpen = openerButton("fresh");
  const freshDlg = dialogEl();
  el.openers = [freshOpen];
  el.dialogs = [freshDlg];
  hook.updated(el);

  freshOpen.click();
  assert.equal(freshDlg.open, true, "fresh opener must open the fresh dialog");
  assert.equal(staleDlg.open, false, "stale dialog must stay closed");
  assert.equal((staleDlg.listeners.close || []).length, 0, "stale close listener removed");
});

// #392: opener href / data-amarra-dialog-src loads into the inner amarra-frame.
test("dialog hook loads opener href into the inner amarra-frame", () => {
  const frame = frameEl();
  const dlg = dialogEl({ frame });
  const open = openerButton("Preview", { href: "/books/1/chapters/6/preview" });
  makeDialog().connect(container([open], [dlg]));

  let prevented = false;
  open.click({
    preventDefault() {
      prevented = true;
    },
  });

  assert.equal(prevented, true, "preventDefault must run before Drive intercepts");
  assert.equal(dlg.open, true);
  assert.equal(frame.attrs.src, "/books/1/chapters/6/preview");
});

test("dialog hook prefers data-amarra-dialog-src over href", () => {
  const frame = frameEl();
  const dlg = dialogEl({ frame });
  const open = openerButton("Preview", {
    href: "/full-page",
    "data-amarra-dialog-src": "/books/1/preview",
  });
  makeDialog().connect(container([open], [dlg]));
  open.click();
  assert.equal(frame.attrs.src, "/books/1/preview");
});

test("dialog hook without href only showModal", () => {
  const frame = frameEl();
  frame.attrs.src = "stale";
  const dlg = dialogEl({ frame });
  const open = openerButton();
  makeDialog().connect(container([open], [dlg]));
  open.click();
  assert.equal(dlg.open, true);
  assert.equal(frame.attrs.src, "stale");
});

test("dialog hook second opener swaps the frame src", () => {
  const frame = frameEl();
  const dlg = dialogEl({ frame });
  const first = openerButton("06", { href: "/chapters/6/preview" });
  const second = openerButton("08", { href: "/chapters/8/preview" });
  makeDialog().connect(container([first, second], [dlg]));
  first.click();
  assert.equal(frame.attrs.src, "/chapters/6/preview");
  second.click();
  assert.equal(frame.attrs.src, "/chapters/8/preview");
});

test("dialog hook loads src onto data-amarra-dialog-body when no frame", () => {
  const body = frameEl();
  body.tagName = "DIV";
  const dlg = dialogEl({ body });
  const open = openerButton("Preview", { href: "/preview" });
  makeDialog().connect(container([open], [dlg]));
  open.click();
  assert.equal(body.attrs.src, "/preview");
});
