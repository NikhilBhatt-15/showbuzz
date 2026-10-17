import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

// On the server, requests go to the gateway over the Docker network and the
// visitor's cookie is forwarded; in the browser they are same-origin.
const apiBase = import.meta.env.SSR ? `${process.env.API_INTERNAL_URL || 'http://localhost:8080'}/api` : '/api';

const baseQuery = fetchBaseQuery({
  baseUrl: apiBase,
  credentials: 'include',
  prepareHeaders: (headers, { extra }) => {
    if (extra?.cookie) headers.set('cookie', extra.cookie);
    return headers;
  },
});

const qs = (obj) =>
  Object.entries(obj)
    .filter(([, v]) => v !== undefined && v !== null && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join('&');

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: ['Me', 'Bookings', 'Txn', 'SeatGrid'],
  keepUnusedDataFor: 60,
  endpoints: (b) => ({
    // catalog
    getRegions: b.query({ query: () => '/regions', keepUnusedDataFor: 3600 }),
    getDiscovery: b.query({
      query: ({ page, region, search = '' }) => `/explore/v1/discover/${page}?region=${region}${search ? `&${search}` : ''}`,
    }),
    getSynopsisStatic: b.query({ query: (code) => `/movies/v1/synopsis/init/static?eventCode=${code}` }),
    getSynopsisDynamic: b.query({ query: ({ code, region }) => `/movies/v1/synopsis/init/dynamic?${qs({ eventCode: code, regionCode: region })}` }),
    getEvent: b.query({ query: ({ code, region }) => `/events/v1/${code}?${qs({ regionCode: region })}` }),
    search: b.query({ query: ({ q, region }) => `/search/v1?${qs({ q, regionCode: region })}` }),

    // showtimes
    getShowtimesStatic: b.query({ query: () => '/movies-data/v4/showtimes-by-event/static', keepUnusedDataFor: 3600 }),
    getShowtimes: b.query({
      query: ({ code, region, date, lang, format }) =>
        `/movies-data/v4/showtimes-by-event/primary-dynamic?${qs({ eventCode: code, regionCode: region, dateCode: date, lang, format })}`,
    }),
    getSeatLayoutPrimary: b.query({ query: (sessionId) => `/movies-data/seatlayout/v1/primary?sessionId=${sessionId}` }),
    getCinemas: b.query({ query: (region) => `/movies-data/v1/cinemas?regionCode=${region}` }),
    getVenueShowtimes: b.query({ query: ({ venue, date }) => `/movies-data/v1/venue-showtimes?${qs({ venueCode: venue, dateCode: date })}` }),

    // transactions
    getSeatGrid: b.query({
      query: ({ session, txn }) => `/trans/v1/seat-layout/${session}${txn ? `?txn=${txn}` : ''}`,
      providesTags: ['SeatGrid'],
      keepUnusedDataFor: 0,
    }),
    getFnb: b.query({ query: () => '/trans/v1/fnb' }),
    getEventInventory: b.query({ query: (code) => `/trans/v1/events/${code}/inventory`, keepUnusedDataFor: 0 }),
    createTxn: b.mutation({ query: (body) => ({ url: '/trans/v1/transactions', method: 'POST', body }) }),
    createEventTxn: b.mutation({ query: (body) => ({ url: '/trans/v1/event-transactions', method: 'POST', body }) }),
    getTxn: b.query({ query: (id) => `/trans/v1/transactions/${id}`, providesTags: ['Txn'], keepUnusedDataFor: 0 }),
    updateTxn: b.mutation({
      query: ({ id, ...body }) => ({ url: `/trans/v1/transactions/${id}`, method: 'PATCH', body }),
      async onQueryStarted({ id }, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(api.util.upsertQueryData('getTxn', id, data));
        } catch {
          /* surfaced by the caller */
        }
      },
    }),
    cancelTxn: b.mutation({ query: (id) => ({ url: `/trans/v1/transactions/${id}`, method: 'DELETE' }), invalidatesTags: ['SeatGrid'] }),
    pay: b.mutation({
      query: ({ id, ...body }) => ({ url: `/trans/v1/transactions/${id}/pay`, method: 'POST', body }),
      invalidatesTags: ['Bookings'],
    }),
    getBookings: b.query({ query: () => '/trans/v1/bookings', providesTags: ['Bookings'] }),
    getBooking: b.query({ query: (id) => `/trans/v1/bookings/${id}`, providesTags: ['Bookings'] }),
    cancelBooking: b.mutation({ query: (id) => ({ url: `/trans/v1/bookings/${id}/cancel`, method: 'POST' }), invalidatesTags: ['Bookings'] }),

    // user — a 401 simply means "signed out"
    getMe: b.query({
      async queryFn(_arg, api, _extra, fetchWithBQ) {
        const res = await fetchWithBQ('/user/v1/me');
        if (res.error?.status === 401) return { data: { user: null } };
        return res.error ? { error: res.error } : { data: res.data };
      },
      providesTags: ['Me'],
    }),
    requestOtp: b.mutation({ query: (body) => ({ url: '/user/v1/otp', method: 'POST', body }) }),
    verifyOtp: b.mutation({ query: (body) => ({ url: '/user/v1/verify', method: 'POST', body }), invalidatesTags: ['Me', 'Bookings'] }),
    updateMe: b.mutation({ query: (body) => ({ url: '/user/v1/me', method: 'PUT', body }), invalidatesTags: ['Me'] }),
    logout: b.mutation({ query: () => ({ url: '/user/v1/logout', method: 'POST' }), invalidatesTags: ['Me', 'Bookings'] }),
  }),
});

export const {
  useGetRegionsQuery,
  useGetDiscoveryQuery,
  useGetSynopsisStaticQuery,
  useGetSynopsisDynamicQuery,
  useGetEventQuery,
  useSearchQuery,
  useLazySearchQuery,
  useGetShowtimesStaticQuery,
  useGetShowtimesQuery,
  useLazyGetShowtimesQuery,
  useGetSeatLayoutPrimaryQuery,
  useGetCinemasQuery,
  useGetVenueShowtimesQuery,
  useGetSeatGridQuery,
  useGetFnbQuery,
  useGetEventInventoryQuery,
  useCreateTxnMutation,
  useCreateEventTxnMutation,
  useGetTxnQuery,
  useUpdateTxnMutation,
  useCancelTxnMutation,
  usePayMutation,
  useGetBookingsQuery,
  useGetBookingQuery,
  useCancelBookingMutation,
  useGetMeQuery,
  useRequestOtpMutation,
  useVerifyOtpMutation,
  useUpdateMeMutation,
  useLogoutMutation,
} = api;

export const errorMessage = (err, fallback = 'Something went wrong. Please try again.') =>
  err?.data?.error?.message || fallback;
export const errorCode = (err) => err?.data?.error?.code;
