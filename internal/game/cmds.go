package game

// The general commands (cmds.c): camping, sleeping in a bed, boarding and leaving
// vehicles, firing cannons, torches, the party order, yelling words of power, pushing,
// escaping combat, mixing reagents and climbing with a grapple.

// Addresses of the commands' tables in the data segment.
const (
	ambushFoesAddress    = 5940  // C: D_1734 (the monsters that ambush a camp)
	reagentNamesAddress  = 6610  // C: D_19d2
	spellReagentsAddress = 7360  // C: D_1cc0 (each spell's reagents, a bit each)
	mantrasAddress       = 8030  // C: D_1f5e
	shadowlordNamesAddr  = 17482 // C: D_444a
	wordsOfPowerAddress  = 17666 // C: D_4502
	sealedEntrances      = 17682 // C: D_4512 (the tile of each dungeon's entrance)
)

const (
	ambushFoeCount = 8
	dungeonCount   = 8
)

// Command tables read from the data segment.
var (
	ambushFoes      []byte   // C: D_1734
	reagentNames    []string // C: D_19d2
	spellReagents   []byte   // C: D_1cc0
	mantras         []string // C: D_1f5e
	shadowlordNames []string // C: D_444a
	wordsOfPower    []string // C: D_4502
	entranceTiles   []byte   // C: D_4512
)

// loadCommandTables reads the commands' tables.
func loadCommandTables() {
	ambushFoes = segment.Bytes(ambushFoesAddress, ambushFoeCount)
	reagentNames = messages(reagentNamesAddress, reagentTypes)
	spellReagents = segment.Bytes(spellReagentsAddress, spellCount)
	mantras = messages(mantrasAddress, shrineCount)
	shadowlordNames = messages(shadowlordNamesAddr, shadowlordCount)
	wordsOfPower = messages(wordsOfPowerAddress, dungeonCount)
	entranceTiles = segment.Bytes(sealedEntrances, dungeonCount)
}

// messages reads count messages from a table of their addresses.
func messages(table, count int) []string {
	list := make([]string, count)
	for i := range list {
		list[i] = message(int(segment.Word(table + i*2)))
	}
	return list
}

// Flags of a camp (the kind of special map).
const (
	campInDungeon = 2
	campNoGhost   = 130
	bardSong      = 52
	restHours     = 5
	restCooldownH = 14
	apparitionPct = 25
	maxRestHP     = 63
	bardPlaying   = actor95
)

// campHere camps hours with member guard on watch (-1 none); it tells whether the camp
// was ambushed.
// C: CMDS_0000
func campHere(kind, guard, hours int) bool {
	lastHour := int(hour)
	bard := -1
	ambush := -1
	spellIcon, protection = 0, 0
	active := activeMember
	activeMember = noMember
	if kind&campInDungeon != 0 {
		enterDungeonRoom(int(dungeonCell(int(mapX), int(mapY), int(level))), 0)
	} else {
		setUpFighters()
	}
	combatant = noCombatant
	wakeAt := hours + int(hour)
	if wakeAt >= hoursPerDay {
		wakeAt -= hoursPerDay
	}
	if guard > restHours || guard < -1 {
		guard = -1
	}
	var poisoned [8]int
	var watch *fighter
	for i := range int(partySize) {
		slot := 0
		for slot = 0; slot < combatantSlots; slot++ {
			if i == int(combatants[slot].entity) {
				watch = &combatants[slot]
				break
			}
		}
		switch party[i].status {
		case statusPoisoned:
			poisoned[i]++
			fallthrough
		case statusGood, statusSleeping:
			if i != guard {
				sleepUnlessPoisoned(slot)
			}
			if party[i].class == classBard && i == guard && soundOn {
				bard = slot
			}
		}
	}
	updateFrame()
	if bard != -1 {
		a := &actors[combatants[bard].actorSlot]
		tile, anim := a.tile, a.animTile
		a.tile, a.animTile = bardPlaying, bardPlaying
		songNote = 1
		sleepFrames(bardSong)
		a.tile, a.animTile = tile, anim
		if int(combatants[bard].entity) != guard {
			sleepUnlessPoisoned(bard)
		}
		updateFrame()
	}
	printMessage(msgZzzzzzLine2)
	var guardTile byte
	if guard != -1 {
		for slot := range combatantSlots {
			if int(combatants[slot].entity) == guard {
				watch = &combatants[slot]
				break
			}
		}
		guardTile = actors[watch.actorSlot].tile
	}
	for wakeAt != int(hour) {
		if hour >= hoursPerDay {
			hour -= hoursPerDay
		}
		updateFrame()
		regenerate()
		showStatus()
		if lastHour != int(hour) && randomRange(0, 63) == 0 {
			seedRandom(clockSeed())
			ambush = int(ambushFoes[randomRange(0, 7)])
			printMessage(msgAmbushedLine)
			if guard > -1 {
				for slot := range maxPartySize {
					member := int(combatants[slot].entity)
					if member < int(partySize) {
						wakeFighter(slot)
						if poisoned[member] != 0 {
							party[member].status = statusPoisoned
						} else {
							party[member].status = statusGood
						}
					}
				}
			}
			showStatus()
			if kind&campInDungeon != 0 {
				enterDungeonRoom(int(dungeonCell(int(mapX), int(mapY), int(level))), 1)
			} else {
				setUpCombat(kind, ambush)
			}
			break
		}
		lastHour = int(hour)
		passTime(5)
		waitTicks(1)
		if savedMapID < firstDungeonMap {
			mapID = savedMapID
			showSky()
			mapID = combatMapID
		}
		if guard != -1 && randomRange(0, 3) == 2 {
			moveX, moveY = int16(watch.x), int16(watch.y)
			switch randomRange(0, 3) {
			case 0:
				moveY--
			case 1:
				moveY++
			case 2:
				moveX++
			case 3:
				moveX--
			}
			if onCombatMap(int(moveX), int(moveY)) && fighterMayStand(int(guardTile), int(moveX), int(moveY)) {
				watch.x = byte(moveX)
				actors[watch.actorSlot].x = byte(moveX)
				watch.y = byte(moveY)
				actors[watch.actorSlot].y = byte(moveY)
				updateFrame()
			}
		}
	}
	if ambush > -1 {
		return true
	}
	if restCooldown < 1 && hours > restHours {
		printMessage(msgPartyRestedLine)
		for i := range int(partySize) {
			m := &party[i]
			if poisoned[i] != 0 || m.status == statusDead || i == guard {
				continue
			}
			m.hp += int16(randomRange(1, maxRestHP))
			if m.hp > m.maxHP {
				m.hp = m.maxHP
			}
			switch m.class {
			case classAvatar, classMage:
				m.magic = m.intel
			case classBard:
				m.magic = m.intel >> 1
			}
		}
		if kind&campNoGhost == 0 && randomRange(0, 99) < apparitionPct {
			apparitionMonth = month
			apparition()
		}
		restCooldown = restCooldownH
	} else {
		printMessage(msgNoEffectLine)
	}
	for i := range int(partySize) {
		if party[i].status == statusSleeping {
			party[i].status = statusGood
		}
	}
	activeMember = active
	showStatus()
	return false
}

const (
	bedHourLimit = 23
	settleTurns  = 16
)

// sleepInBed sleeps the given hours in a settlement's bed.
// C: CMDS_0552_HoleUpCmd
func sleepInBed() {
	printMessage(msgForHowManyHoursSp)
	key := byte(0)
	for {
		key = byte(getChar())
		if key == keySpace || (key >= '0' && key <= '9') {
			break
		}
	}
	if key == keySpace || key == '0' {
		return
	}
	printChar(int(key))
	printChar('\n')
	wakeAt := int(hour) + int(key) - '0'
	if wakeAt > bedHourLimit {
		wakeAt -= bedHourLimit
	}
	for range settleTurns {
		movePeople(int(hour))
		updateFrame()
		if townMode == townModeAttack {
			return
		}
	}
	active := activeMember
	for i := range int(partySize) {
		if party[i].status == statusGood {
			party[i].status = statusSleeping
		}
	}
	showStatus()
	printMessage(msgZzzzzzzLine)
	setPenColour(0)
	fillRect(viewLeft, viewLeft, viewRight, viewRight)
	last := hour
	for wakeAt != int(hour) {
		passTime(10)
		if last != hour && (hour == lampLitHour || hour == dawnHour) {
			switchLamps()
		}
		last = hour
		showSky()
		passTurn()
		showStatus()
		placeAllPeople()
		if actorAt(int(mapX), int(mapY), int(level)) != 0 {
			wakeAt = -1
			break
		}
		waitTicks(1)
	}
	if wakeAt == -1 {
		printMessage(msgThrownOutOfBedLine)
	}
	for i := range int(partySize) {
		if party[i].status == statusSleeping {
			party[i].status = statusGood
		}
	}
	activeMember = active
	mapX++
	viewChanged = 1
	actors[0].x++
	showStatus()
	updateFrame()
}

// onFoot tells whether the party walks (or carries a carpet), else says it must.
// C: CMDS_06ee
func onFoot() bool {
	if partyTile != actorAvatar && partyTile != actor29 {
		printMessage(msgOnFootLine4)
		return false
	}
	return true
}

// canLeaveShip tells whether the party may leave its ship as it is.
// C: CMDS_070c
func canLeaveShip() bool {
	switch partyTile {
	case actorFlyingCarpet, actorFlyingCarpet + 1, actorAvatar, actor29,
		actorSkiff, actorSkiff + 1, actorSkiff + 2, actorSkiff + 3:
		return true
	}
	printMessage(msgOnFootLine4)
	return false
}

// landNearby tells whether there is land beside the party.
// C: CMDS_073e
func landNearby() bool {
	return landAt(4, viewCentre) || landAt(6, viewCentre) || landAt(viewCentre, 6) || landAt(viewCentre, 4)
}

// landAt tells whether view row, column can be walked or holds something to stand on.
// C: CMDS_0788
func landAt(row, column int) bool {
	if tile := int(viewTile(column, row)); tile != 0 {
		return canEnter(actorAvatar, tile)
	}
	tile := int(actorAtView(column, row))
	if tile == actorCarpet {
		return true
	}
	tile &= shapeMask
	return tile == actorAvatar || tile == actorFrigate36 || tile == actorHorse || tile == actorSkiff
}

const (
	shipDamaged       = 10
	firstDungeonInner = 33
	lastDungeonInner  = 40
)

// boardCommand boards the horse, carpet, skiff or ship the party stands on.
// C: CMDS_07f6_BoardCmd
func boardCommand() int {
	if mapID > settlements && mapID < firstDungeonInner+dungeonCount {
		printMessage(msgNotHereLine)
		return 1
	}
	tile := byte(actorAt(int(mapX), int(mapY), int(level)))
	slot := int(moveX)
	switch {
	case tile&pairMask == actorHorse:
		if mapID != 0 {
			who := personShownBy(slot)
			if who != -1 && npcs[who].mood != 0 {
				printMessage(msgNayLine)
				return 1
			}
		}
		if !onFoot() {
			return 1
		}
		printMessage(msgHorseLine)
		partyTile = tile + 2
	case tile == actorCarpet:
		if !onFoot() {
			return 1
		}
		printMessage(msgCarpetLine)
		partyTile = actorFlyingCarpet
	case tile&shapeMask == actorSkiff:
		if !onFoot() {
			return 1
		}
		printMessage(msgSkiffLine)
		partyTile = tile
	case tile&shapeMask == actorFrigate36:
		if !canLeaveShip() {
			return 1
		}
		printMessage(msgShipLine)
		hull := actors[slot].hull
		if hull < shipDamaged {
			printMessage(msgDangerShipBadlyDamagedLine)
		}
		actors[0].hull = hull
		skiffs := actors[slot].target
		if partyTile&pairMask == actorFlyingCarpet {
			magicCarpets++
		}
		if partyTile&shapeMask == actorSkiff {
			skiffs++
		}
		if skiffs == 0 {
			printMessage(msgWarningNoSkiffsOnBoardLine)
		}
		partyTile = tile
		actors[0].target = skiffs
		statusChanged = true
	default:
		printMessage(msgWhatLine)
		return 0
	}
	setActor(0, 0, 0, 0, 0, 0, slot)
	viewChanged |= 2
	return 1
}

const (
	broadsideRange = 3
	cannonTarget   = 236
	maxShipDamage  = 20
	hullSunk       = 127
)

// fireBroadside fires the ship's cannons sideways.
// C: CMDS_0962
func fireBroadside() {
	if partyTile < actorFrigate32 || partyTile > actorFrigate32End {
		printMessage(msgWhatLine)
		return
	}
	if !selectDirection() {
		return
	}
	dx, dy := int(moveX), int(moveY)
	if (dx == 0 && partyTile&1 == 0) || (dx != 0 && partyTile&1 != 0) {
		printMessage(msgFireBroadsidesOnlyLine)
		return
	}
	x, y := int(mapX), int(mapY)
	sweepTone(cannonFrom, cannonTo, cannonStep, cannonTime)
	for range broadsideRange {
		x += dx
		y += dy
		target := actorAt(x, y, int(level))
		if roamsOutdoors(target) && target&shapeMask != cannonTarget {
			slot := int(moveX)
			tx := int(actors[slot].x) - int(mapX) + viewCentre
			ty := int(actors[slot].y) - int(mapY) + viewCentre
			if cannonFire(viewCentre, viewCentre, tx, ty, 1) != 0 {
				updateFrame()
				explodeAt(x, y)
				actors[slot].hull -= byte(randomRange(1, maxShipDamage))
				if actors[slot].hull > hullSunk {
					setActor(0, 0, 0, 0, 0, 0, slot)
					viewChanged |= 2
				}
			}
			return
		}
	}
	cannonFire(viewCentre, viewCentre, dx*broadsideRange+viewCentre, dy*broadsideRange+viewCentre, 1)
}

const (
	cannonFrom     = 1000
	cannonTo       = 200
	cannonStep     = 5
	cannonTime     = 300
	cannonReach    = 5
	cannonShape    = mapTileCannon180
	rubbleTile     = mapTile68
	cannonKarma    = 5
	firstFolk      = 28
	monsterShapes  = 120
	monsterMask    = 248
	wispShape      = 47
	shotHorse      = 16
	shotHorseRight = 17
)

// fireCommand fires a cannon beside the party, or the ship's broadsides.
// C: CMDS_0aea_FireCmd
func fireCommand() {
	if mapID > settlements && mapID < firstDungeonInner+dungeonCount {
		printMessage(msgWhatLine)
		return
	}
	if mapID == 0 {
		fireBroadside()
		return
	}
	setTile(doorOpen, doorX, doorY)
	var cannon byte
	var dx, dy int
	switch {
	case viewTile(viewCentre, viewCentre-1)&shapeMask == cannonShape:
		cannon, dx, dy = viewTile(viewCentre, viewCentre-1), 0, -1
	case viewTile(viewCentre+1, viewCentre)&shapeMask == cannonShape:
		cannon, dx, dy = viewTile(viewCentre+1, viewCentre), 1, 0
	case viewTile(viewCentre, viewCentre+1)&shapeMask == cannonShape:
		cannon, dx, dy = viewTile(viewCentre, viewCentre+1), 0, 1
	case viewTile(viewCentre-1, viewCentre)&shapeMask == cannonShape:
		cannon, dx, dy = viewTile(viewCentre-1, viewCentre), -1, 0
	default:
		printMessage(msgWhatLine)
		return
	}
	x, y := dx+int(mapX), dy+int(mapY)
	fromX, fromY := dx+viewCentre, dy+viewCentre
	toX, toY := fromX, fromY
	switch cannon & 3 {
	case 0:
		dx, dy = 0, -1
	case 1:
		dx, dy = 1, 0
	case 2:
		dx, dy = 0, 1
	case 3:
		dx, dy = -1, 0
	}
	printMessage(msgBooomLine)
	sweepTone(cannonFrom, cannonTo, cannonStep, cannonTime)
	raiseAlarm()
	door, hit := false, false
	victim := 0
	for steps := cannonReach; !door && !hit; {
		steps--
		if steps <= 0 {
			break
		}
		x += dx
		y += dy
		toX += dx
		toY += dy
		target := topActorAt(x, y, int(level))
		if target == 0 {
			switch *tileAt(x, y) {
			case mapTile151, mapTile152, mapTile153, mapTileDoor184, mapTileDoor185, mapTileDoor186, mapTileDoor187:
				door = true
			}
			// the original's test of the actor below always holds for one there
		} else if target >= firstFolk || target == shotHorse || target == shotHorseRight {
			hit = true
			victim = int(moveX)
		}
	}
	cannonFire(fromX, fromY, toX, toY, 1)
	if door || hit {
		explodeAt(x, y)
	}
	if door {
		printMessage(msgDoorDestroyedLine)
		*tileAt(x, y) = rubbleTile
		viewChanged = 1
		doorOpen = 0
	}
	if hit && victim != 0 {
		setActor(0, 0, 0, 0, 0, 0, victim)
		viewChanged |= 2
		if karma > cannonKarma {
			karma -= cannonKarma
		} else {
			karma = 0
		}
		who := personShownBy(victim)
		if who == -1 {
			return
		}
		markKilled(who)
		removePerson(who)
	}
	if hit && victim == 0 {
		hurtParty()
	}
}

const (
	torchInDungeon = 112
	torchOutside   = 240
)

// igniteTorch lights a torch.
// C: CMDS_0d98_IgniteTorchCmd
func igniteTorch() {
	if torches == 0 {
		printMessage(msgNoneOwnedLine)
		return
	}
	torches--
	if mapID > settlements && mapID < firstDungeonInner+dungeonCount {
		increaseByte(&torchLight, randomRange(0, 15)+torchInDungeon, lowByteMask)
	} else {
		torchLight = torchOutside
	}
}

// newOrderCommand swaps two party members (not the Avatar).
// C: CMDS_0ddc_NewOrderCmd
func newOrderCommand() {
	printMessage(msgSwapSp)
	first := selectOne()
	if first == -1 {
		printMessage(msgNobodyLine)
		return
	}
	printString(party[first].nameString())
	if first == 0 {
		printMessage(msgNewlines2)
		printString(party[0].nameString())
		printMessage(msgMustLeadLine)
		return
	}
	printMessage(msgWithSp)
	second := selectOne()
	if second == -1 {
		printMessage(msgNobodyLine)
		return
	}
	printString(party[second].nameString())
	if second == 0 {
		printMessage(msgNewlines2)
		printString(party[0].nameString())
		printMessage(msgMustLeadLine)
		return
	}
	printMessage(msgChar33Newline)
	party[first], party[second] = party[second], party[first]
	statusChanged = true
}

// xitCommand leaves the vehicle the party is in.
// C: CMDS_0eb4_XitCmd
func xitCommand() {
	// the original's test for the dungeons can never hold (sic)
	tile := *tileAt(int(mapX), int(mapY))
	var left byte
	skiffs := byte(0)
	switch partyTile & shapeMask {
	case actorAvatar:
		printMessage(msgWhatLine2)
		return
	case actorFrigate32:
		printMessage(msgUnderSailLine)
		return
	case actorFlyingCarpet:
		if landNearby() || canEnter(actorAvatar, int(tile)) {
			printMessage(msgCarpetLine2)
			left = actorCarpet
			partyTile = actorAvatar
			break
		}
		if !landNearby() {
			printMessage(msgNoLandNearbyLine)
		} else {
			printMessage(msgNotHereLine)
		}
		return
	case actorHorse:
		printMessage(msgHorseLine2)
		left = partyTile - 2
		partyTile = actorAvatar
	case actorSkiff:
		switch {
		case !landNearby():
			printMessage(msgNoLandNearbyLine)
			return
		case tile&pairMask == mapTile106:
			printMessage(msgNotHereLine)
			return
		}
		printMessage(msgSkiffLine2)
		left = partyTile
		partyTile = actorAvatar
	case actorFrigate36:
		printMessage(msgShipLine2)
		switch {
		case landNearby():
			left = partyTile
			partyTile = actorAvatar
			skiffs = actors[0].target
		case actors[0].target != 0:
			left = partyTile
			partyTile += 4
			skiffs = actors[0].target - 1
		case magicCarpets != 0:
			magicCarpets--
			left = partyTile
			partyTile = actorFlyingCarpet
			skiffs = actors[0].target
		default:
			printMessage(msgNoSkiffsOnBoardLine)
			return
		}
	}
	slot := freeActorSlot()
	setActor(left, left, mapX, mapY, level, actors[0].hull, slot)
	// fix: the original leaves a stale value as the skiffs of what is left but a ship.
	actors[slot].target = skiffs
	showStatus()
}

const (
	shadowlordKeepFirst = 30
	shadowlordKeepLast  = 32
	shadowlordRise      = 2
	summonFreq          = 10400
	summonDuration      = 30000
	summonWidth         = 2000
	summonStep          = 2
	summonFlash         = 508
	summonFlashX        = 5
	summonFlashY        = 3
)

// summonShadowlordByName calls a shadowlord into its flame by its name.
// C: CMDS_1030
func summonShadowlordByName(words []byte) int {
	if mapID < shadowlordKeepFirst || mapID > shadowlordKeepLast {
		printMessage(msgNoEffectLine2)
		return 1
	}
	which := 0
	for which = 0; which < shadowlordCount; which++ {
		if findWord([]byte(shadowlordNames[which]+"\x00"), words) > -1 {
			break
		}
	}
	if which == shadowlordCount || mapY < shadowlordRise || shrineVisits[which] == lowByteMask {
		printMessage(msgNoEffectLine2)
		return 1
	}
	for slot := range actorSlots {
		if actors[slot].tile == actorShadowlord {
			printMessage(msgNoEffectLine2)
			return 1
		}
	}
	lastSpell = byte(which)
	slot := freeActorSlot()
	y := mapY - shadowlordRise
	setActor(actorShadowlord, actorShadowlord, mapX, y, level, 0, slot)
	who := npcSlots - 1
	for who = npcSlots - 1; who >= 0; who-- {
		if npcTypes[who] == 0 {
			break
		}
	}
	n := &npcs[who]
	n.moveState = 1
	n.actorAt = uint16(slot)
	n.x, n.y, n.z = int16(mapX), int16(y), uint16(level)
	s := &schedules[who]
	for i := range scheduleTimes {
		s.time[i] = 0
	}
	for i := range schedulePoints {
		s.kind[i] = scheduleGuard
		s.x[i], s.y[i], s.z[i] = mapX, y, level
	}
	npcTypes[who] = shadowlordKind
	printMessage(msgAShadowlordAppearsLine)
	pulse(summonFreq, 1, summonDuration, summonWidth, summonStep)
	actors[slot].tile, actors[slot].animTile = actorCircle, actorCircle
	revealTile(summonFlash, summonFlashX, summonFlashY)
	actors[slot].tile, actors[slot].animTile = actorShadowlord, actorShadowlord
	return 0
}

const shrineRuinedBit = 127

// meditate asks the virtue and mantra at a ruined shrine, restoring it when right.
// C: CMDS_1202
func meditate(shrine, x, y int) {
	right := true
	printMessage(msgUponWhatVirtueDostThouMeditate)
	var answer [talkInputSize]byte
	copy(answer[:], readString(talkInputLen))
	if findWord([]byte(shrineNames[shrine]+"\x00"), answer[:]) == -1 {
		right = false
	}
	for range 3 {
		printMessage(msgMantra)
		answer = [talkInputSize]byte{}
		copy(answer[:], readString(talkInputLen))
		if findWord([]byte(mantras[shrine]+"\x00"), answer[:]) == -1 {
			right = false
		}
	}
	if right && int(shrineX[shrine]) == x && int(shrineY[shrine]) == y {
		shrineRunes[shrine] &= shrineRuinedBit
		printMessage(msgTheShrineIsRestoredLine)
		shakeScreen()
		*tileAt(x, y) = mapTileShrine
		viewChanged |= 2
		return
	}
	printChar('\n')
}

const (
	sealedTile      = mapTile223
	dungeonPlaceOff = 32
	sealBit         = 128
)

// speakWordOfPower speaks a word: a dungeon's word unseals (or seals) its entrance, and
// beside a ruined shrine one meditates.
// C: CMDS_12c8
func speakWordOfPower(words []byte) {
	spoken := false
	for i := 0; i < dungeonCount; i++ {
		if findWord([]byte(wordsOfPower[i]+"\x00"), words) <= -1 {
			continue
		}
		printMessage(msgAWordOfPowerIsUttered)
		shakeScreen()
		var dx, dy int
		var tile byte
		matches := func(t byte) bool { return entranceTiles[i] == t || t == sealedTile || t == mapTileRuins }
		switch {
		case matches(viewTile(4, viewCentre)):
			tile, dx, dy = viewTile(4, viewCentre), -1, 0
		case matches(viewTile(viewCentre, 6)):
			tile, dx, dy = viewTile(viewCentre, 6), 0, 1
		case matches(viewTile(6, viewCentre)):
			tile, dx, dy = viewTile(6, viewCentre), 1, 0
		case matches(viewTile(viewCentre, 4)):
			tile, dx, dy = viewTile(viewCentre, 4), 0, -1
		default:
			// the original leaves the loop here: neither said nor sealed
			if !spoken {
				printMessage(msgNoEffectLine2)
			}
			return
		}
		if tile == mapTileRuins {
			meditate(i, dx+int(mapX), dy+int(mapY))
			return
		}
		x, y := dx+int(mapX), dy+int(mapY)
		if x == int(placeX[i+dungeonPlaceOff]) && y == int(placeY[i+dungeonPlaceOff]) {
			shrineStates[i] ^= sealBit
			*tileAt(x, y) ^= entranceTiles[i] ^ sealedTile
			viewChanged |= 2
		}
		spoken = true
		break
	}
	if !spoken {
		printMessage(msgNoEffectLine2)
	}
}

const yellLength = 30

// yellCommand furls or hoists the sails, or yells a word.
// C: CMDS_1418_YellCmd
func yellCommand() int {
	result := 1
	if partyTile&monsterMask == actorFrigate32 && mapID < firstMapCombat {
		if partyTile&shapeMask == actorFrigate32 {
			printMessage(msgFurlLine)
			partyTile += 4
		} else {
			printMessage(msgHoistLine)
			partyTile -= 4
		}
		return result
	}
	printMessage(msgWhat)
	var words [yellLength + 2]byte
	copy(words[:], readString(yellLength))
	if words[0] == 0 {
		printMessage(msgNothingLine)
		return result
	}
	printChar('\n')
	switch {
	case mapID >= 1 && mapID <= settlements:
		result = summonShadowlordByName(words[:])
	case mapID == 0:
		speakWordOfPower(words[:])
	default:
		printMessage(msgNoEffectLine2)
	}
	return result
}

// pushable tells whether a tile can be pushed.
// C: CMDS_14ba_Pushable
func pushable(tile byte) bool {
	switch tile {
	case mapTile91, mapTileChair144, mapTileChair145, mapTileChair146, mapTileChair147, mapTileDesk,
		mapTileBarrel, mapTileVanity, mapTile169, mapTileDresser, mapTile174, mapTileTrunk,
		mapTileCannon180, mapTileCannon181, mapTileCannon182, mapTileCannon183:
		return true
	}
	return false
}

// facing turns a chair or cannon tile to face the way it was moved (back when pulled).
// C: CMDS_1504
func turnedTo(dx, dy, tile int, pulled bool) int {
	if dx == 1 && dy == 0 {
		tile++
	}
	if dx == 0 && dy == 1 {
		tile += 2
	}
	if dx == -1 && dy == 0 {
		tile += 3
	}
	if pulled {
		tile ^= 2
	}
	return tile
}

// pushed moves tile from x1, y1 onto x2, y2, leaving under behind.
// C: CMDS_1548_Pushed
func pushed(tile, under, x1, y1, x2, y2, dx, dy int) {
	printMessage(msgPushedLine)
	*tileAt(x2, y2) = byte(tile)
	*tileAt(x1, y1) = byte(under)
	tile &= shapeMask
	if tile == mapTileChair144 || tile == mapTileCannon180 {
		*tileAt(x2, y2) = byte(turnedTo(dx, dy, tile, false))
	}
}

// pulled draws tile from x2, y2 to where the party stood, x1, y1.
// C: CMDS_15b0_Pulled
func pulled(floor, tile, x1, y1, x2, y2, dx, dy int) {
	printMessage(msgPulledLine)
	*tileAt(x1, y1) = byte(tile)
	*tileAt(x2, y2) = byte(floor)
	tile &= shapeMask
	if tile == mapTileChair144 || tile == mapTileCannon180 {
		*tileAt(x1, y1) = byte(turnedTo(dx, dy, tile, true))
	}
}

// pushCommand pushes (or pulls) furniture or a cannon.
// C: CMDS_161a_PushCmd
func pushCommand() {
	setTile(doorOpen, doorX, doorY)
	if !selectDirection() {
		return
	}
	dx, dy := int(moveX), int(moveY)
	var savedX, savedY byte
	if mapID >= firstMapCombat {
		savedX, savedY = mapX, mapY
		mapX, mapY = combatants[combatant].x, combatants[combatant].y
	}
	x, y := dx+int(mapX), dy+int(mapY)
	if mapID >= firstMapCombat && blockedInCombat(x, y) {
		return
	}
	tile := int(*tileAt(x, y))
	if actorAt(x, y, int(level)) != 0 || !pushable(byte(tile)) {
		printMessage(msgWonTBudgeLine)
		return
	}
	floor := mapTile68
	if tile&shapeMask == mapTileCannon180 {
		floor = mapTile69
	}
	beyondX, beyondY := x+dx, y+dy
	beyond := int(*tileAt(beyondX, beyondY))
	here := int(*tileAt(int(mapX), int(mapY)))
	switch {
	case actorAt(beyondX, beyondY, int(level)) == 0 && beyond == floor:
		pushed(tile, beyond, x, y, beyondX, beyondY, dx, dy)
	case here == floor:
		pulled(here, tile, int(mapX), int(mapY), x, y, dx, dy)
	default:
		printMessage(msgWonTBudgeLine2)
		return
	}
	mapX = byte(int(mapX) + dx)
	mapY = byte(int(mapY) + dy)
	viewChanged = 1
	if mapID >= firstMapCombat {
		slot := combatants[combatant].entity
		combatants[combatant].x += byte(dx)
		combatants[combatant].y += byte(dy)
		actors[slot].x += byte(dx)
		actors[slot].y += byte(dy)
		mapX, mapY = savedX, savedY
		updateFrame()
	}
}

const (
	escapeFrom = 1200
	escapeTo   = 2000
	escapeTime = 40
	noEscape   = 128
)

// escapeCombat flees the combat map when the party may.
// C: CMDS_17ec_Escape
func escapeCombat() int {
	printMessage(msgEscape)
	stuck := 0
	alive := false
	for slot := range combatantSlots {
		if combatants[slot].flags&(fighterPlayer|fighterDead) == fighterPlayer {
			alive = true
			break
		}
	}
	if alive {
		if combatKind&noEscape != 0 {
			printMessage(msgNotHereLine2)
			stuck = 1
		} else if battleWon == 0 {
			printMessage(msgNotYetLine)
			stuck = 1
		}
	}
	if stuck == 0 {
		printChar('!')
		for slot := range combatantSlots {
			if combatants[slot].flags != 0 {
				removeFromCombat(-slot - 1)
				updateFrame()
			}
		}
		for slot := range actorSlots {
			if actors[slot].tile != 0 {
				removeFromCombat(slot + 1)
				updateFrame()
			}
		}
		sweepTone(escapeFrom, escapeTo, 1, escapeTime)
	}
	statusChanged = true
	return stuck
}

// Mixing reagents.
const (
	reagentWindow   = 1
	reagentTop      = 1
	reagentLeft     = 24
	reagentRight    = 38
	reagentFullR    = 39
	reagentBottom   = 9
	reagentTick     = 15
	reagentMark     = 3
	mixKey          = 'M'
	cancelledMix    = -1
	mixDigits       = 2
	maxMixtures     = 99
	arrowLeftGlyph  = 27
	arrowRightGlyph = 26
	arrowUpGlyph    = 24
	arrowDownGlyph  = 25
)

// chooseReagents lets the player tick reagents; it returns their bits (-1 cancelled).
// C: CMDS_18be
func chooseReagents() int {
	var owned [reagentTypes]int
	count := 0
	for i := range reagentTypes {
		if reagents[i] != 0 {
			owned[count] = i
			count++
		}
	}
	selectWindow(reagentWindow)
	setWindowSize(reagentWindow, reagentLeft, reagentTop, reagentRight, reagentBottom)
	printChar(charClearWindow)
	setWindowSize(reagentWindow, reagentLeft, reagentTop, reagentFullR, reagentBottom)
	clearIconRow()
	showCaption(msgReagents)
	for i := 0; i != count; i++ {
		printChar('\n')
		printChar(' ')
		printNumber(int(reagents[owned[i]]), 2, '0')
		printChar(' ')
		printString(reagentNames[owned[i]])
	}
	at, chosen := 0, 0
	highlightMember(1)
	for done := false; !done; {
		selectWindow(windowScroll)
		switch getChar() {
		case keyLeft, keyUp:
			if at > 0 {
				highlightMember(at + 1)
				at--
				highlightMember(at + 1)
			}
		case keyRight, keyDown:
			if at < count-1 {
				highlightMember(at + 1)
				at++
				highlightMember(at + 1)
			}
		case mixKey:
			done = true
			printMessage(msgNewlines2)
		case keyEnter, keySpace:
			bit := 128 >> owned[at]
			chosen ^= bit
			selectWindow(reagentWindow)
			setTextPosition(reagentMark, at+1)
			printChar(charToggleInverse)
			if bit&chosen != 0 {
				printChar(reagentTick)
			} else {
				printChar(' ')
			}
			printChar(charToggleInverse)
		case keyEscape:
			chosen = cancelledMix
			done = true
			printChar('\n')
		}
	}
	highlightMember(at + 1)
	selectWindow(windowScroll)
	return chosen
}

// mixAmount asks how many mixtures to make of the chosen reagents.
// C: CMDS_1a70
func mixAmount(chosen int) int {
	for {
		enough := true
		printMessage(msgHowMuchSp)
		amount := readNumber(mixDigits)
		if amount != 0 {
			bit := 128
			for i := range reagentTypes {
				if chosen&bit != 0 && int(reagents[i]) < amount {
					printMessage(msgInsufficientReagentsLine)
					enough = false
					break
				}
				bit >>= 1
			}
		}
		if enough {
			return amount
		}
	}
}

// mixCommand mixes reagents into spells.
// C: CMDS_1ad8_MixCmd
func mixCommand() {
	total := 0
	for i := range reagentTypes {
		total += int(reagents[i])
	}
	if total == 0 {
		printMessage(msgNoReagentsOwnedLine)
		return
	}
	printMessage(msgForWhatSpell)
	spell := selectSpell()
	if spell == -1 {
		printMessage(msgNoneLine2)
	} else {
		printChar('\n')
		printChar(arrowLeftGlyph)
		printChar(',')
		printChar(arrowRightGlyph)
		printChar(',')
		printChar(arrowUpGlyph)
		printChar(',')
		printChar(arrowDownGlyph)
		printMessage(msgToMoveReturnSelectsTypeM)
		chosen := chooseReagents()
		if chosen >= 0 {
			if amount := mixAmount(chosen); amount > 0 {
				mixReagents(spell, chosen, amount)
			}
		}
	}
	clearCaption()
	clearIconBar()
	showStatus()
}

// mixReagents uses up the reagents and makes the mixtures, or springs a mishap.
func mixReagents(spell, chosen, amount int) {
	if chosen == 0 {
		printMessage(msgNothingToMixLine)
		return
	}
	printMessage(msgMixingLine)
	if mapID > settlements {
		waitTicks(10)
	} else {
		sleepFrames(10)
	}
	bit := 128
	for i := range reagentTypes {
		if chosen&bit != 0 {
			reagents[i] -= byte(amount)
		}
		bit >>= 1
	}
	if spell >= 0 && int(spellReagents[spell]) == chosen {
		printMessage(msgDoneLine)
		spellMixtures[spell] += byte(amount)
		if spellMixtures[spell] > maxMixtures {
			spellMixtures[spell] = maxMixtures
		}
		return
	}
	printChar('\n')
	firstAbleMember()
	springTrap(int(moveX))
}

const (
	mountainTile = 13
	cliffTile    = 12
	climbDex     = 30
	fallDamage   = 5
)

// klimbOutdoors climbs a cliff with the grapple.
// C: CMDS_1c20_KlimbCmd
func klimbOutdoors() {
	if grapple == 0 {
		printMessage(msgWithWhatLine)
		return
	}
	if partyTile != actorAvatar {
		printMessage(msgOnFootLine)
		return
	}
	if !selectDirection() {
		return
	}
	dx, dy := int(moveX), int(moveY)
	tile := *tileAt(int(mapX)+dx, int(mapY)+dy)
	switch {
	case tile == mountainTile:
		printMessage(msgImpassableLine)
	case tile != cliffTile:
		printMessage(msgNotClimbableLine)
	default:
		for i := range int(partySize) {
			if party[i].status != statusDead && int(party[i].dex) < randomRange(1, climbDex) {
				printMessage(msgFellLine)
				hurtMember(i, randomRange(1, fallDamage))
			}
		}
		moveParty(dx, dy)
	}
}

// dungeonCell is the dungeon map's cell at x, y on level.
func dungeonCell(x, y, level int) byte {
	return dungeonMap[level*dungeonSize*dungeonSize+y*dungeonSize+x]
}
