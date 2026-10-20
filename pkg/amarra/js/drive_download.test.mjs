import test from "node:test";
import assert from "node:assert/strict";
import {
  downloadBlobResponse,
  filenameFromDisposition,
  isMutationMethod,
} from "./drive_download.mjs";

test("isMutationMethod separates body-carrying verbs from GET/HEAD", () => {
  assert.equal(isMutationMethod("GET"), false);
  assert.equal(isMutationMethod("get"), false);
  assert.equal(isMutationMethod("HEAD"), false);
  assert.equal(isMutationMethod("POST"), true);
  assert.equal(isMutationMethod("PUT"), true);
  assert.equal(isMutationMethod("PATCH"), true);
  assert.equal(isMutationMethod("DELETE"), true);
  assert.equal(isMutationMethod(undefined), false);
});

test("filenameFromDisposition reads quoted and extended filenames", () => {
  assert.equal(filenameFromDisposition('attachment; filename="memorial.pdf"'), "memorial.pdf");
  assert.equal(filenameFromDisposition("attachment; filename=memorial.pdf"), "memorial.pdf");
  assert.equal(
    filenameFromDisposition("attachment; filename*=UTF-8''dossi%C3%AA.pdf"),
    "dossiê.pdf"
  );
});

test("filenameFromDisposition falls back to the URL and sanitizes separators", () => {
  assert.equal(filenameFromDisposition("", "http://a/projetos/1/pdf/memorial"), "memorial");
  assert.equal(filenameFromDisposition("", "http://a/export?filter=paid"), "export");
  assert.equal(filenameFromDisposition("", "http://a/"), "download");
  assert.equal(filenameFromDisposition('attachment; filename="../etc/passwd"'), "_etc_passwd");
});

test("downloadBlobResponse clicks an anchor, names the file and revokes the URL", async () => {
  const blob = { size: 3 };
  const order = [];
  const anchor = {
    download: "",
    href: "",
    appended: false,
    removed: false,
    click() {
      order.push(`click:${this.href}:${this.download}`);
    },
    remove() {
      this.removed = true;
    },
  };
  const doc = {
    createElement(tag) {
      assert.equal(tag, "a");
      return anchor;
    },
    body: {
      appendChild(el) {
        el.appended = true;
      },
    },
  };

  const result = await downloadBlobResponse(
    {
      url: "http://a/export",
      headers: {
        get: (name) =>
          name === "content-disposition" ? 'attachment; filename="filtro.pdf"' : null,
      },
      blob: async () => {
        order.push("blob");
        return blob;
      },
    },
    {
      document: doc,
      createObjectURL(candidate) {
        assert.equal(candidate, blob);
        order.push("create");
        return "blob:1";
      },
      revokeObjectURL(url) {
        order.push(`revoke:${url}`);
      },
      schedule(fn) {
        order.push("schedule");
        fn();
      },
    }
  );

  assert.equal(result.filename, "filtro.pdf");
  assert.equal(anchor.href, "blob:1");
  assert.equal(anchor.download, "filtro.pdf");
  assert.equal(anchor.appended, true);
  assert.equal(anchor.removed, true);
  assert.deepEqual(order, [
    "blob",
    "create",
    "click:blob:1:filtro.pdf",
    "schedule",
    "revoke:blob:1",
  ]);
});
