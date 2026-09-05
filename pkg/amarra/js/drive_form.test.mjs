import test from "node:test";
import assert from "node:assert/strict";
import { confirmOk, requestMethod, disableSubmit, restoreSubmit } from "./drive_form.mjs";

test("confirmOk is true without data-amarra-confirm", () => {
  assert.equal(
    confirmOk({
      getAttribute() {
        return null;
      },
    }),
    true
  );
});

test("confirmOk uses confirm callback for data-amarra-confirm", () => {
  const el = {
    getAttribute(n) {
      return n === "data-amarra-confirm" ? "Delete?" : null;
    },
  };
  assert.equal(
    confirmOk(el, () => false),
    false
  );
  assert.equal(
    confirmOk(el, () => true),
    true
  );
});

test("requestMethod reads data-amarra-method then _method", () => {
  assert.equal(
    requestMethod({
      getAttribute(n) {
        return n === "data-amarra-method" ? "delete" : null;
      },
    }),
    "DELETE"
  );
  assert.equal(
    requestMethod({
      getAttribute() {
        return null;
      },
      querySelector(sel) {
        return sel === 'input[name="_method"]' ? { value: "patch" } : null;
      },
    }),
    "PATCH"
  );
  assert.equal(
    requestMethod(
      {
        getAttribute() {
          return null;
        },
        querySelector() {
          return null;
        },
      },
      "post"
    ),
    "POST"
  );
});

test("disableSubmit swaps label and restoreSubmit reverts", () => {
  const btn = {
    disabled: false,
    textContent: "Save",
    getAttribute(n) {
      return n === "data-amarra-disable-with" ? "Saving..." : null;
    },
  };
  const prev = disableSubmit(btn);
  assert.equal(btn.disabled, true);
  assert.equal(btn.textContent, "Saving...");
  restoreSubmit(prev);
  assert.equal(btn.disabled, false);
  assert.equal(btn.textContent, "Save");
});
