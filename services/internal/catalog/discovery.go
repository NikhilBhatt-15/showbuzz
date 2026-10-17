package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

// The discovery API is server-driven UI, modelled on BookMyShow's
// getDiscoveryData payload: the backend decides which widgets ("listings") a
// page shows, in what order, with which cards and style IDs; the web client
// only knows how to render widget types.

type Image struct {
	URL     string `json:"url"`
	AltText string `json:"altText"`
}

type Text struct {
	Text    string `json:"text"`
	StyleID string `json:"styleId"`
}

type Card struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	CtaURL string `json:"ctaUrl"`
	Image  Image  `json:"image"`
	Tag    string `json:"tag,omitempty"`
	Text   []Text `json:"text"`
}

type Widget struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // banner-carousel | horizontal-carousel | category-carousel | image-strip | coming-soon-strip | grid
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Theme    string `json:"theme,omitempty"`
	CtaText  string `json:"ctaText,omitempty"`
	CtaURL   string `json:"ctaUrl,omitempty"`
	StyleID  string `json:"styleId"`
	Cards    []Card `json:"cards"`
}

type FilterItem struct {
	Code     string `json:"code"`
	Label    string `json:"label"`
	Selected bool   `json:"selected"`
}

type Filter struct {
	QueryParam  string       `json:"queryParam"`
	Title       string       `json:"title"`
	QuickFilter bool         `json:"quickFilter"`
	Items       []FilterItem `json:"items"`
}

type Crumb struct {
	Label string `json:"label"`
	Route string `json:"route"`
}

type Style struct {
	Font      string `json:"font,omitempty"`
	FontColor string `json:"fontColor,omitempty"`
	Padding   string `json:"padding,omitempty"`
}

type Discovery struct {
	PageID      string           `json:"pageId"`
	FilterRoute string           `json:"filterRoute"`
	Region      Region           `json:"region"`
	Header      map[string]any   `json:"header"`
	Meta        map[string]any   `json:"meta"`
	Listings    []Widget         `json:"listings"`
	Filters     []Filter         `json:"filters"`
	Breadcrumbs []Crumb          `json:"breadCrumbs"`
	Styles      map[string]Style `json:"styles"`
	Empty       map[string]any   `json:"empty,omitempty"`
}

var styles = map[string]Style{
	"widget":            {Padding: "32,0,32,0"},
	"widget-dark":       {Padding: "32,0,32,0"},
	"card-title":        {Font: "subtitle-large", FontColor: "#222222"},
	"card-title-dark":   {Font: "subtitle-large", FontColor: "#FFFFFF"},
	"card-muted":        {Font: "body-regular", FontColor: "#666666"},
	"card-muted-dark":   {Font: "body-regular", FontColor: "#CCCCCC"},
	"widget-title":      {Font: "title", FontColor: "#222222"},
	"widget-title-dark": {Font: "title", FontColor: "#FFFFFF"},
	"widget-cta":        {Font: "body-regular", FontColor: "#DC3558"},
}

type cardOpts struct {
	overlay string // rating | release
	dark    bool
	meta    string // genres | certlang | release
}

func movieCard(m movie, region Region, o cardOpts) Card {
	img := "/api/img/poster/" + m.Code + ".svg"
	if o.overlay != "" {
		img += "?overlay=" + o.overlay
	}
	title, muted := "card-title", "card-muted"
	if o.dark {
		title, muted = "card-title-dark", "card-muted-dark"
	}
	var sub string
	switch o.meta {
	case "certlang":
		sub = m.Certificate + " • " + strings.Join(m.Languages, ", ")
	case "release":
		sub = "Releasing on " + m.ReleaseDate.Format("2 Jan, 2006")
	default:
		sub = strings.Join(m.Genres, "/")
	}
	c := Card{ID: m.Code, Kind: "movie", CtaURL: "/movies/" + region.Slug + "/" + m.Slug + "/" + m.Code,
		Image: Image{URL: img, AltText: m.Title}, Text: []Text{{m.Title, title}, {sub, muted}}}
	if m.Promoted && !o.dark {
		c.Tag = "PROMOTED"
	}
	return c
}

func eventCard(e event, region Region, dark bool) Card {
	title, muted := "card-title", "card-muted"
	if dark {
		title, muted = "card-title-dark", "card-muted-dark"
	}
	venue := e.VenueName
	if e.RegionCode == nil {
		venue += ": " + region.Name
	}
	c := Card{ID: e.Code, Kind: "event", CtaURL: "/events/" + e.Slug + "/" + e.Code,
		Image: Image{URL: "/api/img/poster/" + e.Code + ".svg?overlay=date", AltText: e.Title},
		Text:  []Text{{e.Title, title}, {venue, muted}, {e.Genre, muted}}}
	if e.Promoted {
		c.Tag = "PROMOTED"
	}
	return c
}

// queryMovieCards returns movies with shows in region from today on. Extra
// SQL may reference $1 (region code) and $2 (today) plus its own args.
func (s *Server) queryMovieCards(ctx context.Context, region Region, extra string, o cardOpts, args ...any) ([]Card, error) {
	movies, err := s.queryMovies(ctx, region, extra, args...)
	if err != nil {
		return nil, err
	}
	cards := make([]Card, 0, len(movies))
	for _, m := range movies {
		cards = append(cards, movieCard(m, region, o))
	}
	return cards, nil
}

func (s *Server) queryMovies(ctx context.Context, region Region, extra string, args ...any) ([]movie, error) {
	q := `SELECT ` + movieCols + ` FROM movies m WHERE EXISTS (
		SELECT 1 FROM shows s JOIN venues v ON v.code = s.venue_code
		WHERE s.movie_code = m.code AND v.region_code = $1 AND s.date_code >= $2) ` + extra +
		` ORDER BY m.promoted DESC, m.popularity DESC`
	rows, err := s.db.Query(ctx, q, append([]any{region.Code, platform.DateCode(time.Now())}, args...)...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanMovie)
}

func (s *Server) upcomingMovies(ctx context.Context) ([]movie, error) {
	rows, err := s.db.Query(ctx, `SELECT `+movieCols+` FROM movies WHERE release_date > $1 ORDER BY release_date, popularity DESC`,
		platform.Today().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanMovie)
}

func (s *Server) eventsIn(ctx context.Context, region Region, category string) ([]event, error) {
	q := `SELECT ` + eventCols + ` FROM events WHERE (region_code IS NULL OR region_code = $1) AND starts_at > now()`
	args := []any{region.Code}
	if category != "" {
		q += ` AND category = $2`
		args = append(args, category)
	}
	rows, err := s.db.Query(ctx, q+` ORDER BY promoted DESC, popularity DESC`, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, scanEvent)
}

func selected(q url.Values, param string) []string {
	v := q.Get(param)
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

func buildFilter(param, title string, quick bool, values []string, sel []string) Filter {
	f := Filter{QueryParam: param, Title: title, QuickFilter: quick}
	for _, v := range values {
		code := strings.ReplaceAll(strings.ToLower(v), " ", "-")
		f.Items = append(f.Items, FilterItem{Code: code, Label: v, Selected: slices.Contains(sel, code)})
	}
	return f
}

func matchesAny(values []string, sel []string) bool {
	if len(sel) == 0 {
		return true
	}
	for _, v := range values {
		if slices.Contains(sel, strings.ReplaceAll(strings.ToLower(v), " ", "-")) {
			return true
		}
	}
	return false
}

func distinct(lists ...[]string) []string {
	var out []string
	for _, l := range lists {
		for _, v := range l {
			if !slices.Contains(out, v) {
				out = append(out, v)
			}
		}
	}
	return out
}

var categoryPages = map[string]struct{ Category, Label string }{
	"events": {"EVENTS", "Events"}, "plays": {"PLAYS", "Plays"}, "sports": {"SPORTS", "Sports"}, "activities": {"ACTIVITIES", "Activities"},
}

func (s *Server) discover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := chi.URLParam(r, "page")
	q := r.URL.Query()
	region, err := s.resolveRegion(ctx, q.Get("region"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}

	cacheKey := "discover:" + page + ":" + region.Code + ":" + q.Encode()
	if cached, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
		w.Header().Set("X-Discovery-Cache", "HIT")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		platform.Cache(w, 30)
		_, _ = w.Write(cached)
		return
	}

	var d *Discovery
	switch {
	case page == "home":
		d, err = s.homePage(ctx, region)
	case page == "movies":
		d, err = s.moviesPage(ctx, region, q)
	case page == "upcoming-movies":
		d, err = s.upcomingPage(ctx, region, q)
	default:
		cp, ok := categoryPages[page]
		if !ok {
			platform.Err(w, 404, "NOT_FOUND", "Unknown discovery page")
			return
		}
		d, err = s.categoryPage(ctx, region, page, cp.Category, cp.Label, q)
	}
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	d.Region = region
	d.Styles = styles
	b, _ := json.Marshal(d)
	_ = s.rdb.Set(ctx, cacheKey, b, 60*time.Second).Err()
	w.Header().Set("X-Discovery-Cache", "MISS")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	platform.Cache(w, 30)
	_, _ = w.Write(b)
}

func (s *Server) homePage(ctx context.Context, region Region) (*Discovery, error) {
	d := &Discovery{PageID: "home", FilterRoute: "/home/" + region.Slug,
		Header: map[string]any{"title": "Movie Tickets, Plays, Sports, Events & Cinemas near " + region.Name},
		Meta:   map[string]any{"screen_name": "home"}}

	var hero []Card
	for _, b := range banners(region.Slug) {
		hero = append(hero, Card{ID: b.ID, Kind: "banner", CtaURL: b.URL, Image: Image{URL: "/api/img/banner/" + b.ID + ".svg", AltText: b.Title}})
	}
	d.Listings = append(d.Listings, Widget{ID: "HERO_BANNERS", Type: "banner-carousel", StyleID: "widget", Cards: hero})

	movies, err := s.queryMovieCards(ctx, region, "", cardOpts{overlay: "rating"})
	if err != nil {
		return nil, err
	}
	if len(movies) > 0 {
		d.Listings = append(d.Listings, Widget{ID: "RECOMMENDED_MOVIES", Type: "horizontal-carousel", Title: "Recommended Movies",
			CtaText: "See All", CtaURL: "/explore/movies-" + region.Slug, StyleID: "widget", Cards: movies})
	}
	d.Listings = append(d.Listings, Widget{ID: "STREAM_LEADIN", Type: "image-strip", StyleID: "widget",
		Cards: []Card{{ID: "stream", Kind: "banner", CtaURL: "/explore/upcoming-movies-" + region.Slug, Image: Image{URL: "/api/img/strip/stream", AltText: "Stream"}}}})

	events, err := s.eventsIn(ctx, region, "")
	if err != nil {
		return nil, err
	}
	var tileCards []Card
	for _, t := range tiles {
		n := 0
		for _, e := range events {
			if e.Category == t.Category && (t.Genre == "" || e.Genre == t.Genre) {
				n++
			}
		}
		if n == 0 {
			continue
		}
		page := strings.ToLower(t.Category)
		link := "/explore/" + page + "-" + region.Slug
		if t.Genre != "" {
			link += "?categories=" + strings.ReplaceAll(strings.ToLower(t.Genre), " ", "-")
		}
		tileCards = append(tileCards, Card{ID: t.Slug, Kind: "category", CtaURL: link,
			Image: Image{URL: fmt.Sprintf("/api/img/tile/%s.svg?count=%d", t.Slug, n*5), AltText: t.Name}})
	}
	if len(tileCards) > 0 {
		d.Listings = append(d.Listings, Widget{ID: "LIVE_EVENTS", Type: "category-carousel", Title: "The Best of Live Events", StyleID: "widget", Cards: tileCards})
	}

	upcoming, err := s.upcomingMovies(ctx)
	if err != nil {
		return nil, err
	}
	var prem []Card
	for _, m := range upcoming {
		prem = append(prem, movieCard(m, region, cardOpts{overlay: "release", dark: true, meta: "certlang"}))
	}
	if len(prem) > 0 {
		d.Listings = append(d.Listings, Widget{ID: "PREMIERES", Type: "horizontal-carousel", Theme: "dark", Title: "Premieres",
			Subtitle: "Brand new releases every Friday", CtaText: "See All", CtaURL: "/explore/upcoming-movies-" + region.Slug,
			StyleID: "widget-dark", Cards: prem})
	}

	addEvents := func(id, title, cta string, keep func(event) bool) {
		var cards []Card
		for _, e := range events {
			if keep(e) {
				cards = append(cards, eventCard(e, region, false))
			}
		}
		if len(cards) > 0 {
			d.Listings = append(d.Listings, Widget{ID: id, Type: "horizontal-carousel", Title: title, CtaText: "See All", CtaURL: cta, StyleID: "widget", Cards: cards})
		}
	}
	addEvents("POPULAR_EVENTS", "Popular Events", "/explore/events-"+region.Slug, func(e event) bool { return e.Category == "EVENTS" })
	addEvents("SPORTS", "Top Games & Sport Events", "/explore/sports-"+region.Slug, func(e event) bool { return e.Category == "SPORTS" })
	addEvents("LAUGHTER", "Laughter Therapy", "/explore/events-"+region.Slug+"?categories=comedy-shows", func(e event) bool { return e.Genre == "Comedy Shows" })
	addEvents("ACTIVITIES", "Explore Fun Activities", "/explore/activities-"+region.Slug, func(e event) bool { return e.Category == "ACTIVITIES" })
	addEvents("PLAYS", "Theatre & Plays", "/explore/plays-"+region.Slug, func(e event) bool { return e.Category == "PLAYS" })
	return d, nil
}

func (s *Server) moviesPage(ctx context.Context, region Region, q url.Values) (*Discovery, error) {
	all, err := s.queryMovies(ctx, region, "")
	if err != nil {
		return nil, err
	}
	langs, genres, formats := selected(q, "languages"), selected(q, "genres"), selected(q, "formats")
	var allLangs, allGenres, allFormats []string
	var cards []Card
	for _, m := range all {
		allLangs = distinct(allLangs, m.Languages)
		allGenres = distinct(allGenres, m.Genres)
		allFormats = distinct(allFormats, m.Formats)
		if matchesAny(m.Languages, langs) && matchesAny(m.Genres, genres) && matchesAny(m.Formats, formats) {
			cards = append(cards, movieCard(m, region, cardOpts{overlay: "rating", meta: "certlang"}))
		}
	}
	slices.Sort(allGenres)
	d := &Discovery{PageID: "movies", FilterRoute: "/movies-" + region.Slug,
		Header: map[string]any{"title": "Movies In " + region.Name},
		Meta:   map[string]any{"screen_name": "movies_listing"},
		Filters: []Filter{
			buildFilter("languages", "Languages", true, allLangs, langs),
			buildFilter("genres", "Genres", false, allGenres, genres),
			buildFilter("formats", "Format", false, allFormats, formats),
		},
		Breadcrumbs: []Crumb{{"Home", "/explore/home/" + region.Slug}, {"Movies in " + region.Name, ""}},
	}
	d.Listings = append(d.Listings, Widget{ID: "COMING_SOON", Type: "coming-soon-strip", Title: "Coming Soon", CtaText: "Explore Upcoming Movies",
		CtaURL: "/explore/upcoming-movies-" + region.Slug, StyleID: "widget"})
	d.Listings = append(d.Listings, Widget{ID: "MOVIES_GRID", Type: "grid", StyleID: "widget", Cards: cards})
	if len(cards) == 0 {
		d.Empty = emptyFilters()
	}
	return d, nil
}

func emptyFilters() map[string]any {
	return map[string]any{"title": "Caught up in the filter maze", "message": "Your filter combination didn't fetch any results. Kindly reset to discover more.", "cta": "Clear Filters"}
}

func (s *Server) upcomingPage(ctx context.Context, region Region, q url.Values) (*Discovery, error) {
	all, err := s.upcomingMovies(ctx)
	if err != nil {
		return nil, err
	}
	langs, genres := selected(q, "languages"), selected(q, "genres")
	var allLangs, allGenres []string
	var cards []Card
	for _, m := range all {
		allLangs = distinct(allLangs, m.Languages)
		allGenres = distinct(allGenres, m.Genres)
		if matchesAny(m.Languages, langs) && matchesAny(m.Genres, genres) {
			cards = append(cards, movieCard(m, region, cardOpts{overlay: "release", meta: "release"}))
		}
	}
	slices.Sort(allGenres)
	d := &Discovery{PageID: "upcoming-movies", FilterRoute: "/upcoming-movies-" + region.Slug,
		Header: map[string]any{"title": "Upcoming Movies"},
		Meta:   map[string]any{"screen_name": "upcoming_movies_listing"},
		Filters: []Filter{
			buildFilter("languages", "Languages", true, allLangs, langs),
			buildFilter("genres", "Genres", false, allGenres, genres),
		},
		Breadcrumbs: []Crumb{{"Home", "/explore/home/" + region.Slug}, {"Upcoming Movies", ""}},
		Listings:    []Widget{{ID: "UPCOMING_GRID", Type: "grid", StyleID: "widget", Cards: cards}},
	}
	if len(cards) == 0 {
		d.Empty = emptyFilters()
	}
	return d, nil
}

var priceBands = []struct {
	Code, Label string
	Min, Max    int
}{
	{"free", "Free", 0, 0}, {"0-500", "Under ₹500", 1, 500}, {"501-2000", "₹501 - ₹2000", 501, 2000}, {"above-2000", "Above ₹2000", 2001, 1 << 30},
}

func (s *Server) categoryPage(ctx context.Context, region Region, page, category, label string, q url.Values) (*Discovery, error) {
	events, err := s.eventsIn(ctx, region, category)
	if err != nil {
		return nil, err
	}
	dates, genres, langs, prices := selected(q, "date"), selected(q, "categories"), selected(q, "languages"), selected(q, "price")
	today := platform.Today()
	dateMatch := func(e event) bool {
		if len(dates) == 0 {
			return true
		}
		day := time.Date(e.StartsAt.In(platform.IST).Year(), e.StartsAt.In(platform.IST).Month(), e.StartsAt.In(platform.IST).Day(), 0, 0, 0, 0, platform.IST)
		for _, d := range dates {
			switch d {
			case "today":
				if day.Equal(today) {
					return true
				}
			case "tomorrow":
				if day.Equal(today.AddDate(0, 0, 1)) {
					return true
				}
			case "this-weekend":
				if wd := day.Weekday(); (wd == time.Saturday || wd == time.Sunday) && day.Sub(today) < 7*24*time.Hour {
					return true
				}
			}
		}
		return false
	}
	priceMatch := func(e event) bool {
		if len(prices) == 0 {
			return true
		}
		p := e.priceFrom()
		for _, b := range priceBands {
			if slices.Contains(prices, b.Code) && p >= b.Min && p <= b.Max {
				return true
			}
		}
		return false
	}
	var allGenres, allLangs []string
	var cards []Card
	for _, e := range events {
		allGenres = distinct(allGenres, []string{e.Genre})
		allLangs = distinct(allLangs, e.Languages)
		if dateMatch(e) && matchesAny([]string{e.Genre}, genres) && matchesAny(e.Languages, langs) && priceMatch(e) {
			cards = append(cards, eventCard(e, region, false))
		}
	}
	dateF := Filter{QueryParam: "date", Title: "Date"}
	for _, it := range []FilterItem{{"today", "Today", false}, {"tomorrow", "Tomorrow", false}, {"this-weekend", "This Weekend", false}} {
		it.Selected = slices.Contains(dates, it.Code)
		dateF.Items = append(dateF.Items, it)
	}
	priceF := Filter{QueryParam: "price", Title: "Price"}
	for _, b := range priceBands {
		priceF.Items = append(priceF.Items, FilterItem{b.Code, b.Label, slices.Contains(prices, b.Code)})
	}
	d := &Discovery{PageID: page, FilterRoute: "/" + page + "-" + region.Slug,
		Header: map[string]any{"title": label + " In " + region.Name},
		Meta:   map[string]any{"screen_name": page + "_listing"},
		Filters: []Filter{dateF, buildFilter("categories", "Categories", true, allGenres, genres),
			buildFilter("languages", "Languages", false, allLangs, langs), priceF},
		Breadcrumbs: []Crumb{{"Home", "/explore/home/" + region.Slug}, {label + " in " + region.Name, ""}},
		Listings:    []Widget{{ID: strings.ToUpper(page) + "_GRID", Type: "grid", StyleID: "widget", Cards: cards}},
	}
	if len(cards) == 0 {
		d.Empty = emptyFilters()
	}
	return d, nil
}
