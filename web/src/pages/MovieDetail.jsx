import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import styled from 'styled-components';
import { useGetSynopsisDynamicQuery, useGetSynopsisStaticQuery, useLazyGetShowtimesQuery } from '../store/api.js';
import { Button, Container, EmptyState, ErrorState, Modal, Skeleton, Spinner } from '../components/ui.jsx';
import { Carousel, PosterCard } from '../components/Widgets.jsx';
import { ChevronRight, PlayIcon, ShareIcon, StarFilled, ThumbUp } from '../components/icons.jsx';
import { compactCount, durationLabel, fullDate } from '../lib/format.js';
import { Title } from '../lib/head.jsx';
import { pushToast } from '../store/ui.js';
import { useRegion } from '../lib/region.js';
import { track } from '../lib/clickstream.js';

const Hero = styled.section`
  position: relative;
  color: #fff;
  background-color: #1a1a1a;
  background-image: linear-gradient(90deg, #1a1a1a 24.97%, #1a1a1a 38.3%, rgba(26, 26, 26, 0.04) 97.47%, #1a1a1a 100%), url(${(p) => p.$bg});
  background-size: cover;
  background-position: right center;
  .inner {
    display: flex;
    gap: 40px;
    padding: 32px 0;
    align-items: center;
    position: relative;
    @media (max-width: 720px) {
      flex-direction: column;
      align-items: flex-start;
    }
  }
  .poster {
    width: 261px;
    flex: 0 0 auto;
    border-radius: 12px;
    overflow: hidden;
    position: relative;
    img {
      width: 100%;
      aspect-ratio: 400 / 660;
      object-fit: cover;
    }
    .trailer {
      position: absolute;
      top: 45%;
      left: 50%;
      transform: translateX(-50%);
      background: rgba(0, 0, 0, 0.75);
      border-radius: 20px;
      padding: 6px 12px;
      font-size: 12px;
      display: flex;
      gap: 4px;
      align-items: center;
    }
    .status {
      background: #000;
      text-align: center;
      font-size: 12px;
      padding: 6px;
    }
  }
  h1 {
    font-size: 36px;
    font-weight: 700;
    margin-bottom: 24px;
  }
  .rating {
    display: flex;
    align-items: center;
    gap: 12px;
    background: #333;
    border-radius: 12px;
    padding: 16px 20px;
    max-width: 450px;
    font-size: 20px;
    font-weight: 700;
    margin-bottom: 24px;
    small {
      font-size: 16px;
      font-weight: 400;
    }
    button {
      margin-left: auto;
      background: #fff;
      color: #000;
      border: none;
      border-radius: 6px;
      padding: 8px 14px;
      font-size: 14px;
    }
  }
  .chips {
    display: flex;
    gap: 8px;
    margin-bottom: 16px;
    flex-wrap: wrap;
    span {
      background: #e5e5e5;
      color: #000;
      border-radius: 2px;
      padding: 3px 8px;
      font-size: 14px;
      font-weight: 500;
    }
  }
  .meta {
    font-size: 16px;
    margin-bottom: 32px;
  }
  .share {
    position: absolute;
    top: 32px;
    right: 0;
    background: rgba(0, 0, 0, 0.5);
    color: #fff;
    border: none;
    border-radius: 8px;
    padding: 10px 16px;
    font-size: 16px;
    display: flex;
    gap: 8px;
    align-items: center;
  }
`;

const Body = styled(Container)`
  padding: 32px 16px 48px;
  max-width: 1240px;
  section {
    padding: 24px 0;
    border-bottom: 1px solid ${(p) => p.theme.border};
    max-width: 925px;
  }
  h2 {
    font-size: 24px;
    margin-bottom: 16px;
  }
  .about {
    font-size: 16px;
    line-height: 1.5;
    color: #333;
  }
  .offers {
    display: flex;
    gap: 16px;
    overflow-x: auto;
  }
  .offer {
    flex: 0 0 300px;
    background: ${(p) => p.theme.amberSoft};
    border: 1px dashed #f5c26b;
    border-radius: 8px;
    padding: 12px 16px;
    b {
      display: block;
      margin-bottom: 4px;
    }
    span {
      font-size: 13px;
      color: #666;
    }
  }
  .person {
    flex: 0 0 auto;
    width: 120px;
    text-align: center;
    img {
      width: 120px;
      height: 120px;
      border-radius: 50%;
      margin-bottom: 10px;
    }
    div {
      font-size: 14px;
      color: #333;
    }
    small {
      color: #666;
      font-size: 12px;
    }
  }
  .reviews {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
    @media (max-width: 720px) {
      grid-template-columns: 1fr;
    }
  }
  .review {
    border: 1px solid ${(p) => p.theme.border};
    border-radius: 8px;
    padding: 16px;
    .top {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;
      font-size: 14px;
    }
    .tags {
      display: flex;
      gap: 6px;
      margin-bottom: 8px;
      span {
        font-weight: 700;
        font-size: 13px;
      }
    }
    p {
      color: #666;
      font-size: 14px;
      line-height: 1.5;
    }
    .foot {
      color: #999;
      font-size: 12px;
      margin-top: 12px;
    }
  }
`;

const FormatList = styled.div`
  padding: 24px;
  h3 {
    font-size: 18px;
    margin-bottom: 4px;
  }
  .lead {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
  }
  .lang {
    padding: 14px 0;
    border-top: 1px solid ${(p) => p.theme.border};
  }
  .lang b {
    display: block;
    margin-bottom: 10px;
    font-weight: 500;
  }
  .opts {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .opts button {
    border: 1px solid ${(p) => p.theme.brand};
    background: #fff;
    color: ${(p) => p.theme.brand};
    border-radius: 20px;
    padding: 6px 14px;
    font-size: 13px;
    &:hover {
      background: ${(p) => p.theme.brandSoft};
    }
  }
`;

export function MovieDetailPage() {
  const { city, code } = useParams();
  const region = useRegion();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const st = useGetSynopsisStaticQuery(code);
  const dy = useGetSynopsisDynamicQuery({ code, region: city });
  const [picker, setPicker] = useState(false);
  const [loadFormats, formats] = useLazyGetShowtimesQuery();

  if (st.isError) {
    return st.error?.status === 404 ? (
      <EmptyState title="Movie not found" message="This title may have been removed." cta="Go Home" onCta={() => navigate('/')} />
    ) : (
      <ErrorState onRetry={st.refetch} />
    );
  }
  if (!st.data) return <Container style={{ padding: 32 }}><Skeleton $h="420px" /></Container>;

  const { movie, cast, crew, posterUrl, backdropUrl } = st.data;
  const d = dy.data;
  const comingSoon = st.data.comingSoon;
  const buyBase = `/movies/${city}/${movie.slug}/buytickets/${movie.code}`;

  const book = async () => {
    track('book_tickets_clicked', { event_code: movie.code });
    if (d && !d.hasShows) {
      dispatch(pushToast(`${movie.title} is not playing in ${region.name} right now.`, 'error'));
      return;
    }
    setPicker(true);
    const res = await loadFormats({ code: movie.code, region: city }, true);
    if (res.data?.formats?.length === 1) {
      setPicker(false);
      navigate(buyBase);
    }
  };

  const byLang = (formats.data?.formats || []).reduce((acc, f) => ((acc[f.language] = acc[f.language] || []).push(f), acc), {});

  return (
    <>
      <Title>{`${movie.title} (${new Date(movie.releaseDate).getFullYear()}) - Movie | Reviews, Cast & Release Date in ${region.name}`}</Title>
      <Hero $bg={backdropUrl}>
        <Container className="inner">
          <div className="poster">
            <img src={posterUrl} alt={movie.title} />
            <span className="trailer">
              <PlayIcon size={12} /> Trailer
            </span>
            <div className="status">{comingSoon ? `Releasing on ${fullDate(movie.releaseDate)}` : 'In cinemas'}</div>
          </div>
          <div>
            <h1>{movie.title}</h1>
            {movie.rating > 0 ? (
              <div className="rating">
                <StarFilled size={24} /> {movie.rating.toFixed(1)}/10 <small>({compactCount(movie.votes)}+ Votes)</small>
                <ChevronRight size={16} />
                <button onClick={() => dispatch(pushToast('Thanks! Ratings open after you watch the movie.'))}>Rate now</button>
              </div>
            ) : (
              <div className="rating">
                <ThumbUp size={24} /> {compactCount(movie.likes)}+ <small>are interested</small>
                <button onClick={() => dispatch(pushToast("Great! We'll let you know when bookings open."))}>I'm interested</button>
              </div>
            )}
            <div className="chips">
              <span>{movie.formats.join(', ')}</span>
              <span>{movie.languages.join(', ')}</span>
            </div>
            <div className="meta">
              {durationLabel(movie.durationMin)} • {movie.genres.join(', ')} • {movie.certificate} • {fullDate(movie.releaseDate)}
            </div>
            <Button $size="lg" onClick={book}>
              Book tickets
            </Button>
          </div>
          <button
            className="share"
            onClick={() => {
              navigator.clipboard?.writeText(window.location.href);
              dispatch(pushToast('Link copied to clipboard'));
            }}
          >
            <ShareIcon /> Share
          </button>
        </Container>
      </Hero>

      <Body>
        <section>
          <h2>About the movie</h2>
          <p className="about">{movie.synopsis}</p>
        </section>
        {d?.offers && (
          <section>
            <h2>Top offers for you</h2>
            <div className="offers">
              {d.offers.map((o) => (
                <div className="offer" key={o.title}>
                  <b>🏷️ {o.title}</b>
                  <span>{o.text}</span>
                </div>
              ))}
            </div>
          </section>
        )}
        <section>
          <h2>Cast</h2>
          <Carousel gap={24} arrowTop="30%">
            {cast.map((p) => (
              <div className="person" key={p.name}>
                <img src={p.avatar} alt={p.name} />
                <div>{p.name}</div>
                <small>{p.role}</small>
              </div>
            ))}
          </Carousel>
        </section>
        <section>
          <h2>Crew</h2>
          <Carousel gap={24} arrowTop="30%">
            {crew.map((p) => (
              <div className="person" key={p.name + p.role}>
                <img src={p.avatar} alt={p.name} />
                <div>{p.name}</div>
                <small>{p.role}</small>
              </div>
            ))}
          </Carousel>
        </section>
        {d?.reviews?.length > 0 && (
          <section>
            <h2>Top reviews</h2>
            <p style={{ color: '#666', marginBottom: 16, fontSize: 14 }}>Summary of {compactCount(movie.votes)} reviews.</p>
            <div className="reviews">
              {d.reviews.map((r, i) => (
                <div className="review" key={i}>
                  <div className="top">
                    <b>{r.user}</b>
                    <span>
                      <StarFilled size={14} /> {r.rating}/10
                    </span>
                  </div>
                  <div className="tags">
                    {r.tags.map((t) => (
                      <span key={t}>{t}</span>
                    ))}
                  </div>
                  <p>{r.text}</p>
                  <div className="foot">
                    👍 {r.likes} · {r.ago}
                  </div>
                </div>
              ))}
            </div>
          </section>
        )}
        {d?.similar?.length > 0 && (
          <section style={{ borderBottom: 'none', maxWidth: 'none' }}>
            <h2>You might also like</h2>
            <Carousel arrowTop="38%">
              {d.similar.map((c, k) => (
                <PosterCard key={c.id} card={c} widgetId="similar" position={k + 1} />
              ))}
            </Carousel>
          </section>
        )}
        <p style={{ fontSize: 11, marginTop: 24 }}>
          <Link to={`/explore/home/${city}`}>Home</Link> → <Link to={`/explore/movies-${city}`}>Movies in {region.name}</Link> → {movie.title}
        </p>
      </Body>

      <Modal open={picker} onClose={() => setPicker(false)} width={520} label="Select language and format">
        <FormatList>
          <h3>{movie.title}</h3>
          <p className="lead">Select language and format</p>
          {formats.isFetching || !formats.data ? (
            <Spinner $size={28} />
          ) : (
            Object.entries(byLang).map(([lang, list]) => (
              <div className="lang" key={lang}>
                <b>{lang}</b>
                <div className="opts">
                  {list.map((f) => (
                    <button key={f.label} onClick={() => navigate(`${buyBase}?lang=${encodeURIComponent(f.language)}&format=${encodeURIComponent(f.format)}`)}>
                      {f.format}
                    </button>
                  ))}
                </div>
              </div>
            ))
          )}
        </FormatList>
      </Modal>
    </>
  );
}

