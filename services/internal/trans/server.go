// Package trans is the transactional core — the equivalent of BookMyShow's
// doTrans service. It renders seat grids with live status, locks seats for a
// pending transaction, prices the order (tickets, convenience fee + GST, F&B),
// takes a (simulated) payment and commits the booking.
package trans

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/data"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

const (
	MaxTickets      = 10
	convFeePerSeat  = 25.0 // movie convenience fee base, per ticket
	eventFeePercent = 0.06 // event convenience fee base
	gstRate         = 0.18 // IGST on the convenience fee
	cutoffAfter     = 15 * time.Minute
	cancelBefore    = 20 * time.Minute
)

type Server struct {
	db      *pgxpool.Pool
	rdb     *redis.Client
	pub     *platform.Producer
	holdTTL time.Duration
}

func New(db *pgxpool.Pool, rdb *redis.Client, pub *platform.Producer) *Server {
	return &Server{db: db, rdb: rdb, pub: pub, holdTTL: time.Duration(platform.EnvInt("SEAT_HOLD_SECONDS", 600)) * time.Second}
}

func (s *Server) Routes(r chi.Router) {
	r.Route("/api/trans/v1", func(r chi.Router) {
		r.Get("/seat-layout/{session}", s.seatLayout)
		r.Get("/fnb", s.fnbMenu)
		r.Get("/events/{code}/inventory", s.eventInventory)
		r.Post("/transactions", s.createMovieTxn)
		r.Post("/event-transactions", s.createEventTxn)
		r.Get("/transactions/{id}", s.getTxn)
		r.Patch("/transactions/{id}", s.updateTxn)
		r.Delete("/transactions/{id}", s.cancelTxn)
		r.With(platform.RequireUser).Post("/transactions/{id}/pay", s.pay)
		r.With(platform.RequireUser).Get("/bookings", s.listBookings)
		r.With(platform.RequireUser).Get("/bookings/{id}", s.getBooking)
		r.With(platform.RequireUser).Post("/bookings/{id}/cancel", s.cancelBooking)
	})
}

// ---- show lookups ----------------------------------------------------------

type showInfo struct {
	SessionID    int64
	MovieCode    string
	MovieSlug    string
	Title        string
	Certificate  string
	VenueCode    string
	VenueName    string
	VenueAddress string
	RegionCode   string
	StateName    string
	Cancellation bool
	FnB          bool
	ScreenName   string
	Template     string
	StartsAt     time.Time
	DateCode     string
	Language     string
	Format       string
	Prices       map[string]int
}

func (s *Server) loadShow(ctx context.Context, session int64) (showInfo, error) {
	var sh showInfo
	err := s.db.QueryRow(ctx, `
		SELECT s.session_id, m.code, m.slug, m.title, m.certificate, v.code, v.name, v.address, v.region_code, r.state_name,
			v.cancellation, v.fnb, sc.name, sc.layout_template, s.starts_at, s.date_code, s.language, s.format, s.prices
		FROM shows s JOIN movies m ON m.code = s.movie_code JOIN venues v ON v.code = s.venue_code
		JOIN screens sc ON sc.id = s.screen_id JOIN regions r ON r.code = v.region_code
		WHERE s.session_id = $1`, session).Scan(&sh.SessionID, &sh.MovieCode, &sh.MovieSlug, &sh.Title, &sh.Certificate,
		&sh.VenueCode, &sh.VenueName, &sh.VenueAddress, &sh.RegionCode, &sh.StateName, &sh.Cancellation, &sh.FnB,
		&sh.ScreenName, &sh.Template, &sh.StartsAt, &sh.DateCode, &sh.Language, &sh.Format, &sh.Prices)
	return sh, err
}

func (sh showInfo) bookable() bool { return time.Now().Before(sh.StartsAt.Add(cutoffAfter)) }

// ---- seat layout -----------------------------------------------------------

// Seat status codes match BookMyShow's seatLegends: 1 available, 2 sold,
// 4 bestseller. 88 (selected) is applied client-side.
const (
	SeatAvailable  = 1
	SeatSold       = 2
	SeatBestseller = 4
)

type gridSeat struct {
	ID     string `json:"id"`
	N      int    `json:"n"`
	C      int    `json:"c"`
	Status int    `json:"s"`
}

type gridRow struct {
	Label string     `json:"label"`
	Seats []gridSeat `json:"seats"`
}

type gridArea struct {
	Code  string    `json:"code"`
	Name  string    `json:"name"`
	Price int       `json:"price"`
	Rows  []gridRow `json:"rows"`
}

func (s *Server) seatLayout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, _ := strconv.ParseInt(chi.URLParam(r, "session"), 10, 64)
	sh, err := s.loadShow(ctx, session)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Err(w, 404, "NOT_FOUND", "Show not found")
		return
	} else if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	layout, ok := data.GetLayout(sh.Template)
	if !ok {
		platform.Err(w, 500, "INTERNAL", "unknown layout")
		return
	}
	taken := map[string]bool{}
	rows, err := s.db.Query(ctx, `SELECT seat_id FROM booking_seats WHERE session_id = $1`, session)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			taken[id] = true
		}
	}
	rows.Close()
	locked, err := lockedSeats(ctx, s.rdb, session)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	// A transaction re-opening its own layout sees its seats as available.
	mine := map[string]bool{}
	if id := r.URL.Query().Get("txn"); id != "" {
		if t, err := s.loadTxn(ctx, id); err == nil && t.SessionID == session {
			for _, seat := range t.Seats {
				mine[seat] = true
			}
		}
	}
	for _, seat := range locked {
		if !mine[seat] {
			taken[seat] = true
		}
	}

	var areas []gridArea
	available := 0
	for _, a := range layout.Areas {
		ga := gridArea{Code: a.Code, Name: a.Name, Price: sh.Prices[a.Code]}
		for _, row := range a.Rows {
			gr := gridRow{Label: row.Label}
			for _, seat := range row.Seats {
				st := SeatAvailable
				switch {
				case taken[seat.ID]:
					st = SeatSold
				case seat.Best:
					st = SeatBestseller
				}
				if st != SeatSold {
					available++
				}
				gr.Seats = append(gr.Seats, gridSeat{ID: seat.ID, N: seat.N, C: seat.C, Status: st})
			}
			ga.Rows = append(ga.Rows, gr)
		}
		areas = append(areas, ga)
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{
		"sessionId":  session,
		"bookable":   sh.bookable(),
		"cols":       layout.Cols,
		"areas":      areas,
		"available":  available,
		"maxTickets": MaxTickets,
		"screenName": sh.ScreenName,
	})
}

// ---- F&B -------------------------------------------------------------------

type fnbItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Category string `json:"category"`
	Price    int    `json:"price"`
	Veg      bool   `json:"veg"`
	Image    string `json:"image"`
}

var fnbMenu = []fnbItem{
	{"COMB", "Couple Combo", "Large salted popcorn + 2 Cola (450ml)", "Combos", 620, true, ""},
	{"POPS", "Salted Popcorn (Regular)", "Classic butter-salted popcorn, 90g", "Popcorn", 290, true, ""},
	{"POPC", "Caramel Popcorn (Large)", "Sweet caramel-coated popcorn, 150g", "Popcorn", 410, true, ""},
	{"COKE", "Cola (450ml)", "Chilled fountain cola", "Beverages", 250, true, ""},
	{"COFF", "Cold Coffee", "Thick, creamy and chilled", "Beverages", 280, true, ""},
	{"WATR", "Mineral Water", "500ml bottle", "Beverages", 80, true, ""},
	{"NACH", "Nachos with Cheese", "Corn tortilla chips with warm cheese dip", "Snacks", 330, true, ""},
	{"PUFF", "Veg Puff", "Flaky pastry with spiced vegetables", "Snacks", 180, true, ""},
	{"SAMO", "Samosa (2 pcs)", "With mint chutney", "Snacks", 160, true, ""},
	{"HOTD", "Chicken Hot Dog", "Grilled sausage, mustard and onions", "Snacks", 300, false, ""},
}

func init() {
	for i := range fnbMenu {
		fnbMenu[i].Image = "/api/img/fnb/" + fnbMenu[i].Code + ".svg"
	}
}

func (s *Server) fnbMenu(w http.ResponseWriter, _ *http.Request) {
	platform.Cache(w, 600)
	platform.JSON(w, 200, map[string]any{"categories": []string{"All", "Combos", "Popcorn", "Beverages", "Snacks"}, "items": fnbMenu})
}

// ---- transactions ----------------------------------------------------------

type Line struct {
	AreaCode string   `json:"areaCode"`
	AreaName string   `json:"areaName"`
	Price    float64  `json:"price"`
	Seats    []string `json:"seats"`
}

type FnbLine struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Qty   int     `json:"qty"`
}

type Snapshot struct {
	Title        string    `json:"title"`
	Certificate  string    `json:"certificate,omitempty"`
	Language     string    `json:"language,omitempty"`
	Format       string    `json:"format,omitempty"`
	EventCode    string    `json:"eventCode"`
	Slug         string    `json:"slug"`
	VenueCode    string    `json:"venueCode,omitempty"`
	VenueName    string    `json:"venueName"`
	VenueAddress string    `json:"venueAddress,omitempty"`
	ScreenName   string    `json:"screenName,omitempty"`
	StartsAt     time.Time `json:"startsAt"`
	DateCode     string    `json:"dateCode,omitempty"`
	StateName    string    `json:"stateName,omitempty"`
	Poster       string    `json:"poster"`
	Cancellation bool      `json:"cancellation"`
	FnB          bool      `json:"fnb"`
	TierName     string    `json:"tierName,omitempty"`
	Lines        []Line    `json:"lines"`
}

type Txn struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"` // MOVIE | EVENT
	SessionID int64     `json:"sessionId,omitempty"`
	Seats     []string  `json:"seats,omitempty"`
	EventCode string    `json:"eventCode,omitempty"`
	Tier      string    `json:"tier,omitempty"`
	Qty       int       `json:"qty"`
	Snapshot  Snapshot  `json:"snapshot"`
	Fnb       []FnbLine `json:"fnb"`
	Donation  bool      `json:"donation"`
	UserID    string    `json:"userId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Pricing struct {
	Ticket   float64 `json:"ticketAmount"`
	ConvBase float64 `json:"convBase"`
	GST      float64 `json:"gst"`
	ConvFee  float64 `json:"convFee"`
	SubTotal float64 `json:"subTotal"`
	Fnb      float64 `json:"fnbAmount"`
	Donation float64 `json:"donation"`
	Total    float64 `json:"total"`
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (t *Txn) Pricing() Pricing {
	var p Pricing
	for _, l := range t.Snapshot.Lines {
		p.Ticket += l.Price * float64(len(l.Seats))
	}
	if t.Kind == "EVENT" {
		p.Ticket = t.Snapshot.Lines[0].Price * float64(t.Qty)
		p.ConvBase = round2(p.Ticket * eventFeePercent)
	} else {
		p.ConvBase = convFeePerSeat * float64(t.Qty)
	}
	p.GST = round2(p.ConvBase * gstRate)
	p.ConvFee = round2(p.ConvBase + p.GST)
	p.SubTotal = round2(p.Ticket + p.ConvFee)
	for _, f := range t.Fnb {
		p.Fnb += f.Price * float64(f.Qty)
	}
	if t.Donation {
		p.Donation = float64(t.Qty) // ₹1 per ticket, like BookASmile
	}
	p.Total = round2(p.SubTotal + p.Fnb + p.Donation)
	return p
}

func txnKey(id string) string { return "txn:" + id }

func (s *Server) saveTxn(ctx context.Context, t *Txn) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, txnKey(t.ID), b, time.Until(t.ExpiresAt)+2*time.Minute).Err()
}

var errExpired = errors.New("transaction expired")

func (s *Server) loadTxn(ctx context.Context, id string) (*Txn, error) {
	b, err := s.rdb.Get(ctx, txnKey(id)).Bytes()
	if err == redis.Nil {
		return nil, errExpired
	} else if err != nil {
		return nil, err
	}
	var t Txn
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func newID(prefix string, n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		k, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[k.Int64()]
	}
	return prefix + string(b)
}

func (s *Server) txnView(t *Txn) map[string]any {
	return map[string]any{
		"id":          t.ID,
		"kind":        t.Kind,
		"sessionId":   t.SessionID,
		"seats":       t.Seats,
		"eventCode":   t.EventCode,
		"tier":        t.Tier,
		"qty":         t.Qty,
		"snapshot":    t.Snapshot,
		"fnb":         t.Fnb,
		"donation":    t.Donation,
		"pricing":     t.Pricing(),
		"expiresAt":   t.ExpiresAt,
		"serverTime":  time.Now(),
		"holdSeconds": int(s.holdTTL.Seconds()),
	}
}

func (s *Server) createMovieTxn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		SessionID int64    `json:"sessionId"`
		Seats     []string `json:"seats"`
	}
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	slices.Sort(req.Seats)
	req.Seats = slices.Compact(req.Seats)
	if len(req.Seats) == 0 || len(req.Seats) > MaxTickets {
		platform.Err(w, 400, "BAD_REQUEST", fmt.Sprintf("Select between 1 and %d seats", MaxTickets))
		return
	}
	sh, err := s.loadShow(ctx, req.SessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Err(w, 404, "NOT_FOUND", "Show not found")
		return
	} else if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if !sh.bookable() {
		platform.Err(w, 409, "SHOW_CUTOFF", "Sorry, you cannot book tickets as you've exceeded the cinema's booking cutoff time.")
		return
	}

	layout, _ := data.GetLayout(sh.Template)
	areaName := map[string]string{}
	for _, a := range layout.Areas {
		areaName[a.Code] = a.Name
	}
	byArea := map[string][]string{}
	for _, seat := range req.Seats {
		a, ok := data.SeatArea(sh.Template, seat)
		if !ok {
			platform.Err(w, 400, "BAD_SEAT", "Unknown seat "+seat)
			return
		}
		byArea[a] = append(byArea[a], seat)
	}

	var sold []string
	rows, err := s.db.Query(ctx, `SELECT seat_id FROM booking_seats WHERE session_id = $1 AND seat_id = ANY($2)`, sh.SessionID, req.Seats)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			sold = append(sold, id)
		}
	}
	rows.Close()
	if len(sold) > 0 {
		platform.JSON(w, 409, map[string]any{"error": platform.APIError{Code: "SEATS_UNAVAILABLE",
			Message: "Sorry, seat " + strings.Join(sold, ", ") + " was just booked by someone else. Please choose different seats."}, "seats": sold})
		return
	}

	t := &Txn{ID: newID("T", 10), Kind: "MOVIE", SessionID: sh.SessionID, Seats: req.Seats, Qty: len(req.Seats),
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(s.holdTTL), Fnb: []FnbLine{}}
	if u := platform.User(ctx); u != nil {
		t.UserID = u.UserID
	}
	conflict, err := lockSeats(ctx, s.rdb, sh.SessionID, t.ID, req.Seats, t.ExpiresAt)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if conflict != "" {
		platform.JSON(w, 409, map[string]any{"error": platform.APIError{Code: "SEATS_UNAVAILABLE",
			Message: "Sorry, seat " + conflict + " is being booked by someone else. Please choose different seats."}, "seats": []string{conflict}})
		return
	}

	// Lines are ordered by area as laid out (back to front), seats naturally.
	var lines []Line
	for _, a := range layout.Areas {
		if seats := byArea[a.Code]; len(seats) > 0 {
			slices.SortFunc(seats, seatOrder)
			lines = append(lines, Line{AreaCode: a.Code, AreaName: a.Name, Price: float64(sh.Prices[a.Code]), Seats: seats})
		}
	}
	t.Snapshot = Snapshot{Title: sh.Title, Certificate: sh.Certificate, Language: sh.Language, Format: sh.Format,
		EventCode: sh.MovieCode, Slug: sh.MovieSlug, VenueCode: sh.VenueCode, VenueName: sh.VenueName, VenueAddress: sh.VenueAddress,
		ScreenName: sh.ScreenName, StartsAt: sh.StartsAt, DateCode: sh.DateCode, StateName: sh.StateName,
		Poster: "/api/img/poster/" + sh.MovieCode + ".svg", Cancellation: sh.Cancellation, FnB: sh.FnB, Lines: lines}
	if err := s.saveTxn(ctx, t); err != nil {
		_ = releaseSeats(ctx, s.rdb, sh.SessionID, t.ID, req.Seats)
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 201, s.txnView(t))
}

// seatOrder sorts "J9" before "J10".
func seatOrder(a, b string) int {
	ra, na := splitSeat(a)
	rb, nb := splitSeat(b)
	if ra != rb {
		return strings.Compare(ra, rb)
	}
	return na - nb
}

var seatRe = regexp.MustCompile(`^([A-Z]+)(\d+)$`)

func splitSeat(s string) (string, int) {
	m := seatRe.FindStringSubmatch(s)
	if m == nil {
		return s, 0
	}
	n, _ := strconv.Atoi(m[2])
	return m[1], n
}

type eventRow struct {
	Code, Slug, Title, VenueName, Category string
	RegionCode                             *string
	StartsAt                               time.Time
	Tiers                                  []data.Tier
}

func (s *Server) loadEvent(ctx context.Context, code string) (eventRow, error) {
	var e eventRow
	err := s.db.QueryRow(ctx, `SELECT code, slug, title, venue_name, category, region_code, starts_at, tiers FROM events WHERE code = $1`, code).
		Scan(&e.Code, &e.Slug, &e.Title, &e.VenueName, &e.Category, &e.RegionCode, &e.StartsAt, &e.Tiers)
	return e, err
}

func (s *Server) soldCount(ctx context.Context, event, tier string) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(qty), 0) FROM event_sales WHERE event_code = $1 AND tier = $2`, event, tier).Scan(&n)
	return n, err
}

func (s *Server) eventInventory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	e, err := s.loadEvent(ctx, chi.URLParam(r, "code"))
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Event not found")
		return
	}
	var tiers []map[string]any
	for _, t := range e.Tiers {
		sold, err := s.soldCount(ctx, e.Code, t.Code)
		if err != nil {
			platform.Err(w, 500, "INTERNAL", err.Error())
			return
		}
		held, _ := heldTickets(ctx, s.rdb, e.Code, t.Code)
		avail := max(0, t.Capacity-sold-held)
		tiers = append(tiers, map[string]any{"code": t.Code, "name": t.Name, "price": t.Price, "note": t.Note,
			"available": avail, "fastFilling": avail > 0 && float64(avail) < float64(t.Capacity)*0.2, "maxPerBooking": min(MaxTickets, avail)})
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{"eventCode": e.Code, "startsAt": e.StartsAt, "tiers": tiers})
}

func (s *Server) createEventTxn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		EventCode  string `json:"eventCode"`
		Tier       string `json:"tier"`
		Qty        int    `json:"qty"`
		RegionName string `json:"regionName"`
	}
	if err := platform.Decode(r, &req); err != nil || req.Qty < 1 || req.Qty > MaxTickets {
		platform.Err(w, 400, "BAD_REQUEST", fmt.Sprintf("Select between 1 and %d tickets", MaxTickets))
		return
	}
	e, err := s.loadEvent(ctx, req.EventCode)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Event not found")
		return
	}
	if time.Now().After(e.StartsAt) {
		platform.Err(w, 409, "EVENT_STARTED", "Bookings for this event have closed.")
		return
	}
	i := slices.IndexFunc(e.Tiers, func(t data.Tier) bool { return t.Code == req.Tier })
	if i < 0 {
		platform.Err(w, 400, "BAD_TIER", "Unknown ticket type")
		return
	}
	tier := e.Tiers[i]
	sold, err := s.soldCount(ctx, e.Code, tier.Code)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	t := &Txn{ID: newID("T", 10), Kind: "EVENT", EventCode: e.Code, Tier: tier.Code, Qty: req.Qty,
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(s.holdTTL), Fnb: []FnbLine{}}
	if u := platform.User(ctx); u != nil {
		t.UserID = u.UserID
	}
	ok, remaining, err := holdTickets(ctx, s.rdb, e.Code, tier.Code, t.ID, req.Qty, sold, tier.Capacity, t.ExpiresAt)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if !ok {
		platform.Err(w, 409, "SOLD_OUT", fmt.Sprintf("Only %d tickets left in %s", max(0, remaining), tier.Name))
		return
	}
	venue := e.VenueName
	if e.RegionCode == nil && req.RegionName != "" {
		venue += ": " + req.RegionName
	}
	t.Snapshot = Snapshot{Title: e.Title, EventCode: e.Code, Slug: e.Slug, VenueName: venue, StartsAt: e.StartsAt,
		Poster: "/api/img/poster/" + e.Code + ".svg", TierName: tier.Name,
		Lines: []Line{{AreaCode: tier.Code, AreaName: tier.Name, Price: float64(tier.Price)}}}
	if err := s.saveTxn(ctx, t); err != nil {
		_ = releaseHold(ctx, s.rdb, e.Code, tier.Code, t.ID)
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 201, s.txnView(t))
}

func (s *Server) txnFromRequest(w http.ResponseWriter, r *http.Request) *Txn {
	t, err := s.loadTxn(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, errExpired) || (err == nil && time.Now().After(t.ExpiresAt)) {
		platform.Err(w, 410, "TXN_EXPIRED", "Your session has expired. The seats you selected have been released.")
		return nil
	} else if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return nil
	}
	if u := platform.User(r.Context()); t.UserID != "" && u != nil && u.UserID != t.UserID {
		platform.Err(w, 403, "FORBIDDEN", "This transaction belongs to another account")
		return nil
	}
	return t
}

func (s *Server) getTxn(w http.ResponseWriter, r *http.Request) {
	if t := s.txnFromRequest(w, r); t != nil {
		platform.NoStore(w)
		platform.JSON(w, 200, s.txnView(t))
	}
}

func (s *Server) updateTxn(w http.ResponseWriter, r *http.Request) {
	t := s.txnFromRequest(w, r)
	if t == nil {
		return
	}
	var req struct {
		Fnb []struct {
			Code string `json:"code"`
			Qty  int    `json:"qty"`
		} `json:"fnb"`
		Donation *bool `json:"donation"`
	}
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	if req.Fnb != nil {
		if t.Kind != "MOVIE" || !t.Snapshot.FnB {
			platform.Err(w, 400, "FNB_UNAVAILABLE", "Food & beverages are not available at this venue")
			return
		}
		t.Fnb = []FnbLine{}
		for _, it := range req.Fnb {
			i := slices.IndexFunc(fnbMenu, func(m fnbItem) bool { return m.Code == it.Code })
			if i < 0 || it.Qty < 0 || it.Qty > 10 {
				platform.Err(w, 400, "BAD_ITEM", "Invalid food item")
				return
			}
			if it.Qty > 0 {
				t.Fnb = append(t.Fnb, FnbLine{Code: it.Code, Name: fnbMenu[i].Name, Price: float64(fnbMenu[i].Price), Qty: it.Qty})
			}
		}
	}
	if req.Donation != nil {
		t.Donation = *req.Donation
	}
	if err := s.saveTxn(r.Context(), t); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 200, s.txnView(t))
}

func (s *Server) release(ctx context.Context, t *Txn) {
	if t.Kind == "MOVIE" {
		_ = releaseSeats(ctx, s.rdb, t.SessionID, t.ID, t.Seats)
	} else {
		_ = releaseHold(ctx, s.rdb, t.EventCode, t.Tier, t.ID)
	}
	_ = s.rdb.Del(ctx, txnKey(t.ID)).Err()
}

func (s *Server) cancelTxn(w http.ResponseWriter, r *http.Request) {
	t, err := s.loadTxn(r.Context(), chi.URLParam(r, "id"))
	if err == nil {
		s.release(r.Context(), t)
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- payment & booking -----------------------------------------------------

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	phoneRe = regexp.MustCompile(`^[6-9]\d{9}$`)
	upiRe   = regexp.MustCompile(`^[\w.\-]{2,}@[a-zA-Z]{2,}$`)
)

type payRequest struct {
	Method string `json:"method"` // CARD | UPI | NETBANKING | WALLET
	Card   struct {
		Number string `json:"number"`
		Name   string `json:"name"`
		Expiry string `json:"expiry"`
		CVV    string `json:"cvv"`
	} `json:"card"`
	UPI   string `json:"upi"`
	Bank  string `json:"bank"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// simulateGateway stands in for the payment aggregator. Test card numbers
// ending in 0002 and the UPI handle fail@upi are declined.
func simulateGateway(req payRequest) (string, *platform.APIError) {
	switch req.Method {
	case "CARD":
		num := strings.ReplaceAll(req.Card.Number, " ", "")
		if len(num) < 12 || len(num) > 19 || len(req.Card.CVV) < 3 || !regexp.MustCompile(`^\d{2}/\d{2}$`).MatchString(req.Card.Expiry) {
			return "", &platform.APIError{Code: "INVALID_CARD", Message: "Please enter valid card details"}
		}
		if strings.HasSuffix(num, "0002") {
			return "", &platform.APIError{Code: "PAYMENT_DECLINED", Message: "Your bank declined this transaction. Please try another payment method."}
		}
	case "UPI":
		if !upiRe.MatchString(req.UPI) {
			return "", &platform.APIError{Code: "INVALID_UPI", Message: "Please enter a valid UPI ID"}
		}
		if req.UPI == "fail@upi" {
			return "", &platform.APIError{Code: "PAYMENT_DECLINED", Message: "UPI request was declined by the payer."}
		}
	case "NETBANKING", "WALLET":
	default:
		return "", &platform.APIError{Code: "INVALID_METHOD", Message: "Choose a payment option"}
	}
	return newID("PAY", 12), nil
}

func (s *Server) pay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := platform.User(ctx)
	t := s.txnFromRequest(w, r)
	if t == nil {
		return
	}
	var req payRequest
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	if req.Email == "" {
		req.Email = user.Email
	}
	if req.Phone == "" {
		req.Phone = user.Phone
	}
	if !emailRe.MatchString(req.Email) {
		platform.Err(w, 400, "INVALID_EMAIL", "Please enter a valid email address")
		return
	}
	if req.Phone != "" && !phoneRe.MatchString(req.Phone) {
		platform.Err(w, 400, "INVALID_PHONE", "Please enter a valid 10-digit mobile number")
		return
	}

	// Make sure our hold is still alive before charging.
	var alive bool
	var err error
	if t.Kind == "MOVIE" {
		alive, err = verifySeats(ctx, s.rdb, t.SessionID, t.ID, t.Seats)
	} else {
		alive, err = verifyHold(ctx, s.rdb, t.EventCode, t.Tier, t.ID)
	}
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if !alive {
		platform.Err(w, 410, "TXN_EXPIRED", "Your session has expired. The seats you selected have been released.")
		return
	}

	payRef, perr := simulateGateway(req)
	if perr != nil {
		platform.JSON(w, 402, map[string]any{"error": perr})
		return
	}

	pr := t.Pricing()
	snap, _ := json.Marshal(t.Snapshot)
	fnb, _ := json.Marshal(t.Fnb)
	bookingID := newID("", 7)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	defer tx.Rollback(ctx)
	var session, event, tier any
	if t.Kind == "MOVIE" {
		session = t.SessionID
	} else {
		event, tier = t.EventCode, t.Tier
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO bookings (id, user_id, kind, session_id, event_code, tier, seats, qty, ticket_amount, conv_fee, gst,
			fnb_amount, donation, total, payment_method, payment_ref, status, contact_email, contact_phone, snapshot, fnb)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,'CONFIRMED',$17,$18,$19,$20)`,
		bookingID, user.UserID, t.Kind, session, event, tier, append([]string{}, t.Seats...), t.Qty, pr.Ticket, pr.ConvBase, pr.GST,
		pr.Fnb, pr.Donation, pr.Total, req.Method, payRef, req.Email, req.Phone, snap, fnb); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if t.Kind == "MOVIE" {
		batch := &pgx.Batch{}
		for _, seat := range t.Seats {
			batch.Queue(`INSERT INTO booking_seats (session_id, seat_id, booking_id) VALUES ($1,$2,$3)`, t.SessionID, seat, bookingID)
		}
		if err := tx.SendBatch(ctx, batch).Close(); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				s.release(ctx, t)
				platform.Err(w, 409, "SEATS_UNAVAILABLE", "One of your seats was booked by someone else. Any amount debited will be refunded within 5-7 working days.")
				return
			}
			platform.Err(w, 500, "INTERNAL", err.Error())
			return
		}
	} else {
		if _, err := tx.Exec(ctx, `INSERT INTO event_sales (booking_id, event_code, tier, qty) VALUES ($1,$2,$3,$4)`, bookingID, t.EventCode, t.Tier, t.Qty); err != nil {
			platform.Err(w, 500, "INTERNAL", err.Error())
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	s.release(ctx, t)

	b, err := s.fetchBooking(ctx, bookingID, user.UserID)
	if err == nil {
		s.pub.Publish(ctx, platform.TopicBookingConfirmed, bookingID, b)
	}
	platform.NoStore(w)
	platform.JSON(w, 201, map[string]any{"bookingId": bookingID, "booking": b})
}

type Booking struct {
	ID            string          `json:"id"`
	Kind          string          `json:"kind"`
	Status        string          `json:"status"`
	Seats         []string        `json:"seats"`
	Qty           int             `json:"qty"`
	Snapshot      Snapshot        `json:"snapshot"`
	Fnb           []FnbLine       `json:"fnb"`
	Pricing       Pricing         `json:"pricing"`
	PaymentMethod string          `json:"paymentMethod"`
	PaymentRef    string          `json:"paymentRef"`
	ContactEmail  string          `json:"contactEmail"`
	ContactPhone  string          `json:"contactPhone"`
	CreatedAt     time.Time       `json:"createdAt"`
	Cancellable   bool            `json:"cancellable"`
	Raw           json.RawMessage `json:"-"`
}

const bookingCols = `id, kind, status, seats, qty, snapshot, fnb, ticket_amount::float8, conv_fee::float8, gst::float8,
	fnb_amount::float8, donation::float8, total::float8, payment_method, payment_ref, contact_email, contact_phone, created_at`

func scanBooking(r pgx.CollectableRow) (Booking, error) {
	var b Booking
	err := r.Scan(&b.ID, &b.Kind, &b.Status, &b.Seats, &b.Qty, &b.Snapshot, &b.Fnb, &b.Pricing.Ticket, &b.Pricing.ConvBase,
		&b.Pricing.GST, &b.Pricing.Fnb, &b.Pricing.Donation, &b.Pricing.Total, &b.PaymentMethod, &b.PaymentRef,
		&b.ContactEmail, &b.ContactPhone, &b.CreatedAt)
	b.Pricing.ConvFee = round2(b.Pricing.ConvBase + b.Pricing.GST)
	b.Pricing.SubTotal = round2(b.Pricing.Ticket + b.Pricing.ConvFee)
	b.Cancellable = b.Status == "CONFIRMED" && b.Kind == "MOVIE" && b.Snapshot.Cancellation &&
		time.Until(b.Snapshot.StartsAt) > cancelBefore
	return b, err
}

func (s *Server) fetchBooking(ctx context.Context, id, userID string) (Booking, error) {
	rows, err := s.db.Query(ctx, `SELECT `+bookingCols+` FROM bookings WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return Booking{}, err
	}
	return pgx.CollectExactlyOneRow(rows, scanBooking)
}

func (s *Server) listBookings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := s.db.Query(ctx, `SELECT `+bookingCols+` FROM bookings WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`,
		platform.User(ctx).UserID)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	list, err := pgx.CollectRows(rows, scanBooking)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{"bookings": list})
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request) {
	b, err := s.fetchBooking(r.Context(), chi.URLParam(r, "id"), platform.User(r.Context()).UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Err(w, 404, "NOT_FOUND", "Booking not found")
		return
	} else if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 200, b)
}

func (s *Server) cancelBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := platform.User(ctx)
	b, err := s.fetchBooking(ctx, chi.URLParam(r, "id"), user.UserID)
	if err != nil {
		platform.Err(w, 404, "NOT_FOUND", "Booking not found")
		return
	}
	if !b.Cancellable {
		platform.Err(w, 409, "NOT_CANCELLABLE", "This booking can no longer be cancelled")
		return
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE bookings SET status = 'CANCELLED' WHERE id = $1 AND status = 'CONFIRMED'`, b.ID)
	if err != nil || tag.RowsAffected() != 1 {
		platform.Err(w, 409, "NOT_CANCELLABLE", "This booking can no longer be cancelled")
		return
	}
	if _, err := tx.Exec(ctx, `DELETE FROM booking_seats WHERE booking_id = $1`, b.ID); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if err := tx.Commit(ctx); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	b.Status = "CANCELLED"
	b.Cancellable = false
	refund := round2(b.Pricing.Ticket + b.Pricing.Fnb)
	s.pub.Publish(ctx, platform.TopicBookingCancelled, b.ID, map[string]any{"booking": b, "refund": refund})
	platform.JSON(w, 200, map[string]any{"booking": b, "refund": refund})
}
