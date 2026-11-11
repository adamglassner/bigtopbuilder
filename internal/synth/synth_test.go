package synth

import (
	"testing"
	"time"
)

func TestMake(t *testing.T) {
	start := time.Now()
	m := March()
	fx := Effects()
	t.Logf("march %.1fs of audio, %d effects, made in %v", float64(len(m))/4/sampleRate, len(fx), time.Since(start))
	if len(m) == 0 || len(m)%4 != 0 || len(fx) < 10 {
		t.Fatal("bad audio")
	}
	for _, n := range []string{"click", "place", "coin", "drumroll", "applause", "polite", "flub", "gate", "fanfare", "error", "sell", "cash", "close", "tab"} {
		if len(fx[n]) == 0 {
			t.Fatalf("missing %s", n)
		}
	}
}
