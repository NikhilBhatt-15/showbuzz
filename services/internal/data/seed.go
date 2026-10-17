package data

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

//go:embed schema.sql
var schemaSQL string

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Slug turns a title into a URL slug ("Kaal Bhairav" -> "kaal-bhairav").
func Slug(s string) string {
	return strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func rngFor(parts ...string) *rand.Rand {
	h := fnv.New64a()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	s := h.Sum64()
	return rand.New(rand.NewPCG(s, s^0x9e3779b97f4a7c15))
}

// Migrate applies the (idempotent) schema.
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, schemaSQL)
	return err
}

// SeedCatalog upserts regions, movies, events, venues and screens. Release
// dates and event dates are anchored to today so the demo always looks live.
func SeedCatalog(ctx context.Context, db *pgxpool.Pool, today time.Time) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, r := range Regions {
		if _, err := tx.Exec(ctx, `
			INSERT INTO regions (code, name, slug, state_name, state_code, is_popular, sort, lat, lng)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name, slug=EXCLUDED.slug, is_popular=EXCLUDED.is_popular, sort=EXCLUDED.sort`,
			r.Code, r.Name, r.Slug, r.StateName, r.StateCode, r.Popular, i, r.Lat, r.Lng); err != nil {
			return fmt.Errorf("region %s: %w", r.Code, err)
		}
	}

	for _, m := range Movies {
		cast, _ := json.Marshal(m.Cast)
		crew, _ := json.Marshal(m.Crew)
		release := today.AddDate(0, 0, m.ReleaseDays).Format("2006-01-02")
		if _, err := tx.Exec(ctx, `
			INSERT INTO movies (code, slug, title, tagline, languages, formats, genres, certificate, duration_min,
				release_date, synopsis, cast_members, crew, rating, votes, likes, promoted, hue, popularity)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
			ON CONFLICT (code) DO UPDATE SET slug=EXCLUDED.slug, title=EXCLUDED.title, tagline=EXCLUDED.tagline,
				languages=EXCLUDED.languages, formats=EXCLUDED.formats, genres=EXCLUDED.genres,
				certificate=EXCLUDED.certificate, duration_min=EXCLUDED.duration_min, release_date=EXCLUDED.release_date,
				synopsis=EXCLUDED.synopsis, cast_members=EXCLUDED.cast_members, crew=EXCLUDED.crew,
				promoted=EXCLUDED.promoted, hue=EXCLUDED.hue, popularity=EXCLUDED.popularity`,
			m.Code, Slug(m.Title), m.Title, m.Tagline, m.Languages, m.Formats, m.Genres, m.Certificate, m.DurationMin,
			release, m.Synopsis, cast, crew, m.Rating, m.Votes, m.Likes, m.Promoted, m.Hue, m.Popularity); err != nil {
			return fmt.Errorf("movie %s: %w", m.Code, err)
		}
	}

	for _, e := range Events {
		tiers, _ := json.Marshal(e.Tiers)
		starts := today.AddDate(0, 0, e.DayOffset).Add(time.Duration(e.Hour)*time.Hour + time.Duration(e.Min)*time.Minute)
		var region any
		if e.Region != "" {
			region = e.Region
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (code, slug, title, category, genre, languages, age_limit, duration_min, synopsis,
				region_code, venue_name, starts_at, tiers, promoted, hue, popularity)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			ON CONFLICT (code) DO UPDATE SET title=EXCLUDED.title, slug=EXCLUDED.slug, starts_at=EXCLUDED.starts_at,
				tiers=EXCLUDED.tiers, venue_name=EXCLUDED.venue_name, synopsis=EXCLUDED.synopsis,
				region_code=EXCLUDED.region_code, promoted=EXCLUDED.promoted`,
			e.Code, Slug(e.Title), e.Title, e.Category, e.Genre, e.Languages, e.AgeLimit, e.DurationMin, e.Synopsis,
			region, e.VenueName, starts, tiers, e.Promoted, e.Hue, e.Popularity); err != nil {
			return fmt.Errorf("event %s: %w", e.Code, err)
		}
	}

	malls := []string{"Skyline Mall", "Central Plaza", "City Centre", "Metro Square", "Orbit Mall", "Riverside Mall", "Grand Arcade", "Sunrise Mall"}
	premium := map[string]string{"CSTR": "IMAX", "MVDM": "DOLBY CINEMA", "GLXY": "4DX", "RGLS": "IMAX", "SLVR": ""}
	for _, r := range Regions {
		n := 3
		if r.Popular {
			n = 5
		}
		rng := rngFor("venues", r.Code)
		locs := Localities[r.Code]
		for i := 0; i < n; i++ {
			chain := Chains[(i+int(rng.IntN(5)))%len(Chains)]
			loc := locs[i%len(locs)]
			mall := malls[rng.IntN(len(malls))]
			code := chain.Code[:2] + r.Code[:2] + strconv.Itoa(i+1)
			name := fmt.Sprintf("%s: %s, %s", chain.Name, mall, loc)
			lat := r.Lat + (rng.Float64()-0.5)*0.12
			lng := r.Lng + (rng.Float64()-0.5)*0.12
			if _, err := tx.Exec(ctx, `
				INSERT INTO venues (code, name, slug, region_code, locality, address, company_code, company_name, lat, lng,
					cancellation, fnb, mticket, popularity)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
				ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name, slug=EXCLUDED.slug, locality=EXCLUDED.locality,
					address=EXCLUDED.address, company_code=EXCLUDED.company_code, company_name=EXCLUDED.company_name`,
				code, name, Slug(name), r.Code, loc,
				fmt.Sprintf("%d%s Floor, %s, %s, %s, %s", 2+rng.IntN(3), ordinal(2+rng.IntN(3)), mall, loc, r.Name, r.StateName),
				chain.Code, chain.Name, lat, lng, rng.IntN(3) != 0, chain.Code != "SLVR", true, 50+rng.IntN(50)); err != nil {
				return fmt.Errorf("venue %s: %w", code, err)
			}
			screens := []struct{ name, tpl, fmt string }{{"Screen 1", "classic", ""}, {"Screen 2", "compact", ""}}
			if r.Popular {
				pf := premium[chain.Code]
				label := "Screen 3"
				if pf != "" {
					label = "Audi 3 - " + pf
				}
				screens = append(screens, struct{ name, tpl, fmt string }{label, "large", pf})
			}
			for _, s := range screens {
				if _, err := tx.Exec(ctx, `
					INSERT INTO screens (venue_code, name, layout_template, premium_format) VALUES ($1,$2,$3,$4)
					ON CONFLICT (venue_code, name) DO NOTHING`, code, s.name, s.tpl, s.fmt); err != nil {
					return fmt.Errorf("screen %s/%s: %w", code, s.name, err)
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	}
	return "th"
}

type screenRow struct {
	id                          int
	venue, region, tpl, premium string
	chain                       string
}

type movieRow struct {
	code       string
	languages  []string
	formats    []string
	duration   int
	popularity int
	release    string // yyyymmdd
}

var basePrices = map[string]map[string]int{
	"classic": {"RC": 550, "PR": 230, "EX": 200, "NM": 180},
	"compact": {"GD": 240, "SL": 190},
	"large":   {"PL": 650, "GD": 340, "SL": 280},
}

var classicAttributes = []string{"", "", "2K LASER DOLBY 7.1", "ATMOS", "4K DOLBY ATMOS"}

// EnsureShows generates showtimes for every screen for each of the next
// `days` days starting at `from`, skipping screen-days that already have
// shows. New shows get a realistic share of pre-sold seats.
func EnsureShows(ctx context.Context, db *pgxpool.Pool, from time.Time, days int) (int, error) {
	screens, err := loadScreens(ctx, db)
	if err != nil {
		return 0, err
	}
	movies, err := loadMovies(ctx, db)
	if err != nil {
		return 0, err
	}
	tierOf := map[string]float64{}
	for _, c := range Chains {
		tierOf[c.Code] = c.Tier
	}

	created := 0
	for d := 0; d < days; d++ {
		date := from.AddDate(0, 0, d)
		dateCode := date.Format("20060102")

		existing := map[int]bool{}
		rows, err := db.Query(ctx, `SELECT DISTINCT screen_id FROM shows WHERE date_code = $1`, dateCode)
		if err != nil {
			return created, err
		}
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return created, err
			}
			existing[id] = true
		}
		rows.Close()

		var playing []movieRow
		for _, m := range movies {
			if m.release <= dateCode {
				playing = append(playing, m)
			}
		}
		if len(playing) == 0 {
			continue
		}

		for _, sc := range screens {
			if existing[sc.id] {
				continue
			}
			n, err := scheduleScreen(ctx, db, sc, playing, date, dateCode, d, tierOf[sc.chain])
			if err != nil {
				return created, fmt.Errorf("schedule screen %d on %s: %w", sc.id, dateCode, err)
			}
			created += n
		}
	}
	return created, nil
}

func loadScreens(ctx context.Context, db *pgxpool.Pool) ([]screenRow, error) {
	rows, err := db.Query(ctx, `
		SELECT s.id, s.venue_code, v.region_code, s.layout_template, s.premium_format, v.company_code
		FROM screens s JOIN venues v ON v.code = s.venue_code ORDER BY s.id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (screenRow, error) {
		var s screenRow
		err := r.Scan(&s.id, &s.venue, &s.region, &s.tpl, &s.premium, &s.chain)
		return s, err
	})
}

func loadMovies(ctx context.Context, db *pgxpool.Pool) ([]movieRow, error) {
	rows, err := db.Query(ctx, `SELECT code, languages, formats, duration_min, popularity, release_date FROM movies`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (movieRow, error) {
		var m movieRow
		var release time.Time
		err := r.Scan(&m.code, &m.languages, &m.formats, &m.duration, &m.popularity, &release)
		m.release = release.Format("20060102")
		return m, err
	})
}

// moviesForScreen ranks the playing movies for one screen: regional language
// fit and popularity first, with a venue-specific shuffle so cinemas differ.
func moviesForScreen(sc screenRow, playing []movieRow, rng *rand.Rand) []movieRow {
	regionLangs := RegionLanguages[sc.region]
	type scored struct {
		m     movieRow
		score float64
	}
	var list []scored
	for _, m := range playing {
		if sc.premium != "" && !slices.ContainsFunc(m.formats, func(f string) bool { return strings.HasPrefix(f, sc.premium) }) {
			continue
		}
		s := float64(m.popularity) + rng.Float64()*45
		for _, l := range m.languages {
			if slices.Contains(regionLangs, l) {
				s += 40
				break
			}
		}
		list = append(list, scored{m, s})
	}
	slices.SortFunc(list, func(a, b scored) int {
		switch {
		case a.score > b.score:
			return -1
		case a.score < b.score:
			return 1
		}
		return strings.Compare(a.m.code, b.m.code)
	})
	n := 3
	if sc.tpl == "large" {
		n = 2
	}
	out := make([]movieRow, 0, n)
	for i := 0; i < len(list) && len(out) < n; i++ {
		out = append(out, list[i].m)
	}
	return out
}

func pickLanguage(m movieRow, region string, rng *rand.Rand) string {
	var fits []string
	for _, l := range m.languages {
		if slices.Contains(RegionLanguages[region], l) {
			fits = append(fits, l)
		}
	}
	if len(fits) == 0 {
		return m.languages[0]
	}
	// Prefer the original language, but screen dubbed versions sometimes.
	if slices.Contains(fits, m.languages[0]) && rng.IntN(3) != 0 {
		return m.languages[0]
	}
	return fits[rng.IntN(len(fits))]
}

func pickFormat(m movieRow, sc screenRow, rng *rand.Rand) (format, attribute string) {
	if sc.premium != "" {
		for _, f := range m.formats {
			if strings.HasPrefix(f, sc.premium) {
				attr := map[string]string{"IMAX": "IMAX WITH LASER", "DOLBY CINEMA": "DOLBY VISION + ATMOS", "4DX": "4DX MOTION SEATS"}[sc.premium]
				return f, attr
			}
		}
	}
	if slices.Contains(m.formats, "3D") && rng.IntN(2) == 0 {
		return "3D", ""
	}
	if sc.tpl == "classic" {
		return "2D", classicAttributes[rng.IntN(len(classicAttributes))]
	}
	return "2D", ""
}

func scheduleScreen(ctx context.Context, db *pgxpool.Pool, sc screenRow, playing []movieRow, date time.Time, dateCode string, daysAhead int, tier float64) (int, error) {
	rng := rngFor("schedule", strconv.Itoa(sc.id), dateCode)
	lineup := moviesForScreen(sc, playing, rngFor("lineup", sc.venue, strconv.Itoa(sc.id), dateCode))
	if len(lineup) == 0 {
		return 0, nil
	}
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, platform.IST)
	start := day.Add(9*time.Hour + time.Duration(rng.IntN(10)*5)*time.Minute)
	last := day.Add(23*time.Hour + 15*time.Minute)
	weekend := date.Weekday() == time.Friday || date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
	total := SeatCount(sc.tpl)

	layout, _ := GetLayout(sc.tpl)
	var allSeats []string
	for _, a := range layout.Areas {
		for _, r := range a.Rows {
			for _, s := range r.Seats {
				allSeats = append(allSeats, s.ID)
			}
		}
	}

	created := 0
	for i := 0; !start.After(last); i++ {
		m := lineup[i%len(lineup)]
		format, attr := pickFormat(m, sc, rng)
		lang := pickLanguage(m, sc.region, rng)

		prices := map[string]int{}
		for area, base := range basePrices[sc.tpl] {
			p := float64(base) * tier
			switch {
			case strings.HasPrefix(format, "IMAX"), strings.HasPrefix(format, "DOLBY"), format == "4DX":
				p += 180
			case format == "3D":
				p += 40
			}
			if weekend {
				p *= 1.15
			}
			if start.Hour() >= 18 {
				p += 20
			}
			prices[area] = int(math.Round(p/10) * 10)
		}
		pj, _ := json.Marshal(prices)

		var sessionID int64
		err := db.QueryRow(ctx, `
			INSERT INTO shows (movie_code, venue_code, screen_id, starts_at, date_code, language, format, attribute, prices, total_seats)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (screen_id, starts_at) DO NOTHING RETURNING session_id`,
			m.code, sc.venue, sc.id, start, dateCode, lang, format, attr, pj, total).Scan(&sessionID)
		if err == nil {
			created++
			if err := presell(ctx, db, sessionID, allSeats, presoldShare(rng, daysAhead, m.popularity, start.Hour())); err != nil {
				return created, err
			}
		} else if err != pgx.ErrNoRows {
			return created, err
		}

		next := start.Add(time.Duration(m.duration+25) * time.Minute)
		if r := next.Minute() % 5; r != 0 {
			next = next.Add(time.Duration(5-r) * time.Minute)
		}
		start = next
	}
	return created, nil
}

// presoldShare models demand: closer dates, popular titles and evening shows
// sell more; a few near-term shows are fully sold out.
func presoldShare(rng *rand.Rand, daysAhead, popularity, hour int) float64 {
	if daysAhead <= 1 && rng.IntN(25) == 0 {
		return 1
	}
	f := 0.45 - 0.07*float64(daysAhead) + float64(popularity)/100*0.25 + (rng.Float64()-0.5)*0.3
	if hour >= 18 && hour <= 22 {
		f += 0.15
	}
	return math.Max(0.02, math.Min(0.97, f))
}

func presell(ctx context.Context, db *pgxpool.Pool, sessionID int64, seats []string, share float64) error {
	n := int(float64(len(seats)) * share)
	if n == 0 {
		return nil
	}
	rng := rngFor("presell", strconv.FormatInt(sessionID, 10))
	perm := rng.Perm(len(seats))
	rows := make([][]any, 0, n)
	for _, idx := range perm[:n] {
		rows = append(rows, []any{sessionID, seats[idx], "SYSTEM"})
	}
	_, err := db.CopyFrom(ctx, pgx.Identifier{"booking_seats"}, []string{"session_id", "seat_id", "booking_id"}, pgx.CopyFromRows(rows))
	return err
}

// Bootstrap migrates, seeds and generates a week of shows.
func Bootstrap(ctx context.Context, db *pgxpool.Pool, today time.Time) error {
	if err := Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	var seeded bool
	if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM movies)`).Scan(&seeded); err != nil {
		return err
	}
	if !seeded {
		if err := SeedCatalog(ctx, db, today); err != nil {
			return fmt.Errorf("seed catalog: %w", err)
		}
		slog.Info("catalog seeded")
	}
	n, err := EnsureShows(ctx, db, today, 7)
	if err != nil {
		return fmt.Errorf("ensure shows: %w", err)
	}
	slog.Info("shows ensured", "created", n)
	return nil
}
