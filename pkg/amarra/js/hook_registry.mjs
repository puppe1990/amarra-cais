const defs = new Map();
const mounted = new Map();

export function register(name, def) {
  if (!name || !def) return;
  defs.set(name, def);
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
    const name = el.getAttribute?.("amarra-hook") || "";
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

export function dispatchLivePush(event, payload) {
  for (const [el, cur] of mounted) {
    cur.def.handleEvent?.(event, payload, el);
  }
}

function collect(root) {
  const out = [];
  if (root.hasAttribute?.("amarra-hook")) out.push(root);
  const list = root.querySelectorAll?.("[amarra-hook]");
  if (list) {
    for (const el of list) out.push(el);
  }
  return out;
}
