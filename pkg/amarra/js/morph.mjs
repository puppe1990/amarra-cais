/**
 * Idiomorph wrapper. Tests inject morphFn.
 * When Idiomorph is absent (node tests / missing vendor), children are replaced
 * via innerHTML so morph(el, html) still has a defined contract.
 */
export function morph(el, html, morphFn) {
  if (!el) return;
  if (typeof morphFn === "function") return morphFn(el, html);
  const lib = globalThis.Idiomorph;
  if (lib && typeof lib.morph === "function") {
    return lib.morph(el, html, { morphStyle: "innerHTML" });
  }
  if ("innerHTML" in el) el.innerHTML = html ?? "";
}
