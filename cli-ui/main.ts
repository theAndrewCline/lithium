import { Command } from "https://deno.land/x/cliffy@v0.25.7/mod.ts";

const cmd = await new Command()
  .name("lithium")
  .version("0.1.0")
  .description("Commandline ui for lithium todos")
  .parse(Deno.args);
