import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled from 'styled-components';
import { useGetShowtimesQuery, useGetShowtimesStaticQuery } from '../store/api.js';
import { Button, Container, EmptyState, ErrorState, Modal, Skeleton } from '../components/ui.jsx';
import { ChevronDown, FoodIcon, HeartIcon, InfoIcon, MoonIcon, SearchIcon, SunIcon, TicketIcon } from '../components/icons.jsx';
import { Legend, PillRow, ShowtimePill } from '../components/Showtime.jsx';
import { durationLabel } from '../lib/format.js';
import { Title } from '../lib/head.jsx';
import { openCityModal, pushToast } from '../store/ui.js';
import { useClickOutside } from '../lib/hooks.js';
import { track } from '../lib/clickstream.js';

const Head = styled.div`
  background: #fff;
  padding: 24px 0 16px;
  h1 {
    font-size: 34px;
    font-weight: 500;
    margin-bottom: 12px;
  }
  .tags {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    span {
      border: 1px solid #666;
      border-radius: 16px;
      font-size: 12px;
      padding: 2px 8px;
      color: #333;
    }
  }
`;

const Bar = styled.div`
  background: #fff;
  border-top: 1px solid ${(p) => p.theme.border};
  border-bottom: 1px solid ${(p) => p.theme.border};
  position: sticky;
  top: 0;
  z-index: 20;
  .inner {
    display: flex;
    align-items: stretch;
    min-height: 84px;
    @media (max-width: 900px) {
      flex-direction: column;
    }
  }
  .dates {
    display: flex;
    gap: 4px;
    align-items: center;
    padding: 8px 0;
    overflow-x: auto;
  }
  .filters {
    margin-left: auto;
    display: flex;
    align-items: stretch;
    border-left: 1px solid ${(p) => p.theme.border};
    @media (max-width: 900px) {
      margin-left: 0;
      overflow-x: auto;
      border-left: none;
      min-height: 44px;
    }
  }
`;

const DateBtn = styled.button`
  width: 56px;
  border: none;
  border-radius: 8px;
  padding: 8px 0;
  background: ${(p) => (p.$active ? p.theme.brand : 'transparent')};
  color: ${(p) => (p.$active ? '#fff' : p.disabled ? '#ccc' : '#333')};
  font-size: 12px;
  line-height: 1.5;
  b {
    display: block;
    font-size: 18px;
    font-weight: ${(p) => (p.$active ? 700 : 500)};
  }
`;

const DropWrap = styled.div`
  position: relative;
  border-right: 1px solid ${(p) => p.theme.border};
  > button {
    height: 100%;
    background: #fff;
    border: none;
    padding: 0 14px;
    font-size: 13px;
    color: #333;
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
    border-bottom: 2px solid ${(p) => (p.$active ? p.theme.brand : 'transparent')};
    svg {
      width: 14px;
      color: ${(p) => p.theme.brand};
    }
  }
  .menu {
    position: absolute;
    top: 100%;
    right: 0;
    min-width: 220px;
    background: #fff;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
    border-radius: 0 0 8px 8px;
    padding: 12px;
    z-index: 30;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .opt {
    border: 1px solid ${(p) => p.theme.border};
    background: #fff;
    border-radius: 6px;
    padding: 8px 12px;
    font-size: 12px;
    color: #333;
    text-align: left;
    small {
      display: block;
      color: #999;
      font-size: 10px;
    }
    &[data-on='true'] {
      border-color: ${(p) => p.theme.brand};
      color: ${(p) => p.theme.brand};
      background: ${(p) => p.theme.brandSoft};
    }
  }
`;

function Dropdown({ label, active, children }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);
  const close = useCallback(() => setOpen(false), []);
  useClickOutside(ref, close, open);
  return (
    <DropWrap ref={ref} $active={active}>
      <button onClick={() => setOpen((o) => !o)}>
        {label} <ChevronDown />
      </button>
      {open && <div className="menu" onClick={(e) => e.target.closest('.opt[data-close]') && setOpen(false)}>{children}</div>}
    </DropWrap>
  );
}

const List = styled.div`
  background: ${(p) => p.theme.bgSoft};
  padding: 16px 0 48px;
  .sub {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 0 16px;
    font-size: 11px;
    color: #666;
    .marks {
      display: flex;
      gap: 16px;
      span {
        display: inline-flex;
        gap: 4px;
        align-items: center;
      }
    }
  }
  .search {
    display: flex;
    align-items: center;
    gap: 8px;
    background: #fff;
    border-radius: 8px;
    padding: 8px 12px;
    margin-bottom: 12px;
    input {
      border: none;
      outline: none;
      flex: 1;
      font-size: 14px;
    }
  }
`;

const VenueRow = styled.div`
  background: #fff;
  display: flex;
  gap: 24px;
  padding: 24px;
  border-bottom: 1px solid ${(p) => p.theme.border};
  &:first-of-type {
    border-radius: 8px 8px 0 0;
  }
  .info {
    width: 340px;
    flex: 0 0 auto;
    display: flex;
    gap: 12px;
  }
  .logo {
    width: 36px;
    height: 36px;
    border-radius: 4px;
    border: 1px solid ${(p) => p.theme.border};
  }
  .name {
    font-size: 15px;
    font-weight: 500;
    color: #333;
    display: flex;
    gap: 6px;
    align-items: flex-start;
    a:hover {
      text-decoration: underline;
    }
    svg {
      flex: 0 0 auto;
      color: #999;
      cursor: help;
    }
  }
  .meta {
    font-size: 12px;
    color: #666;
    margin-top: 6px;
  }
  .icons {
    display: flex;
    gap: 10px;
    margin-top: 10px;
    span {
      display: inline-flex;
      gap: 4px;
      align-items: center;
      font-size: 11px;
      color: ${(p) => p.theme.green};
    }
    span.food {
      color: ${(p) => p.theme.amber};
    }
  }
  .heart {
    background: none;
    border: none;
    color: ${(p) => p.theme.light};
    margin-left: auto;
    align-self: flex-start;
    &[data-on='true'] {
      color: ${(p) => p.theme.brand};
      svg {
        fill: ${(p) => p.theme.brand};
      }
    }
  }
  @media (max-width: 900px) {
    flex-direction: column;
    .info {
      width: auto;
    }
  }
`;

const Warn = styled.div`
  padding: 28px;
  text-align: center;
  h3 {
    margin-bottom: 12px;
  }
  p {
    color: #666;
    line-height: 1.5;
    margin-bottom: 24px;
    font-size: 14px;
  }
  .actions {
    display: flex;
    gap: 12px;
    justify-content: center;
  }
`;

const FAVORITES_KEY = 'sb_fav_venues';
const loadFavs = () => {
  try {
    return JSON.parse(localStorage.getItem(FAVORITES_KEY) || '[]');
  } catch {
    return [];
  }
};

export function ShowtimesPage() {
  const { city, slug, code, date } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const q = new URLSearchParams(location.search);
  const args = { code, region: city, date: date || '', lang: q.get('lang') || '', format: q.get('format') || '' };
  const { data, isLoading, isFetching, isError, refetch } = useGetShowtimesQuery(args);
  const { data: statics } = useGetShowtimesStaticQuery();
  const [prices, setPrices] = useState([]);
  const [times, setTimes] = useState([]);
  const [sort, setSort] = useState('relevance');
  const [term, setTerm] = useState('');
  const [searching, setSearching] = useState(false);
  const [favs, setFavs] = useState([]);
  useEffect(() => setFavs(loadFavs()), []);
  const [adultGate, setAdultGate] = useState(null);

  const venues = useMemo(() => {
    if (!data) return [];
    const inPrice = (s) =>
      !prices.length || s.categories.some((c) => prices.some((b) => c.price >= b.from && c.price <= b.to));
    const hour = (s) => Number(new Intl.DateTimeFormat('en-GB', { hour: 'numeric', hour12: false, timeZone: 'Asia/Kolkata' }).format(new Date(s.startsAt)));
    const inTime = (s) => !times.length || times.some((t) => hour(s) >= t.from && hour(s) < t.to);
    let list = data.venues
      .map((v) => ({ ...v, showtimes: v.showtimes.filter((s) => inPrice(s) && inTime(s)) }))
      .filter((v) => v.showtimes.length && (!term || v.name.toLowerCase().includes(term.toLowerCase())));
    if (sort === 'name') list = [...list].sort((a, b) => a.name.localeCompare(b.name));
    // Favourite cinemas float to the top, as on BookMyShow.
    return [...list].sort((a, b) => favs.includes(b.code) - favs.includes(a.code));
  }, [data, prices, times, sort, term, favs]);

  if (isError) return <ErrorState onRetry={refetch} />;
  if (isLoading || !data) {
    return (
      <Container style={{ padding: 24 }}>
        <Skeleton $h="60px" $w="50%" />
        <div style={{ height: 24 }} />
        <Skeleton $h="300px" />
      </Container>
    );
  }

  const ev = data.event;
  const goDate = (d) => navigate(`/movies/${city}/${slug}/buytickets/${code}/${d}${location.search}`);
  const goFormat = (f) => navigate(`/movies/${city}/${slug}/buytickets/${code}/${data.dateCode}?lang=${encodeURIComponent(f.language)}&format=${encodeURIComponent(f.format)}`);
  const toggleFav = (vc) => {
    const next = favs.includes(vc) ? favs.filter((x) => x !== vc) : [...favs, vc];
    setFavs(next);
    try {
      localStorage.setItem(FAVORITES_KEY, JSON.stringify(next));
    } catch {
      /* ignore */
    }
  };
  const openShow = (v, s) => {
    track('showtime_selected', { event_code: code, venue_code: v.code, show_session_id: s.sessionId, show_time: s.time });
    const dest = `/movies/${city}/seat-layout/${code}/${v.code}/${s.sessionId}/${data.dateCode}`;
    if (ev.certificate === 'A') setAdultGate(dest);
    else navigate(dest);
  };
  const toggle = (list, setList, item) => setList(list.some((x) => x.code === item.code) ? list.filter((x) => x.code !== item.code) : [...list, item]);
  const selectedFormat = data.formats?.find((f) => f.selected);

  return (
    <>
      <Title>{`${ev.title} Movie Showtimes in ${data.region.name}`}</Title>
      <Head>
        <Container>
          <h1>
            <Link to={`/movies/${city}/${slug}/${code}`}>{ev.title}</Link>
          </h1>
          <div className="tags">
            <span>Movie runtime: {durationLabel(ev.durationMin)}</span>
            <span>{ev.certificate}</span>
            {ev.genres.map((g) => (
              <span key={g}>{g}</span>
            ))}
          </div>
        </Container>
      </Head>
      <Bar>
        <Container className="inner">
          <div className="dates">
            {data.dates.map((d) => (
              <DateBtn key={d.code} $active={d.code === data.dateCode} disabled={!d.enabled} onClick={() => goDate(d.code)}>
                {d.day}
                <b>{d.date}</b>
                {d.month}
              </DateBtn>
            ))}
          </div>
          <div className="filters">
            {data.formats?.length > 0 && (
              <Dropdown label={selectedFormat?.label || 'Format'} active>
                {data.formats.map((f) => (
                  <button key={f.label} className="opt" data-close data-on={f.selected} onClick={() => goFormat(f)}>
                    {f.label}
                  </button>
                ))}
              </Dropdown>
            )}
            <Dropdown label="Price Range" active={prices.length > 0}>
              {statics?.priceBuckets.map((b) => (
                <button key={b.code} className="opt" data-on={prices.some((x) => x.code === b.code)} onClick={() => toggle(prices, setPrices, b)}>
                  {b.label}
                </button>
              ))}
            </Dropdown>
            <Dropdown label="Preferred Time" active={times.length > 0}>
              {statics?.timeBuckets.map((b) => (
                <button key={b.code} className="opt" data-on={times.some((x) => x.code === b.code)} onClick={() => toggle(times, setTimes, b)}>
                  {b.label}
                  <small>{b.sub}</small>
                </button>
              ))}
            </Dropdown>
            <Dropdown label="Sort By" active={sort !== 'relevance'}>
              {[
                ['relevance', 'Relevance'],
                ['name', 'Cinema name (A-Z)'],
              ].map(([k, l]) => (
                <button key={k} className="opt" data-close data-on={sort === k} onClick={() => setSort(k)}>
                  {l}
                </button>
              ))}
            </Dropdown>
            <DropWrap>
              <button onClick={() => setSearching((s) => !s)} aria-label="Search cinemas">
                <SearchIcon size={20} />
              </button>
            </DropWrap>
          </div>
        </Container>
      </Bar>
      <List style={{ opacity: isFetching ? 0.6 : 1 }}>
        <Container>
          <div className="sub">
            <div className="marks">
              <span>
                <MoonIcon size={12} /> Late night shows
              </span>
              <span>
                <SunIcon size={12} /> Early morning shows
              </span>
            </div>
            <Legend>
              <span>Available</span>
              <span className="ff">Fast filling</span>
            </Legend>
          </div>
          {searching && (
            <label className="search">
              <SearchIcon size={16} />
              <input autoFocus placeholder="Search cinema by name or area" value={term} onChange={(e) => setTerm(e.target.value)} />
            </label>
          )}
          {data.venues.length === 0 ? (
            <EmptyState title={statics?.errors.noShows.title || 'No shows'} message={statics?.errors.noShows.message} cta="Change Location" onCta={() => dispatch(openCityModal())} />
          ) : venues.length === 0 ? (
            <EmptyState
              title={statics?.errors.filter.title}
              message={statics?.errors.filter.message}
              cta="Clear Filters"
              onCta={() => {
                setPrices([]);
                setTimes([]);
                setTerm('');
              }}
            />
          ) : (
            venues.map((v) => (
              <VenueRow key={v.code}>
                <div className="info">
                  <img className="logo" src={v.logo} alt="" />
                  <div>
                    <div className="name">
                      <Link to={`/cinemas/${city}/${v.slug}/buytickets/${v.code}/${data.dateCode}`}>{v.name}</Link>
                      <span title={v.address}>
                        <InfoIcon size={14} />
                      </span>
                    </div>
                    <div className="meta">{v.cancellation ? 'Cancellation available' : 'Non-cancellable'}</div>
                    <div className="icons">
                      {v.mticket && (
                        <span>
                          <TicketIcon size={14} /> M-Ticket
                        </span>
                      )}
                      {v.fnb && (
                        <span className="food">
                          <FoodIcon size={14} /> Food &amp; Beverage
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <PillRow>
                  {v.showtimes.map((s) => (
                    <ShowtimePill key={s.sessionId} show={s} onClick={() => openShow(v, s)} />
                  ))}
                </PillRow>
                <button className="heart" data-on={favs.includes(v.code)} onClick={() => toggleFav(v.code)} aria-label="Favourite cinema">
                  <HeartIcon />
                </button>
              </VenueRow>
            ))
          )}
        </Container>
      </List>
      <Modal open={!!adultGate} onClose={() => setAdultGate(null)} width={440} label="Content advisory">
        <Warn>
          <h3>This movie is rated A (Adults only)</h3>
          <p>You must be 18 years or older to watch this movie. Please carry a valid government ID proof — entry may be denied otherwise.</p>
          <div className="actions">
            <Button $variant="outline" onClick={() => setAdultGate(null)}>
              Cancel
            </Button>
            <Button
              onClick={() => {
                dispatch(pushToast('Please carry a valid ID proof to the cinema.'));
                navigate(adultGate);
              }}
            >
              Accept
            </Button>
          </div>
        </Warn>
      </Modal>
    </>
  );
}
