import test from "node:test";
import assert from "node:assert/strict";
import { makeBulk } from "./hook_bulk.mjs";

function checkbox(checked = false) {
  return {
    checked,
    indeterminate: false,
    listeners: {},
    addEventListener(type, fn) {
      (this.listeners[type] ??= []).push(fn);
    },
    removeEventListener(type, fn) {
      this.listeners[type] = (this.listeners[type] || []).filter((f) => f !== fn);
    },
    change() {
      (this.listeners.change || []).forEach((fn) => fn({}));
    },
  };
}

function bar() {
  return {
    hidden: true,
    count: { textContent: "" },
    querySelector(sel) {
      return sel === "[data-amarra-bulk-count]" ? this.count : null;
    },
  };
}

function container(rows, barEl) {
  return {
    rows,
    barEl,
    querySelector(sel) {
      if (sel === "[data-amarra-bulk-all]") return this.all;
      if (sel === "[data-amarra-bulk-bar]") return barEl ?? null;
      return null;
    },
    querySelectorAll(sel) {
      return sel === "[data-amarra-bulk-row]" ? [...rows] : [];
    },
  };
}

function setup(rowCount = 2, withBar = true) {
  const all = checkbox();
  const rows = Array.from({ length: rowCount }, () => checkbox());
  const barEl = withBar ? bar() : null;
  const el = container(rows, barEl);
  el.all = all;
  makeBulk().connect(el);
  return { all, rows, barEl, el };
}

test("bulk hook select-all checks and unchecks every row on the page", () => {
  const { all, rows } = setup();
  all.checked = true;
  all.change();
  assert.deepEqual(
    rows.map((r) => r.checked),
    [true, true]
  );

  all.checked = false;
  all.change();
  assert.deepEqual(
    rows.map((r) => r.checked),
    [false, false]
  );
});

test("bulk hook marks the header indeterminate when some rows are selected", () => {
  const { all, rows } = setup();
  rows[0].checked = true;
  rows[0].change();
  assert.equal(all.indeterminate, true);
  assert.equal(all.checked, false);

  rows[1].checked = true;
  rows[1].change();
  assert.equal(all.indeterminate, false);
  assert.equal(all.checked, true);
});

test("bulk hook shows the bar with the selected count when count > 0", () => {
  const { rows, barEl } = setup(3);
  rows[0].checked = true;
  rows[1].checked = true;
  rows[1].change();
  assert.equal(barEl.hidden, false);
  assert.equal(barEl.count.textContent, "2");

  rows[0].checked = false;
  rows[0].change();
  rows[1].checked = false;
  rows[1].change();
  assert.equal(barEl.hidden, true);
  assert.equal(barEl.count.textContent, "0");
});

test("bulk hook works without a bar and unbinds on disconnect", () => {
  const { all, rows, el } = setup(1, false);
  all.checked = true;
  all.change();
  assert.equal(rows[0].checked, true);

  makeBulk().disconnect(el);
  rows[0].checked = false;
  rows[0].change();
  // header no longer reacts after disconnect
  assert.equal(all.checked, true);
});
