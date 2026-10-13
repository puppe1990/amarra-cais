import { restoreSvgAttributeCase } from "./svg_attrs.mjs";

/**
 * Idiomorph wrapper. Tests inject morphFn.
 * When Idiomorph is absent (node tests / missing vendor), children are replaced
 * via innerHTML so morph(el, html) still has a defined contract.
 */
export function morph(el, html, morphFn, opts = {}) {
  if (!el) return;
  const parent = el.parentNode ?? null;
  const result = morphInto(el, html, morphFn, opts);
  // #330: an outerHTML morph detaches the target, so the replacement lives in the
  // old parent — normalize whichever subtree now holds the morphed nodes.
  restoreSvgAttributeCase(el.isConnected === false && parent ? parent : el);
  return result;
}

function morphInto(el, html, morphFn, opts) {
  if (typeof morphFn === "function") return morphFn(el, html);
  const style = opts.morphStyle || "innerHTML";
  const lib = globalThis.Idiomorph;
  if (lib && typeof lib.morph === "function") {
    return lib.morph(el, html, { morphStyle: style });
  }
  if (style === "outerHTML" && "outerHTML" in el) {
    el.outerHTML = html ?? "";
    return;
  }
  if ("innerHTML" in el) el.innerHTML = html ?? "";
}
