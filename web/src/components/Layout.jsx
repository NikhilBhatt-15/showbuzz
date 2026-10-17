import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled from 'styled-components';
import { Container } from './ui.jsx';
import { ChevronDown, MenuIcon, SearchIcon, UserIcon } from './icons.jsx';
import { openCityModal, openLogin, toggleSidebar } from '../store/ui.js';
import { useGetMeQuery, useLazySearchQuery } from '../store/api.js';
import { useRegion } from '../lib/region.js';
import { useClickOutside, useDebounced } from '../lib/hooks.js';

export function Logo({ light }) {
  return (
    <LogoWrap to="/" aria-label="showbuzz home" $light={light}>
      show<span>buzz</span>
    </LogoWrap>
  );
}

const LogoWrap = styled(Link)`
  font-family: 'Arial Narrow', Roboto, Arial, sans-serif;
  font-size: 26px;
  font-weight: 700;
  letter-spacing: -0.5px;
  color: ${(p) => (p.$light ? '#fff' : '#333')};
  display: inline-flex;
  align-items: center;
  user-select: none;
  span {
    display: inline-block;
    background: ${(p) => p.theme.brand};
    color: #fff;
    padding: 0 5px 2px;
    margin-left: 2px;
    border-radius: 4px;
    transform: rotate(-8deg) translateY(2px);
    font-size: 22px;
  }
`;

const Bar = styled.header`
  background: #fff;
`;

const Top = styled(Container)`
  height: 64px;
  display: flex;
  align-items: center;
  gap: 24px;
`;

const Right = styled.div`
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 24px;
`;

const CityBtn = styled.button`
  background: none;
  border: none;
  font-size: 14px;
  color: #333;
  display: flex;
  align-items: center;
  gap: 6px;
  svg {
    width: 14px;
  }
`;

const SignIn = styled.button`
  background: ${(p) => p.theme.brand};
  border: none;
  color: #fff;
  border-radius: 4px;
  font-size: 12px;
  padding: 6px 16px;
  white-space: nowrap;
  &:hover {
    background: ${(p) => p.theme.brandDark};
  }
`;

const Hi = styled.button`
  background: none;
  border: none;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #333;
  font-size: 14px;
  span.av {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: ${(p) => p.theme.bg};
    display: grid;
    place-items: center;
  }
`;

const IconBtn = styled.button`
  background: none;
  border: none;
  color: #333;
  padding: 4px;
  display: flex;
`;

const Sub = styled.nav`
  background: ${(p) => p.theme.bg};
  border-bottom: 1px solid ${(p) => p.theme.border};
`;

const SubInner = styled(Container)`
  height: 40px;
  display: flex;
  align-items: center;
  gap: 20px;
  font-size: 14px;
  color: #333;
  a:hover {
    color: ${(p) => p.theme.brand};
  }
  .right {
    margin-left: auto;
    display: flex;
    gap: 24px;
    font-size: 12px;
  }
  @media (max-width: 720px) {
    overflow-x: auto;
    .right {
      display: none;
    }
  }
`;

export function Header() {
  const dispatch = useDispatch();
  const region = useRegion();
  const { data: me } = useGetMeQuery();
  const user = me?.user;
  return (
    <Bar>
      <Top>
        <Logo />
        <SearchBox region={region} />
        <Right>
          <CityBtn onClick={() => dispatch(openCityModal())} aria-label="Change city">
            {region.name} <ChevronDown />
          </CityBtn>
          {user ? (
            <Hi onClick={() => dispatch(toggleSidebar(true))}>
              <span className="av">
                <UserIcon size={18} />
              </span>
              Hi, {user.name || 'Guest'}
            </Hi>
          ) : (
            <SignIn onClick={() => dispatch(openLogin())}>Sign in</SignIn>
          )}
          <IconBtn onClick={() => dispatch(toggleSidebar(true))} aria-label="Open menu">
            <MenuIcon size={26} />
          </IconBtn>
        </Right>
      </Top>
      <Sub>
        <SubInner>
          <Link to={`/explore/movies-${region.slug}`}>Movies</Link>
          <Link to={`/explore/upcoming-movies-${region.slug}`}>Stream</Link>
          <Link to={`/explore/events-${region.slug}`}>Events</Link>
          <Link to={`/explore/plays-${region.slug}`}>Plays</Link>
          <Link to={`/explore/sports-${region.slug}`}>Sports</Link>
          <Link to={`/explore/activities-${region.slug}`}>Activities</Link>
          <div className="right">
            <Link to={`/${region.slug}/cinemas`}>Cinemas</Link>
            <Link to="/offers">Offers</Link>
            <Link to="/offers">Gift Cards</Link>
          </div>
        </SubInner>
      </Sub>
    </Bar>
  );
}

// ---- search ----------------------------------------------------------------

const SearchWrap = styled.div`
  position: relative;
  flex: 0 1 548px;
  @media (max-width: 720px) {
    display: none;
  }
`;

const SearchInput = styled.label`
  display: flex;
  align-items: center;
  gap: 10px;
  border: 1px solid ${(p) => p.theme.border};
  border-radius: 4px;
  height: 36px;
  padding: 0 12px;
  color: #777;
  input {
    border: none;
    outline: none;
    flex: 1;
    font-size: 14px;
    color: #333;
    background: transparent;
  }
`;

const Results = styled.div`
  position: absolute;
  top: 42px;
  left: 0;
  right: 0;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.18);
  z-index: 500;
  padding: 8px 0;
  max-height: 460px;
  overflow-y: auto;
  h6 {
    margin: 8px 16px;
    font-size: 12px;
    color: #999;
    font-weight: 500;
    text-transform: uppercase;
  }
  a {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 8px 16px;
    &:hover,
    &[data-active='true'] {
      background: ${(p) => p.theme.bg};
    }
  }
  img {
    width: 32px;
    height: 48px;
    object-fit: cover;
    border-radius: 4px;
  }
  .t {
    font-size: 14px;
    color: #333;
  }
  .s {
    font-size: 12px;
    color: #999;
  }
  .none {
    padding: 16px;
    color: #666;
  }
`;

function SearchBox({ region }) {
  const [q, setQ] = useState('');
  const [open, setOpen] = useState(false);
  const debounced = useDebounced(q.trim(), 200);
  const [trigger, { data, isFetching }] = useLazySearchQuery();
  const ref = useRef(null);
  const navigate = useNavigate();
  const close = useCallback(() => setOpen(false), []);
  useClickOutside(ref, close, open);

  useEffect(() => {
    if (debounced.length >= 2) trigger({ q: debounced, region: region.code }, true);
  }, [debounced, region.code, trigger]);

  const results = debounced.length >= 2 ? data?.results || [] : [];
  const groups = results.reduce((acc, r) => ((acc[r.type] = acc[r.type] || []).push(r), acc), {});

  return (
    <SearchWrap ref={ref}>
      <SearchInput>
        <SearchIcon size={18} />
        <input
          value={q}
          placeholder="Search for Movies, Events, Plays, Sports and Activities"
          onChange={(e) => {
            setQ(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && q.trim()) {
              setOpen(false);
              navigate(`/search?q=${encodeURIComponent(q.trim())}`);
            }
          }}
          aria-label="Search"
        />
      </SearchInput>
      {open && debounced.length >= 2 && (
        <Results>
          {!isFetching && results.length === 0 && <div className="none">No results for “{debounced}”</div>}
          {Object.entries(groups).map(([type, items]) => (
            <div key={type}>
              <h6>{type}</h6>
              {items.map((r) => (
                <Link key={r.url} to={r.url} onClick={() => setOpen(false)}>
                  {r.image ? <img src={r.image} alt="" /> : <span style={{ width: 32 }} />}
                  <span>
                    <div className="t">{r.title}</div>
                    <div className="s">{r.subtitle}</div>
                  </span>
                </Link>
              ))}
            </div>
          ))}
        </Results>
      )}
    </SearchWrap>
  );
}

// ---- footer ----------------------------------------------------------------

const Foot = styled.footer`
  background: ${(p) => p.theme.footer};
  color: #a5a5a5;
  margin-top: 0;
`;

const ListShow = styled.div`
  background: ${(p) => p.theme.footerDark};
  color: #fff;
  > div {
    display: flex;
    align-items: center;
    gap: 16px;
    min-height: 88px;
    flex-wrap: wrap;
    padding-top: 12px;
    padding-bottom: 12px;
  }
  b {
    font-size: 18px;
  }
  span {
    font-size: 16px;
    color: #ddd;
  }
  a {
    margin-left: auto;
    background: ${(p) => p.theme.brand};
    padding: 12px 20px;
    border-radius: 6px;
    font-weight: 500;
  }
`;

const Help = styled(Container)`
  display: flex;
  justify-content: space-around;
  padding: 32px 16px;
  text-align: center;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  div {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }
`;

const Links = styled(Container)`
  border-top: 1px solid #4a4a4f;
  padding: 24px 16px;
  h5 {
    margin: 0 0 8px;
    font-size: 12px;
    text-transform: uppercase;
    color: #a5a5a5;
    font-weight: 500;
  }
  p {
    font-size: 11px;
    line-height: 20px;
    color: #666;
    margin-bottom: 16px;
  }
  a:hover {
    color: #fff;
  }
`;

const Bottom = styled.div`
  text-align: center;
  padding: 24px 16px 32px;
  font-size: 11px;
  color: #666;
  .logo-line {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 24px;
    &::before,
    &::after {
      content: '';
      flex: 1;
      border-top: 1px solid #4a4a4f;
    }
  }
`;

export function Footer() {
  const region = useRegion();
  return (
    <Foot>
      <ListShow>
        <Container>
          <span style={{ fontSize: 28 }}>🎪</span>
          <b>List your Show</b>
          <span>Got a show, event, activity or a great experience? Partner with us &amp; get listed on showbuzz</span>
          <Link to="/offers">Contact today!</Link>
        </Container>
      </ListShow>
      <Help>
        <div>
          <span style={{ fontSize: 32 }}>🎧</span>24/7 Customer Care
        </div>
        <div>
          <span style={{ fontSize: 32 }}>🎟️</span>Resend Booking Confirmation
        </div>
        <div>
          <span style={{ fontSize: 32 }}>✉️</span>Subscribe to the Newsletter
        </div>
      </Help>
      <Links>
        <h5>Movies now showing in {region.name}</h5>
        <p>
          <Link to={`/explore/movies-${region.slug}`}>All movies</Link> | <Link to={`/explore/upcoming-movies-${region.slug}`}>Upcoming movies</Link> |{' '}
          <Link to={`/${region.slug}/cinemas`}>Cinemas in {region.name}</Link>
        </p>
        <h5>Events in {region.name}</h5>
        <p>
          <Link to={`/explore/events-${region.slug}`}>Events</Link> | <Link to={`/explore/plays-${region.slug}`}>Plays</Link> |{' '}
          <Link to={`/explore/sports-${region.slug}`}>Sports</Link> | <Link to={`/explore/activities-${region.slug}`}>Activities</Link>
        </p>
        <h5>Help</h5>
        <p>About Us | Current Openings | Press Release | Press Coverage | Sitemap | FAQs | Terms and Conditions | Privacy Policy</p>
      </Links>
      <Bottom>
        <div className="logo-line">
          <Logo light />
        </div>
        Copyright {new Date().getFullYear()} © showbuzz — an educational replica of a ticketing platform. All titles, venues and people are fictional.
        <br />
        No real tickets are sold and no real payments are processed.
      </Bottom>
    </Foot>
  );
}
