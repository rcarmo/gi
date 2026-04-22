import jsyaml from "js-yaml";

(globalThis as Record<string, unknown>).jsyaml = jsyaml;
