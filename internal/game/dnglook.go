package game

// The dungeon's Look and gem view, the rooms' combat maps and the dungeon's graphics
// (dnglook.c).

// Addresses of the tables in the data segment.
const (
	roomCentreTilesAddress = 9290  // C: D_244a (by cell kind: what stands in a room's middle)
	campPartyXAddress      = 9298  // C: D_2452
	campPartyYAddress      = 9304  // C: D_2458
	partyStartsAAddress    = 9310  // C: D_245e
	partyStartsBAddress    = 9316  // C: D_2464
	partyStartsCAddress    = 9322  // C: D_246a
	partyStartsDAddress    = 9328  // C: D_2470
	foeStartsAAddress      = 9334  // C: D_2476
	foeStartsBAddress      = 9350  // C: D_2486
	foeStartsCAddress      = 9366  // C: D_2496
	foeStartsDAddress      = 9382  // C: D_24a6
	campFoeXAddress        = 9398  // C: D_24b6
	campFoeYAddress        = 9414  // C: D_24c6
	unmarkedRoomsAddress   = 14394 // C: D_383a (rooms that stay unwon)
	unmarkedRoomCount      = 14400 // C: D_3840
	thingRangeAddress      = 14402 // C: D_3842 (a room thing's value: its range...)
	thingBaseAddress       = 14416 // C: D_3850 (...and its least)
	roomFoesAddress        = 14430 // C: D_385e (the foes a room's random foe tiles stand for)
)

const (
	roomStarts     = 6
	roomFoeStarts  = 16
	thingKinds     = 13
	roomFoeChoices = 8
)

var (
	roomCentreTiles []byte // C: D_244a
	campPartyX      []byte // C: D_2452
	campPartyY      []byte // C: D_2458
	partyStartsA    []byte // C: D_245e
	partyStartsB    []byte // C: D_2464
	partyStartsC    []byte // C: D_246a
	partyStartsD    []byte // C: D_2470
	foeStartsA      []byte // C: D_2476
	foeStartsB      []byte // C: D_2486
	foeStartsC      []byte // C: D_2496
	foeStartsD      []byte // C: D_24a6
	campFoeX        []byte // C: D_24b6
	campFoeY        []byte // C: D_24c6
	unmarkedRooms   []byte // C: D_383a
	thingRange      []byte // C: D_3842
	thingBase       []byte // C: D_3850
	roomFoes        []byte // C: D_385e
)

// loadDungeonLookTables reads the tables of the rooms.
func loadDungeonLookTables() {
	roomCentreTiles = segment.Bytes(roomCentreTilesAddress, cellKinds)
	campPartyX = segment.Bytes(campPartyXAddress, roomStarts)
	campPartyY = segment.Bytes(campPartyYAddress, roomStarts)
	partyStartsA = segment.Bytes(partyStartsAAddress, roomStarts)
	partyStartsB = segment.Bytes(partyStartsBAddress, roomStarts)
	partyStartsC = segment.Bytes(partyStartsCAddress, roomStarts)
	partyStartsD = segment.Bytes(partyStartsDAddress, roomStarts)
	foeStartsA = segment.Bytes(foeStartsAAddress, roomFoeStarts)
	foeStartsB = segment.Bytes(foeStartsBAddress, roomFoeStarts)
	foeStartsC = segment.Bytes(foeStartsCAddress, roomFoeStarts)
	foeStartsD = segment.Bytes(foeStartsDAddress, roomFoeStarts)
	campFoeX = segment.Bytes(campFoeXAddress, roomFoeStarts)
	campFoeY = segment.Bytes(campFoeYAddress, roomFoeStarts)
	unmarkedRooms = segment.Bytes(unmarkedRoomsAddress, int(segment.Bytes(unmarkedRoomCount, 1)[0]))
	thingRange = segment.Bytes(thingRangeAddress, thingKinds)
	thingBase = segment.Bytes(thingBaseAddress, thingKinds)
	roomFoes = segment.Bytes(roomFoesAddress, roomFoeChoices)
}

// The combat map's rows of start places (beyond the 11x11 map, from column 11).
const (
	rowPartyEast    = 1
	rowPartyWest    = 2
	rowPartyNorth   = 4
	roomCentre      = 5
	roomEdge        = viewSize - 1
	campCentreTile  = 179
	campCombat      = 4 // the combatKind bit of a camp
	roomFoeTileBase = 64
	roomFoeRandom   = 236 // foe tiles 236..239 stand for one of roomFoes
	roomTileMask    = 252
	roomThingFirst  = 64  // tiles below are things, not foes
	roomFieldTile   = 180 // 180..183, fields
	roomChestTile   = 232 // 232..235
	thingTreasure   = 1
	thingGold       = 2
	thingTableFirst = 3
	thingLast       = 16
	treasureLevel   = 3
	treasureBase    = 7
	goldLevel       = 10
	roomCombat      = 3 // enterDungeonRoom's kind for a room behind a door
	normalTimeTiles = 0
)

// Wall and floor tiles of the rooms by the dungeon's look.
const (
	remainsWallTile  = 79
	remainsFloorTile = 69
	caveWallTile     = 77
	caveFloorTile    = 5
)

// Glyphs of the gem view (the second font unless noted).
const (
	gemSeenTrap     = 24 // first font
	gemLadderUp     = 46
	gemLadderDown   = 45
	gemLadders      = 47
	gemChest        = 112
	gemTrap         = 25 // first font
	gemPit          = 113
	gemSeenPit      = 18 // first font
	gemOtherTrap    = 114
	gemDoor         = 115
	gemPlainWall    = 127 // first font
	gemWall         = 116
	gemRubble       = 117
	gemSecret       = 118
	gemDoors        = 119
	gemParty        = 96
	gemRadius       = 11
	gemSize         = 22
	gemQueueLength  = 256
	gemFlagRows     = 23
	gemTrapCell     = 96
	gemPitCell      = 97
	gemPitSeenCell  = 105
	gemSeenTrapCell = 104
)

// lookInDungeon has a member look at (and maybe drink from) the cell ahead.
// C: DNGLOOK_0000_LookCmdInDungeon
func lookInDungeon() {
	member := chooseMember()
	if member == -1 {
		return
	}
	if torchLight == 0 && lightSpell == 0 {
		printMessage(msgYouSeeDarknessLine)
		return
	}
	if !lookDirection(int(facing)) {
		return
	}
	cell := int(*dungeonCellAt(int(moveX)&(dungeonSize-1), int(moveY)&(dungeonSize-1), int(level)))
	printMessage(msgYouSeeLine)
	if cell == cellPit {
		cell = cellNothing
	}
	switch cell & cellKindMask {
	case cellField:
		switch cell {
		case cellSleepField:
			printMessage(msgASleepFieldLine)
		case cellPoison:
			printMessage(msgAPoisonGasFieldLine)
		case cellFireField:
			printMessage(msgAWallOfFireLine)
		case cellElectric:
			printMessage(msgAnElectricFieldLine)
		default:
			printMessage(msgAnEnergyFieldLine)
		}
		return
	case cellRubble:
		switch int(dungeonLook) & cellLow {
		case lookStalactite:
			printMessage(msgADrippingStalactiteLine)
		case lookCaveIn:
			printMessage(msgACavedInPassageLine)
		default:
			if randomRange(1, 255) == 255 {
				printMessage(msgAnUnfortunateSoftwarePirateLine)
			} else {
				printMessage(msgALessFortunateAdventurerLine)
			}
		}
		return
	}
	switch cell & cellKindMask {
	case cellNothing:
		printMessage(msgAPassageLine)
	case cellLadderUp:
		printMessage(msgAnUpLadderLine)
	case cellLadderDown:
		printMessage(msgADownLadderLine)
	case cellLadders:
		printMessage(msgALadderLine)
	case cellChest:
		printMessage(msgAWoodenChestLine)
	case cellFountain:
		printMessage(msgAFountainLine)
	case cellTrap:
		printMessage(msgAPitLine)
	case cellOpenChest:
		printMessage(msgAnOpenChestLine)
	case cellImpossible:
		printMessage(msgNothingOfNoteLine)
	case cellDoor, cellDoors, cellDoorsTwo:
		printMessage(msgAHeavyDoorLine)
	case cellWall, cellSecret:
		printMessage(msgAWallLine)
	}
	if cell&cellKindMask != cellFountain {
		return
	}
	printMessage(msgWillYouDrinkLine)
	answer := 0
	for answer != 'Y' && answer != 'N' {
		answer = getChar() & 255
	}
	if answer == 'N' {
		printMessage(msgNoLine2)
		return
	}
	printMessage(msgYesGulpLine)
	switch cell {
	case cellFountain:
		printMessage(msgCuredLine)
		party[member].status = statusGood
	case cellFountain + 1:
		printMessage(msgHealedLine)
		party[member].hp = party[member].maxHP
	case cellFountain + 2:
		printMessage(msgPoisonedLine)
		party[member].status = statusPoisoned
	default:
		printMessage(msgBadTasteLine)
		hurtMember(member, randomRange(0, 7))
	}
}

// drawGemField draws a field in the gem view: stripes of the four field colours.
// C: DNGLOOK_0284
func drawGemField(x, y int) {
	for i, colour := range []int{magentaColour, redColour, blueColour, greenColour} {
		setPenColour(colour + brightOffset)
		line(x+1, y+i*2, x+6, y+i*2)
		line(x+1, y+i*2+1, x+6, y+i*2+1)
	}
}

// gemViewCell draws cell (x, y) of the gem view, the party at its middle, once. It returns
// whether the view spreads on through it.
// C: DNGLOOK_0340
func gemViewCell(x, y int) int {
	spreads := 0
	if x < 0 || x >= gemSize || y < 0 || y >= gemSize {
		return 0
	}
	if combatMap[y*viewStride+x] != 0 {
		spreads = 1
		combatMap[y*viewStride+x] = 0
		cx := (x + int(mapX) - gemRadius) & (dungeonSize - 1)
		cy := (y + int(mapY) - gemRadius) & (dungeonSize - 1)
		cell := *dungeonCellAt(cx, cy, int(level))
		setTextPosition(x+1, y+1)
		px, py := x*8+8, y*8+8
		setPenColour(brightWhiteColour)
		switch cell >> 4 {
		case cellNothing >> 4:
			if cell&cellSeen != 0 {
				setTextColour(lightGrayColour)
				printChar(gemSeenTrap)
			}
		case cellLadderUp >> 4:
			selectFont(1)
			setTextColour(lightGrayColour)
			printChar(gemLadderUp)
		case cellLadderDown >> 4:
			selectFont(1)
			setTextColour(lightGrayColour)
			printChar(gemLadderDown)
		case cellLadders >> 4:
			selectFont(1)
			setTextColour(lightGrayColour)
			printChar(gemLadders)
		case cellChest >> 4:
			selectFont(1)
			setTextColour(brightYellow)
			printChar(gemChest)
		case cellFountain >> 4:
			line(px+1, py+4, px+6, py+4)
			line(px+2, py+5, px+5, py+5)
			line(px+1, py+6, px+2, py+6)
			line(px+5, py+6, px+6, py+6)
			setPenColour(blueColour + brightOffset)
			plot(px+1, py+2)
			plot(px+2, py+1)
			line(px+3, py+2, px+4, py+2)
			line(px+3, py+3, px+4, py+3)
			plot(px+5, py+1)
			plot(px+6, py+2)
		case cellTrap >> 4:
			setTextColour(lightGrayColour)
			switch cell {
			case gemTrapCell:
				printChar(gemTrap)
			case gemPitCell, gemPitSeenCell:
				selectFont(1)
				printChar(gemPit)
			case gemSeenTrapCell:
				printChar(gemSeenPit)
			default:
				selectFont(1)
				setTextColour(redColour + brightOffset)
				printChar(gemOtherTrap)
			}
		case cellField >> 4:
			drawGemField(px, py)
		case cellDoor >> 4, cellDoorsTwo >> 4:
			selectFont(1)
			setTextColour(brightYellow)
			printChar(gemDoor)
		case cellWall >> 4:
			if cell == cellWall {
				printChar(gemPlainWall)
			} else {
				selectFont(1)
				printChar(gemWall)
			}
			spreads = 0
		case cellRubble >> 4:
			selectFont(1)
			setTextBackground(blueColour)
			printChar(gemRubble)
			setTextBackground(0)
			spreads = 0
		case cellSecret >> 4:
			selectFont(1)
			switch videoDriver {
			case 0:
				setTextBackground(blueColour)
			case 1, 2:
				setTextBackground(lightGrayColour)
			}
			printChar(gemSecret)
			setTextBackground(0)
			spreads = 0
		case cellDoors >> 4:
			selectFont(1)
			setTextColour(brightYellow)
			printChar(gemDoors)
		}
	}
	setTextColour(brightWhiteColour)
	selectFont(0)
	return spreads
}

// gemNeighbours are the eight steps to a cell's neighbours, in the order the view spreads.
var gemNeighbours = [8][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}}

// viewGemInDungeon shows the map of the level around the party as far as it is open.
// C: DNGLOOK_06a8_ViewCmd
func viewGemInDungeon() {
	setPenColour(0)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	var queueX, queueY [gemQueueLength]byte // C: D_a528, D_a628
	for i := range gemFlagRows * viewStride {
		combatMap[i] = 255
	}
	head, tail := 0, 0
	queueX[tail], queueY[tail] = gemRadius, gemRadius
	tail++
	combatMap[gemRadius*viewStride+gemRadius] = 0
	setPenColour(brightWhiteColour)
	selectWindow(windowFrame)
	setTextPosition(gemRadius+1, gemRadius+1)
	selectFont(1)
	setTextColour(greenColour + brightOffset)
	if videoDriver == cgaComposite {
		setTextColour(brightWhiteColour)
	}
	printChar(gemParty)
	setTextColour(brightWhiteColour)
	selectFont(0)
	for tail != head {
		for _, step := range gemNeighbours {
			x, y := int(queueX[head])+step[0], int(queueY[head])+step[1]
			if gemViewCell(x, y) != 0 {
				queueX[tail], queueY[tail] = byte(x), byte(y)
				tail = (tail + 1) % gemQueueLength
			}
		}
		head = (head + 1) % gemQueueLength
	}
	selectWindow(windowScroll)
	waitForKey(0)
	setPenColour(0)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	redrawDungeon()
}

// roomBit returns the byte and bit of dungeonSeen that mark room won in this dungeon.
func roomBit(room int) (int, byte) {
	n := dungeonNumber()*roomsPerDungeon + room
	return n / 8, 1 << (n & 7)
}

// markRoomWon remembers the room as won, unless it is one that never stays so.
// C: DNGLOOK_0844
func markRoomWon(room int) {
	for _, r := range unmarkedRooms {
		if int(r) == int(mapID&cellLow)*roomsPerDungeon+room {
			return
		}
	}
	at, bit := roomBit(room)
	dungeonSeen[at] |= bit
}

// roomWon tells whether the room has been won.
// C: DNGLOOK_08d4
func roomWon(room int) bool {
	at, bit := roomBit(room)
	return dungeonSeen[at]&bit != 0
}

// revealSeenRooms turns the doors of rooms won into plain doors.
// C: DNGLOOK_093a
func revealSeenRooms() {
	for i, cell := range dungeonMap {
		if cell&cellKindMask == cellDoorsTwo && roomWon(int(cell&cellLow)) {
			dungeonMap[i] &= roomClearedMask
		}
	}
}

// roomLine sets count cells of the combat map from (x, y) on in steps (dx, dy) to tile.
func roomLine(x, y, dx, dy, count int, tile byte) {
	for range count {
		*combatCell(x, y) = tile
		x += dx
		y += dy
	}
}

// wallRoomSide closes a side of the room with blackness; standing in a doorway the walls
// either side of it come in too.
// C: DNGLOOK_097e
func wallRoomSide(kind byte, side int) {
	switch side {
	case faceNorth:
		roomLine(0, 0, 1, 0, viewSize, hiddenTile)
	case faceEast:
		roomLine(roomEdge, 0, 0, 1, viewSize, hiddenTile)
	case faceSouth:
		roomLine(0, roomEdge, 1, 0, viewSize, hiddenTile)
	case faceWest:
		roomLine(0, 0, 0, 1, viewSize, hiddenTile)
	}
	if side == faceNorth && kind == cellDoors {
		*combatCell(5, 2), *combatCell(5, 8) = roomWallTile, roomWallTile
	}
	if side == faceWest && kind == cellDoors {
		*combatCell(2, 5), *combatCell(8, 5) = roomWallTile, roomWallTile
	}
}

// doorRoomSide opens a doorway in a side of the room.
// C: DNGLOOK_0a48
func doorRoomSide(side int) {
	switch side {
	case faceNorth:
		roomLine(3, 1, 1, 0, 5, moongateFloor)
	case faceEast:
		roomLine(9, 3, 0, 1, 5, moongateFloor)
	case faceSouth:
		roomLine(3, 9, 1, 0, 5, moongateFloor)
	case faceWest:
		roomLine(1, 3, 0, 1, 5, moongateFloor)
	}
}

// openRoomSide opens a side of the room to the corridor.
// C: DNGLOOK_0aee
func openRoomSide(side int) {
	switch side {
	case faceNorth:
		roomLine(2, 1, 1, 0, 7, moongateFloor)
	case faceEast:
		roomLine(9, 2, 0, 1, 7, moongateFloor)
	case faceSouth:
		roomLine(2, 9, 1, 0, 7, moongateFloor)
	case faceWest:
		roomLine(1, 2, 0, 1, 7, moongateFloor)
	}
}

// buildRoomSide shapes a side of the corridor fight's map by the cell beyond it.
// C: DNGLOOK_0b9e
func buildRoomSide(side int) {
	x, y := int(mapX), int(mapY)
	switch side {
	case faceNorth:
		y--
	case faceEast:
		x++
	case faceSouth:
		y++
	case faceWest:
		x--
	}
	kind := *dungeonCellAt(x&(dungeonSize-1), y&(dungeonSize-1), int(level)) & cellKindMask
	if kind < cellDoor {
		openRoomSide(side)
	}
	if kind == cellWall || kind == cellRubble || kind == cellSecret {
		wallRoomSide(*dungeonCellAt(int(mapX), int(mapY), int(level))&cellKindMask, side)
	} else {
		doorRoomSide(side)
	}
}

// buildRoomWalls builds the walls of a corridor fight's map around where the party stands.
// C: DNGLOOK_0c6c
func buildRoomWalls() {
	roomLine(0, 1, 1, 0, viewSize, roomWallTile)
	roomLine(0, 9, 1, 0, viewSize, roomWallTile)
	for i := range viewSize {
		*combatCell(1, i), *combatCell(9, i) = roomWallTile, roomWallTile
	}
	*combatCell(0, 0), *combatCell(roomEdge, 0) = hiddenTile, hiddenTile
	*combatCell(0, roomEdge), *combatCell(roomEdge, roomEdge) = hiddenTile, hiddenTile
	kind := *dungeonCellAt(int(mapX), int(mapY), int(level)) & cellKindMask
	if kind != cellNothing && kind < cellField {
		kind >>= 4
		if tile := roomCentreTiles[kind]; tile != 0 {
			*combatCell(roomCentre, roomCentre) = tile
		}
		if kind == cellLadders>>4 {
			moongateFoe = 1
			changeTimeTiles(normalTimeTiles)
		} else {
			moongateFoe = 0
		}
	}
	for side := range 4 {
		buildRoomSide(side)
	}
}

// drawDungeonFrame lays out the map of a fight in a dungeon corridor (or a camp there):
// the floor, the walls, and where the party and the foes start; it picks the foes.
// C: DNGLOOK_0d3e
func drawDungeonFrame() {
	for row := range viewSize {
		roomLine(0, row, 1, 0, viewSize, moongateFloor)
	}
	camp := combatKind&campCombat != 0
	if camp {
		*combatCell(roomCentre, roomCentre) = campCentreTile
	}
	buildRoomWalls()
	for i := range triggerCount {
		*combatCell(specialColumn+i, rowTriggers) = hiddenTile
		*combatCell(triggerYColumn+i, rowTriggers) = hiddenTile
	}
	for i := range roomStarts {
		if camp {
			for _, row := range []int{rowPartySouth, rowPartyWest, rowPartyNorth, rowPartyEast} {
				*combatCell(specialColumn+i, row) = campPartyX[i]
				*combatCell(partyYColumn+i, row) = campPartyY[i]
			}
			continue
		}
		*combatCell(specialColumn+i, rowPartySouth) = partyStartsD[i]
		*combatCell(specialColumn+i, rowPartyWest) = partyStartsC[i]
		*combatCell(specialColumn+i, rowPartyNorth) = partyStartsA[i]
		*combatCell(specialColumn+i, rowPartyEast) = partyStartsB[i]
		*combatCell(partyYColumn+i, rowPartySouth) = partyStartsB[i]
		*combatCell(partyYColumn+i, rowPartyWest) = partyStartsA[i]
		*combatCell(partyYColumn+i, rowPartyNorth) = partyStartsC[i]
		*combatCell(partyYColumn+i, rowPartyEast) = partyStartsD[i]
	}
	for i := range roomFoeStarts {
		x, y := combatCell(specialColumn+i, rowMonsterX), combatCell(specialColumn+i, rowMonsterY)
		switch {
		case camp:
			*y, *x = campFoeY[i], campFoeX[i]
		case facing == faceNorth:
			*x, *y = foeStartsA[i], foeStartsC[i]
		case facing == faceEast:
			*x, *y = foeStartsB[i], foeStartsD[i]
		case facing == faceSouth:
			*x, *y = foeStartsD[i], foeStartsB[i]
		case facing == faceWest:
			*x, *y = foeStartsC[i], foeStartsA[i]
		}
	}
	var order [roomFoeStarts]int
	for i := range order {
		order[i] = i
	}
	for i := range roomFoeStarts {
		j := randomRange(0, roomFoeStarts-1)
		order[i], order[j] = order[j], order[i]
		*combatCell(specialColumn+i, rowMonsterTiles) = 0
	}
	var foe int
	if camp {
		foe = int(wandererFoes[randomRange(0, wandererKinds-1)])
	} else {
		foe = int(actors[wandererSlot].hull)
	}
	most := int(foes[foe].maxCount)
	count := randomRange(1, most)
	if most == 8 || most == roomFoeStarts {
		count = most
	}
	for i := range count {
		*combatCell(specialColumn+order[i], rowMonsterTiles) = byte(foe*4 + roomFoeTileBase)
	}
}

// leaveRoom moves the party on the way it left the fight: a step, or a level up or down.
// C: DNGLOOK_0fda
func leaveRoom() {
	switch exitDirection {
	case dirUp:
		mapY--
		facing, dungeonMode = faceNorth, faceNorth
		if mapY > dungeonSize-1 {
			mapY = dungeonSize - 1
		}
	case dirRight:
		mapX++
		facing, dungeonMode = faceEast, faceEast
		if mapX > dungeonSize-1 {
			mapX = 0
		}
	case dirDown:
		mapY++
		facing, dungeonMode = faceSouth, faceSouth
		if mapY > dungeonSize-1 {
			mapY = 0
		}
	case dirLeft:
		mapX--
		facing, dungeonMode = faceWest, faceWest
		if mapX > dungeonSize-1 {
			mapX = dungeonSize - 1
		}
	case exitUp:
		if level == 0 {
			leaveDungeon()
			mapID = 0
		} else {
			level--
			dungeonMode = dungeonModeDown
		}
	case exitDown:
		if level == lastLevel {
			leaveDungeon()
			mapID = 0
		} else {
			level++
			dungeonMode = dungeonModeUp
		}
	}
}

// restoreDungeonLook loads the dungeon's pictures (again) and blanks the view; fresh
// clears the wanderer.
// C: DNGLOOK_109e
func restoreDungeonLook(fresh int) {
	if mapID == 0 {
		return
	}
	// the original frees the tile set here (ULTIMA_102e_UnloadTileset) to make room; the
	// port keeps it loaded
	if dungeonGraphics&dungeonPicturesLoaded == 0 {
		dungeonPictures = loadPictures(resourceFiles[itemPicturesFile+int(dungeonLook)])
		itemPictures = loadPictures(resourceFiles[itemPicturesFile])
	}
	dungeonGraphics = dungeonPicturesLoaded
	if fresh != 0 {
		foe := &actors[wandererSlot]
		foe.tile, foe.animTile = 0, 0
		foe.hull, foe.x, foe.y = noWanderer, noWanderer, noWanderer
	}
	setPenColour(0)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	selectPage(1)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	selectPage(0)
}

// hideDungeonView drops the dungeon's pictures and brings back the tiles, before a fight
// or on leaving.
// C: DNGLOOK_1130
func hideDungeonView() {
	if dungeonGraphics&dungeonPicturesLoaded != 0 {
		monsterPictures = nil
		itemPictures = nil
		dungeonPictures = nil
	}
	if dungeonGraphics&worldTilesLoaded == 0 {
		for !loadTiles(tileFile) {
		}
	}
	if dungeonGraphics&timeTilesNormal == 0 {
		changeTileSet(normalTimeTiles)
	}
	dungeonGraphics = leftDungeonGraphics
}

// enterDungeonRoom places the party and the room's foes and things on the room's map. kind
// 1 keeps the party where it is; 0 and 2 place no things; 3 places a room's things only
// behind a room door.
// C: DNGLOOK_117e
func enterDungeonRoom(cell, kind int) {
	if kind != 1 {
		row := rowPartyNorth
		switch facing {
		case faceNorth, dungeonModeUp:
			row = rowPartySouth
		case faceEast:
			row = rowPartyWest
		case faceWest:
			row = rowPartyEast
		}
		setUpFighters()
		for i := range int(partySize) {
			x, y := *combatCell(specialColumn+i, row), *combatCell(partyYColumn+i, row)
			combatants[i].x, actors[i].x = x, x
			combatants[i].y, actors[i].y = y, y
		}
	}
	if kind <= 0 || !(cell >= cellDoorsTwo && kind == roomCombat || kind < roomCombat) {
		return
	}
	var picks [4]byte
	for i := range picks {
		picks[i] = roomFoes[randomRange(0, roomFoeChoices-1)]
	}
	for i := range roomFoeStarts {
		tile := *combatCell(specialColumn+i, rowMonsterTiles)
		if tile == 0 {
			continue
		}
		var place, entity int
		if tile < roomThingFirst || tile&roomTileMask == roomFieldTile || tile&roomTileMask == roomChestTile {
			place, entity = placeBareActor, int(tile)
		} else {
			place, entity = placeFoe, int(tile-roomFoeTileBase)/4
		}
		if tile&roomTileMask == roomFoeRandom {
			entity = int(picks[tile&3])
		}
		slot := placeFighter(entity, place,
			int(*combatCell(specialColumn+i, rowMonsterX)), int(*combatCell(specialColumn+i, rowMonsterY)), int(level))
		// fix: the original writes before the actors when there was no room for the thing
		if place != placeBareActor || slot < 0 {
			continue
		}
		switch {
		case entity == thingTreasure:
			actors[slot].hull = level*treasureLevel + treasureBase
		case entity == thingGold:
			actors[slot].hull = byte(randomRange(1, int(level)*goldLevel+goldLevel))
		case entity < thingLast:
			actors[slot].hull = thingBase[entity-thingTableFirst] + byte(randomRange(0, int(thingRange[entity-thingTableFirst])-1))
		}
	}
}
