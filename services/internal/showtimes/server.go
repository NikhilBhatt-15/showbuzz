// Package showtimes serves showtime listings, cinema listings and the
// seat-layout "primary" payload (show strip + category prices). Availability
// combines committed seats in Postgres with live seat locks in Redis.
package showtimes

import (
	"context"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/data"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

// CutoffAfterStart mirrors BookMyShow's cutOffDateTime: bookings stay open
// until 15 minutes after the show starts.
const CutoffAfterStart = 15 * time.Minute

type Server struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func New(db *pgxpool.Pool, rdb *redis.Client) *Server { return &Server{db: db, rdb: rdb} }

func (s *Server) Routes(r chi.Router) {
	r.Get("/api/movies-data/v4/showtimes-by-event/static", s.static)
	r.Get("/api/movies-data/v4/showtimes-by-event/primary-dynamic", s.primaryDynamic)
	r.Get("/api/movies-data/seatlayout/v1/primary", s.seatLayoutPrimary)
	r.Get("/api/movies-data/v1/cinemas", s.cinemas)
	r.Get("/api/movies-data/v1/venue-showtimes", s.venueShowtimes)
}

// LocksKey is the Redis sorted set of locked seats for a show (member = seat
// ID, score = lock expiry in unix ms). Shared with the trans service.
func LocksKey(sessionID int64) string { return "locks:" + strconv.FormatInt(sessionID, 10) }

type show struct {
	SessionID  int64
	MovieCode  string
	VenueCode  string
	ScreenID   int
	ScreenName string
	Template   string
	StartsAt   time.Time
	DateCode   string
	Language   string
	Format     string
	Attribute  string
	Prices     map[string]int
	Total      int
}

const showCols = `s.session_id, s.movie_code, s.venue_code, s.screen_id, sc.name, sc.layout_template, s.starts_at, s.date_code,
	s.language, s.format, s.attribute, s.prices, s.total_seats`

func scanShow(r pgx.CollectableRow) (show, error) {
	var sh show
	err := r.Scan(&sh.SessionID, &sh.MovieCode, &sh.VenueCode, &sh.ScreenID, &sh.ScreenName, &sh.Template, &sh.StartsAt,
		&sh.DateCode, &sh.Language, &sh.Format, &sh.Attribute, &sh.Prices, &sh.Total)
	return sh, err
}

func (sh show) bookable(now time.Time) bool { return now.Before(sh.StartsAt.Add(CutoffAfterStart)) }

type Category struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Status string `json:"status"`
}

type Showtime struct {
	SessionID  int64      `json:"sessionId"`
	Time       string     `json:"time"`
	StartsAt   time.Time  `json:"startsAt"`
	CutoffAt   time.Time  `json:"cutoffAt"`
	Language   string     `json:"language"`
	Format     string     `json:"format"`
	Attribute  string     `json:"attribute"`
	ScreenName string     `json:"screenName"`
	Status     string     `json:"status"` // AVAILABLE | FAST_FILLING | SOLD_OUT
	Available  int        `json:"available"`
	Total      int        `json:"total"`
	Late       bool       `json:"lateNight,omitempty"`
	Early      bool       `json:"earlyMorning,omitempty"`
	Categories []Category `json:"categories"`
}

func status(avail, total int) string {
	switch {
	case avail <= 0:
		return "SOLD_OUT"
	case float64(avail)/float64(total) < 0.25:
		return "FAST_FILLING"
	}
	return "AVAILABLE"
}

// availability returns, per show, the number of taken (sold + locked) seats
// in each area.
func (s *Server) availability(ctx context.Context, shows []show) (map[int64]map[string]int, error) {
	out := make(map[int64]map[string]int, len(shows))
	if len(shows) == 0 {
		return out, nil
	}
	ids := make([]int64, len(shows))
	tpl := make(map[int64]string, len(shows))
	for i, sh := range shows {
		ids[i] = sh.SessionID
		tpl[sh.SessionID] = sh.Template
		out[sh.SessionID] = map[string]int{}
	}
	taken := make(map[int64]map[string]bool, len(shows))
	rows, err := s.db.Query(ctx, `SELECT session_id, seat_id FROM booking_seats WHERE session_id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var seat string
		if err := rows.Scan(&id, &seat); err != nil {
			rows.Close()
			return nil, err
		}
		if taken[id] == nil {
			taken[id] = map[string]bool{}
		}
		taken[id][seat] = true
	}
	rows.Close()

	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	pipe := s.rdb.Pipeline()
	cmds := make(map[int64]*redis.StringSliceCmd, len(ids))
	for _, id := range ids {
		cmds[id] = pipe.ZRangeByScore(ctx, LocksKey(id), &redis.ZRangeBy{Min: "(" + now, Max: "+inf"})
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}
	for id, cmd := range cmds {
		for _, seat := range cmd.Val() {
			if taken[id] == nil {
				taken[id] = map[string]bool{}
			}
			taken[id][seat] = true
		}
	}
	for id, seats := range taken {
		for seat := range seats {
			if a, ok := data.SeatArea(tpl[id], seat); ok {
				out[id][a]++
			}
		}
	}
	return out, nil
}

func buildShowtime(sh show, taken map[string]int) Showtime {
	layout, _ := data.GetLayout(sh.Template)
	counts := data.AreaSeatCounts(sh.Template)
	st := Showtime{SessionID: sh.SessionID, Time: sh.StartsAt.In(platform.IST).Format("03:04 PM"), StartsAt: sh.StartsAt,
		CutoffAt: sh.StartsAt.Add(CutoffAfterStart), Language: sh.Language, Format: sh.Format, Attribute: sh.Attribute,
		ScreenName: sh.ScreenName, Total: sh.Total}
	hour := sh.StartsAt.In(platform.IST).Hour()
	st.Late = hour >= 23
	st.Early = hour < 10
	// Categories are listed cheapest first, as in the "How many seats?" dialog.
	for i := len(layout.Areas) - 1; i >= 0; i-- {
		a := layout.Areas[i]
		avail := counts[a.Code] - taken[a.Code]
		st.Available += avail
		cs := status(avail, counts[a.Code])
		st.Categories = append(st.Categories, Category{Code: a.Code, Name: a.Name, Price: sh.Prices[a.Code], Status: cs})
	}
	st.Status = status(st.Available, st.Total)
	return st
}

func (s *Server) static(w http.ResponseWriter, _ *http.Request) {
	platform.Cache(w, 3600)
	platform.JSON(w, 200, map[string]any{
		"seatLegends": []map[string]string{
			{"status": "1", "label": "Available"}, {"status": "2", "label": "Sold"},
			{"status": "4", "label": "Bestseller", "infoText": "Bestsellers are the most in demand seats for any cinema"},
			{"status": "88", "label": "Selected"},
		},
		"availabilityLegends": []map[string]string{{"status": "AVAILABLE", "label": "AVAILABLE"}, {"status": "FAST_FILLING", "label": "FAST FILLING"}},
		"timeBuckets": []map[string]any{
			{"code": "morning", "label": "Morning", "sub": "12:00-11:59 AM", "from": 0, "to": 12},
			{"code": "afternoon", "label": "Afternoon", "sub": "12:00-3:59 PM", "from": 12, "to": 16},
			{"code": "evening", "label": "Evening", "sub": "4:00-6:59 PM", "from": 16, "to": 19},
			{"code": "night", "label": "Night", "sub": "7:00-11:59 PM", "from": 19, "to": 24},
		},
		"priceBuckets": []map[string]any{
			{"code": "0-200", "label": "₹0 - ₹200", "from": 0, "to": 200},
			{"code": "201-400", "label": "₹201 - ₹400", "from": 201, "to": 400},
			{"code": "401-600", "label": "₹401 - ₹600", "from": 401, "to": 600},
			{"code": "601-1000", "label": "₹601 - ₹1000", "from": 601, "to": 1000},
			{"code": "1001+", "label": "₹1001+", "from": 1001, "to": 100000},
		},
		"errors": map[string]any{
			"filter":  map[string]string{"title": "Caught up in the filter maze", "message": "Your filter combination didn't fetch any results. Kindly reset to discover more.", "cta": "Clear Filters"},
			"noShows": map[string]string{"title": "Uh-oh! We couldn't find anything.", "message": "There are no shows available in this region.", "cta": "Change Location"},
		},
		"maxTickets":              10,
		"cutoffMinutesAfterStart": int(CutoffAfterStart.Minutes()),
	})
}

type movieInfo struct {
	Code        string   `json:"code"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Certificate string   `json:"certificate"`
	Genres      []string `json:"genres"`
	Languages   []string `json:"languages"`
	Formats     []string `json:"formats"`
	DurationMin int      `json:"durationMin"`
}

func (s *Server) movie(ctx context.Context, code string) (movieInfo, error) {
	var m movieInfo
	err := s.db.QueryRow(ctx, `SELECT code, slug, title, certificate, genres, languages, formats, duration_min FROM movies WHERE code = $1`, code).
		Scan(&m.Code, &m.Slug, &m.Title, &m.Certificate, &m.Genres, &m.Languages, &m.Formats, &m.DurationMin)
	return m, err
}

type regionInfo struct {
	Code, Name, Slug string
}

func (s *Server) region(ctx context.Context, key string) (regionInfo, error) {
	var g regionInfo
	err := s.db.QueryRow(ctx, `SELECT code, name, slug FROM regions WHERE code = upper($1) OR slug = lower($1)
		ORDER BY sort LIMIT 1`, key).Scan(&g.Code, &g.Name, &g.Slug)
	if err == pgx.ErrNoRows {
		err = s.db.QueryRow(ctx, `SELECT code, name, slug FROM regions ORDER BY sort LIMIT 1`).Scan(&g.Code, &g.Name, &g.Slug)
	}
	return g, err
}

type venueInfo struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Locality     string  `json:"locality"`
	Address      string  `json:"address"`
	CompanyCode  string  `json:"companyCode"`
	Cancellation bool    `json:"cancellation"`
	FnB          bool    `json:"fnb"`
	MTicket      bool    `json:"mticket"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	Logo         string  `json:"logo"`
	Popularity   int     `json:"-"`
}

const venueCols = `code, name, slug, locality, address, company_code, cancellation, fnb, mticket, lat, lng, popularity`

func scanVenue(r pgx.CollectableRow) (venueInfo, error) {
	var v venueInfo
	err := r.Scan(&v.Code, &v.Name, &v.Slug, &v.Locality, &v.Address, &v.CompanyCode, &v.Cancellation, &v.FnB, &v.MTicket, &v.Lat, &v.Lng, &v.Popularity)
	v.Logo = "/api/img/logo/" + v.CompanyCode + ".svg"
	return v, err
}

type dateItem struct {
	Code    string `json:"code"`
	Day     string `json:"day"`
	Date    string `json:"date"`
	Month   string `json:"month"`
	Enabled bool   `json:"enabled"`
}

func dateStrip(enabled map[string]bool) []dateItem {
	var out []dateItem
	for i := 0; i < 7; i++ {
		d := platform.Today().AddDate(0, 0, i)
		code := d.Format("20060102")
		out = append(out, dateItem{Code: code, Day: strings.ToUpper(d.Format("Mon")), Date: d.Format("02"),
			Month: strings.ToUpper(d.Format("Jan")), Enabled: enabled[code]})
	}
	return out
}

func (s *Server) primaryDynamic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	m, err := s.movie(ctx, q.Get("eventCode"))
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Movie not found")
		return
	}
	g, err := s.region(ctx, q.Get("regionCode"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	now := time.Now()

	rows, err := s.db.Query(ctx, `SELECT `+showCols+` FROM shows s JOIN screens sc ON sc.id = s.screen_id JOIN venues v ON v.code = s.venue_code
		WHERE s.movie_code = $1 AND v.region_code = $2 AND s.date_code >= $3 AND s.date_code < $4 ORDER BY s.starts_at`,
		m.Code, g.Code, platform.DateCode(now), platform.DateCode(now.AddDate(0, 0, 7)))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	all, err := pgx.CollectRows(rows, scanShow)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	enabled := map[string]bool{}
	for _, sh := range all {
		if sh.bookable(now) {
			enabled[sh.DateCode] = true
		}
	}
	dates := dateStrip(enabled)
	dateCode := q.Get("dateCode")
	if !enabled[dateCode] {
		dateCode = ""
		for _, d := range dates {
			if d.Enabled {
				dateCode = d.Code
				break
			}
		}
	}

	// Language/format combinations available on the chosen date.
	type combo struct{ Lang, Format string }
	comboCount := map[combo]int{}
	var day []show
	for _, sh := range all {
		if sh.DateCode == dateCode && sh.bookable(now) {
			day = append(day, sh)
			comboCount[combo{sh.Language, sh.Format}]++
		}
	}
	var combos []combo
	for c := range comboCount {
		combos = append(combos, c)
	}
	sort.Slice(combos, func(i, j int) bool {
		if comboCount[combos[i]] != comboCount[combos[j]] {
			return comboCount[combos[i]] > comboCount[combos[j]]
		}
		return combos[i].Lang+combos[i].Format < combos[j].Lang+combos[j].Format
	})
	sel := combo{q.Get("lang"), q.Get("format")}
	if _, ok := comboCount[sel]; !ok && len(combos) > 0 {
		sel = combos[0]
	}
	var formatOpts []map[string]any
	for _, c := range combos {
		formatOpts = append(formatOpts, map[string]any{"language": c.Lang, "format": c.Format,
			"label": c.Lang + " - " + c.Format, "selected": c == sel})
	}

	var chosen []show
	for _, sh := range day {
		if sh.Language == sel.Lang && sh.Format == sel.Format {
			chosen = append(chosen, sh)
		}
	}
	taken, err := s.availability(ctx, chosen)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}

	vrows, err := s.db.Query(ctx, `SELECT `+venueCols+` FROM venues WHERE region_code = $1`, g.Code)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	venues, err := pgx.CollectRows(vrows, scanVenue)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	byVenue := map[string][]Showtime{}
	for _, sh := range chosen {
		byVenue[sh.VenueCode] = append(byVenue[sh.VenueCode], buildShowtime(sh, taken[sh.SessionID]))
	}
	type venueShows struct {
		venueInfo
		Showtimes []Showtime `json:"showtimes"`
	}
	var list []venueShows
	for _, v := range venues {
		if st := byVenue[v.Code]; len(st) > 0 {
			list = append(list, venueShows{v, st})
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].Popularity > list[j].Popularity })

	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{
		"event":    m,
		"region":   map[string]string{"code": g.Code, "name": g.Name, "slug": g.Slug},
		"dateCode": dateCode,
		"dates":    dates,
		"formats":  formatOpts,
		"selected": map[string]string{"language": sel.Lang, "format": sel.Format},
		"venues":   list,
	})
}

func (s *Server) seatLayoutPrimary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	sessionID, _ := strconv.ParseInt(q.Get("sessionId"), 10, 64)
	rows, err := s.db.Query(ctx, `SELECT `+showCols+` FROM shows s JOIN screens sc ON sc.id = s.screen_id WHERE s.session_id = $1`, sessionID)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	cur, err := pgx.CollectExactlyOneRow(rows, scanShow)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Show not found")
		return
	}
	m, err := s.movie(ctx, cur.MovieCode)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	vrows, err := s.db.Query(ctx, `SELECT `+venueCols+` FROM venues WHERE code = $1`, cur.VenueCode)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	v, err := pgx.CollectExactlyOneRow(vrows, scanVenue)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	rows, err = s.db.Query(ctx, `SELECT `+showCols+` FROM shows s JOIN screens sc ON sc.id = s.screen_id
		WHERE s.movie_code = $1 AND s.venue_code = $2 AND s.date_code = $3 AND s.language = $4 AND s.format = $5 ORDER BY s.starts_at`,
		cur.MovieCode, cur.VenueCode, cur.DateCode, cur.Language, cur.Format)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	siblings, err := pgx.CollectRows(rows, scanShow)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	now := time.Now()
	siblings = slices.DeleteFunc(siblings, func(sh show) bool { return !sh.bookable(now) && sh.SessionID != cur.SessionID })
	taken, err := s.availability(ctx, siblings)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	var strip []Showtime
	var current Showtime
	for _, sh := range siblings {
		st := buildShowtime(sh, taken[sh.SessionID])
		if sh.SessionID == cur.SessionID {
			current = st
		}
		strip = append(strip, st)
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{
		"event":      m,
		"venue":      v,
		"dateCode":   cur.DateCode,
		"show":       current,
		"bookable":   cur.bookable(now),
		"showTimes":  strip,
		"maxTickets": 10,
	})
}

func (s *Server) cinemas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	g, err := s.region(ctx, r.URL.Query().Get("regionCode"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	rows, err := s.db.Query(ctx, `SELECT `+venueCols+` FROM venues WHERE region_code = $1 ORDER BY popularity DESC`, g.Code)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	venues, err := pgx.CollectRows(rows, scanVenue)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.Cache(w, 60)
	platform.JSON(w, 200, map[string]any{"region": map[string]string{"code": g.Code, "name": g.Name, "slug": g.Slug},
		"dateCode": platform.DateCode(time.Now()), "venues": venues})
}

func (s *Server) venueShowtimes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	vrows, err := s.db.Query(ctx, `SELECT `+venueCols+` FROM venues WHERE code = $1`, q.Get("venueCode"))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	v, err := pgx.CollectExactlyOneRow(vrows, scanVenue)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Cinema not found")
		return
	}
	now := time.Now()
	rows, err := s.db.Query(ctx, `SELECT `+showCols+` FROM shows s JOIN screens sc ON sc.id = s.screen_id
		WHERE s.venue_code = $1 AND s.date_code >= $2 AND s.date_code < $3 ORDER BY s.starts_at`,
		v.Code, platform.DateCode(now), platform.DateCode(now.AddDate(0, 0, 7)))
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	all, err := pgx.CollectRows(rows, scanShow)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	enabled := map[string]bool{}
	for _, sh := range all {
		if sh.bookable(now) {
			enabled[sh.DateCode] = true
		}
	}
	dates := dateStrip(enabled)
	dateCode := q.Get("dateCode")
	if !enabled[dateCode] {
		for _, d := range dates {
			if d.Enabled {
				dateCode = d.Code
				break
			}
		}
	}
	var day []show
	for _, sh := range all {
		if sh.DateCode == dateCode && sh.bookable(now) {
			day = append(day, sh)
		}
	}
	taken, err := s.availability(ctx, day)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	type group struct {
		Movie     movieInfo  `json:"movie"`
		Language  string     `json:"language"`
		Format    string     `json:"format"`
		Showtimes []Showtime `json:"showtimes"`
	}
	var groups []*group
	idx := map[string]*group{}
	for _, sh := range day {
		key := sh.MovieCode + "|" + sh.Language + "|" + sh.Format
		gr, ok := idx[key]
		if !ok {
			m, err := s.movie(ctx, sh.MovieCode)
			if err != nil {
				continue
			}
			gr = &group{Movie: m, Language: sh.Language, Format: sh.Format}
			idx[key] = gr
			groups = append(groups, gr)
		}
		gr.Showtimes = append(gr.Showtimes, buildShowtime(sh, taken[sh.SessionID]))
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{"venue": v, "dateCode": dateCode, "dates": dates, "movies": groups})
}
