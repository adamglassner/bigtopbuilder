package circus

import (
	"errors"
	"math"
	"math/rand"
)

// Performer is one member of the troupe (or someone looking to join it).
type Performer struct {
	ID     int
	Name   string
	Act    *ActDef
	Skill  float64 // 0.8 (green) to 1.25 (star)
	Energy float64 // 0-100; each show uses some, a night's sleep restores it
	Tent   *Attraction
	Look   int
	Skin   int
	Shows  int // shows performed, all time
	Flubs  int
}

// Stars is the performer's skill as 1-5 stars.
func (p *Performer) Stars() int {
	return int(math.Round(clamp((p.Skill-0.8)/0.45*4+1, 1, 5)))
}

// Tired reports whether they're running low and might flub.
func (p *Performer) Tired() bool { return p.Energy < TiredAt }

var (
	ErrTooTall   = errors.New("The trapeze needs the Big Top")
	ErrNoSlot    = errors.New("That show is full")
	ErrNotTroupe = errors.New("Hire them first")
	ErrFull      = errors.New("Your troupe is full (12 performers)")
)

var stageNames = map[string][]string{
	"clown":    {"Bingo", "Pickles", "Dottie", "Wobbles", "Sprocket", "Noodle", "Patches", "Bubbles"},
	"stilts":   {"Tall Tilly", "Long Lenny", "Lofty Lou", "Stretch", "High Hattie"},
	"juggler":  {"Marvelo", "Juggling Jo", "Fingers Finn", "The Great Gus", "Ottilie", "Rinaldo"},
	"strong":   {"Iron Ida", "Mighty Mabel", "Brunhilde", "Big Bertha", "Steel Sadie"},
	"acrobats": {"The Flying Fitzes", "The Tumbling Tates", "The Bounding Bros.", "The Somersault Sisters"},
	"fire":     {"Ember", "Blaze Bonnet", "Pyro Pete", "Scorchy", "Cinder"},
	"trapeze":  {"The Amazing Aurelia", "Fearless Felix", "Skylark Sue", "The Great Zephyr"},
}

func (s *Sim) newPerformer(act *ActDef, r *rand.Rand) *Performer {
	s.nextPerf++
	names := stageNames[act.ID]
	name := names[r.Intn(len(names))]
	for tries := 0; tries < 10 && s.nameTaken(name); tries++ {
		name = names[r.Intn(len(names))]
	}
	return &Performer{ID: s.nextPerf, Name: name, Act: act, Skill: 0.82 + r.Float64()*0.4, Energy: 100, Look: r.Intn(8), Skin: r.Intn(5)}
}

func (s *Sim) nameTaken(n string) bool {
	for _, p := range append(append([]*Performer{}, s.Troupe...), s.Hireable...) {
		if p.Name == n {
			return true
		}
	}
	return false
}

// HireCost is what it costs to sign a performer; better ones ask for more.
func (p *Performer) HireCost() int {
	return int(math.Round(float64(p.Act.Hire)*(0.6+p.Skill*0.4)/10) * 10)
}

// Wage is what they're paid each night.
func (p *Performer) Wage() int { return int(math.Round(float64(p.Act.Wage) * (0.7 + p.Skill*0.3))) }

// refreshHireable puts a fresh set of performers on the noticeboard.
func (s *Sim) refreshHireable() {
	s.Hireable = nil
	pool := []*ActDef{}
	for _, a := range Acts {
		pool = append(pool, a)
	}
	s.Rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	for i := 0; i < ForHire; i++ {
		s.Hireable = append(s.Hireable, s.newPerformer(pool[i%len(pool)], s.Rand))
	}
}

// Hire signs a performer from the noticeboard.
func (s *Sim) Hire(p *Performer) error {
	idx := -1
	for i, h := range s.Hireable {
		if h == p {
			idx = i
		}
	}
	if idx < 0 {
		return ErrNotTroupe
	}
	if len(s.Troupe) >= MaxTroupe {
		return ErrFull
	}
	if s.Cash < p.HireCost() {
		return ErrCash
	}
	s.Cash -= p.HireCost()
	s.Report.Hiring += p.HireCost()
	s.Hireable = append(s.Hireable[:idx], s.Hireable[idx+1:]...)
	s.Troupe = append(s.Troupe, p)
	return nil
}

// Dismiss lets a performer go (no refund).
func (s *Sim) Dismiss(p *Performer) {
	s.Unbook(p)
	for i, q := range s.Troupe {
		if q == p {
			s.Troupe = append(s.Troupe[:i], s.Troupe[i+1:]...)
			return
		}
	}
}

// Book puts a performer in a tent's lineup at slot i (moving them from wherever they were).
func (s *Sim) Book(p *Performer, a *Attraction, i int) error {
	if a.Def.Cat != CatTent || i < 0 || i >= len(a.Lineup) {
		return ErrNoSlot
	}
	if p.Act.Tall && !a.Def.Tall {
		return ErrTooTall
	}
	if old := a.Lineup[i]; old != nil && old != p {
		old.Tent = nil
	}
	s.Unbook(p)
	a.Lineup[i] = p
	p.Tent = a
	return nil
}

// Unbook takes a performer out of whatever show they're in.
func (s *Sim) Unbook(p *Performer) {
	if p.Tent != nil {
		for i, q := range p.Tent.Lineup {
			if q == p {
				p.Tent.Lineup[i] = nil
			}
		}
	}
	p.Tent = nil
}

// ClearSlot empties slot i of a tent's lineup.
func (s *Sim) ClearSlot(a *Attraction, i int) {
	if i >= 0 && i < len(a.Lineup) && a.Lineup[i] != nil {
		a.Lineup[i].Tent = nil
		a.Lineup[i] = nil
	}
}

// MoveSlot swaps slot i with slot i+d, to change the running order.
func (s *Sim) MoveSlot(a *Attraction, i, d int) {
	j := i + d
	if i < 0 || j < 0 || i >= len(a.Lineup) || j >= len(a.Lineup) {
		return
	}
	a.Lineup[i], a.Lineup[j] = a.Lineup[j], a.Lineup[i]
}

/* ---------- the show ---------- */

// ActScore is one act's part in a show.
type ActScore struct {
	P     *Performer
	Value float64
	Note  string // why it scored the way it did
	Good  bool   // the note is praise
}

// ShowScore rates a tent's show as it's booked right now.
type ShowScore struct {
	Acts  []ActScore
	Total float64
	Stars float64 // 0-5
	Fun   float64 // happiness a guest gets from watching
}

// Score works out how good a tent's show is: each act's appeal, their skill
// and how fresh they are, plus pacing: a laugh to warm the crowd up, no two
// acts of the same kind back to back, and a thrilling finale.
func (a *Attraction) Score() ShowScore {
	var sc ShowScore
	var acts []*Performer
	for _, p := range a.Lineup {
		if p != nil {
			acts = append(acts, p)
		}
	}
	if len(acts) == 0 {
		return sc
	}
	for i, p := range acts {
		v := p.Act.Appeal * p.Skill * (0.6 + 0.4*p.Energy/100)
		note, good := "", false
		switch {
		case i > 0 && acts[i-1].Act.Style == p.Act.Style:
			v *= 0.72
			note = "Too much " + StyleNames[p.Act.Style] + " in a row"
		case i == len(acts)-1 && len(acts) > 1 && p.Act.Style == Thrills:
			v *= 1.3
			note, good = "Big finish!", true
		case i == 0 && len(acts) > 1 && p.Act.Style == Laughs:
			v *= 1.15
			note, good = "Warms up the crowd", true
		case i > 0 && i < len(acts)-1:
			note, good = "Good change of pace", true
		}
		if p.Tired() {
			note, good = "Tired: might flub", false
		}
		sc.Acts = append(sc.Acts, ActScore{P: p, Value: v, Note: note, Good: good})
		sc.Total += v
	}
	sc.Stars = clamp(sc.Total/(float64(a.Def.Slots)*32)*5, 0, 5)
	sc.Fun = clamp(8+sc.Total*0.36, 0, 62)
	return sc
}

// Booked is how many acts a tent has in its lineup.
func (a *Attraction) Booked() int {
	n := 0
	for _, p := range a.Lineup {
		if p != nil {
			n++
		}
	}
	return n
}

// NewPerformer makes a performer of the given act (for tests and screenshots).
func (s *Sim) NewPerformer(act string) *Performer { return s.newPerformer(ActByID(act), s.Rand) }
