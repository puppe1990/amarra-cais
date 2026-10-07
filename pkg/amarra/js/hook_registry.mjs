const defs = new Map();
const mounted = new Map();
// Tag names that get a builtin without an amarra-hook attribute (e.g. select
// search is on by default). Kept separate from defs so scan() can pick them up.
const auto = new Map();

export function register(name, def) {
  if (!name || !def) return;
  defs.set(name, def);
}

export function registerAuto(tag, name) {
  if (tag && name) auto.set(String(tag).toUpperCase(), name);
}

export function reset() {
  defs.clear();
  mounted.clear();
}

export function scan(root) {
  if (!root) return;
  const found = collect(root);
  const seen = new Set(found);
  for (const el of found) {
    const name = hookName(el);
    const def = defs.get(name);
    const cur = mounted.get(el);
    if (cur && cur.name === name) {
      def?.updated?.(el);
      continue;
    }
    if (cur) {
      cur.def.disconnect?.(el);
      mounted.delete(el);
    }
    if (!def) continue;
    mounted.set(el, { name, def });
    def.connect?.(el);
  }
  for (const [el, cur] of [...mounted]) {
    if (seen.has(el)) continue;
    cur.def.disconnect?.(el);
    mounted.delete(el);
  }
}

function hookName(el) {
  const explicit = el.getAttribute?.("amarra-hook") || "";
  if (explicit) return explicit;
  return auto.get(el.tagName?.toUpperCase?.()) ?? "";
}

export function dispatchLivePush(event, payload) {
  for (const [el, cur] of mounted) {
    cur.def.handleEvent?.(event, payload, el);
  }
}

function collect(root) {
  const out = [];
  const push = (el) => {
    if (el && !out.includes(el)) out.push(el);
  };
  if (root.hasAttribute?.("amarra-hook") || auto.has(root.tagName?.toUpperCase?.())) push(root);
  for (const sel of ["[amarra-hook]", ...auto.keys()]) {
    const list = root.querySelectorAll?.(sel);
    if (!list) continue;
    for (const el of list) push(el);
  }
  return out;
}
