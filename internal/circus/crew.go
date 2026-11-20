package circus

import "math"

// Sweeper is one of the grounds crew: during the night they walk to the
// nearest litter and sweep it up.
type Sweeper struct {
	ID    int
	Name  string
	Look  int
	Skin  int
	X, Y  float64 // tile coordinates
	Swept int     // pieces of litter tonight

	cx, cy, nx, ny int
	moving         bool
	target         int // tile they're heading for (-1 = none)
	sweepT         float64
	idleT          float64
	Sweeping       bool // broom moving, for the screen
}

var crewNames = []string{"Gus", "Hattie", "Moe", "Pearl", "Ike", "Nell", "Otto", "Birdie", "Clem", "Dot"}

// HireSweeper adds someone to the grounds crew.
func (s *Sim) HireSweeper() error {
	if len(s.Crew) >= MaxCrew {
		return ErrCrewFull
	}
	if s.Cash < SweeperHire {
		return ErrCash
	}
	s.Cash -= SweeperHire
	s.Report.Hiring += SweeperHire
	s.nextPerf++
	name := crewNames[s.Rand.Intn(len(crewNames))]
	for tries := 0; tries < 12; tries++ {
		taken := false
		for _, c := range s.Crew {
			if c.Name == name {
				taken = true
			}
		}
		if !taken {
			break
		}
		name = crewNames[s.Rand.Intn(len(crewNames))]
	}
	c := &Sweeper{ID: s.nextPerf, Name: name, Look: s.Rand.Intn(8), Skin: s.Rand.Intn(5), target: -1}
	s.Crew = append(s.Crew, c)
	if s.Phase == PhaseNight {
		s.placeAtGate(c)
	}
	return nil
}

// DismissSweeper lets the last sweeper go.
func (s *Sim) DismissSweeper() {
	if n := len(s.Crew); n > 0 {
		s.Crew = s.Crew[:n-1]
	}
}

func (s *Sim) placeAtGate(c *Sweeper) {
	c.cx, c.cy = s.GateX, s.GateY
	c.X, c.Y = float64(s.GateX)+.5, float64(s.GateY)+.5
	c.moving, c.target, c.Swept, c.Sweeping = false, -1, 0, false
}

func (s *Sim) crewToGate() {
	for _, c := range s.Crew {
		s.placeAtGate(c)
	}
}

// nearestLitter finds the closest littered tile nobody else is going for.
func (s *Sim) nearestLitter(c *Sweeper) int {
	taken := map[int]bool{}
	for _, o := range s.Crew {
		if o != c && o.target >= 0 {
			taken[o.target] = true
		}
	}
	f := s.bfs(s.idx(c.cx, c.cy))
	best, bd := -1, int16(math.MaxInt16)
	for i, l := range s.Litter {
		if l > 0 && f[i] >= 0 && f[i] < bd && !taken[i] {
			best, bd = i, f[i]
		}
	}
	return best
}

func (c *Sweeper) update(s *Sim, dt float64) {
	c.Sweeping = false
	if !s.in(c.cx, c.cy) || !s.Free(c.cx, c.cy) {
		s.placeAtGate(c) // something got built on top of them
	}
	if c.moving {
		tx, ty := float64(c.nx)+.5, float64(c.ny)+.5
		dx, dy := tx-c.X, ty-c.Y
		d := math.Hypot(dx, dy)
		step := 1.35 * dt
		if s.IsPath(c.nx, c.ny) {
			step *= 1.3
		}
		if d <= step {
			c.X, c.Y = tx, ty
			c.cx, c.cy = c.nx, c.ny
			c.moving = false
		} else {
			c.X += dx / d * step
			c.Y += dy / d * step
		}
		return
	}
	cur := s.idx(c.cx, c.cy)
	if c.target >= 0 && s.Litter[c.target] == 0 {
		c.target = -1
	}
	if c.target < 0 {
		c.idleT -= dt
		if c.idleT > 0 {
			return
		}
		c.target = s.nearestLitter(c)
		if c.target < 0 {
			c.idleT = 1.5 // nothing to sweep; look again soon
			return
		}
	}
	if cur == c.target {
		c.Sweeping = true
		c.sweepT += dt
		if c.sweepT >= 0.7 {
			c.sweepT = 0
			s.Litter[cur]--
			c.Swept++
			s.Report.Swept++
		}
		return
	}
	f := s.flow(c.target)
	d := f[cur]
	if d < 0 {
		c.target = -1
		return
	}
	for _, n := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		x, y := c.cx+n[0], c.cy+n[1]
		if j := s.idx(x, y); s.in(x, y) && f[j] >= 0 && f[j]+s.stepCost(j) == d {
			c.nx, c.ny, c.moving = x, y, true
			return
		}
	}
	c.target = -1
}

// BuyPosters puts posters up around town for a bigger crowd tonight.
func (s *Sim) BuyPosters() error {
	if s.Posters || s.Phase != PhaseBuild {
		return ErrPosters
	}
	if s.Cash < PosterCost {
		return ErrCash
	}
	s.Cash -= PosterCost
	s.Report.Posters += PosterCost
	s.Posters = true
	return nil
}
