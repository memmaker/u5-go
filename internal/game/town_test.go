package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// TestTown creates a character and journeys onward into the hut it starts in, then walks
// around and out of it.
func TestTown(t *testing.T) {
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
	snapshot(t, "town0")
	t.Logf("map %d level %d at %d,%d", mapID, level, mapX, mapY)
	if mapID == 0 {
		t.Fatal("not in a settlement")
	}
	for i, scan := range []byte{host.ScanDown, host.ScanDown, host.ScanDown, host.ScanLeft, host.ScanLeft, host.ScanUp} {
		host.KeyQueue = append(host.KeyQueue, host.Key{Scan: scan})
		if !runTicks(30) {
			t.Fatalf("the game ended at step %d", i)
		}
	}
	snapshot(t, "town1")
	t.Logf("map %d level %d at %d,%d hour %d", mapID, level, mapX, mapY, hour)
	for range 40 {
		press(" ")
		if !runTicks(10) {
			t.Fatal("the game ended while waiting")
		}
	}
	snapshot(t, "town2")
}

// TestTalk enters Britain from Britannia and talks to one of its people.
func TestTalk(t *testing.T) {
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
	const britain = 2
	mapID, level = 0, 0
	mapX, mapY = placeX[britain-1], placeY[britain-1]
	if err := os.WriteFile(filepath.Join(dir, "SAVED.GAM"), encodeSave(), 0o644); err != nil {
		t.Fatal(err)
	}
	startGame2(t, dir, "J")
	runTicks(100)
	press("E")
	runTicks(200)
	snapshot(t, "britain0")
	if mapID != britain {
		t.Fatalf("map %d", mapID)
	}
	who := -1
	for w := 1; w < npcSlots; w++ {
		n := npcs[w]
		if n.actorAt != 0 && n.mood != 0 && n.mood < scriptLetter && actors[n.actorAt].z == level {
			who = w
			break
		}
	}
	if who < 0 {
		t.Fatal("nobody to talk to")
	}
	a := actors[npcs[who].actorAt]
	mapX, mapY = a.x+1, a.y
	viewChanged = 1
	t.Logf("talking to %d (tile %d) at %d,%d", who, a.tile, a.x, a.y)
	press("T")
	host.KeyQueue = append(host.KeyQueue, host.Key{Scan: host.ScanLeft})
	runTicks(100)
	snapshot(t, "talk0")
	press("name\r")
	runTicks(100)
	snapshot(t, "talk1")
	press("job\r")
	runTicks(100)
	snapshot(t, "talk2")
	press("\r")
	runTicks(100)
	snapshot(t, "talk3")
}
