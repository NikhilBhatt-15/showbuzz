const TZ = 'Asia/Kolkata';

export const rupee = (n, decimals = 0) =>
  '₹' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: decimals, maximumFractionDigits: decimals });

export const rs = (n) => 'Rs. ' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

const fmt = (opts) => new Intl.DateTimeFormat('en-IN', { timeZone: TZ, ...opts });

export const showDateLong = (iso) => {
  const d = new Date(iso);
  const day = fmt({ weekday: 'short' }).format(d);
  const date = fmt({ day: 'numeric', month: 'long', year: 'numeric' }).format(d);
  return `${day}, ${date}`;
};

export const showTime = (iso) => fmt({ hour: '2-digit', minute: '2-digit', hour12: true }).format(new Date(iso)).toUpperCase();

export const shortDate = (iso) => fmt({ weekday: 'short', day: '2-digit', month: 'short' }).format(new Date(iso));

export const fullDate = (iso) => fmt({ day: 'numeric', month: 'short', year: 'numeric' }).format(new Date(iso));

export const durationLabel = (min) => `${Math.floor(min / 60)}h ${min % 60}m`;

export const compactCount = (n) => {
  if (n >= 1000) {
    const v = n / 1000;
    return (v >= 100 ? v.toFixed(0) : v.toFixed(1).replace(/\.0$/, '')) + 'K';
  }
  return String(n);
};

export const dateCodeToday = () => {
  const parts = fmt({ year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(new Date());
  const get = (t) => parts.find((p) => p.type === t).value;
  return `${get('year')}${get('month')}${get('day')}`;
};

export const mmss = (secs) => {
  const s = Math.max(0, Math.floor(secs));
  return `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`;
};

export const seatSummary = (lines = []) =>
  lines.filter((l) => l.seats?.length).map((l) => `${l.areaName} - ${l.seats.join(', ')}`).join(' | ');
