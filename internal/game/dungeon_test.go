package game

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// startInDungeon starts a game from INIT.GAM at the entrance of dungeon place and enters it.
func startInDungeon(t *testing.T, place int) {
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
	mapID, level = 0, 0
	mapX, mapY = placeX[place], placeY[place]
	// the entrances stay sealed until their words of power are spoken
	for i := range shrineStates {
		shrineStates[i] = 1
	}
	if err := os.WriteFile(filepath.Join(dir, "SAVED.GAM"), encodeSave(), 0o644); err != nil {
		t.Fatal(err)
	}
	startGame2(t, dir, "J")
	runTicks(100)
	press("E")
	alive := runTicks(200)
	snapshot(t, "dungeon-enter")
	if int(mapID) != place+1 {
		t.Fatalf("alive %v map %d, want %d", alive, mapID, place+1)
	}
}

// TestDungeonWalk lights a torch in Deceit, walks, turns, looks and views a gem.
func TestDungeonWalk(t *testing.T) {
	startInDungeon(t, firstDungeon)
	key := func(c byte) host.Key {
		host.PushChar(c)
		k := host.KeyQueue[len(host.KeyQueue)-1]
		host.KeyQueue = host.KeyQueue[:len(host.KeyQueue)-1]
		return k
	}
	step(t, "ignite", key('i'))
	if torchLight == 0 && lightSpell == 0 {
		t.Fatal("no light")
	}
	snapshot(t, "dungeon-lit")
	startX, startY, startFacing := mapX, mapY, facing
	step(t, "turn right", scan(host.ScanRight))
	if facing != (startFacing+1)&3 {
		t.Fatalf("facing %d after turning right from %d", facing, startFacing)
	}
	step(t, "turn left", scan(host.ScanLeft))
	step(t, "advance", scan(host.ScanUp))
	step(t, "back up", scan(host.ScanDown))
	t.Logf("start (%d,%d) now (%d,%d) level %d facing %d", startX, startY, mapX, mapY, level, facing)
	step(t, "look", key('l'), key('1'), key('\r'), scan(host.ScanUp))
	snapshot(t, "dungeon-look")
	if mapID < firstDungeonMap {
		t.Fatalf("left the dungeon: map %d", mapID)
	}
}

// TestDungeonFuzz presses random keys in a dungeon and fails on a panic.
func TestDungeonFuzz(t *testing.T) {
	if testing.Short() {
		t.Skip("long")
	}
	for _, place := range []int{firstDungeon, firstDungeon + 3, firstDungeon + 6} {
		startInDungeon(t, place)
		r := rand.New(rand.NewSource(int64(place)))
		keys := []host.Key{scan(host.ScanUp), scan(host.ScanUp), scan(host.ScanUp), scan(host.ScanUp), scan(host.ScanLeft),
			scan(host.ScanRight), scan(host.ScanDown)}
		for _, c := range "aaillvxyn1 \r" {
			host.PushChar(byte(c))
			keys = append(keys, host.KeyQueue[len(host.KeyQueue)-1])
			host.KeyQueue = host.KeyQueue[:len(host.KeyQueue)-1]
		}
		for i := 0; i < 3000; i++ {
			host.KeyQueue = append(host.KeyQueue, keys[r.Intn(len(keys))])
			if !runTicks(15) {
				break
			}
		}
		t.Logf("place %d: map %d level %d at (%d,%d)", place, mapID, level, mapX, mapY)
	}
}
