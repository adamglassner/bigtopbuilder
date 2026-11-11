package main

import (
	"image/color"
	"math"
	"sort"

	ct "github.com/adamglassner/bigtopbuilder/internal/circus"
)

// The lot: 24 x 14 tiles, laid out left of the side panel and above the build bar.
const (
	TS = 46.0 // tile size
	LX = 44.0 // lot left
	LY = 78.0 // lot top
)

var lotW, lotH = 24 * TS, 14 * TS

func tilePx(x, y float64) (float64, float64) { return LX + x*TS, LY + y*TS }

func (g *game) mouseTile() (int, int, bool) {
	tx, ty := math.Floor((g.mx-LX)/TS), math.Floor((g.my-LY)/TS)
	in := g.mx >= LX && g.my >= LY && g.mx < LX+lotW && g.my < LY+lotH
	return int(tx), int(ty), in
}

var shirts = []color.RGBA{cRed, cTeal, cGold, cBlue, cPurple, cOrange, cPink, hex(0x6fbf5a)}
var skins = []color.RGBA{hex(0xf1c7a1), hex(0xe0a87c), hex(0xc68857), hex(0x8d5a3b), hex(0x5e3a26)}
var prizeCols = []color.RGBA{cPink, cBlue, cGold, cPurple, cMint}

func (g *game) darkness() float64 {
	s := g.sim
	if s == nil || s.Phase != ct.PhaseNight {
		return 0
	}
	return math.Max(0, math.Min(1, (s.T+0.02)/0.22))
}

func (g *game) drawLot(p pen) {
	s := g.sim
	// field around the lot
	p.rect(-200, 0, 1360, 900, hex(0x2c5a37))
	for i := 0; i < 60; i++ { // a few tufts outside the fence
		x := math.Mod(float64(i*197), 1150)
		y := 10 + math.Mod(float64(i*131), 720)
		if x > LX-8 && x < LX+lotW+8 && y > LY-8 && y < LY+lotH+8 {
			continue
		}
		p.circle(x, y, 3, hex(0x2f6340))
	}
	// the road in to the gate
	gx, gy := tilePx(0, float64(s.GateY))
	p.rect(-200, gy+4, 200+LX, TS-8, cDirt)
	// grass
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			px, py := tilePx(float64(x), float64(y))
			c := cGrass
			if (x+y)%2 == 1 {
				c = cGrass2
			}
			p.rect(px, py, TS+0.5, TS+0.5, c)
		}
	}
	// worn paths where guests walk most
	for y := 0; y < s.H; y++ {
		for x := 0; x < s.W; x++ {
			t := s.Traffic[y*s.W+x]
			if t < 4 {
				continue
			}
			k := math.Min(0.8, float64(t)/110)
			px, py := tilePx(float64(x), float64(y))
			p.rrect(px-4, py-4, TS+8, TS+8, 18, withA(cDirt, k))
		}
	}
	// the gate tile is a dirt path
	p.rect(gx, gy+4, TS*2, TS-8, mix(cDirt, cGrass, 0.2))
	if s.Phase == ct.PhaseBuild || g.placing != nil {
		for x := 0; x <= s.W; x++ {
			px, _ := tilePx(float64(x), 0)
			p.line(px, LY, px, LY+lotH, 1, withA(cWhite, 0.07))
		}
		for y := 0; y <= s.H; y++ {
			_, py := tilePx(0, float64(y))
			p.line(LX, py, LX+lotW, py, 1, withA(cWhite, 0.07))
		}
	}
	// litter
	for i, l := range s.Litter {
		if l == 0 {
			continue
		}
		x, y := i%s.W, i/s.W
		for k := 0; k < int(l); k++ {
			px, py := tilePx(float64(x)+0.2+math.Mod(float64(k*37+i*11), 60)/100, float64(y)+0.25+math.Mod(float64(k*53+i*7), 55)/100)
			p.rect(px, py, 5, 4, []color.RGBA{cCream, cRed, cGold}[(k+i)%3])
		}
	}
	// fence
	fc := hex(0x8a5a32)
	p.line(LX, LY, LX+lotW, LY, 3, fc)
	p.line(LX, LY+lotH, LX+lotW, LY+lotH, 3, fc)
	p.line(LX+lotW, LY, LX+lotW, LY+lotH, 3, fc)
	p.line(LX, LY, LX, gy, 3, fc)
	p.line(LX, gy+TS, LX, LY+lotH, 3, fc)
	for x := 0; x <= s.W; x++ {
		px, _ := tilePx(float64(x), 0)
		p.rect(px-2, LY-4, 4, 8, fc)
		p.rect(px-2, LY+lotH-4, 4, 8, fc)
	}
}

// drawGate: the ticket booth and arch, drawn among the other things so it sorts by depth.
func (g *game) drawGate(p pen) {
	gx, gy := tilePx(0, float64(g.sim.GateY))
	// ticket booth just outside the fence
	bx, by := gx-38, gy-6
	p.ellipse(bx+16, by+2, 20, 5, cShadow)
	p.rect(bx, by-34, 32, 36, cCream)
	p.rect(bx+5, by-26, 22, 12, cNavy)
	p.rect(bx-3, by-44, 38, 12, cRed)
	for i := 0; i < 4; i++ {
		p.rect(bx-3+float64(i)*10, by-44, 5, 12, cCream)
	}
	p.text("TICKETS", body(7, 700), bx+16, by-6, alCenter, cRed)
	// arch posts either side of the path
	for _, y := range []float64{gy - 2, gy + TS + 2} {
		p.rect(gx-4, y-46, 8, 46, cRedDk)
		p.circle(gx, y-48, 6, cGold)
	}
	p.line(gx, gy-48, gx, gy+TS-46, 7, cRed)
}

func (g *game) drawAttraction(p pen, d *ct.Def, x, y float64, a *ct.Attraction) {
	w, h := float64(d.W)*TS, float64(d.H)*TS
	cx, by := x+w/2, y+h-4
	t := g.clock
	switch d.ID {
	case "ringtent", "bigtop":
		drawTent(p, d, cx, by, w, a, t)
	case "threering", "cannon":
		p.rrect(x+6, y+6, w-12, h-12, 8, cPanel2)
		p.text("?", display(40), cx, y+h/2, alCenter, cMuted)
	case "funhouse":
		drawFunhouse(p, x, y, w, h, t, a)
	case "mirrors":
		drawMirrors(p, x, y, w, h, t)
	case "fortune":
		drawFortune(p, x, y, w, h, t)
	case "popcorn", "cottoncandy", "hotdogs", "soda":
		drawCart(p, d, x, y, w, h, t)
	case "bench":
		p.ellipse(cx, by-2, 18, 4, cShadow)
		p.rect(cx-17, by-16, 34, 5, cBrown)
		p.rect(cx-17, by-24, 34, 5, hex(0x9a6238))
		p.rect(cx-15, by-11, 3, 10, hex(0x3a2a1a))
		p.rect(cx+12, by-11, 3, 10, hex(0x3a2a1a))
	case "trash":
		p.ellipse(cx, by-2, 11, 4, cShadow)
		p.rrect(cx-9, by-26, 18, 24, 3, hex(0x2f7d5a))
		p.rect(cx-9, by-18, 18, 2, hex(0x235e43))
		p.rrect(cx-11, by-30, 22, 6, 3, hex(0x3d9a70))
	case "lamp":
		p.ellipse(cx, by-2, 9, 3, cShadow)
		p.rect(cx-2.5, by-64, 5, 62, hex(0x2a2a3a))
		p.rect(cx-6, by-6, 12, 6, hex(0x2a2a3a))
		p.rrect(cx-8, by-74, 16, 12, 5, hex(0x2a2a3a))
		p.circle(cx, by-66, 5, cGold)
	}
	if a != nil && d.Cat == ct.CatTent && a.Running {
		// the crowd's excitement, over the tent
		_, top := tentTop(d, cx, by, w)
		bw := 64.0
		ex := a.Excitement()
		p.rrect(cx-bw/2, top-44, bw, 9, 4.5, withA(cNight, 0.85))
		p.rrect(cx-bw/2, top-44, math.Max(9, bw*ex), 9, 4.5, mix(cTeal, cGold, ex))
		if i := a.CurrentAct(); i >= 0 && i < len(a.Show.Acts) {
			name := a.Show.Acts[i].P.Act.Name
			c := cCream
			if a.Flubbed(i) {
				name, c = "Oops!", cBad
			}
			f := body(10, 700)
			tw := textW(name, f) + 12
			p.rrect(cx-tw/2, top-62, tw, 16, 8, withA(cNight, 0.85))
			p.text(name, f, cx, top-54, alCenter, c)
		}
	}
}

func (g *game) drawGuest(p pen, gu *ct.Guest) {
	x, y := tilePx(gu.X, gu.Y)
	k := 1.0
	if gu.Kid {
		k = 0.78
	}
	walking := gu.State == ct.Walking
	ph := 0.0
	if walking {
		ph = math.Sin(g.clock*12+float64(gu.ID)) * 3 * k
	}
	p.ellipse(x, y, 5.5*k, 2.2, cShadow)
	leg := hex(0x2a2f55)
	p.line(x-2*k, y-8*k, x-2*k+ph*0.6, y, 2.4*k, leg)
	p.line(x+2*k, y-8*k, x+2*k-ph*0.6, y, 2.4*k, leg)
	shirt := shirts[gu.Look%len(shirts)]
	p.rrect(x-5*k, y-18*k, 10*k, 11*k, 3*k, shirt)
	skin := skins[gu.Skin%len(skins)]
	p.circle(x, y-22*k, 4.4*k, skin)
	hair := []color.RGBA{hex(0x3a2a1a), hex(0x6a4a2a), hex(0xd9b56a), hex(0x1a1a1a), hex(0xa0522d)}[(gu.Look+gu.Skin)%5]
	p.ellipse(x, y-24.5*k, 4.6*k, 2.4*k, hair)
	if gu.Balloon > 0 {
		bc := prizeCols[(gu.Balloon-1)%len(prizeCols)]
		bx := x + 7*k + math.Sin(g.clock*2+float64(gu.ID))*1.5
		p.line(x+4*k, y-14*k, bx, y-34*k, 1, cTan)
		p.ellipse(bx, y-40*k, 5.5, 6.5, bc)
		p.circle(bx-1.6, y-42*k, 1.5, withA(cWhite, 0.6))
	}
	if g.selGuest == gu {
		p.ring(x, y-12*k, 15, 2, cGold)
	}
}

func (g *game) drawThought(p pen, gu *ct.Guest) {
	if gu.ThoughtT <= 0 || gu.Thought == "" || gu.State == ct.Riding {
		return
	}
	x, y := tilePx(gu.X, gu.Y)
	a := math.Min(1, gu.ThoughtT/0.4)
	q := p.fade(a)
	f := body(11, 600)
	w := textW(gu.Thought, f) + 14
	bx, by := x-w/2, y-50
	c := cCream
	tc := cNavy
	q.rrect(bx, by, w, 20, 9, c)
	q.poly(c, x-4, by+19, x+4, by+19, x, by+26)
	q.text(gu.Thought, f, x, by+10, alCenter, tc)
}

// drawWorld draws the lot and everything on it, sorted back to front.
func (g *game) drawWorld(p pen) {
	s := g.sim
	g.drawLot(p)
	type item struct {
		y    float64
		draw func()
	}
	var items []item
	_, gy := tilePx(0, float64(s.GateY))
	items = append(items, item{gy + TS + 2, func() { g.drawGate(p) }})
	for _, a := range s.Attr {
		a := a
		x, y := tilePx(float64(a.X), float64(a.Y))
		items = append(items, item{y + float64(a.Def.H)*TS, func() {
			q := p
			if g.sel == a {
				q.rrect(x+2, y+2, float64(a.Def.W)*TS-4, float64(a.Def.H)*TS-4, 8, withA(cGold, 0.18))
			}
			g.drawAttraction(q, a.Def, x, y, a)
		}})
	}
	for _, gu := range s.Guests {
		gu := gu
		if gu.State == ct.Riding {
			continue
		}
		_, y := tilePx(gu.X, gu.Y)
		items = append(items, item{y, func() { g.drawGuest(p, gu) }})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].y < items[j].y })
	g.drawBeams(p)
	for _, it := range items {
		it.draw()
	}
	g.drawNight(p)
	if s.Weather == ct.Drizzle && s.Phase == ct.PhaseNight {
		for i := 0; i < 140; i++ {
			fx := math.Mod(float64(i*173)+g.clock*40, 1180)
			fy := math.Mod(float64(i*97)+g.clock*520, 760)
			p.line(fx, fy, fx-3, fy+12, 1, withA(hex(0xb4c8ff), 0.28))
		}
	}
	var placed [][4]float64
	for i := len(s.Guests) - 1; i >= 0; i-- {
		gu := s.Guests[i]
		if gu.ThoughtT <= 0 || gu.Thought == "" || gu.State == ct.Riding {
			continue
		}
		x, y := tilePx(gu.X, gu.Y)
		w := textW(gu.Thought, body(11, 600)) + 14
		r := [4]float64{x - w/2, y - 50, w, 26}
		clash := false
		for _, o := range placed {
			if r[0] < o[0]+o[2] && o[0] < r[0]+r[2] && r[1] < o[1]+o[3] && o[1] < r[1]+r[3] {
				clash = true
				break
			}
		}
		if clash || len(placed) >= 10 {
			continue
		}
		placed = append(placed, r)
		g.drawThought(p, gu)
	}
}

// drawNight darkens the grounds as the evening goes on and lights them with the tents.
func (g *game) drawNight(p pen) {
	d := g.darkness()
	s := g.sim
	if d <= 0 {
		return
	}
	dusk := mix(hex(0x6a3050), hex(0x0a0d38), d)
	p.rect(-200, 0, 1376, 900, withA(dusk, 0.18+0.42*d))
	warm := color.RGBA{255, 200, 120, 255}
	gx, gy := tilePx(0, float64(s.GateY))
	p.glow(gx, gy+TS/2, 2.6*TS, warm, 0.55*d)
	for _, a := range s.Attr {
		if a.Def.Light <= 0 {
			continue
		}
		cx, cy := a.Center()
		x, y := tilePx(cx, cy)
		r := (a.Def.Light + float64(a.Def.W)/2) * TS
		k := 0.45
		if a.Def.ID == "lamp" {
			p.glow(x, y-56, 34, cGold, 1.1*d)
			k = 0.6
		}
		p.glow(x, y, r, warm, k*d)
		g.drawBulbs(p, a, d)
	}
}

// drawBulbs: the strings of lights on tents and stalls, bright after dark.
func (g *game) drawBulbs(p pen, a *ct.Attraction, d float64) {
	x, y := tilePx(float64(a.X), float64(a.Y))
	w, h := float64(a.Def.W)*TS, float64(a.Def.H)*TS
	cx, by := x+w/2, y+h-4
	t := g.clock
	cols := []color.RGBA{cGold, cRed, cMint, cPink}
	bulb := func(bx, byy float64, i int) {
		on := math.Sin(t*3+float64(i)*0.9) > -0.3
		if !on {
			return
		}
		c := cols[i%len(cols)]
		p.circle(bx, byy, 2.2, mix(c, cWhite, 0.4))
		p.glow(bx, byy, 9, c, 0.5*d)
	}
	switch a.Def.ID {
	case "ringtent", "bigtop":
		_, apex := tentRoof(cx, by, w, map[bool]float64{true: 46, false: 34}[a.Def.ID == "bigtop"], map[bool]float64{true: 92, false: 64}[a.Def.ID == "bigtop"], lookOf(a.Def.ID).peaks)
		eave := by - map[bool]float64{true: 46, false: 34}[a.Def.ID == "bigtop"]
		half := w*0.46 + 6
		i := 0
		for _, ap := range apex {
			for _, side := range []float64{-1, 1} {
				ex := ap[0] + side*half*0.9
				if len(apex) == 2 {
					ex = ap[0] + side*half*0.55
				}
				for k := 1; k <= 6; k++ {
					f := float64(k) / 7
					bulb(ap[0]+(ex-ap[0])*f, ap[1]+(eave-ap[1])*f, i)
					i++
				}
			}
		}
	case "funhouse", "mirrors", "fortune":
		for i := 0; i < 6; i++ {
			bulb(x+8+float64(i)*(w-16)/5, by-84, i)
		}
	case "popcorn", "cottoncandy", "hotdogs", "soda":
		for i := 0; i < 5; i++ {
			bulb(x+10+float64(i)*(w-20)/4, by-86, i)
		}
	}
}

// drawBeams: spotlights sweep the sky over every tent with a show on.
func (g *game) drawBeams(p pen) {
	d := g.darkness()
	if d <= 0 {
		return
	}
	for _, a := range g.sim.Attr {
		if a.Def.Cat != ct.CatTent || !a.Running {
			continue
		}
		x, y := tilePx(float64(a.X), float64(a.Y))
		w, h := float64(a.Def.W)*TS, float64(a.Def.H)*TS
		drawSpotlights(p, a.Def, x+w/2, y+h-4, w, g.clock+float64(a.ID), d)
	}
}

// drawGhost shows what's about to be built under the mouse.
func (g *game) drawGhost(p pen) {
	d := g.placing
	if d == nil {
		return
	}
	tx, ty, in := g.mouseTile()
	if !in {
		return
	}
	x0, y0 := tx-d.W/2, ty-d.H/2
	err := g.sim.CanPlace(d, x0, y0)
	ok := err == nil
	col := withA(cMint, 0.35)
	if !ok {
		col = withA(cRed, 0.4)
	}
	for yy := y0; yy < y0+d.H; yy++ {
		for xx := x0; xx < x0+d.W; xx++ {
			px, py := tilePx(float64(xx), float64(yy))
			p.rect(px+1, py+1, TS-2, TS-2, col)
		}
	}
	if d.Visitable() {
		ex, ey := x0+d.W/2, y0+d.H
		px, py := tilePx(float64(ex), float64(ey))
		p.rect(px+6, py+6, TS-12, TS-12, withA(cGold, 0.35))
		p.poly(cGold, px+TS/2, py+10, px+TS/2-8, py+22, px+TS/2+8, py+22)
	}
	x, y := tilePx(float64(x0), float64(y0))
	g.drawAttraction(p.fade(0.75), d, x, y, nil)
	if !ok && err != nil {
		f := body(12, 600)
		w := textW(err.Error(), f) + 20
		mx, my := g.mx, g.my-34
		p.rrect(mx-w/2, my-12, w, 24, 8, cRedDk)
		p.text(err.Error(), f, mx, my, alCenter, cCream)
	}
}
