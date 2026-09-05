import { csrfTokenFromMeta } from "./hook.mjs";
import { morph } from "./morph.mjs";

export function wsURL(loc, { view, topic } = {}) {
  if (!loc?.host) return "";
  const proto = loc.protocol === "https:" ? "wss:" : "ws:";
  const q = new URLSearchParams();
  if (view) q.set("view", view);
  if (topic) q.set("topic", topic);
  const qs = q.toString();
  return `${proto}//${loc.host}/amarra/live${qs ? `?${qs}` : ""}`;
}

export function liveRoot(el) {
  return el?.closest?.("[amarra-live]") ?? null;
}

export function eventName(el, kind) {
  if (!el?.getAttribute) return "";
  if (kind === "click") return el.getAttribute("amarra-click") || "";
  if (kind === "change") return el.getAttribute("amarra-change") || "";
  if (kind === "submit") return el.getAttribute("amarra-submit") || "";
  return "";
}

export function formPayload(form, FormDataCtor = FormData) {
  if (!form || !FormDataCtor) return {};
  const data = {};
  for (const [k, v] of new FormDataCtor(form)) {
    data[k] = v;
  }
  return data;
}

export function applyLiveMessage(msg, root, morphFn = morph) {
  if (!msg || !root) return;
  if (msg.type !== "ok" && msg.type !== "morph") return;
  const html = msg.html ?? "";
  if (msg.target) {
    const el = root.querySelector?.(`#${cssEscape(msg.target)}`) || root;
    morphFn(el, html);
    return;
  }
  morphFn(root, html);
}

function cssEscape(id) {
  if (typeof CSS !== "undefined" && CSS.escape) return CSS.escape(id);
  return String(id).replace(/[^a-zA-Z0-9_-]/g, "\\$&");
}

export function start(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc || typeof doc.addEventListener !== "function") return;
  if (doc.documentElement?.dataset?.amarraLive === "true") return;
  if (doc.documentElement?.dataset) doc.documentElement.dataset.amarraLive = "true";

  const WS = opts.WebSocket ?? (typeof WebSocket !== "undefined" ? WebSocket : null);
  const location = opts.location ?? (typeof window !== "undefined" ? window.location : null);
  const sockets = new Map();

  function connect(root) {
    if (!root || !WS || !location) return;
    const view = root.getAttribute("amarra-live");
    if (!view) return;
    const topic = root.getAttribute("data-amarra-topic") || view;
    const url = wsURL(location, { view, topic });
    const ws = new WS(url);
    sockets.set(root, ws);
    ws.addEventListener("open", () => {
      const csrf = opts.csrfToken ?? csrfTokenFromMeta(doc);
      ws.send(JSON.stringify({ type: "join", csrf }));
    });
    ws.addEventListener("message", (ev) => {
      let msg;
      try {
        msg = JSON.parse(ev.data);
      } catch {
        return;
      }
      applyLiveMessage(msg, root, opts.morphFn);
    });
    ws.addEventListener("close", () => {
      sockets.delete(root);
      const wait = opts.reconnectMs ?? 1000;
      if (wait < 0) return;
      setTimeout(() => {
        if (!doc.contains?.(root) && root.isConnected === false) return;
        connect(root);
      }, wait);
    });
  }

  doc.querySelectorAll?.("[amarra-live]").forEach((el) => connect(el));

  function sendFrom(el, kind, extra) {
    const root = liveRoot(el);
    if (!root) return false;
    const name = eventName(kind === "submit" ? el : el.closest?.(`[amarra-${kind}]`) || el, kind);
    if (!name) return false;
    const ws = sockets.get(root);
    if (!ws || ws.readyState !== 1) return false;
    const payload = extra ?? {};
    ws.send(JSON.stringify({ type: "event", event: name, payload, ref: String(Date.now()) }));
    return true;
  }

  doc.addEventListener(
    "click",
    (ev) => {
      const btn = ev.target?.closest?.("[amarra-click]");
      if (!btn || !liveRoot(btn)) return;
      ev.preventDefault();
      sendFrom(btn, "click", { value: btn.value ?? btn.textContent ?? "" });
    },
    true
  );

  doc.addEventListener(
    "change",
    (ev) => {
      const el = ev.target?.closest?.("[amarra-change]");
      if (!el || !liveRoot(el)) return;
      sendFrom(el, "change", { value: el.value ?? "" });
    },
    true
  );

  doc.addEventListener(
    "submit",
    (ev) => {
      const form = ev.target?.closest?.("form[amarra-submit]");
      if (!form || !liveRoot(form)) return;
      ev.preventDefault();
      sendFrom(form, "submit", formPayload(form, opts.FormData));
    },
    true
  );
}
