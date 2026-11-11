package main

import (
	"image/color"
	"math"

	ct "github.com/adamglassner/bigtopbuilder/internal/circus"
)

// All of the game's pictures are drawn with vector shapes: the tents,
// sideshows, carts and the performers in their costumes.

/* ---------- tents ---------- */

// tentLook is how a tent is painted.
type tentLook struct {
	stripe, base, trim color.RGBA
	peaks              int
}

func lookOf(id string) tentLook {
	if id == "bigtop" {
		return tentLook{stripe: cRed, base: cCream, trim: cGold, peaks: 2}
	}
	return tentLook{stripe: cBlue, base: cCream, trim: cGold, peaks: 1}
}

// tentRoof returns the apexes and the eave line of a tent.
func tentRoof(cx, by, w, wallH, roofH float64, peaks int) (eaveY float64, apex [][2]float64) {
	eaveY = by - wallH
	if peaks == 1 {
		return eaveY, [][2]float64{{cx, eaveY - roofH}}
	}
	return eaveY, [][2]float64{{cx - w*0.2, eaveY - roofH}, {cx + w*0.2, eaveY - roofH}}
}

// drawTent draws a striped circus tent standing on (cx, by).
func drawTent(p pen, d *ct.Def, cx, by, w float64, a *ct.Attraction, t float64) {
	lk := lookOf(d.ID)
	half := w * 0.46
	wallH := 34.0
	roofH := 64.0
	if lk.peaks == 2 {
		wallH, roofH = 46, 92
	}
	eave, apex := tentRoof(cx, by, w, wallH, roofH, lk.peaks)
	running := a != nil && a.Running
	sway := 0.0
	if d.ID == "bigtop" || running {
		sway = math.Sin(t*2.2) * 1.5
	}
	p.ellipse(cx, by+2, half+10, 12, cShadow)
	// walls: vertical stripes
	n := 10
	if lk.peaks == 2 {
		n = 14
	}
	sw := half * 2 / float64(n)
	for i := 0; i < n; i++ {
		c := lk.base
		if i%2 == 0 {
			c = lk.stripe
		}
		p.rect(cx-half+float64(i)*sw, eave, sw+0.5, wallH, c)
	}
	p.rect(cx-half, by-5, half*2, 5, mix(lk.stripe, cBlack, 0.35))
	// the doorway with its flaps tied back
	dw := 16.0
	if lk.peaks == 2 {
		dw = 22
	}
	doorTop := by - wallH + 8
	glow := cNight
	if running {
		glow = mix(cGold, cOrange, 0.5+0.5*math.Sin(t*5))
	}
	p.poly(glow, cx-dw, by-1, cx-dw, doorTop+dw*0.5, cx, doorTop, cx+dw, doorTop+dw*0.5, cx+dw, by-1)
	p.poly(lk.trim, cx-dw-4, by-1, cx-dw-4, doorTop+dw*0.4, cx-dw+5, by-1)
	p.poly(lk.trim, cx+dw+4, by-1, cx+dw+4, doorTop+dw*0.4, cx+dw-5, by-1)
	// roof: stripes fanning out from each peak
	if lk.peaks == 1 {
		ax, ay := apex[0][0]+sway, apex[0][1]
		m := 12
		for i := 0; i < m; i++ {
			x0 := cx - half - 6 + float64(i)*(half*2+12)/float64(m)
			x1 := x0 + (half*2+12)/float64(m)
			c := lk.base
			if i%2 == 0 {
				c = lk.stripe
			}
			p.poly(c, x0, eave, x1, eave, ax, ay)
		}
	} else {
		// two poles: a fan of stripes down each end, upright panels between them
		a0, a1 := apex[0], apex[1]
		l, r := cx-half-6, cx+half+6
		p.poly(lk.base, l, eave, a0[0]+sway, a0[1], a1[0]+sway, a1[1], r, eave)
		fan := func(ax, ay, x0, x1 float64, m int, startStripe bool) {
			for i := 0; i < m; i++ {
				if (i%2 == 0) != startStripe {
					continue
				}
				ea := x0 + float64(i)*(x1-x0)/float64(m)
				eb := ea + (x1-x0)/float64(m)
				p.poly(lk.stripe, ea, eave, eb, eave, ax, ay)
			}
		}
		fan(a0[0]+sway, a0[1], l, a0[0], 6, true)
		fan(a1[0]+sway, a1[1], a1[0], r, 6, false)
		m := 6
		pw := (a1[0] - a0[0]) / float64(m)
		for i := 1; i < m; i += 2 {
			x0 := a0[0] + float64(i)*pw
			p.poly(lk.stripe, x0+sway, a0[1], x0+pw+sway, a0[1], x0+pw, eave, x0, eave)
		}
		p.line(a0[0]+sway, a0[1], a1[0]+sway, a1[1], 2.5, lk.trim)
	}
	// scalloped valance along the eave
	k := 12
	if lk.peaks == 2 {
		k = 16
	}
	vw := (half*2 + 12) / float64(k)
	p.rect(cx-half-6, eave-3, half*2+12, 6, lk.trim)
	for i := 0; i < k; i++ {
		x := cx - half - 6 + (float64(i)+0.5)*vw
		c := lk.trim
		if i%2 == 1 {
			c = lk.stripe
		}
		p.circle(x, eave+3, vw/2, c)
	}
	// poles and pennants
	for i, ap := range apex {
		ax, ay := ap[0]+sway, ap[1]
		p.line(ax, ay, ax, ay-22, 2.5, cTan)
		fl := math.Sin(t*6+float64(i)) * 4
		fc := cGold
		if i == 1 || lk.peaks == 1 {
			fc = cRed
		}
		if lk.peaks == 1 {
			fc = cGold
		}
		p.poly(fc, ax, ay-22, ax+18, ay-17+fl, ax, ay-12)
		p.circle(ax, ay-1, 4, lk.trim)
	}
	// marquee sign over the door
	name := "RING TENT"
	if lk.peaks == 2 {
		name = "BIG TOP"
	}
	sy := eave - 2
	tw := textW(name, head(10)) + 14
	p.rrect(cx-tw/2, sy-9, tw, 18, 4, cRed)
	p.rrectStroke(cx-tw/2, sy-9, tw, 18, 4, 1.5, cGold)
	p.text(name, head(10), cx, sy, alCenter, cCream)
	if a != nil && d.Slots > 0 && a.Booked() == 0 {
		p.rrect(cx-34, by-wallH/2-8, 68, 16, 4, withA(cNight, 0.85))
		p.text("NO SHOW", head(9), cx, by-wallH/2, alCenter, cMuted)
	}
}

// tentTop is the highest point of a tent, for spotlights and labels.
func tentTop(d *ct.Def, cx, by, w float64) (float64, float64) {
	lk := lookOf(d.ID)
	wallH, roofH := 34.0, 64.0
	if lk.peaks == 2 {
		wallH, roofH = 46, 92
	}
	_, apex := tentRoof(cx, by, w, wallH, roofH, lk.peaks)
	return cx, apex[0][1]
}

// drawSpotlights sweeps two beams over a tent while its show is on.
func drawSpotlights(p pen, d *ct.Def, cx, by, w, t, dark float64) {
	_, top := tentTop(d, cx, by, w)
	for i, side := range []float64{-1, 1} {
		ang := side*0.35 + math.Sin(t*0.9+float64(i)*2)*0.3
		bx := cx + side*w*0.3
		l := 220.0
		ex, ey := bx+math.Sin(ang)*l, top-math.Cos(ang)*l
		perp := 26.0
		px, py := math.Cos(ang)*perp, math.Sin(ang)*perp
		q := p.fade(0.05 + 0.07*dark)
		q.poly(cCream, bx-2, top+10, bx+2, top+10, ex+px, ey+py, ex-px, ey-py)
	}
}

/* ---------- sideshows ---------- */

func drawFunhouse(p pen, x, y, w, h, t float64, a *ct.Attraction) {
	cx, by := x+w/2, y+h-4
	p.ellipse(cx, by, w*0.5, 8, cShadow)
	// wobbly facade
	cols := []color.RGBA{cGold, cPink, cTeal, cPurple}
	p.poly(cPurple, x+4, by, x+4, by-70, x+18, by-80, x+34, by-72, x+52, by-86, x+70, by-74, x+w-4, by-82, x+w-4, by)
	for i := 0; i < 5; i++ {
		c := cols[i%len(cols)]
		yy := by - 14 - float64(i)*12
		p.rect(x+6, yy, w-12, 4, withA(c, 0.6))
	}
	// clown face over the door
	fx, fy := cx, by-44
	p.circle(fx, fy, 22, cCream)
	p.circle(fx-9, fy-6, 4, cNavy)
	p.circle(fx+9, fy-6, 4, cNavy)
	p.circle(fx, fy+1, 5, cRed)
	open := 0.0
	if a != nil && a.Running {
		open = 0.5 + 0.5*math.Sin(t*4)
	}
	p.ellipse(fx, fy+13+open*2, 13, 5+open*4, cRedDk)
	p.circle(fx-22, fy-14, 8, cOrange)
	p.circle(fx+22, fy-14, 8, cOrange)
	// door below
	p.rrect(cx-10, by-22, 20, 22, 4, cNight)
	p.rrect(cx-30, by-104, 60, 18, 4, cRed)
	p.rrectStroke(cx-30, by-104, 60, 18, 4, 1.5, cGold)
	p.text("FUNHOUSE", head(10), cx, by-95, alCenter, cCream)
}

func drawMirrors(p pen, x, y, w, h, t float64) {
	cx, by := x+w/2, y+h-4
	p.ellipse(cx, by, w*0.5, 8, cShadow)
	p.rect(x+4, by-74, w-8, 74, hex(0x3a2a6a))
	p.poly(hex(0x5a3a9a), x, by-74, cx, by-96, x+w, by-74)
	for i := 0; i < 3; i++ {
		mx := x + 12 + float64(i)*(w-24)/3
		mw := (w-24)/3 - 6
		wob := math.Sin(t*2+float64(i)) * 2
		p.rrect(mx, by-64, mw, 46, 6, hex(0x9ad0e8))
		p.rrect(mx+3, by-60, mw-6, 38, 5, hex(0xc9eaf6))
		// a funny reflection
		p.ellipse(mx+mw/2, by-40, 5+wob+float64(i)*2, 12-float64(i)*3, hex(0x7ab0c8))
		p.circle(mx+mw/2, by-54+float64(i)*2, 4, hex(0x7ab0c8))
	}
	p.rect(x+4, by-12, w-8, 12, hex(0x2a1d4f))
	p.rrect(cx-30, by-112, 60, 16, 4, cPurple)
	p.text("MIRRORS", head(10), cx, by-104, alCenter, cCream)
}

func drawFortune(p pen, x, y, w, h, t float64) {
	cx, by := x+w/2, y+h-4
	p.ellipse(cx, by, w*0.46, 7, cShadow)
	p.rrect(x+10, by-62, w-20, 60, 6, cPlum)
	p.poly(cPurple, x+4, by-60, cx, by-92, x+w-4, by-60)
	for i := 0; i < 5; i++ {
		sx := x + 16 + float64(i)*(w-32)/4
		star(p, sx, by-70+float64(i%2)*8, 3.5, cGold)
	}
	p.rect(x+18, by-50, w-36, 26, cPlumDk)
	pulse := 0.6 + 0.4*math.Sin(t*2.5)
	p.glow(cx, by-36, 22, cMint, 0.5*pulse)
	p.circle(cx, by-36, 9, mix(cMint, cWhite, 0.4))
	p.rect(cx-8, by-28, 16, 4, cGold)
	p.rect(x+12, by-24, w-24, 6, cGold)
	p.rrect(cx-26, by-110, 52, 16, 4, cPlumDk)
	p.text("FORTUNES", head(9), cx, by-102, alCenter, cGold)
}

/* ---------- food carts ---------- */

func foodAccent(id string) color.RGBA {
	switch id {
	case "popcorn":
		return cRed
	case "cottoncandy":
		return cPink
	case "hotdogs":
		return cOrange
	case "soda":
		return cTeal
	}
	return cGold
}

func drawCart(p pen, d *ct.Def, x, y, w, h, t float64) {
	acc := foodAccent(d.ID)
	cx, by := x+w/2, y+h-4
	p.ellipse(cx, by, w*0.46, 7, cShadow)
	// the cart body
	p.rrect(x+8, by-40, w-16, 28, 5, acc)
	p.rect(x+8, by-34, w-16, 4, cCream)
	p.rect(x+8, by-20, w-16, 3, mix(acc, cBlack, 0.3))
	// wheels
	for _, wx := range []float64{x + 20, x + w - 20} {
		p.circle(wx, by-9, 8, hex(0x2a2440))
		p.ring(wx, by-9, 8, 2, cGold)
		p.circle(wx, by-9, 2.5, cGold)
	}
	// striped canopy on poles
	p.line(x+12, by-40, x+12, by-74, 2, cTan)
	p.line(x+w-12, by-40, x+w-12, by-74, 2, cTan)
	n := 6
	sw := (w - 12) / float64(n)
	for i := 0; i < n; i++ {
		c := acc
		if i%2 == 1 {
			c = cCream
		}
		p.rect(x+6+float64(i)*sw, by-84, sw+0.5, 10, c)
		p.circle(x+6+(float64(i)+0.5)*sw, by-74, sw/2, c)
	}
	// what's for sale, on the counter
	top := by - 40
	switch d.ID {
	case "popcorn":
		p.rect(cx-16, top-26, 32, 26, withA(hex(0xcfe8f0), 0.6))
		p.strokeRect(cx-16, top-26, 32, 26, 1.5, cGold)
		for i := 0; i < 14; i++ {
			px := cx - 12 + math.Mod(float64(i*7), 24)
			py := top - 4 - math.Mod(float64(i*5), 14)
			p.circle(px, py, 3, hex(0xfff3c4))
		}
	case "cottoncandy":
		p.ellipse(cx, top-4, 16, 5, hex(0xd8d8e0))
		spin := t * 4
		for i := 0; i < 3; i++ {
			sx := cx - 14 + float64(i)*14
			p.line(sx, top-2, sx, top-18, 2, cCream)
			p.circle(sx, top-24+math.Sin(spin+float64(i))*0.8, 8, []color.RGBA{cPink, hex(0x9ad0ff), cPink}[i])
		}
	case "hotdogs":
		p.rrect(cx-18, top-12, 36, 10, 5, hex(0xe8b46a))
		p.rrect(cx-20, top-15, 40, 6, 3, hex(0xb0502a))
		p.line(cx-14, top-13, cx+14, top-13, 1.5, cGold)
	case "soda":
		for i := 0; i < 4; i++ {
			bx := cx - 15 + float64(i)*10
			c := []color.RGBA{cOrange, hex(0x6a3a1a), cRed, cTeal}[i]
			p.rrect(bx-3.5, top-18, 7, 18, 2.5, c)
			p.rect(bx-1.5, top-23, 3, 6, c)
		}
	}
	name := map[string]string{"popcorn": "POPCORN", "cottoncandy": "COTTON CANDY", "hotdogs": "HOT DOGS", "soda": "SODA POP"}[d.ID]
	f := head(9)
	tw := textW(name, f) + 12
	p.rrect(cx-tw/2, by-100, tw, 15, 4, cNavy)
	p.rrectStroke(cx-tw/2, by-100, tw, 15, 4, 1.5, acc)
	p.text(name, f, cx, by-92.5, alCenter, cCream)
}

/* ---------- performers ---------- */

var costumeOf = map[string][2]color.RGBA{
	"clown":    {cGold, cRed},
	"stilts":   {cTeal, cCream},
	"juggler":  {cPurple, cGold},
	"strong":   {cRed, cGold},
	"acrobats": {cBlue, cCream},
	"fire":     {hex(0x9a2f2a), cOrange},
	"trapeze":  {cPink, cCream},
}

// drawPerformer draws a performer standing with their feet at (x, y), k times life size.
func drawPerformer(p pen, act string, x, y, k, t float64, look, skinN int) {
	skin := skins[skinN%len(skins)]
	cc := costumeOf[act]
	body, acc := cc[0], cc[1]
	hair := []color.RGBA{hex(0x3a2a1a), hex(0x6a4a2a), hex(0xd9b56a), hex(0x1a1a1a), hex(0xa0522d)}[(look+skinN)%5]
	p.ellipse(x, y, 13*k, 3.5*k, cShadow)
	figure := func(x, y, k float64, arms [4]float64) {
		// legs
		p.line(x-4*k, y-22*k, x-5*k, y, 4*k, mix(body, cBlack, 0.3))
		p.line(x+4*k, y-22*k, x+5*k, y, 4*k, mix(body, cBlack, 0.3))
		// body
		p.rrect(x-8*k, y-44*k, 16*k, 24*k, 5*k, body)
		p.rect(x-8*k, y-26*k, 16*k, 4*k, acc)
		// arms (angles from straight down, in radians)
		for i, side := range []float64{-1, 1} {
			sx, sy := x+side*8*k, y-40*k
			a := arms[i*2]
			ex, ey := sx+side*math.Sin(a)*16*k, sy+math.Cos(a)*16*k
			p.line(sx, sy, ex, ey, 3.5*k, body)
			p.circle(ex, ey, 2.6*k, skin)
		}
		// head
		p.circle(x, y-51*k, 7.5*k, skin)
		p.ellipse(x, y-56*k, 7.5*k, 3.6*k, hair)
	}
	switch act {
	case "clown":
		// baggy suit, big shoes, wig and nose
		p.ellipse(x-7*k, y-1*k, 7*k, 3*k, cRed)
		p.ellipse(x+7*k, y-1*k, 7*k, 3*k, cRed)
		figure(x, y, k, [4]float64{0.6, 0, 0.6, 0})
		for i := 0; i < 5; i++ {
			p.circle(x-5*k+float64(i%3)*5*k, y-40*k+float64(i/3)*9*k, 1.8*k, []color.RGBA{cRed, cBlue, cTeal}[i%3])
		}
		p.circle(x, y-51*k, 7.5*k, cCream)
		for _, d := range []float64{-1, 1} {
			p.circle(x+d*8*k, y-56*k, 4.5*k, cOrange)
		}
		p.circle(x, y-50*k, 2.6*k, cRed)
		p.ellipse(x, y-46.5*k, 3.5*k, 1.4*k, cRed)
		p.circle(x-2.8*k, y-53.5*k, 1*k, cNavy)
		p.circle(x+2.8*k, y-53.5*k, 1*k, cNavy)
	case "stilts":
		// long striped trouser legs on stilts and a top hat
		for _, d := range []float64{-1, 1} {
			p.line(x+d*4*k, y-48*k, x+d*6*k, y, 5*k, acc)
			for i := 0; i < 5; i++ {
				yy := y - 44*k + float64(i)*9*k
				p.line(x+d*(4+float64(i)*0.4)*k-2.5*k, yy, x+d*(4+float64(i)*0.4)*k+2.5*k, yy, 2.5*k, body)
			}
		}
		figure(x, y-44*k, k, [4]float64{2.4, 0, 2.6, 0})
		p.rect(x-6*k, y-114*k, 12*k, 12*k, cNavy)
		p.rect(x-9*k, y-103*k, 18*k, 3*k, cNavy)
		p.rect(x-6*k, y-106*k, 12*k, 2.5*k, cRed)
	case "juggler":
		figure(x, y, k, [4]float64{1.9, 0, 1.9, 0})
		// diamond pattern
		for i := 0; i < 3; i++ {
			dx := x - 4*k + float64(i)*4*k
			p.poly(acc, dx, y-40*k, dx+2*k, y-36*k, dx, y-32*k, dx-2*k, y-36*k)
		}
		for i := 0; i < 3; i++ {
			ph := t*3.2 + float64(i)*2*math.Pi/3
			bx := x + math.Cos(ph)*12*k
			by := y - 66*k - math.Abs(math.Sin(ph))*18*k
			p.circle(bx, by, 3.4*k, []color.RGBA{cRed, cGold, cTeal}[i])
		}
	case "strong":
		figure(x, y, k, [4]float64{2.9, 0, 2.9, 0})
		p.rect(x-3*k, y-44*k, 6*k, 16*k, acc)
		bar := y - 72*k
		p.line(x-24*k, bar, x+24*k, bar, 3*k, hex(0x6a6e78))
		for _, d := range []float64{-1, 1} {
			p.circle(x+d*24*k, bar, 9*k, hex(0x2a2c31))
			p.ring(x+d*24*k, bar, 9*k, 1.5*k, hex(0x6a6e78))
		}
	case "acrobats":
		// one standing on the other's shoulders
		figure(x, y, k*0.95, [4]float64{2.2, 0, 2.2, 0})
		top := y - 58*k
		q := p
		_ = q
		figure(x, top, k*0.8, [4]float64{2.6, 0, 2.6, 0})
		p.circle(x, top-62*k, 2*k, cGold)
		for i := 0; i < 4; i++ {
			sp := math.Sin(t*5 + float64(i))
			if sp > 0.3 {
				star(p, x-18*k+float64(i)*12*k, y-90*k+float64(i%2)*30*k, 2.5*k, cGold)
			}
		}
	case "fire":
		figure(x, y, k, [4]float64{2.2, 0, 0.9, 0})
		// torch in the left hand, a fireball from the mouth
		hx, hy := x-8*k-math.Sin(2.2)*16*k, y-40*k+math.Cos(2.2)*16*k
		p.line(hx, hy, hx, hy-12*k, 3*k, cBrown)
		flick := 1 + 0.2*math.Sin(t*14)
		p.circle(hx, hy-15*k, 4*flick*k, cOrange)
		p.circle(hx, hy-16*k, 2.2*flick*k, cGold)
		burst := 0.6 + 0.4*math.Sin(t*3)
		for i := 0; i < 5; i++ {
			fx := x + 10*k + float64(i)*6*k*burst
			fy := y - 50*k - float64(i)*2*k
			p.circle(fx, fy, (3+float64(i))*k*burst, []color.RGBA{cGold, cOrange, cRed}[i%3])
		}
	case "trapeze":
		// hanging from the bar
		bar := y - 82*k
		swing := math.Sin(t*1.6) * 6 * k
		p.line(x-12*k+swing, bar, x-12*k, bar-40*k, 1.5*k, cTan)
		p.line(x+12*k+swing, bar, x+12*k, bar-40*k, 1.5*k, cTan)
		p.line(x-14*k+swing, bar, x+14*k+swing, bar, 3*k, cGold)
		figure(x+swing, y-6*k, k, [4]float64{3.0, 0, 3.0, 0})
		for i := 0; i < 3; i++ {
			p.circle(x+swing-4*k+float64(i)*4*k, y-38*k, 1.4*k, cWhite)
		}
	default:
		figure(x, y, k, [4]float64{0.4, 0, 0.4, 0})
	}
}
