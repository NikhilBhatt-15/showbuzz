import { useEffect, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import styled from 'styled-components';
import { Button, Modal } from './ui.jsx';
import {
  BagIcon, BellIcon, CardIcon, ChevronRight, CloseIcon, GearIcon, GiftIcon, HelpIcon, LogoutIcon, PinIcon, SearchIcon, SmileIcon, TvIcon, UserIcon,
} from './icons.jsx';
import { closeCityModal, closeLogin, openLogin, pushToast, setRegion, toggleSidebar } from '../store/ui.js';
import { errorMessage, useGetMeQuery, useGetRegionsQuery, useLogoutMutation, useRequestOtpMutation, useVerifyOtpMutation } from '../store/api.js';
import { pathForCity, writeRegionCookie } from '../lib/region.js';
import { track } from '../lib/clickstream.js';

// ---- city picker -----------------------------------------------------------

const CityBody = styled.div`
  padding: 20px 24px 28px;
  .search {
    display: flex;
    align-items: center;
    gap: 10px;
    border: 1px solid ${(p) => p.theme.border};
    border-radius: 4px;
    padding: 0 12px;
    height: 40px;
    color: #999;
    margin-right: 28px;
    input {
      flex: 1;
      border: none;
      outline: none;
      font-size: 14px;
    }
  }
  .detect {
    background: none;
    border: none;
    color: ${(p) => p.theme.brand};
    display: flex;
    gap: 6px;
    align-items: center;
    margin: 14px 0 6px;
    font-size: 14px;
  }
  h4 {
    text-align: center;
    font-weight: 400;
    font-size: 14px;
    color: #333;
    margin: 20px 0 12px;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 8px;
    @media (max-width: 600px) {
      grid-template-columns: repeat(3, 1fr);
    }
  }
  .city {
    background: none;
    border: none;
    padding: 8px 0;
    color: #666;
    font-size: 13px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    border-radius: 8px;
    img {
      width: 64px;
      height: 64px;
    }
    &:hover,
    &[data-active='true'] {
      color: ${(p) => p.theme.brand};
      background: ${(p) => p.theme.brandSoft};
    }
  }
  .toggle {
    display: block;
    margin: 16px auto 0;
    background: none;
    border: none;
    color: ${(p) => p.theme.brand};
    font-size: 14px;
  }
  .others {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px 16px;
    margin-top: 12px;
    button {
      text-align: left;
      background: none;
      border: none;
      color: #666;
      font-size: 13px;
      padding: 2px 0;
      &:hover {
        color: ${(p) => p.theme.brand};
      }
    }
  }
`;

export function CityModal() {
  const open = useSelector((s) => s.ui.cityModalOpen);
  const current = useSelector((s) => s.ui.regionSlug);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { data } = useGetRegionsQuery();
  const [q, setQ] = useState('');
  const [showAll, setShowAll] = useState(false);
  const [detecting, setDetecting] = useState(false);

  const choose = (r) => {
    writeRegionCookie(r.slug);
    dispatch(setRegion(r.slug));
    track('region_selected', { region_code: r.code });
    navigate(pathForCity(location.pathname, r.slug) + (location.pathname.startsWith('/explore/') ? location.search.replace(/[?&]pickCity=1/, '') : ''));
  };

  const detect = () => {
    if (!navigator.geolocation || !data) return;
    setDetecting(true);
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        setDetecting(false);
        const all = [...data.popular, ...data.others];
        const near = all.reduce((best, r) => {
          const d = (r.lat - coords.latitude) ** 2 + (r.lng - coords.longitude) ** 2;
          return !best || d < best.d ? { r, d } : best;
        }, null);
        if (near) choose(near.r);
      },
      () => {
        setDetecting(false);
        dispatch(pushToast('Location permission denied. Please pick your city.', 'error'));
      },
    );
  };

  const all = data ? [...data.popular, ...data.others] : [];
  const filtered = q ? all.filter((r) => r.name.toLowerCase().includes(q.toLowerCase())) : null;

  return (
    <Modal open={open} onClose={() => dispatch(closeCityModal())} width={760} top label="Select your city">
      <CityBody>
        <label className="search">
          <SearchIcon size={18} />
          <input autoFocus placeholder="Search for your city" value={q} onChange={(e) => setQ(e.target.value)} />
        </label>
        <button className="detect" onClick={detect} disabled={detecting}>
          <PinIcon size={16} /> {detecting ? 'Detecting…' : 'Detect my location'}
        </button>
        {filtered ? (
          <div className="others">
            {filtered.map((r) => (
              <button key={r.code} onClick={() => choose(r)}>
                {r.name}
              </button>
            ))}
            {filtered.length === 0 && <span style={{ color: '#999' }}>No city found</span>}
          </div>
        ) : (
          <>
            <h4>Popular Cities</h4>
            <div className="grid">
              {data?.popular.map((r) => (
                <button key={r.code} className="city" data-active={r.slug === current} onClick={() => choose(r)}>
                  <img src={r.icon} alt="" />
                  {r.name}
                </button>
              ))}
            </div>
            {showAll && (
              <>
                <h4>Other Cities</h4>
                <div className="others">
                  {data?.others.map((r) => (
                    <button key={r.code} onClick={() => choose(r)}>
                      {r.name}
                    </button>
                  ))}
                </div>
              </>
            )}
            <button className="toggle" onClick={() => setShowAll((v) => !v)}>
              {showAll ? 'Hide All Cities' : 'View All Cities'}
            </button>
          </>
        )}
      </CityBody>
    </Modal>
  );
}

// ---- login -----------------------------------------------------------------

const LoginBody = styled.div`
  padding: 28px 32px 24px;
  h3 {
    text-align: center;
    font-size: 18px;
    font-weight: 500;
    margin-bottom: 24px;
  }
  .provider {
    width: 100%;
    border: 1px solid ${(p) => p.theme.border};
    background: #fff;
    border-radius: 8px;
    padding: 12px;
    font-size: 14px;
    color: #333;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    &:disabled {
      opacity: 0.5;
    }
  }
  .or {
    text-align: center;
    color: #999;
    font-size: 12px;
    margin: 16px 0;
  }
  .field {
    display: flex;
    align-items: center;
    border-bottom: 1px solid #ccc;
    padding: 8px 0;
    gap: 8px;
    font-size: 16px;
    &:focus-within {
      border-color: ${(p) => p.theme.brand};
    }
    input {
      flex: 1;
      border: none;
      outline: none;
      font-size: 16px;
    }
  }
  .terms {
    font-size: 11px;
    color: #999;
    text-align: center;
    margin-top: 32px;
    line-height: 1.5;
  }
  .otp {
    display: flex;
    gap: 10px;
    justify-content: center;
    margin: 24px 0 12px;
    input {
      width: 44px;
      height: 52px;
      border: 1px solid #ccc;
      border-radius: 8px;
      text-align: center;
      font-size: 22px;
      &:focus {
        outline: none;
        border-color: ${(p) => p.theme.brand};
      }
    }
  }
  .sub {
    text-align: center;
    color: #666;
    font-size: 14px;
    line-height: 1.5;
  }
  .err {
    color: #d32f2f;
    font-size: 13px;
    text-align: center;
    min-height: 18px;
    margin: 8px 0;
  }
  .dev {
    background: ${(p) => p.theme.amberSoft};
    border: 1px dashed ${(p) => p.theme.amber};
    color: #7a5200;
    font-size: 12px;
    padding: 8px 12px;
    border-radius: 6px;
    text-align: center;
    margin-bottom: 8px;
  }
  .resend {
    text-align: center;
    font-size: 13px;
    color: #999;
    margin-top: 12px;
    button {
      background: none;
      border: none;
      color: ${(p) => p.theme.brand};
      font-size: 13px;
    }
  }
`;

export function LoginModal() {
  const open = useSelector((s) => s.ui.loginOpen);
  const next = useSelector((s) => s.ui.loginNext);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [step, setStep] = useState('start'); // start | email | otp
  const [identifier, setIdentifier] = useState('');
  const [otp, setOtp] = useState(['', '', '', '', '', '']);
  const [sent, setSent] = useState(null);
  const [error, setError] = useState('');
  const [cooldown, setCooldown] = useState(0);
  const [requestOtp, { isLoading: sending }] = useRequestOtpMutation();
  const [verifyOtp, { isLoading: verifying }] = useVerifyOtpMutation();
  const boxes = useRef([]);

  useEffect(() => {
    if (!open) {
      setStep('start');
      setIdentifier('');
      setOtp(['', '', '', '', '', '']);
      setError('');
      setSent(null);
    }
  }, [open]);

  useEffect(() => {
    if (cooldown <= 0) return undefined;
    const id = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [cooldown]);

  const send = async (e) => {
    e?.preventDefault();
    setError('');
    try {
      const res = await requestOtp({ identifier }).unwrap();
      setSent(res);
      setStep('otp');
      setCooldown(res.resendIn || 30);
      setTimeout(() => boxes.current[0]?.focus(), 50);
    } catch (err) {
      setError(errorMessage(err));
    }
  };

  const verify = async (code) => {
    setError('');
    try {
      const res = await verifyOtp({ identifier, otp: code }).unwrap();
      track('login_success', { channel: sent?.channel });
      dispatch(pushToast(`Welcome${res.user.name ? ', ' + res.user.name : ''}! You're signed in.`, 'success'));
      const dest = next;
      dispatch(closeLogin());
      if (dest) navigate(dest);
    } catch (err) {
      setError(errorMessage(err));
      setOtp(['', '', '', '', '', '']);
      boxes.current[0]?.focus();
    }
  };

  const setDigit = (i, v) => {
    const digits = v.replace(/\D/g, '');
    const copy = [...otp];
    if (digits.length > 1) {
      digits.slice(0, 6).split('').forEach((d, k) => (copy[k] = d));
    } else {
      copy[i] = digits;
    }
    setOtp(copy);
    const filled = copy.join('');
    if (filled.length === 6) verify(filled);
    else if (digits && i < 5) boxes.current[Math.min(5, i + digits.length)]?.focus();
  };

  const isPhone = /^\d+$/.test(identifier);

  return (
    <Modal open={open} onClose={() => dispatch(closeLogin())} width={460} label="Sign in">
      <LoginBody>
        {step === 'start' && (
          <form onSubmit={send}>
            <h3>Get Started</h3>
            <button type="button" className="provider" disabled title="Not available in this demo">
              <span style={{ fontWeight: 700, color: '#4285F4' }}>G</span> Continue with Google
            </button>
            <button type="button" className="provider" onClick={() => setStep('email')}>
              ✉️ Continue with Email
            </button>
            <button type="button" className="provider" disabled title="Not available in this demo">
               Continue with Apple
            </button>
            <div className="or">OR</div>
            <label className="field">
              <span>🇮🇳 +91</span>
              <input
                inputMode="numeric"
                maxLength={10}
                placeholder="Continue with mobile number"
                value={isPhone ? identifier : ''}
                onChange={(e) => setIdentifier(e.target.value.replace(/\D/g, ''))}
              />
            </label>
            <div className="err">{error}</div>
            {isPhone && identifier.length === 10 && (
              <Button $block disabled={sending}>
                {sending ? 'Sending OTP…' : 'Continue'}
              </Button>
            )}
            <p className="terms">I agree to the Terms &amp; Conditions and Privacy Policy</p>
          </form>
        )}
        {step === 'email' && (
          <form onSubmit={send}>
            <h3>Login with Email</h3>
            <label className="field">
              <input autoFocus type="email" placeholder="Enter your email address" value={identifier} onChange={(e) => setIdentifier(e.target.value.trim())} />
            </label>
            <div className="err">{error}</div>
            <Button $block disabled={sending || !identifier.includes('@')}>
              {sending ? 'Sending OTP…' : 'Continue'}
            </Button>
            <div className="resend">
              <button type="button" onClick={() => setStep('start')}>
                Use mobile number instead
              </button>
            </div>
          </form>
        )}
        {step === 'otp' && (
          <div>
            <h3>Verify your {sent?.channel === 'sms' ? 'Mobile Number' : 'Email Address'}</h3>
            <p className="sub">
              Enter OTP sent to <b>{sent?.channel === 'sms' ? '+91 ' : ''}{sent?.identifier}</b>
            </p>
            <div className="otp" onPaste={(e) => setDigit(0, e.clipboardData.getData('text'))}>
              {otp.map((d, i) => (
                <input
                  key={i}
                  ref={(el) => (boxes.current[i] = el)}
                  inputMode="numeric"
                  maxLength={6}
                  value={d}
                  aria-label={`OTP digit ${i + 1}`}
                  onChange={(e) => setDigit(i, e.target.value)}
                  onKeyDown={(e) => e.key === 'Backspace' && !otp[i] && i > 0 && boxes.current[i - 1]?.focus()}
                />
              ))}
            </div>
            <div className="err">{verifying ? 'Verifying…' : error}</div>
            {sent?.devOtp && (
              <div className="dev">
                Demo mode — your OTP is <b>{sent.devOtp}</b>
                {sent.channel === 'email' && <> (also delivered to the Mailpit inbox on port 8025)</>}
              </div>
            )}
            <div className="resend">
              {cooldown > 0 ? (
                <>Resend OTP in 00:{String(cooldown).padStart(2, '0')}</>
              ) : (
                <>
                  Didn't receive OTP? <button onClick={send}>Resend OTP</button>
                </>
              )}
            </div>
          </div>
        )}
      </LoginBody>
    </Modal>
  );
}

// ---- sidebar ---------------------------------------------------------------

const Drawer = styled.aside`
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 380px;
  max-width: 100vw;
  background: #fff;
  z-index: 1001;
  transform: translateX(${(p) => (p.$open ? '0' : '100%')});
  transition: transform 0.25s ease;
  display: flex;
  flex-direction: column;
  box-shadow: -4px 0 20px rgba(0, 0, 0, 0.15);
  .head {
    padding: 20px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    h3 {
      font-size: 22px;
    }
    a {
      font-size: 12px;
      color: #333;
      text-decoration: underline;
    }
    button {
      background: none;
      border: none;
    }
  }
  .promo {
    margin: 0 16px 8px;
    border: 1px solid ${(p) => p.theme.border};
    border-radius: 8px;
    padding: 16px;
    display: flex;
    gap: 12px;
    align-items: center;
    font-size: 14px;
  }
  nav {
    overflow-y: auto;
    flex: 1;
  }
  .item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 14px 24px;
    border-bottom: 1px solid #f0f0f0;
    color: #333;
    width: 100%;
    background: none;
    border-left: none;
    border-right: none;
    border-top: none;
    text-align: left;
    font-size: 14px;
    &:hover {
      background: ${(p) => p.theme.bg};
    }
    .label {
      flex: 1;
    }
    small {
      display: block;
      color: #999;
      font-size: 11px;
      margin-top: 2px;
    }
  }
`;

const Shade = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
`;

export function Sidebar() {
  const open = useSelector((s) => s.ui.sidebarOpen);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { data } = useGetMeQuery();
  const [logout] = useLogoutMutation();
  const user = data?.user;
  const close = () => dispatch(toggleSidebar(false));
  const go = (path, needsLogin) => {
    close();
    if (needsLogin && !user) dispatch(openLogin(path));
    else navigate(path);
  };
  const items = [
    { icon: BellIcon, label: 'Notifications', path: '/my-profile/purchase-history', login: true },
    { icon: BagIcon, label: 'Your Orders', sub: 'View all your bookings & purchases', path: '/my-profile/purchase-history', login: true },
    { icon: TvIcon, label: 'Stream Library', sub: 'Rented & Purchased Movies', path: '/offers' },
    { icon: CardIcon, label: 'Play Credit Card', sub: 'View your Play Credit Card details and offers', path: '/offers' },
    { icon: HelpIcon, label: 'Help & Support', sub: 'View commonly asked queries and Chat', path: '/offers' },
    { icon: GearIcon, label: 'Accounts & Settings', sub: 'Location, Payments, Permissions & More', path: '/my-profile/edit', login: true },
    { icon: GiftIcon, label: 'Rewards', sub: 'View your rewards & unlock new ones', path: '/offers' },
    { icon: SmileIcon, label: 'BookASmile', path: '/offers' },
  ];
  return (
    <>
      {open && <Shade onClick={close} />}
      <Drawer $open={open} aria-hidden={!open}>
        <div className="head">
          <div>
            <h3>Hey{user?.name ? `, ${user.name}` : '!'}</h3>
            {user && (
              <Link to="/my-profile/edit" onClick={close}>
                Edit Profile ›
              </Link>
            )}
          </div>
          <button onClick={close} aria-label="Close menu">
            <CloseIcon />
          </button>
        </div>
        {!user && (
          <div className="promo">
            <span style={{ fontSize: 28 }}>🎁</span>
            <span style={{ flex: 1 }}>Unlock special offers &amp; great benefits</span>
            <Button $variant="outline" $size="sm" onClick={() => dispatch(openLogin())}>
              Login / Register
            </Button>
          </div>
        )}
        <nav>
          {items.map(({ icon: Ico, label, sub, path, login }) => (
            <button key={label} className="item" onClick={() => go(path, login)}>
              <Ico size={20} />
              <span className="label">
                {label}
                {sub && <small>{sub}</small>}
              </span>
              <ChevronRight size={16} />
            </button>
          ))}
          {user && (
            <button
              className="item"
              onClick={async () => {
                await logout();
                close();
                dispatch(pushToast('You have been signed out.'));
              }}
            >
              <LogoutIcon size={20} />
              <span className="label">Sign out</span>
            </button>
          )}
          {!user && (
            <button className="item" onClick={() => go('/', false)}>
              <UserIcon size={20} />
              <span className="label">Guest browsing</span>
            </button>
          )}
        </nav>
      </Drawer>
    </>
  );
}
