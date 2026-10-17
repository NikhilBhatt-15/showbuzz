import { createContext, useContext, useEffect } from 'react';

// Minimal document-head management: on the server <Title> records the page
// title (and status) into a context object; in the browser it sets
// document.title.
export const HeadContext = createContext(null);

export function Title({ children, status }) {
  const ctx = useContext(HeadContext);
  const text = `${children} | showbuzz`;
  if (ctx) {
    ctx.title = text;
    if (status) ctx.status = status;
  }
  useEffect(() => {
    document.title = text;
  }, [text]);
  return null;
}
