# showbuzz

A full-stack, Docker-based replica of the **BookMyShow** ticketing platform. It reproduces the user flows, page layouts and backend architecture: city selection, discovery, movie and event pages, showtimes, live seat maps with seat locking, F&B add-ons, OTP sign-in, checkout, e-tickets and cancellations.

The architecture follows what BookMyShow's live web app actually does, captured first-hand with REA browser instrumentation. See [docs/RESEARCH.md](docs/RESEARCH.md) for the evidence.

> **Educational project.** showbuzz uses its own brand and procedurally generated artwork. Every movie, person, cinema chain and venue is fictional. No real tickets or payments are involved.

## Features

- **Server-rendered React** with Redux Toolkit + RTK Query. Data is prefetched per route on the server, then hydrated from `window.__INITIAL_STATE__`.
- **Server-driven discovery UI**: the API returns widgets, cards, styles and filters, and the client renders them generically (home, movies, upcoming, events, plays, sports, activities).
- **City picker** with popular-city icons, "view all cities" and geolocation detection. The city is kept in the URL and a cookie.
- **Movie pages**: dark hero with rating box, formats and languages, cast and crew, reviews, offers, similar titles, and a language/format picker.
- **Showtimes**: 7-day date strip, language-format selector, price and time filters, sorting, cinema search and favourite cinemas. Pills are coloured by availability (available / fast filling / sold out) and show category prices on hover. An A-certificate advisory appears for adult films.
- **Seat layout**: a "How many seats?" dialog with vehicle illustrations, a multi-tier auditorium map, BookMyShow-style auto-selection of contiguous seats, bestseller seats, a live refresh every 15 seconds, and a sibling-showtime strip.
- **Seat locking**: atomic multi-seat Redis Lua locks with a 10-minute hold and lazy expiry. Concurrent buyers can never get the same seat, and a Postgres primary key is the final guard.
- **Checkout**: booking summary with convenience fee + 18% IGST breakdown, F&B menu with steppers, BookASmile donation, countdown timer, and session-expiry handling.
- **Payment**: simulated UPI, card, net banking and wallets, including decline scenarios.
- **E-ticket** with QR code, purchase history, profile editing, and **cancellation with refund**.
- **Events**: general-admission ticket tiers with capacity holds.
- **Passwordless sign-in**: OTP by e-mail or mobile, JWT in an httpOnly cookie.
- **Event-driven notifications**: Kafka topics consumed by a notifier that e-mails OTPs, tickets and cancellation receipts.
- **Search** across movies, events and cinemas (Postgres trigram).
- **Edge gateway** with caching (`X-Edge-Cache`), rate limiting and prefix routing.
- **Clickstream** analytics batched from the browser into Kafka.

## Architecture

```text
                         ┌──────────────────────────── Nginx gateway :8080 ─────────────────────────────┐
 browser ──── HTTP ────▶ │ edge cache · rate limits · routes /api/* by prefix · everything else → web   │
                         └──────┬──────────┬────────────┬──────────────┬──────────────┬─────────────────┘
                                │          │            │              │              │
                  web (React SSR, :3000)   │            │              │              │
                  RTK Query prefetch ──────┘            │              │              │
                                   catalog :8081   showtimes :8082   trans :8083    user :8084
                                   discovery, details, showtimes,    seat grid,     OTP login,
                                   search, images,  cinemas,         locks, txns,   profiles
                                   clickstream      seat primary     payment,
                                                                     bookings
                                        │               │              │              │
                         ┌──────────────┴───────────────┴──────────────┴──────────────┴──────┐
                         │ Postgres 16 (catalog, shows, bookings) · Redis 7 (locks, txns, OTP, │
                         │ cache) · Redpanda (Kafka API: booking.*, otp.requested, clickstream)│
                         └───────────────────────────────────┬───────────────────────────────┘
                                                             │ consume
                                                     notify ─┴─▶ Mailpit (SMTP + web inbox :8025)
```

| Layer | Tech | Mirrors on BookMyShow |
|---|---|---|
| Edge | Nginx 1.27 | Cloudflare + bmscdn CDN |
| Web | React 18, Vite 5 SSR on Express, Redux Toolkit 2 + RTK Query, styled-components 6, React Router 6 | Custom React SSR, RTK Query slices, styled-components |
| Services | Go 1.24, chi, pgx, go-redis, franz-go | Go microservices |
| Data | Postgres 16, Redis 7, Redpanda 24 | MySQL/Postgres, Redis, Kafka |
| Mail | Mailpit | Notification service |

### API surface (selected)

| Endpoint | Service | Notes |
|---|---|---|
| `GET /api/explore/v1/discover/{page}?region=` | catalog | Server-driven listings, filters, styles |
| `GET /api/movies/v1/synopsis/init/static?eventCode=` | catalog | Cacheable details |
| `GET /api/movies/v1/synopsis/init/dynamic?eventCode=&regionCode=` | catalog | Ratings, reviews, availability in city |
| `GET /api/movies-data/v4/showtimes-by-event/primary-dynamic` | showtimes | Dates, formats, venues, live availability |
| `GET /api/movies-data/seatlayout/v1/primary?sessionId=` | showtimes | Show strip + category prices |
| `GET /api/trans/v1/seat-layout/{session}` | trans | Grid with status codes 1/2/4 |
| `POST /api/trans/v1/transactions` | trans | Lock seats → pending transaction |
| `PATCH /api/trans/v1/transactions/{id}` | trans | F&B and donation |
| `POST /api/trans/v1/transactions/{id}/pay` | trans | Pay + commit booking |
| `POST /api/trans/v1/bookings/{id}/cancel` | trans | Cancel + refund, frees seats |
| `POST /api/user/v1/otp`, `POST /api/user/v1/verify` | user | Passwordless sign-in |

## Install & run

Requires Docker Desktop (or Docker Engine with Compose v2).

```bash
git clone https://github.com/NikhilBhatt-15/showbuzz.git
cd showbuzz
docker compose up -d --build
```

| URL | What |
|---|---|
| http://localhost:8080 | The app |
| http://localhost:8025 | Mailpit inbox (OTP and ticket e-mails) |

The first start seeds the catalog and generates a rolling seven-day schedule: about 7,300 shows across 80 cinemas in 20 cities, with realistic pre-sold seats. This takes about a minute.

### Demo tips

- **Sign in** with any e-mail address or 10-digit mobile number. With `DEV_SHOW_OTP=true` (the default) the OTP is shown in the dialog, and e-mail OTPs also land in Mailpit.
- **Cards:** `4111 1111 1111 1111` succeeds and `4000 0000 0000 0002` is declined. **UPI:** any valid ID succeeds and `fail@upi` is declined.
- Galaxy, Cinestar, Regal and Moviedome cinemas sell F&B. Silverline does not.
- Set `SEAT_HOLD_SECONDS=60 docker compose up -d trans` to see seat holds expire quickly.

### Configuration

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Host port for the gateway |
| `JWT_SECRET` | `change-me-in-production` | Session signing key |
| `DEV_SHOW_OTP` | `true` | Return the OTP in the API response (demo only) |
| `SEAT_HOLD_SECONDS` | `600` | Seat lock / transaction lifetime |

## Build & develop

```bash
# Backend: unit tests, vet, build
cd services
go test ./...
go vet ./...
go build ./cmd/showbuzz

# Frontend: production build (client + SSR bundles)
cd web
npm install
npm run build
# Run SSR locally against the dockerised gateway:
PROXY_API=http://localhost:8080 API_INTERNAL_URL=http://localhost:8080 NODE_ENV=production node server.js   # → :3000
```

A single Go binary hosts every service: `showbuzz seed|catalog|showtimes|trans|user|notify`.

## How it works

1. **Rendering.** Express matches the URL against `src/routes.jsx`. Each route's `loadData` dispatches RTK Query fetches to the gateway, forwarding the visitor's cookies. Once they resolve, React renders to HTML with styled-components' collected CSS, and the Redux state is serialised into `window.__INITIAL_STATE__`. The browser hydrates from that cache, so it doesn't refetch.
2. **Discovery.** The catalog service assembles pages as widget lists. The home page gets hero banners, recommended movies, live-event category tiles, premieres and event rails. Listings get a grid plus filters, and Redis caches each result for 60 seconds. Posters are SVGs rendered on request, with the rating, date or release bar baked in, and the gateway caches them.
3. **Showtimes.** Shows live in Postgres. Live availability = layout seats − committed seats (`booking_seats`) − unexpired Redis locks (`ZRANGEBYSCORE locks:{session} (now +inf`). Bookings close 15 minutes after the show starts.
4. **Seat locking.** `POST /transactions` runs one Lua script that sweeps expired locks, fails if any requested seat is held by another transaction, and otherwise `ZADD`s every seat with its expiry. The transaction itself is a Redis JSON blob whose TTL matches the hold.
5. **Payment & commit.** `pay` re-verifies ownership of every lock and calls the simulated gateway. In one Postgres transaction it inserts the booking and its `booking_seats` rows; a unique-key violation aborts the booking. It then releases the locks and publishes `booking.confirmed`.
6. **Notifications.** `notify` is a Kafka consumer group. It sends e-ticket, cancellation and OTP e-mails over SMTP, and logs simulated SMS.

## License

MIT
