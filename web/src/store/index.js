import { configureStore } from '@reduxjs/toolkit';
import { api } from './api.js';
import ui from './ui.js';

export function createStore(preloadedState, extra = {}) {
  return configureStore({
    reducer: { [api.reducerPath]: api.reducer, ui },
    preloadedState,
    middleware: (getDefault) =>
      getDefault({ thunk: { extraArgument: extra }, serializableCheck: false, immutableCheck: false }).concat(api.middleware),
  });
}
