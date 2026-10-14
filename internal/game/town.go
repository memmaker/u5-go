package game

// The settlements (town.c): entering a towne, dwelling, castle or keep, walking its
// levels, its people's arrests and attacks, and the settlement's main loop.

// Addresses of the settlement tables in the data segment.
const (
	shadowlordRowsAddress  = 5030  // C: D_13a6 (the row a shadowlord appears on, by settlement)
	levelOffsetsAddress    = 7706  // C: D_1e1a (each settlement's first level in its file)
	mapFilesAddress        = 9810  // C: D_2652 (the settlement files, by kind)
	drunkKeysAddress       = 10050 // C: D_2742 (the ways a drunk party staggers)
	harpsichordFreqAddress = 10054 // C: D_2746
	harpsichordTuneAddress = 10074 // C: D_275a
	shadowlordVicesAddress = 10204 // C: D_27dc
	hostilePeopleAddress   = 10438 // C: D_28c6 (32 bits a settlement: its hostile people)
	townVersionAddress     = 10409 // the version string Ctrl-V prints
	personBlocksAddress    = 13950 // C: D_367e (a bit per map tile, set when it stops a person)
)

const (
	settlementKinds   = 4
	shadowlordCount   = 3
	harpsichordKeys   = 10
	harpsichordLength = 13
	drunkKeyCount     = 4
	settlementLevel   = settlementSize * settlementSize // bytes of a settlement level
	settlementsOfKind = 8
)

// Settlement tables read from the data segment.
var (
	shadowlordRows   []byte   // C: D_13a6
	levelOffsets     []byte   // C: D_1e1a
	mapFiles         []string // C: D_2652
	drunkKeys        []byte   // C: D_2742
	harpsichordFreqs []uint16 // C: D_2746
	harpsichordTune  []byte   // C: D_275a
	shadowlordVices  []string // C: D_27dc
	hostilePeople    []uint32 // C: D_28c6
	personBlocks     []byte   // C: D_367e
)

// Settlement state outside the saved game.
var (
	harpsichordPlayed int  // C: D_2767 (the notes of the tune played right so far)
	freshArrival      bool // C: D_a9bc (skip the moongate check on the first turn)
)

// loadTownTables reads the settlement tables.
func loadTownTables() {
	shadowlordRows = segment.Bytes(shadowlordRowsAddress, settlementsOfKind)
	levelOffsets = segment.Bytes(levelOffsetsAddress, settlements)
	mapFiles = make([]string, settlementKinds)
	for i := range mapFiles {
		mapFiles[i] = message(int(segment.Word(mapFilesAddress + i*2)))
	}
	drunkKeys = segment.Bytes(drunkKeysAddress, drunkKeyCount)
	harpsichordFreqs = segment.Words(harpsichordFreqAddress, harpsichordKeys)
	harpsichordTune = segment.Bytes(harpsichordTuneAddress, harpsichordLength)
	shadowlordVices = make([]string, shadowlordCount)
	for i := range shadowlordVices {
		shadowlordVices[i] = message(int(segment.Word(shadowlordVicesAddress + i*2)))
	}
	words := segment.Words(hostilePeopleAddress, settlements*2)
	hostilePeople = make([]uint32, settlements)
	for i := range hostilePeople {
		hostilePeople[i] = uint32(words[i*2]) | uint32(words[i*2+1])<<16
	}
	personBlocks = segment.Bytes(personBlocksAddress, mapTileCount/bitsPerByte)
}

// Settlements with their own rules.
const (
	mapYew          = 4  // the jail
	mapCastle       = 17 // Lord British's castle, its harpsichord
	mapPalace       = 18 // Blackthorn's palace
	mapArarat       = 25 // the keep in the Underworld
	mapStonegate    = 29 // the shadowlords' keep
	harpsichordRoom = 2  // the castle level of the harpsichord
	jailX, jailY    = 25, 4
	wakeHour        = 8
)

// Tiles of the settlements.
const (
	lampPost        = mapTile135 // the tile above a lamp that lights at night
	lampToggle      = mapTile221 // the bits that turn a lamp on or off
	drawbridgeTile  = mapTile72  // a drawbridge, either half
	fireplaceMarker = mapTile42
	ladderOrStair   = mapTile134
	climbable1      = mapTile76
	climbable2      = mapTile202
	climbable3      = mapTile203
	horseFence1     = mapTile162
	horseFence2     = mapTile67
	harpsichordTile = 141
	secretWallX     = 17
	secretWallY     = 13
	secretWallBits  = 11
	bedTile1        = mapTile132
	bedTile2        = mapTile133
	shadowlordX     = 15
	dawnHour        = 5
	nightHour       = 20
	lampLitHour     = 20
	guardAlarm      = 128
)

// Person kinds (moveState) and moods.
const (
	moodFleeing    = 253
	moodHostile    = 254
	shadowlordKind = 252
	scheduleGuard  = 6
	scheduleHunt   = 7
	scheduleFlee   = 3
	scheduleKinds  = 3
	hostileHull    = 255
	waitingHull    = 30
	beggarKind     = 1
	lowPersonTile  = 47 // the people below this tile flee rather than hunt
)

// personKilled tells whether person who was killed in this settlement before.
// C: TOWN_0000_CheckNpcKilled
func personKilled(who int) bool {
	if npcTypes[who] == actorSandalwoodBox || npcTypes[who] >= actorWizard {
		return killedBits()&(1<<who) != 0
	}
	return false
}

// killedBits returns this settlement's killed people.
func killedBits() uint32 {
	at := (int(mapID) - 1) * 4
	return uint32(npcKilled[at]) | uint32(npcKilled[at+1])<<8 | uint32(npcKilled[at+2])<<16 | uint32(npcKilled[at+3])<<24
}

// markKilled records person who as killed, unless a guard or a person of no account.
// C: TOWN_0052_SetNpcKilled
func markKilled(who int) {
	if who < 0 || who >= npcSlots {
		return
	}
	kind := int(npcTypes[who] & shapeMask)
	if (kind != actorGuard && kind < actor128) || kind == actorShard {
		bits := killedBits() | 1<<who
		at := (int(mapID) - 1) * 4
		for i := range 4 {
			npcKilled[at+i] = byte(bits >> (8 * i))
		}
	}
}

// removePerson takes person who off the map.
// C: TOWN_00b0_DespawnNpc
func removePerson(who int) {
	n := &npcs[who]
	a := &actors[n.actorAt]
	a.tile, a.animTile, a.x, a.y, a.z, a.state, a.target = 0, 0, 0, 0, 0, 0, 0
	n.moveState, n.x, n.y, n.z, n.mood = 0, 0, 0, 0, 0
	s := &schedules[who]
	s.kind[0], s.kind[1], s.kind[2] = 0, 0, 0
	npcTypes[who] = 0
	viewChanged |= 2
}

// personShownBy returns the person the actor in slot shows, or -1.
// C: TOWN_011e
func personShownBy(slot int) int {
	for who := range npcSlots {
		if int(npcs[who].actorAt) == slot && npcTypes[who] != 0 && npcs[who].moveState != 0 {
			return who
		}
	}
	return -1
}

// switchLamps lights or puts out the lamps, and raises or lowers the drawbridges.
// C: TOWN_0170
func switchLamps() {
	for x := range settlementSize {
		for y := range settlementSize {
			if *tileAt(x, y) == lampPost {
				*tileAt(x, y+1) ^= lampToggle
			}
		}
	}
	if *tileAt(int(mapX), int(mapY))&pairMask != drawbridgeTile {
		for i := range int(changedCount) {
			if hour == dawnHour {
				*tileAt(int(changedX[i]), int(changedY[i])) = changedTile[i]
			} else {
				*tileAt(int(changedX[i]), int(changedY[i])) = mapTileWater3
			}
		}
	}
	viewChanged |= 2
}

// blightSettlement fells the trees and spoils the crops where a shadowlord haunts.
// C: TOWN_0212
func blightSettlement() {
	if shadowlordHere == noShadowlord {
		return
	}
	seedRandom(uint16(day))
	for y := range settlementSize {
		for x := range settlementSize {
			switch mapTiles[y*settlementSize+x] {
			case mapTileTree:
				if randomRange(0, 7) != 0 {
					mapTiles[y*settlementSize+x] = mapTileStump
				}
			case mapTileCrops:
				if randomRange(0, 7) != 0 {
					mapTiles[y*settlementSize+x] = mapTileCropsPicked
				}
			}
		}
	}
	viewChanged |= 2
	seedRandom(clockSeed())
}

const noShadowlord = 255

// summonShadowlord puts the shadowlord that haunts this settlement, if any, on the map.
// C: TOWN_02ae
func summonShadowlord() {
	shadowlordHere = noShadowlord
	if mapY != 4 {
		for i := range shadowlordCount {
			if shrineVisits[i] == mapID {
				shadowlordHere = byte(i)
				break
			}
		}
	}
	if shadowlordHere == noShadowlord {
		return
	}
	blightSettlement()
	for i := range actorSlots {
		if actors[i].tile == actorShadowlord {
			return
		}
	}
	slot := freeActorSlot()
	a := &actors[slot]
	who := npcSlots - 1
	for i := npcSlots - 1; i >= 0; i-- {
		if npcTypes[i] == 0 {
			who = i
			break
		}
	}
	n := &npcs[who]
	n.moveState = 1
	n.actorAt = uint16(slot)
	a.x = shadowlordX
	n.x = shadowlordX
	a.y = shadowlordRows[mapID-1]
	n.y = int16(a.y)
	n.z, a.state, a.hull, a.z, a.target = 0, 0, 0, 0, 0
	a.tile, a.animTile = actorShadowlord, actorShadowlord
	s := &schedules[who]
	for i := range scheduleTimes {
		s.time[i] = 0
	}
	for i := range schedulePoints {
		s.kind[i] = scheduleGuard
		s.x[i] = shadowlordX
		s.y[i] = a.y
		s.z[i] = 0
	}
	npcTypes[who] = shadowlordKind
}

// loadLevel reads the settlement's current level and sets it up.
// C: TOWN_0408
func loadLevel(placePeople bool) {
	showMapName()
	showSky()
	setWind(-1)
	doorOpen = 0
	chunkX, chunkY = 0, 0
	file := mapFiles[(mapID-1)>>3]
	index := int(levelOffsets[mapID-1]) + int(int8(level))
	readFile(file, mapTiles[:], index*settlementLevel)
	changedCount = 0
	fireX, fireY, secondFireX, secondFireY = -1, -1, -1, -1
	for x := range settlementSize {
		for y := range settlementSize {
			tile := *tileAt(x, y)
			if tile&pairMask == drawbridgeTile {
				changedX[changedCount] = byte(x)
				changedY[changedCount] = byte(y)
				changedTile[changedCount] = tile
				changedCount++
			}
			if tile == fireplaceMarker {
				if fireX == -1 {
					fireX, fireY = int16(x), int16(y)
				} else {
					secondFireX, secondFireY = int16(x), int16(y)
				}
			}
		}
	}
	if hour < dawnHour || hour > duskHour {
		switchLamps()
	}
	updateVisibility()
	passTime(0)
	blightSettlement()
	if placePeople {
		placeAllPeople()
	}
	viewChanged = 1
}

// takeStairs moves the party up or down when it steps onto stairs facing their way.
// C: TOWN_052e
func takeStairs(direction, tile int) {
	if tile&shapeMask != mapTileStair {
		return
	}
	updateFrame()
	switch tile - mapTileStair {
	case direction:
		level++
		printMessage(msgUpLine)
	case direction ^ 2:
		level--
		printMessage(msgDownLine)
	default:
		return
	}
	loadLevel(true)
}

// turnVehicle turns the party's horse, carpet or ship to direction (0 north, 1 east, 2
// south, 3 west).
// C: TOWN_057c
func turnVehicle(direction int) {
	switch partyTile & shapeMask {
	case actorHorse:
		printMessage(msgRideSp)
		if direction == 1 {
			partyTile = actorRidingHorse
		} else if direction == 3 {
			partyTile = actorRidingHorse + 1
		}
	case actorFlyingCarpet:
		printMessage(msgFlySp)
		if direction == 1 {
			partyTile = actorFlyingCarpet
		} else if direction == 3 {
			partyTile = actorFlyingCarpet + 1
		}
	case actorSkiff:
		printMessage(msgRowSp)
		partyTile = byte(direction) + partyTile&shapeMask
	case actorFrigate32, actorFrigate36:
		partyTile = byte(direction) + partyTile&shapeMask
	}
}

// stepInTown moves the party one step; it tells whether the party left the settlement.
// C: TOWN_0600
func stepInTown(key int) bool {
	if partyTile == actorAvatar || partyTile&pairMask == actorRidingHorse {
		footstep()
	}
	dx, dy := 0, 0
	leaving := false
	direction := 0
	switch key {
	case keyUp:
		dy--
		leaving = mapY < 1
		direction = 0
		turnVehicle(0)
		echoStep(msgNorthLine)
	case keyDown:
		dy++
		leaving = mapY > settlementSize-2
		direction = 2
		turnVehicle(2)
		echoStep(msgSouthLine)
	case keyRight:
		dx++
		leaving = mapX > settlementSize-2
		direction = 1
		turnVehicle(1)
		echoStep(msgEastLine)
	case keyLeft:
		dx--
		leaving = mapX < 1
		direction = 3
		turnVehicle(3)
		echoStep(msgWestLine)
	}
	free := true
	target := int(viewTile(dx+viewCentre, dy+viewCentre))
	other := actorAt(dx+int(mapX), dy+int(mapY), int(level))
	if other != 0 {
		free = false
		if partyTile >= actor48 || partyTile < actorFrigate32 {
			if (other >= actorFrigate36 && other < actorPirate) || other == actorCarpet || other&pairMask == actorHorse ||
				other == actorSleep || other == actorDead {
				free = true
			}
		} else if partyTile >= actorSkiff && other >= actorFrigate36 && other < actorSkiff {
			free = true
		}
	}
	if !free || !canEnter(int(partyTile), target) {
		printMessage(msgBlockedLine)
		tone(blockedFreq, blockedTime)
		clearKeyboard()
		return false
	}
	if !leaving {
		mapX = byte(int(mapX) + dx)
		mapY = byte(int(mapY) + dy)
		viewChanged = 1
		if partyTile&pairMask == actorRidingHorse {
			footstep()
		}
		takeStairs(direction, target)
		return false
	}
	printMessage(msgDostThouWishToLeaveSp)
	answer := 0
	for answer != 'Y' && answer != 'N' && answer != keyEscape {
		answer = getChar()
	}
	if answer != 'Y' {
		printMessage(msgNoLine)
		return false
	}
	printMessage(msgYesExitToLine)
	if mapID == mapArarat {
		printMessage(msgUnderworldLine)
		level = underworldLevel
	} else {
		printMessage(msgBritanniaLine)
		level = 0
	}
	doorOpen = 0
	mapX = placeX[mapID-1]
	mapY = placeY[mapID-1]
	mapID = 0
	return true
}

const (
	blockedFreq = 165
	blockedTime = 200
)

// alarmPerson sets person who on the party: guards and the shadowlord hunt it.
// C: TOWN_085e
func alarmPerson(who int) {
	kind := byte(scheduleGuard)
	if npcTypes[who] >= lowPersonTile {
		kind = scheduleHunt
	}
	s := &schedules[who]
	for i := range scheduleTimes {
		s.time[i] = 0
	}
	for i := range schedulePoints {
		s.kind[i] = kind
	}
}

// frightenPerson makes a townsperson flee the party.
// C: TOWN_08d4
func frightenPerson(who int) {
	moving := false
	for i := range scheduleTimes {
		if schedules[who].time[i] != 0 {
			moving = true
		}
	}
	if npcTypes[who] < actor116 && npcTypes[who] >= actorWizard && (npcs[who].mood == moodHostile || moving) {
		npcs[who].mood = moodFleeing
		for i := range schedulePoints {
			schedules[who].kind[i] = scheduleFlee
		}
	}
}

// raiseAlarm calls the guards on the party and scatters the townsfolk.
// C: TOWN_0958
func raiseAlarm() {
	for who := range npcSlots {
		if npcs[who].moveState == 0 {
			continue
		}
		kind := int(npcTypes[who])
		if kind == actorShadowlord || kind == actor216 || kind == actorGuard {
			alarmPerson(who)
		} else if randomRange(0, 255) < guardAlarm {
			frightenPerson(who)
		}
	}
}

// fightPerson fights person who on the combat map, then puts the settlement back.
// C: TOWN_09bc
func fightPerson(who int) {
	markKilled(who)
	fightActor(int(npcs[who].actorAt))
	removePerson(who)
	loadLevel(false)
	summonShadowlord()
}

// attackInTown is the Attack command in a settlement.
// C: TOWN_09e6_AttackCmd
func attackInTown() int {
	turned := 1
	printMessage(msgAttack)
	if *tileAt(int(mapX), int(mapY)) < mapTilePoison && partyTile != actorAvatar {
		printMessage(msgOnFootLine)
		return 0
	}
	if !selectDirection() {
		return turned
	}
	x := int(mapX) + int(moveX)
	y := int(mapY) + int(moveY)
	if *tileAt(x, y) == mapTileMirror {
		*tileAt(x, y) = mapTileMirrorBroken
		printMessage(msgBrokenLine)
		for limit := mirrorNoiseFrom; limit < mirrorNoiseTo; limit += mirrorNoiseStep {
			whiteNoise(mirrorNoiseRate, mirrorNoiseTime, limit)
		}
		viewChanged |= 2
		return turned
	}
	target := actorAt(x, y, int(level))
	nothing := true
	who := 0
	if target != 0 {
		who = personShownBy(int(moveX))
		if target >= actorWizard && (target < actor232 || target >= actor240) && target&shapeMask != actorShard {
			nothing = false
		}
	}
	if nothing {
		printMessage(msgNothingToAttackLine)
		return turned
	}
	if target < actor128 {
		decreaseByte(&karma, 5)
		raiseAlarm()
	} else if target&shapeMask == actor216 {
		raiseAlarm()
	}
	switch *tileAt(x, y) {
	case bedTile1, bedTile2, mapTileMirrorBroken, mapTileBed:
		if target == actorBlackthorn {
			printMessage(msgMissedLine)
			break
		}
		printMessage(msgMurderedLine)
		decreaseByte(&karma, 5)
		explodeAt(x, y)
		if who >= 0 {
			markKilled(who)
			removePerson(who)
		}
	default:
		if who >= 0 {
			markKilled(who)
			fightPerson(who)
		}
	}
	return turned
}

const (
	mirrorNoiseFrom = 2000
	mirrorNoiseTo   = 20000
	mirrorNoiseStep = 1000
	mirrorNoiseRate = 40
	mirrorNoiseTime = 120
)

// klimbInTown is the Klimb command in a settlement.
// C: TOWN_0b82_KlimbCmd
func klimbInTown() int {
	climbed := 0
	printMessage(msgKlimb)
	if partyTile&pairMask == actorRidingHorse {
		printMessage(msgOnFootLine2)
		return climbed
	}
	switch *tileAt(int(mapX), int(mapY)) {
	case mapTileLadderUp:
		takeStairs(0, mapTileStair)
		climbed = 1
	case ladderOrStair, mapTileLadderDown:
		takeStairs(2, mapTileStair)
		climbed = 1
	}
	if climbed == 0 && selectDirection() {
		switch *tileAt(int(mapX)+int(moveX), int(mapY)+int(moveY)) {
		case climbable1, climbable2, climbable3:
			mapX = byte(int(mapX) + int(moveX))
			mapY = byte(int(mapY) + int(moveY))
			viewChanged = 1
			updateFrame()
			climbed = 1
		default:
			printMessage(msgWhatLine)
		}
	}
	return climbed
}

// fencedIn tells whether x, y is a horse's fence.
// C: TOWN_0c4a
func fencedIn(x, y int) bool {
	tile := *tileAt(x, y)
	return tile == horseFence1 || tile == horseFence2
}

// wanderHorses moves the settlement's loose horses.
// C: TOWN_0c78
func wanderHorses() {
	for slot := range actorSlots {
		a := &actors[slot]
		tile := a.tile
		if tile&pairMask != actorHorse || a.z != level || randomRange(0, 1) != 0 {
			continue
		}
		x, y := int(a.x), int(a.y)
		if fencedIn(x, y+1) || fencedIn(x+1, y) || fencedIn(x, y-1) || fencedIn(x-1, y) {
			continue
		}
		if randomRange(0, 1) != 0 {
			step := randomRange(0, 1)*2 - 1
			x += step
			if step > 0 {
				tile = actorHorse
			} else {
				tile = actorHorse + 1
			}
		} else {
			y += randomRange(0, 1)*2 - 1
		}
		if x < settlementSize && y < settlementSize && x >= 0 && y >= 0 &&
			canEnter(actorHorse, int(*tileAt(x, y))) && topActorAt(x, y, int(level)) == 0 {
			a.tile, a.animTile = tile, tile
			a.x, a.y = byte(x), byte(y)
			viewChanged |= 2
		}
	}
}

// readTownKey redraws the frame when asked and reads the next key; a drunk party may
// stagger instead.
// C: TOWN_0dc4
func readTownKey(redraw bool) int {
	defer asking(promptCommand)()
	if redraw {
		updateFrame()
		if statusChanged {
			showStatus()
			statusChanged = false
		}
		if !reusePrompt() {
			printChar('\n')
			printOpenCaption()
		}
	}
	key := getChar()
	if drunkTurns != 0 && randomRange(0, 1) != 0 {
		raiseAlarm()
		drunkTurns--
		printMessage(msgHicLine)
		key = int(drunkKeys[randomRange(0, 3)])
	}
	return key
}

// Results of a command.
const (
	noTurn        = 0
	tookTurn      = 1
	attackedGuard = 2
	readAgain     = 3
)

// numberKey plays the harpsichord when the party stands before one, else picks the active
// member.
// C: TOWN_0e34
func numberKey(key int) int {
	result := readAgain
	if viewTile(viewCentre, viewCentre+1) != harpsichordTile {
		return setActiveMember(key)
	}
	note := key - '0'
	if soundOn {
		instrumentNote(int(harpsichordFreqs[note]), harpsichordDuration, harpsichordWidth, harpsichordStep)
	}
	switch {
	case int(harpsichordTune[harpsichordPlayed]) == note:
		harpsichordPlayed++
		if harpsichordPlayed == harpsichordLength {
			harpsichordPlayed = 0
			if mapID == mapCastle && level == harpsichordRoom {
				mapTiles[secretWallY*settlementSize+secretWallX] ^= secretWallBits
				shakeScreen()
				viewChanged = 1
			}
		}
	case harpsichordPlayed == 10 && note == 8:
		harpsichordPlayed = 3
	case harpsichordPlayed == 11 && note == 7:
		harpsichordPlayed = 2
	case note == int(harpsichordTune[0]):
		harpsichordPlayed = 1
	default:
		harpsichordPlayed = 0
	}
	return result
}

const (
	harpsichordDuration = 4000
	harpsichordWidth    = 20000
	harpsichordStep     = -4
)

// townHazards wakes sleepers now and then and works the trapdoors, poison and fire under
// the party.
// C: TOWN_0f02
func townHazards() {
	for i := range int(partySize) {
		if party[i].status == statusSleeping && randomRange(0, 15) == 15 {
			party[i].status = statusGood
		}
	}
	for again := true; again; {
		again = false
		tile := *tileAt(int(mapX), int(mapY))
		switch {
		case tile == mapTileTrapdoor && partyTile&pairMask != actorFlyingCarpet:
			printMessage(msgATrapdoorLine)
			saved := partyTile
			partyTile = 0
			updateFrame()
			hurtParty()
			partyTile = saved
			if mapID == mapStonegate {
				setPenColour(0)
				fillRect(viewLeft, viewLeft, viewRight, viewRight)
				sweepTone(trapdoorFrom, trapdoorTo, trapdoorStep, trapdoorTime)
				for i := range mapTiles {
					mapTiles[i] = mapTileLava
				}
				viewChanged = 1
				actors = [actorSlots]actor{}
				for i := range int(partySize) {
					party[i].hp = 0
					party[i].status = statusDead
					whiteNoise(trapdoorNoiseRate, trapdoorNoiseTime, trapdoorNoiseLimit)
					showStatus()
				}
			} else {
				level--
				loadLevel(true)
				again = true
			}
		case tile == mapTilePoison && partyTile == actorAvatar:
			for i := range int(partySize) {
				m := &party[i]
				if m.status != statusDead && m.status != statusPoisoned && int(m.dex) < randomRange(0, 29) {
					printMessage(msgPoisonedLine)
					m.status = statusPoisoned
					showStatus()
				}
			}
		case tile == mapTileFireplace || tile == mapTileLava:
			updateFrame()
			printMessage(msgBurningLine)
			hurtParty()
		}
	}
	passTurn()
}

const (
	trapdoorFrom       = 1000
	trapdoorTo         = 250
	trapdoorStep       = 40
	trapdoorTime       = 30000
	trapdoorNoiseRate  = 40
	trapdoorNoiseTime  = 3000
	trapdoorNoiseLimit = 500
)

// beginVermin has a hostile person send the party off.
// C: TOWN_10da
func beginVermin(who int) {
	printMessage(msgBegoneVerminLine)
	hurtParty()
	frightenPerson(who)
}

// followsSchedule tells (at random) whether a townsperson is swayed by a shadowlord.
// C: TOWN_10f2
func followsSchedule(who int) bool {
	swayed := false
	for i := range scheduleTimes {
		if schedules[who].time[i] != 0 {
			swayed = true
		}
	}
	// fix: the original looks at person 4's tile (its loop counter) rather than who's.
	kind := npcTypes[who]
	if kind < actorWizard || kind >= actor116 {
		swayed = false
	}
	if randomRange(0, 1) != 0 {
		swayed = false
	}
	return swayed
}

// swayTownsfolk turns the townsfolk hateful (hatred) or fearful (cowardice).
// C: TOWN_1156
func swayTownsfolk() {
	switch shadowlordHere {
	case 1:
		for who := range npcSlots {
			if followsSchedule(who) {
				alarmPerson(who)
				npcs[who].mood = moodHostile
			}
		}
	case 2:
		for who := range npcSlots {
			if followsSchedule(who) {
				frightenPerson(who)
				npcs[who].mood = moodFleeing
			}
		}
	}
}

// announceShadowlord tells of the shadowlord's vice in the air.
// C: TOWN_11b8
func announceShadowlord(which int) {
	printMessage(msgAnAirOfLine)
	printString(shadowlordVices[which])
	printMessage(msgDothSurroundTheeLine)
	pulse(shadowlordFreq, 1, shadowlordDuration, shadowlordWidth, 1)
}

const (
	shadowlordFreq     = 6600
	shadowlordDuration = 60000
	shadowlordWidth    = 2000
)

// enterSettlement sets up the settlement mapID; fresh when the party arrives from outside.
// C: TOWN_11f0_Entry
func enterSettlement(fresh bool) {
	freshArrival = true
	viewChanged = 1
	if fresh {
		for slot := 1; slot < actorSlots; slot++ {
			actors[slot].tile = 0
		}
		drunkTurns = 0
		mapShown = 1
		loadPeople()
		placePeopleAt(int(hour))
	}
	shadowlordHere = noShadowlord
	loadLevel(fresh)
	summonShadowlord()
	for who := range npcSlots {
		if personKilled(who) {
			removePerson(who)
		}
	}
	if mapID == mapStonegate && sceptre != 0 {
		removePerson(stonegateWarden)
	}
	showStatus()
	if firstAbleMember() < 0 {
		return
	}
	updateFrame()
	if mapID == mapStonegate {
		for i := shadowlordCount - 1; i >= 0; i-- {
			if shrineVisits[i] < 128 {
				announceShadowlord(i)
			}
		}
	} else if shadowlordHere != noShadowlord {
		announceShadowlord(int(shadowlordHere))
		swayTownsfolk()
	}
}

const stonegateWarden = 9

// arrest has a guard arrest the party; it tells whether the party fights instead.
// C: TOWN_12ae
func arrest() bool {
	if mapID == mapPalace {
		if firstAbleMember() >= 0 {
			captureByBlackthorn()
			enterSettlement(true)
		}
		return false
	}
	printMessage(msgThouArtUnderArrestLine)
	printMessage(msgWiltThouComeQuietly)
	answer := 0
	for answer != 'N' && answer != 'Y' {
		answer = getChar() & lowByteMask
	}
	if answer == 'N' {
		printMessage(msgNoThenDefendThyselfRogueLine)
		raiseAlarm()
		return true
	}
	printMessage(msgYesTheGuardStrikesTheeUnconscious)
	setPenColour(0)
	printMessage(msgThouDostAwakenToLine)
	mapID = mapYew
	mapX, mapY = jailX, jailY
	viewChanged = 1
	for hour != wakeHour {
		passTime(20)
	}
	level, keys = 0, 0
	enterSettlement(true)
	return false
}

// meetPerson handles the person who came up to the party this turn.
// C: TOWN_1352
func meetPerson(wasAttack int) {
	fight := false
	n := &npcs[npcTurn]
	a := &actors[n.actorAt]
	updateFrame()
	switch {
	case townMode == townModeAttack:
		switch {
		case n.mood == moodHostile:
			printChar('\n')
			beginVermin(int(npcTurn))
		case a.tile != actorGuard:
			fight = true
		default:
			fight = arrest()
		}
	case wasAttack == 0:
		if n.mood != 0 && talkToPerson(int(npcTurn)) != 0 {
			fight = arrest()
		}
	default:
		fight = arrest()
	}
	if !fight {
		return
	}
	if actors[npcs[npcTurn].actorAt].tile >= actorWizard {
		printMessage(msgAttackedLine)
		fightPerson(int(npcTurn))
	} else {
		removePerson(int(npcTurn))
	}
}

// What a person who reached the party wants (townMode).
const (
	townModeAttack = 'a'
	townModeTalk   = 't'
)

// townMainLoop plays the settlement until the party leaves it.
// C: TOWN_141e_MainLoop
func townMainLoop() {
	horseSkip, quickSkip := false, false
	for left := false; !left; {
		turned := tookTurn
		key := 0
		switch firstAbleMember() {
		case onlySleepers:
			printChar('\n')
			printOpenCaption()
			printMessage(msgZzzzzzLine)
		case noneAble:
			partyDied()
			turned = noTurn
		default:
			if freshArrival {
				freshArrival = false
			} else if enterMoongate() != 0 {
				shrine()
			}
			if mapID == 0 {
				return
			}
			for redraw := true; ; redraw = false {
				key = readTownKey(redraw)
				switch {
				case key < ' ':
					switch key {
					case keyCtrlK:
						printNumber(int(karma), 1, ' ')
						printChar('\n')
						turned = noTurn
					case keyCtrlE:
						printMessage(msgExitToDosSp)
						key = getChar()
						if key == 'Y' {
							exitGame()
						}
						printMessage(msgNLine)
					case keyCtrlV:
						printMessage(townVersionAddress)
						printChar('\n')
						turned = noTurn
					case keyCtrlS:
						printMessage(msgSoundSp)
						if soundOn {
							printMessage(msgOffLine)
						} else {
							printMessage(msgOnLine)
						}
						soundOn = !soundOn
						turned = noTurn
					case keyLeft, keyRight, keyUp, keyDown:
						left = stepInTown(key)
						turned = tookTurn
						if left {
							turned = noTurn
						}
					default:
						printMessage(msgWhatLine)
						turned = noTurn
					}
				case key < '0' || key > '9':
					turned = doCommand(key)
				default:
					turned = numberKey(key)
				}
				if turned != readAgain {
					break
				}
			}
		}
		if mapID == 0 {
			left = true
			continue
		}
		if turned == noTurn || firstAbleMember() == noneAble {
			continue
		}
		before := hour
		passTime(1)
		if hour != before && (hour == lampLitHour || hour == dawnHour) {
			switchLamps()
		}
		townHazards()
		if doorOpen != 0 {
			doorTurns--
			if doorTurns == 0 {
				setTile(doorOpen, doorX, doorY)
			}
		}
		actors[0].x, actors[0].y, actors[0].z = mapX, mapY, level
		people := partyTile < actorRidingHorse || partyTile >= actorCircle || key == ' '
		if !people {
			horseSkip = !horseSkip
			people = !horseSkip
		}
		if people && spellIcon != 'T' {
			if spellIcon == 'Q' {
				quickSkip = !quickSkip
				people = !quickSkip
			}
			if people {
				wanderHorses()
				if turned < attackedGuard {
					movePeople(int(hour))
				}
				if npcTurn != 0 || turned == attackedGuard {
					meetPerson(turned - 1)
				}
			}
		}
	}
}

// placeAllPeople puts the people of the current level on the map.
// C: TOWN_1694
func placeAllPeople() {
	for slot := 1; slot < actorSlots; slot++ {
		setActor(0, 0, 0, 0, 0, 0, slot)
		npcs[slot].actorAt = 0
	}
	for who := 1; who < npcSlots; who++ {
		if npcTypes[who] == 0 {
			continue
		}
		place := schedulePlace(who, int(hour))
		s := &schedules[who]
		placePerson(who, int(s.x[place]), int(s.y[place]), int(s.z[place]))
		npcs[who].moveState = 1
		npcs[who].place = uint16(place)
		movementPos[who] = -1
	}
}

// placePerson moves person who to x, y, z, showing or hiding them as the level allows.
// C: TOWN_1726
func placePerson(who, x, y, z int) {
	if npcTypes[who] == 0 {
		return
	}
	n := &npcs[who]
	if z != int(level) && n.actorAt != 0 {
		actors[n.actorAt].tile = 0
		n.actorAt = 0
	}
	if z == int(level) && n.actorAt == 0 {
		slot := freeActorSlot()
		hull := byte(0)
		switch {
		case npcTypes[who] == beggarKind:
			hull = waitingHull
		case hostilePeople[mapID-1]&(1<<who) != 0:
			hull = hostileHull
		}
		setActor(npcTypes[who], npcTypes[who], byte(x), byte(y), byte(z), hull, slot)
		actors[slot].state = 0
		n.actorAt = uint16(slot)
	}
	if z == int(level) && n.actorAt != 0 {
		a := &actors[n.actorAt]
		a.x, a.y, a.z = byte(x), byte(y), byte(z)
	}
	n.x, n.y, n.z = int16(x), int16(y), uint16(z)
	n.moveState = 1
}
