#!/usr/bin/env node
/**
 * Minimal JS build using esbuild.
 * Produces the framework IIFE bundles from testable sources.
 *
 * For now this wires the new logic/ and produces placeholder
 * updated assets. Full extraction of cais-core/chat happens over time.
 */
import { build } from "esbuild";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, resolve } from "path";

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dirname, "..");

const outDir = resolve(root, "pkg/cais/pwa/assets");

// #320: stamp the bundle with the framework version so doctor can detect a
// stale vendored amarra.js against the app's go.mod.
function frameworkVersion() {
  const changelog = readFileSync(resolve(root, "CHANGELOG.md"), "utf8");
  const m = changelog.match(/^## \[(\d+\.\d+\.\d+)\]/m);
  return m ? m[1] : "0.0.0";
}
const banner = { js: `/* amarra-cais v${frameworkVersion()} */` };

async function main() {
  await build({
    entryPoints: [resolve(root, "pkg/amarra/js/entry.mjs")],
    bundle: true,
    format: "iife",
    outfile: resolve(outDir, "amarra.js"),
    banner,
    minify: false,
    sourcemap: false,
  });

  console.log("js-build: produced amarra.js");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
