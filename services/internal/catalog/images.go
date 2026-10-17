package catalog

import (
	"fmt"
	"hash/fnv"
	"html"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/data"
	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

// The image service plays the role of BookMyShow's ImageKit CDN: artwork is
// rendered on request and overlays (rating bar, date bar) are baked into the
// image itself, like BMS's `l-image` overlay transforms.

var (
	movieByCode = map[string]data.Movie{}
	eventByCode = map[string]data.Event{}
)

func init() {
	for _, m := range data.Movies {
		movieByCode[m.Code] = m
	}
	for _, e := range data.Events {
		eventByCode[e.Code] = e
	}
}

func seeded(s string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(s))
	v := h.Sum64()
	return rand.New(rand.NewPCG(v, v>>1))
}

func esc(s string) string { return html.EscapeString(s) }

func writeSVG(w http.ResponseWriter, svg string) {
	w.Header().Set("Content-Type", "image/svg+xml")
	platform.Cache(w, 86400)
	_, _ = w.Write([]byte(svg))
}

// wrap splits title into at most maxLines lines of roughly width chars.
func wrap(title string, width, maxLines int) []string {
	words := strings.Fields(title)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
		} else if len(cur)+1+len(w) <= width {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) > maxLines {
		lines = append(lines[:maxLines-1], strings.Join(lines[maxLines-1:], " "))
	}
	return lines
}

func compact(n int) string {
	switch {
	case n >= 1000:
		v := float64(n) / 1000
		if v >= 100 {
			return fmt.Sprintf("%.0fK", v)
		}
		return strings.TrimSuffix(fmt.Sprintf("%.1f", v), ".0") + "K"
	}
	return strconv.Itoa(n)
}

const starPath = `M12 2l2.9 6.9 7.1.6-5.4 4.7 1.7 7-6.3-3.9-6.3 3.9 1.7-7L2 9.5l7.1-.6z`
const thumbPath = `M2 21h4V9H2v12zm20-11c0-1.1-.9-2-2-2h-6.3l1-4.6v-.3c0-.4-.2-.8-.4-1.1L13.2 1 6.6 7.6C6.2 7.9 6 8.4 6 9v10c0 1.1.9 2 2 2h9c.8 0 1.5-.5 1.8-1.2l3-7.1c.1-.2.2-.5.2-.7v-2z`

func ratingBar(m data.Movie, y, w int) string {
	if m.Rating > 0 {
		return fmt.Sprintf(`<rect x="0" y="%d" width="%d" height="64" fill="#000"/>
<path d="%s" transform="translate(20 %d) scale(1.25)" fill="#F84464"/>
<text x="62" y="%d" font-family="Roboto,Arial,sans-serif" font-size="28" fill="#fff">%.1f/10  %s Votes</text>`,
			y, w, starPath, y+17, y+42, m.Rating, compact(m.Votes)+"+")
	}
	return fmt.Sprintf(`<rect x="0" y="%d" width="%d" height="64" fill="#000"/>
<path d="%s" transform="translate(20 %d) scale(1.25)" fill="#1EA83C"/>
<text x="62" y="%d" font-family="Roboto,Arial,sans-serif" font-size="28" fill="#fff">%s Likes</text>`,
		y, w, thumbPath, y+17, y+42, compact(m.Likes)+"+")
}

func dateBar(label string, y, w int) string {
	return fmt.Sprintf(`<rect x="0" y="%d" width="%d" height="64" fill="#000"/>
<text x="20" y="%d" font-family="Roboto,Arial,sans-serif" font-size="28" fill="#fff">%s</text>`, y, w, y+42, esc(label))
}

// motif draws seeded decorative shapes behind a poster title.
func motif(r *rand.Rand, hue int, w, h int) string {
	var b strings.Builder
	cx := 60 + r.IntN(w-120)
	cy := 90 + r.IntN(h/3)
	rad := 70 + r.IntN(70)
	fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="hsl(%d,85%%,62%%)" opacity=".55"/>`, cx, cy, rad, (hue+40)%360)
	fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="none" stroke="hsl(%d,90%%,75%%)" stroke-width="2" opacity=".35"/>`, cx, cy, rad+24, (hue+40)%360)
	for i := 0; i < 5; i++ {
		x := r.IntN(w)
		fmt.Fprintf(&b, `<polygon points="%d,0 %d,0 %d,%d" fill="#fff" opacity=".05"/>`, x, x+30+r.IntN(40), x+r.IntN(160)-80, h)
	}
	// Skyline / ridge silhouette.
	base := h*55/100 + r.IntN(h/10)
	pts := fmt.Sprintf("0,%d ", h)
	x := 0
	for x < w {
		y := base - r.IntN(h/5)
		pts += fmt.Sprintf("%d,%d %d,%d ", x, y, x+18+r.IntN(30), y)
		x += 18 + r.IntN(40)
	}
	pts += fmt.Sprintf("%d,%d", w, h)
	fmt.Fprintf(&b, `<polygon points="%s" fill="hsl(%d,40%%,6%%)" opacity=".85"/>`, pts, hue)
	// A figure silhouette.
	fx := w/2 + r.IntN(w/3) - w/6
	fy := base - 40
	fmt.Fprintf(&b, `<g fill="hsl(%d,30%%,4%%)"><circle cx="%d" cy="%d" r="16"/><path d="M%d %d q24 -10 48 0 l10 120 h-68z"/></g>`,
		hue, fx, fy-70, fx-24, fy-50)
	return b.String()
}

func posterSVG(title, tagline, credit string, hue int, overlay string) string {
	const w, h = 400, 660
	r := seeded(title)
	lines := wrap(strings.ToUpper(title), 11, 3)
	size := 58
	for _, l := range lines {
		if len(l) > 9 {
			size = 46
		}
	}
	var t strings.Builder
	y := 480 - (len(lines)-1)*size
	for i, l := range lines {
		fmt.Fprintf(&t, `<text x="200" y="%d" text-anchor="middle" font-family="Impact,'Arial Narrow',Oswald,sans-serif" font-weight="900" font-size="%d" letter-spacing="1" fill="#fff" stroke="hsl(%d,70%%,20%%)" stroke-width="1">%s</text>`,
			y+i*size, size, hue, esc(l))
	}
	bar := ""
	if overlay != "" {
		bar = overlay
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" preserveAspectRatio="xMidYMid slice">
<defs><linearGradient id="g" x1="0" y1="0" x2="0" y2="1">
<stop offset="0" stop-color="hsl(%d,65%%,38%%)"/><stop offset=".6" stop-color="hsl(%d,60%%,16%%)"/><stop offset="1" stop-color="hsl(%d,50%%,6%%)"/>
</linearGradient></defs>
<rect width="%d" height="%d" fill="url(#g)"/>%s
<text x="200" y="44" text-anchor="middle" font-family="Georgia,serif" font-style="italic" font-size="14" fill="#fff" opacity=".85">%s</text>
%s
<text x="200" y="%d" text-anchor="middle" font-family="'Arial Narrow',Arial,sans-serif" font-size="10" letter-spacing="2" fill="#fff" opacity=".6">%s</text>
%s</svg>`, w, h, hue, (hue+25)%360, (hue+40)%360, w, h, motif(r, hue, w, h), esc(tagline), t.String(), 565, esc(strings.ToUpper(credit)), bar)
}

func (s *Server) poster(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(chi.URLParam(r, "code"), ".svg")
	overlay := r.URL.Query().Get("overlay")
	if m, ok := movieByCode[code]; ok {
		credit := ""
		if len(m.Crew) > 0 {
			credit = "a " + m.Crew[0].Name + " film"
		}
		bar := ""
		switch overlay {
		case "rating":
			bar = ratingBar(m, 596, 400)
		case "release":
			bar = dateBar("Releasing "+platform.Today().AddDate(0, 0, m.ReleaseDays).Format("2 Jan"), 596, 400)
		}
		writeSVG(w, posterSVG(m.Title, m.Tagline, credit, m.Hue, bar))
		return
	}
	if e, ok := eventByCode[code]; ok {
		bar := ""
		if overlay == "date" {
			bar = dateBar(platform.Today().AddDate(0, 0, e.DayOffset).Format("Mon, 2 Jan")+" onwards", 596, 400)
		}
		writeSVG(w, posterSVG(e.Title, e.Genre, e.VenueName, e.Hue, bar))
		return
	}
	http.NotFound(w, r)
}

// landscape renders wide artwork: home banners, movie backdrops, event heroes.
func landscapeSVG(w, h int, title, tagline, cta string, hue int, seed string) string {
	r := seeded(seed)
	var b strings.Builder
	for i := 0; i < 7; i++ {
		x := w/2 + r.IntN(w/2)
		fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="hsl(%d,80%%,60%%)" opacity=".%d"/>`, x, r.IntN(h), 30+r.IntN(h/2), (hue+r.IntN(60))%360, 10+r.IntN(25))
	}
	for i := 0; i < 6; i++ {
		x := w/3 + r.IntN(w*2/3)
		fmt.Fprintf(&b, `<polygon points="%d,0 %d,0 %d,%d" fill="#fff" opacity=".04"/>`, x, x+40, x-120+r.IntN(240), h)
	}
	lines := wrap(title, 18, 2)
	size := h / 6
	var t strings.Builder
	ty := h/2 - (len(lines)-1)*size/2
	for i, l := range lines {
		fmt.Fprintf(&t, `<text x="%d" y="%d" font-family="Impact,'Arial Narrow',sans-serif" font-size="%d" fill="#fff">%s</text>`, w/14, ty+i*size, size, esc(l))
	}
	ctaSVG := ""
	if cta != "" {
		cy := ty + len(lines)*size - size/3
		fmt.Fprintf(&t, `<text x="%d" y="%d" font-family="Roboto,Arial,sans-serif" font-size="%d" fill="#fff" opacity=".9">%s</text>`, w/14, cy, h/16, esc(tagline))
		ctaSVG = fmt.Sprintf(`<rect x="%d" y="%d" rx="6" width="%d" height="%d" fill="#fff"/><text x="%d" y="%d" text-anchor="middle" font-family="Roboto,Arial,sans-serif" font-weight="700" font-size="%d" fill="#DC3558">%s</text>`,
			w/14, cy+h/14, h*2/3, h/7, w/14+h/3, cy+h/14+h/10, h/18, esc(cta))
	} else if tagline != "" {
		fmt.Fprintf(&t, `<text x="%d" y="%d" font-family="Roboto,Arial,sans-serif" font-size="%d" fill="#fff" opacity=".85">%s</text>`, w/14, ty+len(lines)*size-size/4, h/15, esc(tagline))
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" preserveAspectRatio="xMidYMid slice">
<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="0">
<stop offset="0" stop-color="hsl(%d,60%%,10%%)"/><stop offset=".55" stop-color="hsl(%d,60%%,28%%)"/><stop offset="1" stop-color="hsl(%d,70%%,42%%)"/>
</linearGradient></defs>
<rect width="%d" height="%d" fill="url(#g)"/>%s%s%s</svg>`, w, h, hue, (hue+20)%360, (hue+45)%360, w, h, b.String(), t.String(), ctaSVG)
}

// Banners are the home page hero carousel (promotion-serving).
type banner struct {
	ID, Title, Tagline, CTA, URL string
	Hue                          int
}

func banners(regionSlug string) []banner {
	return []banner{
		{"tidebreaker", "Tidebreaker", "Now showing in IMAX & 4DX", "Book Now", "/movies/" + regionSlug + "/tidebreaker/ET00510001", 200},
		{"ravi-kapoor", "Ravi Kapoor: Unfiltered", "Live on tour · Limited seats", "Book Now", "/events/ravi-kapoor-unfiltered-tour/ET00520001", 40},
		{"iron-monsoon", "Iron Monsoon", "The storm is here. Experience it in 3D", "Book Now", "/movies/" + regionSlug + "/iron-monsoon/ET00510013", 210},
		{"offer", "Flat 50% Off", "On your first booking with SB Pay · Use code FIRST50", "Know More", "/offers", 345},
		{"dandiya", "Dandiya Dhamaka", "Nine nights of garba & dandiya", "Book Now", "/events/dandiya-dhamaka/ET00520015", 330},
	}
}

func (s *Server) bannerImage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSuffix(chi.URLParam(r, "id"), ".svg")
	for _, b := range banners("") {
		if b.ID == id {
			writeSVG(w, landscapeSVG(1240, 300, b.Title, b.Tagline, b.CTA, b.Hue, b.ID))
			return
		}
	}
	http.NotFound(w, r)
}

func (s *Server) backdrop(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(chi.URLParam(r, "code"), ".svg")
	if m, ok := movieByCode[code]; ok {
		writeSVG(w, landscapeSVG(1280, 480, "", "", "", m.Hue, m.Code))
		return
	}
	if e, ok := eventByCode[code]; ok {
		writeSVG(w, landscapeSVG(1280, 720, e.Title, e.Genre, "", e.Hue, e.Code))
		return
	}
	http.NotFound(w, r)
}

// Category tiles for "The Best of Live Events".
type tile struct {
	Slug, Name, Emoji, Category, Genre string
	Hue                                int
}

var tiles = []tile{
	{"comedy-shows", "Comedy Shows", "🎤", "EVENTS", "Comedy Shows", 280},
	{"music-shows", "Music Shows", "🎸", "EVENTS", "Music Shows", 330},
	{"theatre", "Theatre", "🎭", "PLAYS", "", 10},
	{"kids", "Kids", "🎈", "ACTIVITIES", "Kids", 40},
	{"workshops", "Workshops", "🏺", "ACTIVITIES", "Workshops", 160},
	{"amusement-park", "Amusement Parks", "🎢", "ACTIVITIES", "Amusement Park", 200},
	{"adventure", "Adventure", "🏔️", "ACTIVITIES", "Adventure", 120},
	{"sports", "Sports", "🏏", "SPORTS", "", 220},
	{"festivals", "Festivals", "🪔", "EVENTS", "Festivals", 25},
}

func (s *Server) tileImage(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSuffix(chi.URLParam(r, "slug"), ".svg")
	count := r.URL.Query().Get("count")
	for _, t := range tiles {
		if t.Slug != slug {
			continue
		}
		sub := ""
		if count != "" {
			sub = count + "+ Events"
		}
		writeSVG(w, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 224 224">
<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="hsl(%d,75%%,55%%)"/><stop offset="1" stop-color="hsl(%d,75%%,35%%)"/></linearGradient></defs>
<rect width="224" height="224" fill="url(#g)"/>
<circle cx="180" cy="40" r="70" fill="#fff" opacity=".08"/>
<text x="20" y="44" font-family="Roboto,Arial,sans-serif" font-weight="900" font-size="22" fill="#fff">%s</text>
<text x="20" y="70" font-family="Roboto,Arial,sans-serif" font-size="13" fill="#fff" opacity=".9">%s</text>
<text x="200" y="200" text-anchor="end" font-size="96">%s</text></svg>`, t.Hue, (t.Hue+30)%360, esc(strings.ToUpper(t.Name)), esc(sub), t.Emoji))
		return
	}
	http.NotFound(w, r)
}

func initials(name string) string {
	var out []rune
	for _, f := range strings.Fields(name) {
		out = append(out, []rune(f)[0])
		if len(out) == 2 {
			break
		}
	}
	return strings.ToUpper(string(out))
}

func (s *Server) avatar(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	h := fnv.New32a()
	h.Write([]byte(name))
	hue := int(h.Sum32() % 360)
	writeSVG(w, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 120 120">
<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="hsl(%d,45%%,62%%)"/><stop offset="1" stop-color="hsl(%d,50%%,38%%)"/></linearGradient></defs>
<rect width="120" height="120" fill="url(#g)"/>
<circle cx="60" cy="48" r="22" fill="#fff" opacity=".25"/><path d="M18 120c4-28 22-40 42-40s38 12 42 40z" fill="#fff" opacity=".25"/>
<text x="60" y="70" text-anchor="middle" font-family="Roboto,Arial,sans-serif" font-weight="700" font-size="34" fill="#fff">%s</text></svg>`, hue, (hue+30)%360, esc(initials(name))))
}

// cityIcon draws a simple line-art skyline unique to each city.
func (s *Server) cityIcon(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSuffix(chi.URLParam(r, "slug"), ".svg")
	rng := seeded(slug)
	var b strings.Builder
	x := 6
	for x < 70 {
		bw := 8 + rng.IntN(8)
		bh := 14 + rng.IntN(26)
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d"/>`, x, 70-bh, bw, bh)
		x += bw + 2
	}
	switch rng.IntN(3) {
	case 0: // dome
		fmt.Fprintf(&b, `<path d="M28 38 a12 12 0 0 1 24 0z"/><line x1="40" y1="20" x2="40" y2="26"/>`)
	case 1: // arch
		fmt.Fprintf(&b, `<path d="M26 70 V34 h28 V70 M32 70 V46 a8 8 0 0 1 16 0 V70"/>`)
	default: // tower
		fmt.Fprintf(&b, `<path d="M36 70 L40 14 L44 70 M34 40 h12 M33 54 h14"/>`)
	}
	writeSVG(w, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 80 80"><g fill="none" stroke="#666" stroke-width="1.5" stroke-linejoin="round">%s<line x1="2" y1="70" x2="78" y2="70"/></g></svg>`, b.String()))
}

func (s *Server) chainLogo(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(chi.URLParam(r, "code"), ".svg")
	name := code
	hue := 0
	for i, c := range data.Chains {
		if c.Code == code {
			name, hue = c.Name, i*70+340
		}
	}
	writeSVG(w, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
<rect width="64" height="64" rx="8" fill="#fff"/><text x="32" y="30" text-anchor="middle" font-family="Georgia,serif" font-style="italic" font-weight="700" font-size="15" fill="hsl(%d,70%%,40%%)">%s</text>
<text x="32" y="46" text-anchor="middle" font-family="Roboto,Arial,sans-serif" font-size="7" letter-spacing="1" fill="#666">%s</text></svg>`,
		hue%360, esc(strings.ToLower(strings.Fields(name)[0])), esc(strings.ToUpper(strings.Join(strings.Fields(name)[1:], " ")))))
}

func (s *Server) stripImage(w http.ResponseWriter, _ *http.Request) {
	writeSVG(w, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1240 120" preserveAspectRatio="xMidYMid slice">
<defs><linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="#1F2533"/><stop offset="1" stop-color="#3B1D3A"/></linearGradient></defs>
<rect width="1240" height="120" fill="url(#g)"/>
<text x="60" y="72" font-family="Impact,'Arial Narrow',sans-serif" font-size="44" letter-spacing="6" fill="#fff">STREAM</text>
<text x="300" y="70" font-family="Roboto,Arial,sans-serif" font-weight="700" font-size="26" fill="#fff">Endless Entertainment Anytime. Anywhere!</text>
<circle cx="1120" cy="60" r="34" fill="#F84464"/><path d="M1110 42 L1136 60 L1110 78z" fill="#fff"/></svg>`)
}

func (s *Server) fnbImage(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSuffix(chi.URLParam(r, "code"), ".svg")
	emoji := map[string]string{"POPS": "🍿", "POPC": "🍿", "COKE": "🥤", "NACH": "🌮", "COMB": "🍿", "COFF": "☕", "PUFF": "🥐", "HOTD": "🌭", "WATR": "💧", "SAMO": "🥟"}[code]
	if emoji == "" {
		emoji = "🍽️"
	}
	h := fnv.New32a()
	h.Write([]byte(code))
	writeSVG(w, fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 160 160"><rect width="160" height="160" rx="12" fill="hsl(%d,80%%,92%%)"/><text x="80" y="108" text-anchor="middle" font-size="76">%s</text></svg>`, h.Sum32()%360, emoji))
}
