import { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled from 'styled-components';
import {
  errorCode, errorMessage, useCancelBookingMutation, useCancelTxnMutation, useGetBookingQuery, useGetBookingsQuery, useGetFnbQuery,
  useGetMeQuery, useGetTxnQuery, usePayMutation, useUpdateMeMutation, useUpdateTxnMutation,
} from '../store/api.js';
import { Button, Card, Container, EmptyState, ErrorState, Modal, Spinner } from '../components/ui.jsx';
import { CardIcon, CheckIcon, ChevronDown, ChevronLeft, ChevronUp, ClockIcon, ShieldIcon } from '../components/icons.jsx';
import { mmss, rs, rupee, seatSummary, showDateLong, showTime } from '../lib/format.js';
import { useCountdown } from '../lib/hooks.js';
import { Title } from '../lib/head.jsx';
import { openLogin, pushToast } from '../store/ui.js';
import { track } from '../lib/clickstream.js';
import { useRegionSlug } from '../lib/region.js';

// ---- shared chrome ---------------------------------------------------------

const TopBar = styled.header`
  background: ${(p) => p.theme.dark};
  color: #fff;
  .row {
    max-width: 1240px;
    margin: 0 auto;
    padding: 14px 16px;
    display: flex;
    align-items: center;
    gap: 16px;
  }
  button.back {
    background: none;
    border: none;
    color: #fff;
  }
  h1 {
    font-size: 18px;
    font-weight: 500;
  }
  .sub {
    font-size: 12px;
    color: #ccc;
    margin-top: 2px;
  }
  .timer {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    background: rgba(255, 255, 255, 0.1);
    padding: 6px 12px;
    border-radius: 20px;
    &[data-low='true'] {
      background: ${(p) => p.theme.brand};
    }
  }
`;

function BookingTop({ txn, onBack }) {
  const left = useCountdown(txn?.expiresAt, txn?.serverTime);
  const s = txn?.snapshot;
  return (
    <TopBar>
      <div className="row">
        {onBack && (
          <button className="back" onClick={onBack} aria-label="Back">
            <ChevronLeft size={24} />
          </button>
        )}
        <div>
          <h1>{s?.title}</h1>
          <div className="sub">
            {s?.venueName} {s && `| ${showDateLong(s.startsAt)} | ${showTime(s.startsAt)}`}
          </div>
        </div>
        {left !== null && (
          <div className="timer" data-low={left < 120} role="timer">
            <ClockIcon size={14} /> {mmss(left)} left
          </div>
        )}
      </div>
    </TopBar>
  );
}

function ExpiredModal({ open, onOk }) {
  return (
    <Modal open={open} closable={false} width={420} label="Session expired">
      <div style={{ padding: 28, textAlign: 'center' }}>
        <div style={{ fontSize: 48 }}>⏰</div>
        <h3 style={{ margin: '12px 0' }}>Your session has expired</h3>
        <p style={{ color: '#666', marginBottom: 24, lineHeight: 1.5 }}>The seats you selected have been released. Please select your seats again.</p>
        <Button onClick={onOk}>Select seats again</Button>
      </div>
    </Modal>
  );
}

function useTxn(id) {
  const q = useGetTxnQuery(id);
  const left = useCountdown(q.data?.expiresAt, q.data?.serverTime);
  const expired = q.error?.status === 410 || (left !== null && left <= 0);
  return { ...q, expired };
}

function retryPath(txn, city) {
  const s = txn?.snapshot;
  if (!s) return '/';
  if (txn.kind === 'EVENT') return `/events/${s.slug}/${s.eventCode}/book`;
  return `/movies/${city}/seat-layout/${s.eventCode}/${s.venueCode}/${txn.sessionId}/${s.dateCode}`;
}

const Page = styled.div`
  background: ${(p) => p.theme.bg};
  min-height: calc(100vh - 64px);
  padding: 24px 0 64px;
`;

const Cols = styled(Container)`
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 24px;
  align-items: start;
  @media (max-width: 960px) {
    grid-template-columns: 1fr;
  }
`;

// ---- order summary ---------------------------------------------------------

const Summary = styled(Card)`
  padding: 20px 20px 0;
  position: sticky;
  top: 16px;
  h4 {
    font-size: 12px;
    letter-spacing: 2px;
    color: ${(p) => p.theme.brand};
    margin-bottom: 16px;
  }
  .line {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    font-size: 14px;
    color: #333;
    margin-bottom: 4px;
  }
  .mut {
    font-size: 12px;
    color: #666;
    margin-bottom: 12px;
  }
  .fee {
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .breakup {
    background: ${(p) => p.theme.bg};
    padding: 8px 12px;
    margin: 4px 0 12px;
    border-radius: 4px;
    font-size: 12px;
    color: #666;
    .line {
      font-size: 12px;
      color: #666;
    }
  }
  hr {
    border: none;
    border-top: 1px dashed #ccc;
    margin: 12px 0;
  }
  .smile {
    background: ${(p) => p.theme.bgSoft};
    border-radius: 6px;
    padding: 10px 12px;
    font-size: 12px;
    color: #333;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: 12px 0;
    button {
      background: none;
      border: none;
      color: ${(p) => p.theme.brand};
      font-size: 12px;
      font-weight: 500;
    }
  }
  .state {
    font-size: 11px;
    color: #999;
    margin-bottom: 12px;
  }
  .payable {
    background: #fffcdc;
    margin: 0 -20px;
    padding: 14px 20px;
    display: flex;
    justify-content: space-between;
    font-weight: 500;
    font-size: 15px;
  }
  .ticket-type {
    font-size: 12px;
    color: #666;
    padding: 12px 0;
    label {
      margin-right: 16px;
      cursor: pointer;
    }
  }
  .cta {
    margin: 0 -20px;
    padding: 16px 20px 20px;
  }
`;

function PriceSummary({ txn, onToggleSmile, cta }) {
  const [open, setOpen] = useState(false);
  const p = txn.pricing;
  const s = txn.snapshot;
  return (
    <Summary>
      <h4>BOOKING SUMMARY</h4>
      {txn.kind === 'EVENT' ? (
        <>
          <div className="line">
            <span>
              {s.tierName} ({txn.qty} Tickets)
            </span>
            <span>{rs(p.ticketAmount)}</span>
          </div>
          <div className="mut">{s.venueName}</div>
        </>
      ) : (
        <>
          {s.lines.map((l) => (
            <div className="line" key={l.areaCode}>
              <span>
                {l.areaName} - {l.seats.join(', ')} ({l.seats.length} Tickets)
              </span>
              <span>{rs(l.price * l.seats.length)}</span>
            </div>
          ))}
          <div className="mut">{s.screenName}</div>
        </>
      )}
      <div className="line">
        <span className="fee" onClick={() => setOpen((o) => !o)}>
          Convenience fees {open ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        </span>
        <span>{rs(p.convFee)}</span>
      </div>
      {open && (
        <div className="breakup">
          <div className="line">
            <span>Base Amount</span>
            <span>{rs(p.convBase)}</span>
          </div>
          <div className="line">
            <span>Integrated GST (IGST) @ 18%</span>
            <span>{rs(p.gst)}</span>
          </div>
        </div>
      )}
      <hr />
      <div className="line">
        <span>Sub total</span>
        <span>{rs(p.subTotal)}</span>
      </div>
      {txn.fnb.length > 0 && (
        <>
          <div className="line" style={{ marginTop: 12 }}>
            <span>Food &amp; Beverage</span>
            <span>{rs(p.fnbAmount)}</span>
          </div>
          <div className="mut">{txn.fnb.map((f) => `${f.qty} x ${f.name}`).join(', ')}</div>
        </>
      )}
      {onToggleSmile ? (
        <div className="smile">
          <span>
            Contribution to BookASmile
            <br />
            <small style={{ color: '#999' }}>(₹1 per ticket will be added)</small>
          </span>
          <button onClick={onToggleSmile}>{txn.donation ? 'Remove' : 'Add Rs. ' + txn.qty}</button>
        </div>
      ) : (
        txn.donation && (
          <div className="line" style={{ marginTop: 12 }}>
            <span>BookASmile</span>
            <span>{rs(p.donation)}</span>
          </div>
        )
      )}
      {s.stateName && <div className="state">Your current State is {s.stateName}</div>}
      <div className="payable">
        <span>Amount Payable</span>
        <span>{rs(p.total)}</span>
      </div>
      {cta}
    </Summary>
  );
}

const FnbPanel = styled(Card)`
  padding: 20px;
  .head {
    display: flex;
    align-items: baseline;
    gap: 12px;
    margin-bottom: 16px;
    h2 {
      font-size: 20px;
    }
    span {
      color: #999;
      font-size: 13px;
    }
  }
  .tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 16px;
    flex-wrap: wrap;
    button {
      border: 1px solid ${(p) => p.theme.border};
      background: #fff;
      border-radius: 20px;
      padding: 6px 14px;
      font-size: 13px;
      &[data-on='true'] {
        background: ${(p) => p.theme.dark};
        color: #fff;
        border-color: ${(p) => p.theme.dark};
      }
    }
  }
  .items {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 16px;
  }
  .item {
    border: 1px solid ${(p) => p.theme.border};
    border-radius: 8px;
    padding: 12px;
    display: flex;
    gap: 12px;
    img {
      width: 72px;
      height: 72px;
      border-radius: 8px;
    }
    .n {
      font-size: 14px;
      font-weight: 500;
      display: flex;
      gap: 6px;
      align-items: center;
    }
    .d {
      font-size: 11px;
      color: #999;
      margin: 4px 0 8px;
    }
    .row {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }
  .veg {
    width: 12px;
    height: 12px;
    border: 1px solid ${(p) => p.theme.green};
    display: inline-grid;
    place-items: center;
    &::after {
      content: '';
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: ${(p) => p.theme.green};
    }
    &[data-veg='false'] {
      border-color: #b33;
      &::after {
        background: #b33;
      }
    }
  }
`;

const Stepper = styled.div`
  display: inline-flex;
  align-items: center;
  border: 1px solid ${(p) => p.theme.brand};
  border-radius: 4px;
  color: ${(p) => p.theme.brand};
  button {
    background: none;
    border: none;
    color: inherit;
    width: 28px;
    height: 28px;
    font-size: 16px;
  }
  span {
    min-width: 20px;
    text-align: center;
    font-size: 13px;
  }
`;

export function OrderSummaryPage() {
  const { txn: id } = useParams();
  const city = useRegionSlug();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const txn = useTxn(id);
  const { data: me } = useGetMeQuery();
  const { data: menu } = useGetFnbQuery(undefined, { skip: !txn.data?.snapshot?.fnb });
  const [updateTxn] = useUpdateTxnMutation();
  const [cancelTxn] = useCancelTxnMutation();
  const [tab, setTab] = useState('All');
  const [confirmBack, setConfirmBack] = useState(false);

  if (txn.isLoading) return <Spinner />;
  if (txn.expired || txn.error?.status === 410) {
    return (
      <>
        <BookingTop txn={txn.data} />
        <ExpiredModal open onOk={() => navigate(retryPath(txn.data, city))} />
      </>
    );
  }
  if (txn.isError || !txn.data) return <ErrorState />;
  const t = txn.data;

  const qtyOf = (code) => t.fnb.find((f) => f.code === code)?.qty || 0;
  const setFnb = async (code, qty) => {
    const items = [...t.fnb.filter((f) => f.code !== code).map((f) => ({ code: f.code, qty: f.qty })), ...(qty > 0 ? [{ code, qty }] : [])];
    try {
      await updateTxn({ id, fnb: items }).unwrap();
    } catch (err) {
      dispatch(pushToast(errorMessage(err), 'error'));
    }
  };
  const proceed = () => {
    track('booking_summary_proceed', { txn_id: id, amount: t.pricing.total });
    const next = `/booking/payment/${id}`;
    if (!me?.user) dispatch(openLogin(next));
    else navigate(next);
  };
  const leave = async () => {
    await cancelTxn(id);
    navigate(retryPath(t, city));
  };

  const items = (menu?.items || []).filter((i) => tab === 'All' || i.category === tab);

  return (
    <>
      <Title>Booking Summary</Title>
      <BookingTop txn={t} onBack={() => setConfirmBack(true)} />
      <Page>
        <Cols>
          <div>
            {t.kind === 'MOVIE' && t.snapshot.fnb ? (
              <FnbPanel>
                <div className="head">
                  <h2>Grab a bite!</h2>
                  <span>Prebook your meal and save more!</span>
                </div>
                <div className="tabs">
                  {menu?.categories.map((c) => (
                    <button key={c} data-on={c === tab} onClick={() => setTab(c)}>
                      {c}
                    </button>
                  ))}
                </div>
                <div className="items">
                  {items.map((i) => (
                    <div className="item" key={i.code}>
                      <img src={i.image} alt="" />
                      <div style={{ flex: 1 }}>
                        <div className="n">
                          <span className="veg" data-veg={String(i.veg)} aria-label={i.veg ? 'Veg' : 'Non-veg'} /> {i.name}
                        </div>
                        <div className="d">{i.desc}</div>
                        <div className="row">
                          <b>{rupee(i.price)}</b>
                          {qtyOf(i.code) ? (
                            <Stepper>
                              <button onClick={() => setFnb(i.code, qtyOf(i.code) - 1)} aria-label="Remove one">
                                −
                              </button>
                              <span>{qtyOf(i.code)}</span>
                              <button onClick={() => setFnb(i.code, qtyOf(i.code) + 1)} aria-label="Add one">
                                +
                              </button>
                            </Stepper>
                          ) : (
                            <Button $variant="outline" $size="sm" onClick={() => setFnb(i.code, 1)}>
                              Add
                            </Button>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </FnbPanel>
            ) : (
              <Card style={{ padding: 24 }}>
                <h2 style={{ fontSize: 20, marginBottom: 12 }}>{t.snapshot.title}</h2>
                <p style={{ color: '#666', lineHeight: 1.6 }}>
                  {t.snapshot.venueName}
                  <br />
                  {showDateLong(t.snapshot.startsAt)} | {showTime(t.snapshot.startsAt)}
                  <br />
                  {t.qty} x {t.snapshot.tierName || seatSummary(t.snapshot.lines)}
                </p>
              </Card>
            )}
          </div>
          <PriceSummary
            txn={t}
            onToggleSmile={() => updateTxn({ id, donation: !t.donation })}
            cta={
              <>
                {t.kind === 'MOVIE' && (
                  <div className="ticket-type">
                    <label>
                      <input type="radio" defaultChecked name="tt" /> M-Ticket
                    </label>
                    <label>
                      <input type="radio" name="tt" /> Box Office Pickup
                    </label>
                  </div>
                )}
                <div className="cta">
                  <Button $block $size="lg" onClick={proceed}>
                    Proceed
                  </Button>
                </div>
              </>
            }
          />
        </Cols>
      </Page>
      <Modal open={confirmBack} onClose={() => setConfirmBack(false)} width={400} label="Cancel transaction">
        <div style={{ padding: 28, textAlign: 'center' }}>
          <h3 style={{ marginBottom: 12 }}>Are you sure?</h3>
          <p style={{ color: '#666', marginBottom: 24 }}>Your seats will be released and you will be taken back to seat selection.</p>
          <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
            <Button $variant="outline" onClick={() => setConfirmBack(false)}>
              No
            </Button>
            <Button onClick={leave}>Yes, cancel</Button>
          </div>
        </div>
      </Modal>
    </>
  );
}

// ---- payment ---------------------------------------------------------------

const PayCard = styled(Card)`
  margin-bottom: 16px;
  h3 {
    font-size: 16px;
    padding: 16px 20px;
    border-bottom: 1px solid ${(p) => p.theme.border};
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .contact {
    padding: 20px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    @media (max-width: 600px) {
      grid-template-columns: 1fr;
    }
  }
  label {
    font-size: 12px;
    color: #666;
    display: block;
  }
  input,
  select {
    width: 100%;
    border: 1px solid #ccc;
    border-radius: 4px;
    padding: 10px 12px;
    font-size: 14px;
    margin-top: 6px;
    &:focus {
      outline: none;
      border-color: ${(p) => p.theme.brand};
    }
  }
  .methods {
    display: grid;
    grid-template-columns: 220px 1fr;
    min-height: 320px;
    @media (max-width: 700px) {
      grid-template-columns: 1fr;
    }
  }
  .tabs {
    background: ${(p) => p.theme.bg};
    button {
      display: block;
      width: 100%;
      text-align: left;
      border: none;
      background: none;
      padding: 16px 20px;
      font-size: 14px;
      color: #333;
      border-left: 3px solid transparent;
      &[data-on='true'] {
        background: #fff;
        border-left-color: ${(p) => p.theme.brand};
        color: ${(p) => p.theme.brand};
        font-weight: 500;
      }
      &:disabled {
        color: #bbb;
      }
    }
  }
  .panel {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 440px;
  }
  .pair {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }
  .banks,
  .wallets {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
    button {
      border: 1px solid ${(p) => p.theme.border};
      background: #fff;
      border-radius: 6px;
      padding: 12px;
      font-size: 13px;
      &[data-on='true'] {
        border-color: ${(p) => p.theme.brand};
        background: ${(p) => p.theme.brandSoft};
      }
    }
  }
  .hint {
    font-size: 12px;
    color: #999;
    line-height: 1.5;
  }
  .err {
    color: #d32f2f;
    font-size: 13px;
    background: #fdecea;
    padding: 10px 12px;
    border-radius: 4px;
  }
`;

const OrderCard = styled(Summary)`
  .title {
    font-size: 16px;
    font-weight: 500;
    margin-bottom: 4px;
  }
  .qtybox {
    float: right;
    text-align: center;
    border: 1px solid ${(p) => p.theme.border};
    border-radius: 4px;
    padding: 4px 10px;
    font-size: 18px;
    font-weight: 700;
    small {
      display: block;
      font-size: 10px;
      font-weight: 400;
      color: #666;
    }
  }
`;

const METHODS = [
  { code: 'UPI', label: 'Pay by any UPI App' },
  { code: 'CARD', label: 'Debit/Credit Card' },
  { code: 'NETBANKING', label: 'Net Banking' },
  { code: 'WALLET', label: 'Mobile Wallets' },
  { code: 'GIFT', label: 'Gift Voucher', disabled: true },
];
const BANKS = ['National Bank', 'Metro Bank', 'Coastal Bank', 'Union Federal', 'Heritage Bank', 'Prime Co-op'];
const WALLETS = ['SB Pay', 'QuickWallet', 'PayFast', 'Freecash'];

export function PaymentPage() {
  const { txn: id } = useParams();
  const city = useRegionSlug();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const txn = useTxn(id);
  const { data: me, isLoading: meLoading } = useGetMeQuery();
  const [pay, { isLoading: paying }] = usePayMutation();
  const [method, setMethod] = useState('UPI');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [card, setCard] = useState({ number: '', name: '', expiry: '', cvv: '' });
  const [upi, setUpi] = useState('');
  const [bank, setBank] = useState(BANKS[0]);
  const [error, setError] = useState('');

  useEffect(() => {
    if (me?.user) {
      setEmail((e) => e || me.user.email);
      setPhone((p) => p || me.user.phone);
    }
  }, [me]);
  useEffect(() => {
    if (!meLoading && !me?.user) dispatch(openLogin(`/booking/payment/${id}`));
  }, [me, meLoading, dispatch, id]);

  if (txn.isLoading) return <Spinner />;
  if (txn.expired) {
    return (
      <>
        <BookingTop txn={txn.data} />
        <ExpiredModal open onOk={() => navigate(retryPath(txn.data, city))} />
      </>
    );
  }
  if (txn.isError || !txn.data) return <ErrorState />;
  const t = txn.data;
  const s = t.snapshot;

  const submit = async (e) => {
    e.preventDefault();
    setError('');
    if (!me?.user) {
      dispatch(openLogin(`/booking/payment/${id}`));
      return;
    }
    track('payment_initiated', { txn_id: id, method, amount: t.pricing.total });
    try {
      const res = await pay({ id, method, email, phone, card, upi, bank }).unwrap();
      track('booking_confirmed', { booking_id: res.bookingId, amount: t.pricing.total });
      navigate(`/booking/confirmation/${res.bookingId}`, { replace: true });
    } catch (err) {
      if (errorCode(err) === 'TXN_EXPIRED') txn.refetch();
      setError(errorMessage(err));
    }
  };

  const fmtCard = (v) => v.replace(/\D/g, '').slice(0, 16).replace(/(.{4})/g, '$1 ').trim();
  const fmtExp = (v) => v.replace(/\D/g, '').slice(0, 4).replace(/^(\d{2})(\d)/, '$1/$2');

  return (
    <>
      <Title>Payment</Title>
      <BookingTop txn={t} onBack={() => navigate(`/booking/summary/${id}`)} />
      <Page>
        <Cols as="form" onSubmit={submit}>
          <div>
            <PayCard>
              <h3>Share your Contact Details</h3>
              <div className="contact">
                <label>
                  Email address
                  <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="name@example.com" />
                </label>
                <label>
                  Mobile number
                  <input inputMode="numeric" maxLength={10} value={phone} onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))} placeholder="10-digit mobile" />
                </label>
              </div>
            </PayCard>
            <PayCard>
              <h3>
                <CardIcon size={18} /> Payment options
              </h3>
              <div className="methods">
                <div className="tabs">
                  {METHODS.map((m) => (
                    <button type="button" key={m.code} data-on={m.code === method} disabled={m.disabled} onClick={() => setMethod(m.code)}>
                      {m.label}
                    </button>
                  ))}
                </div>
                <div className="panel">
                  {method === 'UPI' && (
                    <>
                      <label>
                        Enter UPI ID
                        <input value={upi} onChange={(e) => setUpi(e.target.value.trim())} placeholder="yourname@okbank" />
                      </label>
                      <p className="hint">Demo: any valid UPI ID succeeds; use fail@upi to simulate a decline.</p>
                    </>
                  )}
                  {method === 'CARD' && (
                    <>
                      <label>
                        Card Number
                        <input inputMode="numeric" value={card.number} onChange={(e) => setCard({ ...card, number: fmtCard(e.target.value) })} placeholder="XXXX XXXX XXXX XXXX" />
                      </label>
                      <label>
                        Name on the card
                        <input value={card.name} onChange={(e) => setCard({ ...card, name: e.target.value })} />
                      </label>
                      <div className="pair">
                        <label>
                          Expiry (MM/YY)
                          <input inputMode="numeric" value={card.expiry} onChange={(e) => setCard({ ...card, expiry: fmtExp(e.target.value) })} placeholder="MM/YY" />
                        </label>
                        <label>
                          CVV
                          <input inputMode="numeric" type="password" maxLength={4} value={card.cvv} onChange={(e) => setCard({ ...card, cvv: e.target.value.replace(/\D/g, '') })} />
                        </label>
                      </div>
                      <p className="hint">Demo: 4111 1111 1111 1111 succeeds; 4000 0000 0000 0002 is declined.</p>
                    </>
                  )}
                  {method === 'NETBANKING' && (
                    <div className="banks">
                      {BANKS.map((b) => (
                        <button type="button" key={b} data-on={b === bank} onClick={() => setBank(b)}>
                          {b}
                        </button>
                      ))}
                    </div>
                  )}
                  {method === 'WALLET' && (
                    <div className="wallets">
                      {WALLETS.map((b) => (
                        <button type="button" key={b} data-on={b === bank} onClick={() => setBank(b)}>
                          {b}
                        </button>
                      ))}
                    </div>
                  )}
                  {error && <div className="err">{error}</div>}
                  <Button $size="lg" disabled={paying}>
                    {paying ? 'Processing payment…' : `Make Payment ${rs(t.pricing.total)}`}
                  </Button>
                  <p className="hint" style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
                    <ShieldIcon size={14} /> This is a simulated gateway. No money is charged.
                  </p>
                </div>
              </div>
            </PayCard>
          </div>
          <OrderCard>
            <h4>ORDER SUMMARY</h4>
            <div className="qtybox">
              {t.qty}
              <small>Tickets</small>
            </div>
            <div className="title">{s.title}</div>
            <div className="mut">
              {[s.certificate, s.language, s.format].filter(Boolean).join(', ')}
              <br />
              {s.venueName}
              <br />
              {t.kind === 'MOVIE' ? 'M-Ticket' : 'E-Ticket'}
            </div>
            <div className="line" style={{ marginBottom: 12 }}>
              <b>{showDateLong(s.startsAt)}</b>
              <b>{showTime(s.startsAt)}</b>
            </div>
            <div className="mut">{t.kind === 'MOVIE' ? seatSummary(s.lines) : s.tierName}</div>
            <hr />
            <div className="line">
              <span>Sub total</span>
              <span>{rs(t.pricing.subTotal)}</span>
            </div>
            {t.fnb.length > 0 && (
              <div className="line">
                <span>Food &amp; Beverage</span>
                <span>{rs(t.pricing.fnbAmount)}</span>
              </div>
            )}
            {t.donation && (
              <div className="line">
                <span>BookASmile</span>
                <span>{rs(t.pricing.donation)}</span>
              </div>
            )}
            <div className="payable" style={{ marginTop: 12 }}>
              <span>Amount Payable</span>
              <span>{rs(t.pricing.total)}</span>
            </div>
          </OrderCard>
        </Cols>
      </Page>
    </>
  );
}

// ---- confirmation ----------------------------------------------------------

const Ticket = styled.div`
  max-width: 720px;
  margin: 0 auto;
  .ok {
    text-align: center;
    margin-bottom: 24px;
    .tick {
      width: 64px;
      height: 64px;
      border-radius: 50%;
      background: ${(p) => p.theme.green};
      color: #fff;
      display: grid;
      place-items: center;
      margin: 0 auto 12px;
    }
    h2 {
      font-size: 22px;
    }
    p {
      color: #666;
      margin-top: 6px;
    }
  }
  .card {
    background: #fff;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  }
  .main {
    display: flex;
    gap: 20px;
    padding: 24px;
    img.poster {
      width: 110px;
      border-radius: 8px;
      aspect-ratio: 400 / 660;
      object-fit: cover;
    }
    h3 {
      font-size: 20px;
      margin-bottom: 4px;
    }
    .m {
      color: #666;
      font-size: 13px;
      line-height: 1.7;
    }
  }
  .perf {
    position: relative;
    border-top: 2px dashed #ddd;
    margin: 0 16px;
    &::before,
    &::after {
      content: '';
      position: absolute;
      top: -14px;
      width: 26px;
      height: 26px;
      border-radius: 50%;
      background: ${(p) => p.theme.bg};
    }
    &::before {
      left: -30px;
    }
    &::after {
      right: -30px;
    }
  }
  .bottom {
    display: flex;
    gap: 24px;
    padding: 24px;
    align-items: center;
    img.qr {
      width: 132px;
      height: 132px;
    }
    .k {
      color: #999;
      font-size: 12px;
    }
    .v {
      font-size: 15px;
      margin-bottom: 10px;
    }
    .id {
      font-size: 22px;
      font-weight: 700;
      letter-spacing: 3px;
    }
  }
  .status {
    display: inline-block;
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 10px;
    background: ${(p) => p.theme.greenSoft};
    color: ${(p) => p.theme.green};
    &[data-s='CANCELLED'] {
      background: #fdecea;
      color: #d32f2f;
    }
  }
  .actions {
    display: flex;
    gap: 12px;
    justify-content: center;
    margin-top: 24px;
    flex-wrap: wrap;
  }
`;

function QR({ text }) {
  const [src, setSrc] = useState('');
  useEffect(() => {
    let live = true;
    import('qrcode').then((m) => m.toDataURL(text, { margin: 1, width: 264 })).then((url) => live && setSrc(url));
    return () => {
      live = false;
    };
  }, [text]);
  return src ? <img className="qr" src={src} alt={`QR code for booking ${text}`} /> : <div style={{ width: 132, height: 132, background: '#f3f3f3' }} />;
}

export function ConfirmationPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { data: me, isLoading: meLoading } = useGetMeQuery();
  const { data: b, isLoading, isError, error } = useGetBookingQuery(id, { skip: !me?.user });
  const [cancelBooking, { isLoading: cancelling }] = useCancelBookingMutation();
  const [ask, setAsk] = useState(false);

  if (meLoading || isLoading) return <Spinner />;
  if (!me?.user) return <EmptyState art="ticket" title="Sign in to view your ticket" cta="Sign in" onCta={() => dispatch(openLogin(`/booking/confirmation/${id}`))} />;
  if (isError) return error?.status === 404 ? <EmptyState art="ticket" title="Booking not found" cta="Go Home" onCta={() => navigate('/')} /> : <ErrorState />;
  const s = b.snapshot;

  const doCancel = async () => {
    try {
      const res = await cancelBooking(b.id).unwrap();
      setAsk(false);
      dispatch(pushToast(`Booking cancelled. ${rs(res.refund)} will be refunded in 5-7 working days.`, 'success'));
    } catch (err) {
      dispatch(pushToast(errorMessage(err), 'error'));
    }
  };

  return (
    <Page>
      <Title>{`Booking ${b.id}`}</Title>
      <Container>
        <Ticket>
          <div className="ok">
            {b.status === 'CONFIRMED' ? (
              <>
                <div className="tick">
                  <CheckIcon size={34} />
                </div>
                <h2>Your booking is confirmed!</h2>
                <p>
                  We've sent your {b.kind === 'MOVIE' ? 'M-Ticket' : 'E-Ticket'} to <b>{b.contactEmail}</b>
                </p>
              </>
            ) : (
              <h2>This booking has been cancelled</h2>
            )}
          </div>
          <div className="card">
            <div className="main">
              <img className="poster" src={s.poster} alt="" />
              <div>
                <h3>{s.title}</h3>
                <div className="m">
                  {[s.certificate, s.language, s.format].filter(Boolean).join(' | ')}
                  <br />
                  <b style={{ color: '#333' }}>
                    {showDateLong(s.startsAt)} | {showTime(s.startsAt)}
                  </b>
                  <br />
                  {s.venueName}
                  {s.screenName && ` · ${s.screenName}`}
                  <br />
                  {b.kind === 'MOVIE' ? seatSummary(s.lines) : `${b.qty} x ${s.tierName}`}
                  {b.fnb.length > 0 && (
                    <>
                      <br />
                      🍿 {b.fnb.map((f) => `${f.qty} x ${f.name}`).join(', ')}
                    </>
                  )}
                </div>
              </div>
            </div>
            <div className="perf" />
            <div className="bottom">
              <QR text={b.id} />
              <div>
                <div className="k">BOOKING ID</div>
                <div className="id">{b.id}</div>
                <div className="k" style={{ marginTop: 10 }}>
                  STATUS
                </div>
                <div className="v">
                  <span className="status" data-s={b.status}>
                    {b.status}
                  </span>
                </div>
                <div className="k">AMOUNT PAID</div>
                <div className="v">
                  {rs(b.pricing.total)} via {b.paymentMethod}
                </div>
              </div>
            </div>
          </div>
          <div className="actions">
            <Button $variant="outline" onClick={() => navigate('/my-profile/purchase-history')}>
              View all bookings
            </Button>
            {b.cancellable && (
              <Button $variant="outline" onClick={() => setAsk(true)}>
                Cancel booking
              </Button>
            )}
            <Button onClick={() => navigate('/')}>Back to Home</Button>
          </div>
        </Ticket>
      </Container>
      <Modal open={ask} onClose={() => setAsk(false)} width={420} label="Cancel booking">
        <div style={{ padding: 28, textAlign: 'center' }}>
          <h3 style={{ marginBottom: 12 }}>Cancel this booking?</h3>
          <p style={{ color: '#666', marginBottom: 24, lineHeight: 1.5 }}>
            You will receive a refund of {rs(b.pricing.ticketAmount + b.pricing.fnbAmount)}. Convenience fees are non-refundable.
          </p>
          <Button onClick={doCancel} disabled={cancelling}>
            {cancelling ? 'Cancelling…' : 'Yes, cancel booking'}
          </Button>
        </div>
      </Modal>
    </Page>
  );
}

// ---- purchase history & profile -------------------------------------------

const Orders = styled.div`
  max-width: 860px;
  margin: 0 auto;
  h1 {
    font-size: 24px;
    margin-bottom: 20px;
  }
  .o {
    background: #fff;
    border-radius: 8px;
    display: flex;
    gap: 16px;
    padding: 16px;
    margin-bottom: 12px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
    img {
      width: 72px;
      border-radius: 6px;
      aspect-ratio: 400 / 660;
      object-fit: cover;
    }
    .t {
      font-size: 16px;
      font-weight: 500;
    }
    .m {
      color: #666;
      font-size: 13px;
      line-height: 1.6;
      margin-top: 2px;
    }
    .r {
      margin-left: auto;
      text-align: right;
      display: flex;
      flex-direction: column;
      gap: 8px;
      align-items: flex-end;
    }
  }
`;

export function OrdersPage() {
  const dispatch = useDispatch();
  const { data: me, isLoading: meLoading } = useGetMeQuery();
  const { data, isLoading, isError } = useGetBookingsQuery(undefined, { skip: !me?.user });
  if (meLoading || isLoading) return <Spinner />;
  if (!me?.user) return <EmptyState art="ticket" title="Sign in to see your orders" cta="Sign in" onCta={() => dispatch(openLogin('/my-profile/purchase-history'))} />;
  if (isError) return <ErrorState />;
  return (
    <Page>
      <Title>Purchase History</Title>
      <Container>
        <Orders>
          <h1>Your Orders</h1>
          {data.bookings.length === 0 ? (
            <EmptyState art="ticket" title="No bookings yet" message="Your movie and event bookings will show up here." />
          ) : (
            data.bookings.map((b) => (
              <div className="o" key={b.id}>
                <img src={b.snapshot.poster} alt="" />
                <div>
                  <div className="t">{b.snapshot.title}</div>
                  <div className="m">
                    {showDateLong(b.snapshot.startsAt)} | {showTime(b.snapshot.startsAt)}
                    <br />
                    {b.snapshot.venueName}
                    <br />
                    {b.kind === 'MOVIE' ? seatSummary(b.snapshot.lines) : `${b.qty} x ${b.snapshot.tierName}`}
                    <br />
                    Booking ID: <b>{b.id}</b>
                  </div>
                </div>
                <div className="r">
                  <span style={{ fontSize: 11, fontWeight: 500, color: b.status === 'CONFIRMED' ? '#1EA83C' : '#d32f2f' }}>{b.status}</span>
                  <b>{rs(b.pricing.total)}</b>
                  <Link to={`/booking/confirmation/${b.id}`} style={{ color: '#F84464', fontSize: 13 }}>
                    View ticket ›
                  </Link>
                </div>
              </div>
            ))
          )}
        </Orders>
      </Container>
    </Page>
  );
}

const Form = styled(Card)`
  max-width: 560px;
  margin: 0 auto;
  padding: 28px;
  h1 {
    font-size: 22px;
    margin-bottom: 20px;
  }
  label {
    display: block;
    font-size: 12px;
    color: #666;
    margin-bottom: 16px;
  }
  input {
    width: 100%;
    border: 1px solid #ccc;
    border-radius: 4px;
    padding: 10px 12px;
    font-size: 14px;
    margin-top: 6px;
    &:disabled {
      background: #f5f5f5;
    }
  }
`;

export function ProfilePage() {
  const dispatch = useDispatch();
  const { data: me, isLoading } = useGetMeQuery();
  const [save, { isLoading: saving }] = useUpdateMeMutation();
  const [form, setForm] = useState({ name: '', email: '', phone: '' });
  useEffect(() => {
    if (me?.user) setForm({ name: me.user.name, email: me.user.email, phone: me.user.phone });
  }, [me]);
  if (isLoading) return <Spinner />;
  if (!me?.user) return <EmptyState title="Sign in to edit your profile" cta="Sign in" onCta={() => dispatch(openLogin('/my-profile/edit'))} />;
  const u = me.user;
  return (
    <Page>
      <Title>Edit Profile</Title>
      <Container>
        <Form
          as="form"
          onSubmit={async (e) => {
            e.preventDefault();
            try {
              await save(form).unwrap();
              dispatch(pushToast('Profile updated', 'success'));
            } catch (err) {
              dispatch(pushToast(errorMessage(err), 'error'));
            }
          }}
        >
          <h1>Profile Details</h1>
          <label>
            Name
            <input value={form.name} maxLength={60} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Your name" />
          </label>
          <label>
            Email address
            <input type="email" value={form.email} disabled={u.identifier === u.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
          </label>
          <label>
            Mobile number
            <input value={form.phone} disabled={u.identifier === u.phone} maxLength={10} onChange={(e) => setForm({ ...form, phone: e.target.value.replace(/\D/g, '') })} />
          </label>
          <Button disabled={saving}>{saving ? 'Saving…' : 'Save Changes'}</Button>
        </Form>
      </Container>
    </Page>
  );
}
