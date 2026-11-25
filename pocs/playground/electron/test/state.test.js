const { test } = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { readState, writeState, restoreBounds, toggleMaximize, DEFAULT } = require("../state.js");

const screen = { x: 0, y: 0, width: 1920, height: 1080 };

test("the window reopens exactly where it was closed", () => {
  const saved = { bounds: { x: 200, y: 120, width: 1000, height: 700 } };
  assert.deepEqual(restoreBounds(saved, [screen]), saved.bounds);
});

test("a window saved on a monitor that is gone is not restored off screen", () => {
  const saved = { bounds: { x: 4000, y: 100, width: 1000, height: 700 } };
  assert.deepEqual(restoreBounds(saved, [screen]), { width: 1000, height: 700 });
});

test("first launch uses the default size", () => {
  assert.deepEqual(restoreBounds({}, [screen]), DEFAULT);
});

test("double click fills the work area and remembers where it was", () => {
  const current = { x: 100, y: 100, width: 800, height: 600 };
  const r = toggleMaximize(current, screen, null);
  assert.deepEqual(r.bounds, screen);
  assert.deepEqual(r.restore, current);
});

test("second double click goes back to the exact previous bounds", () => {
  const restore = { x: 100, y: 100, width: 800, height: 600 };
  const r = toggleMaximize({ ...screen }, screen, restore);
  assert.deepEqual(r.bounds, restore);
  assert.equal(r.restore, null);
});

test("state survives a write and read", () => {
  const file = path.join(fs.mkdtempSync(path.join(os.tmpdir(), "gp-")), "s.json");
  writeState(file, { bounds: { x: 1, y: 2, width: 3, height: 4 }, fullScreen: true });
  assert.deepEqual(readState(file), { bounds: { x: 1, y: 2, width: 3, height: 4 }, fullScreen: true });
  assert.deepEqual(readState(file + ".missing"), {});
});
