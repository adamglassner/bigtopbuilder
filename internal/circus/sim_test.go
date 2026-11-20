package circus

import "testing"

func build(t *testing.T, s *Sim, id string, x, y int) *Attraction {
	t.Helper()
	a, err := s.Place(DefByID(id), x, y)
	if err != nil {
		t.Fatalf("place %s at %d,%d: %v", id, x, y, err)
	}
	return a
}

func hire(t *testing.T, s *Sim, act string) *Performer {
	t.Helper()
	p := s.newPerformer(ActByID(act), s.Rand)
	s.Hireable = append(s.Hireable, p)
	if err := s.Hire(p); err != nil {
		t.Fatalf("hire %s: %v", act, err)
	}
	return p
}

func TestStartingTroupe(t *testing.T) {
	s := New(1)
	if len(s.Troupe) != 2 || len(s.Hireable) != ForHire {
		t.Fatalf("troupe %d, for hire %d", len(s.Troupe), len(s.Hireable))
	}
	if s.Cash != StartCash {
		t.Fatal("cash")
	}
}

func TestNeedsAShow(t *testing.T) {
	s := New(1)
	if err := s.OpenGates(); err != ErrNoShow {
		t.Fatalf("want ErrNoShow, got %v", err)
	}
	a := build(t, s, "ringtent", 5, 3)
	if err := s.OpenGates(); err != ErrNoShow {
		t.Fatalf("empty tent should not count, got %v", err)
	}
	if err := s.Book(s.Troupe[0], a, 0); err != nil {
		t.Fatal(err)
	}
	if err := s.OpenGates(); err != nil {
		t.Fatal(err)
	}
}

func TestTrapezeNeedsBigTop(t *testing.T) {
	s := New(2)
	s.Cash = 1 << 20
	ring := build(t, s, "ringtent", 5, 3)
	big := build(t, s, "bigtop", 12, 3)
	tr := hire(t, s, "trapeze")
	if err := s.Book(tr, ring, 0); err != ErrTooTall {
		t.Fatalf("want ErrTooTall, got %v", err)
	}
	if err := s.Book(tr, big, 4); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Place(DefByID("bigtop"), 3, 9); err != ErrOnlyOne {
		t.Fatalf("second big top: %v", err)
	}
}

func TestBookingMovesPerformer(t *testing.T) {
	s := New(3)
	s.Cash = 1 << 20
	a := build(t, s, "ringtent", 3, 2)
	b := build(t, s, "ringtent", 10, 2)
	p := s.Troupe[0]
	s.Book(p, a, 0)
	s.Book(p, b, 2)
	if a.Lineup[0] != nil || b.Lineup[2] != p || p.Tent != b {
		t.Fatal("performer should move between tents")
	}
	s.Sell(b)
	if p.Tent != nil {
		t.Fatal("selling a tent frees its acts")
	}
}

func TestPacing(t *testing.T) {
	s := New(4)
	s.Cash = 1 << 20
	a := build(t, s, "ringtent", 3, 2)
	clown := hire(t, s, "clown")
	clown2 := hire(t, s, "clown")
	fire := hire(t, s, "fire")
	for _, p := range []*Performer{clown, clown2, fire} {
		p.Skill = 1
	}
	s.Book(clown, a, 0)
	s.Book(fire, a, 1)
	s.Book(clown2, a, 2)
	weak := a.Score().Total
	s.MoveSlot(a, 1, 1) // clown, clown, fire: samey middle but a big finish
	s.MoveSlot(a, 0, 0)
	good := a.Score()
	if good.Acts[2].Note != "Big finish!" {
		t.Fatalf("finale note: %q", good.Acts[2].Note)
	}
	if good.Acts[1].Note == "" || good.Acts[1].Good {
		t.Fatalf("two clowns in a row should be marked down: %q", good.Acts[1].Note)
	}
	_ = weak
	tired := *fire
	fire.Energy = 10
	if !fire.Tired() || a.Score().Total >= good.Total {
		t.Fatal("tired acts score lower")
	}
	_ = tired
}

func TestNightsGoOn(t *testing.T) {
	s := New(42)
	ring := build(t, s, "ringtent", 4, 2)
	s.Book(s.Troupe[0], ring, 0)
	s.Book(s.Troupe[1], ring, 1)
	build(t, s, "popcorn", 10, 3)
	build(t, s, "soda", 13, 3)
	build(t, s, "funhouse", 16, 3)
	build(t, s, "lamp", 9, 9)
	build(t, s, "trash", 12, 10)
	for night := 1; night <= 6; night++ {
		if s.Night != night || s.Phase != PhaseBuild {
			t.Fatalf("expected the morning of night %d, got night %d phase %v", night, s.Night, s.Phase)
		}
		if err := s.OpenGates(); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 20000 && s.Phase == PhaseNight; i++ {
			s.Update(0.1)
		}
		if s.Phase != PhaseResults {
			t.Fatalf("night %d never ended", night)
		}
		r := s.Reports[len(s.Reports)-1]
		t.Logf("night %d: %d guests, %d shows avg %.1f stars, flubs %d, profit %d, happy %.0f, rep %.2f, cash %d", night, r.Guests, r.ShowsRun, r.AvgStars(), r.Flubs, r.Profit(), r.AvgHappy, r.RepAfter, s.Cash)
		if r.Guests < 20 || r.ShowsRun < 3 || r.Shows == 0 {
			t.Fatalf("night %d too quiet: %+v", night, r)
		}
		s.Next()
	}
}

func TestBrokeEndsTheRun(t *testing.T) {
	s := New(7)
	ring := build(t, s, "ringtent", 4, 2)
	s.Book(s.Troupe[0], ring, 0)
	s.OpenGates()
	for s.Phase == PhaseNight {
		s.Update(0.1)
	}
	s.Cash = BrokeAt - 1
	s.Next()
	if s.Phase != PhaseBroke {
		t.Fatalf("phase %v", s.Phase)
	}
}

func TestPaths(t *testing.T) {
	s := New(11)
	start := s.Cash
	if err := s.LayPath(5, 5); err != nil {
		t.Fatal(err)
	}
	if s.LayPath(5, 5) != ErrPathHere || s.Cash != start-PathCost {
		t.Fatal("one path per tile")
	}
	if !s.LiftPath(5, 5) || s.Cash != start || s.IsPath(5, 5) {
		t.Fatal("lifting a path refunds it")
	}
	// a building can go over a path, and takes it up
	s.LayPath(10, 4)
	a := build(t, s, "popcorn", 10, 3)
	if s.IsPath(10, 4) || a == nil {
		t.Fatal("building should take up the path")
	}
	if s.LayPath(10, 4) != ErrSpace {
		t.Fatal("no paths under buildings")
	}

	// A path along the line from (2,2) to (2,8) makes it the cheapest way there.
	f := s.flow(s.idx(2, 8))
	grass := f[s.idx(2, 2)]
	for y := 2; y <= 8; y++ {
		s.LayPath(2, y)
	}
	f = s.flow(s.idx(2, 8))
	if f[s.idx(2, 2)] >= grass {
		t.Fatalf("path route should be cheaper: %d vs %d", f[s.idx(2, 2)], grass)
	}
	// from (3,2), stepping left onto the path should be a valid first step
	cur := s.idx(3, 2)
	ok := false
	for _, n := range []int{s.idx(2, 2), s.idx(3, 3)} {
		if f[n]+s.stepCost(n) == f[cur] && n == s.idx(2, 2) {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("guests should head for the path: f(3,2)=%d f(2,2)=%d", f[cur], f[s.idx(2, 2)])
	}
}

func TestFullBigTopIsProfitable(t *testing.T) {
	s := New(9)
	s.Cash = 20000
	big := build(t, s, "bigtop", 6, 2)
	ring := build(t, s, "ringtent", 13, 2)
	for i, act := range []string{"stilts", "juggler", "acrobats", "strong", "trapeze"} {
		s.Book(hire(t, s, act), big, i)
	}
	s.Book(s.Troupe[0], ring, 0)
	s.Book(s.Troupe[1], ring, 1)
	s.Book(hire(t, s, "fire"), ring, 2)
	for i, id := range []string{"popcorn", "hotdogs", "soda", "cottoncandy"} {
		build(t, s, id, 3+i*3, 10)
	}
	build(t, s, "mirrors", 18, 3)
	start := s.Cash
	s.OpenGates()
	for s.Phase == PhaseNight {
		s.Update(0.1)
	}
	r := s.Reports[0]
	t.Logf("%d guests, shows %d (%.1f stars), income %d, costs %d, profit %d", r.Guests, r.ShowsRun, r.AvgStars(), r.Income(), r.Costs(), r.Profit())
	if s.Cash-start <= -r.Hiring || r.AvgStars() < 3 {
		t.Fatalf("a full bill should pay for its night: cash %d -> %d", start, s.Cash)
	}
}

func TestSaveAndRestore(t *testing.T) {
	s := New(5)
	s.Cash = 50000
	big := build(t, s, "threering", 6, 2)
	build(t, s, "cannon", 14, 3)
	build(t, s, "statue", 3, 9)
	build(t, s, "flowers", 4, 9)
	s.LayPath(2, 7)
	s.LayPath(3, 7)
	tr := hire(t, s, "trapeze")
	s.Book(tr, big, 6)
	s.Book(s.Troupe[0], big, 0)
	if err := s.HireSweeper(); err != nil {
		t.Fatal(err)
	}
	s.BuyPosters()
	b, err := s.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if n, _, ok := SavedNight(b); !ok || n != 1 {
		t.Fatal("peek")
	}
	r, err := Restore(b)
	if err != nil {
		t.Fatal(err)
	}
	if r.Cash != s.Cash || len(r.Attr) != 4 || len(r.Troupe) != 3 || len(r.Crew) != 1 || !r.Posters || !r.IsPath(3, 7) {
		t.Fatalf("restored: cash %d attr %d troupe %d crew %d", r.Cash, len(r.Attr), len(r.Troupe), len(r.Crew))
	}
	var rb *Attraction
	for _, a := range r.Attr {
		if a.Def.ID == "threering" {
			rb = a
		}
	}
	if rb == nil || rb.Lineup[6] == nil || rb.Lineup[6].Name != tr.Name || rb.Lineup[0] == nil || rb.Booked() != 2 {
		t.Fatal("lineup not restored")
	}
	if _, err := r.Place(DefByID("cannon"), 1, 1); err != ErrOnlyOne {
		t.Fatalf("restored layout should block a second cannon: %v", err)
	}
	// and it still plays
	r.OpenGates()
	for r.Phase == PhaseNight {
		r.Update(0.1)
	}
	if r.Reports[0].Guests == 0 {
		t.Fatal("no guests after restore")
	}
}

func TestCrewSweepsLitter(t *testing.T) {
	run := func(crew int) int {
		s := New(21)
		ring := build(t, s, "ringtent", 4, 2)
		s.Book(s.Troupe[0], ring, 0)
		s.Book(s.Troupe[1], ring, 1)
		build(t, s, "popcorn", 10, 3)
		build(t, s, "hotdogs", 13, 3)
		for i := 0; i < crew; i++ {
			s.HireSweeper()
		}
		s.OpenGates()
		for s.Phase == PhaseNight {
			s.Update(0.1)
		}
		r := s.Reports[0]
		t.Logf("crew %d: litter left %d, swept %d, happy %.0f", crew, r.Litter, r.Swept, r.AvgHappy)
		return r.Litter
	}
	if a, b := run(0), run(2); b >= a || b > 10 {
		t.Fatalf("sweepers should clear litter: %d without, %d with", a, b)
	}
}

func TestPostersAndPractice(t *testing.T) {
	night := func(posters bool) (int, *Performer) {
		s := New(33)
		ring := build(t, s, "ringtent", 4, 2)
		s.Book(s.Troupe[0], ring, 0)
		s.Book(s.Troupe[1], ring, 1)
		if posters {
			if err := s.BuyPosters(); err != nil {
				t.Fatal(err)
			}
			if s.BuyPosters() != ErrPosters {
				t.Fatal("posters twice")
			}
		}
		skill := s.Troupe[0].Skill
		s.OpenGates()
		for s.Phase == PhaseNight {
			s.Update(0.1)
		}
		if s.Troupe[0].Skill <= skill {
			t.Fatal("acts should improve with practice")
		}
		return s.Reports[0].Guests, s.Troupe[0]
	}
	a, _ := night(false)
	b, _ := night(true)
	t.Logf("guests: %d without posters, %d with", a, b)
	if b <= a {
		t.Fatal("posters should bring a bigger crowd")
	}
}

func TestDecorationsCheerGuests(t *testing.T) {
	run := func(decor bool) float64 {
		s := New(44)
		s.Cash = 20000
		ring := build(t, s, "ringtent", 4, 2)
		s.Book(s.Troupe[0], ring, 0)
		s.Book(s.Troupe[1], ring, 1)
		if decor {
			for _, d := range []struct {
				id   string
				x, y int
			}{{"statue", 3, 8}, {"flowers", 5, 8}, {"flowers", 2, 5}, {"arch", 8, 8}} {
				build(t, s, d.id, d.x, d.y)
			}
		}
		s.OpenGates()
		for s.Phase == PhaseNight {
			s.Update(0.1)
		}
		return s.Reports[0].AvgHappy
	}
	a, b := run(false), run(true)
	t.Logf("happiness %.1f without decorations, %.1f with", a, b)
	if b <= a {
		t.Fatal("decorations should cheer guests up")
	}
}
