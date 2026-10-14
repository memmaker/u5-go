package game

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/ega"
	"github.com/memmaker/u5-go/internal/host"
)

// TestModernPrompts checks that a town's command loop asks for a command, offers a menu,
// and draws a wider map by day than the original's view.
func TestModernPrompts(t *testing.T) {
	startInBritain(t)
	runTicks(20)
	if prompt != promptCommand || !host.InWorld {
		t.Fatalf("prompt %d, in world %v", prompt, host.InWorld)
	}
	m := promptMenu(true)
	if m == nil || len(m.Items) < 3 {
		t.Fatalf("command menu %+v", m)
	}
	for _, it := range m.Items {
		t.Log(it.Label)
	}
	if g := glyphs(); len(g) < 128 {
		t.Fatalf("%d glyphs", len(g))
	}
	light = dayLight
	const cols, rows = 25, 15
	wide := wideView(cols, rows)
	if wide == nil {
		t.Fatal("no wide view by day in a town")
	}
	dir := os.Getenv("U5SHOTS")
	if dir == "" {
		return
	}
	frame := screen.Frame()
	palette := make(color.Palette, 16)
	for i := range palette {
		level := func(c int) uint8 { return uint8(int(frame.Palette[i*3+c]) * math.MaxUint8 / 63) }
		palette[i] = color.RGBA{level(0), level(1), level(2), math.MaxUint8}
	}
	w := cols * ega.TileSize
	img := &image.Paletted{Pix: wide, Stride: w, Rect: image.Rect(0, 0, w, rows*ega.TileSize), Palette: palette}
	f, err := os.Create(filepath.Join(dir, "wide.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, img)
}
