package game

// Segment 5000 and the view helpers of 6000/7000: what of the map the party sees, the
// actors placed on it, and drawing it.

const (
	hiddenTile    = 255 // a view tile out of sight
	actorOverlay  = 0   // a view tile an actor stands on
	viewLast      = viewSize - 1
	actorTileBase = 256 // actor tiles follow the map tiles in the tile set
)

// Addresses of the view tables in the data segment.
const (
	distancesAddress     = 27304 // C: D_6aa8 (the light a tile needs, by distance from the centre)
	opaqueTilesAddress   = 27270 // C: D_6a86
	lightSourcesAddress  = 27290 // C: D_6a9a
	flickerShapesAddress = 8062  // C: D_1f7e
	opaqueTileCount      = 19
	lightSourceCount     = 10
	distanceRow          = 6
	flickerShapeCount    = 16
	flickerShapePoints   = 16
	flickerPairBytes     = 2
	queueSize            = 257
	lightVisible         = -111 // the visibility search's marker for the party's own view
	distanceLimit        = viewCentre
	lightSourceRange     = 10
	viewBufferBytes      = viewSize * viewStride
)

var (
	opaqueTiles   []byte // C: D_6a86
	lightSources  []byte // C: D_6a9a
	flickerShapes []byte // C: D_1f7e (pairs of signed x, y)
)

// loadViewTables reads the view tables.
func loadViewTables() {
	opaqueTiles = segment.Bytes(opaqueTilesAddress, opaqueTileCount)
	lightSources = segment.Bytes(lightSourcesAddress, lightSourceCount)
	flickerShapes = segment.Bytes(flickerShapesAddress, mapTileCount*flickerPairBytes)
	reachableTiles = segment.Bytes(reachableTilesAddress, settlementSize)
}

// lightNeeded returns the light a view tile needs to be seen: less nearer the centre.
// The original reads the table directly, so the data segment is read here too.
// C: ULTIMA_6ff0
func lightNeeded(x, y int) int {
	if x >= viewSize || y >= viewSize {
		return 0
	}
	if x > viewCentre {
		x = viewLast - x
	}
	if y > viewCentre {
		y = viewLast - y
	}
	return int(segment.Byte(distancesAddress + x + y*distanceRow))
}

// shimmer returns a random 0..3 for the frames of furniture with someone on it, 0 while
// time stands still.
// C: ULTIMA_51a0
func shimmer() int {
	if spellIcon == iconTimeStop {
		return 0
	}
	return randomRange(0, 3)
}

// Actor tiles that sit on furniture.
const (
	actorFlyingFirst   = 18
	actorFlyingEnd     = 22
	actorSkiffFirst    = actorSkiff
	actorSkiffEnd      = actorSkiff + 4
	actorPeople        = 64
	actorOnStocksFirst = 96
	actorInStocks      = 100
	actorClimbingUp    = 23
	actorClimbingDown  = 24
	actorInBed         = 26
	actorMirrored      = 60
	actorSeatedFacing  = 52
	actorSeatedBack    = 56
	actorSeated        = 50
	actorSeatedNorth   = 48
	actorMonsterGroup  = 128
	actorSkiffGroup    = 40
)

// placeActor puts actorTile on view cell (column, row) for map position (x, y): people
// and the Avatar take the look of the furniture they stand on.
// C: ULTIMA_51b8
func placeActor(column, row, x, y, actorTile int) {
	if actorTile == actorAvatar || actorTile >= actorRidingHorse && actorTile < actorFlyingEnd ||
		actorTile >= actorPeople || actorTile >= actorSkiffFirst && actorTile < actorSkiffEnd {
		ground := int(*tileAt(x, y))
		if ground == mapTile236 || ground == mapTile10 {
			return
		}
		if ground == mapTile87 {
			viewTiles[row*viewStride+column] = mapTile56
			return
		}
		if ground == mapTile106 || ground == mapTile107 {
			if actorTile&^15 == actorMonsterGroup || actorTile&^3 == actorSkiffGroup {
				return
			}
		} else if actorTile < actorMonsterGroup {
			switch ground {
			case mapTile132:
				actorTile = shimmer() + actorOnStocksFirst
			case mapTile133:
				actorTile = shimmer() + actorInStocks
			case mapTileLadderUp:
				actorTile = actorClimbingUp
			case mapTileLadderDown:
				actorTile = actorClimbingDown
			case mapTileBed:
				actorTile = actorInBed
			case mapTileMirror, mapTileMirror158:
				actorTile = shimmer() + actorMirrored
			case mapTileChair146:
				below := *tileAt(x, y+1)
				if below == mapTileTable154 || below == mapTileTable156 {
					actorTile = shimmer() + actorSeatedFacing
				} else {
					actorTile = actorSeated
				}
			case mapTileChair144:
				above := *tileAt(x, y-1)
				if above == mapTileTable155 || above == mapTileTable156 {
					actorTile = shimmer() + actorSeatedBack
				} else {
					actorTile = actorSeatedNorth
				}
			case mapTileChair145, mapTileChair147:
				actorTile = ground&3 + actorSeatedNorth
			default:
				if *tileAt(x, y-1) == mapTileMirror && row != 0 {
					viewTiles[(row-1)*viewStride+column] = mapTileMirror158
				}
			}
		}
	}
	actorTiles[row*actorMapStride+column] = byte(actorTile)
	viewTiles[row*viewStride+column] = actorOverlay
}

const lampLitDistance = 5

// placeActors puts the party and the visible actors on the view, and lights the lamps
// near the party.
// C: ULTIMA_5394
func placeActors() {
	if mapID < firstMapCombat {
		party := &actors[0]
		party.x, party.y, party.z = mapX, mapY, level
		party.tile, party.animTile = partyTile, partyTile
		for row := range viewSize {
			for column := range viewSize {
				at := &viewTiles[row*viewStride+column]
				if *at == mapTile221 && lightNeeded(column, row) > lampLitDistance {
					*at = mapTile28
				} else if *at == mapTile28 && lightNeeded(column, row) <= lampLitDistance {
					*at = mapTile221
				}
			}
		}
	}
	for slot := actorSlots - 1; slot >= 0; slot-- {
		a := &actors[slot]
		if a.tile == 0 {
			continue
		}
		column, row := a.x, a.y
		if mapID < firstMapCombat {
			column -= mapX - viewCentre
			row -= mapY - viewCentre
			if a.z != level || column > viewLast || row > viewLast {
				continue
			}
		}
		at := &viewTiles[int(row)*viewStride+int(column)]
		if a.animTile == 0 || *at == hiddenTile || *at == mapTile135 {
			continue
		}
		switch {
		case a.tile&^3 == actor232 || a.tile == actorSleep || a.tile == actorDead:
			if *at != actorOverlay {
				actorTiles[int(row)*actorMapStride+int(column)] = a.animTile
				*at = actorOverlay
			}
		case a.animTile == actor29 || a.animTile == actorSleep:
			actorTiles[int(row)*actorMapStride+int(column)] = a.animTile
			*at = actorOverlay
		case a.tile == actorBard:
			if *at == mapTileChair146 {
				actorTiles[int(row)*actorMapStride+int(column)] = a.animTile
				*at = actorOverlay
			} else {
				placeActor(int(column), int(row), int(a.x), int(a.y), int(a.animTile)-8)
			}
		default:
			placeActor(int(column), int(row), int(a.x), int(a.y), int(a.animTile))
		}
	}
}

// Avatar's name's fifth letter that marks the ghost of a test character ('j').
const controlledNameLetter = 4

// computerControlled says whether combatant is steered by the computer: a monster, a
// charmed member, or the test ghost.
// C: ULTIMA_5646
func computerControlled(slot int) bool {
	flags := combatants[slot].flags
	entity := combatants[slot].entity
	if flags&fighterDead != 0 {
		return false
	}
	if flags&fighterPlayer != 0 {
		if entity != 0 && party[entity].name[controlledNameLetter] == 'j' {
			return true
		}
		return flags&fighterCharmed != 0
	}
	return flags&fighterCharmed == 0
}

// drawView draws the view tiles and, in combat, the frame around the fighter whose turn
// it is, blinking, and the crosshair.
// C: ULTIMA_56ac_DrawMap
func drawView() {
	for row := range viewSize {
		for column := range viewSize {
			tile := viewTile(column, row)
			switch {
			case tile == actorOverlay:
				if shown := actorAtView(column, row); shown != actorCircle {
					putTile(actorTileBase+int(shown), column, row)
				}
			case tile == mapTileMoongate && moongateRise != 0 && moongateRise < moongateHeight:
				putMoongate(int(moongateRise), column, row)
			default:
				putTile(int(tileShown[tile]), column, row)
			}
		}
	}
	if mapID < firstMapCombat {
		return
	}
	focusBlink ^= 1
	if focusBlink == 0 || combatant == noCombatant || computerControlled(int(combatant)) {
		return
	}
	turn := &combatants[combatant]
	x := int(turn.x)*tilePixels + viewLeft
	y := int(turn.y)*tilePixels + viewLeft
	setPenColour(brightWhiteColour)
	for i := range 2 {
		for edge := 0; edge < tilePixels-1; edge += tilePixels - 2 {
			horizLine(x, y+i+edge, x+tilePixels-1)
			vertLine(x+i+edge, y, y+tilePixels-1)
		}
	}
	if crosshairOn == 0 {
		return
	}
	x = int(crosshairX)*tilePixels + viewLeft
	y = int(crosshairY)*tilePixels + viewLeft
	for i := range 2 {
		for j := range 2 {
			setPenColour(brightWhiteColour)
			line(j*11+x+2, i*3+y+6, j*3+x+6, i*3+y+6)
			lineTo(j*3+x+6, i*11+y+2)
			setPenColour(0)
			line(j*11+x+2, i*5+y+5, j*5+x+5, i*5+y+5)
			lineTo(j*5+x+5, i*11+y+2)
			line(j*11+x+2, y+i+7, j*3+x+6, y+i+7)
			line(x+j+7, i*3+y+6, x+j+7, i*11+y+2)
		}
	}
}

// skipAnimation marks a frame that must not animate the actors (set after a key was
// read on the Tandy driver).
const skipAnimation = 255

// updateFrame animates the map, works out what is seen, draws it and plays the ambient
// sounds.
// C: ULTIMA_5910_UpdateFrame
func updateFrame() {
	frameKeyHint = noFrameKey
	if spellIcon == iconTimeStop {
		animateFrame = 0
	}
	if mapShown != 0 {
		if animateFrame != 0 {
			if animateFrame != skipAnimation {
				animateActors()
			}
			changeWindSometimes()
			if mapID < firstMapCombat {
				raiseMoongates()
				flickerLights()
			}
		}
		if mapID < firstMapCombat {
			if viewChanged != 0 {
				computeView(int(light), int(mapX)-int(chunkX), int(mapY)-int(chunkY))
				viewChanged = 0
			} else {
				for row := range viewSize {
					for column := range viewSize {
						if viewTile(column, row) == actorOverlay {
							viewTiles[row*viewStride+column] = *tileAt(column+int(mapX)-viewCentre, row+int(mapY)-viewCentre)
						}
					}
				}
			}
		} else {
			copy(viewTiles[:], combatMap[:viewBufferBytes])
		}
		placeActors()
		drawView()
		if animateFrame != 0 {
			playAmbientSound()
		}
	}
	animateFrame = 1
}

// frameKeyHint is cleared each frame (D_545e).
var frameKeyHint byte = noFrameKey // C: D_545e

const noFrameKey = 255

// visibilityQueue is the search's queue of view positions (y, x pairs).
type visibilityQueue struct {
	entries     [queueSize]int
	read, write int
}

func (q *visibilityQueue) push(v int) {
	q.entries[q.write] = v
	q.write++
}

func (q *visibilityQueue) pop() int {
	v := q.entries[q.read]
	q.read++
	return v
}

// Directions the search steps in, from 7 down to 0, each step moving on from the last.
const searchDirections = 8

// searchView spreads out from the view's centre through the tiles that do not block
// sight, writing what is seen into area (rows of 32 from the row top, column left).
// With top lightVisible it works out the party's view by the light; otherwise it marks
// what a light source at (originX, originY) of the loaded map lights, in the combat
// map buffer, using the view as its record of visited tiles.
// C: ULTIMA_5a28
func searchView(lightLevel, originY, originX, top, left int, area []byte) {
	var queue visibilityQueue
	direction := searchDirections - 1
	inside := true
	partyView := false
	if lightLevel <= 0 {
		return
	}
	if top == lightVisible {
		top = 0
		partyView = true
	}
	lightLevel++
	queue.push(viewCentre)
	queue.push(viewCentre)
	area[top*viewStride+left+viewCentre*viewStride+viewCentre] =
		*tileAt(originX+int(chunkX)+viewCentre, originY+int(chunkY)+viewCentre)
	areaRows := settlementSize
	if partyView {
		areaRows = viewSize
	}
	for queue.read != queue.write {
		parentY := queue.pop()
		parentX := queue.pop()
		y, x := parentY, parentX
		for ; direction > -1; direction, inside = direction-1, true {
			switch direction {
			case 0, 1, 7:
				x--
			case 4, 5:
				x++
			case 2, 3:
				y--
			case 6:
				y++
			}
			if x < 0 || x > viewLast || y < 0 || y > viewLast {
				inside = false
			}
			// fix: the original writes outside the buffer here; u5d skips those steps.
			if top+y < 0 || left+x < 0 || top+y >= areaRows || left+x >= viewStride {
				continue
			}
			at := &area[(top+y)*viewStride+left+x]
			if !partyView {
				// fix: the original also clears the visited mark of a step off the
				// view, a stray write outside the 11x11 view.
				if inside && viewTile(x, y) != actorOverlay && top+y <= worldWrapMask {
					viewTiles[y*viewStride+x] = actorOverlay
				} else {
					inside = false
				}
			} else if inside && *at != hiddenTile {
				inside = false
			}
			if !inside {
				continue
			}
			tile := *tileAt(originX+x+int(chunkX), originY+y+int(chunkY))
			switch {
			case lightNeeded(x, y) < lightLevel:
				*at = tile
			case !partyView:
				tile = hiddenTile
			case !letsSightThrough(tile, lightNeeded(x, y)):
				if originY+parentY >= settlementSize || originX+parentX >= settlementSize ||
					originY+y < 0 || originX+parentX < 0 ||
					area[parentY*viewStride+parentX] == 0 ||
					combatMap[(originY+parentY)*viewStride+originX+parentX] == 0 ||
					combatMap[(originY+y)*viewStride+originX+x] == 0 {
					tile = hiddenTile
					*at = hiddenTile
				} else {
					*at = tile
				}
			case originY+y < 0 || originX+x < 0 || originX+x >= settlementSize || originY+y >= settlementSize:
				*at = 0
			case combatMap[(originY+y)*viewStride+originX+x] != 0:
				*at = tile
			default:
				*at = 0
			}
			if letsSightThrough(tile, lightNeeded(x, y)) {
				queue.push(y)
				queue.push(x)
			}
		}
		direction = searchDirections - 1
	}
}

// computeView works out the view around map position (x, y) (relative to the loaded
// chunks) for the light: what is lit and in sight, or with negative light everything.
// C: ULTIMA_5d0a
func computeView(lightLevel, x, y int) {
	for row := range viewSize {
		for column := range viewSize {
			viewTiles[row*viewStride+column] = hiddenTile
		}
	}
	left, top := x-viewCentre, y-viewCentre
	if lightLevel > 0 {
		searchView(lightLevel, top, left, lightVisible, 0, viewTiles[:])
		for row := range viewSize {
			for column := range viewSize {
				at := &viewTiles[row*viewStride+column]
				if *at == 0 {
					*at--
				}
			}
		}
	}
	if lightLevel < 0 {
		for row := range viewSize {
			for column := range viewSize {
				viewTiles[row*viewStride+column] = *tileAt(left+column+int(chunkX), top+row+int(chunkY))
			}
		}
	}
}

// Tiles that let sight through only next to the viewer.
var windowTiles = [...]byte{mapTile75, mapTile74, mapTileDoor186, mapTileDoor187, mapTile152}

// letsSightThrough says whether sight passes tile at distance.
// C: ULTIMA_5dfe
func letsSightThrough(tile byte, distance int) bool {
	for _, w := range windowTiles {
		if tile == w {
			return distance == 1
		}
	}
	for _, opaque := range opaqueTiles {
		if tile == opaque {
			return false
		}
	}
	return true
}

// updateVisibility works out which tiles of the loaded map the light sources on it
// light, into the combat map buffer (0 unlit).
// C: ULTIMA_5e4a
func updateVisibility() {
	for i := range combatMap {
		combatMap[i] = hiddenTile
	}
	var found []int
	for y := range settlementSize {
		for x := range settlementSize {
			tile := *tileAt(int(chunkX)+x, int(chunkY)+y)
			for _, source := range lightSources {
				if tile == source {
					found = append(found, y, x)
					combatMap[y*viewStride+x] = tile
					break
				}
			}
		}
	}
	for i := 0; i < len(found); i += 2 {
		top, left := found[i]-viewCentre, found[i+1]-viewCentre
		for row := range viewSize {
			for column := range viewSize {
				viewTiles[row*viewStride+column] = hiddenTile
			}
		}
		searchView(lightSourceRange, top, left, top, left, combatMap[:])
	}
	for i := range combatMap {
		if combatMap[i] == hiddenTile {
			combatMap[i]++
		}
	}
}
