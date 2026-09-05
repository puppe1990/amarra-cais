import { morph } from "./morph.mjs";
import { csrfTokenFromMeta } from "./hook.mjs";

export function frameHeaders(id, csrfToken) {
  const headers = {
    "Amarra-Frame": String(id ?? ""),
    Accept: "text/html",
  };
  if (csrfToken) headers["X-CSRF-Token"] = csrfToken;
  return headers;
}

export function frameTarget(el) {
  return el?.getAttribute?.("data-amarra-frame") || "";
}

export async function visitIntoFrame(el, href, opts = {}) {
  const id = frameTarget(el);
  if (!id || id === "_top") return false;
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  const frame = doc?.getElementById?.(id) || null;
  if (!frame) return false;
  const prev = frame.getAttribute?.("src");
  if (typeof frame.setAttribute === "function") frame.setAttribute("src", href);
  const tag = frame.tagName ? String(frame.tagName).toUpperCase() : "";
  if (tag === "AMARRA-FRAME" && prev !== href) return true;
  await loadFrame(frame, { ...opts, document: doc });
  return true;
}

export function observeLazy(el, opts = {}) {
  const IO = opts.IntersectionObserver ?? globalThis.IntersectionObserver;
  if (typeof IO !== "function") {
    return loadFrame(el, opts);
  }
  const io = new IO((entries) => {
    if (!entries?.some?.((e) => e.isIntersecting)) return;
    io.disconnect();
    void loadFrame(el, opts);
  });
  io.observe(el);
  return io;
}

export async function loadFrame(el, opts = {}) {
  if (!el) return;
  const src = el.getAttribute?.("src");
  if (!src) return;
  const id = el.getAttribute?.("id") || el.id || "";
  const fetchFn = opts.fetchFn ?? opts.fetch ?? fetch;
  const csrfToken =
    opts.csrfToken ??
    csrfTokenFromMeta(opts.document ?? (typeof document !== "undefined" ? document : ""));
  const res = await fetchFn(src, {
    headers: frameHeaders(id, csrfToken),
    credentials: "same-origin",
    redirect: "follow",
  });
  const html = await res.text();
  (opts.morphFn ?? morph)(el, html);
  const doc = opts.document ?? (typeof document !== "undefined" ? document : null);
  if (doc && typeof doc.dispatchEvent === "function") {
    doc.dispatchEvent(new CustomEvent("amarra:morphed", { bubbles: true }));
  }
  if (el.hasAttribute?.("amarra-push") && opts.history?.pushState) {
    opts.history.pushState({ amarra: true }, "", src);
  }
}

export function define(opts = {}) {
  const registry =
    opts.customElements ?? (typeof customElements !== "undefined" ? customElements : null);
  const Base = opts.HTMLElement ?? (typeof HTMLElement !== "undefined" ? HTMLElement : null);
  if (!registry || !Base || typeof registry.define !== "function") return false;
  if (typeof registry.get === "function" && registry.get("amarra-frame")) return true;

  class AmarraFrame extends Base {
    connectedCallback() {
      if (!this.getAttribute("src")) return;
      if (this.getAttribute("loading") === "lazy") {
        observeLazy(this, opts);
        return;
      }
      loadFrame(this, opts);
    }

    static get observedAttributes() {
      return ["src"];
    }

    attributeChangedCallback(name, prev, next) {
      if (name === "src" && next && prev !== next) loadFrame(this, opts);
    }
  }

  registry.define("amarra-frame", AmarraFrame);
  return true;
}
