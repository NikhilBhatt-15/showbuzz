import { useCallback, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';
import { Container, SectionTitle } from './ui.jsx';
import { ChevronLeft, ChevronRight } from './icons.jsx';
import { track } from '../lib/clickstream.js';

// Renders the discovery API's server-driven "listings". Each widget type maps
// to one component; the server controls order, content and style IDs.

const fonts = {
  'subtitle-large': 'font-size:18px;font-weight:500;',
  'body-regular': 'font-size:16px;font-weight:400;',
  title: 'font-size:24px;font-weight:700;',
};

export function StyledText({ text, styles }) {
  const st = styles?.[text.styleId] || {};
  return (
    <TextLine style={{ color: st.fontColor }} $font={fonts[st.font] || ''}>
      {text.text}
    </TextLine>
  );
}

const TextLine = styled.div`
  ${(p) => p.$font}
  line-height: 1.35;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  & + & {
    margin-top: 4px;
  }
`;

// ---- poster card -----------------------------------------------------------

const PosterLink = styled(Link)`
  display: block;
  width: ${(p) => (p.$fluid ? '100%' : '222px')};
  flex: 0 0 auto;
  .img {
    position: relative;
    border-radius: 8px;
    overflow: hidden;
    aspect-ratio: 400 / 660;
    background: #ddd;
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      transition: transform 0.3s;
    }
  }
  &:hover img {
    transform: scale(1.03);
  }
  .tag {
    position: absolute;
    top: 0;
    right: 0;
    background: ${(p) => p.theme.brand};
    color: #fff;
    font-size: 12px;
    letter-spacing: 0.5px;
    padding: 2px 8px;
    border-bottom-left-radius: 6px;
  }
  .txt {
    padding-top: 12px;
  }
`;

export function PosterCard({ card, styles, fluid, widgetId, position }) {
  return (
    <PosterLink
      to={card.ctaUrl}
      $fluid={fluid}
      onClick={() => track(card.kind === 'movie' ? 'movie_card_clicked' : 'event_card_clicked', { id: card.id, widget_id: widgetId, in_widget_position: position })}
    >
      <div className="img">
        <img src={card.image.url} alt={card.image.altText} loading="lazy" />
        {card.tag && <span className="tag">{card.tag}</span>}
      </div>
      <div className="txt">
        {card.text.map((t, i) => (
          <StyledText key={i} text={t} styles={styles} />
        ))}
      </div>
    </PosterLink>
  );
}

// ---- carousel --------------------------------------------------------------

const Rail = styled.div`
  position: relative;
  .track {
    display: flex;
    gap: ${(p) => p.$gap}px;
    overflow-x: auto;
    scroll-behavior: smooth;
    scrollbar-width: none;
    &::-webkit-scrollbar {
      display: none;
    }
  }
`;

const Arrow = styled.button`
  position: absolute;
  top: ${(p) => p.$top || '40%'};
  ${(p) => (p.$side === 'left' ? 'left: -18px;' : 'right: -18px;')}
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, 0.65);
  color: #fff;
  display: grid;
  place-items: center;
  z-index: 2;
  &:hover {
    background: rgba(0, 0, 0, 0.85);
  }
  @media (max-width: 720px) {
    display: none;
  }
`;

export function Carousel({ children, gap = 32, arrowTop }) {
  const ref = useRef(null);
  const [edges, setEdges] = useState({ start: true, end: false });
  const update = useCallback(() => {
    const el = ref.current;
    if (!el) return;
    setEdges({ start: el.scrollLeft <= 4, end: el.scrollLeft + el.clientWidth >= el.scrollWidth - 4 });
  }, []);
  useEffect(() => {
    update();
    const el = ref.current;
    el?.addEventListener('scroll', update, { passive: true });
    window.addEventListener('resize', update);
    return () => {
      el?.removeEventListener('scroll', update);
      window.removeEventListener('resize', update);
    };
  }, [update]);
  const move = (dir) => ref.current?.scrollBy({ left: dir * (ref.current.clientWidth - 40) });
  return (
    <Rail $gap={gap}>
      {!edges.start && (
        <Arrow $side="left" $top={arrowTop} onClick={() => move(-1)} aria-label="Previous">
          <ChevronLeft />
        </Arrow>
      )}
      <div className="track" ref={ref}>
        {children}
      </div>
      {!edges.end && (
        <Arrow $side="right" $top={arrowTop} onClick={() => move(1)} aria-label="Next">
          <ChevronRight />
        </Arrow>
      )}
    </Rail>
  );
}

// ---- banner carousel (hero) ------------------------------------------------

const Hero = styled.section`
  position: relative;
  overflow: hidden;
  padding: 8px 0 0;
  background: ${(p) => p.theme.bg};
  .slides {
    display: flex;
    gap: 8px;
    transition: transform 0.5s ease;
  }
  .slide {
    flex: 0 0 min(1240px, calc(100vw - 40px));
    aspect-ratio: 1240 / 300;
    border-radius: 8px;
    overflow: hidden;
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }
  .dots {
    position: absolute;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: 6px;
    button {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      border: none;
      padding: 0;
      background: rgba(255, 255, 255, 0.5);
      &[data-active='true'] {
        background: #fff;
      }
    }
  }
  .nav {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    width: 40px;
    height: 80px;
    border: none;
    background: rgba(0, 0, 0, 0.35);
    color: #fff;
    display: grid;
    place-items: center;
    z-index: 2;
  }
  .prev {
    left: 0;
    border-radius: 0 8px 8px 0;
  }
  .next {
    right: 0;
    border-radius: 8px 0 0 8px;
  }
`;

export function BannerCarousel({ cards }) {
  const [i, setI] = useState(0);
  const n = cards.length;
  useEffect(() => {
    const id = setInterval(() => setI((x) => (x + 1) % n), 5000);
    return () => clearInterval(id);
  }, [n]);
  // Centre the active slide and let its neighbours peek on both sides.
  const offset = `calc(50vw - min(620px, 50vw - 20px) - ${i} * (min(1240px, 100vw - 40px) + 8px))`;
  return (
    <Hero>
      <div className="slides" style={{ transform: `translateX(${offset})` }}>
        {cards.map((c, k) => (
          <Link key={c.id} className="slide" to={c.ctaUrl} onClick={() => track('banner_clicked', { id: c.id, position: k + 1 })} aria-hidden={k !== i}>
            <img src={c.image.url} alt={c.image.altText} />
          </Link>
        ))}
      </div>
      <button className="nav prev" onClick={() => setI((i - 1 + n) % n)} aria-label="Previous banner">
        <ChevronLeft />
      </button>
      <button className="nav next" onClick={() => setI((i + 1) % n)} aria-label="Next banner">
        <ChevronRight />
      </button>
      <div className="dots">
        {cards.map((c, k) => (
          <button key={c.id} data-active={k === i} onClick={() => setI(k)} aria-label={`Banner ${k + 1}`} />
        ))}
      </div>
    </Hero>
  );
}

// ---- section wrappers ------------------------------------------------------

const Section = styled.section`
  background: ${(p) => (p.$dark ? p.theme.premiere : 'transparent')};
  padding: 32px 0 ${(p) => (p.$dark ? '40px' : '8px')};
  margin: ${(p) => (p.$dark ? '24px 0' : '0')};
  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 16px;
  }
  .sub {
    color: ${(p) => (p.$dark ? '#ccc' : p.theme.muted)};
    font-size: 16px;
    margin-top: 4px;
  }
  .cta {
    color: ${(p) => p.theme.brandDark};
    font-size: 16px;
  }
  .premiere-logo {
    display: inline-flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 20px;
    color: #fff;
    font-weight: 700;
    letter-spacing: 3px;
    font-size: 14px;
    span {
      background: ${(p) => p.theme.brand};
      border-radius: 4px;
      padding: 4px 8px;
    }
  }
`;

const Tile = styled(Link)`
  flex: 0 0 auto;
  width: 222px;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  img {
    width: 100%;
    height: 100%;
    transition: transform 0.3s;
  }
  &:hover img {
    transform: scale(1.04);
  }
`;

const Strip = styled(Link)`
  display: block;
  border-radius: 8px;
  overflow: hidden;
  margin-top: 24px;
  img {
    width: 100%;
    aspect-ratio: 1240 / 120;
    object-fit: cover;
  }
`;

export function WidgetRenderer({ widget, styles }) {
  const dark = widget.theme === 'dark';
  switch (widget.type) {
    case 'banner-carousel':
      return <BannerCarousel cards={widget.cards} />;
    case 'image-strip':
      return (
        <Container>
          {widget.cards.map((c) => (
            <Strip key={c.id} to={c.ctaUrl}>
              <img src={c.image.url} alt={c.image.altText} />
            </Strip>
          ))}
        </Container>
      );
    case 'category-carousel':
      return (
        <Section>
          <Container>
            <div className="head">
              <SectionTitle>{widget.title}</SectionTitle>
            </div>
            <Carousel gap={24} arrowTop="45%">
              {widget.cards.map((c) => (
                <Tile key={c.id} to={c.ctaUrl} onClick={() => track('category_clicked', { id: c.id })}>
                  <img src={c.image.url} alt={c.image.altText} />
                </Tile>
              ))}
            </Carousel>
          </Container>
        </Section>
      );
    case 'horizontal-carousel':
      return (
        <Section $dark={dark}>
          <Container>
            {dark && (
              <div className="premiere-logo">
                <span>▶</span> PREMIERE
              </div>
            )}
            <div className="head">
              <div>
                <SectionTitle $dark={dark}>{widget.title}</SectionTitle>
                {widget.subtitle && <div className="sub">{widget.subtitle}</div>}
              </div>
              {widget.ctaUrl && (
                <Link className="cta" to={widget.ctaUrl}>
                  {widget.ctaText} ›
                </Link>
              )}
            </div>
            <Carousel arrowTop="38%">
              {widget.cards.map((c, k) => (
                <PosterCard key={c.id} card={c} styles={styles} widgetId={widget.id} position={k + 1} />
              ))}
            </Carousel>
          </Container>
        </Section>
      );
    default:
      return null;
  }
}
