import { createSlice } from '@reduxjs/toolkit';

let toastId = 0;

const ui = createSlice({
  name: 'ui',
  initialState: {
    regionSlug: '',
    cityModalOpen: false,
    loginOpen: false,
    loginNext: null, // path to visit after a successful sign-in
    sidebarOpen: false,
    toasts: [],
  },
  reducers: {
    setRegion(state, { payload }) {
      state.regionSlug = payload;
      state.cityModalOpen = false;
    },
    openCityModal(state) {
      state.cityModalOpen = true;
    },
    closeCityModal(state) {
      state.cityModalOpen = false;
    },
    openLogin(state, { payload }) {
      state.loginOpen = true;
      state.loginNext = payload || null;
      state.sidebarOpen = false;
    },
    closeLogin(state) {
      state.loginOpen = false;
      state.loginNext = null;
    },
    toggleSidebar(state, { payload }) {
      state.sidebarOpen = payload ?? !state.sidebarOpen;
    },
    pushToast: {
      reducer(state, { payload }) {
        state.toasts.push(payload);
      },
      prepare(message, kind = 'info') {
        return { payload: { id: ++toastId, message, kind } };
      },
    },
    dropToast(state, { payload }) {
      state.toasts = state.toasts.filter((t) => t.id !== payload);
    },
  },
});

export const { setRegion, openCityModal, closeCityModal, openLogin, closeLogin, toggleSidebar, pushToast, dropToast } = ui.actions;
export default ui.reducer;
