package game

import "github.com/memmaker/u5-go/internal/ega"

// The dungeons (dungeon.c): the first-person view, walking, the traps, the wandering foe,
// the rooms behind the doors, and Klimb and Attack below ground.

// Addresses of the dungeon tables in the data segment.
const (
	wandererFoesAddress   = 5948  // C: D_173c (the foe type of each wanderer)
	wandererStatesAddress = 5956  // C: D_1744 (its starting pose)
	roomFacingsAddress    = 11382 // C: D_2c76 (the way the party faces on entering a room)
	signFirstAddress      = 11752 // C: D_2de8 (each dungeon's first sign)
	signCountAddress      = 11760 // C: D_2df0 (its number of signs)
	signColumnAddress     = 11768 // C: D_2df8
	signRowAddress        = 11780 // C: D_2e04
	signTextsAddress      = 11792 // C: D_2e10 (pointers)
	wandererStepsAddress  = 11814 // C: D_2e26 (the walking poses)
	wandererXAddress      = 11818 // C: D_2e2a (by depth)
	wandererYAddress      = 11826 // C: D_2e32 (by mirroring and depth)
	fieldTopAddress       = 11842 // C: D_2e42 (by depth)
	fieldBottomAddress    = 11850 // C: D_2e4a
	fieldLinesAddress     = 11858 // C: D_2e52
	fieldWidthAddress     = 11866 // C: D_2e5a
	pieceXAddress         = 11874 // C: D_2e62 (by side and column)
	itemXAddress          = 11890 // C: D_2e72
	itemYAddress          = 11898 // C: D_2e7a (by row and column)
	frontPiecesAddress    = 11914 // C: D_2e8a (by wall kind)
	dripFrontYAddress     = 11920 // C: D_2e90
	dripSideYAddress      = 11930 // C: D_2e9a
	sprayNearAddress      = 11940 // C: D_2ea4 (fountain spray points, by frame)
	spraySecondAddress    = 11988 // C: D_2ed4
	sprayThirdAddress     = 12018 // C: D_2ef2
	sprayFarAddress       = 12042 // C: D_2f0a
	sprayTopAddress       = 12048 // C: D_2f10
	upperPiecesAddress    = 12054 // C: D_2f16 (by cell kind)
	lowerPiecesAddress    = 12062 // C: D_2f1e
	sideXAddress          = 12072 // C: D_2f28 (by facing)
	sideYAddress          = 12076 // C: D_2f2c
	dungeonVersionAddress = 11582 // the version string Ctrl-V prints
)

const (
	wandererKinds = 8
	dungeonSigns  = 11
	depths        = 4
	// drip tables are read past their ten bytes into the next table, as the original does
	dripTableReach = 13
	sprayNearDots  = 8
	sprayDots2     = 5
	sprayDots3     = 4
	sprayFrames    = 3
	cellKinds      = 8
)

var (
	wandererFoes   []byte   // C: D_173c
	wandererStates []byte   // C: D_1744
	roomFacings    []byte   // C: D_2c76
	signFirst      []byte   // C: D_2de8
	signCount      []byte   // C: D_2df0
	signColumn     []byte   // C: D_2df8
	signRow        []byte   // C: D_2e04
	signTexts      []string // C: D_2e10
	wandererSteps  []byte   // C: D_2e26
	wandererX      []int    // C: D_2e2a
	wandererY      [2][]int // C: D_2e32
	fieldTop       []int    // C: D_2e42
	fieldBottom    []int    // C: D_2e4a
	fieldLines     []int    // C: D_2e52
	fieldWidth     []int    // C: D_2e5a
	pieceX         [2][]int // C: D_2e62
	itemX          []int    // C: D_2e72
	itemY          [2][]int // C: D_2e7a
	frontPieces    []byte   // C: D_2e8a
	dripFrontY     []byte   // C: D_2e90
	dripSideY      []byte   // C: D_2e9a
	sprayNear      []byte   // C: D_2ea4 (x, y pairs)
	spraySecond    []byte   // C: D_2ed4
	sprayThird     []byte   // C: D_2ef2
	sprayFar       []byte   // C: D_2f0a
	sprayTop       []byte   // C: D_2f10
	upperPieces    []byte   // C: D_2f16
	lowerPieces    []byte   // C: D_2f1e
	sideX          []int8   // C: D_2f28
	sideY          []int8   // C: D_2f2c
)

// loadDungeonTables reads the dungeon tables.
func loadDungeonTables() {
	wandererFoes = segment.Bytes(wandererFoesAddress, wandererKinds)
	wandererStates = segment.Bytes(wandererStatesAddress, wandererKinds)
	roomFacings = segment.Bytes(roomFacingsAddress, dungeonModeUp+1)
	signFirst = segment.Bytes(signFirstAddress, dungeonCount)
	signCount = segment.Bytes(signCountAddress, dungeonCount)
	signColumn = segment.Bytes(signColumnAddress, dungeonSigns+1)
	signRow = segment.Bytes(signRowAddress, dungeonSigns+1)
	signTexts = messages(signTextsAddress, dungeonSigns)
	wandererSteps = segment.Bytes(wandererStepsAddress, depths)
	wandererX = segment.SignedWords(wandererXAddress, depths)
	fieldTop = segment.SignedWords(fieldTopAddress, depths)
	fieldBottom = segment.SignedWords(fieldBottomAddress, depths)
	fieldLines = segment.SignedWords(fieldLinesAddress, depths)
	fieldWidth = segment.SignedWords(fieldWidthAddress, depths)
	itemX = segment.SignedWords(itemXAddress, depths)
	for side := range 2 {
		wandererY[side] = segment.SignedWords(wandererYAddress+side*depths*2, depths)
		pieceX[side] = segment.SignedWords(pieceXAddress+side*depths*2, depths)
		itemY[side] = segment.SignedWords(itemYAddress+side*depths*2, depths)
	}
	frontPieces = segment.Bytes(frontPiecesAddress, cellKinds-2)
	dripFrontY = segment.Bytes(dripFrontYAddress, dripTableReach)
	dripSideY = segment.Bytes(dripSideYAddress, dripTableReach)
	sprayNear = segment.Bytes(sprayNearAddress, sprayFrames*sprayNearDots*2)
	spraySecond = segment.Bytes(spraySecondAddress, sprayFrames*sprayDots2*2)
	sprayThird = segment.Bytes(sprayThirdAddress, sprayFrames*sprayDots3*2)
	sprayFar = segment.Bytes(sprayFarAddress, sprayFrames*2)
	sprayTop = segment.Bytes(sprayTopAddress, sprayFrames*2)
	upperPieces = segment.Bytes(upperPiecesAddress, cellKinds)
	lowerPieces = segment.Bytes(lowerPiecesAddress, cellKinds)
	sideX, sideY = make([]int8, 4), make([]int8, 4)
	for i, b := range segment.Bytes(sideXAddress, 4) {
		sideX[i] = int8(b)
	}
	for i, b := range segment.Bytes(sideYAddress, 4) {
		sideY[i] = int8(b)
	}
}

// The dungeon's state beyond the saved game.
var (
	// footstepOdd alternates each step; a footstep sounds on every other (D_24e7).
	footstepOdd bool // C: D_24e7
	// footstepLength is the length of the next footstep's noise (D_a9fb).
	footstepLength byte // C: D_a9fb
	// fountainFrame is the fountain spray's frame (D_2f26).
	fountainFrame int // C: D_2f26
	// roomWallTile is the tile a dungeon room's walls are (D_bb14).
	roomWallTile byte // C: D_bb14
	// dungeonGraphics tells what of the dungeon's graphics are loaded (D_bb17).
	dungeonGraphics byte // C: D_bb17
	// the loaded picture files: the walls, the items and the wanderer
	dungeonPictures ega.Pictures // C: D_a9c0
	itemPictures    ega.Pictures // C: D_a9c4
	monsterPictures ega.Pictures // C: D_a9c6
)

// dungeonGraphics bits.
const (
	dungeonPicturesLoaded = 1
	worldTilesLoaded      = 2
	timeTilesNormal       = 4
	leftDungeonGraphics   = worldTilesLoaded | timeTilesNormal
)

// Indexes into resourceFiles.
const (
	itemPicturesFile    = 4
	monsterPicturesFile = 8
)

const (
	roomsFile           = "DUNGEON.CBT"
	roomsPerDungeon     = 16
	roomCombatKind      = 130 // a room's combat
	roomClearedMask     = 175 // clears a room door's "room not won" bits
	noWanderer          = 255
	wandererSlot        = 1
	wandererTrailSlot   = 2 // where the wanderer was the step before
	wandererTries       = 8
	wandererMirrorOdds  = 48 // out of 100: a bat or a gremlin faces the other way
	wandererBat         = 22
	wandererGremlin     = 24
	wandererMirrored    = 255
	wandererTimeStopped = 27 // a foe that does not walk
	wandererHoldStill   = 900
	wandererFar         = 999
	wandererCatchOdds   = 1  // of 0..7: a wanderer walks back onto its trail anyway
	trapOdds            = 30 // a trap catches one whose dexterity is below 1..30
	trapNoiseRate       = 50
	trapNoiseLimit      = 3500
	wakeOdds            = 4 // of 0..63: a sleeper wakes
	stepNoiseLimit      = 20000
	stepNoiseFade       = 4
	stepNoiseLength     = 15
	fieldNoiseLength    = 500
	lowestLevel         = 8 // a pit below level 8 drops to the surface
	lookRemains         = 3 // the dungeons with skeletons in their walls
	ctrlKeyShift        = 250
	exitToDOSAnswer     = 'Y'
	clearTrapMask       = 248
	dripState           = 7
	dripSound           = 5
	dripShine           = 4
	dripOdds            = 4 // of 0..64: a drip starts
	dripSoundFrom       = 3200
	dripSoundTo         = 3500
	dripSoundDepth      = 20
	cyanColour          = 3
	brightCyanColour    = 11
	eyesOdds            = 4
	klimbFailFrom       = 800
	klimbFailTo         = 2000
	klimbFailTime       = 50
	firstTimeStopKey    = '0'
	lastTimeStopKey     = '9'
)

// The dungeon view on the screen.
const (
	viewTop, viewBottom                      = 8, 183
	pageLeft, pageTop, pageRight, pageBottom = 96, 14, 175, 178
	clipLeft, clipTop, clipRight, clipBottom = 40, 42, 152, 183
	fullCopyLeft                             = 16
	halfCopyLeft, halfCopyRight              = 40, 150
	centreX                                  = 96
	mirrorX                                  = 190
	pieceTop                                 = 14
	itemFloor                                = 96
	wallPieces                               = 31
	sidePieceMask                            = 248
	sidePieceRow                             = 8
	cornerPieceMask                          = 252
	cornerPieceRow                           = 24
	doorPiece                                = 12
	sidePieceDoor                            = 4
	sidePieceOpen                            = 16
	sidePieceRubble                          = 20
	nearDripX, farDripX                      = 33, 67
	frontDripX                               = 95
	eyeTop                                   = 87
)

// dungeonNumber is the index of the dungeon the party is in among the dungeons' rooms and
// seen-room bits (the Underworld has none).
func dungeonNumber() int {
	n := int(mapID) - firstDungeonMap
	if n >= 1 {
		n--
	}
	return n
}

// enterRoom fights the room behind the door the party stands in.
// C: DUNGEON_0000
func enterRoom(cell byte) {
	printMessage(msgEnteringRoomLine)
	facing = roomFacings[dungeonMode]
	clearKeyboard()
	room := int(cell & cellLow)
	active := activeMember
	activeMember, combatant = noMember, noCombatant
	clear(combatMap[:combatMapBytes])
	readFile(roomsFile, combatMap[:combatMapBytes], (dungeonNumber()*roomsPerDungeon+room)*combatMapBytes)
	savedMapID = mapID
	x, y := mapX, mapY
	wanderer, trail := actors[wandererSlot], actors[wandererTrailSlot]
	mapID = combatMapID
	exitDirection = 0
	enterDungeonRoom(int(cell), roomCombat)
	updateFrame()
	combatKind = roomCombatKind
	if !combatMainLoop() {
		if exitDirection == exitMirror {
			endgame()
		}
		mapID = savedMapID
		markRoomWon(room)
		*dungeonCellAt(int(x), int(y), int(level)) &= roomClearedMask
	}
	mapY, mapX = y, x
	mapID = savedMapID
	leaveRoom()
	actors[wandererSlot], actors[wandererTrailSlot] = wanderer, trail
	activeMember = active
}

// resumeDungeon sets a new wanderer on the level when fresh, and loads its pictures.
// C: DUNGEON_0134
func resumeDungeon(fresh int) {
	foe := &actors[wandererSlot]
	before := foe.hull
	if fresh != 0 {
		kind := randomRange(0, wandererKinds-1)
		foe.tile, foe.animTile = byte(kind), byte(kind)
		foe.target = 0
		foe.state = wandererStates[kind]
		foe.hull = wandererFoes[kind]
		foe.z = level
		if before != noWanderer {
			monsterPictures = nil
		}
		monsterPictures = nil
		if !placeWanderer() {
			foe.tile, foe.animTile = 0, 0
			foe.hull = noWanderer
		}
	}
	if foe.hull != noWanderer {
		monsterPictures = loadPictures(resourceFiles[int(foe.tile)+monsterPicturesFile])
	}
}

// showLevelAndFacing prints the level and the way the party faces in the frame.
// C: DUNGEON_01d2_PrintWalkDir
func showLevelAndFacing() {
	saved := windowIndex
	selectWindow(windowFrame)
	setTextPosition(12, 0)
	printNumber(int(level)+1, 1, ' ')
	setTextPosition(12, 23)
	switch facing {
	case faceNorth:
		printMessage(msgNorth)
	case faceEast:
		printMessage(msgEast)
	case faceSouth:
		printMessage(msgSouth)
	case faceWest:
		printMessage(msgWest)
	default:
		printMessage(msgSpaceChar63Char63Char63Char63)
	}
	selectWindow(saved)
}

// placeWanderer puts the wanderer somewhere open on the level, away from the party's row
// and column. It returns false when no place was found.
// C: DUNGEON_0252
func placeWanderer() bool {
	foe, trail := &actors[wandererSlot], &actors[wandererTrailSlot]
	for range wandererTries {
		at := randomRange(0, dungeonSize*dungeonSize-1)
		kind := dungeonMap[int(level)*dungeonSize*dungeonSize+at] & cellKindMask
		if kind < cellTrap || kind == cellOpenChest {
			y, x := at/dungeonSize, at%dungeonSize
			if x != int(mapX) && y != int(mapY) {
				foe.x, trail.x = byte(x), byte(x)
				foe.y, trail.y = byte(y), byte(y)
				if (foe.hull == wandererBat || foe.hull == wandererGremlin) && randomRange(0, 99) > wandererMirrorOdds {
					foe.target = wandererMirrored
				}
				return true
			}
		}
	}
	foe.x, trail.x, foe.y, trail.y = noWanderer, noWanderer, noWanderer, noWanderer
	return false
}

// drawDungeonCaptions draws the frame's level and direction captions.
// C: DUNGEON_0332
func drawDungeonCaptions() {
	saved := windowIndex
	setPenColour(blueColour)
	fillRect(40, 0, 152, 7)
	fillRect(48, 185, 152, 191)
	setPenColour(brightWhiteColour)
	line(40, 7, 152, 7)
	line(48, 184, 152, 184)
	selectWindow(windowFrame)
	setTextPosition(10, 0)
	printOpenCaption()
	printMessage(msgLSp)
	printCloseCaption()
	setTextPosition(6, 23)
	printOpenCaption()
	printMessage(msgDirSp)
	printCloseCaption()
	selectWindow(saved)
}

// readDungeonKey waits for a key, redrawing the view (its drips and sprays) meanwhile.
// C: DUNGEON_03d6
func readDungeonKey() int {
	defer asking(promptCommand)()
	if statusChanged {
		showStatus()
		statusChanged = false
	}
	if !reusePrompt() {
		printChar('\n')
		printOpenCaption()
	}
	full := 1
	for {
		key := toUpper(pollKeyWithCursor())
		// Ctrl-A..Ctrl-D typed are not the arrows they share codes with
		if key != 0 && key < keyCtrlE && !navigationKey {
			key += ctrlKeyShift
		}
		if key != 0 {
			return key
		}
		drawDungeon(full)
		if full != 0 {
			copyArea(1, 0, fullCopyLeft, pageTop, pageRight, pageBottom)
		} else {
			copyArea(1, 0, halfCopyLeft, pageTop, halfCopyRight, pageBottom)
		}
		if dungeonLook != lookStalactite {
			full = 0
		}
		dungeonStep()
	}
}

// electricField shocks the party walking into an electric field and throws it back.
// C: DUNGEON_0470_ElectricField
func electricField(way int) {
	redrawDungeon()
	printMessage(msgOuchLine2)
	printMessage(msgElectricFieldLine)
	setPenColour(brightWhiteColour)
	xorFillRect(viewLeft, viewTop, viewRight, viewBottom)
	waitTicks(1)
	xorFillRect(viewLeft, viewTop, viewRight, viewBottom)
	whiteNoise(1, fieldNoiseLength, stepNoiseLimit)
	if way > 0 {
		mapX -= byte(demoDeltaX[facing])
		mapY -= byte(demoDeltaY[facing])
	} else {
		mapX += byte(demoDeltaX[facing])
		mapY += byte(demoDeltaY[facing])
	}
	hurtParty()
	redrawDungeon()
}

// stepInDungeon advances, backs up or turns the party. It returns whether it tried to move.
// C: DUNGEON_0502
func stepInDungeon(key int, cell byte) int {
	way, moved := 0, 0
	switch key {
	case keyUp:
		echoStep(msgAdvanceLine)
		way = 1
	case keyRight:
		if cell&cellKindMask == cellDoors {
			printMessage(msgNotInDoorwayLine)
		} else {
			echoStep(msgTurnRightLine)
			facing = (facing + 1) & 3
		}
	case keyDown:
		echoStep(msgBackUpLine)
		way = -1
	case keyLeft:
		if cell&cellKindMask == cellDoors {
			printMessage(msgNotInDoorwayLine)
		} else {
			echoStep(msgTurnLeftLine)
			facing = (facing + 3) & 3
		}
	default:
		echoStep(msgTurnAroundLine)
		facing = (facing + 2) & 3
	}
	dungeonMode = facing
	if way != 0 {
		moved = 1
		x := demoDeltaX[facing]*way + int(mapX)
		if x < 0 {
			x = dungeonSize - 1
		}
		if x > dungeonSize-1 {
			x = 0
		}
		y := demoDeltaY[facing]*way + int(mapY)
		if y < 0 {
			y = dungeonSize - 1
		}
		if y > dungeonSize-1 {
			y = 0
		}
		next := *dungeonCellAt(x, y, int(level))
		if next == cellElectric {
			mapX, mapY = byte(x), byte(y)
			electricField(way)
			x, y = int(mapX), int(mapY)
		}
		kind := next & cellKindMask
		switch {
		case kind > cellDoor && kind < cellDoors || way == -1 && (kind == cellDoor || kind == cellDoorsTwo):
			printMessage(msgBlockedLine)
		case x == int(actors[wandererSlot].x) && y == int(actors[wandererSlot].y):
			printMessage(msgBlockedLine)
		default:
			footstepLength = stepNoiseLength
			footstepOdd = false
			dungeonStep()
			footstepOdd = false
			mapX = byte(x) & (dungeonSize - 1)
			mapY = byte(y) & (dungeonSize - 1)
		}
	}
	redrawDungeon()
	return moved
}

// dungeonCommand carries out a key's command below ground. It returns whether a turn
// passed.
// C: DUNGEON_06c4_ProcessCommand
func dungeonCommand(key int) int {
	turned := 1
	switch key {
	case keyCtrlK:
		printNumber(int(karma), 1, ' ')
		printChar('\n')
	case keyCtrlE:
		printMessage(msgExitToDosSp)
		if getChar() == exitToDOSAnswer {
			exitGame()
		}
		printMessage(msgNLine)
	case keyCtrlV:
		printMessage(dungeonVersionAddress)
		printChar('\n')
	case keyLeft, keyRight, keyUp, keyDown, keyEnter, '.':
		turned = stepInDungeon(key, *dungeonCellAt(int(mapX), int(mapY), int(level)))
	case keyCtrlS:
		printMessage(msgSoundSp)
		if soundOn {
			printMessage(msgOffLine)
		} else {
			printMessage(msgOnLine)
		}
		soundOn = !soundOn
	default:
		if key >= firstTimeStopKey && key <= lastTimeStopKey {
			setActiveMember(key)
			turned = 0
		} else {
			turned = doCommand(key)
		}
	}
	return turned
}

// moveWanderer moves the wanderer a step at random. It returns whether it reached the
// party (and so attacks).
// C: DUNGEON_07e2
func moveWanderer() bool {
	foe, trail := &actors[wandererSlot], &actors[wandererTrailSlot]
	attacks := false
	if foe.hull == noWanderer {
		return false
	}
	if foe.hull != wandererTimeStopped {
		plan := wandererFar
		x, y := 0, 0
		for range wandererTries {
			way := randomRange(0, 3)
			x = demoDeltaX[way] + int(foe.x)
			if x > dungeonSize-1 {
				x = 0
			}
			if x < 0 {
				x = dungeonSize - 1
			}
			y = demoDeltaY[way] + int(foe.y)
			if y > dungeonSize-1 {
				y = 0
			}
			if y < 0 {
				y = dungeonSize - 1
			}
			kind := *dungeonCellAt(x, y, int(level)) & cellKindMask
			if kind != cellTrap && kind != cellField && kind < cellDoor {
				if x != int(trail.x) || y != int(trail.y) || randomRange(0, 7) == wandererCatchOdds {
					plan = wandererHoldStill
					break
				}
			}
		}
		if plan == wandererHoldStill {
			trail.x, trail.y = foe.x, foe.y
			foe.x, foe.y = byte(x), byte(y)
		}
	}
	if foe.x == mapX && foe.y == mapY {
		attacks = true
		foe.x, foe.y = trail.x, trail.y
	}
	return attacks
}

// sleepTrap puts to sleep those of the party too slow to escape it.
// C: DUNGEON_0948_SleepSpell
func sleepTrap() {
	printMessage(msgSleepSpellLine)
	for row := range int(partySize) {
		if int(party[row].dex) <= randomRange(1, trapOdds) && party[row].status != statusDead {
			party[row].status = statusSleeping
			highlightMember(row)
			whiteNoise(1, trapNoiseRate, trapNoiseLimit)
			statusChanged = true
		}
	}
	*dungeonCellAt(int(mapX), int(mapY), int(level)) &= cellSeen
}

// poisonTrap poisons those of the party too slow to escape it.
// C: DUNGEON_09e6_Poison
func poisonTrap() {
	printMessage(msgPoisonLine)
	for row := range int(partySize) {
		if int(party[row].dex) <= randomRange(1, trapOdds) && party[row].status != statusDead {
			party[row].status = statusPoisoned
			whiteNoise(1, trapNoiseRate, trapNoiseLimit)
		}
	}
}

// fallThroughPit drops the party down pit after pit; below the last level it lands outside.
// C: DUNGEON_0a4c
func fallThroughPit(cell byte) {
	dungeonMode = dungeonModeUp
	for (cell == cellPit || cell == cellPit|cellSeen) && level < lowestLevel {
		printMessage(msgPitTrapLine)
		printMessage(msgFallingLine)
		*dungeonCellAt(int(mapX), int(mapY), int(level)) &= clearTrapMask
		level++
		cell = *dungeonCellAt(int(mapX), int(mapY), int(level))
		if cell < cellImpossible {
			cell |= cellSeen
			*dungeonCellAt(int(mapX), int(mapY), int(level)) = cell
		}
		redrawDungeon()
		printMessage(msgSplatLine)
		hurtParty()
	}
	if level == lowestLevel {
		mapID = 0
	}
	if mapID != 0 {
		kind := *dungeonCellAt(int(mapX), int(mapY), int(level)) & cellKindMask
		if kind == cellDoor || kind == cellDoorsTwo {
			hideDungeonView()
			enterRoom(*dungeonCellAt(int(mapX), int(mapY), int(level)))
			if mapID == 0 {
				return
			}
			restoreDungeonLook(1)
		}
		resumeDungeon(1)
	}
}

// wandererAttacks turns the party toward the wanderer and fights it.
// C: DUNGEON_0b7e
func wandererAttacks() {
	printMessage(msgAttacked)
	trail := &actors[wandererTrailSlot]
	switch {
	case (trail.x-1)&(dungeonSize-1) == mapX:
		dungeonMode = faceEast
	case (trail.x+1)&(dungeonSize-1) == mapX:
		dungeonMode = faceWest
	case (trail.y-1)&(dungeonSize-1) == mapY:
		dungeonMode = faceSouth
	default:
		dungeonMode = faceNorth
	}
	if dungeonMode != facing {
		printMessage(msgFromTheSp)
		switch dungeonMode {
		case faceNorth:
			printMessage(msgNorth2)
		case faceEast:
			printMessage(msgEast2)
		case faceSouth:
			printMessage(msgSouth2)
		case faceWest:
			printMessage(msgWest2)
		}
		facing = dungeonMode
	}
	printMessage(msgChar33Newline)
	redrawDungeon()
	clearKeyboard()
	waitTicks(2)
	hideDungeonView()
	exitDirection = 0
	combatKind = specialDungeonCombat
	drawDungeonFrame()
	enterSpecialMap(specialDungeonCombat, int(actors[wandererSlot].hull), 0)
	leaveRoom()
	restoreDungeonLook(1)
	resumeDungeon(1)
	if mapID != 0 {
		redrawDungeon()
	}
}

// afterDungeonTurn wakes sleepers, moves the wanderer when moved, and springs what the
// party stands on: a room, a field or a trap.
// C: DUNGEON_0c76
func afterDungeonTurn(cell byte, moved int) {
	attacked := false
	kind := cell & cellKindMask
	woke := 0
	for row := range int(partySize) {
		if party[row].status == statusSleeping && randomRange(0, 63) < wakeOdds {
			party[row].status = statusGood
			woke++
		}
	}
	if woke != 0 {
		showStatus()
	}
	if moved != 0 {
		attacked = moveWanderer()
	}
	if attacked {
		wandererAttacks()
		cell = *dungeonCellAt(int(mapX), int(mapY), int(level))
		kind = cell & cellKindMask
	}
	if kind == cellDoorsTwo || kind == cellDoor {
		hideDungeonView()
		for kind == cellDoorsTwo || kind == cellDoor {
			if firstAbleMember() <= noneAble {
				break
			}
			enterRoom(cell)
			if mapID == 0 {
				break
			}
			cell = *dungeonCellAt(int(mapX), int(mapY), int(level))
			kind = cell & cellKindMask
		}
		restoreDungeonLook(1)
		if mapID > settlements {
			redrawDungeon()
		}
	} else {
		switch cell {
		case cellSleepField, cellSleepField | cellSeen:
			sleepTrap()
		case cellPoison, cellPoison | cellSeen:
			poisonTrap()
		case cellFireField, cellFireField | cellSeen:
			printMessage(msgFireLine)
			hurtParty()
		case cellBombTrap, cellBombTrap | cellSeen:
			printMessage(msgBombTrapLine)
			printMessage(msgKaboomLine)
			*dungeonCellAt(int(mapX), int(mapY), int(level)) &= cellSeen
			hurtParty()
		case cellPit, cellPit | cellSeen:
			fallThroughPit(cell)
		}
		showStatus()
	}
	passTurn()
}

// dungeonMainLoop runs the dungeon until the party leaves it or dies. fromWorld is set when
// the party just came down from the surface.
// C: DUNGEON_0e2e_MainLoop
func dungeonMainLoop(fromWorld bool) {
	fresh := 0
	if fromWorld {
		fresh = 1
	}
	going := true
	able := someoneAble
	timeStep := 0
	dungeonGraphics = worldTilesLoaded
	revealSeenRooms()
	kind := *dungeonCellAt(int(mapX), int(mapY), int(level)) & cellKindMask
	drawDungeonCaptions()
	showLevelAndFacing()
	switch dungeon := int(mapID) - settlements; dungeon {
	case 1, 4, 5:
		roomWallTile, moongateFloor, dungeonLook = remainsWallTile, remainsFloorTile, lookRemains
	default:
		roomWallTile, moongateFloor = caveWallTile, caveFloorTile
		if dungeon == 6 || dungeon == 7 {
			dungeonLook = lookCaveIn
		} else {
			dungeonLook = lookStalactite
		}
	}
	if kind == cellDoor || kind == cellDoorsTwo {
		hideDungeonView()
		enterRoom(*dungeonCellAt(int(mapX), int(mapY), int(level)))
	}
	if mapID > settlements {
		restoreDungeonLook(fresh)
		resumeDungeon(fresh)
		redrawDungeon()
	}
	for going && mapID > settlements {
		turned := 1
		able = firstAbleMember()
		if able == onlySleepers {
			printChar('\n')
			printOpenCaption()
			printMessage(msgZzzzzzLine)
		}
		if able < someoneAble {
			going = false
			turned = 0
		}
		// fix: the decompilation goes back to the loop's top when no one is able, which
		// with everyone asleep never lets a turn pass; the turn passes as in the towns.
		if able == someoneAble {
			if spellIcon != iconTimeStop && spellIcon != iconQuickness {
				timeStep = 1
				passTime(1)
			} else if spellIcon == iconQuickness {
				timeStep ^= 1
				if timeStep != 0 {
					passTime(1)
				}
			} else {
				timeStep = 0
			}
			if key := readDungeonKey(); key > -1 {
				turned = dungeonCommand(key)
			}
		} else if able == onlySleepers {
			passTime(1)
		}
		if mapID <= settlements {
			turned = 0
			going = false
		}
		// fix: the original reads the cell before this test, past the dungeon map once
		// the party has left for the surface
		if turned != 0 {
			afterDungeonTurn(*dungeonCellAt(int(mapX), int(mapY), int(level)), timeStep)
		}
		if mapID <= settlements {
			going = false
		}
	}
	if dungeonGraphics&worldTilesLoaded != 0 {
		dungeonGraphics &^= timeTilesNormal
	} else {
		dungeonGraphics |= timeTilesNormal
	}
	hideDungeonView()
	if able < someoneAble {
		partyDied()
	}
}

// dungeonStep sounds a footstep every other call, each fainter than the last.
// C: DUNGEON_1020
func dungeonStep() {
	footstepOdd = !footstepOdd
	if footstepOdd && footstepLength != 0 {
		whiteNoise(1, int(footstepLength), stepNoiseLimit)
		decreaseByte(&footstepLength, stepNoiseFade)
	}
}

// showDungeonSign prints sign number n of this dungeon on the wall ahead.
// C: DUNGEON_104c
func showDungeonSign(n int) {
	dungeon := int(mapID) - firstDungeonMap
	if n > int(signCount[dungeon]) {
		return
	}
	at := int(signFirst[dungeon]) + n - 1
	setWindowSize(windowFrame, int(signColumn[at]), int(signRow[at]), 39, 24)
	selectWindow(windowFrame)
	selectFont(1)
	setTextPosition(0, 0)
	printChar(charToggleInverse)
	printString(signTexts[at])
	printChar(charToggleInverse)
	selectFont(0)
	selectWindow(windowScroll)
	setWindowSize(windowFrame, 0, 0, 39, 24)
}

// visibleCell returns the dungeon cell at (x, y) of the level, without its "seen" bit
// unless it is a wall or door.
// C: DUNGEON_10dc
func visibleCell(x, y int) int {
	cell := int(*dungeonCellAt(x&(dungeonSize-1), y&(dungeonSize-1), int(level)))
	if cell < cellImpossible {
		cell &= notSeenLock
	}
	return cell
}

// drawWanderer draws the wanderer depth cells ahead in its pose, and steps the pose.
// C: DUNGEON_111e
func drawWanderer(depth int) {
	foe := &actors[wandererSlot]
	if spellIcon == iconTimeStop {
		foe.state = wandererStates[foe.tile]
	}
	mirrored := 0
	if foe.target == wandererMirrored {
		mirrored = 1
	}
	pose := int(foe.state) & poseBits
	motion := int(foe.state) & motionBits
	frame := int(foe.state) & cellLow
	upper, lower := 1, 1
	if spellIcon != iconTimeStop {
		upper = boolInt(randomRange(0, 100) < 50)
		lower = boolInt(randomRange(0, 100) < 50)
		switch motion {
		case motionToggle:
			frame = boolInt(frame == 0)
			upper, lower = frame, frame
		case motionWalk:
			frame--
			if frame < 0 {
				frame = 3
			}
			upper = int(wandererSteps[frame]) >> 1
			lower = int(wandererSteps[frame]) & 1
		}
	}
	if pose != 0 {
		lower = boolInt(pose == poseOpposed) ^ upper
	}
	if monsterPictures == nil {
		printMessage(msgHeyWhatSGoingOnHere)
		return
	}
	y := int(wandererY[mirrored][depth])
	putPicture(monsterPictures, upper*3+depth-1, int(wandererX[depth]), y, mirrored)
	putPictureMirrored(monsterPictures, lower*3+depth-1, centreX, y, mirrored)
	foe.state = byte(motion + pose + frame)
}

// The wanderer's pose bits.
const (
	poseBits     = 144
	poseOpposed  = 144
	motionBits   = 96
	motionToggle = 32
	motionWalk   = 96
)

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// drawField draws a field depth cells ahead as flickering lines of its colour.
// C: DUNGEON_127e
func drawField(depth, kind int) {
	saved := video.clipping
	video.clipping = false
	switch kind {
	case 0:
		setPenColour(magentaColour + brightOffset)
	case 1:
		setPenColour(greenColour + brightOffset)
	case 2:
		setPenColour(redColour + brightOffset)
	case 3:
		setPenColour(blueColour + brightOffset)
	}
	for range int(fieldLines[depth]) {
		x := randomRange(int(fieldTop[depth]), int(fieldBottom[depth]-fieldWidth[depth]))
		horizLine(x, randomRange(int(fieldTop[depth]), int(fieldBottom[depth])), int(fieldWidth[depth])+x)
	}
	video.clipping = saved
}

// drawPiece draws a wall piece (below wallPieces) or an item picture, on the right half
// mirrored when right is set; row picks the item's place.
// C: DUNGEON_134a
func drawPiece(piece, row, right int) {
	if piece < wallPieces {
		x := int(pieceX[right][piece&3])
		if right != 0 && (piece&sidePieceMask == sidePieceRow || piece&cornerPieceMask == cornerPieceRow) {
			x = centreX
		}
		if right != 0 {
			putPictureMirrored(dungeonPictures, piece, x, pieceTop, 0)
		} else {
			putPicture(dungeonPictures, piece, x, pieceTop, 0)
		}
		return
	}
	item := ((piece + 1) >> 1) - 16
	column := ((piece + 1) & 6) >> 1
	x := int(itemX[column])
	if right != 0 {
		x = centreX
	}
	var y int
	switch {
	case item >= 8:
		y = int(itemY[row][column])
	case row != 0:
		y = int(itemY[1][column])
	default:
		y = itemFloor
	}
	if right != 0 {
		putPictureMirrored(itemPictures, item, x, y, row)
	} else {
		putPicture(itemPictures, item, x, y, row)
	}
}

// drip draws a stalactite's drop at (x, y) in its state, and returns its next state; the
// last state is the drop's sound.
// C: DUNGEON_145c
func drip(x, y, state, depth int) int {
	if state == dripSound {
		sweepTone(dripSoundFrom, dripSoundTo, 1, -(depth*8 - dripSoundDepth))
		return 0
	}
	switch {
	case state != dripShine:
		setPenColour(blueColour)
	case videoDriver == 0 || videoDriver == cgaComposite:
		setPenColour(cyanColour)
	default:
		setPenColour(brightCyanColour)
	}
	horizLine(x-1, y, x+1)
	vertLine(x, y-1, y+1)
	if state < dripShine {
		setPenColour(blueColour + brightOffset)
		plot(x, y)
	}
	if state != 0 || randomRange(0, 64) < dripOdds {
		state++
	}
	return state
}

// setDrip keeps a stalactite's drop state in its cell.
func setDrip(x, y, state int) {
	cell := dungeonCellAt(x&(dungeonSize-1), y&(dungeonSize-1), int(level))
	*cell &= clearTrapMask
	*cell += byte(state)
}

// drawFront draws the wall depth cells ahead, if any. It returns whether the view goes on
// past it.
// C: DUNGEON_150a
func drawFront(x, y, depth int) int {
	cell := visibleCell(x, y)
	if cell < cellDoor {
		return 1
	}
	piece := int(frontPieces[(cell-cellDoor)>>4]) + depth
	if depth == 0 {
		piece = doorPiece
	}
	drawPiece(piece, 0, 0)
	drawPiece(piece, 0, 1)
	if (depth == 1 || depth == 2) && dungeonLook == lookStalactite && cell&cellKindMask == cellRubble {
		state := cell & dripState
		state = drip(frontDripX, int(dripFrontY[state+(depth-1)*5])+pieceTop, state, depth)
		setDrip(x, y, state)
	}
	if depth == 1 {
		if cell&cellKindMask == cellWall && cell&cellLow != 0 {
			showDungeonSign(cell & cellLow)
		} else if dungeonLook == lookRemains && cell&cellKindMask == cellRubble && randomRange(0, 64) < eyesOdds {
			// eyes glint in the dark
			setPenColour(redColour + brightOffset)
			horizLine(92, eyeTop, 93)
			horizLine(91, eyeTop+1, 93)
			horizLine(97, eyeTop, 98)
			horizLine(97, eyeTop+1, 99)
		}
	}
	if cell&cellKindMask == cellDoors && depth == 0 {
		video.clipping = true
		return 1
	}
	return 0
}

// drawSide draws the wall or opening beside the way depth cells ahead, on the right when
// right is set.
// C: DUNGEON_1682
func drawSide(x, y, right, depth int) {
	cell := visibleCell(x, y)
	if cell < cellDoor {
		drawPiece(depth+sidePieceOpen, 0, right)
		return
	}
	switch cell & cellKindMask {
	case cellRubble:
		drawPiece(depth+sidePieceRubble, 0, right)
		if depth >= 2 || dungeonLook != lookStalactite {
			break
		}
		state := cell & dripState
		dripX := nearDripX
		if depth != 0 {
			dripX = farDripX
		}
		if right != 0 {
			dripX = mirrorX - dripX
		}
		state = drip(dripX, int(dripSideY[state+depth*5])+pieceTop, state, depth)
		setDrip(x, y, state)
	case cellDoor, cellDoors, cellDoorsTwo:
		drawPiece(depth+sidePieceDoor, 0, right)
	default:
		drawPiece(depth, 0, right)
	}
}

// drawSpray draws frame of a fountain's spray depth cells ahead.
// C: DUNGEON_1786
func drawSpray(depth, frame int) {
	dim := videoDriver == 0 || videoDriver == cgaComposite
	if dim {
		setPenColour(0)
	} else {
		setPenColour(blueColour)
	}
	dot := func(table []byte, i, dx, mirror int) {
		x, y := int(table[i*2]), int(table[i*2+1])
		plot(x+dx, y+itemFloor)
		plot(mirror-x, y+itemFloor)
	}
	switch depth {
	case 0:
		if !dim {
			setPenColour(blueColour + brightOffset)
		}
		for i := range sprayNearDots {
			x, y := int(sprayNear[(frame*sprayNearDots+i)*2]), int(sprayNear[(frame*sprayNearDots+i)*2+1])
			plot(x, y)
			plot(mirrorX-x, y)
		}
		x, y := int(sprayTop[frame*2]), int(sprayTop[frame*2+1])
		plot(x, y)
		plot(mirrorX-x, y)
	case 1:
		for i := range sprayDots2 {
			dot(spraySecond, frame*sprayDots2+i, 72, 118)
		}
	case 2:
		for i := range sprayDots3 {
			dot(sprayThird, frame*sprayDots3+i, 80, 110)
		}
	case 3:
		dot(sprayFar, frame, 88, 102)
	}
}

// drawCellContents draws what stands in the cell depth cells ahead: ladders, chests,
// fountains, fields, the wanderer.
// C: DUNGEON_1952
func drawCellContents(x, y, depth int) {
	cell := visibleCell(x, y)
	kind := cell >> 4
	if kind > 0 && kind < 8 && (kind != cellTrap>>4 || cell&dripState == 0) {
		if upperPieces[kind] != 0 {
			drawPiece(int(upperPieces[kind])+depth*2, 1, 0)
			drawPiece(int(upperPieces[kind])+depth*2, 1, 1)
		}
		if lowerPieces[kind] != 0 {
			drawPiece(int(lowerPieces[kind])+depth*2, 0, 0)
			drawPiece(int(lowerPieces[kind])+depth*2, 0, 1)
		}
		if kind == cellFountain>>4 {
			drawSpray(depth, fountainFrame)
		}
	} else if kind == cellField>>4 {
		drawField(depth, cell&dripState)
	}
	if kind < cellImpossible>>4 && *dungeonCellAt(x&(dungeonSize-1), y&(dungeonSize-1), int(level))&cellSeen != 0 {
		// the trap door of a seen trap
		trapDoor := int(lowerPieces[cellTrap>>4])
		drawPiece(trapDoor+depth*2, 1, 0)
		drawPiece(trapDoor+depth*2, 1, 1)
	}
	if depth != 0 && x&(dungeonSize-1) == int(actors[wandererSlot].x) && y&(dungeonSize-1) == int(actors[wandererSlot].y) {
		drawWanderer(depth)
	}
	if depth == 0 {
		fountainFrame++
		if fountainFrame > sprayFrames-1 {
			fountainFrame = 0
		}
	}
}

// drawDungeon draws the view ahead on page 1, the near walls beside the party when full.
// C: DUNGEON_1a90
func drawDungeon(full int) {
	frameKeyHint = noFrameKey
	selectPage(1)
	setPenColour(0)
	fillRect(pageLeft, pageTop, pageRight, pageBottom)
	setClipWindow(clipLeft, clipTop, clipRight, clipBottom)
	video.clipping = false
	if lightSpell != 0 || torchLight != 0 {
		x, y := int(mapX), int(mapY)
		dx, dy := demoDeltaX[facing], demoDeltaY[facing]
		sx, sy := int(byte(sideX[facing])), int(byte(sideY[facing]))
		depth := 0
		for ; depth < depths; depth++ {
			if drawFront(x, y, depth) == 0 {
				break
			}
			if depth != 0 || visibleCell(int(mapX), int(mapY)) != cellDoors && full != 0 {
				drawSide(sx+x, sy+y, 0, depth)
				drawSide(x-sx, y-sy, 1, depth)
			}
			x += dx
			y += dy
		}
		for depth--; depth >= 0; depth-- {
			x -= dx
			y -= dy
			drawCellContents(x, y, depth)
		}
	} else {
		fillRect(viewLeft, viewTop, viewRight, viewBottom)
	}
	selectPage(0)
	video.clipping = false
}

// redrawDungeon draws the view and the status and shows them.
// C: DUNGEON_1be0
func redrawDungeon() {
	showStatus()
	drawDungeon(1)
	copyArea(1, 0, fullCopyLeft, pageTop, pageRight, pageBottom)
	dungeonStep()
	showLevelAndFacing()
}

// canKlimb tells whether the party may come out on toLevel; by magic only onto an empty
// cell.
// C: DUNGEON_1c0c
func canKlimb(toLevel, magic int) bool {
	kind := *dungeonCellAt(int(mapX), int(mapY), toLevel) & cellKindMask
	// the original also lists the wall kinds, which kind != cellNothing already takes in
	return magic == 0 || kind == cellNothing
}

// dungeonKlimb climbs way levels (1 down, -1 up). It returns whether the party climbed out
// of the dungeon.
// C: DUNGEON_1c6a_Klimb
func dungeonKlimb(way, magic int) int {
	if way > 0 {
		printMessage(msgDownLine)
	} else {
		printMessage(msgUpLine)
	}
	if way > 0 && level == lastLevel || way < 0 && level == 0 {
		return 1
	}
	if !canKlimb(way+int(level), magic) {
		printMessage(msgFailedLine)
		sweepTone(klimbFailFrom, klimbFailTo, 1, klimbFailTime)
		return 0
	}
	level = byte(int(level) + way)
	if way > 0 {
		dungeonMode = dungeonModeUp
	} else {
		dungeonMode = dungeonModeDown
	}
	resumeDungeon(1)
	redrawDungeon()
	return 0
}

// leaveDungeon puts the party outside the dungeon's entrance, in Britannia or the
// Underworld.
// C: DUNGEON_1d08_Exit
func leaveDungeon() {
	mapX = placeX[mapID-1]
	mapY = placeY[mapID-1]
	printMessage(msgExitToSp)
	if level != 0 {
		level = underworldLevel
		printMessage(msgUnderworldLine2)
	} else {
		level = 0
		printMessage(msgBritanniaLine2)
	}
	mapID = 0
}

// attackInDungeon fights the wanderer standing ahead.
// C: DUNGEON_1d4a_AttackCmd
func attackInDungeon() int {
	printMessage(msgAttackLine)
	x := (demoDeltaX[facing] + int(mapX)) & (dungeonSize - 1)
	y := (demoDeltaY[facing] + int(mapY)) & (dungeonSize - 1)
	if x != int(actors[wandererSlot].x) || y != int(actors[wandererSlot].y) {
		printMessage(msgWhatLine)
		return 0
	}
	combatKind = specialDungeonCombat
	hideDungeonView()
	buildRoomWalls()
	drawDungeonFrame()
	enterSpecialMap(specialDungeonCombat, int(actors[wandererSlot].hull), 0)
	switch exitDirection {
	case exitUp:
		if level != 0 {
			level--
		} else {
			leaveDungeon()
		}
	case exitDown:
		if level < lastLevel {
			level++
		} else {
			leaveDungeon()
		}
	}
	restoreDungeonLook(1)
	if mapID != 0 {
		resumeDungeon(1)
		redrawDungeon()
	}
	return 0
}

// klimbInDungeon climbs the ladder (or, with the grapple, the hole) the party stands on.
// C: DUNGEON_1e10_KlimbCmd
func klimbInDungeon() int {
	up, down := false, false
	kind := visibleCell(int(mapX), int(mapY)) & cellKindMask
	hole := *dungeonCellAt(int(mapX), int(mapY), int(level)) & cellSeen
	if kind == cellLadderUp || kind == cellLadders || hole != 0 && grapple != 0 {
		up = true
	}
	if kind == cellLadderDown || kind == cellLadders || kind == cellTrap {
		down = true
	}
	switch {
	case up && down:
		printMessage(msgKlimbUD)
		for up && down {
			switch getChar() {
			case keyUp, 'U':
				down = false
			case keyDown, 'D':
				up = false
			case keySpace:
				printMessage(msgPassLine)
				up, down = false, false
			}
		}
	case up || down:
		printMessage(msgKlimb)
	case hole != 0:
		printMessage(msgKlimbWithWhatLine)
		return 0
	default:
		printMessage(msgKlimbWhatLine)
		return 0
	}
	if up && dungeonKlimb(-1, 0) != 0 || down && dungeonKlimb(1, 0) != 0 {
		leaveDungeon()
	}
	return 1
}
