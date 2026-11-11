package circus

import "math"

// AttrStats counts what a building did.
type AttrStats struct {
	Customers, Revenue, Costs int
	Shows, Flubs, Refused     int
}

// Attraction is one tent, sideshow, stand or extra on the grounds.
type Attraction struct {
	ID     int
	Def    *Def
	X, Y   int // top-left tile
	Price  int
	Lineup []*Performer // tents: the running order (nil = empty slot)

	Queue   []*Guest
	Riding  []*Guest // in the audience, or inside
	Running bool
	Timer   float64 // time left on this show or visit
	Length  float64 // how long this show or visit runs
	loadT   float64
	Anim    float64 // seconds it has been running, for the screen

	Show      ShowScore // the show that's on (tents)
	flubbed   []bool    // which acts flubbed in the show that's on
	LastShow  float64   // stars of the last show, for the screen
	announced int       // the act last announced (-1 = none yet)

	Night, Total AttrStats
}

// Open reports whether guests can go in: tents need at least one act.
func (a *Attraction) Open() bool {
	if a.Def.Cat == CatTent {
		return a.Booked() > 0
	}
	return a.Def.Visitable()
}

// Entrance is the tile guests line up on: just below the middle of the footprint.
func (a *Attraction) Entrance() (int, int) { return a.X + a.Def.W/2, a.Y + a.Def.H }

func (a *Attraction) entranceCenter() (float64, float64) {
	x, y := a.Entrance()
	return float64(x) + .5, float64(y) + .5
}

// Center of the footprint, in tile coordinates.
func (a *Attraction) Center() (float64, float64) {
	return float64(a.X) + float64(a.Def.W)/2, float64(a.Y) + float64(a.Def.H)/2
}

// CurrentAct is which act of the running show is on (-1 if none).
func (a *Attraction) CurrentAct() int {
	if !a.Running || len(a.Show.Acts) == 0 || a.Length <= 0 {
		return -1
	}
	return min(len(a.Show.Acts)-1, int(a.Anim/a.Length*float64(len(a.Show.Acts))))
}

// Flubbed reports whether act i of the running show went wrong.
func (a *Attraction) Flubbed(i int) bool { return i >= 0 && i < len(a.flubbed) && a.flubbed[i] }

// Excitement is how worked up the audience is right now (0-1): it builds act
// by act toward the show's rating.
func (a *Attraction) Excitement() float64 {
	if !a.Running || len(a.Show.Acts) == 0 || a.Length <= 0 {
		return 0
	}
	prog := a.Anim / a.Length * float64(len(a.Show.Acts))
	sum := 0.0
	for i, ac := range a.Show.Acts {
		part := clamp(prog-float64(i), 0, 1)
		v := ac.Value
		if a.Flubbed(i) {
			v *= 0.4
		}
		sum += v * part
	}
	return clamp(sum/(float64(a.Def.Slots)*32), 0, 1)
}

// QueueSpot is where the k-th guest in line stands, in tile coordinates.
func (a *Attraction) QueueSpot(k int) (float64, float64) {
	ex, ey := a.entranceCenter()
	row, col := k/4, k%4
	if row%2 == 1 {
		col = 3 - col
	}
	return ex + (float64(col)-1.5)*0.23, ey - 0.12 + float64(row)*0.22
}

func (a *Attraction) dequeue(g *Guest) {
	for i, q := range a.Queue {
		if q == g {
			a.Queue = append(a.Queue[:i], a.Queue[i+1:]...)
			return
		}
	}
}

func (a *Attraction) update(s *Sim, dt float64) {
	if a.Running {
		a.Anim += dt
		a.Timer -= dt
		if a.Def.Cat == CatTent {
			s.showProgress(a)
		}
		if a.Timer <= 0 {
			a.finish(s)
		}
		return
	}
	if len(a.Queue) == 0 {
		a.loadT = 0
		return
	}
	a.loadT += dt
	wait := 0.3
	if a.Def.Cat == CatTent {
		wait = 9.0 // wait for an audience
	} else if a.Def.Cat == CatSideshow {
		wait = 2.5
	}
	if len(a.Queue) >= a.Def.Capacity || a.loadT >= wait {
		a.start(s)
	}
}

func (a *Attraction) start(s *Sim) {
	if a.Def.Cat == CatTent && a.Booked() == 0 {
		return
	}
	n := min(a.Def.Capacity, len(a.Queue))
	a.Riding = append(a.Riding[:0], a.Queue[:n]...)
	a.Queue = append([]*Guest{}, a.Queue[n:]...)
	cx, cy := a.Center()
	for _, g := range a.Riding {
		g.State = Riding
		g.wait = 0
		g.X, g.Y = cx, cy
		if a.Price > 0 && g.Money >= a.Price {
			g.Money -= a.Price
			g.Spent += a.Price
			s.Cash += a.Price
			a.Night.Revenue += a.Price
			a.Total.Revenue += a.Price
			switch a.Def.Cat {
			case CatTent:
				s.Report.Shows += a.Price
			case CatSideshow:
				s.Report.Sideshows += a.Price
			case CatFood:
				s.Report.Food += a.Price
			}
		}
		if a.Def.Cat != CatExtra {
			a.Night.Customers++
			a.Total.Customers++
		}
	}
	if a.Def.Cat != CatFood && a.Def.Cat != CatExtra && a.Price > 0 {
		s.emit(Event{Kind: "coin", X: cx, Y: cy - float64(a.Def.H)/2, A: a, Text: "+$" + itoa(a.Price*n)})
	}
	a.Running = true
	a.announced = -1
	a.Anim = 0
	a.Length = a.Def.Cycle * (0.9 + s.Rand.Float64()*0.2)
	if a.Def.Cat == CatTent {
		a.Length = a.Def.Cycle * (0.6 + 0.4*float64(a.Booked())/float64(a.Def.Slots))
		a.Show = a.Score()
		a.flubbed = make([]bool, len(a.Show.Acts))
		for i, ac := range a.Show.Acts {
			p := ac.P
			if p.Tired() && s.chance(0.3+(TiredAt-p.Energy)/100) {
				a.flubbed[i] = true
			}
		}
		s.emit(Event{Kind: "showstart", X: cx, Y: cy, A: a})
	}
	a.Timer = a.Length
	a.loadT = 0
}

// showProgress announces each act as it comes on, and any flubs.
func (s *Sim) showProgress(a *Attraction) {
	i := a.CurrentAct()
	if i < 0 || i == a.announced {
		return
	}
	a.announced = i
	if a.Flubbed(i) {
		p := a.Show.Acts[i].P
		cx, cy := a.Center()
		s.emit(Event{Kind: "flub", X: cx, Y: cy, A: a, Text: p.Name + " " + flubLine(s, p.Act.ID)})
	}
}

func flubLine(s *Sim, act string) string {
	switch act {
	case "juggler":
		return "dropped a club!"
	case "clown":
		return "slipped on their own banana peel!"
	case "stilts":
		return "wobbled off a stilt!"
	case "strong":
		return "couldn't lift the piano!"
	case "acrobats":
		return "toppled out of the pyramid!"
	case "fire":
		return "only managed a puff of smoke!"
	case "trapeze":
		return "missed the catch and hit the net!"
	}
	return pick(s, "missed a cue!", "fumbled the act!")
}

func (a *Attraction) finish(s *Sim) {
	a.Running = false
	riders := a.Riding
	a.Riding = nil
	a.announced = -1
	d := a.Def
	cx, cy := a.Center()
	fun := d.Fun
	stars := 0.0
	if d.Cat == CatTent {
		total := 0.0
		flubs := 0
		for i, ac := range a.Show.Acts {
			v := ac.Value
			if a.Flubbed(i) {
				v *= 0.4
				flubs++
				ac.P.Flubs++
				a.Night.Flubs++
			}
			total += v
			ac.P.Energy = clamp(ac.P.Energy-ShowEnergy*(1.2-0.3*ac.P.Skill), 0, 100)
			ac.P.Shows++
		}
		fun = clamp(8+total*0.36, 0, 62)
		stars = clamp(total/(float64(d.Slots)*32)*5, 0, 5)
		a.LastShow = stars
		a.Night.Shows++
		a.Total.Shows++
		s.Report.ShowsRun++
		s.Report.StarSum += stars
		if stars > s.Report.BestStars {
			s.Report.BestStars = stars
		}
		s.Report.Flubs += flubs
		s.emit(Event{Kind: "applause", X: cx, Y: cy - float64(d.H)/2, A: a, V: stars})
	}
	for _, g := range riders {
		g.Activities++
		g.Visits[a.ID]++
		g.State = Idle
		g.X, g.Y = a.entranceCenter()
		g.X += g.ox * 0.6
		g.cx, g.cy = a.Entrance()
		g.moving = false
		switch d.Cat {
		case CatTent:
			g.Happy += fun * (0.75 + s.Rand.Float64()*0.3)
			g.Energy = clamp(g.Energy+12, 0, 100) // a sit-down
			if g.Kid && g.Balloon == 0 && s.chance(0.35) {
				g.Balloon = 1 + s.Rand.Intn(5) // a souvenir balloon on the way out
			}
			switch {
			case stars >= 4.2:
				g.think(s, pick(s, "Bravo! Bravo!", "Best show ever!", "What a finale!", "I want to run away with the circus!"), 0.5)
			case stars >= 2.8:
				g.think(s, pick(s, "Great show!", "That was fun!", "Did you see that?!"), 0.3)
			default:
				g.think(s, pick(s, "Bit of a dull show.", "Is that it?", "Meh."), 0.4)
			}
		case CatSideshow:
			g.Happy += fun * (0.55 + s.Rand.Float64()*0.25)
			g.Energy -= 3
			g.think(s, pick(s, "Ha! That was weird!", "Spooky!", "Let's go again!", "Did you see my reflection?"), 0.25)
		case CatFood:
			g.Hunger = clamp(g.Hunger-d.Hunger, 0, 100)
			g.Thirst = clamp(g.Thirst-d.Thirst, 0, 100)
			g.Happy += fun * 0.7
			g.litterT = 3 + s.Rand.Float64()*7
			goods := int(math.Round(d.Goods))
			s.Cash -= goods
			s.Report.Goods += goods
			a.Night.Costs += goods
			g.think(s, pick(s, "Yum!", "Mmm, "+d.Name+"!", "Just what I needed."), 0.25)
			if a.Price > 0 {
				s.emit(Event{Kind: "coin", X: cx, Y: cy - 1, A: a, Text: "+$" + itoa(a.Price)})
			}
		case CatExtra:
			g.Energy = clamp(g.Energy+d.Rest, 0, 100)
			g.Happy += 3
		}
	}
	for _, g := range riders {
		if g.State == Idle {
			g.decide(s)
		}
	}
}

func pick(s *Sim, opts ...string) string { return opts[s.Rand.Intn(len(opts))] }

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
