package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	ct "github.com/adamglassner/bigtopbuilder/internal/circus"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type hit struct {
	x, y, w, h float64
	fn         func()
	drag       func(mx float64) // sliders
}

type floater struct {
	x, y, t float64
	text    string
	good    bool
	stars   float64 // a show's rating, drawn as stars instead of text
}

type toastMsg struct {
	title, body, kind string
	t                 float64
	a                 *ct.Attraction
}

type game struct {
	sim      *ct.Sim
	save     saveData
	screen   string // "title" or "play"
	panel    string // "", "menu", "help", "settings"
	speed    int    // 1, 2 or 3
	paused   bool
	tab      ct.Cat
	placing  *ct.Def
	sel      *ct.Attraction
	selGuest *ct.Guest
	slotPick int // tent slot being booked in the inspector (-1 = none)

	cannonFired  map[int]bool // cannons that have gone BOOM this show
	lastRunSave  float64
	runPeek      *runPeek
	noSave       bool // screenshot runs don't touch the saved circus
	brokeCleared bool
	painting     bool // dragging to lay (or lift) paths
	paintAdd     bool // true = laying paths, false = picking them up
	paintAt      [2]int

	hits     []hit
	dragging func(mx float64)
	mx, my   float64
	scale    float64
	offX     float64
	offY     float64

	clock      float64
	floaters   []floater
	toasts     []toastMsg
	log        []string
	quit       bool
	canvas     *ebiten.Image
	fixedMouse bool // screenshots
	heroShot   bool
	runStart   time.Time
}

func newGame() *game {
	g := &game{save: loadSave(), screen: "title", speed: 1, scale: 1, slotPick: -1}
	return g
}

func (g *game) newRun() {
	g.startRun(ct.New(time.Now().UnixNano()))
	g.save.Stats.Runs++
	g.persist()
	g.note("Welcome to " + g.sim.Town.Name + "! " + g.sim.Troupe[0].Name + " the Clown and " + g.sim.Troupe[1].Name + " the Juggler are ready to perform.")
	g.note("Raise a tent, then click it to book your acts.")
	g.saveRun()
}

// continueRun picks up the saved circus on the morning it was saved.
func (g *game) continueRun() {
	b, err := os.ReadFile(runPath())
	if err != nil {
		g.toast("Couldn't find your saved circus", "", "warn")
		return
	}
	s, err := ct.Restore(b)
	if err != nil {
		g.toast("Couldn't load your saved circus", err.Error(), "warn")
		return
	}
	g.startRun(s)
	g.note("Welcome back to " + s.Town.Name + "! It's the morning of night " + itoa(s.Night) + ".")
}

func (g *game) startRun(s *ct.Sim) {
	g.sim = s
	g.screen = "play"
	g.panel = ""
	g.placing, g.sel, g.selGuest = nil, nil, nil
	g.paused, g.speed = false, 1
	g.tab = ct.CatTent
	g.slotPick = -1
	g.floaters, g.toasts, g.log = nil, nil, nil
	g.runStart = time.Now()
	g.cannonFired = map[int]bool{}
	g.lastRunSave = g.clock
	g.brokeCleared = false
}

/* ---------- the saved run ---------- */

// runPath is where the circus in progress is kept, next to the settings.
func runPath() string {
	dir := appDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, bundleName, "run.json")
}

// saveRun saves the circus; only mornings can be saved.
func (g *game) saveRun() {
	if g.sim == nil || g.sim.Phase != ct.PhaseBuild || g.noSave || runPath() == "" {
		return
	}
	b, err := g.sim.Snapshot()
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(runPath()), 0o755)
	tmp := runPath() + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		os.Rename(tmp, runPath())
	}
	g.lastRunSave = g.clock
	g.runPeek = nil
}

// savedRun says which night the saved circus is on (ok=false if there isn't one).
func (g *game) savedRun() (night, cash int, ok bool) {
	if g.runPeek == nil {
		g.runPeek = &runPeek{}
		if p := runPath(); p != "" && !g.noSave {
			if b, err := os.ReadFile(p); err == nil {
				g.runPeek.night, g.runPeek.cash, g.runPeek.ok = ct.SavedNight(b)
			}
		}
	}
	return g.runPeek.night, g.runPeek.cash, g.runPeek.ok
}

func (g *game) deleteRun() {
	if p := runPath(); p != "" && !g.noSave {
		os.Remove(p)
	}
	g.runPeek = nil
}

type runPeek struct {
	night, cash int
	ok          bool
}

// applyTestQuery sets up a scene for automated screenshots.
func (g *game) applyTestQuery(q string) {
	if q == "" {
		return
	}
	has := func(k string) bool { return strings.Contains(q, k) }
	g.heroShot = has("hero")
	g.noSave = true
	defer func() {
		switch {
		case has("panel=help"):
			g.panel = "help"
		case has("panel=settings"):
			g.panel = "settings"
		case has("panel=troupe"):
			g.panel = "troupe"
		}
	}()
	if !has("play") && !has("night") && !has("results") && !has("done") {
		return
	}
	g.newRun()
	s := g.sim
	s.Rand.Seed(5)
	if has("build") {
		s.Cash = 1 << 20
		for _, b := range []struct {
			id   string
			x, y int
		}{{"bigtop", 8, 1}, {"ringtent", 3, 2}, {"ringtent", 15, 2}, {"funhouse", 19, 2}, {"mirrors", 4, 8}, {"fortune", 7, 9},
			{"popcorn", 11, 8}, {"cottoncandy", 14, 8}, {"hotdogs", 17, 8}, {"soda", 20, 7}, {"lamp", 13, 6}, {"lamp", 3, 12}, {"trash", 10, 11},
			{"trash", 16, 11}, {"bench", 12, 11}, {"bench", 20, 11}, {"lamp", 22, 12}} {
			s.Place(ct.DefByID(b.id), b.x, b.y)
		}
		var tents []*ct.Attraction
		for _, a := range s.Attr {
			if a.Def.Cat == ct.CatTent {
				tents = append(tents, a)
			}
		}
		book := func(a *ct.Attraction, acts ...string) {
			for i, id := range acts {
				pf := &ct.Performer{}
				for _, h := range s.Hireable {
					if h.Act.ID == id {
						pf = h
					}
				}
				if pf.Act == nil {
					s.Hireable = append(s.Hireable, s.NewPerformer(id))
					pf = s.Hireable[len(s.Hireable)-1]
				}
				s.Hire(pf)
				s.Book(pf, a, i)
			}
		}
		book(tents[0], "stilts", "juggler", "acrobats", "strong", "trapeze")
		s.Book(s.Troupe[0], tents[1], 0)
		s.Book(s.Troupe[1], tents[1], 1)
		fire := s.NewPerformer("fire")
		s.Hireable = append(s.Hireable, fire)
		s.Hire(fire)
		s.Book(fire, tents[1], 2)
		book(tents[2], "clown", "acrobats")
		s.Cash = 4120
		if has("inspect") {
			g.sel = tents[0]
		}
		if has("pick") {
			g.sel, g.slotPick = tents[0], 4
		}
	}
	if has("v3") {
		s.Cash = 1 << 20
		for _, b := range []struct {
			id   string
			x, y int
		}{{"threering", 7, 1}, {"ringtent", 2, 1}, {"cannon", 14, 2}, {"funhouse", 19, 2},
			{"popcorn", 4, 8}, {"cottoncandy", 7, 8}, {"hotdogs", 11, 8}, {"soda", 14, 8},
			{"flowers", 1, 5}, {"flowers", 5, 5}, {"flowers", 13, 5}, {"flowers", 17, 5}, {"statue", 12, 5}, {"arch", 18, 9}, {"statue", 22, 8},
			{"lamp", 2, 9}, {"lamp", 17, 11}, {"lamp", 21, 12}, {"lamp", 10, 5}, {"trash", 10, 11}, {"trash", 6, 11}, {"bench", 13, 11}} {
			if _, err := s.Place(ct.DefByID(b.id), b.x, b.y); err != nil {
				println("v3:", b.id, err.Error())
			}
		}
		for x := 1; x < 23; x++ {
			s.LayPath(x, 6)
		}
		for _, c := range [][2]int{{9, 5}, {3, 4}, {3, 5}, {15, 4}, {15, 5}, {20, 4}, {20, 5}, {5, 7}, {8, 7}, {12, 7}, {15, 7}} {
			s.LayPath(c[0], c[1])
		}
		for _, a := range s.Attr {
			switch a.Def.ID {
			case "threering":
				for i, id := range []string{"clown", "juggler", "acrobats", "strong", "stilts", "fire", "trapeze"} {
					pf := s.NewPerformer(id)
					s.Hireable = append(s.Hireable, pf)
					s.Hire(pf)
					s.Book(pf, a, i)
				}
				if has("inspect") {
					g.sel = a
				}
			case "ringtent":
				s.Book(s.Troupe[0], a, 0)
				s.Book(s.Troupe[1], a, 1)
			}
		}
		s.HireSweeper()
		s.HireSweeper()
		s.HireSweeper()
		s.Cash = 6240
	}
	if has("paths") {
		s.Cash += 5000
		for x := 1; x < 23; x++ {
			s.LayPath(x, 7)
		}
		for y := 1; y < 13; y++ {
			s.LayPath(12, y)
			s.LayPath(6, y)
			s.LayPath(18, y)
		}
		s.Cash -= 5000
	}
	if has("night") || has("results") {
		s.OpenGates()
		target := 0.55
		if has("late") {
			target = 0.85
		}
		for s.T < target {
			s.Update(0.05)
		}
		s.Events = nil
		for _, a := range s.Attr {
			if a.Def.ID == "bigtop" || a.Def.ID == "threering" {
				g.sel = a
			}
		}
	}
	if has("results") {
		for s.Phase == ct.PhaseNight {
			s.Update(0.1)
		}
		s.Events = nil
	}
	if has("place") {
		g.placing = ct.DefByID("ringtent")
		g.mx, g.my = 640, 420
		g.fixedMouse = true
	}
	if has("guest") && len(s.Guests) > 0 {
		g.sel = nil
		g.selGuest = s.Guests[len(s.Guests)/2]
	}
	if has("paused") {
		g.paused = true
	}
	if has("tab=food") {
		g.tab = ct.CatFood
	}
	if has("tab=extra") {
		g.tab = ct.CatExtra
	}
	if has("tab=side") {
		g.tab = ct.CatSideshow
	}
}

func (g *game) Layout(ow, oh int) (int, int) {
	k := ebiten.Monitor().DeviceScaleFactor()
	return int(float64(ow) * k), int(float64(oh) * k)
}

func (g *game) Update() error {
	if g.quit {
		return ebiten.Termination
	}
	const dt = 1.0 / 60
	g.clock += dt
	if !g.fixedMouse {
		cx, cy := ebiten.CursorPosition()
		g.mx, g.my = float64(cx)/g.scale-g.offX, float64(cy)/g.scale-g.offY
	}

	g.keys()
	g.mouse()

	if g.screen == "play" && g.sim != nil {
		s := g.sim
		if s.Phase == ct.PhaseNight && g.panel == "" && !g.paused {
			mul := []float64{1, 1, 2, 4}[g.speed]
			s.Update(dt * mul)
		}
		g.events()
		g.cannonSounds()
		switch {
		case s.Phase == ct.PhaseBuild && g.clock-g.lastRunSave > 4:
			g.saveRun() // every few seconds through the morning
		case s.Phase == ct.PhaseBroke && !g.brokeCleared:
			g.deleteRun()
			g.brokeCleared = true
		}
		if g.selGuest != nil && !g.guestAlive(g.selGuest) {
			g.selGuest = nil
		}
	}
	for i := range g.floaters {
		g.floaters[i].t += dt
		g.floaters[i].y -= dt * 22
	}
	for len(g.floaters) > 0 && g.floaters[0].t > 1.4 {
		g.floaters = g.floaters[1:]
	}
	keep := g.toasts[:0]
	for _, t := range g.toasts {
		t.t += dt
		if t.t < 6 {
			keep = append(keep, t)
		}
	}
	g.toasts = keep

	level := 1.0
	if g.screen == "play" && g.sim != nil {
		switch g.sim.Phase {
		case ct.PhaseBuild:
			level = 0.55
		case ct.PhaseNight:
			level = 1
			if g.paused {
				level = 0.35
			}
		default:
			level = 0.45
		}
	}
	snd.update(dt, level)
	return nil
}

func (g *game) guestAlive(gu *ct.Guest) bool {
	for _, x := range g.sim.Guests {
		if x == gu {
			return true
		}
	}
	return false
}

func (g *game) keys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) || (inpututil.IsKeyJustPressed(ebiten.KeyEnter) && ebiten.IsKeyPressed(ebiten.KeyAlt)) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.save.Settings.Music = !g.save.Settings.Music
		snd.setMusic(g.save.Settings.Music)
		g.persist()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		switch {
		case g.placing != nil:
			g.placing = nil
		case g.slotPick >= 0:
			g.slotPick = -1
		case g.panel != "":
			g.panel = ""
		case g.sel != nil || g.selGuest != nil:
			g.sel, g.selGuest, g.slotPick = nil, nil, -1
		case g.screen == "play":
			g.panel = "menu"
		}
	}
	if g.screen != "play" || g.panel != "" || g.sim == nil {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.sim.Phase == ct.PhaseNight {
		g.paused = !g.paused
	}
	for i, k := range []ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3} {
		if inpututil.IsKeyJustPressed(k) {
			g.speed, g.paused = i+1, false
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.placing = nil
	}
	if g.sel != nil && (inpututil.IsKeyJustPressed(ebiten.KeyDelete)) {
		g.sellSelected()
	}
}

func (g *game) mouse() {
	if g.painting {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.placing != nil {
			g.paintPath()
			return
		}
		g.painting = false
		return
	}
	if g.dragging != nil {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.dragging(g.mx)
			return
		}
		g.dragging = nil
		g.persist()
	}
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	for i := len(g.hits) - 1; i >= 0; i-- {
		h := g.hits[i]
		if g.mx >= h.x && g.my >= h.y && g.mx < h.x+h.w && g.my < h.y+h.h {
			if h.drag != nil {
				g.dragging = h.drag
				h.drag(g.mx)
				return
			}
			if h.fn != nil {
				h.fn()
			}
			return
		}
	}
	if g.screen == "play" && g.panel == "" && g.sim != nil {
		g.clickWorld()
	}
}

func (g *game) clickWorld() {
	s := g.sim
	if s.Phase != ct.PhaseBuild && s.Phase != ct.PhaseNight {
		return
	}
	tx, ty, in := g.mouseTile()
	if g.placing != nil {
		if !in {
			return
		}
		d := g.placing
		if d.ID == "path" {
			g.painting = true
			g.paintAdd = !s.IsPath(tx, ty)
			g.paintAt = [2]int{-1, -1}
			g.paintPath()
			return
		}
		a, err := s.Place(d, tx-d.W/2, ty-d.H/2)
		if err != nil {
			snd.play("error")
			return
		}
		snd.play("place")
		x, y := tilePx(float64(a.X)+float64(d.W)/2, float64(a.Y))
		g.floaters = append(g.floaters, floater{x: x, y: y, text: "-" + money(d.Cost)})
		if d.ID == "threering" {
			g.unlock("threering")
			g.note("Three rings, seven acts! Click it to book the biggest show in town.")
		} else if d.ID == "cannon" {
			g.unlock("cannon")
		} else if d.ID == "bigtop" {
			g.unlock("bigtop")
			g.note("The Big Top is up! Click it to book up to five acts.")
		} else if d.Cat == ct.CatTent {
			g.note("Click the " + d.Name + " to book acts for its show.")
		}
		if d.Cat != ct.CatExtra || s.Cash < d.Cost {
			g.placing = nil
			g.sel, g.selGuest, g.slotPick = a, nil, -1
		}
		return
	}
	// a guest under the mouse?
	var best *ct.Guest
	bd := 14.0
	for _, gu := range s.Guests {
		if gu.State == ct.Riding {
			continue
		}
		x, y := tilePx(gu.X, gu.Y)
		if d := math.Hypot(g.mx-x, g.my-(y-12)); d < bd {
			best, bd = gu, d
		}
	}
	g.slotPick = -1
	if best != nil {
		g.selGuest, g.sel = best, nil
		snd.play("click")
		return
	}
	if !in {
		g.sel, g.selGuest = nil, nil
		return
	}
	if a := s.AttrAt(tx, ty); a != nil {
		g.sel, g.selGuest = a, nil
		snd.play("click")
		return
	}
	g.sel, g.selGuest = nil, nil
}

// paintPath lays (or lifts) a path on the tile under the mouse while dragging.
func (g *game) paintPath() {
	s := g.sim
	tx, ty, in := g.mouseTile()
	if !in || g.paintAt == [2]int{tx, ty} {
		return
	}
	g.paintAt = [2]int{tx, ty}
	x, y := tilePx(float64(tx)+.5, float64(ty))
	if g.paintAdd {
		if s.LayPath(tx, ty) == nil {
			snd.play("tab")
			g.floaters = append(g.floaters, floater{x: x, y: y, text: "-" + money(ct.PathCost)})
		}
		return
	}
	if s.LiftPath(tx, ty) {
		snd.play("tab")
		g.floaters = append(g.floaters, floater{x: x, y: y, text: "+" + money(ct.PathCost), good: true})
	}
}

// cannonSounds plays the BOOM as each cannon fires.
func (g *game) cannonSounds() {
	for _, a := range g.sim.Attr {
		if a.Def.ID != "cannon" {
			continue
		}
		prog, running := cannonPhase(a)
		switch {
		case !running:
			g.cannonFired[a.ID] = false
		case prog >= 0.3 && !g.cannonFired[a.ID]:
			g.cannonFired[a.ID] = true
			snd.play("boom")
		}
	}
}

func (g *game) sellSelected() {
	if g.sel == nil {
		return
	}
	a := g.sel
	back := g.sim.Sell(a)
	x, y := tilePx(float64(a.X)+float64(a.Def.W)/2, float64(a.Y))
	g.floaters = append(g.floaters, floater{x: x, y: y, text: "+" + money(back), good: true})
	snd.play("sell")
	g.sel, g.slotPick = nil, -1
}

func (g *game) openGates() {
	if err := g.sim.OpenGates(); err != nil {
		g.toast(err.Error(), "", "warn")
		snd.play("error")
		return
	}
	g.placing = nil
	g.paused = false
	g.slotPick = -1
	g.unlock("opening")
}

func (g *game) toast(title, body, kind string) {
	g.toasts = append(g.toasts, toastMsg{title: title, body: body, kind: kind})
	if len(g.toasts) > 3 {
		g.toasts = g.toasts[len(g.toasts)-3:]
	}
}

func (g *game) note(s string) {
	g.log = append(g.log, s)
	if len(g.log) > 5 {
		g.log = g.log[len(g.log)-5:]
	}
}

// events turns what happened in the simulation into sounds and pop-ups.
func (g *game) events() {
	s := g.sim
	coins := 0
	for _, e := range s.Events {
		x, y := tilePx(e.X, e.Y)
		switch e.Kind {
		case "coin":
			if len(g.floaters) < 30 {
				g.floaters = append(g.floaters, floater{x: x, y: y - 20, text: e.Text, good: true})
			}
			coins++
		case "showstart":
			snd.play("drumroll")
		case "applause":
			if e.V >= 3 {
				snd.play("applause")
			} else {
				snd.play("polite")
			}
			if len(g.floaters) < 30 {
				g.floaters = append(g.floaters, floater{x: x, y: y - 70, stars: math.Max(0.5, e.V), good: e.V >= 2.5})
			}
			if e.V >= 4.5 {
				g.unlock("standing")
			}
			if e.A != nil && e.A.Def.ID == "bigtop" && e.A.Booked() == e.A.Def.Slots {
				g.unlock("fullbill")
			}
			if n := len(e.A.Show.Acts); n > 1 && e.A.Show.Acts[n-1].Note == "Big finish!" {
				g.unlock("finale")
			}
		case "starup":
			snd.play("fanfare")
			g.note(e.Text + " Practice makes perfect.")
			g.unlock("starup")
		case "flub":
			snd.play("flub")
			g.note(e.Text + " Tired acts flub. Rest them or swap them out.")
			g.unlock("mustgoon")
		case "gate":
			snd.play("gate")
			g.note("The gates are open! Night " + itoa(s.Night) + " in " + s.Town.Name + " has begun.")
		}
	}
	if coins > 0 && int(g.clock*10)%3 == 0 {
		snd.play("coin")
	}
	s.Events = s.Events[:0]
	if s.Phase == ct.PhaseResults && !g.resultsSeen() {
		g.onResults()
	}
}

var resultsShown = map[*ct.Sim]int{}

func (g *game) resultsSeen() bool { return resultsShown[g.sim] >= g.sim.Night }

func (g *game) onResults() {
	s := g.sim
	resultsShown[s] = s.Night
	r := s.Reports[len(s.Reports)-1]
	snd.play("close")
	g.sel, g.selGuest, g.placing, g.slotPick = nil, nil, nil, -1
	g.save.Stats.Nights++
	if r.Night > g.save.Stats.BestNight {
		g.save.Stats.BestNight = r.Night
	}
	if s.Cash > g.save.Stats.BestCash {
		g.save.Stats.BestCash = s.Cash
	}
	g.save.Stats.Guests += r.Guests
	if r.Profit() > g.save.Stats.BestProfit {
		g.save.Stats.BestProfit = r.Profit()
	}
	g.persist()
	if r.Guests >= 120 {
		g.unlock("crowd")
	}
	if r.AvgHappy >= 80 {
		g.unlock("pleaser")
	}
	if r.Shows >= 1200 {
		g.unlock("soldout")
	}
	if r.Night >= 3 {
		g.unlock("roadready")
	}
	if r.Night >= 10 {
		g.unlock("longrun")
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	W, H := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())
	s := math.Min(W/logicalW, H/logicalH)
	g.scale = s
	g.offX, g.offY = (W/s-logicalW)/2, (H/s-logicalH)/2
	ss := math.Max(1, math.Min(2, 3200/W))
	cw, ch := int(math.Ceil(W*ss)), int(math.Ceil(H*ss))
	if g.canvas == nil || g.canvas.Bounds().Dx() != cw || g.canvas.Bounds().Dy() != ch {
		if g.canvas != nil {
			g.canvas.Deallocate()
		}
		g.canvas = ebiten.NewImage(cw, ch)
	}
	g.canvas.Fill(cNavy)
	p := pen{dst: g.canvas, s: s * ss, ox: g.offX, oy: g.offY, alpha: 1}
	g.hits = g.hits[:0]
	if g.screen == "title" {
		g.drawTitle(p)
	} else {
		g.drawPlay(p)
	}
	switch g.panel {
	case "menu":
		g.drawMenu(p)
	case "help":
		g.drawHelp(p)
	case "settings":
		g.drawSettings(p)
	case "newrun":
		g.drawNewRun(p)
	case "troupe":
		g.drawTroupe(p)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1/ss, 1/ss)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(g.canvas, op)
}
