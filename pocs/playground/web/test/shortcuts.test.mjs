import { test } from "node:test";
import assert from "node:assert/strict";
import { filterShortcuts, SHORTCUTS, TABS, GROUPS } from "../static/js/shortcuts.js";

test("empty query keeps every shortcut", () => {
  const total = filterShortcuts("").reduce((n, g) => n + g.items.length, 0);
  assert.equal(total, SHORTCUTS.length);
});

test("matching a group title keeps the whole group", () => {
  const res = filterShortcuts("window");
  assert.equal(res.length, 1);
  assert.equal(res[0].items.length, SHORTCUTS.filter((s) => s.group === "window").length);
});

test("matching a description keeps only matching rows", () => {
  const res = filterShortcuts("zoom in");
  assert.deepEqual(res.flatMap((g) => g.items.map((s) => s.id)), ["zoomIn"]);
});

test("matching keys works", () => {
  const ids = filterShortcuts("⌘ K").flatMap((g) => g.items.map((s) => s.id));
  assert.ok(ids.includes("palette"));
});

test("no match returns nothing so the modal can say so", () => {
  assert.equal(filterShortcuts("zzzz").length, 0);
});

test("every tab has a numbered shortcut using the tab icon", () => {
  for (const t of TABS) {
    const s = SHORTCUTS.find((x) => x.id === "tab-" + t.id);
    assert.ok(s, t.id);
    assert.equal(s.icon, t.icon);
    assert.deepEqual(s.keys, ["⌘", t.key]);
  }
});

test("every shortcut belongs to a known group", () => {
  const ids = new Set(GROUPS.map((g) => g.id));
  for (const s of SHORTCUTS) assert.ok(ids.has(s.group), s.id);
});
