import { test } from "node:test";
import assert from "node:assert/strict";
import { highlight, tokenize } from "../static/js/highlight.js";

const kinds = (src) => tokenize(src).filter((t) => t.kind !== "ws").map((t) => [t.kind, t.text]);

test("channel operator is its own token so sends and receives stand out", () => {
  assert.deepEqual(kinds("v := <-ch"), [["id", "v"], ["op", ":="], ["chan", "<-"], ["id", "ch"]]);
});

test("keywords, builtin types and builtin funcs get distinct classes", () => {
  assert.deepEqual(kinds("go make(chan int)"), [["kw", "go"], ["bi", "make"], ["pun", "("], ["kw", "chan"], ["ty", "int"], ["pun", ")"]]);
});

test("package selectors and their calls are told apart", () => {
  assert.deepEqual(kinds("fmt.Println(x)"), [["pkg", "fmt"], ["pun", "."], ["fn", "Println"], ["pun", "("], ["id", "x"], ["pun", ")"]]);
});

test("a method named like a builtin is still a function call", () => {
  assert.deepEqual(kinds("s.close()")[2], ["fn", "close"]);
});

test("strings and comments swallow keywords inside them", () => {
  assert.deepEqual(kinds('"go func" // for range'), [["str", '"go func"'], ["com", "// for range"]]);
  assert.deepEqual(kinds("`raw\nfor`"), [["str", "`raw\nfor`"]]);
});

test("numbers in every base", () => {
  for (const n of ["42", "0x1F", "0b101", "1_000", "3.14", "1e9"]) assert.equal(tokenize(n)[0].kind, "num", n);
});

test("highlight escapes html so code can never inject markup", () => {
  const out = highlight('x := "<script>"');
  assert.ok(!out.includes("<script>"));
  assert.ok(out.includes("&lt;script&gt;"));
});

test("highlight keeps every character so the overlay lines up with the textarea", () => {
  const src = "package main\n\nfunc main() {\n\tch := make(chan int, 1)\n\tch <- 1\n}\n";
  const text = highlight(src).replace(/<[^>]+>/g, "").replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&");
  assert.equal(text, src);
});

test("unterminated string does not hang the tokenizer", () => {
  assert.equal(tokenize('"open').length, 1);
});
