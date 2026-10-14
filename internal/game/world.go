package game

// The map in memory: the loaded part of Britannia (four 16x16 chunks), or a 32x32
// settlement, or the 11x11 combat map; the visible view; and the tile animation table.

const (
	chunkSize      = 16
	chunksAcross   = 2
	settlementSize = 32
	mapBytes       = settlementSize * settlementSize
	viewSize       = 11 // the map view's tiles across and down
	viewCentre     = 5
	viewStride     = 32 // the row length of the view and combat buffers
	actorMapStride = 16
	worldWrapMask  = 31
)

// The map buffers.
var (
	// mapTiles is the loaded map: four chunks of Britannia (top-left, top-right,
	// bottom-left, bottom-right) or a settlement level, rows of 32 (D_6608).
	mapTiles [mapBytes]byte // C: D_6608
	// mapBackup is a copy of the map kept while it changes (D_6708).
	mapBackup [mapBytes]byte // C: D_6708
	// viewTiles is what is visible of the map around the party: 0 marks a tile drawn
	// over by an actor, 255 a hidden one (D_ab02).
	viewTiles [viewSize * viewStride]byte // C: D_ab02
	// actorTiles is the actor shown on each view tile (D_ac64).
	actorTiles [viewSize * actorMapStride]byte // C: D_ac64
	// combatMap is the combat map, rows of 32: the 11x11 tiles and, beyond them, the
	// map's starting positions and triggers (D_ad14).
	combatMap [mapBytes]byte // C: D_ad14
	// tileShown is each map tile's animation frame (D_b11e).
	tileShown [mapTileCount]byte // C: D_b11e
	// animationStep counts the slower animations (D_6a7e).
	animationStep byte // C: D_6a7e
)

// lastMapByte is the byte that reads and writes off a settlement map land in: the
// original returns the address of the map's last byte for them.
const lastMapByte = mapBytes - 1

// viewTile returns the visible tile at view position x, y.
func viewTile(x, y int) byte { return viewTiles[y*viewStride+x] }

// actorAtView returns the actor tile at view position x, y.
func actorAtView(x, y int) byte { return actorTiles[y*actorMapStride+x] }

// tileAt returns the map byte for x, y: on Britannia relative to the loaded chunks, off a
// settlement the map's last byte.
// C: ULTIMA_4402_GetTileAddr
func tileAt(x, y int) *byte {
	switch {
	case mapID >= firstMapCombat:
		if y < 0 {
			// fix: the original reads before the combat map for rows above it; it is
			// given the spare byte instead.
			return &mapTiles[lastMapByte]
		}
		return &combatMap[y*viewStride+x]
	case mapID == 0:
		x = (x - int(chunkX)) & worldWrapMask
		y = (y - int(chunkY)) & worldWrapMask
		chunk := 0
		if y >= chunkSize {
			chunk = chunksAcross
			y -= chunkSize
		}
		if x >= chunkSize {
			chunk++
			x -= chunkSize
		}
		return &mapTiles[chunk*chunkSize*chunkSize+y*chunkSize+x]
	case x < 0 || y < 0 || x >= settlementSize || y >= settlementSize:
		return &mapTiles[lastMapByte]
	}
	return &mapTiles[y*settlementSize+x]
}
