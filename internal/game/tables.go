package game

import "github.com/memmaker/u5-go/internal/data"

// The game's constant tables and messages live in its data segment, DATA.OVL; they are
// read from there rather than copied into the port.

// segment is DATA.OVL.
var segment *data.Segment

// message returns the zero-terminated message at address in the data segment.
func message(address int) string { return segment.String(address) }

// printMessage prints the message at address.
func printMessage(address int) { printString(message(address)) }

// Addresses of the tables in the data segment.
const (
	blockedTilesAddress  = 21716 // C: D_54d4 (a bit per map tile, set when it blocks walkers)
	movementKindsAddress = 21748 // C: D_54f4 (a movement kind per group of four actor tiles)
	skiffWaterAddress    = 21812 // C: D_5534 (the directions a skiff can leave a river tile)
	skiffBridgeAddress   = 21828 // C: D_5544 (the same for bridge tiles)
	trapKindsAddress     = 21918 // C: D_559e
)

// Tables read from the data segment.
var (
	blockedTiles  []byte // C: D_54d4
	movementKinds []byte // C: D_54f4
	skiffWater    []byte // C: D_5534
	skiffBridge   []byte // C: D_5544
	trapKinds     []byte // C: D_559e
)

// loadTables reads the tables of the data segment.
func loadTables(dataOvl []byte) {
	segment = data.NewSegment(dataOvl)
	blockedTiles = segment.Bytes(blockedTilesAddress, mapTileCount/bitsPerByte)
	movementKinds = segment.Bytes(movementKindsAddress, actorTileGroups)
	skiffWater = segment.Bytes(skiffWaterAddress, riverTileCount)
	skiffBridge = segment.Bytes(skiffBridgeAddress, bridgeTileCount)
	trapKinds = segment.Bytes(trapKindsAddress, trapKindCount)
}

const (
	bitsPerByte     = 8
	mapTileCount    = 256
	actorTileGroups = 64
	riverTileCount  = 16
	bridgeTileCount = 4
	trapKindCount   = 8
)
