// #397: copy <template data-amarra-outlet="id"> into #id after a Drive morph.
// Scripts in the template do not run: content is taken from a <template>.
export function applyOutlets(doc, html) {
  if (!doc?.getElementById) return;
  const source = String(html ?? "");
  const re = /<template\b[^>]*\bdata-amarra-outlet=["']([^"']+)["'][^>]*>([\s\S]*?)<\/template>/gi;
  let m;
  while ((m = re.exec(source))) {
    const target = doc.getElementById(m[1]);
    if (!target) continue;
    fillOutlet(doc, target, m[2]);
  }
}

function fillOutlet(doc, target, inner) {
  if (typeof doc.createElement === "function") {
    const tpl = doc.createElement("template");
    tpl.innerHTML = inner;
    const nodes = [...(tpl.content?.childNodes ?? [])];
    if (typeof target.replaceChildren === "function") {
      target.replaceChildren(...nodes);
      return;
    }
  }
  if ("innerHTML" in target) target.innerHTML = inner;
}
