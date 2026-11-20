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
	wallH, roofH       float64
	sign               string
}

func lookOf(id string) tentLook {
	switch id {
	case "bigtop":
		return tentLook{stripe: cRed, base: cCream, trim: cGold, peaks: 2, wallH: 46, roofH: 92, sign: "BIG TOP"}
	case "threering":
		return tentLook{stripe: hex(0x7a3fb4), base: cCream, trim: cGold, peaks: 3, wallH: 50, roofH: 104, sign: "THREE RINGS"}
	}
	return tentLook{stripe: cBlue, base: cCream, trim: cGold, peaks: 1, wallH: 34, roofH: 64, sign: "RING TENT"}
}

// tentRoof returns the apexes and the eave line of a tent.
func tentRoof(cx, by, w, wallH, roofH float64, peaks int) (eaveY float64, apex [][2]float64) {
	eaveY = by - wallH
	switch peaks {
	case 1:
		return eaveY, [][2]float64{{cx, eaveY - roofH}}
	case 3:
		return eaveY, [][2]float64{{cx - w*0.27, eaveY - roofH}, {cx, eaveY - roofH}, {cx + w*0.27, eaveY - roofH}}
	}
	return eaveY, [][2]float64{{cx - w*0.2, eaveY - roofH}, {cx + w*0.2, eaveY - roofH}}
}

// drawTent draws a striped circus tent standing on (cx, by).
func drawTent(p pen, d *ct.Def, cx, by, w float64, a *ct.Attraction, t float64) {
	lk := lookOf(d.ID)
	half := w * 0.46
	wallH, roofH := lk.wallH, lk.roofH
	eave, apex := tentRoof(cx, by, w, wallH, roofH, lk.peaks)
	running := a != nil && a.Running
	sway := 0.0
	if lk.peaks > 1 || running {
		sway = math.Sin(t*2.2) * 1.5
	}
	p.ellipse(cx, by+2, half+10, 12, cShadow)
	// walls: vertical stripes
	n := []int{0, 10, 14, 18}[lk.peaks]
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
	dw := []float64{0, 16, 22, 26}[lk.peaks]
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
		a0, a1 := apex[0], apex[len(apex)-1]
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
	k := []int{0, 12, 16, 20}[lk.peaks]
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
		fc := []color.RGBA{cGold, cRed, cTeal}[i%3]
		p.poly(fc, ax, ay-22, ax+18, ay-17+fl, ax, ay-12)
		p.circle(ax, ay-1, 4, lk.trim)
	}
	// marquee sign over the door
	name := lk.sign
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
	_, apex := tentRoof(cx, by, w, lk.wallH, lk.roofH, lk.peaks)
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
	"clown":     {cGold, cRed},
	"stilts":    {cTeal, cCream},
	"juggler":   {cPurple, cGold},
	"strong":    {cRed, cGold},
	"acrobats":  {cBlue, cCream},
	"daredevil": {cRed, cGold},
	"fire":      {hex(0x9a2f2a), cOrange},
	"trapeze":   {cPink, cCream},
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

/* ---------- the human cannonball ---------- */

// cannonPhase splits a cannonball show into countdown, flight and landing.
func cannonPhase(a *ct.Attraction) (prog float64, firing bool) {
	if a == nil || !a.Running || a.Length <= 0 {
		return 0, false
	}
	return a.Anim / a.Length, true
}

func drawCannon(p pen, x, y, w, h, t float64, a *ct.Attraction) {
	by := y + h - 4
	p.ellipse(x+w/2, by, w*0.5, 9, cShadow)
	// a low stage
	p.rrect(x+4, by-12, w-8, 12, 4, hex(0x6a3a2a))
	p.rect(x+4, by-12, w-8, 3, cGold)
	// the safety net on the right
	nx0, nx1, ny := x+w-58, x+w-8, by-34.0
	p.line(nx0, by-12, nx0, ny-6, 3, hex(0x4a4a5a))
	p.line(nx1, by-12, nx1, ny-6, 3, hex(0x4a4a5a))
	prog, running := cannonPhase(a)
	sag := 0.0
	if running && prog > 0.72 {
		sag = math.Max(0, math.Sin((prog-0.72)/0.28*math.Pi*3)) * 6 * (1 - prog)
	}
	p.poly(withA(cCream, 0.55), nx0, ny-4, nx1, ny-4, nx1-4, ny+6+sag, nx0+4, ny+6+sag)
	for i := 0; i <= 5; i++ {
		xx := nx0 + float64(i)*(nx1-nx0)/5
		p.line(xx, ny-4, xx+(float64(i)-2.5)*0.6, ny+6+sag, 1, withA(cNavy, 0.5))
	}
	// the cannon on its carriage, aimed up and to the right
	cx, cy := x+40, by-22
	ang := -0.62
	cs, sn := math.Cos(ang), math.Sin(ang)
	pt := func(u, v float64) (float64, float64) { return cx + u*cs - v*sn, cy + u*sn + v*cs }
	body := func(u0, u1, r0, r1 float64, c color.RGBA) {
		ax, ay := pt(u0, -r0)
		bx, by2 := pt(u1, -r1)
		ccx, ccy := pt(u1, r1)
		dx, dy := pt(u0, r0)
		p.poly(c, ax, ay, bx, by2, ccx, ccy, dx, dy)
	}
	body(-14, 54, 14, 10, hex(0x2a2c48))
	body(-14, 54, 14, 10, hex(0x2a2c48))
	for _, u := range []float64{2, 30, 50} {
		body(u, u+4, 14, 12, cGold)
	}
	mx, my := pt(56, 0)
	p.circle(mx, my, 10, hex(0x1a1b30))
	p.ring(mx, my, 10, 2.5, cGold)
	sx, sy := pt(16, 0)
	star(p, sx, sy, 5, cRed)
	// carriage and wheels
	p.poly(cRedDk, cx-22, cy+8, cx+20, cy+8, cx+12, cy-6, cx-14, cy-4)
	for _, wx := range []float64{cx - 12, cx + 12} {
		p.circle(wx, by-14, 10, hex(0x6a3a2a))
		p.ring(wx, by-14, 10, 2, cGold)
		for k := 0; k < 4; k++ {
			ang := float64(k)*math.Pi/4 + t*0.2
			p.line(wx-math.Cos(ang)*9, by-14-math.Sin(ang)*9, wx+math.Cos(ang)*9, by-14+math.Sin(ang)*9, 1.4, cGold)
		}
	}
	// the show: fuse, BOOM, flight, net
	landX, landY := (nx0+nx1)/2, ny+2+sag
	switch {
	case !running:
		// the daredevil waits by the cannon, waving
		drawPerformer(p, "daredevil", x+w/2-8, by-12, 0.45, t, 3, 2)
	case prog < 0.3:
		fx, fy := pt(-16, 0)
		if math.Sin(t*30) > 0 {
			p.circle(fx, fy, 3.5, cGold)
			p.circle(fx, fy, 1.8, cWhite)
		}
		p.text(itoa(3-int(prog/0.1)), display(22), x+w/2+6, by-74, alCenter, cGold)
	case prog < 0.72:
		f := (prog - 0.3) / 0.42
		if f < 0.25 {
			sm := 1 - f/0.25
			for k := 0; k < 5; k++ {
				p.circle(mx+float64(k)*5, my-float64(k)*3, (6+float64(k)*3)*(1-sm*0.3), withA(cCream, 0.7*sm))
			}
		}
		px := mx + (landX-mx)*f
		py := my + (landY-my)*f - math.Sin(f*math.Pi)*110
		spin := f * 9
		p.circle(px, py, 6, cRed)
		p.line(px-math.Cos(spin)*7, py-math.Sin(spin)*7, px+math.Cos(spin)*7, py+math.Sin(spin)*7, 3, cGold)
		p.circle(px+math.Cos(spin+1.5)*6, py+math.Sin(spin+1.5)*6, 3.2, hex(0xf2c79e))
		// a dotted trail
		for k := 1; k < 6; k++ {
			ff := f - float64(k)*0.035
			if ff < 0 {
				break
			}
			tx := mx + (landX-mx)*ff
			ty := my + (landY-my)*ff - math.Sin(ff*math.Pi)*110
			p.circle(tx, ty, 2, withA(cCream, 0.5-float64(k)*0.08))
		}
	default:
		p.circle(landX, landY-6, 6, cRed)
		p.circle(landX+5, landY-10, 3.2, hex(0xf2c79e))
		if math.Sin(t*6) > 0 {
			star(p, landX-12, landY-22, 3, cGold)
			star(p, landX+14, landY-26, 3, cGold)
		}
	}
	f := head(9)
	tw := textW("HUMAN CANNONBALL", f) + 14
	p.rrect(x+w/2-tw/2, by-100, tw, 15, 4, cRed)
	p.rrectStroke(x+w/2-tw/2, by-100, tw, 15, 4, 1.5, cGold)
	p.text("HUMAN CANNONBALL", f, x+w/2, by-92.5, alCenter, cCream)
}

/* ---------- decorations ---------- */

func drawFlowers(p pen, cx, by, t float64, seed int) {
	p.ellipse(cx, by-6, 19, 9, hex(0x5a3a22))
	p.ellipse(cx, by-8, 17, 7, hex(0x6e4a2c))
	cols := []color.RGBA{cRed, cGold, cPink, cPurple, cOrange, cCream}
	for k := 0; k < 9; k++ {
		fx := cx - 13 + math.Mod(float64(k*11+seed*7), 26)
		fy := by - 8 - math.Mod(float64(k*7+seed*3), 10) - 6
		sway := math.Sin(t*1.5+float64(k)) * 0.8
		p.line(fx, by-8, fx+sway, fy, 1.4, hex(0x3d8a3d))
		p.ellipse(fx+sway-3, fy+3, 3, 1.5, hex(0x4fa04a))
		c := cols[(k+seed)%len(cols)]
		for q := 0; q < 5; q++ {
			a := float64(q) * 2 * math.Pi / 5
			p.circle(fx+sway+math.Cos(a)*2.4, fy+math.Sin(a)*2.4, 2, c)
		}
		p.circle(fx+sway, fy, 1.4, cGold)
	}
}

func drawArch(p pen, x, by, w, t float64) {
	cols := []color.RGBA{cRed, cGold, cTeal, cPink, cPurple, cOrange}
	l, r := x+10, x+w-10
	top := by - 86.0
	mid := (l + r) / 2
	rad := (r - l) / 2
	p.ellipse(l, by-2, 8, 3, cShadow)
	p.ellipse(r, by-2, 8, 3, cShadow)
	// the columns
	k := 0
	for _, cx := range []float64{l, r} {
		for yy := by - 6; yy > top+rad; yy -= 10 {
			p.circle(cx-3, yy, 6, cols[k%len(cols)])
			p.circle(cx+3, yy-5, 6, cols[(k+2)%len(cols)])
			k++
		}
	}
	// the arch
	for i := 0; i <= 12; i++ {
		a := math.Pi - float64(i)*math.Pi/12
		bx := mid + math.Cos(a)*rad
		bby := top + rad - math.Sin(a)*rad
		bob := math.Sin(t*2+float64(i)) * 1.2
		p.circle(bx, bby+bob, 7, cols[i%len(cols)])
		p.circle(bx-2, bby-2+bob, 2, withA(cWhite, 0.55))
	}
}

func drawStatue(p pen, cx, by, t float64) {
	p.ellipse(cx, by-2, 18, 5, cShadow)
	// stone plinth
	p.rect(cx-15, by-22, 30, 20, hex(0x8a8aa0))
	p.rect(cx-17, by-26, 34, 6, hex(0xa6a6bc))
	p.rect(cx-17, by-6, 34, 4, hex(0x6e6e84))
	gold, dk, lt := hex(0xe2b53e), hex(0xa8801f), hex(0xffe08a)
	b := by - 26
	// legs, coat, arm raised with a cane, top hat
	p.line(cx-3, b, cx-3, b-14, 3.5, dk)
	p.line(cx+3, b, cx+3, b-14, 3.5, dk)
	p.poly(gold, cx-8, b-10, cx+8, b-10, cx+6, b-30, cx-6, b-30)
	p.poly(dk, cx-8, b-10, cx-3, b-10, cx-1, b-28, cx-6, b-30)
	p.line(cx+6, b-28, cx+15, b-40, 3, gold)
	p.line(cx+15, b-40, cx+18, b-56, 1.6, dk)
	p.circle(cx+18, b-57, 2, lt)
	p.line(cx-6, b-28, cx-11, b-17, 3, dk)
	p.circle(cx, b-35, 5, gold)
	p.rect(cx-5, b-48, 10, 10, dk)
	p.rect(cx-8, b-39, 16, 2.5, dk)
	// a glint that wanders over the gold
	gl := 0.5 + 0.5*math.Sin(t*1.3)
	p.circle(cx-3+gl*6, b-26+gl*4, 1.6, withA(cWhite, 0.8))
}

/* ---------- grounds crew ---------- */

func (g *game) drawSweeper(p pen, c *ct.Sweeper) {
	x, y := tilePx(c.X, c.Y)
	ph := math.Sin(g.clock*10+float64(c.ID)) * 2.6
	p.ellipse(x, y, 5.5, 2.2, cShadow)
	overall := hex(0x2f7d5a)
	p.line(x-2, y-8, x-2+ph*0.5, y, 2.4, overall)
	p.line(x+2, y-8, x+2-ph*0.5, y, 2.4, overall)
	p.rrect(x-5, y-18, 10, 11, 3, overall)
	p.rect(x-5, y-17, 10, 3, cCream)
	skin := skins[c.Skin%len(skins)]
	p.circle(x, y-22, 4.4, skin)
	p.ellipse(x, y-25, 5, 2.2, hex(0x2f7d5a))
	p.rect(x, y-26, 6, 1.8, hex(0x235e43))
	// the broom
	sw := 0.0
	if c.Sweeping {
		sw = math.Sin(g.clock*14) * 4
		p.circle(x+9+sw, y-1, 2.5, withA(cTan, 0.6))
		p.circle(x+13-sw, y-3, 1.8, withA(cTan, 0.5))
	}
	p.line(x+3, y-16, x+8+sw, y, 1.6, cBrown)
	p.poly(hex(0xd9b56a), x+5+sw, y, x+12+sw, y, x+10+sw, y-4, x+6+sw, y-4)
}
