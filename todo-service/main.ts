import { Application, Router } from "https://deno.land/x/oak@v11.1.0/mod.ts";

const router = new Router();

router.get("/", (ctx) => {
  ctx.response.body = JSON.stringify({
    hello: "there",
  });
});

const app = new Application();

app.use(router.routes());
app.use(router.allowedMethods());

app.listen({ port: 8080 });
