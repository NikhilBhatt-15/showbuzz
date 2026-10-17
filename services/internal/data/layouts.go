package data

import (
	"strconv"
	"sync"
)

// Seat is one bookable seat. C is its column on the auditorium grid, so gaps
// between C values render as aisles. Seat IDs are row label + number ("J12"),
// unique within a screen.
type Seat struct {
	ID   string `json:"id"`
	N    int    `json:"n"`
	C    int    `json:"c"`
	Best bool   `json:"best,omitempty"`
}

type Row struct {
	Label string `json:"label"`
	Seats []Seat `json:"seats"`
}

// Area is a price category ("PREMIUM", "RECLINER"...). Areas are listed from
// the back of the auditorium to the front, the way the seat map is drawn.
type Area struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Rows []Row  `json:"rows"`
}

type Layout struct {
	Template string `json:"template"`
	Cols     int    `json:"cols"`
	Areas    []Area `json:"areas"`
}

// block is a run of adjacent seats: starting column and seat count.
type block struct{ col, count int }

// row numbers seats right-to-left (rightmost seat is 1), matching how
// BookMyShow numbers most Indian multiplex rows.
func row(label string, blocks []block, best func(col int) bool) Row {
	total := 0
	for _, b := range blocks {
		total += b.count
	}
	r := Row{Label: label}
	n := total
	for _, b := range blocks {
		for i := 0; i < b.count; i++ {
			c := b.col + i
			r.Seats = append(r.Seats, Seat{ID: label + strconv.Itoa(n), N: n, C: c, Best: best != nil && best(c)})
			n--
		}
	}
	return r
}

func rows(labels string, blocks []block, best func(col int) bool) []Row {
	out := make([]Row, 0, len(labels))
	for _, l := range labels {
		out = append(out, row(string(l), blocks, best))
	}
	return out
}

func classic() Layout {
	premium := []block{{0, 5}, {7, 11}, {20, 3}}
	mid := func(c int) bool { return c >= 10 && c <= 14 }
	return Layout{Template: "classic", Cols: 23, Areas: []Area{
		{Code: "RC", Name: "RECLINER", Rows: []Row{row("K", []block{{0, 2}, {3, 2}, {7, 2}, {10, 2}, {13, 2}, {16, 2}, {21, 2}}, nil)}},
		{Code: "PR", Name: "PREMIUM", Rows: append(rows("J", premium, nil), rows("IHGF", premium, mid)...)},
		{Code: "EX", Name: "EXECUTIVE", Rows: rows("EDC", []block{{0, 5}, {7, 11}}, nil)},
		{Code: "NM", Name: "NORMAL", Rows: rows("BA", []block{{0, 5}, {7, 11}}, nil)},
	}}
}

func compact() Layout {
	b := []block{{0, 4}, {6, 10}, {18, 4}}
	return Layout{Template: "compact", Cols: 22, Areas: []Area{
		{Code: "GD", Name: "GOLD", Rows: rows("JIHG", b, func(c int) bool { return c >= 9 && c <= 12 })},
		{Code: "SL", Name: "SILVER", Rows: rows("FEDCBA", b, nil)},
	}}
}

func large() Layout {
	wide := []block{{0, 6}, {8, 14}, {24, 6}}
	return Layout{Template: "large", Cols: 30, Areas: []Area{
		{Code: "PL", Name: "PLATINUM", Rows: rows("NM", []block{{2, 4}, {8, 14}, {24, 4}}, nil)},
		{Code: "GD", Name: "GOLD", Rows: rows("LKJIH", wide, func(c int) bool { return c >= 12 && c <= 17 })},
		{Code: "SL", Name: "SILVER", Rows: rows("GFEDCBA", wide, nil)},
	}}
}

var (
	layoutsOnce sync.Once
	layouts     map[string]Layout
	seatIndex   map[string]map[string]string // template -> seat ID -> area code
)

func buildLayouts() {
	layouts = map[string]Layout{"classic": classic(), "compact": compact(), "large": large()}
	seatIndex = map[string]map[string]string{}
	for name, l := range layouts {
		idx := map[string]string{}
		for _, a := range l.Areas {
			for _, r := range a.Rows {
				for _, s := range r.Seats {
					idx[s.ID] = a.Code
				}
			}
		}
		seatIndex[name] = idx
	}
}

// GetLayout returns the seat layout for a template name.
func GetLayout(template string) (Layout, bool) {
	layoutsOnce.Do(buildLayouts)
	l, ok := layouts[template]
	return l, ok
}

// SeatArea maps a seat ID to its area code; ok is false for unknown seats.
func SeatArea(template, seatID string) (string, bool) {
	layoutsOnce.Do(buildLayouts)
	a, ok := seatIndex[template][seatID]
	return a, ok
}

// SeatCount is the number of seats in a template.
func SeatCount(template string) int {
	layoutsOnce.Do(buildLayouts)
	return len(seatIndex[template])
}

// AreaSeatCounts returns seats per area code for a template.
func AreaSeatCounts(template string) map[string]int {
	layoutsOnce.Do(buildLayouts)
	out := map[string]int{}
	for _, a := range seatIndex[template] {
		out[a]++
	}
	return out
}
