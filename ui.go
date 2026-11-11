package main

import (
	"image/color"
	"math"

	ct "github.com/adamglassner/bigtopbuilder/internal/circus"
)

type btnStyle int

const (
	bGold btnStyle = iota
	bRed
	bDark
	bTeal
)

func (g *game) hover(x, y, w, h float64) bool {
	return g.dragging == nil && g.mx >= x && g.my >= y && g.mx < x+w && g.my < y+h
}

func (g *game) onClick(x, y, w, h float64, fn func()) {
	g.hits = append(g.hits, hit{x: x, y: y, w: w, h: h, fn: fn})
}

// block stops clicks reaching anything drawn before it (for dialogs).
func (g *game) block() { g.onClick(-1000, -1000, 4000, 4000, func() {}) }

func (g *game) button(p pen, x, y, w, h float64, label string, st btnStyle, enabled bool, fn func()) {
	hov := enabled && g.hover(x, y, w, h)
	bg, fg, edge := cGold, cNavy, mix(cGold, cBlack, 0.3)
	switch st {
	case bRed:
		bg, fg, edge = cRed, cCream, cRedDk
	case bDark:
		bg, fg, edge = cPanel2, cCream, cLine
	case bTeal:
		bg, fg, edge = cTeal, cNavy, hex(0x1f7f74)
	}
	if !enabled {
		bg, fg, edge = mix(cPanel, cNavy, 0.3), mix(cMuted, cNavy, 0.4), cLine
	}
	if hov {
		bg = mix(bg, cWhite, 0.15)
	}
	p.rrect(x, y+3, w, h, 8, edge)
	p.rrect(x, y, w, h, 8, bg)
	fs := math.Min(18, h*0.42)
	p.text(label, head(fs), x+w/2, y+h/2, alCenter, fg)
	if enabled && fn != nil {
		g.onClick(x, y, w, h+3, func() { snd.play("click"); fn() })
	}
}

func (g *game) closeX(p pen, x, y float64, fn func()) {
	g.button(p, x, y, 30, 28, "", bDark, true, fn)
	p.line(x+10, y+9, x+20, y+19, 2, cCream)
	p.line(x+20, y+9, x+10, y+19, 2, cCream)
}

func (g *game) stepper(p pen, x, y float64, label string, val string, dec, inc func()) {
	p.text(label, body(13, 600), x, y+16, alLeft, cTan)
	bx := x + 236 - 120
	g.button(p, bx, y, 34, 32, "", bDark, true, dec)
	p.rect(bx+11, y+15, 12, 2.5, cCream)
	p.rrect(bx+38, y, 44, 32, 6, cNavy)
	p.text(val, head(16), bx+60, y+16, alCenter, cGold)
	g.button(p, bx+86, y, 34, 32, "", bDark, true, inc)
	p.rect(bx+97, y+15, 12, 2.5, cCream)
	p.rect(bx+101.75, y+10.25, 2.5, 12, cCream)
}

func (g *game) bar(p pen, x, y, w float64, label string, v float64, c color.RGBA) {
	p.text(label, body(12, 500), x, y+6, alLeft, cTan)
	p.rrect(x+84, y, w-84, 12, 6, cNavy)
	p.rrect(x+84, y, math.Max(12, (w-84)*v/100), 12, 6, c)
}

func stars(p pen, x, y, size, v float64) {
	for i := 0; i < 5; i++ {
		cx := x + float64(i)*size*1.15
		fill := math.Max(0, math.Min(1, v-float64(i)))
		star(p, cx, y, size/2, mix(cNavy, cLine, 0.6))
		if fill > 0 {
			q := p
			q.alpha *= 0.35 + 0.65*fill
			star(q, cx, y, size/2*(0.5+0.5*fill), cGold)
		}
	}
}

func star(p pen, cx, cy, r float64, c color.RGBA) {
	pts := []float64{}
	for i := 0; i < 10; i++ {
		rr := r
		if i%2 == 1 {
			rr = r * 0.45
		}
		a := -math.Pi/2 + float64(i)*math.Pi/5
		pts = append(pts, cx+math.Cos(a)*rr, cy+math.Sin(a)*rr)
	}
	p.poly(c, pts...)
}

func (g *game) drawPlay(p pen) {
	s := g.sim
	g.drawWorld(p)
	if g.placing != nil {
		g.drawGhost(p)
	}
	for _, f := range g.floaters {
		c := cGold
		if !f.good {
			c = cBad
		}
		q := p.fade(math.Max(0, 1-f.t/1.4))
		if f.stars > 0 {
			n := int(math.Ceil(f.stars))
			w := float64(n) * 15
			q.rrect(f.x-w/2-6, f.y-10, w+12, 20, 10, withA(cNight, 0.8))
			for i := 0; i < n; i++ {
				fill := math.Min(1, f.stars-float64(i))
				star(q, f.x-w/2+7.5+float64(i)*15, f.y, 6.5*(0.55+0.45*fill), c)
			}
			continue
		}
		q.text(f.text, head(15), f.x+1, f.y+1, alCenter, cNight)
		q.text(f.text, head(15), f.x, f.y, alCenter, c)
	}
	g.drawHUD(p)
	g.drawSide(p)
	g.drawBuildBar(p)
	if g.paused && s.Phase == ct.PhaseNight && g.panel == "" {
		p.rrect(LX+lotW/2-90, LY+16, 180, 40, 10, withA(cNight, 0.85))
		p.text("PAUSED", head(20), LX+lotW/2, LY+36, alCenter, cGold)
	}
	switch s.Phase {
	case ct.PhaseResults:
		g.drawResults(p)
	case ct.PhaseBroke:
		g.drawBroke(p)
	}
	g.drawToasts(p)
}

func (g *game) drawHUD(p pen) {
	s := g.sim
	p.rect(-200, 0, 1840, 64, cNight)
	p.rect(-200, 62, 1840, 3, cRed)
	p.text("Big Top Builder", display(15), 26, 22, alLeft, cGold)
	p.text("CLASSIC", head(12), 27, 44, alLeft, cCream)
	p.rrect(180, 14, 52, 20, 5, cRed)
	p.text("DEMO", head(12), 206, 24, alCenter, cCream)

	p.text(s.Town.Name, body(14, 700), 252, 22, alLeft, cCream)
	sub := "Night " + itoa(s.Night)
	switch s.Phase {
	case ct.PhaseBuild:
		sub += " · Daytime: build, hire and plan the shows"
	case ct.PhaseNight:
		h, m := s.Clock()
		mm := itoa(m)
		if m < 10 {
			mm = "0" + mm
		}
		sub += " · " + itoa(h-12) + ":" + mm + " PM"
		if s.Closing() {
			sub += " · Closing time"
		}
	default:
		sub += " · Closed"
	}
	p.text(sub, body(12, 400), 252, 43, alLeft, cMuted)

	x := 560.0
	stat := func(label, val string, c color.RGBA, w float64) {
		p.text(label, body(11, 400), x, 20, alLeft, cMuted)
		p.text(val, head(17), x, 42, alLeft, c)
		x += w
	}
	cashC := cGold
	if s.Cash < 0 {
		cashC = cBad
	}
	stat("Cash", money(s.Cash), cashC, 112)
	stat("Guests", itoa(len(s.Guests)), cCream, 76)
	hap := "–"
	if len(s.Guests) > 0 {
		hap = itoa(int(s.AvgHappy())) + "%"
	}
	stat("Happiness", hap, cMint, 96)
	p.text("Reputation", body(11, 400), x, 20, alLeft, cMuted)
	stars(p, x+8, 42, 15, s.Rep)
	x += 104
	stat("Weather", ct.WeatherNames[s.Weather], cCream, 84)

	// speed controls
	bx := 1050.0
	if s.Phase == ct.PhaseNight {
		for i := 0; i < 4; i++ {
			i := i
			on := (i == 0 && g.paused) || (i > 0 && !g.paused && g.speed == i)
			st := bDark
			if on {
				st = bGold
			}
			defer func() {
				ic := cCream
				if on {
					ic = cNavy
				}
				cx, cy := bx+float64(i)*48+22, 31.0
				if i == 0 {
					p.rect(cx-6, cy-7, 4, 14, ic)
					p.rect(cx+2, cy-7, 4, 14, ic)
					return
				}
				for k := 0; k < i; k++ {
					tx := cx - float64(i)*4.5 + float64(k)*9
					p.poly(ic, tx-3, cy-7, tx+6, cy, tx-3, cy+7)
				}
			}()
			g.button(p, bx+float64(i)*48, 14, 44, 34, "", st, true, func() {
				if i == 0 {
					g.paused = !g.paused
				} else {
					g.speed, g.paused = i, false
				}
			})
		}
	}
	g.button(p, 1252, 14, 116, 34, "Troupe ("+itoa(len(s.Troupe))+")", bTeal, true, func() { g.panel = "troupe"; g.slotPick = -1 })
	g.button(p, 1376, 14, 52, 34, "Menu", bDark, true, func() { g.panel = "menu" })
}

func (g *game) drawSide(p pen) {
	x, y, w, h := 1164.0, LY, 264.0, lotH
	p.rrect(x, y, w, h, 12, cPanel)
	p.rrectStroke(x, y, w, h, 12, 1, cLine)
	in := x + 16
	iw := w - 32
	switch {
	case g.sel != nil && g.sel.Def.Cat == ct.CatTent && g.slotPick >= 0:
		g.drawSlotPicker(p, in, y+18, iw)
	case g.sel != nil && g.sel.Def.Cat == ct.CatTent:
		g.drawTentInspector(p, in, y+18, iw)
	case g.sel != nil:
		g.drawInspector(p, in, y+18, iw)
	case g.selGuest != nil:
		g.drawGuestPanel(p, in, y+18, iw)
	default:
		g.drawOverview(p, in, y+18, iw)
	}
}

func catLabel(c ct.Cat) string {
	return []string{"Tent", "Sideshow", "Food cart", "Extra"}[c]
}

func (g *game) drawInspector(p pen, x, y, w float64) {
	s := g.sim
	a := g.sel
	d := a.Def
	p.text(catLabel(d.Cat), body(11, 600), x, y+4, alLeft, cMint)
	p.text(d.Name, display(22), x, y+30, alLeft, cGold)
	g.closeX(p, x+w-30, y-6, func() { g.sel = nil })
	y += 54
	if d.Visitable() {
		status, sc := "Open", cMint
		switch {
		case a.Running && d.Cat == ct.CatSideshow:
			status = "Guests inside"
		case s.Phase != ct.PhaseNight:
			status, sc = "Ready for tonight", cTan
		}
		p.text(status, body(13, 600), x, y, alLeft, sc)
		if s.Phase == ct.PhaseNight {
			p.text(itoa(len(a.Queue))+" in line", body(12, 400), x+w, y, alRight, cMuted)
		}
		y += 22
	}
	if d.Cat != ct.CatExtra {
		box := func(bx float64, label, val string) {
			p.rrect(bx, y, (w-12)/3, 46, 6, cNavy)
			p.text(label, body(10, 400), bx+8, y+13, alLeft, cMuted)
			p.text(val, head(16), bx+8, y+32, alLeft, cCream)
		}
		bw := (w-12)/3 + 6
		box(x, "Tonight", money(a.Night.Revenue))
		cust := "Visitors"
		if d.Cat == ct.CatFood {
			cust = "Sold"
		}
		box(x+bw, cust, itoa(a.Night.Customers))
		box(x+2*bw, "All time", money(a.Total.Revenue))
		y += 60
		g.stepper(p, x, y, "Ticket price", money(a.Price), func() { s.SetPrice(a, a.Price-1) }, func() { s.SetPrice(a, a.Price+1) })
		y += 38
		if a.Price > d.FairMax {
			p.text("Guests think over "+money(d.FairMax)+" is too much", body(11, 500), x, y+4, alLeft, cBad)
		} else {
			p.text("Fair price: up to "+money(d.FairMax), body(11, 400), x, y+4, alLeft, cMuted)
		}
		y += 24
	}
	y += 6
	y = p.para(d.Blurb, body(12, 400), x, y, w, 17, cTan) + 6
	if d.Upkeep > 0 {
		p.text("Upkeep "+money(d.Upkeep)+"/night", body(11, 400), x, y, alLeft, cMuted)
		y += 20
	}
	by := LY + lotH - 62
	back := int(float64(d.Cost) * ct.SellBack)
	g.button(p, x, by, w, 40, "Sell for "+money(back), bRed, true, func() { g.sellSelected() })
}

func (g *game) drawGuestPanel(p pen, x, y, w float64) {
	gu := g.selGuest
	k := "Guest"
	if gu.Kid {
		k = "Young guest"
	}
	p.text(k, body(11, 600), x, y+4, alLeft, cMint)
	p.text(gu.Name, display(24), x, y+30, alLeft, cGold)
	g.closeX(p, x+w-30, y-6, func() { g.selGuest = nil })
	y += 56
	p.text(gu.Mood(), head(18), x, y, alLeft, cCream)
	p.text(money(gu.Money)+" left", body(13, 500), x+w, y, alRight, cGold)
	y += 22
	p.text(gu.Doing(), body(12, 400), x, y, alLeft, cTan)
	y += 26
	g.bar(p, x, y, w, "Happiness", gu.Happy, cMint)
	y += 22
	g.bar(p, x, y, w, "Hunger", gu.Hunger, cOrange)
	y += 22
	g.bar(p, x, y, w, "Thirst", gu.Thirst, cBlue)
	y += 22
	g.bar(p, x, y, w, "Energy", gu.Energy, cGold)
	y += 34
	p.text("Spent tonight: "+money(gu.Spent), body(12, 400), x, y, alLeft, cMuted)
	y += 20
	if gu.Balloon > 0 {
		p.text("Got a souvenir balloon", body(12, 400), x, y, alLeft, cPink)
		y += 20
	}
	if gu.ThoughtT > 0 && gu.Thought != "" {
		y += 8
		p.rrect(x, y, w, 40, 10, cCream)
		p.text("“"+gu.Thought+"”", body(13, 600), x+w/2, y+20, alCenter, cNavy)
	}
}

func (g *game) drawOverview(p pen, x, y, w float64) {
	s := g.sim
	p.text("Your circus", body(11, 600), x, y+4, alLeft, cMint)
	p.text(s.Town.Name, display(24), x, y+30, alLeft, cGold)
	y += 56
	switch s.Phase {
	case ct.PhaseBuild:
		y = p.para("Raise tents, book your acts and set up food and sideshows. The gates open at 6 PM.", body(12, 400), x, y, w, 17, cTan) + 10
		if s.Night == 1 {
			has := map[ct.Cat]bool{}
			booked := false
			for _, a := range s.Attr {
				has[a.Def.Cat] = true
				if a.Def.Cat == ct.CatTent && a.Booked() > 0 {
					booked = true
				}
			}
			item := func(done bool, t string) {
				c := cLine
				if done {
					c = cTeal
				}
				p.rrect(x, y-8, 16, 16, 4, c)
				if done {
					p.line(x+4, y, x+7, y+4, 2, cCream)
					p.line(x+7, y+4, x+13, y-4, 2, cCream)
				}
				tc := cCream
				if done {
					tc = cMuted
				}
				p.text(t, body(13, 500), x+26, y, alLeft, tc)
				y += 26
			}
			p.text("Getting started", head(15), x, y, alLeft, cCream)
			y += 26
			item(has[ct.CatTent], "Raise a tent")
			item(booked, "Click it and book your acts")
			item(has[ct.CatFood], "Set up a food cart")
			item(has[ct.CatSideshow], "Add a sideshow")
			item(false, "Open the gates!")
			y += 6
		} else {
			p.text("Forecast tonight: "+ct.WeatherNames[s.Weather], body(13, 600), x, y, alLeft, cCream)
			y += 22
			if s.Weather == ct.Drizzle {
				y = p.para("Rain keeps some folks home and dampens spirits.", body(12, 400), x, y, w, 17, cMuted) + 6
			}
			if len(s.Reports) > 0 {
				r := s.Reports[len(s.Reports)-1]
				p.text("Last night: "+money(r.Profit())+" profit", body(13, 500), x, y, alLeft, cTan)
				y += 22
			}
			y = p.para("Your troupe is rested, and new performers are looking for work.", body(12, 400), x, y, w, 17, cMuted) + 8
		}
		p.text("Troupe: "+itoa(len(s.Troupe))+" · wages "+money(s.Wages())+"/night", body(12, 600), x, y+4, alLeft, cTan)
		y += 18
		g.button(p, x, y, w, 36, "Hire & Manage Troupe", bTeal, true, func() { g.panel = "troupe" })
	case ct.PhaseNight:
		r := s.Report
		row := func(l, v string) {
			p.text(l, body(13, 400), x, y, alLeft, cTan)
			p.text(v, head(16), x+w, y, alRight, cCream)
			y += 24
		}
		row("Guests tonight", itoa(r.Guests))
		row("On the grounds now", itoa(len(s.Guests)))
		row("Taken in", money(r.Income()))
		row("Shows performed", itoa(r.ShowsRun))
		p.text("Average show", body(13, 400), x, y, alLeft, cTan)
		stars(p, x+w-74, y, 13, r.AvgStars())
		y += 30
		y = p.para("Click a tent to watch its show and change the lineup. Tired acts flub, so swap them out.", body(12, 400), x, y, w, 17, cMuted) + 8
	}
	if s.Phase == ct.PhaseBuild || s.Phase == ct.PhaseNight {
		g.stepper(p, x, LY+lotH-66, "Gate admission", money(s.Gate), func() { s.SetGate(s.Gate - 1) }, func() { s.SetGate(s.Gate + 1) })
		p.text("Higher admission, fewer guests.", body(11, 400), x, LY+lotH-22, alLeft, cMuted)
	}
}

func (g *game) drawBuildBar(p pen) {
	s := g.sim
	x0, y0 := LX, 736.0
	p.rrect(x0, y0, 1384, 156, 12, cPanel)
	p.rrectStroke(x0, y0, 1384, 156, 12, 1, cLine)
	// tabs
	tx := x0 + 12
	for c := ct.CatTent; c <= ct.CatExtra; c++ {
		c := c
		label := ct.CatNames[c]
		w := textW(label, head(14)) + 28
		on := g.tab == c
		bg := cNavy
		fg := cMuted
		if on {
			bg, fg = cRed, cCream
		} else if g.hover(tx, y0+10, w, 28) {
			fg = cCream
		}
		p.rrect(tx, y0+10, w, 28, 14, bg)
		p.text(label, head(14), tx+w/2, y0+24, alCenter, fg)
		g.onClick(tx, y0+10, w, 28, func() { g.tab = c; snd.play("tab") })
		tx += w + 8
	}
	// cards
	cx := x0 + 12
	for _, d := range ct.Defs {
		if d.Cat != g.tab {
			continue
		}
		d := d
		cw, ch := 136.0, 98.0
		cy := y0 + 46
		on := g.placing == d
		hov := g.hover(cx, cy, cw, ch)
		bg := cNavy
		if hov && !d.Locked {
			bg = mix(cNavy, cPanel2, 0.7)
		}
		p.rrect(cx, cy, cw, ch, 10, bg)
		if on {
			p.rrectStroke(cx, cy, cw, ch, 10, 3, cGold)
		}
		// a little picture
		thumb := p
		thumb.alpha = 1
		if d.Locked {
			thumb.alpha = 0.35
		}
		g.drawThumb(thumb, d, cx+cw/2, cy+58)
		p.rect(cx, cy+60, cw, 38, withA(cNavy, 0.82))
		nameC := cCream
		if d.Locked {
			nameC = cMuted
		}
		p.text(d.Name, head(13), cx+8, cy+72, alLeft, nameC)
		if d.Locked {
			p.text("Full game", body(11, 500), cx+8, cy+88, alLeft, cMuted)
			p.rrect(cx+cw-40, cy+6, 34, 16, 4, cLine)
			p.text("SOON", head(9), cx+cw-23, cy+14, alCenter, cMuted)
		} else {
			cc := cGold
			if s.Cash < d.Cost {
				cc = cBad
			}
			p.text(money(d.Cost), body(12, 600), cx+8, cy+88, alLeft, cc)
		}
		if !d.Locked {
			g.onClick(cx, cy, cw, ch, func() {
				snd.play("click")
				if g.placing == d {
					g.placing = nil
				} else {
					g.placing, g.sel, g.selGuest = d, nil, nil
				}
			})
		}
		cx += cw + 8
	}
	// notes
	nx := 776.0
	p.text("Ringmaster's notes", head(14), nx, y0+24, alLeft, cTan)
	ny := y0 + 50
	for i := len(g.log) - 1; i >= 0 && ny < y0+150; i-- {
		c := cMuted
		if i == len(g.log)-1 {
			c = cCream
		}
		for _, l := range wrap(g.log[i], body(12, 400), 360) {
			p.text(l, body(12, 400), nx, ny, alLeft, c)
			ny += 17
		}
		ny += 4
	}
	if g.placing != nil && g.placing.ID == "path" {
		p.text("Click and drag to lay paths, or drag over a path to pick it up · Esc to stop", body(12, 600), nx, y0+146, alLeft, cGold)
	} else if g.placing != nil {
		p.text("Click the grounds to build · Right-click or Esc to cancel", body(12, 600), nx, y0+146, alLeft, cGold)
	}
	// the big button
	bx, bw := 1164.0, 252.0
	switch s.Phase {
	case ct.PhaseBuild:
		g.button(p, bx, y0+34, bw, 64, "OPEN THE GATES", bRed, true, g.openGates)
		p.text("Tonight: "+ct.WeatherNames[s.Weather]+" · 6 PM to 11 PM", body(12, 400), bx+bw/2, y0+118, alCenter, cMuted)
		p.text("Troupe wages: "+money(s.Wages()), body(12, 400), bx+bw/2, y0+136, alCenter, cMuted)
	case ct.PhaseNight:
		p.text("Tonight", head(14), bx, y0+24, alLeft, cTan)
		h, m := s.Clock()
		mm := itoa(m)
		if m < 10 {
			mm = "0" + mm
		}
		p.text(itoa(h-12)+":"+mm+" PM", display(30), bx, y0+66, alLeft, cGold)
		p.rrect(bx, y0+96, bw, 14, 7, cNavy)
		p.rrect(bx, y0+96, math.Max(14, bw*math.Min(1, s.T)), 14, 7, cGold)
		msg := "The last show starts before 11 PM"
		if s.Closing() {
			msg = "Closing up. Guests are heading home."
		}
		p.text(msg, body(12, 400), bx, y0+128, alLeft, cMuted)
	}
}

// drawThumb draws a small version of an attraction for its build card.
func (g *game) drawThumb(p pen, d *ct.Def, cx, by float64) {
	k := 0.5
	if d.W >= 3 {
		k = 0.38
	}
	if d.ID == "bigtop" {
		k = 0.29
	}
	if d.W == 1 {
		k = 0.8
	}
	q := p
	q.s = p.s * k
	// map (cx, by) in p's space to q's space
	w, h := float64(d.W)*TS, float64(d.H)*TS
	q.ox = (p.ox+cx)/k - w/2
	q.oy = (p.oy+by)/k - h
	g.drawAttraction(q, d, 0, 0, nil)
}

func (g *game) drawToasts(p pen) {
	y := LY + 12
	modal := g.panel != "" || (g.sim != nil && g.sim.Phase != ct.PhaseBuild && g.sim.Phase != ct.PhaseNight)
	if modal {
		y = 8
	}
	for i := len(g.toasts) - 1; i >= 0; i-- {
		t := g.toasts[i]
		if modal && t.kind != "ach" {
			continue
		}
		a := math.Min(1, math.Min(t.t/0.2, (6-t.t)/0.5))
		q := p.fade(a)
		w := 420.0
		x := LX + lotW/2 - w/2
		h := 52.0
		if t.body == "" {
			h = 36
		}
		bg := cPanel
		edge := cGold
		if t.kind == "warn" {
			edge = cRed
		}
		q.rrect(x, y, w, h, 10, withA(bg, 0.96))
		q.rrect(x, y, 6, h, 3, edge)
		q.text(t.title, head(15), x+18, y+18, alLeft, cCream)
		if t.body != "" {
			q.text(t.body, body(12, 400), x+18, y+38, alLeft, cTan)
		}
		if t.a != nil {
			ta := t.a
			g.onClick(x, y, w, h, func() { g.sel, g.selGuest, g.placing = ta, nil, nil })
		}
		y += h + 8
	}
}
