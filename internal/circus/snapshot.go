package circus

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"
)

// SnapshotVersion bumps when the saved-run format changes incompatibly.
const SnapshotVersion = 1

type savedPerf struct {
	ID                 int
	Name, Act          string
	Skill              float64
	Look, Skin         int
	Shows, Flubs, Rose int
}

type savedAttr struct {
	ID, X, Y, Price int
	Def             string
	Lineup          []int // performer IDs, 0 = empty slot
	Total           AttrStats
}

type savedCrew struct {
	ID, Look, Skin int
	Name           string
}

type savedRun struct {
	Version  int
	Night    int
	Cash     int
	Rep      float64
	Gate     int
	Weather  Weather
	Attr     []savedAttr
	Troupe   []savedPerf
	Hireable []savedPerf
	Crew     []savedCrew
	Paths    []int // tile indexes
	Traffic  []float32
	Reports  []NightReport
	Report   NightReport // money already spent this morning
	Posters  bool
	NextID   int
	NextPerf int
	SavedAt  time.Time
}

// Snapshot saves the run. Only mornings are saved: it returns an error at night.
func (s *Sim) Snapshot() ([]byte, error) {
	if s.Phase != PhaseBuild {
		return nil, errors.New("the circus is only saved in the morning")
	}
	perf := func(p *Performer) savedPerf {
		return savedPerf{ID: p.ID, Name: p.Name, Act: p.Act.ID, Skill: p.Skill, Look: p.Look, Skin: p.Skin, Shows: p.Shows, Flubs: p.Flubs, Rose: p.Rose}
	}
	r := savedRun{Version: SnapshotVersion, Night: s.Night, Cash: s.Cash, Rep: s.Rep, Gate: s.Gate, Weather: s.Weather,
		Traffic: s.Traffic, Reports: s.Reports, Report: s.Report, Posters: s.Posters, NextID: s.nextID, NextPerf: s.nextPerf, SavedAt: time.Now()}
	for _, a := range s.Attr {
		sa := savedAttr{ID: a.ID, X: a.X, Y: a.Y, Price: a.Price, Def: a.Def.ID, Total: a.Total}
		for _, p := range a.Lineup {
			id := 0
			if p != nil {
				id = p.ID
			}
			sa.Lineup = append(sa.Lineup, id)
		}
		r.Attr = append(r.Attr, sa)
	}
	for _, p := range s.Troupe {
		r.Troupe = append(r.Troupe, perf(p))
	}
	for _, p := range s.Hireable {
		r.Hireable = append(r.Hireable, perf(p))
	}
	for _, c := range s.Crew {
		r.Crew = append(r.Crew, savedCrew{ID: c.ID, Look: c.Look, Skin: c.Skin, Name: c.Name})
	}
	for i, p := range s.Path {
		if p {
			r.Paths = append(r.Paths, i)
		}
	}
	return json.Marshal(r)
}

// SavedNight peeks at a saved run: which night it's on and its cash.
func SavedNight(b []byte) (night, cash int, ok bool) {
	var r savedRun
	if json.Unmarshal(b, &r) != nil || r.Version != SnapshotVersion || r.Night < 1 {
		return 0, 0, false
	}
	return r.Night, r.Cash, true
}

// Restore loads a saved run, on the morning it was saved.
func Restore(b []byte) (*Sim, error) {
	var r savedRun
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	if r.Version != SnapshotVersion {
		return nil, errors.New("that save is from a different version")
	}
	s := New(time.Now().UnixNano())
	s.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.Night, s.Cash, s.Rep, s.Gate, s.Weather = r.Night, r.Cash, r.Rep, r.Gate, r.Weather
	s.Reports, s.Report, s.Posters = r.Reports, r.Report, r.Posters
	s.nextID, s.nextPerf = r.NextID, r.NextPerf
	if len(r.Traffic) == len(s.Traffic) {
		copy(s.Traffic, r.Traffic)
	}
	byID := map[int]*Performer{}
	load := func(sp savedPerf) *Performer {
		act := ActByID(sp.Act)
		if act == nil {
			return nil
		}
		p := &Performer{ID: sp.ID, Name: sp.Name, Act: act, Skill: sp.Skill, Energy: 100, Look: sp.Look, Skin: sp.Skin, Shows: sp.Shows, Flubs: sp.Flubs, Rose: sp.Rose}
		byID[p.ID] = p
		return p
	}
	s.Troupe, s.Hireable = nil, nil
	for _, sp := range r.Troupe {
		if p := load(sp); p != nil {
			s.Troupe = append(s.Troupe, p)
		}
	}
	for _, sp := range r.Hireable {
		if p := load(sp); p != nil {
			s.Hireable = append(s.Hireable, p)
		}
	}
	for _, sc := range r.Crew {
		s.Crew = append(s.Crew, &Sweeper{ID: sc.ID, Name: sc.Name, Look: sc.Look, Skin: sc.Skin, target: -1})
	}
	for _, i := range r.Paths {
		if i >= 0 && i < len(s.Path) {
			s.Path[i] = true
		}
	}
	for _, sa := range r.Attr {
		d := DefByID(sa.Def)
		if d == nil || sa.X < 0 || sa.Y < 0 || sa.X+d.W > s.W || sa.Y+d.H > s.H {
			continue
		}
		a := &Attraction{ID: sa.ID, Def: d, X: sa.X, Y: sa.Y, Price: sa.Price, Total: sa.Total, announced: -1}
		if d.Slots > 0 {
			a.Lineup = make([]*Performer, d.Slots)
			for i, id := range sa.Lineup {
				if p := byID[id]; p != nil && i < d.Slots && p.Tent == nil {
					a.Lineup[i] = p
					p.Tent = a
				}
			}
		}
		s.Attr = append(s.Attr, a)
		s.stamp(d, a.X, a.Y, a.ID)
		if a.ID > s.nextID {
			s.nextID = a.ID
		}
	}
	s.Phase = PhaseBuild
	s.layoutChanged()
	return s, nil
}
