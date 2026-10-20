// #366: files returned by a mutation must be saved from the response body.
// Assigning location discarded the POST result and made the browser repeat the
// request as GET, losing filters/body or receiving 405.

export function isMutationMethod(method) {
  const verb = String(method ?? "GET").toUpperCase();
  return verb !== "GET" && verb !== "HEAD";
}

export function filenameFromDisposition(disposition, requestURL = "") {
  const raw = String(disposition ?? "");
  const extended = raw.match(/filename\*\s*=\s*([^;]+)/i);
  if (extended) return sanitizeFilename(decodeFilename(extendedFilename(extended[1])));
  const quoted = raw.match(/filename\s*=\s*"([^"]*)"/i);
  if (quoted) return sanitizeFilename(decodeFilename(quoted[1]));
  const plain = raw.match(/filename\s*=\s*([^;]+)/i);
  if (plain) return sanitizeFilename(decodeFilename(plain[1].trim()));
  return filenameFromURL(requestURL);
}

// RFC 5987: filename*=UTF-8''dossi%C3%AA.pdf. Tolerate a missing charset.
function extendedFilename(value) {
  const name = value.trim().replace(/^"|"$/g, "");
  const at = name.indexOf("''");
  return at >= 0 ? name.slice(at + 2) : name;
}

function decodeFilename(value) {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

function filenameFromURL(url) {
  try {
    const path = new URL(String(url || ""), "http://amarra.invalid").pathname;
    return sanitizeFilename(path.slice(path.lastIndexOf("/") + 1));
  } catch {
    return "download";
  }
}

function sanitizeFilename(value) {
  const clean = String(value ?? "")
    .replace(/[\u0000-\u001f\\/]+/g, "_")
    .replace(/^\.+/, "")
    .trim();
  return clean || "download";
}

export async function downloadBlobResponse(res, opts = {}) {
  const blob = await res.blob();
  // Reading the body can outlive a newer visit (#348): do not start a stale file.
  if (typeof opts.isCurrent === "function" && !opts.isCurrent()) return null;
  const filename = filenameFromDisposition(
    res.headers?.get?.("content-disposition"),
    res.url || opts.url
  );
  const createObjectURL =
    opts.createObjectURL ?? globalThis.URL?.createObjectURL?.bind(globalThis.URL);
  const revokeObjectURL =
    opts.revokeObjectURL ?? globalThis.URL?.revokeObjectURL?.bind(globalThis.URL);
  if (typeof createObjectURL !== "function") {
    throw new Error("download: URL.createObjectURL is unavailable");
  }
  const objectURL = createObjectURL(blob);
  const anchor = downloadAnchor(opts.document, objectURL, filename);
  anchor.click?.();
  anchor.remove?.();
  scheduleRevoke(objectURL, revokeObjectURL, opts.schedule);
  return { filename };
}

function downloadAnchor(doc, objectURL, filename) {
  if (!doc?.createElement) throw new Error("download: document.createElement is unavailable");
  const anchor = doc.createElement("a");
  anchor.href = objectURL;
  anchor.download = filename;
  anchor.rel = "noopener";
  anchor.hidden = true;
  (doc.body ?? doc.documentElement)?.appendChild?.(anchor);
  return anchor;
}

function scheduleRevoke(objectURL, revokeObjectURL, schedule) {
  if (typeof revokeObjectURL !== "function") return;
  const later = schedule ?? globalThis.setTimeout?.bind(globalThis);
  const revoke = () => revokeObjectURL(objectURL);
  if (typeof later === "function") later(revoke, 0);
  else revoke();
}
