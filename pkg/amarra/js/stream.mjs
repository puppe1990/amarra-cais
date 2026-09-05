import { morph } from "./morph.mjs";

const STREAM_KINDS = ["append", "prepend", "replace", "morph", "remove", "toast"];

export function parseSSE(chunk) {
  const events = [];
  const text = String(chunk ?? "").replace(/\r\n/g, "\n");
  for (const block of text.split("\n\n")) {
    if (!block.trim()) continue;
    let kind = "message";
    const dataLines = [];
    for (const line of block.split("\n")) {
      if (!line || line.startsWith(":")) continue;
      if (line.startsWith("event:")) {
        kind = line.slice(6).trim();
        continue;
      }
      if (line.startsWith("data:")) {
        const rest = line.slice(5);
        dataLines.push(rest.startsWith(" ") ? rest.slice(1) : rest);
      }
    }
    events.push({ kind, html: dataLines.join("\n") });
  }
  return events;
}

export function applyOp(op, doc, opts = {}) {
  if (!op) return;
  if (op.kind === "toast") {
    if (doc && typeof doc.dispatchEvent === "function") {
      doc.dispatchEvent(
        new CustomEvent("amarra:toast", { bubbles: true, detail: { message: op.html ?? "" } })
      );
    }
    return;
  }
  const id = op.target || opts.defaultTarget;
  const el = id && doc?.getElementById ? doc.getElementById(id) : null;
  if (!el) return;
  const html = op.html ?? "";
  switch (op.kind) {
    case "append":
      insertHTML(el, "beforeend", html);
      break;
    case "prepend":
      insertHTML(el, "afterbegin", html);
      break;
    case "replace":
      el.outerHTML = html;
      break;
    case "morph":
      (opts.morphFn ?? morph)(el, html);
      break;
    case "remove":
      el.remove?.();
      break;
  }
}

export function connect(url, opts = {}) {
  const ES = opts.EventSource ?? (typeof EventSource !== "undefined" ? EventSource : null);
  if (!ES || !url) return null;
  const src = new ES(url);
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  for (const kind of STREAM_KINDS) {
    src.addEventListener(kind, (ev) => {
      for (const op of parseSSE(sseEnvelope(kind, ev.data))) {
        applyOp(op, doc, opts);
      }
    });
  }
  return src;
}

export function start(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc) return;
  const nodes =
    typeof doc.querySelectorAll === "function" ? doc.querySelectorAll("[data-amarra-stream]") : [];
  for (const el of nodes) {
    const url = el.getAttribute?.("data-amarra-stream");
    if (!url) continue;
    connect(url, {
      ...opts,
      document: doc,
      defaultTarget: el.getAttribute?.("data-amarra-target") || opts.defaultTarget,
    });
  }
}

function sseEnvelope(kind, data) {
  const lines = String(data ?? "")
    .split("\n")
    .map((line) => `data: ${line}`);
  return `event: ${kind}\n${lines.join("\n")}\n\n`;
}

function insertHTML(el, pos, html) {
  if (typeof el.insertAdjacentHTML === "function") {
    el.insertAdjacentHTML(pos, html);
    return;
  }
  if (pos === "afterbegin") el.innerHTML = html + (el.innerHTML || "");
  else el.innerHTML = (el.innerHTML || "") + html;
}
