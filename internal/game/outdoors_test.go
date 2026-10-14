package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// startOutdoors starts a game saved outside Iolo's hut and journeys onward.
func startOutdoors(t *testing.T) {
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
	loadOutdoorTables()
	copy(party[0].name[:], "Avatar\x00")
	mapID, level = 0, 0
	const iolosHut = 12
	mapX, mapY = placeX[iolosHut], placeY[iolosHut]+1
	if err := os.WriteFile(filepath.Join(dir, "SAVED.GAM"), encodeSave(), 0o644); err != nil {
		t.Fatal(err)
	}
	startGame2(t, dir, "J")
}

// startGame2 installs dir and starts the game with keys typed ahead.
func startGame2(t *testing.T, dir, keys string) {
	if err := Install(dir); err != nil {
		t.Fatal(err)
	}
	host.StartHeadless(Run)
	for i := range len(keys) {
		host.PushChar(keys[i])
	}
	runTicks(1)
}

// TestOutdoors walks around Britannia.
func TestOutdoors(t *testing.T) {
	startOutdoors(t)
	runTicks(100)
	snapshot(t, "out1")
	for _, scan := range []byte{host.ScanDown, host.ScanDown, host.ScanRight, host.ScanRight, host.ScanRight, host.ScanDown, host.ScanDown} {
		host.KeyQueue = append(host.KeyQueue, host.Key{Scan: scan})
		runTicks(30)
	}
	snapshot(t, "out2")
	t.Logf("at %d,%d", mapX, mapY)
}
