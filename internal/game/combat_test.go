package game

import (
	"testing"

	"github.com/memmaker/u5-go/internal/host"
)

// TestCombat fights an orc band outdoors until the battle ends.
func TestCombat(t *testing.T) {
	startOutdoors(t)
	runTicks(100)
	orc := -1
	for i, name := range fighterNames {
		if name == "Orc" {
			orc = i
		}
	}
	if orc < 0 {
		t.Fatalf("no orc in %q", foeNames)
	}
	slot := freeActorSlot()
	actors[slot] = actor{tile: byte(actorWizard + orc*4), animTile: byte(actorWizard + orc*4), x: mapX, y: mapY - 1, z: level}
	press("a")
	runTicks(20)
	host.KeyQueue = append(host.KeyQueue, host.Key{Scan: host.ScanUp})
	runTicks(200)
	if mapID != combatMapID {
		t.Fatalf("no combat: map %d", mapID)
	}
	snapshot(t, "combat")
	for turn := 0; turn < 3000 && mapID == combatMapID; turn++ {
		if int(combatant) >= maxPartySize || combatants[combatant].flags&fighterPlayer == 0 {
			runTicks(10)
			continue
		}
		me := &combatants[combatant]
		if crosshairOn != 0 {
			// a second weapon aims too
			for i := firstFoeSlot; i < combatantSlots; i++ {
				f := &combatants[i]
				if f.flags&fighterMonster != 0 && f.flags&fighterDead == 0 && combatDistance(int(f.x), int(f.y), int(me.x), int(me.y)) == 1 {
					step(t, "aim", wayKey(int(f.x)-int(me.x), int(f.y)-int(me.y)))
					break
				}
			}
			step(t, "fire", charKey('\r'))
			step(t, "fire", charKey(keyEscape))
			continue
		}
		nearest, best := -1, noDistance
		for i := firstFoeSlot; i < combatantSlots; i++ {
			f := &combatants[i]
			if f.flags&fighterMonster != 0 && f.flags&fighterDead == 0 {
				if d := combatDistance(int(f.x), int(f.y), int(me.x), int(me.y)); d < best {
					nearest, best = i, d
				}
			}
		}
		if nearest < 0 {
			// won: leave the map northwards, around whoever is in the way
			switch {
			case !blockedFor(int(me.x), int(me.y)-1):
				step(t, "leave", wayKey(0, -1))
			case !blockedFor(int(me.x)-1, int(me.y)):
				step(t, "leave", wayKey(-1, 0))
			default:
				step(t, "leave", wayKey(1, 0))
			}
			continue
		}
		f := &combatants[nearest]
		dx, dy := int(f.x)-int(me.x), int(f.y)-int(me.y)
		if best == 1 {
			step(t, "attack", charKey('a'), wayKey(dx, dy), charKey('\r'))
		} else if sx, sy, ok := firstStep(int(me.x), int(me.y), int(f.x), int(f.y)); ok {
			step(t, "move", wayKey(sx, sy))
		} else {
			step(t, "pass", charKey(' '))
		}
		if turn%25 == 0 {
			snapshot(t, "combat")
		}
	}
	snapshot(t, "combat-end")
	if mapID != 0 {
		for i, f := range combatants {
			if f.flags != 0 {
				t.Logf("%d %+v actor %+v", i, f, actors[f.actorSlot])
			}
		}
		t.Logf("combatant %d", combatant)
		t.Fatalf("still in combat (map %d)", mapID)
	}
}

func wayKey(dx, dy int) host.Key {
	switch {
	case dx < 0 && dy < 0:
		return host.Key{Scan: host.ScanHome}
	case dx > 0 && dy < 0:
		return host.Key{Scan: host.ScanPageUp}
	case dx < 0 && dy > 0:
		return host.Key{Scan: host.ScanEnd}
	case dx > 0 && dy > 0:
		return host.Key{Scan: host.ScanPageDown}
	case dx < 0:
		return host.Key{Scan: host.ScanLeft}
	case dx > 0:
		return host.Key{Scan: host.ScanRight}
	case dy < 0:
		return host.Key{Scan: host.ScanUp}
	}
	return host.Key{Scan: host.ScanDown}
}

// charKey returns the key that types c.
func charKey(c byte) host.Key {
	host.PushChar(c)
	k := host.KeyQueue[len(host.KeyQueue)-1]
	host.KeyQueue = host.KeyQueue[:len(host.KeyQueue)-1]
	return k
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

func blockedFor(x, y int) bool {
	return fighterAt(x, y) >= 0 || !fighterMayStand(int(actors[combatants[combatant].actorSlot].tile), x, y)
}

// firstStep finds the first orthogonal step of a shortest free path from (x, y) to a cell
// next to (tx, ty).
func firstStep(x, y, tx, ty int) (int, int, bool) {
	type cell struct{ x, y, fx, fy int }
	seen := map[[2]int]bool{{x, y}: true}
	queue := []cell{}
	ways := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, w := range ways {
		nx, ny := x+w[0], y+w[1]
		if onCombatMap(nx, ny) && !blockedFor(nx, ny) {
			seen[[2]int{nx, ny}] = true
			queue = append(queue, cell{nx, ny, w[0], w[1]})
		}
	}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if combatDistance(c.x, c.y, tx, ty) == 1 {
			return c.fx, c.fy, true
		}
		for _, w := range ways {
			nx, ny := c.x+w[0], c.y+w[1]
			if onCombatMap(nx, ny) && !seen[[2]int{nx, ny}] && !blockedFor(nx, ny) {
				seen[[2]int{nx, ny}] = true
				queue = append(queue, cell{nx, ny, c.fx, c.fy})
			}
		}
	}
	return 0, 0, false
}
