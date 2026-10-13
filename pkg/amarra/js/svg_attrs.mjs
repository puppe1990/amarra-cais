/**
 * SVG attribute names the HTML parser adjusts to camelCase while the SVG DOM is
 * case-sensitive (WHATWG "adjust SVG attributes"):
 * https://html.spec.whatwg.org/multipage/parsing.html#adjust-svg-attributes
 *
 * Idiomorph parses the morphed HTML with `DOMParser(..., "text/html")`. An engine
 * that does not run the SVG adjustment hands `viewbox` to `morphAttributes`, which
 * then writes the lowercase name and — because the parsed node has no `viewBox` —
 * removes the live camelCase attribute. The `<svg>` keeps its box but stops
 * painting (#330). Restoring the canonical name after the morph closes that gap.
 */
const SVG_NS = "http://www.w3.org/2000/svg";

const CAMEL_CASE_SVG_ATTRS = [
  "attributeName",
  "attributeType",
  "baseFrequency",
  "baseProfile",
  "calcMode",
  "clipPathUnits",
  "diffuseConstant",
  "edgeMode",
  "filterUnits",
  "glyphRef",
  "gradientTransform",
  "gradientUnits",
  "kernelMatrix",
  "kernelUnitLength",
  "keyPoints",
  "keySplines",
  "keyTimes",
  "lengthAdjust",
  "limitingConeAngle",
  "markerHeight",
  "markerUnits",
  "markerWidth",
  "maskContentUnits",
  "maskUnits",
  "numOctaves",
  "pathLength",
  "patternContentUnits",
  "patternTransform",
  "patternUnits",
  "pointsAtX",
  "pointsAtY",
  "pointsAtZ",
  "preserveAlpha",
  "preserveAspectRatio",
  "primitiveUnits",
  "refX",
  "refY",
  "repeatCount",
  "repeatDur",
  "requiredExtensions",
  "requiredFeatures",
  "specularConstant",
  "specularExponent",
  "spreadMethod",
  "startOffset",
  "stdDeviation",
  "stitchTiles",
  "surfaceScale",
  "systemLanguage",
  "tableValues",
  "targetX",
  "targetY",
  "textLength",
  "viewBox",
  "viewTarget",
  "xChannelSelector",
  "yChannelSelector",
  "zoomAndPan",
];

const canonicalByLowerCase = new Map(
  CAMEL_CASE_SVG_ATTRS.map((name) => [name.toLowerCase(), name])
);

/**
 * Rewrites lowercased SVG attribute names on SVG elements under `root` back to
 * their camelCase form, dropping the lowercase twin. The fresh value wins, so a
 * stale camelCase attribute does not survive the morph. Elements outside the SVG
 * namespace are left untouched. Returns how many attributes were restored.
 */
export function restoreSvgAttributeCase(root) {
  if (!root || typeof root.querySelectorAll !== "function") return 0;

  let restored = 0;
  for (const element of [root, ...root.querySelectorAll("*")]) {
    if (!isSvgElement(element)) continue;
    for (const attribute of [...(element.attributes ?? [])]) {
      const canonical = canonicalByLowerCase.get(String(attribute.name).toLowerCase());
      if (!canonical || attribute.name === canonical) continue;
      element.setAttribute(canonical, attribute.value);
      element.removeAttribute(attribute.name);
      restored += 1;
    }
  }
  return restored;
}

function isSvgElement(element) {
  return (
    !!element &&
    element.namespaceURI === SVG_NS &&
    typeof element.setAttribute === "function" &&
    typeof element.removeAttribute === "function"
  );
}
