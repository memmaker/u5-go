package game

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// copyGame copies the game files to a temporary folder, the GOG starting saves among them,
// so a test can write saves without touching the originals.
func copyGame(t *testing.T) string {
	src := gameDir(t)
	dst := t.TempDir()
	for _, from := range []string{src, filepath.Join(src, "__support", "save")} {
		entries, _ := os.ReadDir(from)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			contents, err := os.ReadFile(filepath.Join(from, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dst, e.Name()), contents, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dst
}

// startGame installs a copy of the game and starts it headless.
func startGame(t *testing.T) {
	if err := Install(copyGame(t)); err != nil {
		t.Fatal(err)
	}
	host.StartHeadless(Run)
}

// runTicks steps the game n host ticks; false once it ended.
func runTicks(n int) bool {
	for range n {
		if !host.Step() {
			return false
		}
	}
	return true
}

// press types the keys and lets the game read them.
func press(keys string) {
	for i := range len(keys) {
		host.PushChar(keys[i])
		runTicks(10)
	}
}

// snapshot writes the screen to $U5SHOTS/name.png when U5SHOTS is set.
func snapshot(t *testing.T, name string) {
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
	img := &image.Paletted{Pix: frame.Pixels, Stride: host.ScreenWidth, Rect: image.Rect(0, 0, host.ScreenWidth, host.ScreenHeight), Palette: palette}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// atMenu says whether the main menu waits for a choice (the cursor after "Select: ").
func atMenu() bool { return windowIndex == 0 && window.x == 23 && window.y == 15 }

// toMenu skips the title sequence with spaces until the main menu waits.
func toMenu(t *testing.T) {
	for range 100 {
		if atMenu() {
			return
		}
		host.PushChar(' ')
		for range 20 {
			if runTicks(1); atMenu() {
				return
			}
		}
	}
	t.Fatal("no menu")
}

// TestTitleScreen plays the title sequence, the demo and the menu.
func TestTitleScreen(t *testing.T) {
	startGame(t)
	runTicks(200)
	snapshot(t, "title1")
	runTicks(800)
	snapshot(t, "title2")
	press(" ")
	runTicks(200)
	snapshot(t, "title3")
	press(" ")
	runTicks(100)
	snapshot(t, "menu")
}

// TestCreateCharacter answers the gypsy and checks the new saved game.
func TestCreateCharacter(t *testing.T) {
	startGame(t)
	if !runTicks(100) {
		t.Fatal("the game ended")
	}
	toMenu(t)
	snapshot(t, "create0")
	press("C")
	runTicks(100)
	snapshot(t, "create1")
	press("Avatar\r")
	press("M")
	runTicks(50)
	snapshot(t, "create2")
	press(" ")
	runTicks(50)
	snapshot(t, "create3")
	for range 7 {
		press("A")
		runTicks(20)
	}
	snapshot(t, "create4")
	press(" ")
	runTicks(100)
	snapshot(t, "create5")
	contents, err := disk.ReadFile("SAVED.GAM")
	if err != nil {
		t.Fatal(err)
	}
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	if got := party[0].nameString(); got != "Avatar" {
		t.Fatalf("name %q", got)
	}
	t.Logf("str %d dex %d int %d", party[0].str, party[0].dex, party[0].intel)
}

// TestNewGame creates a character and journeys onward into Britannia.
func TestNewGame(t *testing.T) {
	startGame(t)
	runTicks(100)
	toMenu(t)
	press("C")
	runTicks(100)
	press("Avatar\r")
	press("M")
	runTicks(50)
	press(" ")
	runTicks(50)
	for range 7 {
		press("A")
		runTicks(20)
	}
	press(" ")
	runTicks(100)
	toMenu(t)
	press("J")
	if !runTicks(300) {
		t.Fatal("the game ended")
	}
	snapshot(t, "newgame")
	if mapID == introMap || mapID == journeyMap || party[0].nameString() != "Avatar" {
		t.Fatalf("map %d hero %q", mapID, party[0].nameString())
	}
	t.Logf("map %d level %d at (%d,%d)", mapID, level, mapX, mapY)
}
