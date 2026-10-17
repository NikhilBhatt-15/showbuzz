// Package catalog serves discovery (server-driven listing pages), movie and
// event details, regions, search, clickstream ingestion and artwork.
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/data"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

type Server struct {
	db  *pgxpool.Pool
	rdb *redis.Client
	pub *platform.Producer
}

func New(db *pgxpool.Pool, rdb *redis.Client, pub *platform.Producer) *Server {
	return &Server{db: db, rdb: rdb, pub: pub}
}

func (s *Server) Routes(r chi.Router) {
	r.Get("/api/regions", s.regions)
	r.Get("/api/explore/v1/discover/{page}", s.discover)
	r.Get("/api/movies/v1/synopsis/init/static", s.synopsisStatic)
	r.Get("/api/movies/v1/synopsis/init/dynamic", s.synopsisDynamic)
	r.Get("/api/events/v1/{code}", s.eventDetail)
	r.Get("/api/search/v1", s.search)
	r.Post("/api/clickstream/v1/publish", s.clickstream)

	r.Get("/api/img/poster/{code}", s.poster)
	r.Get("/api/img/backdrop/{code}", s.backdrop)
	r.Get("/api/img/banner/{id}", s.bannerImage)
	r.Get("/api/img/tile/{slug}", s.tileImage)
	r.Get("/api/img/avatar", s.avatar)
	r.Get("/api/img/city/{slug}", s.cityIcon)
	r.Get("/api/img/logo/{code}", s.chainLogo)
	r.Get("/api/img/strip/stream", s.stripImage)
	r.Get("/api/img/fnb/{code}", s.fnbImage)
}

// ---- regions ---------------------------------------------------------------

type Region struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	StateName string  `json:"stateName"`
	StateCode string  `json:"stateCode"`
	Popular   bool    `json:"popular"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Icon      string  `json:"icon"`
}

func (s *Server) loadRegions(ctx context.Context) ([]Region, error) {
	rows, err := s.db.Query(ctx, `SELECT code, name, slug, state_name, state_code, is_popular, lat, lng FROM regions ORDER BY sort`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Region, error) {
		var g Region
		err := r.Scan(&g.Code, &g.Name, &g.Slug, &g.StateName, &g.StateCode, &g.Popular, &g.Lat, &g.Lng)
		g.Icon = "/api/img/city/" + g.Slug + ".svg"
		return g, err
	})
}

// resolveRegion accepts a region code ("MUMBAI") or URL slug ("mumbai").
func (s *Server) resolveRegion(ctx context.Context, key string) (Region, error) {
	regions, err := s.loadRegions(ctx)
	if err != nil {
		return Region{}, err
	}
	for _, g := range regions {
		if strings.EqualFold(g.Code, key) || g.Slug == strings.ToLower(key) {
			return g, nil
		}
	}
	return regions[0], nil
}

func (s *Server) regions(w http.ResponseWriter, r *http.Request) {
	regions, err := s.loadRegions(r.Context())
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	var popular, others []Region
	for _, g := range regions {
		if g.Popular {
			popular = append(popular, g)
		} else {
			others = append(others, g)
		}
	}
	platform.Cache(w, 300)
	platform.JSON(w, 200, map[string]any{"popular": popular, "others": others})
}

// ---- movie details (static vs dynamic split, as on BookMyShow) -------------

type movie struct {
	Code        string          `json:"code"`
	Slug        string          `json:"slug"`
	Title       string          `json:"title"`
	Tagline     string          `json:"tagline"`
	Languages   []string        `json:"languages"`
	Formats     []string        `json:"formats"`
	Genres      []string        `json:"genres"`
	Certificate string          `json:"certificate"`
	DurationMin int             `json:"durationMin"`
	ReleaseDate time.Time       `json:"releaseDate"`
	Synopsis    string          `json:"synopsis"`
	Cast        json.RawMessage `json:"-"`
	Crew        json.RawMessage `json:"-"`
	Rating      float64         `json:"rating"`
	Votes       int             `json:"votes"`
	Likes       int             `json:"likes"`
	Promoted    bool            `json:"promoted"`
	Popularity  int             `json:"popularity"`
}

const movieCols = `code, slug, title, tagline, languages, formats, genres, certificate, duration_min, release_date,
	synopsis, cast_members, crew, rating::float8, votes, likes, promoted, popularity`

func scanMovie(r pgx.CollectableRow) (movie, error) {
	var m movie
	err := r.Scan(&m.Code, &m.Slug, &m.Title, &m.Tagline, &m.Languages, &m.Formats, &m.Genres, &m.Certificate,
		&m.DurationMin, &m.ReleaseDate, &m.Synopsis, &m.Cast, &m.Crew, &m.Rating, &m.Votes, &m.Likes, &m.Promoted, &m.Popularity)
	return m, err
}

func (s *Server) getMovie(ctx context.Context, code string) (movie, error) {
	rows, err := s.db.Query(ctx, `SELECT `+movieCols+` FROM movies WHERE code = $1`, code)
	if err != nil {
		return movie{}, err
	}
	return pgx.CollectExactlyOneRow(rows, scanMovie)
}

func (m movie) comingSoon() bool { return m.ReleaseDate.After(platform.Today()) }

type person struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Avatar string `json:"avatar"`
}

func people(raw json.RawMessage) []person {
	var ps []person
	_ = json.Unmarshal(raw, &ps)
	for i := range ps {
		ps[i].Avatar = "/api/img/avatar?name=" + strings.ReplaceAll(ps[i].Name, " ", "+")
	}
	return ps
}

func (s *Server) synopsisStatic(w http.ResponseWriter, r *http.Request) {
	m, err := s.getMovie(r.Context(), r.URL.Query().Get("eventCode"))
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Err(w, 404, "NOT_FOUND", "Movie not found")
		return
	} else if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.Cache(w, 600)
	platform.JSON(w, 200, map[string]any{
		"movie":       m,
		"cast":        people(m.Cast),
		"crew":        people(m.Crew),
		"posterUrl":   "/api/img/poster/" + m.Code + ".svg",
		"backdropUrl": "/api/img/backdrop/" + m.Code + ".svg",
		"comingSoon":  m.comingSoon(),
	})
}

var reviewTexts = []string{
	"Absolutely gripping from start to finish. The second half is a rollercoaster and the background score gives goosebumps.",
	"Brilliant performances all round. The screenplay could have been tighter in places but the climax more than makes up for it.",
	"Watched it with family and everyone loved it. Worth every rupee — go for the big screen experience!",
	"One of the best films this year. Direction, cinematography and music are all top notch.",
	"Good one-time watch. The lead actor carries the film on his shoulders.",
	"Paisa vasool entertainer! The interval block is the highlight.",
}
var reviewTags = []string{"#Blockbuster", "#GreatActing", "#MustWatch", "#SuperDirection", "#AwesomeStory", "#Wellmade", "#GoodMusic"}
var reviewers = []string{"Aarav", "Diya", "Rohit", "Sneha", "Vikram", "Ananya", "Kunal", "Pooja", "Siddharth", "Meera"}

func (s *Server) synopsisDynamic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("eventCode")
	m, err := s.getMovie(ctx, code)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Movie not found")
		return
	}
	region, err := s.resolveRegion(ctx, r.URL.Query().Get("regionCode"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	var hasShows bool
	_ = s.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM shows s JOIN venues v ON v.code = s.venue_code
		WHERE s.movie_code = $1 AND v.region_code = $2 AND s.date_code >= $3)`, code, region.Code, platform.DateCode(time.Now())).Scan(&hasShows)

	rng := seeded(code)
	var reviews []map[string]any
	if m.Rating > 0 {
		for i := 0; i < 4; i++ {
			score := 7 + rng.IntN(4)
			tags := []string{reviewTags[rng.IntN(len(reviewTags))], reviewTags[rng.IntN(len(reviewTags))]}
			slices.Sort(tags)
			reviews = append(reviews, map[string]any{
				"user":   reviewers[rng.IntN(len(reviewers))],
				"rating": score,
				"tags":   slices.Compact(tags),
				"text":   reviewTexts[rng.IntN(len(reviewTexts))],
				"ago":    fmt.Sprintf("%d Days ago", 1+rng.IntN(9)),
				"likes":  rng.IntN(900),
			})
		}
	}

	similar, err := s.queryMovieCards(ctx, region, `AND m.code <> $3 AND m.genres && $4`, cardOpts{overlay: "rating"}, code, m.Genres)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if len(similar) > 10 {
		similar = similar[:10]
	}

	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{
		"rating":     m.Rating,
		"votes":      m.Votes,
		"likes":      m.Likes,
		"hasShows":   hasShows,
		"comingSoon": m.comingSoon(),
		"reviews":    reviews,
		"offers":     offers(),
		"similar":    similar,
		"region":     region,
	})
}

func offers() []map[string]string {
	return []map[string]string{
		{"title": "SB Pay Wallet Offer", "text": "Flat ₹100 off on 2 tickets. Use code SBPAY100", "icon": "wallet"},
		{"title": "Platinum Credit Card Offer", "text": "Buy 1 get 1 free on movie tickets every Friday", "icon": "card"},
		{"title": "UPI Cashback", "text": "Get up to ₹75 cashback on your first UPI payment", "icon": "upi"},
	}
}

// ---- events ----------------------------------------------------------------

type event struct {
	Code        string      `json:"code"`
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Category    string      `json:"category"`
	Genre       string      `json:"genre"`
	Languages   []string    `json:"languages"`
	AgeLimit    string      `json:"ageLimit"`
	DurationMin int         `json:"durationMin"`
	Synopsis    string      `json:"synopsis"`
	RegionCode  *string     `json:"regionCode"`
	VenueName   string      `json:"venueName"`
	StartsAt    time.Time   `json:"startsAt"`
	Tiers       []data.Tier `json:"tiers"`
	Promoted    bool        `json:"promoted"`
	Popularity  int         `json:"popularity"`
}

const eventCols = `code, slug, title, category, genre, languages, age_limit, duration_min, synopsis, region_code,
	venue_name, starts_at, tiers, promoted, popularity`

func scanEvent(r pgx.CollectableRow) (event, error) {
	var e event
	err := r.Scan(&e.Code, &e.Slug, &e.Title, &e.Category, &e.Genre, &e.Languages, &e.AgeLimit, &e.DurationMin,
		&e.Synopsis, &e.RegionCode, &e.VenueName, &e.StartsAt, &e.Tiers, &e.Promoted, &e.Popularity)
	return e, err
}

func (e event) priceFrom() int {
	min := 0
	for i, t := range e.Tiers {
		if i == 0 || t.Price < min {
			min = t.Price
		}
	}
	return min
}

func (s *Server) eventDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.Query(ctx, `SELECT `+eventCols+` FROM events WHERE code = $1`, chi.URLParam(r, "code"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	e, err := pgx.CollectExactlyOneRow(rows, scanEvent)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Event not found")
		return
	}
	region, _ := s.resolveRegion(ctx, r.URL.Query().Get("regionCode"))
	city := region.Name
	if e.RegionCode != nil {
		if g, err := s.resolveRegion(ctx, *e.RegionCode); err == nil {
			city = g.Name
		}
	}
	platform.Cache(w, 60)
	platform.JSON(w, 200, map[string]any{
		"event":       e,
		"city":        city,
		"venue":       e.VenueName + ": " + city,
		"priceFrom":   e.priceFrom(),
		"posterUrl":   "/api/img/poster/" + e.Code + ".svg",
		"backdropUrl": "/api/img/backdrop/" + e.Code + ".svg",
		"thingsToKnow": []string{
			"Entry allowed for " + strings.ToLower(e.AgeLimit),
			"Please carry a valid ID proof along with you.",
			"No refunds on purchased ticket are possible, even in case of any rescheduling.",
			"Security procedures, including frisking remain the right of the management.",
		},
	})
}

// ---- search ----------------------------------------------------------------

type searchResult struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
	Image    string `json:"image,omitempty"`
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	region, err := s.resolveRegion(ctx, r.URL.Query().Get("regionCode"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	results := []searchResult{}
	if len(q) == 0 {
		platform.JSON(w, 200, map[string]any{"query": q, "results": results})
		return
	}
	like := "%" + q + "%"

	rows, err := s.db.Query(ctx, `
		SELECT code, slug, title, languages, genres FROM movies
		WHERE title ILIKE $1 OR similarity(title, $2) > 0.2
		ORDER BY similarity(title, $2) DESC, popularity DESC LIMIT 6`, like, q)
	if err == nil {
		for rows.Next() {
			var code, slug, title string
			var langs, genres []string
			if rows.Scan(&code, &slug, &title, &langs, &genres) == nil {
				results = append(results, searchResult{"Movies", title, strings.Join(langs, ", "),
					"/movies/" + region.Slug + "/" + slug + "/" + code, "/api/img/poster/" + code + ".svg"})
			}
		}
		rows.Close()
	}
	rows, err = s.db.Query(ctx, `
		SELECT code, slug, title, genre FROM events
		WHERE (region_code IS NULL OR region_code = $3) AND (title ILIKE $1 OR similarity(title, $2) > 0.2)
		ORDER BY similarity(title, $2) DESC LIMIT 5`, like, q, region.Code)
	if err == nil {
		for rows.Next() {
			var code, slug, title, genre string
			if rows.Scan(&code, &slug, &title, &genre) == nil {
				results = append(results, searchResult{"Events", title, genre, "/events/" + slug + "/" + code, "/api/img/poster/" + code + ".svg"})
			}
		}
		rows.Close()
	}
	rows, err = s.db.Query(ctx, `
		SELECT code, slug, name, locality FROM venues
		WHERE region_code = $2 AND name ILIKE $1 ORDER BY popularity DESC LIMIT 5`, like, region.Code)
	if err == nil {
		for rows.Next() {
			var code, slug, name, loc string
			if rows.Scan(&code, &slug, &name, &loc) == nil {
				results = append(results, searchResult{"Cinemas", name, loc,
					"/cinemas/" + region.Slug + "/" + slug + "/buytickets/" + code + "/" + platform.DateCode(time.Now()), ""})
			}
		}
		rows.Close()
	}
	platform.Cache(w, 30)
	platform.JSON(w, 200, map[string]any{"query": q, "results": results})
}

// ---- clickstream -----------------------------------------------------------

func (s *Server) clickstream(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Events []map[string]any `json:"events"`
	}
	if err := platform.Decode(r, &body); err != nil || len(body.Events) > 50 {
		platform.Err(w, 400, "BAD_REQUEST", "expected {events:[...]} with at most 50 events")
		return
	}
	uid := ""
	if u := platform.User(r.Context()); u != nil {
		uid = u.UserID
	}
	for _, e := range body.Events {
		e["received_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		e["user_id"] = uid
		e["user_agent"] = r.UserAgent()
		s.pub.Publish(r.Context(), platform.TopicClickstream, uid, e)
	}
	w.WriteHeader(http.StatusAccepted)
}
