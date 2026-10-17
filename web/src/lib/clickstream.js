// Batched analytics, modelled on BookMyShow's clickstream eventstore: events
// are queued in the browser and flushed to the clickstream API, which
// publishes them to Kafka.
const queue = [];
let timer = null;

function flush() {
  timer = null;
  if (!queue.length) return;
  const body = JSON.stringify({ events: queue.splice(0, 50) });
  try {
    if (navigator.sendBeacon) navigator.sendBeacon('/api/clickstream/v1/publish', body);
    else fetch('/api/clickstream/v1/publish', { method: 'POST', body, keepalive: true });
  } catch {
    /* analytics must never break the page */
  }
}

export function track(event_name, props = {}) {
  if (typeof window === 'undefined') return;
  queue.push({ event_name, schema_version: '3.0', page: location.pathname, ts: Date.now(), ...props });
  if (!timer) timer = setTimeout(flush, 4000);
}

if (typeof window !== 'undefined') {
  window.addEventListener('pagehide', flush);
}
