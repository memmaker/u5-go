package game

import "github.com/memmaker/u5-go/internal/ega"

// Segments 5000 and 6000: entering the special maps (combat and camping), setting up the
// fighters, and the frame around the screen.

// Kinds of special map (combatKind bits).
const (
	specialCombat        = 0
	specialDungeonCombat = 2
	specialCamp          = 4
	specialFixedFoes     = 128
)

const (
	combatMapID     = 255
	combatMapBytes  = viewSize * viewStride
	combatMapsFile  = "BRIT.CBT"
	partyStarts     = maxPartySize
	monsterStarts   = 16
	firstPlayerSlot = 0
	firstFoeSlot    = maxPartySize
	turnBase        = 36 // a fighter's turn timer starts at this less its dexterity
	maxDexterity    = 30
	ghostFoe        = 8 // foe types that start dead (8 and 9)
	shadowlordFoe   = 9
	foeStatBytes    = 8
	foeTypes        = 48
	hiddenFoe       = 256 // added to a foe type when the encounter was not seen coming
	pirateFoe       = 1
	sceptreFoe      = 252
)

// Combat map rows holding the starting places (combatMap rows beyond the 11x11 map).
const (
	rowNewTiles     = 0
	rowPartySouth   = 3
	rowMonsterTiles = 5
	rowMonsterX     = 6
	rowMonsterY     = 7
	specialColumn   = viewSize
	partyYColumn    = specialColumn + maxPartySize
)

// foeStats are a foe type's statistics (EnemyStats).
type foeStats struct {
	str, dex, intel, defence, attack, maxHP, maxCount, treasure byte
}

// Addresses of the combat tables in the data segment.
const (
	foeStatsAddress    = 5052 // C: D_13bc
	armourAddress      = 5684 // C: D_1634 (the defence of each item)
	foeLeadersAddress  = 5844 // C: D_16d4 (the stronger kind that may lead a group)
	foeNamesAddress    = 6326 // C: D_18b6 (pointers)
	equipmentDefenceOf = itemTypes
)

var (
	foes        [foeTypes]foeStats // C: D_13bc
	itemDefence []byte             // C: D_1634
	foeLeaders  []byte             // C: D_16d4
	foeNames    []string           // C: D_18b6
	savedActors [actorSlots]actor  // C: D_a9fc (the map's actors while on a special map)
	monsterX    [monsterStarts]byte
	monsterY    [monsterStarts]byte
	partyStartX [partyStarts]byte
	partyStartY [partyStarts]byte
	// moongateFoe is set while a moongate's guardian fights (D_bb16); moongateFloor the
	// tile under it (D_bb15).
	moongateFoe   byte // C: D_bb16
	moongateFloor byte // C: D_bb15
)

// The start places of the last combat map loaded.
// C: D_1704, D_1714, D_1724, D_172c

// loadCombatTables reads the foes' statistics and names.
func loadCombatTables() {
	stats := segment.Bytes(foeStatsAddress, foeTypes*foeStatBytes)
	for i := range foes {
		s := stats[i*foeStatBytes:]
		foes[i] = foeStats{s[0], s[1], s[2], s[3], s[4], s[5], s[6], s[7]}
	}
	itemDefence = segment.Bytes(armourAddress, equipmentDefenceOf)
	foeLeaders = segment.Bytes(foeLeadersAddress, foeTypes)
	foeNames = nil
	for _, address := range segment.Words(foeNamesAddress, foeTypes) {
		foeNames = append(foeNames, segment.String(int(address)))
	}
}

// combatCell returns the combat map byte at column, row (rows of 32).
func combatCell(column, row int) *byte { return &combatMap[row*viewStride+column] }

// enterSpecialMap leaves the map for a combat map or camp: kind specialCombat fights the
// actor in slot who; specialCamp camps with guard who for hours; specialDungeonCombat
// fights in a dungeon room. The map and actors come back afterwards.
// C: ULTIMA_5f86_SpecialMapHandler
func enterSpecialMap(kind, who, hours int) {
	combatKind = byte(kind)
	x, y, z := mapX, mapY, level
	savedMapID = mapID
	active := activeMember
	mapID = combatMapID
	savedActors = actors
	fight := true
	switch {
	case kind == specialCombat:
		foe := (int(actors[who].tile) - actorWizard) >> 2
		if actors[who].tile&^3 == actorPirate {
			foe = pirateFoe
		}
		if actors[who].hull >= firstMapCombat {
			foe += hiddenFoe
		}
		setUpCombat(1, foe)
	case kind&specialCamp != 0:
		fight = campHere(kind, who, hours)
	case kind&specialDungeonCombat != 0:
		enterDungeonRoom(0, 2)
	default:
		fight = false
	}
	if fight {
		combatant, activeMember = noCombatant, noMember
		battleWon = 0
		combatMainLoop()
	}
	if moongateFoe != 0 {
		moongateFoe = 0
		changeTimeTiles(ega.TimeLadderMend)
	}
	mapX, mapY, level = x, y, z
	mapID = savedMapID
	viewChanged = 1
	showStatus()
	// fix: with no active member the original reads a member past the party table.
	if active != noMember && party[active].status != statusDead && party[active].status != statusSleeping {
		activeMember = active
	} else {
		activeMember = noMember
	}
	actors = savedActors
}

// specialMap is enterSpecialMap under the name the camping code calls it by.
func specialMap(kind, who, hours int) { enterSpecialMap(kind, who, hours) }

// loadCombatMap reads combat map number from BRIT.CBT with its start places.
// C: ULTIMA_60ec_LoadSpecialMap
func loadCombatMap(number int) {
	readFile(combatMapsFile, combatMap[:combatMapBytes], number*combatMapBytes)
	for i := range partyStarts {
		partyStartX[i] = *combatCell(specialColumn+i, rowPartySouth)
		partyStartY[i] = *combatCell(partyYColumn+i, rowPartySouth)
	}
	for i := range monsterStarts {
		monsterX[i] = *combatCell(specialColumn+i, rowMonsterX)
		monsterY[i] = *combatCell(specialColumn+i, rowMonsterY)
	}
}

// Combat maps of BRIT.CBT.
const (
	combatMapCamp       = 0
	combatMapSwamp      = 1
	combatMapGrass      = 2
	combatMapBrush      = 3
	combatMapForest     = 4
	combatMapHills      = 5
	combatMapMountains  = 6
	combatMapBridge     = 7
	combatMapFloor      = 8
	combatMapSceptre    = 10
	combatMapShipShore  = 11
	combatMapPirateLand = 12
	combatMapShipSea    = 13
	combatMapShipShip   = 14
	combatMapWater      = 15
)

// fightActor fights the actor in slot who on the map of the ground it stands on.
// C: ULTIMA_6150_Combat
func fightActor(who int) {
	kind := int(actors[who].tile &^ 3)
	printChar('\n')
	printChar(charCentre)
	if kind < actorWizard {
		printMessage(msgPirates)
	} else {
		printString(foeNames[(kind-actorWizard)/4])
	}
	printChar(charUncentre)
	printMessage(msgNewlines2)
	ground := int(*tileAt(int(actors[who].x), int(actors[who].y)))
	onWater := ground < mapTilePoison || ground&^1 != mapTile106 && ground&^15 == mapTile96
	var number int
	if kind == sceptreFoe {
		number = combatMapSceptre
		if sceptre != 0 {
			printMessage(msgTheSceptreIsReclaimedLine)
			pulse(4050, 1, 65000, 1, 1)
			sceptre = 0
		}
	} else {
		if byte(kind)&^15 == actorMonsterGroup {
			onWater = true
		}
		switch {
		case partyTile&^7 == actorFrigate32:
			switch {
			case kind == actorPirate:
				number = combatMapShipShip
			case onWater:
				number = combatMapShipShore
			default:
				number = combatMapShipSea
			}
		case kind == actorPirate:
			number = combatMapPirateLand
		case onWater:
			number = combatMapWater
		default:
			number = groundCombatMap(ground)
		}
	}
	loadCombatMap(number)
	enterSpecialMap(specialCombat, who, 0)
	afterCombat(who)
	updateVisibility()
	passTime(0)
}

// groundCombatMap returns the combat map of a land tile.
func groundCombatMap(ground int) int {
	switch ground {
	case mapTileWater1, mapTileWater2, mapTileWater3:
		return combatMapWater
	case mapTilePoison:
		return combatMapSwamp
	case mapTileGrass:
		return combatMapGrass
	case mapTile6, mapTile8:
		return combatMapBrush
	case mapTile7, mapTile30, mapTile31:
		return combatMapForest
	case mapTile9, mapTile10:
		return combatMapHills
	case mapTile11, mapTile12, mapTile13, mapTile14, mapTile15:
		return combatMapMountains
	case mapTile29, mapTile72, mapTile73, mapTile106, mapTile107:
		return combatMapBridge
	case mapTile68:
		return combatMapFloor
	}
	if mapID != 0 {
		return combatMapFloor
	}
	return combatMapGrass
}

// camp camps on the camp map with guard for hours.
// C: ULTIMA_6360_Camping
func camp(guard, hours int) {
	loadCombatMap(combatMapCamp)
	enterSpecialMap(specialCamp, guard, hours)
	updateVisibility()
}

// The frame's corner glyphs.
const (
	glyphCornerTopLeft    = 123
	glyphCornerTopRight   = 124
	glyphCornerBottomLeft = 125
)

// drawFrame draws the frame around the map, the status panel and the scroll.
// C: ULTIMA_637e_DrawFrame
func drawFrame() {
	setPenColour(0)
	fillRect(0, 0, screenRight, screenBottom)
	setPenColour(blueColour)
	fillRect(0, 0, 319, 6)
	fillRect(0, 185, 191, 191)
	fillRect(0, 0, 6, 191)
	fillRect(185, 0, 191, 191)
	fillRect(313, 0, 319, 87)
	fillRect(192, 80, 312, 87)
	fillRect(192, 57, 312, 63)
	setTextColour(blueColour)
	setTextPosition(0, 0)
	printChar(glyphCornerTopLeft)
	setTextPosition(39, 0)
	printChar(glyphCornerTopRight)
	setTextPosition(0, 23)
	printChar(glyphCornerBottomLeft)
	setPenColour(brightWhiteColour)
	line(7, 7, 7, 184)
	lineTo(184, 184)
	lineTo(184, 7)
	lineTo(7, 7)
	line(191, 191, 191, 87)
	lineTo(319, 87)
	line(191, 7, 312, 7)
	lineTo(312, 56)
	lineTo(191, 56)
	lineTo(191, 7)
	line(191, 63, 312, 63)
	lineTo(312, 80)
	lineTo(191, 80)
	lineTo(191, 63)
	setTextColour(brightWhiteColour)
}

// Kinds of fighter placeFighter places.
const (
	placeFoe       = 0
	placeMember    = 1
	placeBareActor = 2
)

const noTarget = 255

// placeFighter puts a foe of type entity, member entity or a bare actor of tile entity at
// (x, y, z) on the combat map. It returns the fighter's slot (for placeBareActor the actor's),
// or -1 when there was no room.
// C: ULTIMA_6506
func placeFighter(entity, kind, x, y, z int) int {
	slot := -1
	var f *fighter
	if kind != placeBareActor {
		first := firstPlayerSlot
		if kind == placeFoe {
			first = firstFoeSlot
		}
		for i := first; i < combatantSlots; i++ {
			if combatants[i].flags != 0 {
				continue
			}
			f = &combatants[i]
			slot = i
			if kind == placeMember && party[entity].status != statusDead {
				f.dex = party[entity].dex
				f.turnTimer = turnBase - f.dex
				f.flags = fighterPlayer
				if party[entity].status != statusGood && party[entity].status != statusPoisoned {
					f.flags |= fighterAsleep
				}
			}
			if kind == placeFoe {
				f.hp = foes[entity].maxHP
				f.dex = byte(random(7) - 4 + int(foes[entity].dex))
				if f.dex > maxDexterity {
					f.dex = foes[entity].dex
				}
				f.turnTimer = turnBase - f.dex
				f.flags = fighterMonster
				if entity == ghostFoe || entity == shadowlordFoe {
					f.flags = fighterDead
				}
			}
			f.entity = byte(entity)
			f.x, f.y = byte(x), byte(y)
			break
		}
	}
	if slot > -1 || kind == placeBareActor {
		i := 0
		for ; i < actorSlots; i++ {
			a := &actors[i]
			if a.tile != 0 {
				continue
			}
			if kind == placeFoe {
				f.actorSlot = byte(i)
				a.tile = byte(entity*4 + actorWizard)
				a.animTile = a.tile
			}
			if kind == placeMember {
				f.actorSlot = byte(i)
			}
			if kind == placeBareActor {
				a.tile, a.animTile = byte(entity), byte(entity)
				slot = i
			}
			a.x, a.y, a.z = byte(x), byte(y), byte(z)
			a.target = noTarget
			if kind == placeFoe {
				a.hull = f.hp
			} else {
				a.hull = byte(entity)
			}
			if kind != placeBareActor {
				f.actorSlot = byte(i)
			}
			break
		}
		if kind != placeBareActor && i == actorSlots && slot > -1 {
			slot = -1
		}
	} else {
		slot = -1
	}
	if slot >= 0 {
		for i := range combatantSlots {
			if a := &actors[combatants[i].actorSlot]; int(a.target) == slot {
				a.target = noTarget
			}
		}
	}
	return slot
}

// Rings that act in combat.
const (
	invisibilityRing = 42
)

// startRings makes a member wearing the ring of invisibility invisible, and lets the
// regeneration ring work.
// C: ULTIMA_6794
func startRings(slot int) {
	f := &combatants[slot]
	if f.flags&fighterPlayer == 0 || f.flags&(fighterDead|fighterAsleep) != 0 {
		return
	}
	switch party[f.entity].equip[ringSlot] {
	case invisibilityRing:
		actors[f.actorSlot].animTile = actor29
		f.flags |= fighterInvisible
	case regenerationRing:
		regenerate()
	}
}

// wakeFighter wakes a sleeping fighter.
// C: ULTIMA_6800
func wakeFighter(slot int) {
	f := &combatants[slot]
	if f.flags&fighterAsleep != 0 {
		a := &actors[f.actorSlot]
		if f.flags&fighterPlayer != 0 {
			party[f.entity].status = statusGood
			if f.flags&fighterInvisible != 0 {
				a.animTile = actor29
			} else {
				a.animTile = a.tile
			}
		} else {
			a.state = 0
		}
		f.flags &^= fighterAsleep
	}
	showStatus()
}

// sleepUnlessPoisoned puts a fighter to sleep unless it is a poisoned member.
// C: ULTIMA_6880
func sleepUnlessPoisoned(slot int) {
	f := &combatants[slot]
	if f.flags&fighterPlayer == 0 || party[f.entity].status != statusPoisoned {
		sleepFighter(slot)
	}
}

// sleepFighter puts a fighter to sleep.
// C: ULTIMA_68ae
func sleepFighter(slot int) {
	f := &combatants[slot]
	a := &actors[f.actorSlot]
	if f.flags&fighterPlayer != 0 {
		if party[f.entity].status == statusDead {
			return
		}
		party[f.entity].status = statusSleeping
		f.flags |= fighterAsleep
		a.animTile = actorSleep
		if slot == int(activeMember) {
			activeMember = noMember
		}
		combatEvent = eventSlept
		showStatus()
		if combatKind&specialCamp == 0 {
			updateFrame()
		}
		return
	}
	f.flags |= fighterAsleep
	combatEvent = eventSlept
	a.state = frozenState
	updateFrame()
}

// frozenState stops an actor's animation.
const frozenState = 255

const (
	ringVanishes      = 11
	moongateGuardBase = 7
	moongateGuardStep = 3
)

// setUpFighters clears the fighters and places the living members by class, losing a
// worn ring now and then, and a moongate's guardian.
// C: ULTIMA_6936
func setUpFighters() {
	ring := 0
	for i := range combatantSlots {
		actors[i] = actor{target: actors[i].target}
		combatants[i] = fighter{}
	}
	for i := range int(partySize) {
		m := &party[i]
		if m.status == statusDead {
			continue
		}
		if m.equip[ringSlot] == invisibilityRing {
			ring = invisibilityRing
		}
		if m.equip[ringSlot] == regenerationRing {
			ring = regenerationRing
		}
		if ring != 0 {
			if randomRange(0, 15) == ringVanishes {
				printMessage(msgARingHasVanishedLine)
				sweepTone(1200, 2000, 1, 40)
				unequip(i, ring)
			}
			ring = 0
		}
		slot := placeFighter(i, placeMember, int(partyStartX[i]), int(partyStartY[i]), int(level))
		a := &actors[combatants[slot].actorSlot]
		a.target = noTarget
		switch m.class {
		case classAvatar:
			a.tile = actor76
		case classFighter, classPaladin, classRanger:
			a.tile = actorFighter
		case classDruid, classMage:
			a.tile = actorWizard
		case classBard, classShepherd, classTinker:
			a.tile = actorBard68
		}
		a.animTile = a.tile
		if m.status == statusSleeping {
			sleepFighter(slot)
		} else {
			startRings(slot)
		}
	}
	if *combatCell(viewCentre, viewCentre) == mapTileMoongate {
		slot := placeFighter(1, placeBareActor, viewCentre, viewCentre, int(level))
		actors[slot].hull = level*moongateGuardStep + moongateGuardBase
		*combatCell(viewCentre, viewCentre) = moongateFloor
	}
}

const (
	maxFoes      = actorSlots - maxPartySize
	crowdLimit   = 26
	leaderChance = 8
	singleFoe    = 1
	fixedEight   = 8
	fixedSixteen = 16
	pirateMapFoe = 12
)

// setUpCombat places the party (unless camping) and a group of foe, shuffling the foes'
// places when camping.
// C: ULTIMA_6bc2
func setUpCombat(kind, foe int) {
	seen := true
	if foe >= hiddenFoe {
		foe -= hiddenFoe
		seen = false
	}
	exitDirection = 0
	if kind&specialCamp == 0 {
		setUpFighters()
	}
	var order [monsterStarts]int
	for i := range order {
		order[i] = i
	}
	if kind&specialCamp != 0 {
		for i := range monsterStarts - 1 {
			j := randomRange(0, monsterStarts-1)
			order[i], order[j] = order[j], order[i]
		}
	}
	// The original tests kind != 128 || kind != 4, which always holds.
	printMessage(msgConflictLine)
	var count int
	if savedMapID != 0 && savedMapID < firstDungeonMap && foe != pirateMapFoe && seen {
		count = singleFoe
	} else {
		count = int(foes[foe].maxCount)
	}
	if count != fixedEight && count != fixedSixteen && count != singleFoe {
		count = randomRange(1, count)
		if ambushed != 0 {
			count = randomRange(1, count)
		}
		updateFrame()
		if count > maxFoes-1 {
			count = crowdLimit
		}
	}
	placeFighter(foe, placeFoe, int(monsterX[order[0]]), int(monsterY[order[0]]), int(level))
	leaders := count/4 + 1
	for i := 1; i < count; i++ {
		kind := foe
		if i < leaders && random(leaderChance) == 0 {
			kind = int(foeLeaders[foe])
		}
		placeFighter(kind, placeFoe, int(monsterX[order[i]]), int(monsterY[order[i]]), int(level))
	}
}

// onCombatMap says whether (x, y) lies on the 11x11 combat map.
// C: ULTIMA_6d82
func onCombatMap(x, y int) bool { return x > -1 && x < viewSize && y > -1 && y < viewSize }

const (
	doomMap        = 40
	mysticArmour   = 15
	bodySlot       = 1
	protectionIcon = 'P'
	protectionGain = 3
)

// memberDefence returns a member's defence. In the original the test for an empty slot
// compares the item's number, unsigned, with 65535 and never passes, so the equipment adds
// nothing: only the protection spell counts. The port keeps that.
// C: ULTIMA_6da8
func memberDefence(row int) int {
	m := &party[row]
	if savedMapID == doomMap && m.equip[bodySlot] != mysticArmour {
		return 0
	}
	defence := 0
	if spellIcon == protectionIcon {
		defence += protectionGain
	}
	return defence
}

// unequip takes item off member. It returns whether it was worn.
// C: ULTIMA_6e60
func unequip(row, item int) bool {
	m := &party[row]
	for slot := range equipSlots {
		if int(m.equip[slot]) != item {
			continue
		}
		m.equip[slot] = noEquipment
		if slot == ringSlot && item == invisibilityRing && mapID >= firstMapCombat && combatant < combatantSlots {
			combatants[combatant].flags &^= fighterInvisible
		}
		return true
	}
	return false
}

// findWord returns where word first appears in text, ignoring case and the high bits, or
// -1. After a partial match the original skips ahead past the letters matched, so a match
// overlapping a false start is missed; that is kept.
// C: ULTIMA_6f1e
func findWord(word, text []byte) int {
	wordLen, textLen := cLen(word), cLen(text)
	if wordLen > textLen {
		return -1
	}
	for at := 0; at <= textLen-wordLen; {
		matched := 0
		for matched < wordLen && foldLetter(text[at+matched]) == foldLetter(word[matched]) {
			matched++
		}
		if matched == wordLen {
			return at
		}
		at += matched + 1
	}
	return -1
}

// cLen returns the length of the zero-terminated b.
func cLen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}

// foldLetter drops a character's high bit and upper-cases letters (and the five
// characters above them, as the original's mask does).
func foldLetter(c byte) byte {
	c &= asciiMask
	if c > lastBeforeLower {
		c &= upperMask
	}
	return c
}

const (
	lastBeforeLower = 96
	upperMask       = 95
)

// changeTileSet has the driver swap in tile changes (mode plus 3).
// C: ULTIMA_6f9e
func changeTileSet(mode int) { screen.UpdateTimeTiles(mode+3, 0, 0) }
