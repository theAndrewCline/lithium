import { serve } from "@hono/node-server";
import { serveStatic } from "@hono/node-server/serve-static";
import { Hono } from "hono";

const app = new Hono();

app.use("*", async (c, next) => {
  c.setRenderer((content) => {
    return c.html(
      <html lang="en-US">
        <head>
          <meta charset="UTF-8" />
          <meta
            name="viewport"
            content="width=device-width, initial-scale=1.0"
          />
          <link href="/public/output.css" rel="stylesheet" />
        </head>
        <body>{content}</body>
      </html>,
    );
  });
  await next();
});

app.use("/public/*", serveStatic({ root: "./dist" }));

app.get("/up", (c) => {
  return c.json({ status: "foo" });
});

app.get("/", (c) => {
  return c.render(
    <h1 className="text-blue-900 font-bold text-4xl">Hello there</h1>,
  );
});

serve(app, ({ port }) => {
  console.log(`listening on port: ${port}`);
});
