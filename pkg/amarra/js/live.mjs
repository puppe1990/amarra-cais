import { csrfTokenFromMeta, dispatchLivePush } from "./hook.mjs";
import { morph } from "./morph.mjs";
import { applyOp } from "./stream.mjs";

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

export function debounceWait(el) {
  const n = parseInt(el?.getAttribute?.("amarra-debounce") || "0", 10);
  return Number.isFinite(n) && n > 0 ? n : 0;
}

export function setLoading(el, root, on) {
  const op = on ? "add" : "remove";
  el?.classList?.[op]?.("amarra-click-loading");
  root?.classList?.[op]?.("amarra-loading");
}

export function applyLiveMessage(msg, root, morphFn = morph, extras = {}) {
  if (!msg || !root) return;
  if (msg.type !== "ok" && msg.type !== "morph") return;
  const html = msg.html ?? "";
  if (html !== "") {
    if (msg.target) {
      const el = root.querySelector?.(`#${cssEscape(msg.target)}`) || root;
      // Named targets send the element itself (g live counter: <span id="count">).
      // innerHTML morph inserts that node into itself and Idiomorph throws (#294).
      morphFn(el, html, undefined, { morphStyle: "outerHTML" });
    } else {
      morphFn(root, html);
    }
  }
  const opFn = extras.applyOp ?? applyOp;
  const doc = extras.document ?? root.ownerDocument;
  for (const op of msg.ops || []) opFn(op, doc, extras);
  if (msg.patch) extras.history?.pushState?.({}, "", msg.patch);
  if (msg.navigate && extras.location) extras.location.href = msg.navigate;
  const pushFn = extras.dispatchPush ?? dispatchLivePush;
  for (const p of msg.pushes || []) pushFn(p.event, p.payload);
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:morphed", { bubbles: true }));
  }
}

function cssEscape(id) {
  if (typeof CSS !== "undefined" && CSS.escape) return CSS.escape(id);
  return String(id).replace(/[^a-zA-Z0-9_-]/g, "\\$&");
}

// #375: a fixed reconnect synchronized every client after an outage, and a
// permanent rejection (bad CSRF, failed mount) retried forever. Backoff is
// bounded and jittered so reconnects spread out.
const LIVE_RECONNECT_CODES = new Set([1003, 1008, 1011]);

export function reconnectDelay(attempt, opts = {}) {
  const base = opts.baseMs ?? 1000;
  const max = opts.maxMs ?? 30000;
  const factor = opts.factor ?? 2;
  const jitter = opts.jitter ?? 0.2;
  const random = opts.random ?? Math.random;
  const capped = Math.min(max, base * factor ** Math.max(0, attempt - 1));
  const spread = capped * jitter;
  return Math.max(0, Math.round(capped - spread + random() * spread * 2));
}

export function liveState(root, state, detail = {}) {
  root?.setAttribute?.("data-amarra-live-state", state);
  if (typeof root?.dispatchEvent !== "function" || typeof CustomEvent !== "function") return;
  root.dispatchEvent(new CustomEvent("amarra:live-state", { bubbles: true, detail: { state, ...detail } }));
}

export function start(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc || typeof doc.addEventListener !== "function") return;
  if (doc.documentElement?.dataset?.amarraLive === "true") return;
  if (doc.documentElement?.dataset) doc.documentElement.dataset.amarraLive = "true";

  const WS = opts.WebSocket ?? (typeof WebSocket !== "undefined" ? WebSocket : null);
  const location = opts.location ?? (typeof window !== "undefined" ? window.location : null);
  const connections = new Map();
  const setTimer = opts.setTimeout ?? setTimeout;
  const clearTimer = opts.clearTimeout ?? clearTimeout;
  const disabled = opts.reconnectMs !== undefined && opts.reconnectMs < 0;

  function removed(root) {
    return root.isConnected === false || (typeof doc.contains === "function" && !doc.contains(root));
  }

  function onClose(root, record, ev) {
    record.ws = null;
    if (!connections.has(root)) return;
    if (removed(root)) {
      root?.removeAttribute?.("data-amarra-live-state");
      connections.delete(root);
      return;
    }
    const code = ev?.code ?? 1006;
    if (LIVE_RECONNECT_CODES.has(code)) {
      record.failed = true;
      liveState(root, "failed", { permanent: true, code, reason: ev?.reason ?? "", attempt: record.attempts });
      return;
    }
    if (disabled) {
      liveState(root, "closed", { code });
      return;
    }
    record.attempts += 1;
    const wait = reconnectDelay(record.attempts, {
      baseMs: opts.reconnectMs ?? 1000,
      maxMs: opts.maxMs,
      jitter: opts.jitter,
      random: opts.random,
    });
    liveState(root, "reconnecting", { code, attempt: record.attempts, retryInMs: wait });
    record.timer = setTimer(() => {
      record.timer = null;
      if (!connections.has(root) || removed(root)) {
        connections.delete(root);
        return;
      }
      open(root, record);
    }, wait);
  }

  function open(root, record) {
    const ws = new WS(record.url);
    record.ws = ws;
    liveState(root, "connecting", { attempt: record.attempts });
    ws.addEventListener("open", () => {
      record.attempts = 0;
      liveState(root, "open");
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
      applyLiveMessage(msg, root, opts.morphFn, {
        document: doc,
        history: opts.history ?? (typeof window !== "undefined" ? window.history : null),
        location,
      });
      setLoading(root._amarraPending, root, false);
      root._amarraPending = null;
    });
    ws.addEventListener("close", (ev) => onClose(root, record, ev));
  }

  function connect(root) {
    if (!root || !WS || !location || connections.has(root)) return;
    const view = root.getAttribute("amarra-live");
    if (!view) return;
    const topic = root.getAttribute("data-amarra-topic") || view;
    const record = {
      url: wsURL(location, { view, topic }),
      ws: null,
      attempts: 0,
      timer: null,
      failed: false,
    };
    connections.set(root, record);
    open(root, record);
  }

  // #113: Drive morphs the page on navigation; rescan on amarra:morphed so a
  // chat reached via link still joins, and close sockets of replaced roots.
  // A pending backoff keeps its record, so the morph cannot open a duplicate.
  function sync() {
    for (const [root, record] of connections) {
      if (!removed(root)) continue;
      if (record.timer) clearTimer(record.timer);
      record.ws?.close?.();
      root?.removeAttribute?.("data-amarra-live-state");
      connections.delete(root);
    }
    doc.querySelectorAll?.("[amarra-live]").forEach((el) => {
      if (!connections.has(el)) connect(el);
    });
  }
  sync();
  if (typeof doc.addEventListener === "function") doc.addEventListener("amarra:morphed", sync);

  function sendFrom(el, kind, extra) {
    const root = liveRoot(el);
    if (!root) return false;
    const name = eventName(kind === "submit" ? el : el.closest?.(`[amarra-${kind}]`) || el, kind);
    if (!name) return false;
    const ws = sockets.get(root);
    if (!ws || ws.readyState !== 1) return false;
    const payload = extra ?? {};
    const send = () => {
      setLoading(el, root, true);
      root._amarraPending = el;
      ws.send(JSON.stringify({ type: "event", event: name, payload, ref: String(Date.now()) }));
    };
    const wait = debounceWait(el);
    if (wait) {
      clearTimeout(el._amarraDebounce);
      el._amarraDebounce = setTimeout(send, wait);
      return true;
    }
    send();
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
