import { useRef, useState } from 'react';
import styled, { css } from 'styled-components';
import { MoonIcon, SunIcon } from './icons.jsx';
import { rupee } from '../lib/format.js';

const color = (theme, status) => (status === 'FAST_FILLING' ? theme.amber : status === 'SOLD_OUT' ? '#ccc' : theme.green);

const Pill = styled.button`
  position: relative;
  width: 107px;
  min-height: 40px;
  border: 1px solid ${(p) => color(p.theme, p.$status)};
  border-left-width: 4px;
  border-radius: 2px;
  background: ${(p) => (p.$active ? p.theme.green : '#fff')};
  color: ${(p) => (p.$active ? '#fff' : p.$status === 'SOLD_OUT' ? '#bbb' : '#666')};
  font-size: 13px;
  padding: 4px 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  line-height: 1.3;
  ${(p) => p.$active && css`border-color: ${p.theme.green};`}
  small {
    font-size: 8px;
    text-transform: uppercase;
    letter-spacing: 0.3px;
  }
  .badge {
    position: absolute;
    top: -8px;
    right: -8px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #888;
    color: #fff;
    display: grid;
    place-items: center;
  }
  &:hover:not(:disabled) {
    background: ${(p) => (p.$active ? p.theme.green : p.theme.greenSoft)};
  }
`;

const Tip = styled.div`
  position: absolute;
  bottom: calc(100% + 10px);
  left: 50%;
  transform: translateX(-50%);
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.2);
  padding: 12px 16px;
  z-index: 50;
  display: flex;
  gap: 18px;
  white-space: nowrap;
  pointer-events: none;
  text-align: center;
  .cat {
    font-size: 11px;
    color: #666;
    text-transform: uppercase;
  }
  .price {
    font-size: 14px;
    color: #333;
    margin: 2px 0;
  }
  .st {
    font-size: 10px;
    color: ${(p) => p.theme.green};
    &[data-s='FAST_FILLING'] {
      color: ${(p) => p.theme.amber};
    }
    &[data-s='SOLD_OUT'] {
      color: #999;
    }
  }
  &::after {
    content: '';
    position: absolute;
    top: 100%;
    left: 50%;
    margin-left: -6px;
    border: 6px solid transparent;
    border-top-color: #fff;
  }
`;

const statusLabel = { AVAILABLE: 'Available', FAST_FILLING: 'Filling fast', SOLD_OUT: 'Sold out' };

export function ShowtimePill({ show, onClick, active, tooltip = true }) {
  const [hover, setHover] = useState(false);
  const timer = useRef(null);
  const sold = show.status === 'SOLD_OUT';
  return (
    <div style={{ position: 'relative' }}
      onMouseEnter={() => (timer.current = setTimeout(() => setHover(true), 250))}
      onMouseLeave={() => {
        clearTimeout(timer.current);
        setHover(false);
      }}
    >
      <Pill $status={show.status} $active={active} disabled={sold} onClick={onClick} aria-label={`${show.time}, ${statusLabel[show.status]}`}>
        {show.time}
        {show.attribute && <small>{show.attribute}</small>}
        {(show.lateNight || show.earlyMorning) && <span className="badge">{show.lateNight ? <MoonIcon size={10} /> : <SunIcon size={10} />}</span>}
      </Pill>
      {tooltip && hover && !sold && (
        <Tip>
          {show.categories.map((c) => (
            <div key={c.code}>
              <div className="price">{rupee(c.price)}</div>
              <div className="cat">{c.name}</div>
              <div className="st" data-s={c.status}>
                {statusLabel[c.status]}
              </div>
            </div>
          ))}
        </Tip>
      )}
    </div>
  );
}

export const PillRow = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 16px 24px;
`;

export const Legend = styled.div`
  display: flex;
  gap: 16px;
  font-size: 10px;
  color: #666;
  text-transform: uppercase;
  span::before {
    content: '';
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 6px;
    background: ${(p) => p.theme.green};
  }
  span.ff::before {
    background: ${(p) => p.theme.amber};
  }
`;
