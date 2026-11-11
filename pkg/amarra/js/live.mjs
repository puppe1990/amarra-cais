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
  if (msg.type !== "ok" && msg.type !== "morph" && msg.type !== "resync") return;
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
  if (
    msg.type === "resync" &&
    doc &&
    typeof doc.dispatchEvent === "function" &&
    typeof CustomEvent === "function"
  ) {
    // The server rendered the authoritative state after dropping events (#376).
    doc.dispatchEvent(
      new CustomEvent("amarra:live-resync", { bubbles: true, detail: { lost: msg.lost ?? 0 } })
    );
  }
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
  root.dispatchEvent(
    new CustomEvent("amarra:live-state", { bubbles: true, detail: { state, ...detail } })
  );
}

export function start(opts = {}) {
  const session = liveSession(opts);
  if (!session) return;
  liveSync(session);
  session.document.addEventListener("amarra:morphed", () => liveSync(session));
  session.document.addEventListener("click", (ev) => onLiveClick(ev, session), true);
  session.document.addEventListener("change", (ev) => onLiveChange(ev, session), true);
  session.document.addEventListener("submit", (ev) => onLiveSubmit(ev, session), true);
}

function liveSession(opts = {}) {
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (!doc || typeof doc.addEventListener !== "function") return null;
  if (doc.documentElement?.dataset?.amarraLive === "true") return null;
  if (doc.documentElement?.dataset) doc.documentElement.dataset.amarraLive = "true";
  return {
    document: doc,
    opts,
    WS: opts.WebSocket ?? (typeof WebSocket !== "undefined" ? WebSocket : null),
    location: opts.location ?? (typeof window !== "undefined" ? window.location : null),
    connections: new Map(),
    setTimer: opts.setTimeout ?? setTimeout,
    clearTimer: opts.clearTimeout ?? clearTimeout,
    disabled: opts.reconnectMs !== undefined && opts.reconnectMs < 0,
  };
}

function liveRemoved(session, root) {
  return (
    root.isConnected === false ||
    (typeof session.document.contains === "function" && !session.document.contains(root))
  );
}

function liveOnClose(session, root, record, ev) {
  record.ws = null;
  if (!session.connections.has(root)) return;
  if (liveRemoved(session, root)) {
    root?.removeAttribute?.("data-amarra-live-state");
    session.connections.delete(root);
    return;
  }
  const code = ev?.code ?? 1006;
  if (LIVE_RECONNECT_CODES.has(code)) {
    record.failed = true;
    liveState(root, "failed", {
      permanent: true,
      code,
      reason: ev?.reason ?? "",
      attempt: record.attempts,
    });
    return;
  }
  if (session.disabled) {
    liveState(root, "closed", { code });
    return;
  }
  record.attempts += 1;
  const wait = reconnectDelay(record.attempts, {
    baseMs: session.opts.reconnectMs ?? 1000,
    maxMs: session.opts.maxMs,
    jitter: session.opts.jitter,
    random: session.opts.random,
  });
  liveState(root, "reconnecting", { code, attempt: record.attempts, retryInMs: wait });
  record.timer = session.setTimer(() => {
    record.timer = null;
    if (!session.connections.has(root) || liveRemoved(session, root)) {
      session.connections.delete(root);
      return;
    }
    liveOpen(session, root, record);
  }, wait);
}

function liveOpen(session, root, record) {
  const ws = new session.WS(record.url);
  record.ws = ws;
  liveState(root, "connecting", { attempt: record.attempts });
  ws.addEventListener("open", () => {
    record.attempts = 0;
    liveState(root, "open");
    const csrf = session.opts.csrfToken ?? csrfTokenFromMeta(session.document);
    ws.send(JSON.stringify({ type: "join", csrf }));
  });
  ws.addEventListener("message", (ev) => {
    let msg;
    try {
      msg = JSON.parse(ev.data);
    } catch {
      return;
    }
    applyLiveMessage(msg, root, session.opts.morphFn, {
      document: session.document,
      history: session.opts.history ?? (typeof window !== "undefined" ? window.history : null),
      location: session.location,
    });
    setLoading(root._amarraPending, root, false);
    root._amarraPending = null;
  });
  ws.addEventListener("close", (ev) => liveOnClose(session, root, record, ev));
}

function liveConnect(session, root) {
  if (!root || !session.WS || !session.location || session.connections.has(root)) return;
  const view = root.getAttribute("amarra-live");
  if (!view) return;
  const topic = root.getAttribute("data-amarra-topic") || view;
  const record = {
    url: wsURL(session.location, { view, topic }),
    ws: null,
    attempts: 0,
    timer: null,
    failed: false,
  };
  session.connections.set(root, record);
  liveOpen(session, root, record);
}

// #113: Drive morphs the page on navigation; rescan on amarra:morphed so a
// chat reached via link still joins, and close sockets of replaced roots.
// A pending backoff keeps its record, so the morph cannot open a duplicate.
function liveSync(session) {
  for (const [root, record] of session.connections) {
    if (!liveRemoved(session, root)) continue;
    if (record.timer) session.clearTimer(record.timer);
    record.ws?.close?.();
    root?.removeAttribute?.("data-amarra-live-state");
    session.connections.delete(root);
  }
  session.document.querySelectorAll?.("[amarra-live]").forEach((el) => {
    if (!session.connections.has(el)) liveConnect(session, el);
  });
}

function liveSendFrom(session, el, kind, extra) {
  const root = liveRoot(el);
  if (!root) return false;
  const name = eventName(kind === "submit" ? el : el.closest?.(`[amarra-${kind}]`) || el, kind);
  if (!name) return false;
  const ws = session.connections.get(root)?.ws;
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

function onLiveClick(ev, session) {
  const btn = ev.target?.closest?.("[amarra-click]");
  if (!btn || !liveRoot(btn)) return;
  ev.preventDefault();
  liveSendFrom(session, btn, "click", { value: btn.value ?? btn.textContent ?? "" });
}

function onLiveChange(ev, session) {
  const el = ev.target?.closest?.("[amarra-change]");
  if (!el || !liveRoot(el)) return;
  liveSendFrom(session, el, "change", { value: el.value ?? "" });
}

function onLiveSubmit(ev, session) {
  const form = ev.target?.closest?.("form[amarra-submit]");
  if (!form || !liveRoot(form)) return;
  ev.preventDefault();
  liveSendFrom(session, form, "submit", formPayload(form, session.opts.FormData));
}
