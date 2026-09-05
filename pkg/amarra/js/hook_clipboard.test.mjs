import test from "node:test";
import assert from "node:assert/strict";
import { makeClipboard } from "./hook_clipboard.mjs";

test("clipboard hook copies data-amarra-copy on click and unbinds on disconnect", () => {
  const copied = [];
  const hook = makeClipboard((text) => copied.push(text));
  const listeners = {};
  const el = {
    getAttribute(n) {
      return n === "data-amarra-copy" ? "secret" : null;
    },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
  };
  hook.connect(el);
  listeners.click[0]();
  assert.deepEqual(copied, ["secret"]);
  hook.disconnect(el);
  assert.equal((listeners.click || []).length, 0);
});
