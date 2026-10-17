import { useEffect } from 'react';
import { matchRoutes, useLocation, useRoutes } from 'react-router-dom';
import { ThemeProvider } from 'styled-components';
import { theme } from './styles/theme.js';
import { GlobalStyle } from './styles/GlobalStyle.js';
import { routes } from './routes.jsx';
import { Footer, Header } from './components/Layout.jsx';
import { CityModal, LoginModal, Sidebar } from './components/Overlays.jsx';
import { Toasts } from './components/ui.jsx';
import { track } from './lib/clickstream.js';

function ScrollAndTrack() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo(0, 0);
    track('page_viewed', { event_type: 'screen_view' });
  }, [pathname]);
  return null;
}

export function App() {
  const location = useLocation();
  const element = useRoutes(routes);
  const match = matchRoutes(routes, location)?.[0]?.route;
  const chrome = match?.chrome !== false;
  return (
    <ThemeProvider theme={theme}>
      <GlobalStyle />
      <ScrollAndTrack />
      {chrome && <Header />}
      <div style={{ minHeight: '60vh' }}>{element}</div>
      {chrome && <Footer />}
      <CityModal />
      <LoginModal />
      <Sidebar />
      <Toasts />
    </ThemeProvider>
  );
}
