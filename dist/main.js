import { jsx as _jsx, jsxs as _jsxs } from "hono/jsx/jsx-runtime";
import { serve } from "@hono/node-server";
import { serveStatic } from "@hono/node-server/serve-static";
import { Hono } from "hono";
const app = new Hono();
app.use("*", async (c, next)=>{
    c.setRenderer((content)=>{
        return c.html(/*#__PURE__*/ _jsxs("html", {
            lang: "en-US",
            children: [
                /*#__PURE__*/ _jsxs("head", {
                    children: [
                        /*#__PURE__*/ _jsx("meta", {
                            charset: "UTF-8"
                        }),
                        /*#__PURE__*/ _jsx("meta", {
                            name: "viewport",
                            content: "width=device-width, initial-scale=1.0"
                        }),
                        /*#__PURE__*/ _jsx("link", {
                            href: "/public/output.css",
                            rel: "stylesheet"
                        })
                    ]
                }),
                /*#__PURE__*/ _jsx("body", {
                    children: content
                })
            ]
        }));
    });
    await next();
});
app.use("/public/*", serveStatic({
    root: "./dist"
}));
app.get("/up", (c)=>{
    return c.json({
        status: "foo"
    });
});
app.get("/", (c)=>{
    return c.render(/*#__PURE__*/ _jsx("h1", {
        className: "text-blue-900 font-bold text-4xl",
        children: "Hello there"
    }));
});
serve(app, ({ port })=>{
    console.log(`listening on port: ${port}`);
});

