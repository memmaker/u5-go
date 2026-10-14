package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// startEndgame loads INIT.GAM and runs the ending, with or without the Sandalwood Box.
func startEndgame(t *testing.T, withBox bool) {
	t.Helper()
	dir := copyGame(t)
	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(filepath.Join(dir, "INIT.GAM"))
	if err != nil {
		t.Fatal(err)
	}
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	copy(party[0].name[:], "Avatar\x00")
	sandalwoodBox = 0
	if withBox {
		sandalwoodBox = 1
	}
	host.StartHeadless(func() {
		defer func() {
			if r := recover(); r != nil && r != errQuit {
				panic(r)
			}
		}()
		for i := range tileShown {
			tileShown[i] = byte(i)
		}
		loadFont("ibm.ch", 0)
		loadFont("runes.ch", 1)
		selectFont(0)
		setWindowSize(0, 0, 0, textColumns-1, textRows-1)
		setWindowSize(1, 24, 1, 39, 9)
		setWindowSize(2, 24, 11, 39, 23)
		for !loadTiles(tileFile) {
		}
		drawFrame()
		selectWindow(2)
		mapID = combatMapID
		mapShown = 1
		showStatus()
		endgame()
	})
	runTicks(1)
}

// TestEndgame plays the good ending through to the certificate.
func TestEndgame(t *testing.T) {
	startEndgame(t, true)
	for range 12 {
		runTicks(400)
		press(" ")
	}
	press("Y")
	for range 30 {
		runTicks(300)
		press(" ")
	}
	snapshot(t, "endgame-certificate")
	if len(certificate) != 0 {
		t.Fatalf("certificate line left unprinted: %q", certificate)
	}
}

// TestEndgameWithoutBox plays the bad ending for a while.
func TestEndgameWithoutBox(t *testing.T) {
	startEndgame(t, false)
	for range 10 {
		runTicks(400)
		press("Y")
	}
	runTicks(500)
	snapshot(t, "endgame-bad")
}
