import { useEffect, useRef, useState } from 'react';

// Seconds remaining until expiresAt, corrected for client/server clock skew.
// Returns null until mounted so SSR and hydration markup agree.
export function useCountdown(expiresAt, serverTime) {
  const skew = useRef(0);
  const [left, setLeft] = useState(null);
  useEffect(() => {
    if (!expiresAt) return undefined;
    if (serverTime) skew.current = Date.now() - new Date(serverTime).getTime();
    const tick = () => setLeft(Math.max(0, (new Date(expiresAt).getTime() - (Date.now() - skew.current)) / 1000));
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [expiresAt, serverTime]);
  return left;
}

export function useDebounced(value, ms = 250) {
  const [v, setV] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setV(value), ms);
    return () => clearTimeout(id);
  }, [value, ms]);
  return v;
}

export function useMounted() {
  const [m, setM] = useState(false);
  useEffect(() => setM(true), []);
  return m;
}

export function useClickOutside(ref, onOutside, active = true) {
  useEffect(() => {
    if (!active) return undefined;
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) onOutside();
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [ref, onOutside, active]);
}
