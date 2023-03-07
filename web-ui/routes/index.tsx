import React from "preact";
import { Head } from "$fresh/runtime.ts";
import Counter from "../islands/Counter.tsx";

export default function Home() {
  return (
    <>
      <Head>
        <h1>Foo</h1>
      </Head>
      <Counter start={0} />
    </>
  );
}
