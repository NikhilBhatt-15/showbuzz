package data

import "testing"

func TestLayoutsHaveUniqueSeatsAndFitGrid(t *testing.T) {
	for _, name := range []string{"classic", "compact", "large"} {
		l, ok := GetLayout(name)
		if !ok {
			t.Fatalf("missing layout %s", name)
		}
		seen := map[string]bool{}
		for _, a := range l.Areas {
			for _, r := range a.Rows {
				prevC := -1
				for _, s := range r.Seats {
					if seen[s.ID] {
						t.Errorf("%s: duplicate seat %s", name, s.ID)
					}
					seen[s.ID] = true
					if s.C <= prevC || s.C >= l.Cols {
						t.Errorf("%s: seat %s at column %d out of order or outside %d cols", name, s.ID, s.C, l.Cols)
					}
					prevC = s.C
				}
				if r.Seats[len(r.Seats)-1].N != 1 {
					t.Errorf("%s row %s: rightmost seat should be numbered 1", name, r.Label)
				}
			}
		}
		if SeatCount(name) != len(seen) {
			t.Errorf("%s: SeatCount %d != %d", name, SeatCount(name), len(seen))
		}
		if _, ok := SeatArea(name, "ZZ99"); ok {
			t.Errorf("%s: unknown seat resolved", name)
		}
	}
}

func TestClassicLayoutMatchesReference(t *testing.T) {
	// The classic template reproduces the 4-tier auditorium captured from
	// BookMyShow: 14 recliners, 19-seat premium rows, 16-seat executive/normal.
	counts := AreaSeatCounts("classic")
	want := map[string]int{"RC": 14, "PR": 5 * 19, "EX": 3 * 16, "NM": 2 * 16}
	for k, v := range want {
		if counts[k] != v {
			t.Errorf("area %s: got %d seats, want %d", k, counts[k], v)
		}
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{"Kaal Bhairav": "kaal-bhairav", "Tiny Titans: Dhol Ka Raaz": "tiny-titans-dhol-ka-raaz", "Pune 411": "pune-411"}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}
