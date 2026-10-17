import { useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled from 'styled-components';
import {
  errorMessage, useCreateEventTxnMutation, useGetCinemasQuery, useGetEventInventoryQuery, useGetEventQuery, useGetVenueShowtimesQuery, useSearchQuery,
} from '../store/api.js';
import { Button, Card, Container, EmptyState, ErrorState, Skeleton, Spinner } from '../components/ui.jsx';
import { ChevronDown, ChevronUp, ClockIcon, FoodIcon, HeartIcon, InfoIcon, PinIcon, TicketIcon, UserIcon } from '../components/icons.jsx';
import { PillRow, ShowtimePill } from '../components/Showtime.jsx';
import { durationLabel, rupee, showDateLong, showTime } from '../lib/format.js';
import { Title } from '../lib/head.jsx';
import { useRegion } from '../lib/region.js';
import { pushToast } from '../store/ui.js';

const Page = styled.div`
  background: ${(p) => p.theme.bg};
  padding: 32px 0 64px;
  min-height: 60vh;
`;

// ---- event detail ----------------------------------------------------------

const EventLayout = styled(Container)`
  h1 {
    font-size: 32px;
    margin-bottom: 8px;
  }
  .tags {
    display: flex;
    gap: 8px;
    margin-bottom: 20px;
    span {
      background: #e5e5e5;
      border-radius: 4px;
      font-size: 13px;
      padding: 2px 8px;
    }
  }
  .grid {
    display: grid;
    grid-template-columns: 1fr 380px;
    gap: 32px;
    align-items: start;
    @media (max-width: 960px) {
      grid-template-columns: 1fr;
    }
  }
  .hero {
    border-radius: 12px;
    width: 100%;
    aspect-ratio: 16 / 9;
    object-fit: cover;
  }
  section {
    margin-top: 32px;
    h2 {
      font-size: 22px;
      margin-bottom: 12px;
    }
    p {
      line-height: 1.6;
      color: #333;
      font-size: 15px;
    }
  }
  .know {
    background: #fff;
    border-radius: 8px;
    padding: 4px 20px;
    li {
      padding: 10px 0;
      color: #333;
      font-size: 14px;
      border-bottom: 1px solid #f0f0f0;
    }
    li:last-child {
      border-bottom: none;
    }
  }
`;

const InfoCard = styled(Card)`
  padding: 24px;
  position: sticky;
  top: 16px;
  .row {
    display: flex;
    gap: 12px;
    align-items: center;
    font-size: 14px;
    color: #333;
    margin-bottom: 16px;
    svg {
      color: #666;
      flex: 0 0 auto;
    }
  }
  .price {
    border-top: 1px solid ${(p) => p.theme.border};
    padding-top: 16px;
    margin-top: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    b {
      font-size: 18px;
    }
    small {
      display: block;
      color: ${(p) => p.theme.amber};
      font-size: 11px;
    }
  }
`;

export function EventDetailPage() {
  const { slug, code } = useParams();
  const region = useRegion();
  const navigate = useNavigate();
  const { data, isLoading, isError, error } = useGetEventQuery({ code, region: region.slug });
  const [open, setOpen] = useState(true);
  if (isLoading) return <Spinner />;
  if (isError) return error?.status === 404 ? <EmptyState title="Event not found" cta="Go Home" onCta={() => navigate('/')} /> : <ErrorState />;
  const e = data.event;
  return (
    <Page>
      <Title>{`${e.title} | ${data.city}`}</Title>
      <EventLayout>
        <h1>{e.title}</h1>
        <div className="tags">
          <span>{e.genre}</span>
          <span>{e.category.charAt(0) + e.category.slice(1).toLowerCase()}</span>
        </div>
        <div className="grid">
          <div>
            <img className="hero" src={data.backdropUrl} alt={e.title} />
            <section>
              <h2>About The Event</h2>
              <p>{e.synopsis}</p>
            </section>
            <section>
              <h2 style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }} onClick={() => setOpen((o) => !o)}>
                Things to Know {open ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
              </h2>
              {open && (
                <ul className="know">
                  {data.thingsToKnow.map((t) => (
                    <li key={t}>{t}</li>
                  ))}
                </ul>
              )}
            </section>
          </div>
          <InfoCard>
            <div className="row">
              <TicketIcon /> {showDateLong(e.startsAt)}
            </div>
            <div className="row">
              <ClockIcon /> {showTime(e.startsAt)}
            </div>
            <div className="row">
              <ClockIcon /> {durationLabel(e.durationMin)}
            </div>
            <div className="row">
              <UserIcon /> Age Limit - {e.ageLimit}
            </div>
            <div className="row">
              <InfoIcon /> {e.languages.join(', ')}
            </div>
            <div className="row">
              <PinIcon /> {data.venue}
            </div>
            <div className="price">
              <div>
                <b>{rupee(data.priceFrom)} onwards</b>
                <small>Filling fast</small>
              </div>
              <Button onClick={() => navigate(`/events/${slug}/${code}/book`)}>Book Now</Button>
            </div>
          </InfoCard>
        </div>
      </EventLayout>
    </Page>
  );
}

// ---- event ticket selection ------------------------------------------------

const Tiers = styled.div`
  max-width: 760px;
  margin: 0 auto;
  h1 {
    font-size: 22px;
    margin-bottom: 4px;
  }
  .sub {
    color: #666;
    margin-bottom: 24px;
  }
  .tier {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    gap: 16px;
    border: 1px solid transparent;
    &[data-on='true'] {
      border-color: ${(p) => p.theme.brand};
    }
    .n {
      font-size: 16px;
      font-weight: 500;
    }
    .note {
      font-size: 12px;
      color: #666;
      margin-top: 4px;
    }
    .ff {
      color: ${(p) => p.theme.amber};
      font-size: 11px;
      margin-top: 4px;
    }
    .so {
      color: #999;
      font-size: 12px;
    }
    .right {
      margin-left: auto;
      display: flex;
      align-items: center;
      gap: 16px;
    }
  }
  .step {
    display: inline-flex;
    align-items: center;
    border: 1px solid ${(p) => p.theme.brand};
    border-radius: 4px;
    button {
      width: 32px;
      height: 32px;
      background: none;
      border: none;
      color: ${(p) => p.theme.brand};
      font-size: 18px;
    }
    span {
      min-width: 24px;
      text-align: center;
    }
  }
  .bar {
    position: sticky;
    bottom: 0;
    background: #fff;
    margin-top: 24px;
    padding: 16px 20px;
    border-radius: 8px;
    box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.08);
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
`;

export function EventBookingPage() {
  const { code } = useParams();
  const region = useRegion();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const ev = useGetEventQuery({ code, region: region.slug });
  const inv = useGetEventInventoryQuery(code);
  const [sel, setSel] = useState({ tier: '', qty: 0 });
  const [create, { isLoading }] = useCreateEventTxnMutation();
  if (ev.isLoading || inv.isLoading) return <Spinner />;
  if (ev.isError || inv.isError) return <ErrorState />;
  const e = ev.data.event;
  const tiers = inv.data.tiers;
  const chosen = tiers.find((t) => t.code === sel.tier);
  const setQty = (t, qty) => setSel({ tier: qty > 0 ? t.code : '', qty: Math.max(0, Math.min(qty, t.maxPerBooking)) });
  const proceed = async () => {
    try {
      const txn = await create({ eventCode: code, tier: sel.tier, qty: sel.qty, regionName: region.name }).unwrap();
      navigate(`/booking/summary/${txn.id}`);
    } catch (err) {
      dispatch(pushToast(errorMessage(err), 'error'));
      inv.refetch();
    }
  };
  return (
    <Page>
      <Title>{`Book tickets - ${e.title}`}</Title>
      <Container>
        <Tiers>
          <h1>{e.title}</h1>
          <p className="sub">
            {ev.data.venue} | {showDateLong(e.startsAt)} | {showTime(e.startsAt)}
          </p>
          {tiers.map((t) => (
            <div className="tier" key={t.code} data-on={t.code === sel.tier}>
              <div>
                <div className="n">{t.name}</div>
                {t.note && <div className="note">{t.note}</div>}
                {t.fastFilling && <div className="ff">Filling fast</div>}
              </div>
              <div className="right">
                <b>{rupee(t.price)}</b>
                {t.available === 0 ? (
                  <span className="so">Sold out</span>
                ) : t.code === sel.tier ? (
                  <span className="step">
                    <button onClick={() => setQty(t, sel.qty - 1)} aria-label="Remove ticket">−</button>
                    <span>{sel.qty}</span>
                    <button onClick={() => setQty(t, sel.qty + 1)} aria-label="Add ticket">+</button>
                  </span>
                ) : (
                  <Button $variant="outline" $size="sm" onClick={() => setQty(t, 1)}>
                    Add
                  </Button>
                )}
              </div>
            </div>
          ))}
          {chosen && (
            <div className="bar">
              <span>
                {sel.qty} x {chosen.name}
                <br />
                <b>{rupee(chosen.price * sel.qty)}</b>
              </span>
              <Button $size="lg" onClick={proceed} disabled={isLoading}>
                {isLoading ? 'Holding tickets…' : 'Proceed'}
              </Button>
            </div>
          )}
        </Tiers>
      </Container>
    </Page>
  );
}

// ---- cinemas ---------------------------------------------------------------

const Venues = styled.div`
  background: #fff;
  border-radius: 8px;
  .v {
    display: flex;
    gap: 16px;
    padding: 20px 24px;
    border-bottom: 1px solid ${(p) => p.theme.border};
    align-items: flex-start;
    img {
      width: 40px;
      height: 40px;
      border: 1px solid ${(p) => p.theme.border};
      border-radius: 4px;
    }
    .n {
      font-size: 16px;
      font-weight: 500;
    }
    .m {
      color: #666;
      font-size: 12px;
      margin-top: 6px;
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }
    .addr {
      color: #999;
      font-size: 12px;
      margin-top: 6px;
    }
    .go {
      margin-left: auto;
      color: ${(p) => p.theme.brand};
      font-size: 13px;
      white-space: nowrap;
    }
  }
`;

export function CinemasPage() {
  const { city } = useParams();
  const region = useRegion();
  const { data, isLoading, isError } = useGetCinemasQuery(city);
  if (isLoading) return <Spinner />;
  if (isError) return <ErrorState />;
  return (
    <Page>
      <Title>{`Cinemas in ${region.name}`}</Title>
      <Container>
        <h1 style={{ fontSize: 24, marginBottom: 20 }}>Cinemas in {data.region.name}</h1>
        <Venues>
          {data.venues.map((v) => (
            <div className="v" key={v.code}>
              <HeartIcon size={18} color="#bbb" />
              <img src={v.logo} alt="" />
              <div>
                <Link className="n" to={`/cinemas/${city}/${v.slug}/buytickets/${v.code}/${data.dateCode}`}>
                  {v.name}
                </Link>
                <div className="m">
                  <span>{v.cancellation ? 'Cancellation available' : 'Non-cancellable'}</span>
                  {v.mticket && (
                    <span style={{ color: '#1EA83C' }}>
                      <TicketIcon size={12} /> M-Ticket
                    </span>
                  )}
                  {v.fnb && (
                    <span style={{ color: '#FFA426' }}>
                      <FoodIcon size={12} /> Food &amp; Beverage
                    </span>
                  )}
                </div>
                <div className="addr">{v.address}</div>
              </div>
              <Link className="go" to={`/cinemas/${city}/${v.slug}/buytickets/${v.code}/${data.dateCode}`}>
                View showtimes ›
              </Link>
            </div>
          ))}
        </Venues>
      </Container>
    </Page>
  );
}

const DateRow = styled.div`
  display: flex;
  gap: 6px;
  margin: 20px 0;
  overflow-x: auto;
  button {
    width: 56px;
    border: none;
    border-radius: 8px;
    padding: 8px 0;
    background: #fff;
    font-size: 12px;
    line-height: 1.5;
    b {
      display: block;
      font-size: 18px;
    }
    &[data-on='true'] {
      background: ${(p) => p.theme.brand};
      color: #fff;
    }
    &:disabled {
      color: #ccc;
    }
  }
`;

const MovieGroup = styled.div`
  background: #fff;
  padding: 20px 24px;
  border-bottom: 1px solid ${(p) => p.theme.border};
  &:first-of-type {
    border-radius: 8px 8px 0 0;
  }
  .t {
    font-size: 16px;
    font-weight: 500;
    margin-bottom: 4px;
  }
  .m {
    color: #666;
    font-size: 12px;
    margin-bottom: 16px;
  }
`;

export function VenueShowtimesPage() {
  const { city, slug, venue, date } = useParams();
  const navigate = useNavigate();
  const { data, isLoading, isError, isFetching } = useGetVenueShowtimesQuery({ venue, date: date || '' });
  if (isLoading) return <Spinner />;
  if (isError) return <ErrorState />;
  const v = data.venue;
  return (
    <Page>
      <Title>{`${v.name} - Showtimes`}</Title>
      <Container style={{ opacity: isFetching ? 0.6 : 1 }}>
        <div style={{ display: 'flex', gap: 16, alignItems: 'center' }}>
          <img src={v.logo} alt="" style={{ width: 56, height: 56, borderRadius: 8 }} />
          <div>
            <h1 style={{ fontSize: 24 }}>{v.name}</h1>
            <div style={{ color: '#666', fontSize: 13, marginTop: 4 }}>{v.address}</div>
          </div>
        </div>
        <DateRow>
          {data.dates.map((d) => (
            <button key={d.code} data-on={d.code === data.dateCode} disabled={!d.enabled} onClick={() => navigate(`/cinemas/${city}/${slug}/buytickets/${venue}/${d.code}`)}>
              {d.day}
              <b>{d.date}</b>
              {d.month}
            </button>
          ))}
        </DateRow>
        {(data.movies || []).length === 0 ? (
          <EmptyState title="No shows on this date" message="Try another date." />
        ) : (
          data.movies.map((g) => (
            <MovieGroup key={g.movie.code + g.language + g.format}>
              <div className="t">
                <Link to={`/movies/${city}/${g.movie.slug}/${g.movie.code}`}>{g.movie.title}</Link>
              </div>
              <div className="m">
                {g.movie.certificate} • {g.language} • {g.format}
              </div>
              <PillRow>
                {g.showtimes.map((s) => (
                  <ShowtimePill key={s.sessionId} show={s} onClick={() => navigate(`/movies/${city}/seat-layout/${g.movie.code}/${venue}/${s.sessionId}/${data.dateCode}`)} />
                ))}
              </PillRow>
            </MovieGroup>
          ))
        )}
      </Container>
    </Page>
  );
}

// ---- search, offers, 404 ---------------------------------------------------

const Results = styled.div`
  max-width: 760px;
  a {
    display: flex;
    gap: 16px;
    align-items: center;
    background: #fff;
    padding: 12px 16px;
    border-radius: 8px;
    margin-bottom: 8px;
    img {
      width: 48px;
      aspect-ratio: 400 / 660;
      object-fit: cover;
      border-radius: 4px;
    }
    small {
      display: block;
      color: #999;
      margin-top: 2px;
    }
    .type {
      margin-left: auto;
      font-size: 11px;
      color: ${(p) => p.theme.brand};
      text-transform: uppercase;
    }
  }
`;

export function SearchPage() {
  const location = useLocation();
  const region = useRegion();
  const q = new URLSearchParams(location.search).get('q') || '';
  const { data, isFetching } = useSearchQuery({ q, region: region.slug }, { skip: q.length < 1 });
  return (
    <Page>
      <Title>{`Search: ${q}`}</Title>
      <Container>
        <h1 style={{ fontSize: 22, marginBottom: 20 }}>Results for “{q}”</h1>
        {isFetching && !data ? (
          <Skeleton $h="200px" />
        ) : !data?.results.length ? (
          <EmptyState art="search" title="Sorry! No results found." message="We got no results for your search. Kindly retry or search for something else." />
        ) : (
          <Results>
            {data.results.map((r) => (
              <Link key={r.url} to={r.url}>
                {r.image ? <img src={r.image} alt="" /> : <span style={{ width: 48 }} />}
                <span>
                  {r.title}
                  <small>{r.subtitle}</small>
                </span>
                <span className="type">{r.type}</span>
              </Link>
            ))}
          </Results>
        )}
      </Container>
    </Page>
  );
}

const OfferGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
  .o {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    border-left: 4px solid ${(p) => p.theme.brand};
    b {
      display: block;
      margin-bottom: 6px;
    }
    span {
      color: #666;
      font-size: 13px;
      line-height: 1.5;
    }
  }
`;

export function OffersPage() {
  const offers = [
    ['FIRST50', 'Flat 50% off on your first booking, up to ₹150.'],
    ['SBPAY100', 'Flat ₹100 off on 2 or more tickets when you pay with SB Pay wallet.'],
    ['FRIDAYBOGO', 'Buy 1 get 1 free on movie tickets every Friday with Platinum Credit Card.'],
    ['UPI75', 'Get up to ₹75 cashback on your first UPI payment.'],
    ['COMBO20', '20% off on F&B combos when booked with your tickets.'],
    ['EVENTS10', '10% off on all comedy shows this month.'],
  ];
  return (
    <Page>
      <Title>Offers</Title>
      <Container>
        <h1 style={{ fontSize: 24, marginBottom: 8 }}>Offers &amp; Gift Cards</h1>
        <p style={{ color: '#666', marginBottom: 24 }}>Demo offers — these codes are illustrative and not applied at checkout.</p>
        <OfferGrid>
          {offers.map(([c, t]) => (
            <div className="o" key={c}>
              <b>{c}</b>
              <span>{t}</span>
            </div>
          ))}
        </OfferGrid>
      </Container>
    </Page>
  );
}

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <Page>
      <Title status={404}>Page not found</Title>
      <EmptyState title="Uh-oh! We couldn't find that page." message="The link may be broken, or the page may have been removed." cta="Go to Home" onCta={() => navigate('/')} />
    </Page>
  );
}

