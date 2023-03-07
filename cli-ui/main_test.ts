import { assertEquals } from "https://deno.land/std@0.178.0/testing/asserts.ts";
import { add } from "./main.ts";

Deno.test(function addTest() {
  assertEquals(add(2, 3), 5);
});

Deno.test(function nextAddTest() {
  assertEquals(add(4, 4), 8);
});

Deno.test(function sanity() {
  assertEquals(true, true);
});
