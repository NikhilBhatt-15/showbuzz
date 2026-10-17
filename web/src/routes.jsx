import { Navigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { api } from './store/api.js';
import { HomePage, LISTING_RE, ListingPage } from './pages/Discovery.jsx';
import { MovieDetailPage } from './pages/MovieDetail.jsx';
import { ShowtimesPage } from './pages/Showtimes.jsx';
import { SeatLayoutPage } from './pages/SeatLayout.jsx';
import { ConfirmationPage, OrderSummaryPage, OrdersPage, PaymentPage, ProfilePage } from './pages/Booking.jsx';
import {
  CinemasPage, EventBookingPage, EventDetailPage, NotFoundPage, OffersPage, SearchPage, VenueShowtimesPage,
} from './pages/Misc.jsx';
import { DEFAULT_REGION } from './lib/region.js';

function RootRedirect() {
  const slug = useSelector((s) => s.ui.regionSlug) || DEFAULT_REGION;
  return <Navigate to={`/explore/home/${slug}`} replace />;
}

const q = (endpoint, arg) => ({ dispatch }) => dispatch(api.endpoints[endpoint].initiate(arg));

// `chrome: false` routes (seat layout, checkout) render without the site
// header/footer, as on BookMyShow. `loadData` runs on the server so the page
// is rendered with its data and hydrated from window.__INITIAL_STATE__.
export const routes = [
  { path: '/', element: <RootRedirect /> },
  {
    path: '/explore/home/:city',
    element: <HomePage />,
    loadData: ({ dispatch, params }) => dispatch(api.endpoints.getDiscovery.initiate({ page: 'home', region: params.city })),
  },
  {
    path: '/explore/:listing',
    element: <ListingPage />,
    loadData: ({ dispatch, params, search }) => {
      const m = params.listing.match(LISTING_RE);
      if (m) dispatch(api.endpoints.getDiscovery.initiate({ page: m[1], region: m[2], search }));
    },
  },
  {
    path: '/movies/:city/:slug/:code',
    element: <MovieDetailPage />,
    loadData: ({ dispatch, params }) => {
      dispatch(api.endpoints.getSynopsisStatic.initiate(params.code));
      dispatch(api.endpoints.getSynopsisDynamic.initiate({ code: params.code, region: params.city }));
    },
  },
  {
    path: '/movies/:city/:slug/buytickets/:code/:date?',
    element: <ShowtimesPage />,
    loadData: ({ dispatch, params, query }) => {
      dispatch(api.endpoints.getShowtimesStatic.initiate());
      dispatch(
        api.endpoints.getShowtimes.initiate({
          code: params.code, region: params.city, date: params.date || '', lang: query.get('lang') || '', format: query.get('format') || '',
        }),
      );
    },
  },
  {
    path: '/movies/:city/seat-layout/:code/:venue/:session/:date',
    element: <SeatLayoutPage />,
    chrome: false,
    loadData: ({ dispatch, params }) => {
      dispatch(api.endpoints.getSeatLayoutPrimary.initiate(params.session));
      dispatch(api.endpoints.getSeatGrid.initiate({ session: params.session }));
    },
  },
  { path: '/:city/cinemas', element: <CinemasPage />, loadData: ({ dispatch, params }) => dispatch(api.endpoints.getCinemas.initiate(params.city)) },
  {
    path: '/cinemas/:city/:slug/buytickets/:venue/:date?',
    element: <VenueShowtimesPage />,
    loadData: ({ dispatch, params }) => dispatch(api.endpoints.getVenueShowtimes.initiate({ venue: params.venue, date: params.date || '' })),
  },
  {
    path: '/events/:slug/:code',
    element: <EventDetailPage />,
    loadData: ({ dispatch, params, region }) => dispatch(api.endpoints.getEvent.initiate({ code: params.code, region })),
  },
  {
    path: '/events/:slug/:code/book',
    element: <EventBookingPage />,
    loadData: ({ dispatch, params, region }) => {
      dispatch(api.endpoints.getEvent.initiate({ code: params.code, region }));
      dispatch(api.endpoints.getEventInventory.initiate(params.code));
    },
  },
  { path: '/booking/summary/:txn', element: <OrderSummaryPage />, chrome: false, loadData: ({ dispatch, params }) => q('getTxn', params.txn)({ dispatch }) },
  { path: '/booking/payment/:txn', element: <PaymentPage />, chrome: false, loadData: ({ dispatch, params }) => q('getTxn', params.txn)({ dispatch }) },
  { path: '/booking/confirmation/:id', element: <ConfirmationPage />, loadData: ({ dispatch, params }) => q('getBooking', params.id)({ dispatch }) },
  { path: '/my-profile/purchase-history', element: <OrdersPage />, loadData: q('getBookings', undefined) },
  { path: '/my-profile/edit', element: <ProfilePage /> },
  { path: '/search', element: <SearchPage /> },
  { path: '/offers', element: <OffersPage /> },
  { path: '*', element: <NotFoundPage /> },
];
