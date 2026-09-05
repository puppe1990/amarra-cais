import { Idiomorph } from "./vendor/idiomorph.js";
import * as drive from "./drive.mjs";
import * as frame from "./frame.mjs";
import * as stream from "./stream.mjs";
import * as hook from "./hook.mjs";
import * as live from "./live.mjs";
import { morph } from "./morph.mjs";

if (typeof globalThis !== "undefined") {
  globalThis.Idiomorph = Idiomorph;
}

export function boot() {
  if (typeof window === "undefined") return;
  hook.start();
  drive.start();
  frame.define();
  stream.start();
  live.start();
  window.amarra = {
    drive,
    live,
    hook,
  };
}

if (typeof window !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
}

export { drive, frame, stream, hook, live, morph };
