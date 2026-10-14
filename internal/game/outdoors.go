package game

// The overworld: loading Britannia and the Underworld a chunk at a time (OUTSUBS), the
// party's moves, sailing, the monsters roaming outdoors and entering places (MAINOUT).

// Addresses of the overworld's tables in the data segment.
const (
	placeNamesAddress    = 7738  // C: D_1e3a (the names of the settlements and dungeons)
	placeXAddress        = 7818  // C: D_1e8a
	placeYAddress        = 7858  // C: D_1eb2
	shrineNamesAddress   = 8014  // C: D_1f4e
	shrineXAddress       = 8046  // C: D_1f6e
	shrineYAddress       = 8054  // C: D_1f76
	windXAddress         = 10742 // C: D_29f6 (the way each wind blows)
	windYAddress         = 10746 // C: D_29fa
	landFoesAddress      = 11200 // C: D_2bc0 (the monsters met on land in Britannia)
	underLandFoesAddress = 11212 // C: D_2bcc (in the Underworld)
	seaFoesAddress       = 11220 // C: D_2bd4
	underSeaFoesAddress  = 11226 // C: D_2bda
	landOddsAddress      = 11228 // C: D_2bdc (out of 256, for each of landFoes)
	underLandOddsAddress = 11240 // C: D_2be8
	seaOddsAddress       = 11248 // C: D_2bf0
	underSeaOddsAddress  = 11254 // C: D_2bf6
	pirateWindAddress    = 11256 // C: D_2bf8 (the turns a pirate ship waits for each wind)
	shadowlordSightAddr  = 11288 // C: D_2c18 (where a Shadowlord sees the party from)
	shrineChunksAddress  = 14438 // C: D_3866 (the chunks with the shrines)
	runeChunksAddress    = 14446 // C: D_386e (the chunks with the ruined shrines)
	chunkIndexAddress    = 14454 // C: D_3876 (where each chunk is in BRIT.DAT, 255 sea)
	shardXAddress        = 14854 // C: D_3a06
	shardYAddress        = 14858 // C: D_3a0a
	shardOwnerAddress    = 14862 // C: D_3a0e
	apparitionNotesAddr  = 14886 // C: D_3a26
	classTilesAddress    = 6878  // C: D_1ade (the tile of each class)
	karmaTextsAddress    = 6772  // C: D_1a74 (where each karma's words are in KARMA.DAT)
	classLettersAddress  = 30560 // the classes' letters, "AMBFDTPRS"
	versionAddress       = 11087 // the version string Ctrl-V prints
)

const (
	placeCount       = 40
	settlementCount  = 32
	shrineCount      = 8
	windCount        = 4
	landFoeKinds     = 12
	underLandKinds   = 8
	seaFoeKinds      = 6
	underSeaKinds    = 2
	pirateFacings    = 4
	sightSize        = 60
	chunkCount       = 256
	shardsLaid       = 3
	apparitionNotes  = 6
	classCount       = 9
	karmaTextCount   = 5
	chunkSize2       = chunkSize * chunkSize // the bytes of a chunk
	lastWorldChunk   = 240                   // the last chunk's position
	worldChunkStride = 4096                  // the bytes of a row of chunks
)

// Overworld tables read from DATA.OVL.
var (
	placeNames     [placeCount]string
	placeX, placeY []byte
	shrineNames    [shrineCount]string
	shrineX        []byte
	shrineY        []byte
	windX, windY   []byte
	landFoes       []byte
	underLandFoes  []byte
	seaFoes        []byte
	underSeaFoes   []byte
	landOdds       []byte
	underLandOdds  []byte
	seaOdds        []byte
	underSeaOdds   []byte
	pirateWind     [pirateFacings][]int
	sightTable     []byte
	shrineChunks   []byte
	runeChunks     []byte
	chunkIndex     []byte
	shardX, shardY []byte
	shardOwner     []byte
	apparitionTune []int
	classTiles     []byte
	karmaTexts     []uint16
	classLetters   string
)

// loadOutdoorTables reads the overworld's tables.
func loadOutdoorTables() {
	for i := range placeNames {
		if at := int(segment.Word(placeNamesAddress + i*2)); at != 0 {
			placeNames[i] = message(at)
		}
	}
	placeX = segment.Bytes(placeXAddress, placeCount)
	placeY = segment.Bytes(placeYAddress, placeCount)
	for i := range shrineNames {
		shrineNames[i] = message(int(segment.Word(shrineNamesAddress + i*2)))
	}
	shrineX = segment.Bytes(shrineXAddress, shrineCount)
	shrineY = segment.Bytes(shrineYAddress, shrineCount)
	windX = segment.Bytes(windXAddress, windCount)
	windY = segment.Bytes(windYAddress, windCount)
	landFoes = segment.Bytes(landFoesAddress, landFoeKinds)
	underLandFoes = segment.Bytes(underLandFoesAddress, underLandKinds)
	seaFoes = segment.Bytes(seaFoesAddress, seaFoeKinds)
	underSeaFoes = segment.Bytes(underSeaFoesAddress, underSeaKinds)
	landOdds = segment.Bytes(landOddsAddress, landFoeKinds)
	underLandOdds = segment.Bytes(underLandOddsAddress, underLandKinds)
	seaOdds = segment.Bytes(seaOddsAddress, seaFoeKinds)
	underSeaOdds = segment.Bytes(underSeaOddsAddress, underSeaKinds)
	for i := range pirateWind {
		pirateWind[i] = segment.SignedWords(pirateWindAddress+i*windCount*2, windCount)
	}
	sightTable = segment.Bytes(shadowlordSightAddr, sightSize)
	shrineChunks = segment.Bytes(shrineChunksAddress, shrineCount)
	runeChunks = segment.Bytes(runeChunksAddress, shrineCount)
	chunkIndex = segment.Bytes(chunkIndexAddress, chunkCount)
	shardX = segment.Bytes(shardXAddress, shardsLaid)
	shardY = segment.Bytes(shardYAddress, shardsLaid)
	shardOwner = segment.Bytes(shardOwnerAddress, shardsLaid)
	apparitionTune = segment.SignedWords(apparitionNotesAddr, apparitionNotes)
	classTiles = segment.Bytes(classTilesAddress, classCount)
	karmaTexts = segment.Words(karmaTextsAddress, karmaTextCount)
	classLetters = message(classLettersAddress)
}

// Map numbers and places of the overworld.
const (
	firstHut         = 13 // places 13 to 17 are huts and lighthouses, entered without a name
	lastHut          = 17
	firstDungeon     = 32 // places 32 to 39 are the dungeons
	doomPlace        = 39
	townStartX       = 15
	townStartY       = 30
	underworldMap    = 255 // level of the Underworld
	dungeonDataStart = 16384
	dungeonBytes     = 512
	fallsX, fallsY   = 54, 138
	codexGateX       = 233
	codexGateY       = 235
	trollToll        = 99
)

// The tiles of the overworld the code tests by number.
const (
	mapTileBridge        = mapTile106 // and 107
	mapTileReef          = mapTile47  // walls of rock at sea
	mapTileShallows      = mapTileWater3
	mapTileDock          = mapTile71
	mapTileSwamp         = mapTilePoison
	mapTileBrush         = mapTile9
	mapTileHills         = mapTile30 // and 31
	mapTileMountainsLow  = mapTile32
	mapTileMountainsHigh = mapTile38
	mapTileUnlit         = 255
	mapTileDeepWater     = mapTileWater1
	mapTileFallsMask     = 212 // the waterfall tiles are 212 to 215
	actorShipMask        = 248
)

// entranceHidden says whether the cave, mine and dungeon tiles of chunk show as tile 223:
// everywhere but in the eight chunks of shrineChunks once their shrine's state is set.
// C: OUTSUBS_0000
func entranceHidden(chunk uint16) bool {
	id := byte(chunk >> bitsPerByte)
	for i, c := range shrineChunks {
		if c == id {
			return shrineStates[i] == 0
		}
	}
	return true
}

// shrineRuined says whether the ruined shrine of chunk shows as ruins.
// C: OUTSUBS_004a
func shrineRuined(chunk uint16) bool {
	id := byte(chunk >> bitsPerByte)
	for i, c := range runeChunks {
		if c == id {
			return shrineRunes[i] >= underworldLevel
		}
	}
	return false
}

// loadChunk reads the chunk at offset (y*4096 + x*16) of name into quadrant of the map.
// C: OUTSUBS_0098
func loadChunk(name string, quadrant int, offset uint16) {
	quarter := mapTiles[quadrant*chunkSize2 : (quadrant+1)*chunkSize2]
	fileAt := int(offset)
	if name[0] == 'B' {
		chunk := chunkIndex[byte(offset>>bitsPerByte)]
		if chunk == mapTileUnlit {
			for i := range quarter {
				quarter[i] = mapTileDeepWater
			}
			return
		}
		fileAt = int(chunk) << bitsPerByte
	}
	readFile(name, quarter, fileAt)
	x := int(offset>>4) & 240
	y := int(offset>>8) & 240
	for row := range chunkSize {
		for column := range chunkSize {
			switch quarter[row*chunkSize+column] {
			case mapTileCave, mapTileMine, mapTileDungeon:
				if entranceHidden(offset) {
					*tileAt(column+x, row+y) = mapTile223
				}
			case mapTileShrine:
				if shrineRuined(offset) {
					*tileAt(column+x, row+y) = mapTileRuins
				}
			}
		}
	}
}

// loadChunks reads the quadrants of the map uncovered by a move of dx, dy (a whole new map
// for -1 both ways), then finds the map's lighthouse fire.
// C: OUTSUBS_01b4
func loadChunks(dx, dy int) {
	offset := uint16(int(chunkX)*chunkSize + int(chunkY)*chunkSize2)
	right := uint16(chunkSize2)
	if chunkX == lastWorldChunk {
		wrap := chunkSize2 - worldChunkStride // back to the first column, in 16 bits
		right = uint16(wrap)
	}
	name := "BRIT.DAT"
	if level >= underworldLevel {
		name = "UNDER.DAT"
	}
	if dx == -1 || dy == -1 {
		loadChunk(name, 0, offset)
	}
	if dx == 1 || dy == -1 {
		loadChunk(name, 1, offset+right)
	}
	if dx == -1 || dy == 1 {
		loadChunk(name, 2, offset+worldChunkStride)
	}
	if dx == 1 || dy == 1 {
		loadChunk(name, 3, offset+right+worldChunkStride)
	}
	fireX, fireY = -1, -1
	for at, tile := range mapTiles {
		if tile == mapTileLighthouse {
			fireX = int16(at & 15)
			fireY = int16(at&240) >> 4
			if at&chunkSize2 != 0 {
				fireX += chunkSize
			}
			if at&(2*chunkSize2) != 0 {
				fireY += chunkSize
			}
			break
		}
	}
}

// shiftChunks moves the quadrants the party walks towards to the opposite side.
// C: OUTSUBS_02c8
func shiftChunks(dx, dy int) {
	to := 0
	if dx < 0 || dy < 0 {
		to = 3
	}
	from := 1
	if dx < 0 || dy > 0 {
		from = 2
	}
	to2 := to + dy + dx*2
	from2 := from + dy + dx*2
	quarter := func(q int) []byte { return mapTiles[q*chunkSize2 : (q+1)*chunkSize2] }
	copy(quarter(to), quarter(from))
	copy(quarter(to2), quarter(from2))
}

// worldSaveFile is the file with the actors of the current world.
// C: OUTSUBS_0368_GetWorldSavefile
func worldSaveFile() string {
	if level == 0 {
		return "BRIT.OOL"
	}
	return "UNDER.OOL"
}

// enterTown enters the settlement the party stands on, kind its kind of place; it returns
// 1 when there is none.
// C: OUTSUBS_0388_EnterTown
func enterTown(kind int) int {
	printMessage(kind)
	place := 0
	for ; place < settlementCount; place++ {
		if placeX[place] == mapX && placeY[place] == mapY {
			break
		}
	}
	if place == settlementCount {
		printMessage(msgWhatTownLine)
		return 1
	}
	if place < firstHut || place > lastHut {
		printMessage(msgNewlines2)
		printChar(charCentre)
		printString(placeNames[place])
		printChar(charUncentre)
	}
	printChar('\n')
	writeFile(worldSaveFile(), actorBytes())
	mapID = byte(place + 1)
	level = 0
	mapX, mapY = townStartX, townStartY
	return 0
}

// fall takes the party down a waterfall, into the Underworld from the falls of the
// Serpent's Spine.
// C: OUTSUBS_0458_Falls
func fall() {
	printMessage(msgFALLSLine)
	moveParty(0, 1)
	sleepFrames(1)
	moveParty(0, 1)
	sweepTone(2500, 800, 1, 300)
	tile := partyTile
	partyTile = 0
	sleepFrames(1)
	for i := range int(partySize) {
		if party[i].status != statusDead && int(party[i].dex) <= randomDamage() {
			hurtMember(i, 1)
		}
	}
	sleepFrames(2)
	partyTile = tile
	if mapX == fallsX && mapY == fallsY {
		printMessage(msgFallingIntoUnderworldLine)
		level = underworldMap
		writeFile("BRIT.OOL", actorBytes())
		readActorBytes("UNDER.OOL")
		writeFile("UNDER.OOL", actorBytes())
		redrawWorld()
	}
}

// Where Lord British's amulet lies in the Underworld until taken.
const (
	amuletSlot       = 28
	firstShardSlot   = 29
	amuletX, amuletY = 105, 225
	amuletOwner      = 243
)

// placeArtifacts lays the amulet and the shards not taken yet in the Underworld.
// C: OUTSUBS_0566
func placeArtifacts() {
	if level == 0 {
		return
	}
	if amulet == 0 {
		a := &actors[amuletSlot]
		a.tile, a.animTile = actorAmuletLb, actorAmuletLb
		a.x, a.y, a.z = amuletX, amuletY, underworldMap
		a.hull = amuletOwner
	}
	for i := range shardsLaid {
		if shards[i] == 0 && shrineVisits[i] < underworldLevel {
			a := &actors[firstShardSlot+i]
			a.tile, a.animTile = actorShard, actorShard
			a.x, a.y, a.z = shardX[i], shardY[i], underworldMap
			a.hull = shardOwner[i]
		}
	}
}

// burn hurts the party on lava.
// C: OUTSUBS_05ee_StepOnBurningTile
func burn() {
	updateFrame()
	printMessage(msgBurningLine)
	hurtParty()
}

// poisonParty poisons the members who fail a dexterity roll in a swamp.
// C: OUTSUBS_05fc_StepOnPoisonTile
func poisonParty() {
	for i := range int(partySize) {
		m := &party[i]
		if m.status != statusDead && m.status != statusPoisoned && int(m.dex) < randomRange(1, maxStat) {
			m.status = statusPoisoned
			printMessage(msgPoisonedLine)
		}
	}
}

// Lord British's apparition.
const (
	apparitionSlot   = 10
	apparitionCell   = 5
	apparitionTile   = actor116
	firstActorShape  = 256
	maxStat          = 30
	hpPerLevel       = 30
	expPerLevelStep  = 100
	karmaPerText     = 20
	karmaTextAbove   = 671 // the words for the highest karma
	karmaTextBytes   = 2000
	karmaTextEntries = 4
)

// apparition is Lord British appearing to heal the party, raise the members who earned it
// a level and speak of the party's karma.
// C: OUTSUBS_0658
func apparition() {
	printMessage(msgAnApparitionLine)
	pulse(2620, 1, 10000, 2500, 6)
	for _, note := range apparitionTune {
		pulse(note, 1, 5000, 200, 13)
	}
	ghost := &actors[apparitionSlot]
	ghost.x, ghost.y = apparitionCell, apparitionCell
	ghost.tile, ghost.animTile = actorCircle, actorCircle
	revealTile(firstActorShape+apparitionTile, apparitionCell, apparitionCell)
	ghost.tile, ghost.animTile = apparitionTile, apparitionTile
	shown := ghost
	for i := range int(partySize) {
		m := &party[i]
		if m.status != statusDead {
			m.hp = m.maxHP
			m.status = statusGood
			for c := range combatants {
				if int(combatants[c].entity) == i {
					shown = &actors[combatants[c].actorSlot]
					break
				}
			}
			tile := classTiles[indexOf(classLetters, m.class)]
			shown.tile, shown.animTile = tile, tile
			ghost.state = 0
			sleepFrames(1)
			pulse(5500, 1, 5000, 200, 13)
			setPenColour(brightWhiteColour)
			xorFillRect(8, 8, 183, 183)
			pulse(5500, 1, 60000, 2500, 1)
			for range 3 {
				ghost.state = 1
				sleepFrames(1)
			}
			newLevel := 1
			for n := int(m.exp) / expPerLevelStep; n > 0; n >>= 1 {
				newLevel++
			}
			if int(m.level) != newLevel {
				m.level = byte(newLevel)
				m.maxHP = int16(newLevel * hpPerLevel)
				m.hp = m.maxHP
				printMessage(msgHailSp)
				printString(m.nameString())
				printMessage(msgForThyValiantDeedsIShall)
				printMessage(msgThouArtNowLevelSp)
				printNumber(newLevel, 1, ' ')
				printMessage(msgAndLine)
				switch randomRange(1, 3) {
				case 1:
					printMessage(msgStronger)
					increaseByte(&m.str, 1, maxStat)
				case 2:
					printMessage(msgQuicker)
					increaseByte(&m.dex, 1, maxStat)
				case 3:
					printMessage(msgWiser)
					increaseByte(&m.intel, 1, maxStat)
				}
				printMessage(msgChar34Space)
				getChar()
				printChar('\n')
			}
		}
		switch m.class {
		case classAvatar, classMage:
			m.magic = m.intel
		case classBard:
			m.magic = m.intel >> 1
		}
		showStatus()
	}
	printMessage(msgNewlineChar34)
	at := karmaTextAbove
	if n := int(karma) / karmaPerText; n < karmaTextEntries {
		at = int(karmaTexts[n])
	}
	text := make([]byte, karmaTextBytes)
	readFile("KARMA.DAT", text, at)
	printString(cString(text))
	printChar('"')
	getChar()
	printMessage(msgTheStrangelyFamiliarOldManVanishes)
	ghost.tile, ghost.animTile = actorCircle, actorCircle
	revealTile(int(combatMap[apparitionCell*settlementSize+apparitionCell]), apparitionCell, apparitionCell)
	ghost.tile, ghost.animTile = 0, 0
	sleepFrames(1)
	passTime(0)
}

// redrawWorld loads the overworld around the party and shows it.
// C: MAINOUT_0000
func redrawWorld() {
	showMapName()
	passTime(0)
	mapShown, viewDirty, viewChanged = 1, 1, 1
	sailHeading = 0
	chunkX = mapX & 240
	if mapX&15 < 8 {
		chunkX = (chunkX - chunkSize) & 240
	}
	chunkY = mapY & 240
	if mapY&15 < 8 {
		chunkY = (chunkY - chunkSize) & 240
	}
	loadChunks(0, 1)
	loadChunks(0, -1)
	passTime(0)
	updateVisibility()
	setWind(-1)
	showSky()
	placeArtifacts()
}

// nearDistance is how close a monster must be to slow the party.
const nearDistance = 6

// pauseNearMonster waits a frame when actor 1 is near the party. (The original's loop
// ends after its first actor.)
// C: MAINOUT_007a
func pauseNearMonster() {
	for slot := 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		dx := abs(int(mapX) - int(a.x))
		dy := abs(int(mapY) - int(a.y))
		if a.tile != 0 && a.z == level && dx < nearDistance && dy < nearDistance {
			sleepFrames(1)
		}
		break
	}
}

// The directions of the party's vehicles, added to the vehicle's first tile.
const (
	faceNorth  = 0
	faceEast   = 1
	faceSouth  = 2
	faceWest   = 3
	facingMask = 3
	shapeMask  = 252 // a vehicle's tile without its facing
	pairMask   = 254 // a horse's or carpet's tile without its facing
)

// steer turns the party's vehicle to face direction, printing how it moves; for a frigate
// it reports whether the turn (or a calm) took the move.
// C: MAINOUT_00da
func steer(direction int) bool {
	took := false
	switch partyTile & shapeMask {
	case actorHorse:
		printMessage(msgRideSp)
		switch direction {
		case faceEast:
			partyTile = actorRidingHorse
		case faceWest:
			partyTile = actorRidingHorse + 1
		}
	case actorFlyingCarpet:
		printMessage(msgFlySp)
		switch direction {
		case faceEast:
			partyTile = actorFlyingCarpet
		case faceWest:
			partyTile = actorFlyingCarpet + 1
		}
	case actorSkiff:
		printMessage(msgRowSp)
		partyTile = byte(direction) + partyTile&shapeMask
	case actorFrigate32, actorFrigate36:
		before := partyTile
		partyTile = byte(direction) + partyTile&shapeMask
		if before != partyTile {
			printMessage(msgHeadSp)
			switch direction {
			case faceNorth:
				printMessage(msgNorthLine)
			case faceSouth:
				printMessage(msgSouthLine)
			case faceEast:
				printMessage(msgEastLine)
			case faceWest:
				printMessage(msgWestLine)
			}
			viewDirty = 1
			took = true
			if actors[0].hull < weakHull {
				printMessage(msgHullWeakLine)
			}
		} else if partyTile < actorFrigate36 && windDirection == 0 {
			took = true
		}
	}
	return took
}

// weakHull is the hull strength below which a ship warns.
const weakHull = 50

// canMoveParty checks the party's step by dx, dy: other actors, the terrain, and at full
// sail the rocks.
// C: MAINOUT_01fe
func canMoveParty(dx, dy int) bool {
	if partyTile&shapeMask == actorFrigate36 {
		printMessage(msgRowingLine)
	}
	ok := true
	other := actorAt(dx+int(mapX), dy+int(mapY), int(level))
	if other != 0 {
		ok = false
		if partyTile >= actor48 || partyTile < actorFrigate32 {
			if (other >= actorFrigate36 && other < actorPirate) || other == actorCarpet || other&pairMask == actorHorse {
				ok = true
			}
		} else if partyTile >= actorSkiff && other >= actorFrigate36 && other < actorSkiff {
			ok = true
		}
	}
	tile := int(viewTile(dx+viewCentre, dy+viewCentre))
	ok = ok && canEnter(int(partyTile), tile)
	if !ok {
		switch {
		case sailHeading != 0:
			if tile == mapTileShallows {
				printMessage(msgBreakingUpLine)
			} else if tile != mapTileDock {
				printMessage(msgCollisionLine)
			}
			if tile == mapTileDock {
				printMessage(msgDockedLine)
				partyTile += 4
			} else {
				whiteNoise(100, 2000, 300)
				damageShip()
			}
			sailHeading = 0
			viewDirty = 1
		case partyTile < actorFrigate32 || other&shapeMask != actorWhirlpool:
			printMessage(msgBlockedLine)
			if tile == mapTileReef {
				printMessage(msgOuchLine)
				hurtParty()
			} else {
				tone(165, 200)
			}
			clearKeyboard()
		}
	}
	return ok
}

// moveParty moves the party by dx, dy, loading new chunks near the map's edge.
// C: MAINOUT_0354
func moveParty(dx, dy int) {
	mapX = byte(int(mapX) + dx)
	mapY = byte(int(mapY) + dy)
	viewChanged = 1
	x := int(mapX-chunkX) & worldWrapMask
	y := int(mapY-chunkY) & worldWrapMask
	const near, far = 5, 26
	if x < near || x > far || y < near || y > far {
		shiftChunks(dx, dy)
		chunkX = byte(dx*chunkSize+int(chunkX)) & 240
		chunkY = byte(dy*chunkSize+int(chunkY)) & 240
		loadChunks(dx, dy)
		updateVisibility()
	}
}

// slowTerrain lets the monsters move once or twice more in rough terrain.
// C: MAINOUT_03e0
func slowTerrain() {
	moved := 0
	tile := int(*tileAt(int(mapX), int(mapY)))
	slowness := 0
	switch {
	case tile == mapTileGrass:
	case tile == mapTileHills || tile == mapTile31:
		slowness = 1
	case tile < mapTileSwamp || tile >= mapTileHut:
	case tile >= mapTileBrush:
		slowness = 2
	default:
		slowness = 1
	}
	switch slowness {
	case 1:
		moved += moveMonstersOutdoors()
		pauseNearMonster()
		if moved == 0 {
			printMessage(msgSlowProgressLine)
		}
		passTime(2)
	case 2:
		moved += moveMonstersOutdoors()
		pauseNearMonster()
		moved += moveMonstersOutdoors()
		pauseNearMonster()
		if moved == 0 {
			printMessage(msgVerySlowLine)
		}
		passTime(4)
	}
}

// walk moves the party one step the way of the arrow key (only turning it in the dark). It
// returns whether the move took a turn.
// C: MAINOUT_0490
func walk(key int, dark bool) int {
	if partyTile&shapeMask == actorFrigate32 {
		if byte(key) != sailHeading {
			sailHeading = byte(key)
			stepCounter = 0
		}
		viewDirty = 0
	}
	if partyTile == actorAvatar || partyTile&pairMask == actorRidingHorse {
		footstep()
	}
	dx, dy := 0, 0
	var direction, name int
	switch key {
	case keyUp:
		dy, direction, name = -1, faceNorth, msgNorthLine
	case keyDown:
		dy, direction, name = 1, faceSouth, msgSouthLine
	case keyRight:
		dx, direction, name = 1, faceEast, msgEastLine
	case keyLeft:
		dx, direction, name = -1, faceWest, msgWestLine
	}
	if steer(direction) {
		return 1
	}
	if sailHeading == 0 {
		echoStep(name)
	}
	ok := canMoveParty(dx, dy)
	if !dark && ok {
		if partyTile&pairMask == actorRidingHorse {
			footstep()
		}
		moveParty(dx, dy)
		slowTerrain()
	}
	if ok {
		return 1
	}
	return 0
}

// sailSkip toggles to let a ship with its sails set move every other turn.
var sailSkip bool // C: D_a524

// readWorldKey updates the screen and reads the next command; at full sail the ship
// keeps going until a key changes course.
// C: MAINOUT_0598
func readWorldKey() int {
	defer asking(promptCommand)()
	updateFrame()
	if statusChanged {
		showStatus()
		statusChanged = false
	}
	if viewTile(viewCentre, viewCentre+1)&shapeMask == mapTileFallsMask {
		fall()
		return 0
	}
	if viewDirty != 0 && !reusePrompt() {
		printChar('\n')
		printOpenCaption()
	}
	viewDirty = 1
	key := 0
	for sailHeading != 0 {
		key = toUpper(pollKeyWithCursor())
		if key != 0 && byte(key) != sailHeading {
			break
		}
		key = int(sailHeading)
		if windDirection != 0 {
			dx, dy := 0, 0
			switch key {
			case keyUp:
				dy = -1
			case keyDown:
				dy = 1
			case keyRight:
				dx = 1
			case keyLeft:
				dx = -1
			}
			against := 1
			if int(int8(windX[windDirection-1])) != dx {
				against++
			}
			if int(int8(windY[windDirection-1])) != dy {
				against++
			}
			if against%3 <= int(stepCounter) {
				stepCounter = 0
				break
			}
		}
		if hmsCapePlans >= underworldLevel {
			sailSkip = !sailSkip
			passTime(1)
		} else {
			passTime(2)
		}
		if !sailSkip || hmsCapePlans < underworldLevel {
			moveMonstersOutdoors()
		}
		sleepFrames(1)
		stepCounter++
	}
	if sailHeading == 0 {
		key = getChar()
	}
	return key
}

// attackOutdoors fights the monster in a direction.
// C: MAINOUT_06ec_AttackCmd
func attackOutdoors() int {
	printMessage(msgAttack)
	if *tileAt(int(mapX), int(mapY)) < mapTileSwamp && (partyTile&shapeMask == actorSkiff || partyTile&pairMask == actorFlyingCarpet) {
		printMessage(msgOnFootLine)
		return 0
	}
	if selectDirection() {
		x := int(mapX) + int(moveX)
		y := int(mapY) + int(moveY)
		target := actorAt(x, y, int(level)) & shapeMask
		if target == actorPirate || (target != actorShard && target != actor232 && target >= actorWizard) {
			fightActor(int(moveX))
		} else {
			printMessage(msgNothingToAttackLine)
		}
	}
	return 0
}

// The dungeons' starting places.
const (
	underDungeonLevel = 7
	underDungeonStart = 7
	topDungeonStart   = 1
	facingEast        = 1
	facingWest        = 3
	dungeonModeDown   = 4
	dungeonModeUp     = 5
)

// enterDungeon enters the dungeon the party stands on, kind its kind of place.
// C: MAINOUT_0790_LoadDungeon
func enterDungeon(kind int) int {
	printMessage(kind)
	place := firstDungeon
	for ; place < placeCount; place++ {
		if placeX[place] == mapX && placeY[place] == mapY {
			break
		}
	}
	if place == placeCount {
		printMessage(msgWhatDungeonLine)
		return 1
	}
	if partyTile != actorAvatar {
		printMessage(msgOnFootLine3)
		// fix: the original returned nothing here (MATCHING_BUILD); no turn passes.
		return 0
	}
	if place == doomPlace {
		if shrineVisits[0]&shrineVisits[1]&shrineVisits[2] < underworldLevel {
			printMessage(msgAttackedAtEntranceLine)
			slot := freeActorSlot()
			actors[slot].tile = actorShadowlord
			fightActor(slot)
			return 0
		}
	} else {
		printMessage(msgNewlines2)
		printChar(charCentre)
		printString(placeNames[place])
		printChar(charUncentre)
	}
	printChar('\n')
	writeFile(worldSaveFile(), actorBytes())
	contents := make([]byte, dungeonBytes)
	readFile("DUNGEON.DAT", contents, place*dungeonBytes-dungeonDataStart)
	copy(dungeonMap[:], contents)
	mapID = byte(place + 1)
	if level != 0 && mapID != doomMap {
		level = underDungeonLevel
		facing = facingWest
		dungeonMode = dungeonModeDown
		mapX, mapY = underDungeonStart, underDungeonStart
	} else {
		level = 0
		mapX, mapY = topDungeonStart, topDungeonStart
		facing = facingEast
		dungeonMode = dungeonModeUp
	}
	return 1
}

// enterCommand enters the place the party stands on.
// C: MAINOUT_08de_EnterCmd
func enterCommand() int {
	result := 1
	printMessage(msgEnterSp)
	switch *tileAt(int(mapX), int(mapY)) {
	case mapTileShrine:
		printMessage(msgTheShrineOfLine)
		i := 0
		for ; i < shrineCount; i++ {
			if shrineX[i] == mapX && shrineY[i] == mapY {
				break
			}
		}
		// fix: the original read past the names when no shrine matched.
		if i < shrineCount {
			printString(shrineNames[i])
		}
		printChar('\n')
		shrine()
	case mapTileHut:
		result = enterTown(msgHut)
	case mapTileCodex:
		printMessage(msgTheShrineOfTheCodexLine)
		shrine()
	case mapTileKeep:
		result = enterTown(msgKeep)
	case mapTileVillage:
		result = enterTown(msgVillage)
	case mapTileTowne:
		result = enterTown(msgTowne)
	case mapTileCastle:
		result = enterTown(msgCastle)
	case mapTileCave:
		result = enterDungeon(msgCave)
	case mapTileMine:
		result = enterDungeon(msgMine)
	case mapTileDungeon:
		result = enterDungeon(msgDungeon)
	case mapTileRuins:
		printMessage(msgRuins)
	case mapTileLighthouse:
		result = enterTown(msgLighthouse)
	case mapTilePalacebt:
		result = enterTown(msgThePalaceOfBlackthorn)
	case mapTileCastlelb:
		result = enterTown(msgTheCastleOfLordBritish)
	default:
		printMessage(msgWhatLine)
		result = 0
	}
	return result
}

// spellNegate is the letter of the spell that keeps magic darkness away.
const spellNegate = 14

// checkDarkness puts out the light on unlit tiles; it returns whether the frame was redrawn
// dark.
// C: MAINOUT_0a1a
func checkDarkness(dark bool) bool {
	if *tileAt(int(mapX), int(mapY)) == mapTileUnlit && spellIcon != spellNegate {
		light = 0
		if !dark {
			updateFrame()
			dark = true
		}
		return dark
	}
	passTime(0)
	return false
}

// quakeChance is the roll that shakes the Underworld.
const quakeChance = 105

// earthquake rarely shakes the Underworld.
// C: MAINOUT_0a60
func earthquake() {
	if level != 0 && randomRange(0, 255) == quakeChance {
		printMessage(msgEarthquakeLine)
		shakeScreen()
		hurtParty()
	}
}

// soundOn turns the sound effects on (Ctrl-S).
var soundOn = true // C: D_a9ce

// worldLoop runs the party's turns in the overworld until it enters a place or dies.
// C: MAINOUT_0a84_MainLoop
func worldLoop() {
	dark := false
	for done := false; !done; {
		turned := 1
		dark = checkDarkness(dark)
		switch firstAbleMember() {
		case onlySleepers:
			printChar('\n')
			printOpenCaption()
			printMessage(msgZzzzzzLine)
		case noneAble:
			writeFile(worldSaveFile(), actorBytes())
			partyDied()
			done = true
			turned = 0
		default:
			if enterMoongate() != 0 {
				shrine()
			}
			if mapID != 0 {
				return
			}
			key := readWorldKey() & lowByteMask
			switch {
			case key < ' ':
				switch key {
				case keyCtrlK:
					printNumber(int(karma), 1, ' ')
					printChar('\n')
					turned = 0
				case keyCtrlE:
					printMessage(msgExitToDosSp)
					if getChar() == 'Y' {
						exitGame()
					}
					printMessage(msgNLine)
				case keyCtrlV:
					printMessage(versionAddress)
					printChar('\n')
					turned = 0
				case keyCtrlS:
					printMessage(msgSoundSp)
					if soundOn {
						printMessage(msgOffLine)
					} else {
						printMessage(msgOnLine)
					}
					soundOn = !soundOn
				case keyLeft, keyRight, keyUp, keyDown:
					turned = walk(key, dark)
				case 0:
					turned = 0
				default:
					printMessage(msgWhatLine)
				}
			case key < '0' || key > '9':
				turned = doCommand(key)
			default:
				turned = setActiveMember(key)
			}
		}
		if partyTile&shapeMask != actorFrigate32 {
			sailHeading = 0
		}
		if mapID != 0 {
			done = true
		} else if turned != 0 {
			passTime(2)
			tile := int(*tileAt(int(mapX), int(mapY)))
			switch {
			case tile&pairMask == mapTileBridge:
				trollBridge()
			case tile == mapTileSwamp && partyTile == actorAvatar:
				poisonParty()
				sleepFrames(1)
			case tile == mapTileLava:
				burn()
			case mapX == codexGateX && mapY == codexGateY && level == 0 && mapID == 0:
				printMessage(msgNewlineChar34)
				if questsActive != 0 {
					printMessage(msgPassSeekerLine)
				} else {
					printMessage(msgThouArtNotUponASacred)
					printMessage(msgPassageDeniedLine)
					mapY++
				}
			}
			earthquake()
			passTurn()
			if tile == mapTileDeepWater && (partyTile&shapeMask == actorSkiff || partyTile&pairMask == actorFlyingCarpet) {
				printMessage(msgRoughSeasLine)
				explodeAt(int(mapX), int(mapY))
				damageShip()
			}
			if tile&shapeMask == mapTileFallsMask {
				fall()
			}
			moveMonstersOutdoors()
		}
	}
}

// boughtShipTiles are the bits of boughtShip: a ship bought waits at the docks.
const (
	boughtSkiff   = 64
	boughtFrigate = 128
	skiffMask     = 63
	newShipHull   = 99
)

// outdoorsMain enters the overworld and plays it.
// C: MAINOUT_0d22_Entry
func outdoorsMain() {
	redrawWorld()
	if boughtShip >= boughtSkiff {
		slot := freeActorSlot()
		a := &actors[slot]
		a.target = boughtShip & skiffMask
		a.x, a.y, a.z = shipX, shipY, 0
		tile := byte(actorSkiff41)
		if boughtShip >= boughtFrigate {
			tile = actorFrigate37
		}
		a.tile, a.animTile = tile, tile
		a.hull = newShipHull
		boughtShip = 0
	}
	worldLoop()
}

// encounterOdds is the chance out of 30 that a monster appears, by terrain and time.
// C: MAINOUT_0d8c
func encounterOdds() int {
	if level >= underworldLevel {
		return 3
	}
	tile := int(*tileAt(int(mapX), int(mapY)))
	odds := 1
	switch {
	case tile >= mapTileMountainsLow && tile <= mapTileMountainsHigh:
		odds = 0
	case tile == mapTileSwamp || (tile >= mapTileBrush && tile <= mapTile15):
		odds = 2
	}
	// The original also tests hour >= 32, which never holds.
	const lateHour, dawnHour = 32, 5
	if hour >= lateHour || hour < dawnHour {
		return odds + 3
	}
	return odds
}

// pickWeighted picks an entry by its odds out of 256.
// C: MAINOUT_0e04
func pickWeighted(odds []byte) int {
	roll := randomRange(0, 255)
	i := 0
	for ; int(odds[i]) <= roll; i++ {
		roll -= int(odds[i])
	}
	return i
}

// foeFor picks the monster that may appear on tile, 0 for none.
// C: MAINOUT_0e4e
func foeFor(tile int) byte {
	switch {
	case tile < mapTileSwamp || (tile >= mapTile96 && tile <= mapTile111) ||
		(tile >= mapTileWaterfall && tile <= mapTileWaterfall+3) ||
		(tile >= mapTileWater228 && tile <= mapTileWater231):
		if randomRange(0, 64) < 16 {
			if level < underworldLevel {
				if tile == mapTileDeepWater && randomRange(0, 7) == 7 {
					return actorWhirlpool
				}
				return seaFoes[pickWeighted(seaOdds)]
			}
			return underSeaFoes[pickWeighted(underSeaOdds)]
		}
	case tile == mapTile7:
		if randomRange(0, 3) == 0 {
			return actor224
		}
	default:
		if tile == mapTileSwamp && level == underworldMap {
			return actor248
		}
		if tile == mapTile12 || tile == mapTile13 {
			return 0
		}
		if tile < mapTileHut || tile&shapeMask == mapTile48 {
			if level < underworldLevel {
				return landFoes[pickWeighted(landOdds)]
			}
			return underLandFoes[pickWeighted(underLandOdds)]
		}
	}
	return 0
}

// spawnDistance keeps new monsters out of sight.
const spawnDistance = 6

// pickSpawnPlace picks a place on the loaded map out of the party's sight, in moveX, moveY.
// C: MAINOUT_0f4e
func pickSpawnPlace() {
	for {
		moveX = int16(byte(randomRange(0, worldWrapMask) + int(chunkX)))
		moveY = int16(byte(randomRange(0, worldWrapMask) + int(chunkY)))
		dx := abs(int(moveX) - int(mapX))
		dy := abs(int(moveY) - int(mapY))
		const wrapped = 250
		if dx > spawnDistance && dy > spawnDistance && dx < wrapped && dy < wrapped {
			return
		}
	}
}

// spawnTries is how often a place for a new monster is tried.
const spawnTries = 128

// spawnMonster puts a new monster somewhere out of sight.
// C: MAINOUT_0fc4
func spawnMonster() {
	var foe byte
	var x, y int
	try := 0
	for ; try < spawnTries; try++ {
		pickSpawnPlace()
		x, y = int(moveX), int(moveY)
		foe = foeFor(int(*tileAt(x, y)))
		if foe != 0 && (foe != actorPirate || *tileAt(x, y)&240 != mapTile96) {
			break
		}
	}
	if try != spawnTries {
		slot := freeActorSlot()
		setActor(foe, foe, byte(x), byte(y), level, 0, slot)
		if foe == actorPirate {
			actors[slot].hull = 100
		}
	}
}

// setActor fills actor slot.
// C: ULTIMA_3a74
func setActor(tile, animTile, x, y, z, hull byte, slot int) {
	a := &actors[slot]
	a.tile, a.animTile, a.x, a.y, a.z, a.hull = tile, animTile, x, y, z, hull
}

// roamsOutdoors says whether an actor tile is a monster moving about outdoors.
// C: MAINOUT_105c
func roamsOutdoors(tile int) bool {
	if tile >= actorPirate && tile <= actor47 {
		return true
	}
	if tile < actorMonstersFirst {
		return false
	}
	if (tile >= actorShard && tile <= actorShard+3) || (tile >= actor232 && tile <= actor232+3) {
		return false
	}
	return true
}

// shipDamageMax is the most a hit takes off a ship's hull.
const shipDamageMax = 30

// damageShip damages the party's ship, sinking it when the hull gives out; on foot the
// party is hurt.
// C: MAINOUT_109e
func damageShip() {
	if partyTile&actorShipMask != actorFrigate32 {
		hurtParty()
		return
	}
	damage := randomRange(1, shipDamageMax)
	if damage < int(actors[0].hull) {
		actors[0].hull -= byte(damage)
		showStatus()
		return
	}
	printMessage(msgShipSunkLine)
	skiffs := int(actors[0].target)
	if skiffs > 0 || magicCarpets != 0 {
		printMessage(msgAbandonShipLine)
		if skiffs > 0 {
			partyTile = partyTile&facingMask + actorSkiff
		} else {
			magicCarpets--
			partyTile = byte(randomRange(0, 1)) + actorFlyingCarpet
		}
		statusChanged = true
		return
	}
	partyTile = 0
	showStatus()
	updateFrame()
	sweepTone(660, 150, 40, 7800)
	printMessage(msgDrowningLine)
	for firstAbleMember() != noneAble {
		explodeAt(int(mapX), int(mapY))
		hurtParty()
	}
}

// pirateBroadside fires a pirate ship's cannons at the party.
// C: MAINOUT_1168
func pirateBroadside(slot, dx, dy int) {
	a := &actors[slot]
	if dx == 0 && (a.animTile == actorPirate || a.animTile == actor46) {
		a.animTile = byte(randomRange(0, 3)&2) + actor45
	}
	if dy == 0 && (a.animTile == actor45 || a.animTile == actor47) {
		a.animTile = byte(randomRange(0, 3)&2) + actorPirate
	}
	updateFrame()
	sweepTone(1300, 300, 5, 100)
	x := int(byte(int(a.x) - int(mapX) + viewCentre))
	y := int(byte(int(a.y) - int(mapY) + viewCentre))
	if cannonFire(x, y, viewCentre, viewCentre, 1) != 0 {
		updateFrame()
		explodeAt(int(mapX), int(mapY))
		damageShip()
	}
}

// Where the whirlpool takes the party.
const whirlpoolX, whirlpoolY = 34, 18

// meetMonster handles a monster that reached the party.
// C: MAINOUT_1248
func meetMonster(slot int) {
	a := &actors[slot]
	switch {
	case a.tile&shapeMask == actorWhirlpool:
		if partyTile == actorAvatar {
			damageShip()
			return
		}
		a.tile, a.animTile = 0, 0
		printMessage(msgWhirlpoolLine)
		tile := partyTile
		partyTile = actorWhirlpool
		updateFrame()
		sweepTone(660, 150, 40, 7800)
		partyTile = tile
		damageShip()
		level = underworldMap
		mapX, mapY = whirlpoolX, whirlpoolY
		redrawWorld()
	case a.tile&shapeMask != actor224:
		updateFrame()
		printMessage(msgAttackedLine)
		if *tileAt(int(mapX), int(mapY)) < mapTileSwamp && (partyTile&pairMask == actorFlyingCarpet || partyTile&shapeMask == actorSkiff) {
			damageShip()
			return
		}
		fightActor(slot)
	default:
		damageShip()
	}
}

// wrapDistance is the distance between two world coordinates, across the wrap.
func wrapDistance(a, b byte) int {
	d := abs(int(a) - int(b))
	if d > int(lowByteMask/2) {
		d = 256 - d
	}
	return d
}

// monsterActs lets a monster next to the party attack, a sea serpent spit or a pirate
// fire; it returns whether it did.
// C: MAINOUT_131a
func monsterActs(slot int) int {
	a := &actors[slot]
	tile := a.tile
	dx := wrapDistance(a.x, mapX)
	dy := wrapDistance(a.y, mapY)
	if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
		meetMonster(slot)
		return 1
	}
	if tile == actor136 || tile == actor220 {
		if dx <= 3 && dy <= 3 && randomRange(0, 7) == 0 {
			updateFrame()
			sweepTone(1300, 300, 5, 100)
			x := int(byte(int(a.x) - int(mapX) + viewCentre))
			y := int(byte(int(a.y) - int(mapY) + viewCentre))
			if cannonFire(x, y, viewCentre, viewCentre, 3) != 0 {
				explodeAt(int(mapX), int(mapY))
				damageShip()
			}
			return 1
		}
	} else if tile&shapeMask == actorPirate && ((dx == 0 && dy < 4) || (dy == 0 && dx < 4)) {
		printMessage(msgBooomLine2)
		pirateBroadside(slot, dx, dy)
		return 1
	}
	return 0
}

// monsterCanEnter says whether the monster in slot may step onto x, y.
// C: MAINOUT_1482
func monsterCanEnter(slot, x, y int) bool {
	if !canEnter(int(actors[slot].tile), int(*tileAt(x, y))) {
		return false
	}
	return topActorAt(x, y, int(level)) == 0
}

// lastMonsterX, lastMonsterY is where the last monster moved from.
var lastMonsterX, lastMonsterY byte // C: D_a526, D_a527

// notBack says whether x, y is not where the last monster came from.
// C: MAINOUT_14c8
func notBack(x, y int) bool {
	return x != int(lastMonsterX) || y != int(lastMonsterY)
}

// seesParty says whether a Shadowlord sees the party.
// C: MAINOUT_14ea
func seesParty(slot int) bool {
	a := &actors[slot]
	dx := wrapDistance(a.x, mapX)
	dy := wrapDistance(a.y, mapY)
	if dx < 6 && dy < 6 {
		return sightTable[dx+dy*viewSize] != 0
	}
	return false
}

// moveMonster moves the monster in slot by dx, dy; pirate ships turn to face their way,
// others may be slowed by the terrain.
// C: MAINOUT_1578
func moveMonster(slot, dx, dy int) {
	a := &actors[slot]
	tile := a.tile
	x := int(a.x) + dx
	y := int(a.y) + dy
	ground := *tileAt(x, y)
	if tile&shapeMask == actorPirate {
		facing := 0
		switch {
		case dx == 0 && dy == -1:
			facing = faceNorth
		case dx == 1 && dy == 0:
			facing = faceEast
		case dx == 0 && dy == 1:
			facing = faceSouth
		case dx == -1 && dy == 0:
			facing = faceWest
		}
		a.tile = byte(facing) + actorPirate
		a.animTile = a.tile
	} else if tile != actor220 && tile != actor148 && tile != actor216 && tile != actor240 {
		switch ground {
		case mapTileSwamp, mapTile6, mapTile7, mapTile8, mapTileHills, mapTile31:
			if randomRange(0, 1) == 0 {
				return
			}
		case mapTileBrush, mapTile10, mapTile11, mapTile12, mapTile13, mapTile14, mapTile15:
			if randomRange(0, 2) != 2 {
				return
			}
		}
	}
	lastMonsterX, lastMonsterY = a.x, a.y
	a.x, a.y = byte(x), byte(y)
	viewChanged |= 2
	if *tileAt(x, y) == mapTileMoongate {
		a.tile, a.animTile = 0, 0
	}
}

// wander moves the monster one step a random way, if it can go there. (The original's
// loop always ends in its first pass.)
// C: MAINOUT_16fc
func wander(slot int) {
	x, y := int(actors[slot].x), int(actors[slot].y)
	switch randomRange(0, 3) {
	case 0:
		if monsterCanEnter(slot, x, y-1) {
			moveMonster(slot, 0, -1)
		}
	case 1:
		if monsterCanEnter(slot, x+1, y) {
			moveMonster(slot, 1, 0)
		}
	case 2:
		if monsterCanEnter(slot, x, y+1) {
			moveMonster(slot, 0, 1)
		}
	case 3:
		if monsterCanEnter(slot, x-1, y) {
			moveMonster(slot, -1, 0)
		}
	}
}

// signedOffset is a world coordinate difference across the wrap, -128..127.
func signedOffset(a, b byte) int { return int(int8(a - b)) }

// chase moves the monster one step towards the party, else wanders.
// C: MAINOUT_17d4
func chase(slot int) {
	x, y := int(actors[slot].x), int(actors[slot].y)
	sx, sy := 0, 0
	if d := signedOffset(actors[slot].x, mapX); d > 0 {
		sx = -1
	} else if d < 0 {
		sx = 1
	}
	if d := signedOffset(actors[slot].y, mapY); d > 0 {
		sy = -1
	} else if d < 0 {
		sy = 1
	}
	tryX := func() bool {
		if sx != 0 && monsterCanEnter(slot, x+sx, y) && notBack(x+sx, y) {
			moveMonster(slot, sx, 0)
			return true
		}
		return false
	}
	tryY := func() bool {
		if sy != 0 && monsterCanEnter(slot, x, y+sy) && notBack(x, y+sy) {
			moveMonster(slot, 0, sy)
			return true
		}
		return false
	}
	if randomRange(0, 1) == 1 {
		if tryX() || tryY() {
			return
		}
	} else if tryY() || tryX() {
		return
	}
	wander(slot)
}

// shadowlordChase is how many steps a Shadowlord follows the party.
const shadowlordChase = 20

// noWindWait is the pirate ship wait that has it sail every turn.
const noWindWait = 4

// moveOneMonster moves a monster by its kind: whirlpools drift every other turn,
// Shadowlords chase while they see the party, pirates sail with the wind.
// C: MAINOUT_198c
func moveOneMonster(slot int) {
	a := &actors[slot]
	tile := a.tile
	switch {
	case tile&shapeMask == actorWhirlpool:
		a.hull ^= 1
		if a.hull == 0 {
			return
		}
		if randomRange(0, 1) == 0 {
			wander(slot)
			return
		}
	case tile == actorShadowlord:
		if seesParty(slot) {
			steps := a.hull
			a.hull++
			if steps < shadowlordChase {
				chase(slot)
				return
			}
		}
	case tile&shapeMask == actorPirate:
		if windDirection == 0 {
			return
		}
		wait := pirateWind[tile-actorPirate][windDirection-1]
		if wait != noWindWait {
			a.target++
			if wait < int(a.target) {
				a.target = 0
				return
			}
		}
	}
	chase(slot)
}

// quickSkip and rideSkip skip the monsters' moves every other turn, under the Quickness
// spell and when riding or flying.
var (
	quickSkip bool // C: D_2c55
	rideSkip  bool // C: D_2c57
)

// moveMonstersOutdoors spawns and moves the monsters; it returns whether one acted against
// the party.
// C: MAINOUT_1a60
func moveMonstersOutdoors() int {
	acted := 0
	if spellIcon == 'T' {
		return 0
	}
	if spellIcon == 'Q' {
		quickSkip = !quickSkip
		if quickSkip {
			return 0
		}
	}
	if partyTile&pairMask == actorRidingHorse || partyTile&pairMask == actorFlyingCarpet {
		rideSkip = !rideSkip
		if rideSkip {
			return 0
		}
	}
	if encounterOdds() > randomRange(1, maxStat) {
		spawnMonster()
	}
	for slot := actorSlots - 1; slot > 0; slot-- {
		if roamsOutdoors(int(actors[slot].tile)) {
			acted += monsterActs(slot)
			if acted == 0 {
				moveOneMonster(slot)
			}
		}
	}
	for slot := actorSlots - 1; slot > 0; slot-- {
		a := &actors[slot]
		if roamsOutdoors(int(a.tile)) && (a.x-chunkX > worldWrapMask || a.y-chunkY > worldWrapMask) {
			setActor(0, 0, 0, 0, 0, 0, slot)
		}
	}
	return acted
}

// payTroll asks a toll of the trolls who caught the party, fighting them on a refusal.
// C: MAINOUT_1b3e
func payTroll() {
	printMessage(msgCaughtTheTrollsDemandASp)
	firstAbleMember()
	toll := -(int(party[moveX].str)*3 - trollToll)
	printNumber(toll, 2, ' ')
	printMessage(msgGpTollDostThouPay)
	var key int
	for {
		key = getChar()
		if key == 'Y' || key == 'N' {
			break
		}
	}
	printChar(key)
	printChar('\n')
	if key == 'Y' {
		gold -= int16(toll)
		statusChanged = true
		if gold >= 0 {
			return
		}
		gold += int16(toll)
	}
	slot := freeActorSlot()
	setActor(actor228, 0, mapX, mapY, 0, 0, slot)
	fightActor(slot)
}

// trollBridge may make the party sneak past trolls under a bridge.
// C: MAINOUT_1be8
func trollBridge() {
	if randomRange(0, 7) != 0 || partyTile != actorAvatar {
		return
	}
	updateFrame()
	printMessage(msgThouSpiethTrollsUnderTheBridge)
	sleepFrames(10)
	for i := range int(partySize) {
		m := &party[i]
		if m.status == statusDead || m.status == statusSleeping {
			continue
		}
		printString(m.nameString())
		printMessage(msgSneaksAcross)
		for range 3 {
			sleepFrames(5)
			printChar('.')
		}
		printMessage(msgNewlines2)
		if int(m.dex) < randomRange(1, maxStat) {
			payTroll()
			return
		}
	}
	printMessage(msgTrollsEvadedLine)
}
