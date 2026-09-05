export function captureScroll(history, y) {
  if (!history?.replaceState) return;
  const prev = history.state && typeof history.state === "object" ? history.state : {};
  history.replaceState({ ...prev, amarra: true, scrollY: y ?? 0 }, "");
}

export function restoreScroll(win, state) {
  const y = state?.scrollY;
  if (typeof y !== "number") return;
  win?.scrollTo?.(0, y);
}

export function focusFirstInvalid(doc) {
  const el = doc?.querySelector?.('[aria-invalid="true"], [data-amarra-invalid]');
  if (el && typeof el.focus === "function") el.focus();
}
