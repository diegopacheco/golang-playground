export default [
  { ignores: ["electron/node_modules/**", "electron/out/**"] },
  {
    files: ["web/**/*.js", "web/**/*.mjs"],
    languageOptions: { ecmaVersion: "latest", sourceType: "module", globals: { window: "readonly", document: "readonly", localStorage: "readonly", fetch: "readonly", navigator: "readonly", getComputedStyle: "readonly", requestAnimationFrame: "readonly", setTimeout: "readonly", clearTimeout: "readonly", Event: "readonly" } },
    rules: { "no-unused-vars": "error", "no-undef": "error" },
  },
  {
    files: ["electron/**/*.js"],
    languageOptions: { ecmaVersion: "latest", sourceType: "commonjs", globals: { require: "readonly", module: "writable", __dirname: "readonly", process: "readonly", console: "readonly", setTimeout: "readonly", clearTimeout: "readonly", setInterval: "readonly", clearInterval: "readonly", fetch: "readonly", window: "readonly", document: "readonly" } },
    rules: { "no-unused-vars": "error", "no-undef": "error" },
  },
];
