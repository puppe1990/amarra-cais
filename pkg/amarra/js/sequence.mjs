// Per-key monotonic sequence for async UI updates (#125): the newest request
// owns the DOM, stale responses are dropped before morph/pushState/hide.
const counters = new WeakMap();
let fallback = 0;

export function bumpSequence(key) {
  if (!key || typeof key !== "object") {
    fallback += 1;
    return fallback;
  }
  const next = (counters.get(key) ?? 0) + 1;
  counters.set(key, next);
  return next;
}

export function isCurrentSequence(key, seq) {
  if (!key || typeof key !== "object") return seq === fallback;
  return seq === (counters.get(key) ?? 0);
}
