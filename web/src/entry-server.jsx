import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router-dom/server';
import { matchRoutes } from 'react-router-dom';
import { Provider } from 'react-redux';
import { ServerStyleSheet } from 'styled-components';
import { createStore } from './store/index.js';
import { api } from './store/api.js';
import { App } from './App.jsx';
import { routes } from './routes.jsx';
import { HeadContext } from './lib/head.jsx';
import { DEFAULT_REGION, REGION_COOKIE, parseCookies } from './lib/region.js';

const escapeJSON = (obj) => JSON.stringify(obj).replace(/</g, '\\u003c').replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029');
const escapeHTML = (s) => s.replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);

export async function render({ url, cookie }) {
  const u = new URL(url, 'http://ssr.local');
  const cookies = parseCookies(cookie);
  const regionSlug = cookies[REGION_COOKIE] || '';

  // First visit: send the visitor to the default city and ask them to pick one.
  if (u.pathname === '/') {
    return { redirect: `/explore/home/${regionSlug || DEFAULT_REGION}${regionSlug ? '' : '?pickCity=1'}`, status: 302 };
  }

  const store = createStore({ ui: { regionSlug, cityModalOpen: false, loginOpen: false, loginNext: null, sidebarOpen: false, toasts: [] } }, { cookie });

  store.dispatch(api.endpoints.getRegions.initiate());
  store.dispatch(api.endpoints.getMe.initiate());
  const matches = matchRoutes(routes, u.pathname) || [];
  for (const m of matches) {
    m.route.loadData?.({
      dispatch: store.dispatch,
      params: m.params,
      query: u.searchParams,
      search: u.search.replace(/^\?/, ''),
      region: regionSlug || DEFAULT_REGION,
    });
  }
  await Promise.all(store.dispatch(api.util.getRunningQueriesThunk()));

  const head = { title: 'showbuzz', status: 200 };
  const sheet = new ServerStyleSheet();
  try {
    const html = renderToString(
      sheet.collectStyles(
        <HeadContext.Provider value={head}>
          <Provider store={store}>
            <StaticRouter location={url}>
              <App />
            </StaticRouter>
          </Provider>
        </HeadContext.Provider>,
      ),
    );
    const isNotFound = matches.at(-1)?.route.path === '*';
    return {
      status: isNotFound ? 404 : head.status,
      html,
      head: `<title>${escapeHTML(head.title)}</title>${sheet.getStyleTags()}`,
      state: `<script>window.__INITIAL_STATE__=${escapeJSON(store.getState())}</script>`,
    };
  } finally {
    sheet.seal();
  }
}
