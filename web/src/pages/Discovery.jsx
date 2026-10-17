import { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import styled from 'styled-components';
import { useGetDiscoveryQuery } from '../store/api.js';
import { Button, Chip, Container, EmptyState, ErrorState, Skeleton } from '../components/ui.jsx';
import { PosterCard, WidgetRenderer } from '../components/Widgets.jsx';
import { ChevronDown, ChevronRight, ChevronUp } from '../components/icons.jsx';
import { Title } from '../lib/head.jsx';
import { openCityModal, setRegion } from '../store/ui.js';
import { writeRegionCookie } from '../lib/region.js';

// Keep the region cookie and store in sync with the city in the URL.
function useSyncRegion(slug) {
  const dispatch = useDispatch();
  const stored = useSelector((s) => s.ui.regionSlug);
  useEffect(() => {
    if (slug && slug !== stored) {
      writeRegionCookie(slug);
      dispatch(setRegion(slug));
    }
  }, [slug, stored, dispatch]);
}

export function HomePage() {
  const { city } = useParams();
  const location = useLocation();
  const dispatch = useDispatch();
  useSyncRegion(city);
  const { data, isLoading, isError, refetch } = useGetDiscoveryQuery({ page: 'home', region: city });
  useEffect(() => {
    if (new URLSearchParams(location.search).get('pickCity')) dispatch(openCityModal());
  }, [location.search, dispatch]);

  if (isError) return <ErrorState onRetry={refetch} />;
  if (isLoading || !data) return <HomeSkeleton />;
  return (
    <div style={{ paddingBottom: 32 }}>
      <Title>{data.header.title}</Title>
      {data.listings.map((w) => (
        <WidgetRenderer key={w.id} widget={w} styles={data.styles} />
      ))}
    </div>
  );
}

function HomeSkeleton() {
  return (
    <Container style={{ padding: '16px 0' }}>
      <Skeleton $h="300px" />
      <div style={{ display: 'flex', gap: 32, marginTop: 40 }}>
        {Array.from({ length: 5 }, (_, i) => (
          <Skeleton key={i} $w="222px" $h="378px" />
        ))}
      </div>
    </Container>
  );
}

// ---- listing (movies / upcoming / events / plays / sports / activities) ----

export const LISTING_RE = /^(movies|upcoming-movies|events|plays|sports|activities)-(.+)$/;

const Page = styled.div`
  background: ${(p) => p.theme.bg};
  padding: 32px 0 48px;
`;

const Layout = styled(Container)`
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 30px;
  @media (max-width: 900px) {
    grid-template-columns: 1fr;
  }
`;

const FilterPanel = styled.aside`
  h2 {
    font-size: 24px;
    margin-bottom: 24px;
  }
  .browse {
    margin-top: 16px;
  }
`;

const Accordion = styled.div`
  background: #fff;
  border-radius: 4px;
  margin-bottom: 12px;
  .bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 16px;
    cursor: pointer;
    user-select: none;
    color: ${(p) => (p.$open ? p.theme.brandDark : '#333')};
    font-size: 14px;
    svg {
      width: 16px;
    }
  }
  .clear {
    margin-left: auto;
    background: none;
    border: none;
    color: ${(p) => (p.$dirty ? p.theme.brandDark : '#666')};
    font-size: 12px;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    padding: 0 16px 20px;
  }
`;

const Main = styled.main`
  h1 {
    font-size: 24px;
    margin-bottom: 16px;
  }
  .quick {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 24px;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 32px 24px;
    @media (max-width: 1100px) {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }
    @media (max-width: 600px) {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
`;

const Soon = styled(Link)`
  display: flex;
  align-items: center;
  background: #fff;
  border-radius: 8px;
  padding: 24px 32px;
  margin-bottom: 32px;
  font-size: 20px;
  font-weight: 500;
  span {
    margin-left: auto;
    color: ${(p) => p.theme.brandDark};
    font-size: 16px;
    font-weight: 400;
    display: flex;
    align-items: center;
    gap: 4px;
  }
`;

const Crumbs = styled(Container)`
  font-size: 11px;
  color: #333;
  padding: 16px 0;
  a:hover {
    text-decoration: underline;
  }
`;

function FilterGroup({ filter, onToggle, onClear, defaultOpen }) {
  const [open, setOpen] = useState(defaultOpen);
  const dirty = filter.items.some((i) => i.selected);
  return (
    <Accordion $open={open} $dirty={dirty}>
      <div className="bar" onClick={() => setOpen((o) => !o)}>
        {open ? <ChevronUp /> : <ChevronDown />}
        {filter.title}
        <button
          className="clear"
          onClick={(e) => {
            e.stopPropagation();
            onClear(filter.queryParam);
          }}
        >
          Clear
        </button>
      </div>
      {open && (
        <div className="chips">
          {filter.items.map((it) => (
            <Chip key={it.code} $active={it.selected} onClick={() => onToggle(filter.queryParam, it.code)}>
              {it.label}
            </Chip>
          ))}
        </div>
      )}
    </Accordion>
  );
}

export function ListingPage() {
  const { listing } = useParams();
  const [, page, city] = listing.match(LISTING_RE) || [];
  const location = useLocation();
  const navigate = useNavigate();
  useSyncRegion(city);
  const search = location.search.replace(/^\?/, '');
  const { data, isLoading, isFetching, isError, refetch } = useGetDiscoveryQuery({ page, region: city, search }, { skip: !page });

  if (!page) return <EmptyState title="Page not found" message="The page you are looking for does not exist." />;
  if (isError) return <ErrorState onRetry={refetch} />;

  const params = new URLSearchParams(location.search);
  const toggle = (param, code) => {
    const cur = new Set((params.get(param) || '').split(',').filter(Boolean));
    cur.has(code) ? cur.delete(code) : cur.add(code);
    if (cur.size) params.set(param, [...cur].join(','));
    else params.delete(param);
    navigate({ search: params.toString() }, { replace: true });
  };
  const clear = (param) => {
    params.delete(param);
    navigate({ search: params.toString() }, { replace: true });
  };

  const grid = data?.listings.find((w) => w.type === 'grid');
  const soon = data?.listings.find((w) => w.type === 'coming-soon-strip');
  const quick = data?.filters.find((f) => f.quickFilter);

  return (
    <>
      <Page>
        {data && <Title>{data.header.title}</Title>}
        <Layout>
          <FilterPanel>
            <h2>Filters</h2>
            {data?.filters.map((f, i) => (
              <FilterGroup key={f.queryParam} filter={f} onToggle={toggle} onClear={clear} defaultOpen={i === 0} />
            ))}
            {page === 'movies' && (
              <Button className="browse" $variant="outline" $block onClick={() => navigate(`/${city}/cinemas`)}>
                Browse by Cinemas
              </Button>
            )}
          </FilterPanel>
          <Main style={{ opacity: isFetching && !isLoading ? 0.6 : 1 }}>
            <h1>{data?.header.title || ' '}</h1>
            {quick && (
              <div className="quick">
                {quick.items.map((it) => (
                  <Chip key={it.code} $pill $active={it.selected} onClick={() => toggle(quick.queryParam, it.code)}>
                    {it.label}
                  </Chip>
                ))}
              </div>
            )}
            {soon && (
              <Soon to={soon.ctaUrl}>
                {soon.title}
                <span>
                  {soon.ctaText} <ChevronRight size={16} />
                </span>
              </Soon>
            )}
            {isLoading || !data ? (
              <div className="grid">
                {Array.from({ length: 8 }, (_, i) => (
                  <Skeleton key={i} $h="340px" />
                ))}
              </div>
            ) : data.empty ? (
              <EmptyState title={data.empty.title} message={data.empty.message} cta={data.empty.cta} onCta={() => navigate({ search: '' })} />
            ) : (
              <div className="grid">
                {grid?.cards.map((c, k) => (
                  <PosterCard key={c.id} card={c} styles={data.styles} fluid widgetId={grid.id} position={k + 1} />
                ))}
              </div>
            )}
          </Main>
        </Layout>
      </Page>
      {data?.breadCrumbs && (
        <Crumbs>
          {data.breadCrumbs.map((c, i) => (
            <span key={i}>
              {i > 0 && ' → '}
              {c.route ? <Link to={c.route}>{c.label}</Link> : c.label}
            </span>
          ))}
        </Crumbs>
      )}
    </>
  );
}
