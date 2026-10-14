package game

import (
	"math/rand"
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// TestShops talks to each merchant of Britain who is open and presses random shop keys.
func TestShops(t *testing.T) {
	startInBritain(t)
	startGold := gold
	r := rand.New(rand.NewSource(1))
	opened := 0
	for who := range npcs {
		n := &npcs[who]
		if n.mood < firstShopMood || n.mood >= moodFleeing || n.moveState == 0 || int(n.z) != int(level) {
			continue
		}
		a := &actors[n.actorAt]
		mapX, mapY = a.x, a.y+1
		shopName = ""
		host.PushChar('T')
		runTicks(10)
		host.KeyQueue = append(host.KeyQueue, scan(host.ScanUp))
		if !runTicks(60) {
			t.Fatal("the game ended")
		}
		if shopName == "" {
			continue
		}
		opened++
		t.Logf("merchant %d: %s (%s)", who, shopName, shopOwner)
		snapshot(t, "shop")
		for range 150 {
			host.PushChar("BSYN123AQ \r"[r.Intn(11)])
			if !runTicks(10) {
				t.Fatal("the game ended")
			}
		}
		for range 40 {
			for _, c := range []byte{'\x1b', '\x1b', 'N', '\x1b'} {
				host.PushChar(c)
				runTicks(5)
			}
		}
		snapshot(t, "shop-left"+string(rune('a'+who)))
	}
	t.Logf("%d shops, gold %d -> %d", opened, startGold, gold)
	if opened == 0 {
		t.Fatal("no shop opened")
	}
}
