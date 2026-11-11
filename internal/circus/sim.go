package circus

import (
	"errors"
	"math"
	"math/rand"
)

type Phase int

const (
	PhaseBuild   Phase = iota // daytime: build, hire and plan the shows
	PhaseNight                // showtime: the gates are open
	PhaseResults              // the night's takings
	PhaseBroke                // ran out of money
)

type Weather int

const (
	Clear Weather = iota
	Breezy
	Drizzle
)

var WeatherNames = [...]string{"Clear", "Breezy", "Drizzle"}

// Event tells the screen something happened (for sounds and pop-ups).
type Event struct {
	Kind string // coin, showstart, applause, flub, gate
	Text string
	X, Y float64 // tile coordinates
	A    *Attraction
	V    float64
}

// NightReport is one night's takings.
type NightReport struct {
	Night                int
	Weather              Weather
	Guests               int
	Admissions, Shows    int
	Sideshows, Food      int
	Goods, Wages, Hiring int
	Upkeep, Cleanup      int
	Litter               int
	AvgHappy             float64
	RepBefore, RepAfter  float64
	PeakGuests           int
	ShowsRun, Flubs      int
	StarSum, BestStars   float64
}

func (r NightReport) Income() int { return r.Admissions + r.Shows + r.Sideshows + r.Food }
func (r NightReport) Costs() int  { return r.Goods + r.Wages + r.Upkeep + r.Cleanup + r.Hiring }
func (r NightReport) Profit() int { return r.Income() - r.Costs() }

// AvgStars is the average rating of tonight's shows.
func (r NightReport) AvgStars() float64 {
	if r.ShowsRun == 0 {
		return 0
	}
	return r.StarSum / float64(r.ShowsRun)
}

// Sim is one run of the demo.
type Sim struct {
	Town   Town
	W, H   int
	Rand   *rand.Rand
	Phase  Phase
	Night  int     // 1-based
	T      float64 // fraction of the night gone (1 = closing time)
	Cash   int
	Rep    float64 // 0-5 stars
	Gate   int     // admission price
	Attr   []*Attraction
	Guests []*Guest

	Troupe   []*Performer
	Hireable []*Performer

	Weather Weather

	Report  NightReport
	Reports []NightReport
	Events  []Event

	Path    []bool // tiles with a path laid
	Litter  []uint8
	Traffic []float32 // footsteps per tile; worn grass
	Lit     []float32 // how lit each tile is at night (0-1)

	GateX, GateY int

	grid      []int // attraction ID per tile, 0 = free
	trashNear []bool
	flows     map[int][]int16
	nextID    int
	nextPerf  int
	hiringDay int // money spent hiring today, counted in tonight's report
	spawnAcc  float64
	closing   bool
	happySum  float64
	departed  int
	Version   int // bumps whenever the layout changes
}

// New starts a fresh demo with a clown and a juggler already in the troupe.
func New(seed int64) *Sim {
	t := DemoTown
	s := &Sim{Town: t, W: t.W, H: t.H, Rand: rand.New(rand.NewSource(seed)), Night: 1, Cash: StartCash, Rep: t.Rep, Gate: DefaultGate}
	n := s.W * s.H
	s.grid = make([]int, n)
	s.Path = make([]bool, n)
	s.Litter = make([]uint8, n)
	s.Traffic = make([]float32, n)
	s.Lit = make([]float32, n)
	s.trashNear = make([]bool, n)
	s.GateX, s.GateY = 0, s.H/2
	s.Weather = Clear // a gentle first night
	for _, id := range []string{"clown", "juggler"} {
		p := s.newPerformer(ActByID(id), s.Rand)
		p.Skill = 0.95 + s.Rand.Float64()*0.1
		s.Troupe = append(s.Troupe, p)
	}
	s.refreshHireable()
	s.layoutChanged()
	return s
}

func (s *Sim) idx(x, y int) int       { return y*s.W + x }
func (s *Sim) in(x, y int) bool       { return x >= 0 && y >= 0 && x < s.W && y < s.H }
func (s *Sim) Free(x, y int) bool     { return s.in(x, y) && s.grid[s.idx(x, y)] == 0 }
func (s *Sim) emit(e Event)           { s.Events = append(s.Events, e) }
func (s *Sim) Closing() bool          { return s.closing }
func (s *Sim) reserved(x, y int) bool { return y == s.GateY && x <= 1 }
func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
func (s *Sim) chance(p float64) bool  { return s.Rand.Float64() < p }
func (s *Sim) AttrAt(x, y int) *Attraction {
	if !s.in(x, y) {
		return nil
	}
	id := s.grid[s.idx(x, y)]
	for _, a := range s.Attr {
		if a.ID == id {
			return a
		}
	}
	return nil
}

/* ---------- building ---------- */

var (
	ErrLocked   = errors.New("Coming in the full game")
	ErrCash     = errors.New("Not enough cash")
	ErrSpace    = errors.New("Something's in the way")
	ErrEdge     = errors.New("Off the edge of the grounds")
	ErrBlocks   = errors.New("That would block the way through")
	ErrGateway  = errors.New("Keep the gate clear")
	ErrDoor     = errors.New("Keep entrances clear")
	ErrNoShow   = errors.New("Build a tent and book at least one act first")
	ErrOnlyOne  = errors.New("You can only raise one Big Top")
	ErrPathHere = errors.New("There's already a path here")
)

// CanPlace checks whether def fits with its top-left tile at (x, y).
func (s *Sim) CanPlace(d *Def, x, y int) error {
	if d.Locked {
		return ErrLocked
	}
	if d.ID == "path" {
		switch {
		case !s.in(x, y):
			return ErrEdge
		case !s.Free(x, y):
			return ErrSpace
		case s.Path[s.idx(x, y)]:
			return ErrPathHere
		case s.Cash < d.Cost:
			return ErrCash
		}
		return nil
	}
	if d.ID == "bigtop" {
		for _, a := range s.Attr {
			if a.Def == d {
				return ErrOnlyOne
			}
		}
	}
	if x < 0 || y < 0 || x+d.W > s.W || y+d.H > s.H {
		return ErrEdge
	}
	for ty := y; ty < y+d.H; ty++ {
		for tx := x; tx < x+d.W; tx++ {
			if s.reserved(tx, ty) {
				return ErrGateway
			}
			if !s.Free(tx, ty) {
				return ErrSpace
			}
			for _, a := range s.Attr {
				if ex, ey := a.Entrance(); a.Def.Visitable() && ex == tx && ey == ty {
					return ErrDoor
				}
			}
		}
	}
	if d.Visitable() {
		ex, ey := x+d.W/2, y+d.H
		if !s.Free(ex, ey) {
			return ErrSpace
		}
	}
	// Try it and make sure every entrance can still be reached from the gate.
	s.stamp(d, x, y, -1)
	ok := s.allReachable()
	s.stamp(d, x, y, 0)
	if !ok {
		return ErrBlocks
	}
	if s.Cash < d.Cost {
		return ErrCash
	}
	return nil
}

func (s *Sim) stamp(d *Def, x, y, id int) {
	for ty := y; ty < y+d.H; ty++ {
		for tx := x; tx < x+d.W; tx++ {
			s.grid[s.idx(tx, ty)] = id
		}
	}
}

func (s *Sim) allReachable() bool {
	f := s.bfs(s.idx(s.GateX, s.GateY))
	for _, a := range s.Attr {
		if !a.Def.Visitable() {
			continue
		}
		ex, ey := a.Entrance()
		if !s.in(ex, ey) || s.grid[s.idx(ex, ey)] != 0 || f[s.idx(ex, ey)] < 0 {
			return false
		}
	}
	return true
}

// Place builds def at (x, y).
func (s *Sim) Place(d *Def, x, y int) (*Attraction, error) {
	if d.ID == "path" {
		return nil, s.LayPath(x, y)
	}
	if err := s.CanPlace(d, x, y); err != nil {
		return nil, err
	}
	s.nextID++
	a := &Attraction{ID: s.nextID, Def: d, X: x, Y: y, Price: d.Price, announced: -1}
	if d.Slots > 0 {
		a.Lineup = make([]*Performer, d.Slots)
	}
	s.Attr = append(s.Attr, a)
	s.stamp(d, x, y, a.ID)
	for ty := y; ty < y+d.H; ty++ { // building over a path takes it up
		for tx := x; tx < x+d.W; tx++ {
			s.Path[s.idx(tx, ty)] = false
		}
	}
	s.Cash -= d.Cost
	s.layoutChanged()
	s.unstickGuests()
	return a, nil
}

// Sell removes a building for part of what it cost. Its acts go back to the troupe.
func (s *Sim) Sell(a *Attraction) int {
	back := int(float64(a.Def.Cost) * SellBack)
	for _, g := range append(append([]*Guest{}, a.Queue...), a.Riding...) {
		g.Target = nil
		g.State = Idle
		g.X, g.Y = a.entranceCenter()
		g.cx, g.cy = a.Entrance()
		g.think(s, "Aw, it's closed.", 0.5)
		g.Happy -= 4
	}
	for i := range a.Lineup {
		s.ClearSlot(a, i)
	}
	a.Queue, a.Riding = nil, nil
	s.stamp(a.Def, a.X, a.Y, 0)
	for i, b := range s.Attr {
		if b == a {
			s.Attr = append(s.Attr[:i], s.Attr[i+1:]...)
			break
		}
	}
	for _, g := range s.Guests {
		if g.Target == a {
			g.Target = nil
			g.State = Idle
		}
	}
	s.Cash += back
	s.layoutChanged()
	return back
}

// IsPath reports whether a path is laid on a tile.
func (s *Sim) IsPath(x, y int) bool { return s.in(x, y) && s.Path[s.idx(x, y)] }

// LayPath lays a path on one tile.
func (s *Sim) LayPath(x, y int) error {
	if err := s.CanPlace(DefByID("path"), x, y); err != nil {
		return err
	}
	s.Path[s.idx(x, y)] = true
	s.Cash -= PathCost
	s.flows = map[int][]int16{}
	s.Version++
	return nil
}

// LiftPath picks a path up again, for what it cost.
func (s *Sim) LiftPath(x, y int) bool {
	if !s.IsPath(x, y) {
		return false
	}
	s.Path[s.idx(x, y)] = false
	s.Cash += PathCost
	s.flows = map[int][]int16{}
	s.Version++
	return true
}

func (s *Sim) SetPrice(a *Attraction, p int) {
	a.Price = int(clamp(float64(p), 0, float64(a.Def.MaxPrice)))
}

func (s *Sim) SetGate(p int) { s.Gate = int(clamp(float64(p), 0, MaxGate)) }

func (s *Sim) layoutChanged() {
	s.flows = map[int][]int16{}
	s.Version++
	for i := range s.Lit {
		s.Lit[i] = 0
		s.trashNear[i] = false
	}
	light := func(cx, cy, r float64) {
		for y := 0; y < s.H; y++ {
			for x := 0; x < s.W; x++ {
				d := math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy)
				if v := float32(1 - d/r); v > s.Lit[s.idx(x, y)] {
					s.Lit[s.idx(x, y)] = v
				}
			}
		}
	}
	light(float64(s.GateX)+.5, float64(s.GateY)+.5, 2.6)
	for _, a := range s.Attr {
		cx, cy := a.Center()
		if a.Def.Light > 0 {
			light(cx, cy, a.Def.Light+float64(a.Def.W)/2)
		}
		if a.Def.ID == "trash" {
			for y := 0; y < s.H; y++ {
				for x := 0; x < s.W; x++ {
					if math.Hypot(float64(x)+.5-cx, float64(y)+.5-cy) <= 3.2 {
						s.trashNear[s.idx(x, y)] = true
					}
				}
			}
		}
	}
}

// IsLit reports whether a tile is lit well enough at night.
func (s *Sim) IsLit(x, y int) bool { return s.in(x, y) && s.Lit[s.idx(x, y)] > 0.12 }

// unstickGuests moves anyone standing where something was just built.
func (s *Sim) unstickGuests() {
	for _, g := range s.Guests {
		if g.State == Queuing || g.State == Riding || g.outside {
			continue
		}
		if s.Free(g.cx, g.cy) {
			if g.moving && !s.Free(g.nx, g.ny) {
				g.moving = false
				g.X, g.Y = float64(g.cx)+.5, float64(g.cy)+.5
			}
			continue
		}
		f := s.bfs(s.idx(g.cx, g.cy))
		best, bd := -1, int16(math.MaxInt16)
		for i, d := range f {
			if d >= 0 && d < bd && s.grid[i] == 0 {
				best, bd = i, d
			}
		}
		if best >= 0 {
			g.cx, g.cy = best%s.W, best/s.W
			g.X, g.Y = float64(g.cx)+.5, float64(g.cy)+.5
			g.moving = false
		}
	}
}

// bfs measures walking distance from tile `from` to every free tile (-1 = can't get there).
// The start tile counts as walkable even when something is built on it.
func (s *Sim) bfs(from int) []int16 {
	d := make([]int16, len(s.grid))
	for i := range d {
		d[i] = -1
	}
	d[from] = 0
	q := []int{from}
	for len(q) > 0 {
		i := q[0]
		q = q[1:]
		x, y := i%s.W, i/s.W
		for _, n := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+n[0], y+n[1]
			if !s.in(nx, ny) {
				continue
			}
			j := s.idx(nx, ny)
			if d[j] < 0 && s.grid[j] == 0 {
				d[j] = d[i] + 1
				q = append(q, j)
			}
		}
	}
	return d
}

// stepCost is what it costs to walk onto a tile: guests go a little out of
// their way to stay on a path.
func (s *Sim) stepCost(i int) int16 {
	if s.Path[i] {
		return 2
	}
	return 3
}

// flow is the walking cost from every tile to `target` (-1 = can't get
// there). A guest on tile i steps to a neighbour n where
// flow[n] + stepCost(n) == flow[i].
func (s *Sim) flow(target int) []int16 {
	if f, ok := s.flows[target]; ok {
		return f
	}
	f := make([]int16, len(s.grid))
	for i := range f {
		f[i] = -1
	}
	f[target] = 0
	buckets := [][]int{{target}}
	for d := 0; d < len(buckets); d++ {
		for k := 0; k < len(buckets[d]); k++ {
			i := buckets[d][k]
			if int(f[i]) != d {
				continue // already reached more cheaply
			}
			x, y := i%s.W, i/s.W
			for _, n := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+n[0], y+n[1]
				if !s.in(nx, ny) {
					continue
				}
				j := s.idx(nx, ny)
				if s.grid[j] != 0 {
					continue
				}
				nd := f[i] + s.stepCost(i)
				if f[j] < 0 || nd < f[j] {
					f[j] = nd
					for len(buckets) <= int(nd) {
						buckets = append(buckets, nil)
					}
					buckets[nd] = append(buckets[nd], j)
				}
			}
		}
	}
	s.flows[target] = f
	return f
}

/* ---------- nights ---------- */

// ShowCount is how many tents have a show booked.
func (s *Sim) ShowCount() int {
	n := 0
	for _, a := range s.Attr {
		if a.Def.Cat == CatTent && a.Booked() > 0 {
			n++
		}
	}
	return n
}

// Wages is what the troupe will be paid tonight.
func (s *Sim) Wages() int {
	w := 0
	for _, p := range s.Troupe {
		w += p.Wage()
	}
	return w
}

// OpenGates starts the night.
func (s *Sim) OpenGates() error {
	if s.Phase != PhaseBuild {
		return nil
	}
	if s.ShowCount() == 0 {
		return ErrNoShow
	}
	s.Phase = PhaseNight
	s.T = 0
	s.closing = false
	s.spawnAcc = 0
	s.happySum, s.departed = 0, 0
	hiring := s.Report.Hiring
	s.Report = NightReport{Night: s.Night, Weather: s.Weather, RepBefore: s.Rep, Hiring: hiring}
	for _, a := range s.Attr {
		a.Night = AttrStats{}
	}
	s.emit(Event{Kind: "gate", Text: "The gates are open!"})
	return nil
}

func (s *Sim) appeal() float64 {
	v := 0.2
	for _, a := range s.Attr {
		switch a.Def.Cat {
		case CatTent:
			if a.Booked() > 0 {
				sc := a.Score()
				v += 0.12 + 0.1*sc.Stars*float64(a.Def.Slots)/3
			}
		case CatSideshow:
			v += 0.12
		case CatFood:
			v += 0.07
		}
	}
	return clamp(v, 0.2, 1.7)
}

// SpawnRate is how many guests arrive per second right now.
func (s *Sim) SpawnRate() float64 {
	curve := 1.0
	switch {
	case s.T < 0.25:
		curve = 0.5 + 2*s.T
	case s.T > 0.7:
		curve = math.Max(0.15, 1-(s.T-0.7)*4)
	}
	w := []float64{1, 0.9, 0.62}[s.Weather]
	gate := clamp(1.3-float64(s.Gate)/9, 0.25, 1.3)
	return 0.62 * (0.45 + s.Rep/4.5) * s.appeal() * w * curve * gate
}

// Update advances the simulation by dt seconds of game time.
func (s *Sim) Update(dt float64) {
	if s.Phase != PhaseNight {
		return
	}
	for dt > 0 {
		step := math.Min(dt, 0.05)
		s.step(step)
		dt -= step
		if s.Phase != PhaseNight {
			return
		}
	}
}

func (s *Sim) step(dt float64) {
	s.T += dt / NightSeconds
	if !s.closing && s.T >= 1 {
		s.closing = true
		for _, g := range s.Guests {
			if g.State != Riding && !g.leaving {
				s.sendHome(g, "Closing time already?")
			}
		}
		for _, a := range s.Attr {
			a.Queue = nil
		}
	}
	if !s.closing && s.T < 0.86 && len(s.Guests) < 260 {
		s.spawnAcc += s.SpawnRate() * dt
		for s.spawnAcc >= 1 {
			s.spawnAcc--
			s.spawn()
		}
	}
	for _, a := range s.Attr {
		a.update(s, dt)
	}
	alive := s.Guests[:0]
	for _, g := range s.Guests {
		g.update(s, dt)
		if !g.gone {
			alive = append(alive, g)
		}
	}
	for i := len(alive); i < len(s.Guests); i++ {
		s.Guests[i] = nil
	}
	s.Guests = alive
	if n := len(s.Guests); n > s.Report.PeakGuests {
		s.Report.PeakGuests = n
	}
	if s.closing && (len(s.Guests) == 0 || s.T > 1.3) {
		s.endNight()
	}
}

func (s *Sim) spawn() {
	g := &Guest{ID: s.Report.Guests + 1 + (s.Night-1)*1000}
	g.Kid = s.chance(0.3)
	g.Name = guestName(s.Rand, g.Kid)
	g.Money = 18 + s.Rand.Intn(30)
	if g.Kid {
		g.Money = 10 + s.Rand.Intn(14)
	}
	g.Happy = 58 + s.Rand.Float64()*14
	g.Hunger = s.Rand.Float64() * 45
	g.Thirst = s.Rand.Float64() * 45
	g.Energy = 70 + s.Rand.Float64()*30
	g.Stay = 4 + s.Rand.Intn(5)
	g.Look = s.Rand.Intn(8)
	g.Skin = s.Rand.Intn(5)
	g.Visits = map[int]int{}
	g.speed = 1.55 + s.Rand.Float64()*0.35
	if g.Kid {
		g.speed += 0.2
	}
	g.cx, g.cy = s.GateX, s.GateY
	g.X, g.Y = -0.9, float64(s.GateY)+.5
	g.outside = true // walking in through the gate
	g.ox, g.oy = (s.Rand.Float64()-.5)*.36, (s.Rand.Float64()-.5)*.36
	s.Cash += s.Gate
	s.Report.Admissions += s.Gate
	s.Report.Guests++
	s.Guests = append(s.Guests, g)
}

func (s *Sim) sendHome(g *Guest, why string) {
	if g.State == Queuing && g.Target != nil {
		g.Target.dequeue(g)
	}
	g.leaving = true
	g.Target = nil
	g.State = Walking
	g.goal = s.idx(s.GateX, s.GateY)
	if why != "" {
		g.think(s, why, 0.35)
	}
}

func (s *Sim) endNight() {
	for _, g := range s.Guests {
		s.depart(g)
	}
	s.Guests = nil
	for _, a := range s.Attr {
		a.Queue, a.Riding, a.Running = nil, nil, false
		s.Report.Upkeep += a.Def.Upkeep
	}
	litter := 0
	for i, l := range s.Litter {
		litter += int(l)
		s.Litter[i] = 0
	}
	r := &s.Report
	r.Litter = litter
	r.Cleanup = litter * CleanupEach
	r.Wages = s.Wages()
	s.Cash -= r.Upkeep + r.Cleanup + r.Wages
	avg := 60.0
	if s.departed > 0 {
		avg = s.happySum / float64(s.departed)
	}
	r.AvgHappy = avg
	delta := (avg-58)/45*0.5 + (r.AvgStars()-2.5)*0.08
	delta -= math.Max(0, float64(litter)-20) * 0.006
	s.Rep = clamp(s.Rep+delta, 0.5, 5)
	r.RepAfter = s.Rep
	s.Reports = append(s.Reports, *r)
	s.Phase = PhaseResults
}

func (s *Sim) depart(g *Guest) {
	s.happySum += clamp(g.Happy, 0, 100)
	s.departed++
}

// Next moves on from the results screen to the next morning.
func (s *Sim) Next() {
	if s.Phase != PhaseResults {
		return
	}
	if s.Cash < BrokeAt {
		s.Phase = PhaseBroke
		return
	}
	s.Night++
	s.Weather = []Weather{Clear, Clear, Breezy, Drizzle}[s.Rand.Intn(4)]
	s.Report = NightReport{}
	for _, p := range s.Troupe {
		p.Energy = 100 // a good night's sleep
	}
	s.refreshHireable()
	s.Phase = PhaseBuild
}

// Totals adds up every night so far.
func (s *Sim) Totals() NightReport {
	var t NightReport
	for _, r := range s.Reports {
		t.Guests += r.Guests
		t.Admissions += r.Admissions
		t.Shows += r.Shows
		t.Sideshows += r.Sideshows
		t.Food += r.Food
		t.Goods += r.Goods
		t.Wages += r.Wages
		t.Hiring += r.Hiring
		t.Upkeep += r.Upkeep
		t.Cleanup += r.Cleanup
		t.Litter += r.Litter
		t.ShowsRun += r.ShowsRun
		t.StarSum += r.StarSum
		t.Flubs += r.Flubs
		t.BestStars = math.Max(t.BestStars, r.BestStars)
		t.AvgHappy += r.AvgHappy / float64(len(s.Reports))
	}
	return t
}

// AvgHappy is the mood of the guests on the grounds right now.
func (s *Sim) AvgHappy() float64 {
	if len(s.Guests) == 0 {
		return 0
	}
	t := 0.0
	for _, g := range s.Guests {
		t += clamp(g.Happy, 0, 100)
	}
	return t / float64(len(s.Guests))
}

// Clock is the time of night as hours and minutes (6:00 PM to 11:00 PM).
func (s *Sim) Clock() (h, m int) {
	mins := 18*60 + int(math.Min(s.T, 1)*300)
	return mins / 60, mins % 60
}

var adultNames = []string{"Amy", "Ben", "Carla", "Dave", "Erin", "Frank", "Gina", "Hal", "Iris", "Joe", "Kim", "Lou", "Mae", "Nate", "Olive", "Pete", "Rosa", "Sam", "Tina", "Vic", "Wanda", "Walt", "Jo", "Ray", "Deb", "Cal", "June", "Russ"}
var kidNames = []string{"Tommy", "Lily", "Max", "Ruby", "Sammy", "Ella", "Jake", "Zoe", "Billy", "Rosie", "Theo", "Maisie"}

func guestName(r *rand.Rand, kid bool) string {
	if kid {
		return kidNames[r.Intn(len(kidNames))]
	}
	return adultNames[r.Intn(len(adultNames))]
}
