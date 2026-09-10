const STATE = "_amarraBulkState";

// #23: select-all on the current page. The bulk action itself is still a
// Drive POST/DELETE of the selected ids — this hook only drives the header
// checkbox (checked/indeterminate), the row checkboxes, and the optional
// bulk bar. Selection never spans pages.
export function makeBulk() {
  return {
    connect(el) {
      if (!el?.querySelector) return;
      const all = el.querySelector("[data-amarra-bulk-all]");
      const rows = [...(el.querySelectorAll?.("[data-amarra-bulk-row]") ?? [])];
      const bar = el.querySelector("[data-amarra-bulk-bar]");
      const count = bar?.querySelector?.("[data-amarra-bulk-count]");
      if (!all || rows.length === 0) return;

      const sync = () => {
        const selected = rows.filter((r) => r.checked).length;
        all.indeterminate = selected > 0 && selected < rows.length;
        all.checked = selected === rows.length;
        if (bar) {
          bar.hidden = selected === 0;
          if (count) count.textContent = String(selected);
        }
      };
      const onAll = () => {
        for (const row of rows) row.checked = all.checked;
        sync();
      };

      all.addEventListener?.("change", onAll);
      const rowUnbinds = [];
      for (const row of rows) {
        const fn = () => sync();
        row.addEventListener?.("change", fn);
        rowUnbinds.push([row, fn]);
      }
      el[STATE] = { all, onAll, rowUnbinds };
    },
    disconnect(el) {
      const st = el?.[STATE];
      if (!st) return;
      st.all.removeEventListener?.("change", st.onAll);
      for (const [row, fn] of st.rowUnbinds) row.removeEventListener?.("change", fn);
      delete el[STATE];
    },
  };
}

export const bulk = makeBulk();
