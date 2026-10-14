package host

import "github.com/hajimehoshi/ebiten/v2"

// Dungeon is what the 3D view shows.
type Dungeon struct{}

// DungeonView returns the dungeon around the party, or nil when the map window shows
// something else.
var DungeonView func() *Dungeon

func (g *game) drawDungeon(screen *ebiten.Image, d *Dungeon) {}
