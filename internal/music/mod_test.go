package music

import (
	"os"
	"testing"
)

// TestModule plays a few seconds of the Amiga module named by U5_AMIGA_MOD.
func TestModule(t *testing.T) {
	path := os.Getenv("U5_AMIGA_MOD")
	if path == "" {
		t.Skip("U5_AMIGA_MOD not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	p, err := newModPlayer(b)
	if err != nil {
		t.Fatal(err)
	}
	p.start(0)
	l, r := make([]float32, rate*5), make([]float32, rate*5)
	p.render(l, r)
	var peak float32
	for i := range l {
		peak = max(peak, l[i], r[i], -l[i], -r[i])
	}
	if peak == 0 || peak > 2 {
		t.Fatalf("peak %v", peak)
	}
	t.Logf("peak %.2f, %d positions", peak, len(p.order))
}

// TestMIDI renders a few seconds of each song in the directory named by U5_XMI_DIR.
func TestMIDI(t *testing.T) {
	dir := os.Getenv("U5_XMI_DIR")
	if dir == "" || embeddedSF2() == nil {
		t.Skip("U5_XMI_DIR not set or no SoundFont")
	}
	p := loadMIDI(dir)
	if p == nil {
		t.Fatal("no songs")
	}
	for n := range songCount {
		p.start(n)
		l, r := make([]float32, rate*3), make([]float32, rate*3)
		p.render(l, r)
		var peak float32
		for i := range l {
			peak = max(peak, l[i], -l[i])
		}
		if peak == 0 {
			t.Errorf("%s silent", SongFiles[n])
		}
	}
}
