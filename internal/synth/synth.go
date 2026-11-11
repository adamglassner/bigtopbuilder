// Package synth makes Big Top Builder Classic's music and sound effects from
// scratch: a circus march and drumrolls, applause, bells and honks.
package synth

import (
	"math"
	"math/rand"
)

const sampleRate = 44100

/* ---------- synthesis ---------- */

type buf []float64 // mono samples

func newBuf(sec float64) buf { return make(buf, int(sec*sampleRate)) }

func freq(midi float64) float64 { return 440 * math.Pow(2, (midi-69)/12) }

// add mixes a voice into b starting at t seconds.
func (b buf) add(t, dur float64, f func(tt, p float64) float64, amp float64) {
	start := int(t * sampleRate)
	n := int(dur * sampleRate)
	for i := 0; i < n && start+i < len(b); i++ {
		if start+i < 0 {
			continue
		}
		tt := float64(i) / sampleRate
		b[start+i] += f(tt, tt/dur) * amp
	}
}

func env(tt, dur, a, r float64) float64 {
	e := 1.0
	if tt < a {
		e = tt / a
	}
	if tt > dur-r {
		e *= math.Max(0, (dur-tt)/r)
	}
	return e
}

// calliope: a breathy steam-organ pipe with vibrato.
func calliope(midi, dur float64) func(tt, p float64) float64 {
	f := freq(midi)
	r := rand.New(rand.NewSource(int64(midi * 100)))
	return func(tt, p float64) float64 {
		vib := 1 + 0.006*math.Sin(2*math.Pi*5.6*tt)
		ph := 2 * math.Pi * f * vib * tt
		v := math.Sin(ph) + 0.45*math.Sin(2*ph) + 0.22*math.Sin(3*ph) + 0.1*math.Sin(4*ph)
		v += (r.Float64() - .5) * 0.06
		return v * env(tt, dur, 0.012, 0.07)
	}
}

func tuba(midi, dur float64) func(tt, p float64) float64 {
	f := freq(midi)
	return func(tt, p float64) float64 {
		ph := 2 * math.Pi * f * tt
		v := math.Sin(ph) + 0.35*math.Sin(2*ph) + 0.18*math.Sin(3*ph)
		return v * env(tt, dur, 0.01, 0.08) * math.Exp(-tt*2.2)
	}
}

func bell(midi, dur float64) func(tt, p float64) float64 {
	f := freq(midi)
	return func(tt, p float64) float64 {
		ph := 2 * math.Pi * f * tt
		v := math.Sin(ph) + 0.5*math.Sin(2.76*ph) + 0.25*math.Sin(5.4*ph)
		return v * math.Exp(-tt*5) * env(tt, dur, 0.002, 0.02)
	}
}

func square(midi, dur float64) func(tt, p float64) float64 {
	f := freq(midi)
	return func(tt, p float64) float64 {
		v := 0.0
		for k := 1.0; k <= 7; k += 2 {
			v += math.Sin(2*math.Pi*f*k*tt) / k
		}
		return v * env(tt, dur, 0.004, 0.03)
	}
}

func sweep(f0, f1, dur float64) func(tt, p float64) float64 {
	ph := 0.0
	last := 0.0
	return func(tt, p float64) float64 {
		f := f0 + (f1-f0)*p
		ph += 2 * math.Pi * f * (tt - last)
		last = tt
		return (math.Sin(ph) + 0.3*math.Sin(3*ph)) * env(tt, dur, 0.005, 0.05)
	}
}

func noise(dur, decay float64) func(tt, p float64) float64 {
	r := rand.New(rand.NewSource(7))
	return func(tt, p float64) float64 {
		return (r.Float64()*2 - 1) * math.Exp(-tt*decay) * env(tt, dur, 0.001, 0.02)
	}
}

// pcm turns mono samples into 16-bit stereo bytes, with a soft limiter.
func (b buf) pcm(gain float64) []byte {
	out := make([]byte, len(b)*4)
	for i, v := range b {
		x := math.Tanh(v * gain)
		s := int16(x * 30000)
		out[i*4], out[i*4+1] = byte(s), byte(s>>8)
		out[i*4+2], out[i*4+3] = byte(s), byte(s>>8)
	}
	return out
}

// March writes "Brookhaven Bandwagon": 32 bars of an oom-pah circus march in C,
// as 16-bit stereo PCM. Written for Big Top Builder Classic by Ironite.
func March() []byte {
	const bpm = 128.0
	beat := 60 / bpm
	type note struct{ m, beats float64 }
	n := func(m, b float64) note { return note{m, b} }
	R := -1.0 // rest
	melody := [][]note{
		{n(67, .5), n(72, .5), n(76, .5), n(79, .5)}, {n(76, 1), n(72, 1)}, {n(74, .5), n(77, .5), n(81, .5), n(77, .5)}, {n(79, 1.5), n(R, .5)},
		{n(76, .5), n(77, .5), n(79, .5), n(76, .5)}, {n(72, .5), n(74, .5), n(76, 1)}, {n(74, .5), n(71, .5), n(67, .5), n(71, .5)}, {n(72, 1.5), n(R, .5)},
		{n(81, .5), n(79, .5), n(77, .5), n(76, .5)}, {n(77, 1), n(74, 1)}, {n(79, .5), n(77, .5), n(76, .5), n(74, .5)}, {n(76, 1), n(72, 1)},
		{n(72, .25), n(74, .25), n(76, .25), n(77, .25), n(79, 1)}, {n(81, .5), n(77, .5), n(74, 1)}, {n(79, .5), n(76, .5), n(74, .5), n(71, .5)}, {n(72, 1), n(R, 1)},
	}
	type chord struct{ root, a, b float64 }
	C, F, G, Dm := chord{48, 64, 67}, chord{41, 65, 69}, chord{43, 62, 67}, chord{38, 62, 65}
	chords := []chord{C, C, Dm, G, C, C, G, C, F, Dm, G, C, C, F, G, C}
	bars := 32
	b := newBuf(float64(bars*2)*beat + 0.01)
	for bar := 0; bar < bars; bar++ {
		k := bar % 16
		t0 := float64(bar*2) * beat
		ch := chords[k]
		// oom: root on one, fifth on two
		b.add(t0, beat*0.8, tuba(ch.root, beat*0.8), 0.34)
		fifth := ch.root + 7
		if fifth > 50 {
			fifth -= 12
		}
		b.add(t0+beat, beat*0.8, tuba(fifth, beat*0.8), 0.3)
		// pah: chords on the off-beats, with a snare tick
		for j := 0; j < 2; j++ {
			tt := t0 + (float64(j)+0.5)*beat
			b.add(tt, beat*0.35, calliope(ch.a, beat*0.35), 0.07)
			b.add(tt, beat*0.35, calliope(ch.b, beat*0.35), 0.07)
			b.add(tt, 0.06, noise(0.06, 70), 0.06)
		}
		t := t0
		up := 0.0
		if bar >= 16 {
			up = 12 // second time round, an octave up with a harmony below
		}
		for _, nt := range melody[k] {
			d := nt.beats * beat
			if nt.m > 0 {
				b.add(t, d*0.88, calliope(nt.m+up-12, d*0.88), 0.2)
				b.add(t, d*0.88, square(nt.m-12, d*0.88), 0.035)
				if bar >= 16 {
					b.add(t, d*0.88, calliope(nt.m-16, d*0.88), 0.08)
				}
			}
			t += d
		}
		if bar%4 == 3 { // a cymbal crash into each phrase
			b.add(t0+2*beat-0.02, 0.5, noise(0.5, 6), 0.05)
		}
	}
	return b.pcm(1.6)
}

// Effects makes every sound effect, by name, as 16-bit stereo PCM.
func Effects() map[string][]byte {
	m := map[string][]byte{}
	mk := func(name string, sec, gain float64, f func(b buf)) {
		b := newBuf(sec)
		f(b)
		m[name] = b.pcm(gain)
	}
	mk("click", 0.06, 1, func(b buf) { b.add(0, 0.05, bell(96, 0.05), 0.35) })
	mk("tab", 0.08, 1, func(b buf) { b.add(0, 0.07, square(84, 0.07), 0.15) })
	mk("place", 0.5, 1, func(b buf) {
		b.add(0, 0.18, sweep(160, 55, 0.18), 0.7)
		b.add(0, 0.05, noise(0.05, 60), 0.3)
		for i, n := range []float64{84, 88, 91} {
			b.add(0.1+float64(i)*0.06, 0.25, bell(n, 0.25), 0.18)
		}
	})
	mk("sell", 0.35, 1, func(b buf) {
		b.add(0, 0.2, sweep(400, 120, 0.2), 0.4)
		b.add(0.12, 0.2, bell(88, 0.2), 0.2)
	})
	mk("coin", 0.3, 1, func(b buf) {
		b.add(0, 0.07, square(83, 0.07), 0.16)
		b.add(0.06, 0.22, square(88, 0.22), 0.16)
	})
	mk("cash", 0.6, 1, func(b buf) {
		b.add(0, 0.08, noise(0.08, 40), 0.35)
		b.add(0.06, 0.5, bell(91, 0.5), 0.3)
		b.add(0.09, 0.5, bell(96, 0.5), 0.2)
	})
	mk("drumroll", 1.3, 1, func(b buf) {
		for i := 0; i < 40; i++ {
			t := float64(i) * 0.026
			b.add(t, 0.04, noise(0.04, 90), 0.08+0.18*float64(i)/40)
		}
		b.add(1.05, 0.25, noise(0.25, 10), 0.4)
		b.add(1.05, 0.2, tuba(36, 0.2), 0.5)
	})
	clap := func(b buf, n int, dur, amp float64, seed int64) {
		r := rand.New(rand.NewSource(seed))
		for i := 0; i < n; i++ {
			t := r.Float64() * dur
			fade := 1.0
			if t > dur*0.7 {
				fade = (dur - t) / (dur * 0.3)
			}
			b.add(t, 0.03, noise(0.03, 120+r.Float64()*80), amp*fade*(0.5+r.Float64()*0.5))
		}
	}
	mk("applause", 2.2, 1, func(b buf) {
		clap(b, 260, 2.1, 0.22, 3)
		b.add(0.1, 0.18, sweep(900, 1500, 0.18), 0.06) // a whistle from the back
	})
	mk("polite", 1.2, 1, func(b buf) { clap(b, 40, 1.1, 0.18, 5) })
	mk("flub", 0.9, 1, func(b buf) {
		b.add(0, 0.7, sweep(1100, 260, 0.7), 0.3) // slide whistle
		b.add(0.7, 0.18, tuba(40, 0.18), 0.4)
	})
	mk("gate", 1.6, 1, func(b buf) {
		// ta-ta-ta-taaa on the horns
		for i, n := range []float64{67, 67, 67, 72} {
			d := 0.12
			if i == 3 {
				d = 0.9
			}
			b.add(float64(i)*0.14, d, square(n, d), 0.13)
			b.add(float64(i)*0.14, d, calliope(n+12, d), 0.12)
		}
		b.add(0.42, 0.9, noise(0.9, 5), 0.05)
	})
	mk("fanfare", 1.2, 1, func(b buf) {
		for i, n := range []float64{67, 72, 76} {
			b.add(float64(i)*0.1, 0.15, square(n, 0.15), 0.12)
		}
		b.add(0.3, 0.8, square(79, 0.8), 0.12)
		b.add(0.3, 0.8, square(84, 0.8), 0.08)
	})
	mk("error", 0.2, 1, func(b buf) { b.add(0, 0.16, square(40, 0.16), 0.3) })
	mk("close", 1.6, 1, func(b buf) {
		for i, n := range []float64{84, 79, 76, 72} {
			b.add(float64(i)*0.22, 0.9, bell(n, 0.9), 0.25)
		}
	})
	return m
}
