# How BookMyShow's web platform works (research notes)

This is the reference showbuzz was built from. Findings come from two kinds of evidence, labelled throughout:

- **Observed**: captured first-hand on 17 Oct 2026 by loading `in.bookmyshow.com` in Chrome under REA's browser-scenario capture (network, DOM, hydration state, screenshots). Pages: home, movie listing, movie detail, showtimes, seat layout.
- **Reported**: from public secondary sources (engineering job posts, conference talks, system-design write-ups). BookMyShow has published little about its own architecture, so treat these as unverified.

Nothing proprietary was copied. No bundles, artwork or payloads are included in this repository.

## 1. Edge and delivery

| Finding | Evidence |
|---|---|
| Cloudflare fronts the site with bot management. A headless browser gets a 403 challenge page (`/cdn-cgi/challenge-platform/...`). A headed browser passes. | Observed |
| Static JS/CSS is served from a separate CDN host, `generic-web-assets-in.bmscdn.com/public/build/chunks/js/…`. | Observed |
| Images go through an ImageKit-style transform CDN, e.g. `assets-in.bmscdn.com/discovery-catalog/events/tr:w-400,h-660,...:l-image,i-discovery...`. The **rating bar on posters is an image overlay** (`l-image` layer) composed by the CDN, not HTML. | Observed |
| Poster aspect ratio is 400×660; listing cards render at about 222×378. | Observed |

**showbuzz equivalent:** an Nginx gateway does prefix routing, caching (`X-Edge-Cache` HIT/MISS) and rate limiting. The catalog image service renders posters with the rating or date bar baked into the image, at the same 400×660 ratio.

## 2. Web frontend

| Finding | Evidence |
|---|---|
| A **custom React SSR** app, not Next.js: there is no `__NEXT_DATA__`, and webpack chunks are named `runtime`, `vendor-primary`, `vendor-secondary`, `app`, `common-code`, `app-explore-Container`. | Observed |
| State is **Redux Toolkit + RTK Query**. The page embeds `window.__INITIAL_STATE__` with RTK Query slices: `exploreApi`, `showtimesFunctionalApi`, `loginApi`, `headerApi`, `sidebarApi`, `seo`, plus regular reducers (`regions`, `cookies`, `appConfig`). | Observed |
| Data is **prefetched on the server**. The listing data (`getDiscoveryData-/explore/movies-mumbai-`) arrives already fulfilled in the HTML, and no XHR fetches it after load. | Observed |
| Styling uses **styled-components** (`sc-*` class names). The font is **Roboto**. Brand colours are `#F84464` (CTAs) and `#DC3558` (accents and links). | Observed |
| Errors go to **Sentry** (`sentry.javascript.react/9.47.1`). Analytics goes to GTM/GA4, CleverTap, Branch, Criteo and Facebook. There is a first-party **clickstream** endpoint `clickstream-in.bookmyshow.com/v1/eventstore/publish`. | Observed |
| In 2018 the web app migrated from legacy PHP to a React PWA. | Reported (ReactFoo 2018 talk) |

**showbuzz equivalent:** Vite + React 18 SSR on Express. It uses Redux Toolkit + RTK Query with route-level `loadData` prefetching, `window.__INITIAL_STATE__` hydration, styled-components SSR, `vendor-primary`/`vendor-secondary` chunking, Roboto with the same colours, and a batched clickstream client that publishes to Kafka.

## 3. Server-driven UI (discovery)

Observed. The discovery payload behind the home and listing pages is **server-driven UI**:

```text
{ scrollId, pageId, filterRoute,
  meta: { analytics, ldSchema, ... },
  header: { title, widget, actions },
  listings: [ { id, type: "flexbox" | carousel…, styleId, cards: [ { id, ctaUrl, image, text[], tags[], analytics } ] } ],
  styles: { "<styleId>": { font, fontColor, padding, ... } },
  filters: [ { queryParam: "languages", title, quickFilterIndex, items: [ { code, label, selected } ] } ],
  breadCrumbs, responseHeader: { cacheTagRouteSlug, skipCdnCache, isListingPage } }
```

The backend decides which widgets appear, in what order, and with which style IDs. The client only knows how to render widget types. Every card carries its own analytics payload.

**showbuzz equivalent:** `GET /api/explore/v1/discover/{home|movies|upcoming-movies|events|plays|sports|activities}` returns the same shape: `listings`, `cards`, a `styles` dictionary, `filters` and `breadCrumbs`. The web app's `WidgetRenderer` renders it generically.

## 4. Static/dynamic API split

Observed. Detail and showtime APIs are split into a cacheable **static** half and a per-request **dynamic** half:

- `GET /api/movies/v1/synopsis/init/static?eventcode=…` and `…/init/dynamic?…&regionCode=…`
- `showtimesFunctionalApi`: `fetchStaticShowtimes` (styles, seat legends, error copy, coachmarks) and `fetchPrimaryDynamic-<event>-<date>-<region>` (dates, language/format options, venue groups, showtime pills)
- `GET /api/movies-data/v4/showtimes-by-event/static?id=sl` and `GET /api/movies-data/seatlayout/v1/primary?eventCode&dateCode&regionCode&venueCode`

Observed data model details:

- Each showtime pill carries `sessionId`, `availStatus`, `cutOffDateTime` (the show start **plus 15 minutes**), `showDateCode` (`yyyymmdd`) and `showTimeCode`, plus category prices in a long-press/hover sheet (NORMAL ₹300 / EXECUTIVE ₹320 / PREMIUM ₹340 / RECLINER ₹850 in the captured show).
- Seat legend status codes: **1 Available, 2 Sold, 4 Bestseller, 88 Selected**.
- Venue cards carry `venueCode`, a cancellation flag, M-Ticket and F&B icons, favourite state, and `sortInfo` (distance, popularity).
- URL patterns:
  - `/explore/home/{city}`
  - `/explore/movies-{city}`
  - `/movies/{city}/{slug}/{ET code}`
  - `/movies/{city}/{slug}/buytickets/{ET code}/{yyyymmdd}`
  - `/movies/{city}/seat-layout/{ET code}/{venue}/{session}/{yyyymmdd}`
  - `/cinemas/{city}/{slug}/buytickets/{venue}/{date}`

**showbuzz equivalent:** the same endpoints, field names, status codes, cutoff rule and URL scheme.

## 5. Transaction core

| Finding | Evidence |
|---|---|
| The seat map comes from a **legacy ASP.NET service**: `POST services-in.bookmyshow.com/doTrans.aspx` with multipart fields `strAppCode=WEB`, `lngTransactionIdentifier=0`, `strCommand=GETSEATLAYOUT`, `strVenueCode`, `strParam1=<sessionId>`, …, `strFormat=json`. | Observed |
| The response is `{"BookMyShow":{"blnSuccess":"true","strData":"<encrypted blob>"}}`. The client decrypts the seat grid itself. (showbuzz does not attempt to reproduce this encryption.) | Observed |
| The "How many seats?" dialog (vehicle illustration, 1–10, category prices) comes before the seat map. Selected seats are green, sold seats grey, bestseller seats have a yellow border, and the screen sits at the bottom labelled "All eyes this way please". | Observed (screenshots) |
| Seats are temporarily held with Redis TTL locks (roughly 5–10 minutes), and the booking is committed after payment. | Reported (system-design write-ups) |
| Backend microservices are in **Go** (current job posts). Older write-ups mention Java/Spring and Node. | Reported |
| Kafka/RabbitMQ carry async work: notifications, ticket generation, inventory updates. Elasticsearch powers search, Redis/Memcached caching, and MySQL/Postgres the transactional data. | Reported |

**showbuzz equivalent:**

- A Go `trans` service that plays doTrans's role.
- Atomic multi-seat locks as Redis Lua scripts: a ZSET of seat → expiry plus a HASH of seat → transaction, swept lazily.
- A 10-minute pending transaction with a countdown.
- Pricing: ₹25 per ticket convenience fee + 18% IGST, F&B, and BookASmile ₹1 per ticket.
- A simulated payment gateway.
- A commit guarded by the `booking_seats (session_id, seat_id)` primary key.
- Kafka (Redpanda) events consumed by a `notify` service that e-mails tickets through Mailpit.
- Search uses Postgres `pg_trgm` instead of Elasticsearch, to keep the stack light.

## 6. What was deliberately not replicated

- The BookMyShow name, logo, artwork, posters and copy. showbuzz uses its own brand, procedurally generated art, and fictional titles, people, chains and venues.
- Third-party ad and tracking pixels, Cloudflare bot management, and the encrypted doTrans payload format.
- Real payments, real cinemas and real inventory.

## Sources (secondary)

- ReactFoo 2018: "Migrating from PHP legacy code to React" (BookMyShow) — https://hasgeek.com/reactfoo/2018-pune/sub/migrating-from-php-legacy-code-to-react-best-pract-RAY8d6BM2QxgpK9wQH5Avo
- BookMyShow SDE II Backend job post (Go microservices, AWS) — https://startup.jobs/software-development-engineer-ii-backend-technologies-bookmyshow-8378026
- "BookMyShow system design" — https://medium.com/@narengowda/bookmyshow-system-design-e268fefb56f5
- "Ticket booking system (BookMyShow) high-level design" — https://dev.to/arghya_majumder/ticket-booking-system-bookmyshow-high-level-system-design-3one
- Educative, "BookMyShow System Design" — https://www.educative.io/blog/bookmyshow-system-design
