package game

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// TestCastSaveLoad casts In Lor in Britain, saves, and journeys on from the save.
func TestCastSaveLoad(t *testing.T) {
	startInBritain(t)
	for i := range spellMixtures {
		spellMixtures[i] = 9
	}
	party[0].magic = 30
	before := spellMixtures
	press("C")
	runTicks(20)
	press("\r")
	runTicks(20)
	for _, c := range "IL\r" {
		press(string(c))
		runTicks(20)
	}
	if !runTicks(200) {
		t.Fatal("the game ended")
	}
	snapshot(t, "cast")
	if lightSpell == 0 {
		t.Fatal("In Lor gave no light")
	}
	used := -1
	for i := range spellMixtures {
		if spellMixtures[i] != before[i] {
			used = i
		}
	}
	if used < 0 {
		t.Fatal("no mixture used")
	}
	x, y, gotGold := mapX, mapY, gold
	press("Q")
	runTicks(20)
	press("Y")
	runTicks(100)
	contents, err := os.ReadFile(filepath.Join(disk.Dir, "SAVED.GAM"))
	if err != nil {
		t.Fatal(err)
	}
	mapX, mapY, gold = 0, 0, 0
	if err := loadSave(contents); err != nil {
		t.Fatal(err)
	}
	if mapX != x || mapY != y || gold != gotGold || spellMixtures[used] != before[used]-1 {
		t.Fatalf("saved at (%d,%d) gold %d, want (%d,%d) gold %d", mapX, mapY, gold, x, y, gotGold)
	}
	startGame2(t, disk.Dir, "J")
	if !runTicks(200) {
		t.Fatal("the game ended")
	}
	snapshot(t, "loaded")
	if mapID != 2 || mapX != x || mapY != y {
		t.Fatalf("loaded map %d at (%d,%d)", mapID, mapX, mapY)
	}
}

// TestFuzz presses random keys, commands included, in Britain and outdoors and fails on a
// panic.
func TestFuzz(t *testing.T) {
	if testing.Short() {
		t.Skip("long")
	}
	keys := []host.Key{scan(host.ScanUp), scan(host.ScanDown), scan(host.ScanLeft), scan(host.ScanRight)}
	for _, c := range "ABCDEFGHIJKLMNOPRSTUVWXYZ0123456789 \r\x1b" {
		host.PushChar(byte(c))
		keys = append(keys, host.KeyQueue[len(host.KeyQueue)-1])
		host.KeyQueue = host.KeyQueue[:len(host.KeyQueue)-1]
	}
	for seed := range int64(4) {
		if seed%2 == 0 {
			startInBritain(t)
		} else {
			startOutdoors(t)
		}
		r := rand.New(rand.NewSource(seed))
		for i := range 4000 {
			host.KeyQueue = append(host.KeyQueue, keys[r.Intn(len(keys))])
			if !runTicks(15) {
				t.Logf("seed %d: the game ended after %d keys", seed, i)
				break
			}
		}
		t.Logf("seed %d: map %d level %d at (%d,%d)", seed, mapID, level, mapX, mapY)
	}
}

// TestQuietSteps walks in Britain: the steps leave the prompt standing rather than a line
// each, and the message history keeps the scroll's lines.
func TestQuietSteps(t *testing.T) {
	startInBritain(t)
	host.History = nil
	for _, s := range []byte{host.ScanLeft, host.ScanRight, host.ScanLeft, host.ScanRight} {
		host.KeyQueue = append(host.KeyQueue, host.Key{Scan: s})
		runTicks(20)
	}
	press("K") // a command that prints, so the history has a line
	runTicks(40)
	press("\x1b")
	runTicks(20)
	lines := append(host.History, string(historyLine))
	directions := 0
	for _, l := range lines {
		if strings.Contains(l, "West") || strings.Contains(l, "East") {
			directions++
		}
	}
	if directions > 1 {
		t.Fatalf("%d direction lines for 4 steps: %q", directions, lines)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Klimb") {
		t.Fatalf("no Klimb in the history: %q", lines)
	}
}
