// Production SSR server. Every page request is rendered on the server with
// its data prefetched through RTK Query, then hydrated on the client from
// window.__INITIAL_STATE__ — the same model BookMyShow's web app uses.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import express from 'express';
import compression from 'compression';

const root = path.dirname(fileURLToPath(import.meta.url));
const port = Number(process.env.PORT || 3000);
// Optional: when running outside Docker, forward /api to a gateway.
const proxyApi = process.env.PROXY_API;

const app = express();
app.disable('x-powered-by');
app.use(compression());

app.get('/healthz', (_req, res) => res.json({ status: 'ok' }));

if (proxyApi) {
  app.use('/api', express.raw({ type: '*/*', limit: '1mb' }), async (req, res) => {
    try {
      const headers = { ...req.headers };
      delete headers.host;
      const upstream = await fetch(proxyApi + req.originalUrl, {
        method: req.method,
        headers,
        body: ['GET', 'HEAD'].includes(req.method) ? undefined : req.body,
        redirect: 'manual',
      });
      res.status(upstream.status);
      upstream.headers.forEach((v, k) => {
        if (!['content-encoding', 'content-length', 'transfer-encoding'].includes(k)) res.setHeader(k, v);
      });
      const cookies = upstream.headers.getSetCookie?.() || [];
      if (cookies.length) res.setHeader('set-cookie', cookies);
      res.send(Buffer.from(await upstream.arrayBuffer()));
    } catch (err) {
      res.status(502).json({ error: { code: 'BAD_GATEWAY', message: String(err) } });
    }
  });
}

app.use('/assets', express.static(path.join(root, 'dist/client/assets'), { maxAge: '30d', immutable: true }));
app.use(express.static(path.join(root, 'dist/client'), { index: false }));

const template = fs.readFileSync(path.join(root, 'dist/client/index.html'), 'utf8');
const { render } = await import('./dist/server/entry-server.js');

app.get('*', async (req, res) => {
  try {
    const result = await render({ url: req.originalUrl, cookie: req.headers.cookie || '' });
    if (result.redirect) return res.redirect(result.status || 302, result.redirect);
    const html = template
      .replace('<!--app-head-->', result.head)
      .replace('<!--app-html-->', result.html)
      .replace('<!--app-state-->', result.state);
    res.status(result.status).set({ 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'private, no-store' }).send(html);
  } catch (err) {
    console.error('SSR failed', req.originalUrl, err);
    res.status(500).send('<h1>Sorry for bug-ging</h1><p>Some issue from our side. Kindly refresh the page.</p>');
  }
});

app.listen(port, () => console.log(`web listening on :${port}`));
