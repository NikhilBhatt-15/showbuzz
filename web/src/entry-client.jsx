import { hydrateRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Provider } from 'react-redux';
import { createStore } from './store/index.js';
import { App } from './App.jsx';

// Rehydrate the Redux/RTK Query state the server rendered with.
const store = createStore(window.__INITIAL_STATE__);
delete window.__INITIAL_STATE__;

hydrateRoot(
  document.getElementById('root'),
  <Provider store={store}>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </Provider>,
);
