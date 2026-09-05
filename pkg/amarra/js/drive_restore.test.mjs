import test from "node:test";
import assert from "node:assert/strict";
import { applyDriveResponse } from "./drive.mjs";
import { captureScroll, restoreScroll } from "./drive_restore.mjs";

test("applyDriveResponse focuses first aria-invalid field on 422", () => {
  let focused = false;
  const field = {
    focus() {
      focused = true;
    },
  };
  const main = { innerHTML: "old" };
  const doc = {
    querySelector(sel) {
      if (sel === "#amarra-main") return main;
      if (sel.includes("aria-invalid")) return field;
      return null;
    },
  };
  applyDriveResponse({
    status: 422,
    html: `<main id="amarra-main"><input aria-invalid="true"></main>`,
    main,
    document: doc,
    morphFn: (el, html) => {
      el.innerHTML = html;
    },
    history: { pushState() {} },
  });
  assert.equal(focused, true);
});

test("captureScroll stores scrollY on the current history entry", () => {
  const calls = [];
  const history = {
    state: { amarra: true },
    replaceState(state) {
      calls.push(state);
    },
  };
  captureScroll(history, 140);
  assert.equal(calls[0].scrollY, 140);
  assert.equal(calls[0].amarra, true);
});

test("restoreScroll scrolls to saved scrollY", () => {
  const to = [];
  restoreScroll(
    {
      scrollTo(_x, y) {
        to.push(y);
      },
    },
    { scrollY: 90 }
  );
  assert.deepEqual(to, [90]);
});
