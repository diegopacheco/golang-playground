import { test } from "node:test";
import assert from "node:assert/strict";
import { wordBeforeCaret, insertionFor, caretPosition, indentFor } from "../static/js/completion.js";

test("query keeps the package so members are looked up in it", () => {
  assert.deepEqual(wordBeforeCaret("x := fmt.Pri", 12), { query: "fmt.Pri", replaceFrom: 9 });
});

test("query right after a dot lists all members and replaces nothing", () => {
  assert.deepEqual(wordBeforeCaret("sync.", 5), { query: "sync.", replaceFrom: 5 });
});

test("plain word query", () => {
  assert.deepEqual(wordBeforeCaret("\tsel", 4), { query: "sel", replaceFrom: 1 });
});

test("functions insert parens with the caret inside", () => {
  assert.deepEqual(insertionFor({ label: "Println", kind: "func" }), { text: "Println()", caretBack: 1 });
  assert.deepEqual(insertionFor({ label: "make", kind: "builtin" }), { text: "make()", caretBack: 1 });
});

test("types and keywords insert only the label", () => {
  assert.deepEqual(insertionFor({ label: "int", kind: "builtin" }), { text: "int", caretBack: 0 });
  assert.deepEqual(insertionFor({ label: "WaitGroup", kind: "type" }), { text: "WaitGroup", caretBack: 0 });
});

test("snippets insert their body", () => {
  assert.equal(insertionFor({ label: "gofunc", kind: "snippet", insert: "go func() {}()" }).text, "go func() {}()");
});

test("caret column expands tabs so the popup sits under the cursor", () => {
  assert.deepEqual(caretPosition("a\n\tb", 4), { line: 1, col: 5 });
  assert.deepEqual(caretPosition("ab\tc", 4), { line: 0, col: 5 });
});

test("enter after an opening brace indents one level deeper", () => {
  assert.equal(indentFor("\tfor {", 6), "\t\t");
  assert.equal(indentFor("\tx := 1", 8), "\t");
  assert.equal(indentFor("\tcase v := <-ch:", 16), "\t\t");
});
