CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS regions (
    code        text PRIMARY KEY,
    name        text NOT NULL,
    slug        text NOT NULL UNIQUE,
    state_name  text NOT NULL,
    state_code  text NOT NULL,
    is_popular  boolean NOT NULL DEFAULT false,
    sort        int NOT NULL DEFAULT 0,
    lat         double precision NOT NULL,
    lng         double precision NOT NULL
);

CREATE TABLE IF NOT EXISTS movies (
    code          text PRIMARY KEY,
    slug          text NOT NULL,
    title         text NOT NULL,
    tagline       text NOT NULL DEFAULT '',
    languages     text[] NOT NULL,
    formats       text[] NOT NULL,
    genres        text[] NOT NULL,
    certificate   text NOT NULL,
    duration_min  int NOT NULL,
    release_date  date NOT NULL,
    synopsis      text NOT NULL,
    cast_members  jsonb NOT NULL DEFAULT '[]',
    crew          jsonb NOT NULL DEFAULT '[]',
    rating        numeric(3,1) NOT NULL DEFAULT 0,
    votes         int NOT NULL DEFAULT 0,
    likes         int NOT NULL DEFAULT 0,
    promoted      boolean NOT NULL DEFAULT false,
    hue           int NOT NULL DEFAULT 0,
    popularity    int NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS movies_title_trgm ON movies USING gin (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS events (
    code          text PRIMARY KEY,
    slug          text NOT NULL,
    title         text NOT NULL,
    category      text NOT NULL,
    genre         text NOT NULL,
    languages     text[] NOT NULL,
    age_limit     text NOT NULL,
    duration_min  int NOT NULL,
    synopsis      text NOT NULL,
    region_code   text REFERENCES regions(code),
    venue_name    text NOT NULL,
    starts_at     timestamptz NOT NULL,
    tiers         jsonb NOT NULL,
    promoted      boolean NOT NULL DEFAULT false,
    hue           int NOT NULL DEFAULT 0,
    popularity    int NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS events_title_trgm ON events USING gin (title gin_trgm_ops);

CREATE TABLE IF NOT EXISTS venues (
    code          text PRIMARY KEY,
    name          text NOT NULL,
    slug          text NOT NULL,
    region_code   text NOT NULL REFERENCES regions(code),
    locality      text NOT NULL,
    address       text NOT NULL,
    company_code  text NOT NULL,
    company_name  text NOT NULL,
    lat           double precision NOT NULL,
    lng           double precision NOT NULL,
    cancellation  boolean NOT NULL,
    fnb           boolean NOT NULL,
    mticket       boolean NOT NULL,
    popularity    int NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS venues_region ON venues(region_code);
CREATE INDEX IF NOT EXISTS venues_name_trgm ON venues USING gin (name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS screens (
    id               serial PRIMARY KEY,
    venue_code       text NOT NULL REFERENCES venues(code),
    name             text NOT NULL,
    layout_template  text NOT NULL,
    premium_format   text NOT NULL DEFAULT '',
    UNIQUE (venue_code, name)
);

CREATE TABLE IF NOT EXISTS shows (
    session_id   bigserial PRIMARY KEY,
    movie_code   text NOT NULL REFERENCES movies(code),
    venue_code   text NOT NULL REFERENCES venues(code),
    screen_id    int NOT NULL REFERENCES screens(id),
    starts_at    timestamptz NOT NULL,
    date_code    text NOT NULL,
    language     text NOT NULL,
    format       text NOT NULL,
    attribute    text NOT NULL DEFAULT '',
    prices       jsonb NOT NULL,
    total_seats  int NOT NULL,
    UNIQUE (screen_id, starts_at)
);
CREATE INDEX IF NOT EXISTS shows_movie_date ON shows(movie_code, date_code);
CREATE INDEX IF NOT EXISTS shows_venue_date ON shows(venue_code, date_code);

CREATE TABLE IF NOT EXISTS users (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier  text NOT NULL UNIQUE,
    email       text NOT NULL DEFAULT '',
    phone       text NOT NULL DEFAULT '',
    name        text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bookings (
    id              text PRIMARY KEY,
    user_id         uuid REFERENCES users(id),
    kind            text NOT NULL,              -- MOVIE | EVENT
    session_id      bigint REFERENCES shows(session_id),
    event_code      text REFERENCES events(code),
    tier            text,
    seats           text[] NOT NULL DEFAULT '{}',
    qty             int NOT NULL,
    ticket_amount   numeric(10,2) NOT NULL,
    conv_fee        numeric(10,2) NOT NULL,
    gst             numeric(10,2) NOT NULL,
    fnb_amount      numeric(10,2) NOT NULL DEFAULT 0,
    donation        numeric(10,2) NOT NULL DEFAULT 0,
    total           numeric(10,2) NOT NULL,
    payment_method  text NOT NULL,
    payment_ref     text NOT NULL,
    status          text NOT NULL,              -- CONFIRMED | CANCELLED
    contact_email   text NOT NULL DEFAULT '',
    contact_phone   text NOT NULL DEFAULT '',
    snapshot        jsonb NOT NULL,
    fnb             jsonb NOT NULL DEFAULT '[]',
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS bookings_user ON bookings(user_id, created_at DESC);

-- The primary key is the last line of defence against double booking: even if
-- a Redis seat lock were lost, two bookings can never commit the same seat.
CREATE TABLE IF NOT EXISTS booking_seats (
    session_id  bigint NOT NULL REFERENCES shows(session_id) ON DELETE CASCADE,
    seat_id     text NOT NULL,
    booking_id  text NOT NULL,
    PRIMARY KEY (session_id, seat_id)
);

CREATE TABLE IF NOT EXISTS event_sales (
    booking_id  text NOT NULL,
    event_code  text NOT NULL REFERENCES events(code),
    tier        text NOT NULL,
    qty         int NOT NULL
);
CREATE INDEX IF NOT EXISTS event_sales_event ON event_sales(event_code, tier);
