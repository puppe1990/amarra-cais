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
      if (this.getAttribute("src")) loadFrame(this, opts);
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
