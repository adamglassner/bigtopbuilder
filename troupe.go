package main

import (
	"image/color"
	"math"

	ct "github.com/adamglassner/bigtopbuilder/internal/circus"
)

var styleCol = [...]color.RGBA{ct.Laughs: cGold, ct.Wonder: cMint, ct.Thrills: cRed}

// tentLabel names where a performer is booked.
func tentLabel(a *ct.Attraction) string {
	if a == nil {
		return "Not booked"
	}
	if a.Def.ID == "bigtop" {
		return "In the Big Top"
	}
	return "In a Ring Tent"
}

func (g *game) energyBar(p pen, x, y, w float64, e float64) {
	p.rrect(x, y, w, 7, 3.5, cNavy)
	c := cMint
	if e < ct.TiredAt {
		c = cBad
	} else if e < 65 {
		c = cGold
	}
	p.rrect(x, y, math.Max(7, w*e/100), 7, 3.5, c)
}

func smallStars(p pen, x, y float64, n int) {
	for i := 0; i < 5; i++ {
		c := mix(cNavy, cLine, 0.6)
		if i < n {
			c = cGold
		}
		star(p, x+float64(i)*11, y, 4.6, c)
	}
}

// portrait draws a performer in a little framed card.
func (g *game) portrait(p pen, pf *ct.Performer, x, y, w, h float64) {
	p.rrect(x, y, w, h, 8, mix(cNavy, styleCol[pf.Act.Style], 0.12))
	tall := map[string]float64{"clown": 62, "stilts": 118, "juggler": 90, "strong": 84, "acrobats": 128, "fire": 64, "trapeze": 128}[pf.Act.ID]
	if tall == 0 {
		tall = 64
	}
	k := (h - 8) / tall
	drawPerformer(p, pf.Act.ID, x+w/2, y+h-4, k, g.clock, pf.Look, pf.Skin)
}

/* ---------- the troupe dialog ---------- */

func (g *game) drawTroupe(p pen) {
	s := g.sim
	x, y := g.dialog(p, 1080, 700)
	p.text("The Troupe", display(34), x+40, y+46, alLeft, cGold)
	p.text(itoa(len(s.Troupe))+" of "+itoa(ct.MaxTroupe)+" performers · wages "+money(s.Wages())+" a night · cash "+money(s.Cash), body(14, 500), x+42, y+80, alLeft, cTan)
	g.button(p, x+1080-150, y+22, 120, 40, "Done", bGold, true, func() { g.panel = "" })

	// your performers, two columns of cards
	lx, ly := x+30, y+108
	p.text("Your performers", head(17), lx, ly, alLeft, cMint)
	cw, ch := 318.0, 82.0
	for i, pf := range s.Troupe {
		pf := pf
		cx := lx + float64(i%2)*(cw+12)
		cy := ly + 18 + float64(i/2)*(ch+8)
		p.rrect(cx, cy, cw, ch, 10, cPanel2)
		g.portrait(p, pf, cx+6, cy+6, 62, ch-12)
		tx := cx + 78
		p.text(fit(pf.Name, head(15), cw-150), head(15), tx, cy+16, alLeft, cCream)
		p.text(pf.Act.Name+" · "+ct.StyleNames[pf.Act.Style], body(11, 600), tx, cy+34, alLeft, styleCol[pf.Act.Style])
		smallStars(p, tx+5, cy+50, pf.Stars())
		p.text(money(pf.Wage())+"/night", body(11, 400), cx+cw-12, cy+50, alRight, cMuted)
		g.energyBar(p, tx, cy+62, 110, pf.Energy)
		lab, lc := tentLabel(pf.Tent), cTan
		if pf.Tent == nil {
			lc = cMuted
		}
		if pf.Tired() {
			lab, lc = "Tired!", cBad
		}
		p.text(lab, body(11, 500), tx+118, cy+66, alLeft, lc)
		g.button(p, cx+cw-62, cy+8, 52, 24, "Let go", bDark, s.Phase != ct.PhaseNight || pf.Tent == nil || !pf.Tent.Running, func() {
			s.Dismiss(pf)
			g.note(pf.Name + " has left the troupe.")
		})
	}
	if len(s.Troupe) == 0 {
		p.para("Nobody yet! Hire some performers from the noticeboard.", body(14, 400), lx, ly+40, 600, 20, cMuted)
	}

	// the noticeboard
	rx, ry := x+700, y+108
	rw := 350.0
	p.text("Looking for work", head(17), rx, ry, alLeft, cGold)
	p.text("New faces every morning", body(11, 400), rx+rw, ry, alRight, cMuted)
	for i, pf := range s.Hireable {
		pf := pf
		cy := ry + 18 + float64(i)*(ch+8)
		p.rrect(rx, cy, rw, ch, 10, cNavy)
		p.rrectStroke(rx, cy, rw, ch, 10, 1, cLine)
		g.portrait(p, pf, rx+6, cy+6, 62, ch-12)
		tx := rx + 78
		p.text(pf.Name, head(15), tx, cy+16, alLeft, cCream)
		p.text(pf.Act.Name+" · "+ct.StyleNames[pf.Act.Style], body(11, 600), tx, cy+34, alLeft, styleCol[pf.Act.Style])
		smallStars(p, tx+5, cy+50, pf.Stars())
		note := money(pf.Wage()) + "/night"
		if pf.Act.Tall {
			note += " · Big Top only"
		}
		p.text(note, body(11, 400), tx, cy+68, alLeft, cMuted)
		can := s.Cash >= pf.HireCost() && len(s.Troupe) < ct.MaxTroupe
		g.button(p, rx+rw-104, cy+22, 94, 36, "Hire "+money(pf.HireCost()), bGold, can, func() {
			if err := s.Hire(pf); err != nil {
				g.toast(err.Error(), "", "warn")
				snd.play("error")
				return
			}
			snd.play("cash")
			g.note(pf.Name + " the " + pf.Act.Name + " joined the troupe!")
			g.unlock("hire")
		})
	}
	if len(s.Hireable) == 0 {
		p.para("Everyone's been signed. More performers turn up tomorrow morning.", body(13, 400), rx, ry+40, rw, 19, cMuted)
	}

	// how shows work
	hy := y + 700 - 74
	p.line(x+30, hy-12, x+1050, hy-12, 1, cLine)
	p.text("Building a good show", head(14), x+30, hy+6, alLeft, cMint)
	p.para("Open with Laughs to warm the crowd up, don't put two acts of the same kind back to back, and close with Thrills for a big finish. "+
		"Every show tires an act out; tired acts flub. Everyone's rested by morning.", body(12, 400), x+30, hy+26, 1020, 17, cTan)
}

/* ---------- the tent inspector ---------- */

func (g *game) drawTentInspector(p pen, x, y, w float64) {
	s := g.sim
	a := g.sel
	d := a.Def
	p.text("Tent · "+itoa(a.Booked())+" of "+itoa(d.Slots)+" acts booked", body(11, 600), x, y+4, alLeft, cMint)
	p.text(d.Name, display(22), x, y+30, alLeft, cGold)
	g.closeX(p, x+w-30, y-6, func() { g.sel = nil })
	y += 54
	status, sc := "Ready for tonight", cTan
	switch {
	case a.Booked() == 0:
		status, sc = "No acts booked: no show!", cBad
	case a.Running:
		status, sc = "Show on: act "+itoa(a.CurrentAct()+1)+" of "+itoa(len(a.Show.Acts)), cGold
	case s.Phase == ct.PhaseNight:
		status, sc = "Letting the audience in", cMint
	}
	p.text(status, body(13, 600), x, y, alLeft, sc)
	if s.Phase == ct.PhaseNight {
		p.text(itoa(len(a.Queue))+" waiting", body(12, 400), x+w, y, alRight, cMuted)
	}
	y += 14
	if a.Running {
		ex := a.Excitement()
		p.rrect(x, y, w, 8, 4, cNavy)
		p.rrect(x, y, math.Max(8, w*ex), 8, 4, mix(cTeal, cGold, ex))
	}
	y += 14
	box := func(bx float64, label, val string) {
		p.rrect(bx, y, (w-12)/3, 46, 6, cNavy)
		p.text(label, body(10, 400), bx+8, y+13, alLeft, cMuted)
		p.text(val, head(16), bx+8, y+32, alLeft, cCream)
	}
	bw := (w-12)/3 + 6
	box(x, "Tonight", money(a.Night.Revenue))
	box(x+bw, "Audience", itoa(a.Night.Customers))
	box(x+2*bw, "Shows", itoa(a.Night.Shows))
	y += 56
	g.stepper(p, x, y, "Ticket price", money(a.Price), func() { s.SetPrice(a, a.Price-1) }, func() { s.SetPrice(a, a.Price+1) })
	y += 36
	if a.Price > d.FairMax {
		p.text("Guests think over "+money(d.FairMax)+" is too much", body(11, 500), x, y+4, alLeft, cBad)
	} else {
		p.text("Fair price: up to "+money(d.FairMax), body(11, 400), x, y+4, alLeft, cMuted)
	}
	y += 26

	sc2 := a.Score()
	p.text("The show", head(15), x, y, alLeft, cCream)
	stars(p, x+w-74, y, 13, sc2.Stars)
	y += 18
	byIdx := map[*ct.Performer]ct.ActScore{}
	for _, as := range sc2.Acts {
		byIdx[as.P] = as
	}
	cur := -1
	if a.Running {
		cur = a.CurrentAct()
	}
	order := 0
	rh := 44.0
	for i, pf := range a.Lineup {
		i, pf := i, pf
		ry := y + float64(i)*(rh+4)
		hov := g.hover(x, ry, w-70, rh)
		bg := cNavy
		if hov {
			bg = mix(cNavy, cPanel2, 0.7)
		}
		p.rrect(x, ry, w, rh, 8, bg)
		p.circle(x+12, ry+rh/2, 9, cPanel2)
		p.text(itoa(i+1), head(11), x+12, ry+rh/2, alCenter, cTan)
		if pf == nil {
			p.text("+ Book an act", body(12, 600), x+30, ry+rh/2, alLeft, cGold)
			g.onClick(x, ry, w, rh, func() { g.slotPick = i; snd.play("click") })
			continue
		}
		if pf.Tent == a && order == cur && a.Running {
			p.rrectStroke(x, ry, w, rh, 8, 2, cGold)
		}
		order++
		g.portrait(p, pf, x+24, ry+3, 30, rh-6)
		p.text(fit(pf.Name, body(12, 700), w-130), body(12, 700), x+60, ry+14, alLeft, cCream)
		as := byIdx[pf]
		nc := cMuted
		note := pf.Act.Name
		if as.Note != "" {
			note = as.Note
			nc = cMint
			if !as.Good {
				nc = cBad
			}
		}
		p.text(note, body(10, 500), x+60, ry+31, alLeft, nc)
		g.onClick(x, ry, w-70, rh, func() { g.slotPick = i; snd.play("click") })
		// running order controls
		bx := x + w - 66
		arrow := func(bx float64, up, ok bool, fn func()) {
			g.button(p, bx, ry+12, 20, 20, "", bDark, ok, fn)
			c := cCream
			if !ok {
				c = cMuted
			}
			if up {
				p.poly(c, bx+10, ry+18, bx+15, ry+25, bx+5, ry+25)
			} else {
				p.poly(c, bx+5, ry+19, bx+15, ry+19, bx+10, ry+26)
			}
		}
		arrow(bx, true, i > 0, func() { s.MoveSlot(a, i, -1) })
		arrow(bx+22, false, i < len(a.Lineup)-1, func() { s.MoveSlot(a, i, 1) })
		g.button(p, bx+44, ry+12, 20, 20, "", bDark, true, func() { s.ClearSlot(a, i) })
		p.line(bx+50, ry+17, bx+58, ry+25, 1.8, cCream)
		p.line(bx+58, ry+17, bx+50, ry+25, 1.8, cCream)
	}
	y += float64(len(a.Lineup))*(rh+4) + 6
	if sc2.Total > 0 {
		p.text("Each guest pays "+money(a.Price)+" · upkeep "+money(d.Upkeep)+"/night", body(11, 400), x, y+4, alLeft, cMuted)
		y += 18
	}
	if y < LY+lotH-120 {
		p.para(d.Blurb, body(12, 400), x, y+6, w, 17, cTan)
	}
	by := LY + lotH - 52
	back := int(float64(d.Cost) * ct.SellBack)
	g.button(p, x, by, w, 36, "Sell for "+money(back), bRed, !a.Running, func() { g.sellSelected() })
}

// drawSlotPicker lists the troupe so the player can book someone into a slot.
func (g *game) drawSlotPicker(p pen, x, y, w float64) {
	s := g.sim
	a := g.sel
	i := g.slotPick
	p.text(a.Def.Name+" · act "+itoa(i+1), body(11, 600), x, y+4, alLeft, cMint)
	p.text("Book an act", display(22), x, y+30, alLeft, cGold)
	g.closeX(p, x+w-30, y-6, func() { g.slotPick = -1 })
	y += 50
	hint := "Pick who performs in this spot."
	switch {
	case i == 0:
		hint = "The opener: Laughs warm the crowd up."
	case i == len(a.Lineup)-1:
		hint = "The finale: Thrills make a big finish."
	}
	p.text(hint, body(11, 500), x, y, alLeft, cTan)
	y += 14
	rh := 46.0
	n := 0
	for _, pf := range s.Troupe {
		pf := pf
		ry := y + float64(n)*(rh+4)
		if ry > LY+lotH-110 {
			break
		}
		n++
		ok := !pf.Act.Tall || a.Def.Tall
		here := pf.Tent == a && a.Lineup[i] == pf
		hov := ok && g.hover(x, ry, w, rh)
		bg := cNavy
		if hov {
			bg = mix(cNavy, cPanel2, 0.8)
		}
		if here {
			bg = mix(cNavy, cGold, 0.18)
		}
		p.rrect(x, ry, w, rh, 8, bg)
		q := p
		if !ok {
			q = p.fade(0.45)
		}
		g.portrait(q, pf, x+4, ry+3, 32, rh-6)
		q.text(fit(pf.Name, body(12, 700), w-110), body(12, 700), x+44, ry+13, alLeft, cCream)
		q.text(pf.Act.Name, body(10, 600), x+44, ry+29, alLeft, styleCol[pf.Act.Style])
		smallStars(q, x+w-56, ry+13, pf.Stars())
		lab, lc := tentLabel(pf.Tent), cMuted
		if !ok {
			lab, lc = "Big Top only", cBad
		} else if pf.Tired() {
			lab, lc = "Tired", cBad
		}
		q.text(lab, body(10, 500), x+w-8, ry+31, alRight, lc)
		if ok {
			g.onClick(x, ry, w, rh, func() {
				if pf.Tent != nil && pf.Tent.Running && pf.Tent != a {
					g.toast(pf.Name+" is on stage right now", "Wait for the show to finish.", "warn")
					snd.play("error")
					return
				}
				if err := s.Book(pf, a, i); err != nil {
					g.toast(err.Error(), "", "warn")
					snd.play("error")
					return
				}
				snd.play("place")
				g.slotPick = -1
			})
		}
	}
	if len(s.Troupe) == 0 {
		p.para("Your troupe is empty.", body(12, 400), x, y+10, w, 17, cMuted)
	}
	by := LY + lotH - 96
	if a.Lineup[i] != nil {
		g.button(p, x, by, w, 36, "Leave this spot empty", bDark, true, func() { s.ClearSlot(a, i); g.slotPick = -1 })
	}
	g.button(p, x, by+44, w, 36, "Hire more performers", bTeal, true, func() { g.panel = "troupe"; g.slotPick = -1 })
}

// fit shortens s with an ellipsis so it's no wider than w.
func fit(s string, f font, w float64) string {
	if textW(s, f) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 1 && textW(string(r)+"…", f) > w {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
