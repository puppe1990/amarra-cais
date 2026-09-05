import test from "node:test";
import assert from "node:assert/strict";
import { makeReveal } from "./hook_reveal.mjs";

function classList(initial = []) {
  const set = new Set(initial);
  return {
    contains: (c) => set.has(c),
    add: (c) => set.add(c),
    remove: (c) => set.delete(c),
    toggle(c, force) {
      if (force === true) set.add(c);
      else if (force === false) set.delete(c);
      else if (set.has(c)) set.delete(c);
      else set.add(c);
    },
  };
}

function control(value, attrs = {}) {
  const listeners = {};
  const all = {
    "data-amarra-reveal-show": "access_keys",
    "data-amarra-reveal-target": "#aws-keys",
    ...attrs,
  };
  return {
    value,
    listeners,
    hidden: false,
    getAttribute(n) {
      return Object.hasOwn(all, n) ? all[n] : null;
    },
    addEventListener(type, fn) {
      (listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      listeners[type] = (listeners[type] || []).filter((f) => f !== fn);
    },
  };
}

test("reveal hook shows the target when the control value matches and hides otherwise", () => {
  const target = { hidden: true, classList: classList(["hidden"]) };
  const hook = makeReveal((sel) => (sel === "#aws-keys" ? target : null));
  const el = control("default_chain");
  hook.connect(el);
  assert.equal(target.hidden, true);
  el.value = "access_keys";
  el.listeners.change[0]();
  assert.equal(target.hidden, false);
  el.value = "default_chain";
  el.listeners.change[0]();
  assert.equal(target.hidden, true);
  hook.disconnect(el);
  assert.equal((el.listeners.change || []).length, 0);
  assert.equal((el.listeners.click || []).length, 0);
});

test("reveal hook applies the current value on connect", () => {
  const target = { hidden: true };
  const hook = makeReveal((sel) => (sel === "#aws-keys" ? target : null));
  hook.connect(control("access_keys"));
  assert.equal(target.hidden, false);
});
