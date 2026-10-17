import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled, { css } from 'styled-components';
import { errorCode, errorMessage, useCreateTxnMutation, useGetSeatGridQuery, useGetSeatLayoutPrimaryQuery } from '../store/api.js';
import { Button, EmptyState, ErrorState, Modal, Spinner } from '../components/ui.jsx';
import { ChevronLeft, EditIcon } from '../components/icons.jsx';
import { rupee, showDateLong } from '../lib/format.js';
import { Title } from '../lib/head.jsx';
import { pushToast } from '../store/ui.js';
import { track } from '../lib/clickstream.js';

const SOLD = 2;
const BEST = 4;

// BookMyShow swaps the vehicle illustration as you change the ticket count.
const vehicles = ['', '🚲', '🛵', '🛺', '🚗', '🚙', '🚙', '🚐', '🚐', '🚌', '🚌'];

const Top = styled.header`
  background: #fff;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 30;
  .row {
    max-width: 1240px;
    margin: 0 auto;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .back {
    background: none;
    border: none;
    padding: 4px;
  }
  h1 {
    font-size: 18px;
    font-weight: 500;
  }
  .sub {
    font-size: 13px;
    color: #666;
    margin-top: 2px;
  }
  .qty {
    margin-left: auto;
    border: 1px solid ${(p) => p.theme.brand};
    color: ${(p) => p.theme.brand};
    background: #fff;
    border-radius: 4px;
    padding: 6px 14px;
    font-size: 12px;
    display: flex;
    gap: 6px;
    align-items: center;
  }
  .strip {
    background: ${(p) => p.theme.bgSoft};
    border-top: 1px solid ${(p) => p.theme.border};
  }
  .strip .row {
    gap: 12px;
    overflow-x: auto;
    padding: 10px 16px;
  }
`;

const StripPill = styled.button`
  flex: 0 0 auto;
  width: 88px;
  height: 36px;
  border-radius: 2px;
  font-size: 12px;
  border: 1px solid ${(p) => (p.$status === 'FAST_FILLING' ? p.theme.amber : p.theme.green)};
  background: ${(p) => (p.$active ? p.theme.green : '#fff')};
  color: ${(p) => (p.$active ? '#fff' : '#666')};
  &:disabled {
    border-color: #ccc;
    color: #ccc;
  }
`;

const Hall = styled.main`
  max-width: 1100px;
  margin: 0 auto;
  padding: 32px 16px 160px;
  overflow-x: auto;
`;

const Grid = styled.div`
  display: grid;
  grid-template-columns: 40px repeat(${(p) => p.$cols}, 30px);
  row-gap: 8px;
  justify-content: center;
  min-width: max-content;
  .area {
    grid-column: 1 / -1;
    text-align: center;
    font-size: 13px;
    color: #666;
    border-bottom: 1px solid ${(p) => p.theme.border};
    padding: 24px 0 8px;
    margin-bottom: 12px;
    &:first-child {
      padding-top: 0;
    }
  }
  .label {
    grid-column: 1;
    font-size: 12px;
    color: #999;
    display: flex;
    align-items: center;
  }
`;

const SeatBtn = styled.button`
  width: 26px;
  height: 26px;
  margin: 0 2px;
  border-radius: 4px;
  font-size: 10px;
  padding: 0;
  border: 1px solid ${(p) => p.theme.green};
  background: #fff;
  color: ${(p) => p.theme.green};
  ${(p) => p.$best && css`border-color: #FFB700;`}
  ${(p) =>
    p.$selected &&
    css`
      background: ${p.theme.green};
      border-color: ${p.theme.green};
      color: #fff;
    `}
  &:hover:not(:disabled) {
    background: ${(p) => p.theme.green};
    color: #fff;
  }
  &:disabled {
    background: ${(p) => p.theme.sold};
    border-color: ${(p) => p.theme.sold};
    color: #fff;
    cursor: default;
  }
`;

const Screen = styled.div`
  margin: 48px auto 0;
  width: min(560px, 90%);
  text-align: center;
  font-size: 12px;
  color: #999;
  svg {
    width: 100%;
    margin-bottom: 8px;
  }
`;

const Foot = styled.footer`
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  background: #fff;
  border-top: 1px solid ${(p) => p.theme.border};
  z-index: 30;
  .legend {
    display: flex;
    justify-content: center;
    gap: 24px;
    font-size: 12px;
    color: #666;
    padding: 12px;
    span {
      display: inline-flex;
      align-items: center;
      gap: 6px;
    }
    i {
      width: 16px;
      height: 16px;
      border-radius: 3px;
      border: 1px solid ${(p) => p.theme.green};
      display: inline-block;
    }
    i.sold {
      background: ${(p) => p.theme.sold};
      border-color: ${(p) => p.theme.sold};
    }
    i.best {
      border-color: #ffb700;
    }
    i.sel {
      background: ${(p) => p.theme.green};
    }
  }
  .pay {
    display: flex;
    justify-content: center;
    padding: 12px 16px 16px;
    box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.06);
    button {
      width: min(360px, 100%);
    }
  }
`;

const Qty = styled.div`
  padding: 28px 24px 0;
  text-align: center;
  h3 {
    font-size: 18px;
    font-weight: 500;
  }
  .vehicle {
    font-size: 84px;
    height: 120px;
    line-height: 120px;
    margin: 12px 0;
  }
  .nums {
    display: flex;
    justify-content: center;
    gap: 6px;
    margin-bottom: 20px;
    flex-wrap: wrap;
    button {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      border: none;
      background: none;
      font-size: 16px;
      color: #333;
      &[data-on='true'] {
        background: ${(p) => p.theme.brand};
        color: #fff;
      }
      &:hover:not([data-on='true']) {
        background: ${(p) => p.theme.brandSoft};
      }
    }
  }
  .cats {
    display: flex;
    justify-content: space-around;
    border-top: 1px solid ${(p) => p.theme.border};
    padding: 16px 0;
    gap: 8px;
    div {
      font-size: 12px;
      color: #666;
      text-transform: uppercase;
    }
    b {
      display: block;
      font-size: 14px;
      color: #333;
      margin: 4px 0;
    }
    small {
      color: ${(p) => p.theme.green};
      font-size: 11px;
    }
    small[data-s='FAST_FILLING'] {
      color: ${(p) => p.theme.amber};
    }
    small[data-s='SOLD_OUT'] {
      color: #999;
    }
  }
  .best {
    background: ${(p) => p.theme.bgSoft};
    margin: 0 -24px;
    padding: 10px;
    font-size: 12px;
    color: #333;
  }
  .go {
    padding: 12px 0 16px;
  }
`;

export function SeatLayoutPage() {
  const { city, code, session, date } = useParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const primary = useGetSeatLayoutPrimaryQuery(session);
  const grid = useGetSeatGridQuery({ session }, { pollingInterval: 15000 });
  const [createTxn, { isLoading: locking }] = useCreateTxnMutation();
  const [qty, setQty] = useState(2);
  const [qtyOpen, setQtyOpen] = useState(true);
  const [selected, setSelected] = useState([]);

  // Reset the selection when the user switches shows.
  useEffect(() => {
    setSelected([]);
  }, [session]);

  // Drop selections that became unavailable during polling.
  const status = useMemo(() => {
    const m = {};
    grid.data?.areas.forEach((a) => a.rows.forEach((r) => r.seats.forEach((s) => (m[s.id] = s.s))));
    return m;
  }, [grid.data]);
  useEffect(() => {
    if (selected.some((id) => status[id] === SOLD)) {
      setSelected((sel) => sel.filter((id) => status[id] !== SOLD));
      dispatch(pushToast('Some seats you picked were just booked by someone else.', 'error'));
    }
  }, [status, selected, dispatch]);

  if (primary.isError || grid.isError) {
    return primary.error?.status === 404 ? <EmptyState title="Show not found" cta="Go back" onCta={() => navigate(-1)} /> : <ErrorState />;
  }
  if (!primary.data || !grid.data) return <Spinner />;

  const p = primary.data;
  const ev = p.event;
  const show = p.show;
  const priceOf = Object.fromEntries(grid.data.areas.map((a) => [a.code, a.price]));
  const areaOf = {};
  grid.data.areas.forEach((a) => a.rows.forEach((r) => r.seats.forEach((s) => (areaOf[s.id] = a.code))));
  const total = selected.reduce((sum, id) => sum + priceOf[areaOf[id]], 0);
  const showtimesUrl = `/movies/${city}/${ev.slug}/buytickets/${code}/${date}?lang=${encodeURIComponent(show.language)}&format=${encodeURIComponent(show.format)}`;

  const pick = (row, seat) => {
    if (seat.s === SOLD) return;
    if (selected.includes(seat.id)) {
      setSelected(selected.filter((x) => x !== seat.id));
      return;
    }
    const base = selected.length >= qty ? [] : selected;
    const need = qty - base.length;
    // Auto-select contiguous seats to the right, within the block.
    const idx = row.seats.findIndex((s) => s.id === seat.id);
    const run = [seat.id];
    for (let i = idx + 1; i < row.seats.length && run.length < need; i++) {
      const prev = row.seats[i - 1];
      const cur = row.seats[i];
      if (cur.c !== prev.c + 1 || cur.s === SOLD || base.includes(cur.id)) break;
      run.push(cur.id);
    }
    setSelected([...base, ...run]);
  };

  const proceed = async () => {
    track('seats_selected', { show_session_id: Number(session), qty: selected.length, amount: total });
    try {
      const txn = await createTxn({ sessionId: Number(session), seats: selected }).unwrap();
      navigate(`/booking/summary/${txn.id}`);
    } catch (err) {
      dispatch(pushToast(errorMessage(err), 'error'));
      if (errorCode(err) === 'SEATS_UNAVAILABLE') {
        setSelected([]);
        grid.refetch();
      }
    }
  };

  if (!p.bookable) {
    return (
      <EmptyState title="Bookings closed" message="Sorry, you cannot book tickets as you've exceeded the cinema's booking cutoff time." cta="Choose another show" onCta={() => navigate(showtimesUrl)} />
    );
  }

  const cols = grid.data.cols;
  return (
    <>
      <Title>{`${ev.title} - ${p.venue.name}`}</Title>
      <Top>
        <div className="row">
          <button className="back" onClick={() => navigate(showtimesUrl)} aria-label="Back to showtimes">
            <ChevronLeft size={24} />
          </button>
          <div>
            <h1>
              {ev.title} - {show.language}
            </h1>
            <div className="sub">
              {p.venue.name} | {showDateLong(show.startsAt)} | {show.time}
            </div>
          </div>
          <button className="qty" onClick={() => setQtyOpen(true)}>
            <EditIcon size={14} /> {qty} Tickets
          </button>
        </div>
        <div className="strip">
          <div className="row">
            {p.showTimes.map((s) => (
              <StripPill
                key={s.sessionId}
                $status={s.status}
                $active={String(s.sessionId) === session}
                disabled={s.status === 'SOLD_OUT'}
                onClick={() => navigate(`/movies/${city}/seat-layout/${code}/${p.venue.code}/${s.sessionId}/${date}`, { replace: true })}
              >
                {s.time}
              </StripPill>
            ))}
          </div>
        </div>
      </Top>

      <Hall>
        <Grid $cols={cols}>
          {grid.data.areas.map((a) => (
            <AreaBlock key={a.code} area={a} selected={selected} onPick={pick} />
          ))}
        </Grid>
        <Screen>
          <svg viewBox="0 0 560 40" aria-hidden="true">
            <defs>
              <linearGradient id="scr" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" stopColor="#B7D6F4" />
                <stop offset="1" stopColor="#E8F2FC" />
              </linearGradient>
            </defs>
            <path d="M20 4 H540 L520 36 H40 Z" fill="url(#scr)" stroke="#A8C8EA" />
          </svg>
          All eyes this way please
        </Screen>
      </Hall>

      <Foot>
        {selected.length > 0 ? (
          <div className="pay">
            <Button $size="lg" onClick={proceed} disabled={locking}>
              {locking ? 'Holding your seats…' : `Pay ${rupee(total)}`}
            </Button>
          </div>
        ) : (
          <div className="legend">
            <span>
              <i /> Available
            </span>
            <span>
              <i className="sold" /> Sold
            </span>
            <span>
              <i className="best" /> Bestseller
            </span>
            <span>
              <i className="sel" /> Selected
            </span>
          </div>
        )}
      </Foot>

      <Modal open={qtyOpen} onClose={() => setQtyOpen(false)} width={480} label="How many seats?">
        <Qty>
          <h3>How many seats?</h3>
          <div className="vehicle" aria-hidden="true">
            {vehicles[qty]}
          </div>
          <div className="nums">
            {Array.from({ length: p.maxTickets }, (_, i) => i + 1).map((n) => (
              <button key={n} data-on={n === qty} onClick={() => setQty(n)}>
                {n}
              </button>
            ))}
          </div>
          <div className="cats">
            {show.categories.map((c) => (
              <div key={c.code}>
                {c.name}
                <b>{rupee(c.price)}</b>
                <small data-s={c.status}>{c.status === 'FAST_FILLING' ? 'FILLING FAST' : c.status.replace('_', ' ')}</small>
              </div>
            ))}
          </div>
          {grid.data.areas.some((a) => a.rows.some((r) => r.seats.some((s) => s.s === BEST))) && (
            <div className="best">
              Book the <span style={{ border: '1px solid #FFB700', padding: '0 6px', borderRadius: 3 }}>Bestseller Seats</span> in this cinema at no extra cost!
            </div>
          )}
          <div className="go">
            <Button
              $block
              $size="lg"
              onClick={() => {
                setQtyOpen(false);
                if (selected.length > qty) setSelected(selected.slice(0, qty));
              }}
            >
              Select Seats
            </Button>
          </div>
        </Qty>
      </Modal>
    </>
  );
}

function AreaBlock({ area, selected, onPick }) {
  return (
    <>
      <div className="area">
        {rupee(area.price)} {area.name}
      </div>
      {area.rows.map((r) => (
        <Row key={r.label} row={r} selected={selected} onPick={onPick} />
      ))}
    </>
  );
}

function Row({ row, selected, onPick }) {
  return (
    <>
      <div className="label">{row.label}</div>
      {row.seats.map((s) => (
        <SeatBtn
          key={s.id}
          style={{ gridColumn: s.c + 2 }}
          disabled={s.s === SOLD}
          $best={s.s === BEST}
          $selected={selected.includes(s.id)}
          onClick={() => onPick(row, s)}
          aria-label={`Seat ${s.id}${s.s === SOLD ? ', sold' : ''}`}
          aria-pressed={selected.includes(s.id)}
        >
          {s.n}
        </SeatBtn>
      ))}
    </>
  );
}
