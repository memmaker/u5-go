package game

// Searching, jimmying locks, opening and getting things (sjog.c), with the combat
// moves that share the overlay.

// Addresses of the overlay's tables in the data segment.
const (
	itemNamesAddress     = 6134  // C: D_17f6
	herbXAddress         = 15974 // C: D_3e66 (where the herbs grow)
	herbYAddress         = 15978 // C: D_3e6a
	herbReagentAddress   = 15982 // C: D_3e6e
	herbNamesAddress     = 15986 // C: D_3e72
	hiddenTileAddress    = 15992 // C: D_3e78
	hiddenQualityAddress = 16106 // C: D_3eea
	hiddenMapAddress     = 16220 // C: D_3f5c
	hiddenLevelAddress   = 16334 // C: D_3fce
	hiddenXAddress       = 16448 // C: D_4040
	hiddenYAddress       = 16562 // C: D_40b2
	lootKindsAddress     = 16676 // C: D_4124
	lootOddsAddress      = 16684 // C: D_412c
	lootAmountsAddress   = 16692 // C: D_4134
	chestItemsAddress    = 16700 // C: D_413c
	chestItemOddsAddress = 16748 // C: D_416c
	potionNamesAddress   = 16796 // C: D_419c
	scrollNamesAddress   = 16812 // C: D_41ac
	dungeonLootOdds      = 16828 // C: D_41bc
	dungeonLootAmounts   = 16836 // C: D_41c4
	dungeonLootKinds     = 16844 // C: D_41cc
)

const (
	herbPlaces    = 3
	hiddenObjects = 114
	lootKinds     = 8
	dungeonLoot   = 7
)

// The overlay's tables.
var (
	itemNames       []string // C: D_17f6
	herbX           []byte   // C: D_3e66
	herbY           []byte   // C: D_3e6a
	herbReagent     []byte   // C: D_3e6e
	herbNames       []string // C: D_3e72
	hiddenTiles     []byte   // C: D_3e78
	hiddenQuality   []byte   // C: D_3eea
	hiddenMap       []byte   // C: D_3f5c
	hiddenLevel     []byte   // C: D_3fce
	hiddenX         []byte   // C: D_4040
	hiddenY         []byte   // C: D_40b2
	lootTiles       []byte   // C: D_4124
	lootOdds        []byte   // C: D_412c
	lootAmounts     []byte   // C: D_4134
	chestItems      []byte   // C: D_413c
	chestItemOdds   []byte   // C: D_416c
	potionNames     []string // C: D_419c
	scrollNames     []string // C: D_41ac
	dungeonOdds     []byte   // C: D_41bc
	dungeonAmounts  []byte   // C: D_41c4
	dungeonLootTile []byte   // C: D_41cc
)

// loadSjogTables reads the overlay's tables.
func loadSjogTables() {
	itemNames = messages(itemNamesAddress, itemTypes)
	herbX = segment.Bytes(herbXAddress, herbPlaces)
	herbY = segment.Bytes(herbYAddress, herbPlaces)
	herbReagent = segment.Bytes(herbReagentAddress, herbPlaces)
	herbNames = messages(herbNamesAddress, herbPlaces)
	hiddenTiles = segment.Bytes(hiddenTileAddress, hiddenObjects)
	hiddenQuality = segment.Bytes(hiddenQualityAddress, hiddenObjects)
	hiddenMap = segment.Bytes(hiddenMapAddress, hiddenObjects)
	hiddenLevel = segment.Bytes(hiddenLevelAddress, hiddenObjects)
	hiddenX = segment.Bytes(hiddenXAddress, hiddenObjects)
	hiddenY = segment.Bytes(hiddenYAddress, hiddenObjects)
	lootTiles = segment.Bytes(lootKindsAddress, lootKinds)
	lootOdds = segment.Bytes(lootOddsAddress, lootKinds)
	lootAmounts = segment.Bytes(lootAmountsAddress, lootKinds)
	chestItems = segment.Bytes(chestItemsAddress, itemTypes)
	chestItemOdds = segment.Bytes(chestItemOddsAddress, itemTypes)
	potionNames = messages(potionNamesAddress, potionTypes)
	scrollNames = messages(scrollNamesAddress, scrollTypes)
	dungeonOdds = segment.Bytes(dungeonLootOdds, lootKinds)
	dungeonAmounts = segment.Bytes(dungeonLootAmounts, lootKinds)
	dungeonLootTile = segment.Bytes(dungeonLootKinds, lootKinds)
}

// The directions of a combat move.
const (
	dirLeft  = 1
	dirRight = 2
	dirUp    = 3
	dirDown  = 4
)

// Dungeon cells: the high nibble is the kind, bit 8 marks a cell seen.
const (
	cellKindMask   = 240
	cellSeen       = 8
	cellLow        = 15
	cellNothing    = 0
	cellLadderUp   = 16
	cellLadderDown = 32
	cellLadders    = 48
	cellChest      = 64
	cellFountain   = 80
	cellTrap       = 96
	cellPit        = 97
	cellBombTrap   = 98
	cellOpenChest  = 112
	cellField      = 128
	cellSleepField = 128
	cellPoison     = 129
	cellFireField  = 130
	cellElectric   = 131
	cellImpossible = 144
	cellDoor       = 160
	cellWall       = 176
	cellRubble     = 192
	cellSecret     = 208
	cellDoors      = 224
	cellDoorsTwo   = 240
	chestTrapBits  = 7
	notSeenLock    = 247
	lookCaveIn     = 2
	lookStalactite = 1
	dungeonView    = 8
	dungeonViewEnd = 183
	lastLevel      = 7
)

// lastFreeSlot is the last empty actor slot (0 when none).
// C: SJOG_0000
func lastFreeSlot() int {
	for i := actorSlots - 1; i >= 0; i-- {
		if actors[i].tile == 0 {
			return i
		}
	}
	return 0
}

// stepFacing sets moveX, moveY one step from x, y the way given.
// C: SJOG_002a
func stepFacing(way, x, y int) {
	moveX, moveY = int16(x), int16(y)
	switch way {
	case 0:
		moveY--
	case 1:
		moveX++
	case 2:
		moveY++
	case 3:
		moveX--
	}
}

// lookDirection asks which way to look in a dungeon facing the way given.
// C: SJOG_006c
func lookDirection(facing int) bool {
	defer asking(promptDir)()
	printMessage(msgDir)
	key := 0
	for {
		key = getChar()
		if key == keySpace || key == keyUp || key == keyDown || key == keyRight || key == keyLeft {
			break
		}
	}
	switch key {
	case keySpace:
		printMessage(msgPassLine2)
		return false
	case keyUp:
		printMessage(msgAheadLine)
		stepFacing(facing, int(mapX), int(mapY))
	case keyDown:
		printMessage(msgHereLine)
		moveX, moveY = int16(mapX), int16(mapY)
	case keyRight:
		printMessage(msgRightLine)
		stepFacing((facing+1)%4, int(mapX), int(mapY))
	case keyLeft:
		printMessage(msgLeftLine)
		stepFacing((facing+3)%4, int(mapX), int(mapY))
	}
	return true
}

// describeFind names what was found.
// C: SJOG_012a
func describeFind(tile int) {
	switch tile {
	case actorChest:
		printMessage(msgAChestLine)
	case actorGold:
		printMessage(msgASackOfGoldLine)
	case actorPotion:
		printMessage(msgAPotionLine)
	case actorScroll:
		printMessage(msgAScrollLine)
	case actorWeapon:
		printMessage(msgAWeaponLine)
	case actorShield:
		printMessage(msgAShieldLine)
	case actorKey:
		printMessage(msgARingOfKeysLine)
	case actorGem:
		printMessage(msgAGemLine)
	case actorHelm:
		printMessage(msgAHelmLine)
	case actorRing:
		printMessage(msgARingLine)
	case actorArmour:
		printMessage(msgSomeArmourLine)
	case actorAmulet:
		printMessage(msgAnAmuletLine)
	case actorTorch:
		printMessage(msgSomeTorchesLine)
	case actorFood:
		printMessage(msgSomeFoodLine)
	case actorMoonstone:
		printMessage(msgAStrangeRockLine)
	case actorSleep:
		printMessage(msgARottingBodyLine)
	case actorDead:
		printMessage(msgAMoldyCorpseLine)
	default:
		printMessage(msgNothingOfNoteLine2)
	}
}

const (
	plagueOdds    = 31
	plagueRoll    = 19
	plagueRate    = 40
	plagueTime    = 3000
	plagueLimit   = 500
	corpseFindMax = 3
)

// searchBody searches the body in actor slot as member who.
// C: SJOG_01f2
func searchBody(slot, who int) {
	if randomRange(0, 7) != 0 {
		setActor(0, 0, 0, 0, 0, 0, slot)
		if randomRange(0, plagueOdds) == plagueRoll {
			printMessage(msgPlagueLine)
			whiteNoise(plagueRate, plagueTime, plagueLimit)
			party[who].status = statusPoisoned
			statusChanged = true
			return
		}
		switch randomRange(0, randomRange(0, 3)) {
		case 0:
			printMessage(msgNothingLine2)
		case 1:
			printMessage(msgWormsLine)
		case 2:
			printMessage(msgGutsLine)
		case 3:
			printMessage(msgABloodyPulpLine)
		}
		return
	}
	a := &actors[slot]
	if randomRange(0, 3) == 0 {
		printMessage(msgFoodLine)
		a.tile, a.animTile = actorFood, actorFood
	} else {
		printMessage(msgGoldLine)
		a.tile, a.animTile = actorGold, actorGold
	}
	a.hull = byte(randomRange(1, corpseFindMax))
	viewChanged |= 2
	updateFrame()
}

// Bits of a chest's lock (an actor's hull).
const (
	chestTrapped  = 128
	chestLevel    = 127
	trapSkill     = 30
	simpleTrap    = 10
	complexTrap   = 20
	trapCheckTime = 30
)

// examineChest tells member who what trap the chest in slot holds.
// C: SJOG_02ea
func examineChest(slot, who int) {
	lock := int(actors[slot].hull)
	var odds int
	if lock&chestTrapped == 0 {
		odds = -(int(party[who].intel) - trapSkill) >> 1
	} else {
		odds = (lock&chestLevel - int(party[who].intel) + trapSkill) >> 1
	}
	right := odds <= randomRange(1, trapSkill)
	if right == (lock&chestTrapped == 0) {
		printMessage(msgNoTrapLine)
		return
	}
	level := lock & chestLevel
	switch {
	case right && level < simpleTrap:
		printMessage(msgASimpleTrapLine)
	case right && level > complexTrap:
		printMessage(msgAComplexTrapLine)
	default:
		printMessage(msgATrapLine)
	}
}

// digMoonstone digs up a buried moonstone at x, y, z.
// C: SJOG_03a8
func digMoonstone(x, y, z int) bool {
	for i := moonstones - 1; i >= 0; i-- {
		if int(moonstoneX[i]) != x || int(moonstoneY[i]) != y || int(moonstoneLevel[i]) != z || moonstoneState[i] != mapID {
			continue
		}
		// the original asks this the same way 32 times over
		if actorAt(x, y, z) == actorMoonstone && int(actors[moveX].hull) == i {
			continue
		}
		slot := freeActorSlot()
		setActor(actorMoonstone, actorMoonstone, byte(x), byte(y), byte(z), byte(i), slot)
		printMessage(msgAStrangeRockLine)
		viewChanged |= 2
		return true
	}
	return false
}

const (
	sprigsMin = 2
	sprigsMax = 15
)

// pickHerbs gathers the herbs growing at x, y, once a day at midnight.
// C: SJOG_045a
func pickHerbs(x, y int) bool {
	for i := range herbPlaces {
		if int(herbX[i]) != x || int(herbY[i]) != y || hour != 0 || harvestDay[i] == day {
			continue
		}
		harvestDay[i] = day
		reagent := herbReagent[i]
		sprigs := randomRange(sprigsMin, sprigsMax)
		reagents[reagent] += byte(sprigs)
		if reagents[reagent] > maxItem {
			reagents[reagent] = maxItem
		}
		if sprigs < 10 {
			printNumber(sprigs, 1, ' ')
		} else {
			printNumber(sprigs, 2, ' ')
		}
		printMessage(msgSprigsOfLine)
		printString(herbNames[i])
		printMessage(msgNewline)
		return true
	}
	return false
}

// Hidden objects with a rule of their own.
const (
	hiddenKeys       = 13
	hiddenSkullKeys  = 14
	hiddenGlass      = 15
	hiddenGlassItem  = 39
	hiddenListLength = 113
)

// findHidden turns up the object hidden at x, y, if not yet taken.
// C: SJOG_0514
func findHidden(x, y int) {
	found := false
	i := 0
	for i = 0; i < hiddenListLength; i++ {
		if hiddenMap[i] != mapID || hiddenLevel[i] != level || int(hiddenX[i]) != x || int(hiddenY[i]) != y {
			continue
		}
		if (i == hiddenKeys && keys == 0 && actorAt(x, y, int(level)) == 0) ||
			(i == hiddenSkullKeys && day != skullKeyDay) ||
			(i == hiddenGlass && equipment[hiddenGlassItem] == 0 && actorAt(x, y, int(level)) == 0) {
			found = true
			break
		}
		bit := byte(1 << (i & 7))
		if eventFlags[i>>3]&bit == 0 && (i < hiddenKeys || i > hiddenGlass) {
			eventFlags[i>>3] |= bit
			found = true
			break
		}
	}
	if i == hiddenSkullKeys {
		skullKeyDay = day
	}
	if !found {
		printMessage(msgNothingOfNoteLine)
		return
	}
	slot := freeActorSlot()
	tile := hiddenTiles[i]
	setActor(tile, tile, hiddenX[i], hiddenY[i], hiddenLevel[i], hiddenQuality[i], slot)
	viewChanged |= 2
	describeFind(int(tile))
}

// searchInDungeon searches a dungeon cell.
// C: SJOG_0646_SearchInDungeon
func searchInDungeon() {
	who := chooseMember() + 1
	if who == 0 {
		return
	}
	if torchLight == 0 && lightSpell == 0 {
		printMessage(msgYouFindDarknessLine)
		return
	}
	if !lookDirection(int(facing)) {
		return
	}
	x, y := int(moveX)&7, int(moveY)&7
	cell := &dungeonMap[int(level)*dungeonSize*dungeonSize+y*dungeonSize+x]
	kind := int(*cell)
	dex := int(party[who].dex)
	skill := (int(level)*2 - dex + trapSkill) >> 1
	printMessage(msgYouFindLine)
	switch kind & cellKindMask {
	case cellNothing:
		printMessage(msgNothingOfNoteLine2)
	case cellLadderUp, cellLadderDown, cellLadders:
		printMessage(msgNothingHiddenOnTheLadderLine)
	case cellChest:
		var trap int
		if randomRange(1, trapSkill) > skill {
			if kind == cellChest {
				printMessage(msgNoTrapLine2)
				break
			}
			trap = int(level)
		} else {
			trap = randomRange(1, 8)
		}
		switch {
		case trap < 4:
			printMessage(msgASimpleTrapLine2)
		case trap >= 7:
			printMessage(msgAComplexTrapLine2)
		default:
			printMessage(msgATrapLine2)
		}
	case cellFountain:
		printMessage(msgNothingHiddenOnTheFountainLine)
	case cellTrap:
		switch kind {
		case cellTrap:
			printMessage(msgNothingHiddenInThePitLine)
		case cellPit:
			printMessage(msgAPitLine2)
			*cell = byte(kind&cellSeen + cellTrap)
			if level < lastLevel {
				dungeonMap[(int(level)+1)*dungeonSize*dungeonSize+y*dungeonSize+x] |= cellSeen
			}
			drawDungeon(1)
			reveal(dungeonView, dungeonView, dungeonViewEnd, dungeonViewEnd)
		case cellBombTrap:
			if skill < randomRange(1, trapSkill) {
				printMessage(msgABombTrapLine)
				*cell &= cellSeen
			} else {
				printMessage(msgNothingOfNoteLine2)
			}
		}
	case cellOpenChest:
		printMessage(msgTreasureLine)
	case cellField:
		switch kind {
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
	case cellImpossible:
		printMessage(msgThisTileIsImpossibleLine)
	case cellDoor:
		printMessage(msgNothingHiddenOnTheDoorLine)
	case cellWall:
		printMessage(msgNothingHiddenOnTheWallLine)
	case cellRubble:
		switch dungeonLook & cellLow {
		case lookCaveIn:
			printMessage(msgNothingInTheCavedInPassage)
		case lookStalactite:
			printMessage(msgNothingOnTheStalactiteLine)
		default:
			printMessage(msgNothingHiddenOnTheSkeletonLine)
			printMessage(msgItCrumblesAwayLine)
			*cell = byte(kind&cellSeen + cellWall)
			drawDungeon(1)
			reveal(dungeonView, dungeonView, dungeonViewEnd, dungeonViewEnd)
		}
	case cellSecret:
		printMessage(msgAHiddenDoorLine)
		*cell = byte(kind&cellSeen + cellDoors)
		drawDungeon(1)
		reveal(dungeonView, dungeonView, dungeonViewEnd, dungeonViewEnd)
	case cellDoors, cellDoorsTwo:
		printMessage(msgNothingHiddenOnTheDoorLine)
	}
}

// searchCommand searches beside the party.
// C: SJOG_095c_SearchCmd
func searchCommand() {
	if inDungeon() {
		searchInDungeon()
		return
	}
	if !selectDirection() {
		return
	}
	x, y := int(mapX)+int(moveX), int(mapY)+int(moveY)
	who := chooseMember()
	if who == -1 {
		return
	}
	tile := int(*tileAt(x, y))
	slot := 1
	for slot = 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		if int(a.x) == x && int(a.y) == y && (mapID >= firstMapCombat || a.z == level) && a.tile == actorChest {
			break
		}
	}
	if slot < actorSlots {
		printMessage(msgThouDostFindLine)
		examineChest(slot, who)
		return
	}
	if topActorAt(x, y, int(level)) == actorDead {
		printMessage(msgThouDostFindLine)
		searchBody(int(moveX), who)
		return
	}
	switch tile {
	case mapTileStump:
		printMessage(msgInTheStumpT)
	case mapTileShelf:
		printMessage(msgOnTheShelfT)
	case mapTileBookshelf, mapTileBookshelf + 1:
		printMessage(msgInTheBookshelfT)
	case mapTileWell:
		printMessage(msgNearTheWellT)
	case mapTileDesk:
		printMessage(msgInTheDeskT)
	case mapTileBarrel:
		printMessage(msgInTheBarrelT)
	case mapTileVanity:
		printMessage(msgInTheVanityT)
	case mapTileBed, mapTileBed + 1:
		printMessage(msgUnderTheBedT)
	case mapTileDresser:
		printMessage(msgInTheDresserT)
	case mapTileTrunk:
		printMessage(msgInTheTrunkT)
	case mapTileFireplace:
		printMessage(msgInTheFireplaceT)
	case mapTileBrazier:
		printMessage(msgInTheBrazierT)
	case mapTileWall:
		printMessage(msgInTheWallT)
	default:
		printMessage(msgT)
	}
	printMessage(msgHouDostFindLine)
	if tile == mapTileHiddenDoor {
		printMessage(msgAHiddenDoorLine2)
		if level < basementLevels {
			*tileAt(x, y) = mapTileDoor185
		} else {
			*tileAt(x, y) = mapTileDoor184
		}
		viewChanged |= 2
		return
	}
	if (tile == mapTileMoongate || !digMoonstone(x, y, int(level))) && !pickHerbs(x, y) {
		findHidden(x, y)
	}
}

const (
	keyBreakFrom = 800
	keyBreakTo   = 2000
	keyBreakTime = 50
)

// pickChestLock picks the lock of the chest in slot.
// C: SJOG_0baa
func pickChestLock(slot int) {
	who := chooseMember()
	if who == -1 {
		return
	}
	lock := actors[slot].hull
	if lock < chestTrapped {
		printMessage(msgKeyBrokeLine)
	} else {
		odds := byte((int(lock&chestLevel) - int(party[who].dex) + trapSkill) >> 1)
		if int(odds) < randomRange(1, trapSkill) {
			printMessage(msgSuccessLine)
			actors[slot].hull &= chestLevel
			return
		}
		printMessage(msgKeyBrokeLine)
	}
	sweepTone(keyBreakFrom, keyBreakTo, 1, keyBreakTime)
	keys--
}

// jimmyInDungeon picks the lock of a dungeon chest.
// C: SJOG_0c3e_JimmyInDungeon
func jimmyInDungeon() {
	printMessage(msgNewline)
	who := chooseMember()
	if who == -1 {
		return
	}
	x, y := int(mapX)&7, int(mapY)&7
	cell := &dungeonMap[int(level)*dungeonSize*dungeonSize+y*dungeonSize+x]
	kind := int(*cell)
	skill := (int(level)*2 - int(party[who].dex) + trapSkill) >> 1
	switch {
	case kind&notSeenLock == cellChest:
		if keys == 0 {
			printMessage(msgNoKeysLine)
		} else {
			printMessage(msgKeyBrokeLine)
			keys--
		}
	case kind&cellKindMask == cellChest:
		switch {
		case keys == 0:
			printMessage(msgNoKeysLine)
		case skill < randomRange(1, trapSkill):
			printMessage(msgChestUnlockedLine)
			*cell = byte(kind&cellSeen + cellChest)
		default:
			printMessage(msgKeyBrokeLine)
			keys--
		}
	case kind&cellKindMask == cellOpenChest:
		printMessage(msgAlreadyOpenLine)
	default:
		printMessage(msgWhatLine)
	}
}

const (
	jimmySkill     = 29
	freedSchedule  = 5
	freedKarma     = 2
	stocksLeft     = mapTile132
	stocksRight    = mapTile133
	lockedDoorLeft = mapTile151
	lockedDoorTwo  = mapTile152
	heavyDoor      = mapTile153
	openFloor      = mapTile68
)

// jimmyCommand picks a lock.
// C: SJOG_0d4a_JimmyCmd
func jimmyCommand() {
	if inDungeon() {
		jimmyInDungeon()
		return
	}
	if keys == 0 {
		printMessage(msgNoKeysLine2)
		return
	}
	if !selectDirection() {
		return
	}
	x, y := int(mapX)+int(moveX), int(mapY)+int(moveY)
	tile := int(*tileAt(x, y))
	switch tile {
	case mapTileDoor185, mapTileDoor187:
		who := chooseMember()
		if who == -1 {
			break
		}
		if int(party[who].dex) <= randomRange(0, jimmySkill) {
			printMessage(msgKeyBrokeLine)
			keys--
			break
		}
		*tileAt(x, y) = byte(tile - 1)
		viewChanged |= 2
		printMessage(msgUnlockedLine)
	case lockedDoorLeft, lockedDoorTwo:
		printMessage(msgKeyBrokeLine)
		keys--
	case stocksLeft, stocksRight:
		if mapID < firstMapCombat && actorAt(x, y, int(level)) == 0 {
			printMessage(msgNoOneIsThereLine)
			break
		}
		slot := int(moveX)
		who := chooseMember()
		if who == -1 {
			break
		}
		if int(party[who].dex) <= randomRange(0, jimmySkill) {
			printMessage(msgKeyBrokeLine)
			keys--
			break
		}
		if mapID < firstMapCombat-1 {
			person := personShownBy(slot)
			if person == -1 {
				printMessage(msgCouldnTFindThisNpcLine)
				break
			}
			if npcs[person].mood != 0 {
				npcs[person].mood = 0
			}
			if !personKilled(person) {
				for i := range schedulePoints {
					schedules[person].kind[i] = freedSchedule
				}
				printMessage(msgIThankTheeLine)
				increaseByte(&karma, freedKarma, maxItem)
			}
			markKilled(person)
			break
		}
		*tileAt(x, y) = openFloor
		viewChanged |= 2
		printMessage(msgUnlockedLine2)
	default:
		slot := 1
		for slot = 1; slot < actorSlots; slot++ {
			a := &actors[slot]
			if int(a.x) == x && int(a.y) == y && (mapID >= firstMapCombat || a.z == level) && a.tile == actorChest {
				break
			}
		}
		if slot < actorSlots {
			pickChestLock(slot)
			break
		}
		printMessage(msgNoLockLine)
	}
}

// dropLoot puts an item of kind tile (amount, or a random one for chests and gold)
// beside the opened chest.
// C: SJOG_0f88
func dropLoot(tile, amount, x, y, z, level int, found *bool) {
	if tile == actorPotion || tile == actorScroll {
		amount--
	}
	switch tile {
	case actorChest:
		amount = randomRange(1, level)
	case actorGold:
		amount = randomRange(1, level*3)
	}
	slot := lastFreeSlot()
	if slot == 0 {
		return
	}
	setActor(byte(tile), byte(tile), byte(x), byte(y), byte(z), byte(amount), slot)
	if mapID < firstMapCombat {
		actors[slot].target = 0
		actors[slot].z = byte(z)
	} else {
		actors[slot].target = combatantSlots
	}
	viewChanged |= 2
	updateFrame()
	if !*found {
		printMessage(msgFoundLine)
		*found = true
	}
	describeFind(tile)
}

// dropTreasure spills the plain loot of a chest of the given level.
// C: SJOG_1040
func dropTreasure(level, x, y, z int, found *bool) {
	for i := lootKinds - 1; i >= 0; i-- {
		if int(lootOdds[i]) > level || int(lootOdds[i]) > randomRange(1, trapSkill) {
			continue
		}
		amount := 1
		if lootAmounts[i] != 1 {
			amount = randomRange(1, int(lootAmounts[i]))
		}
		dropLoot(int(lootTiles[i]), amount, x, y, z, level, found)
	}
}

// dropEquipment spills the equipment of a chest of the given level.
// C: SJOG_10b8
func dropEquipment(level, x, y, z int, found *bool) {
	for tries := level / 2; tries >= 0; tries-- {
		item := randomRange(0, itemTypes-1)
		if int(chestItemOdds[item]) > level || int(chestItemOdds[item]) > randomRange(1, trapSkill) {
			continue
		}
		dropLoot(int(chestItems[item]), item, x, y, z, level, found)
	}
}

const openChestKarma = 2

// openChest opens the chest at x, y, z.
// C: SJOG_112c_OpenChest
func openChest(x, y, z int) {
	found := false
	slot := 1
	for slot = 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		if int(a.x) != x || int(a.y) != y || (mapID < firstMapCombat && int(a.z) != z) {
			continue
		}
		if a.tile == actorChest {
			break
		}
		if a.tile == actorSandalwoodBox {
			printMessage(msgCanTLine)
			return
		}
	}
	if slot == actorSlots {
		printMessage(msgNothingToOpenLine)
		return
	}
	who := chooseMember()
	if who == -1 {
		return
	}
	lock := actors[slot].hull
	setActor(0, 0, 0, 0, 0, 0, slot)
	viewChanged |= 2
	if mapID >= 1 && mapID <= settlements {
		if karma > openChestKarma {
			karma -= openChestKarma
		} else {
			karma = 0
		}
	}
	if lock >= chestTrapped {
		lock &= chestLevel
		printMessage(msgTrappedLine)
		springTrap(who)
		if mapID >= firstMapCombat && party[who].status == statusDead {
			for i := range combatantSlots {
				f := &combatants[i]
				if f.flags&fighterPlayer != 0 && int(f.entity) == who {
					f.flags |= fighterDead
					a := &actors[f.actorSlot]
					a.tile, a.animTile = actorSleep, actorSleep
					break
				}
			}
			if who == int(activeMember) {
				activeMember = noMember
			}
			updateFrame()
		}
	}
	dropTreasure(int(lock), x, y, z, &found)
	dropEquipment(int(lock), x, y, z, &found)
	if !found {
		printMessage(msgChestEmptyLine)
	}
}

// openInDungeon opens a dungeon chest.
// C: SJOG_12d4_OpenInDungeon
func openInDungeon() {
	cell := &dungeonMap[int(level)*dungeonSize*dungeonSize+int(mapY&7)*dungeonSize+int(mapX&7)]
	kind := int(*cell)
	switch kind & cellKindMask {
	case cellChest:
		who := chooseMember()
		if who == -1 {
			return
		}
		if kind&chestTrapBits != 0 {
			springTrap(who)
		}
		*cell = byte(kind&cellSeen + cellOpenChest)
		printMessage(msgChestOpenedLine2)
	case cellOpenChest:
		printMessage(msgAlreadyOpenLine2)
	default:
		printMessage(msgWhatLine)
	}
}

const doorStaysOpen = 4

// basementLevels marks the levels below ground, where the hidden doors open.
const basementLevels = 128

// openCommand opens a door or chest.
// C: SJOG_1374_OpenCmd
func openCommand() {
	if inDungeon() {
		openInDungeon()
		return
	}
	setTile(doorOpen, doorX, doorY)
	if !selectDirection() {
		return
	}
	x, y := int(mapX)+int(moveX), int(mapY)+int(moveY)
	tile := *tileAt(x, y)
	switch tile {
	case mapTileTrunk:
		printMessage(msgItSOpenLine)
	case heavyDoor:
		printMessage(msgTooHeavyLine)
	case lockedDoorLeft, lockedDoorTwo, mapTileDoor185, mapTileDoor187:
		printMessage(msgLockedLine)
	case mapTileDoor184, mapTileDoor186:
		doorOpen = tile
		doorTurns = doorStaysOpen
		doorX, doorY = byte(x), byte(y)
		*tileAt(x, y) = openFloor
		viewChanged = 1
		printMessage(msgOpenedLine)
	default:
		openChest(x, y, int(level))
	}
}

// Special items' marks.
const (
	carpetPersonCastle = 22
	sandalwoodFlag     = 67
	sandalwoodBit      = 128
	capePlans          = 255
	oddKeys            = 128
	plentifulAmmo      = 5
	arrowsItem         = 27
	quarrelsItem       = 29
	shardMask          = 3
	owned              = 255
	notInActor         = 32
)

// getObject picks up the item of kind tile with the amount (or kind) given, from actor
// slot (32 none).
// C: SJOG_1458_GetObject
func getObject(tile, amount, slot int) {
	switch tile {
	case actorChest:
		printMessage(msgOpenItFirstLine)
		return
	case actorMoonstone:
		printMessage(msgAMoonstoneLine)
		moonstoneState[amount] = owned
	case actorCarpet:
		printMessage(msgAMagicCarpetLine)
		magicCarpets++
		if magicCarpets == maxItem+1 {
			magicCarpets = maxItem
		}
		if mapID == mapCastle {
			removePerson(carpetPersonCastle)
		}
	case actorFood:
		printNumber(amount, 1, ' ')
		printMessage(msgFoodLine2)
		increaseInt(&food, amount, maxFood)
	case actorSandalwoodBox:
		printMessage(msgASandalwoodBoxLine)
		sandalwoodBox = owned
		npcKilled[sandalwoodFlag] |= sandalwoodBit
	case actorTorch:
		printNumber(amount, 1, ' ')
		printMessage(msgTorch)
		if amount == 1 {
			printMessage(msgChar33Newline)
		} else {
			printMessage(msgEsLine)
		}
		increaseByte(&torches, amount, maxItem)
	case actorGem:
		printNumber(amount, 1, ' ')
		printMessage(msgGem)
		if amount == 1 {
			printMessage(msgChar33Newline)
		} else {
			printMessage(msgSLine)
		}
		increaseByte(&gems, amount, maxItem)
	case actorKey:
		if amount >= oddKeys {
			amount &= chestLevel
			printNumber(amount, 1, ' ')
			printMessage(msgOddKey)
			increaseByte(&skullKeys, amount, maxItem)
		} else {
			printNumber(amount, 1, ' ')
			printMessage(msgKey)
			increaseByte(&keys, amount, maxItem)
		}
		if amount == 1 {
			printMessage(msgChar33Newline)
		} else {
			printMessage(msgSLine)
		}
	case actorScroll:
		if amount == owned {
			printMessage(msgThePlansForTheHmsCape)
			hmsCapePlans = owned
			break
		}
		printMessage(msgAScrollSp)
		selectFont(1)
		printString(scrollNames[amount&7])
		selectFont(0)
		printMessage(msgChar33Newline)
		scrolls[amount]++
		if scrolls[amount] == maxItem+1 {
			scrolls[amount] = maxItem
		}
	case actorGold:
		printNumber(amount, 1, ' ')
		printMessage(msgGoldLine2)
		increaseInt(&gold, amount, maxFood)
	case actorPotion:
		printMessage(msgASp2)
		printString(potionNames[amount])
		printMessage(msgPotionLine2)
		potions[amount]++
		if potions[amount] == maxItem+1 {
			potions[amount] = maxItem
		}
	case actorWeapon, actorShield, actorHelm, actorRing, actorArmour, actorAmulet:
		if amount == arrowsItem || amount == quarrelsItem {
			increaseByte(&equipment[amount], plentifulAmmo, maxItem)
		} else {
			equipment[amount]++
			if equipment[amount] == maxItem+1 {
				equipment[amount] = maxItem
			}
		}
		printString(itemNames[amount])
		printMessage(msgChar33Newline)
	case actorShard:
		shards[amount&shardMask] = owned
		printMessage(msgTheShardOfLine)
		switch amount & shardMask {
		case 0:
			printMessage(msgFalsehoodLine)
		case 1:
			printMessage(msgHatredLine)
		default:
			printMessage(msgCowardiceLine)
		}
	case actorCrown:
		crown = owned
		printMessage(msgTheCrownOfLordBritishLine)
		markKilled(personShownBy(slot))
		removePerson(personShownBy(slot))
	case actorSceptre:
		sceptre = owned
		printMessage(msgTheSceptreOfLordBritishLine)
	case actorAmuletLb:
		amulet = owned
		printMessage(msgTheAmuletOfLordBritishLine)
	default:
		printMessage(msgNothingToGetLine)
		return
	}
	if slot < actorSlots {
		setActor(0, 0, 0, 0, 0, 0, slot)
	}
	viewChanged |= 2
	statusChanged = true
}

const (
	dungeonPotions = 5
	dungeonScrolls = 6
	dungeonGold    = 1
)

// getInDungeon takes the contents of an open dungeon chest.
// C: SJOG_179e_GetInDungeon
func getInDungeon() {
	printMessage(msgGetLine)
	cell := &dungeonMap[int(level)*dungeonSize*dungeonSize+int(mapY&7)*dungeonSize+int(mapX&7)]
	kind := int(*cell)
	switch kind & cellKindMask {
	case cellChest:
		printMessage(msgMustOpenFirstLine)
	case cellOpenChest:
		*cell &= cellSeen
		printMessage(msgContentsOfChestYouFindLine)
		for i := range dungeonLoot {
			if int(dungeonOdds[i]) > randomRange(1, int(level)*4+4) {
				continue
			}
			switch i {
			case dungeonPotions:
				getObject(actorPotion, randomRange(0, 7), actorSlots)
			case dungeonScrolls:
				getObject(actorScroll, randomRange(0, 7), actorSlots)
			default:
				var amount int
				if i == dungeonGold {
					amount = randomRange(1, int(level)<<3)
				} else {
					amount = randomRange(1, int(dungeonAmounts[i]))
				}
				getObject(int(dungeonLootTile[i]), amount, actorSlots)
			}
		}
	default:
		printMessage(msgNotHereLine3)
	}
}

const (
	torchHolderLeft  = mapTile176
	torchHolderRight = mapTile177
	borrowedTorch    = 100
	tableEmpty       = mapTileTable149
	tablePlateBelow  = mapTileTable154
	tablePlateAbove  = mapTileTable155
	tablePlates      = mapTileTable156
)

// eat takes a bite from a table: a ration more, a little karma less.
func eat(x, y int, left byte) {
	*tileAt(x, y) = left
	viewChanged |= 2
	printMessage(msgMmmmmLine)
	increaseInt(&food, 1, maxFood)
	if karma != 0 {
		karma--
	}
}

// getCommand picks up what lies beside the party.
// C: SJOG_18ce_GetCmd
func getCommand() {
	if inDungeon() {
		getInDungeon()
		return
	}
	if !selectDirection() {
		return
	}
	dx, dy := int(moveX), int(moveY)
	x, y := dx+int(mapX), dy+int(mapY)
	printMessage(msgNewline)
	slot := 1
	tile := 0
	for slot = 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		if int(a.x) != x || int(a.y) != y || (mapID < firstMapCombat && a.z != level) {
			continue
		}
		tile = int(a.tile)
		if tile < actorHorse || tile == actorMoonstone || tile == actorCarpet || tile&shapeMask == actorShard {
			break
		}
	}
	if slot < actorSlots {
		getObject(tile, int(actors[slot].hull), slot)
		return
	}
	switch *tileAt(x, y) {
	case torchHolderLeft, torchHolderRight:
		*tileAt(x, y) = openFloor
		viewChanged = 1
		if mapID < firstMapCombat {
			updateVisibility()
		}
		torchLight = borrowedTorch
		printMessage(msgBorrowedLine)
		sweepTone(keyBreakFrom, keyBreakTo, 1, keyBreakTime)
		updateFrame()
	case mapTileCrops:
		*tileAt(x, y) = mapTileCropsPicked
		viewChanged |= 2
		printMessage(msgCropsPickedLine)
		increaseInt(&food, 1, maxFood)
		statusChanged = true
		if karma != 0 {
			karma--
		}
	case tablePlateBelow:
		if dy != 1 {
			printMessage(msgCanTReachPlateLine)
			break
		}
		eat(x, y, tableEmpty)
		statusChanged = true
	case tablePlateAbove:
		if dy != -1 {
			printMessage(msgCanTReachPlateLine)
			break
		}
		eat(x, y, tableEmpty)
	case tablePlates:
		if dx == 1 || dx == -1 {
			printMessage(msgCanTReachPlateLine)
			break
		}
		left := *tileAt(x, y)
		if dy == 1 {
			left = tablePlateAbove
		}
		if dy == -1 {
			left = tablePlateBelow
		}
		eat(x, y, left)
	default:
		printMessage(msgNothingToGetLine)
	}
}

// unequipAll takes item off every member; it counts those who wore it.
// C: SJOG_1b34
func unequipAll(item int) int {
	count := 0
	for i := range int(partySize) {
		if unequip(i, item) {
			count++
		}
	}
	return count
}

// countSides leaves the number of fighters the computer steers in moveX and the player's in
// moveY.
// C: SJOG_1b6c
func countSides() {
	moveX, moveY = 0, 0
	for i := range combatantSlots {
		if combatants[i].flags != 0 && combatants[i].flags&fighterDead == 0 {
			if computerControlled(i) {
				moveX++
			} else {
				moveY++
			}
		}
	}
}

const (
	sameExitTone = 165
	sameExitTime = 200
	leaveFrom    = 1200
	leaveTo      = 2000
	leaveTime    = 40
	exitUp       = 5
	exitDown     = 6
	exitMirror   = 77
	combatEdge   = 10
	shipCombat   = 128
)

// leaveCombat takes combatant slot off the map through the exit given.
// C: SJOG_1bb2_CombatExit
func leaveCombat(slot, exit int) bool {
	if partyTile&monsterMask == actorFrigate32 {
		printMessage(msgStayWithShipLine)
		return false
	}
	if combatants[slot].flags&fighterPlayer != 0 {
		if exitDirection == 0 {
			exitDirection = byte(exit)
		} else if exit != int(exitDirection) && combatKind&shipCombat != 0 {
			printMessage(msgAllMustUseTheSameExit)
			tone(sameExitTone, sameExitTime)
			return false
		}
	}
	focusBlink = 1
	countSides()
	if moveX == 0 {
		printMessage(msgLeaveLine)
	} else {
		printMessage(msgEscapeLine)
	}
	sweepTone(leaveFrom, leaveTo, 1, leaveTime)
	activeMember = noMember
	removeFromCombat(-slot - 1)
	showStatus()
	updateFrame()
	return true
}

const combatTrapsKind = 130

// moveInCombat moves combatant slot the way given, or off the map.
// C: SJOG_1c56_CombatMovePlayer
func moveInCombat(slot, way int) bool {
	dx, dy := 0, 0
	f := &combatants[slot]
	a := &actors[f.actorSlot]
	switch way {
	case dirUp:
		dy = -1
		printMessage(msgNorthLine)
	case dirDown:
		dy = 1
		printMessage(msgSouthLine)
	case dirRight:
		dx = 1
		printMessage(msgEastLine)
	case dirLeft:
		dx = -1
		printMessage(msgWestLine)
	}
	x, y := int(f.x)+dx, int(f.y)+dy
	if x > combatEdge || y > combatEdge || x < 0 || y < 0 {
		return leaveCombat(slot, way)
	}
	if fighterMayStand(int(a.tile), x, y) {
		f.x, a.x = byte(x), byte(x)
		f.y, a.y = byte(y), byte(y)
		footstep()
		if combatKind&combatTrapsKind != 0 {
			blockedInCombat(x, y)
		}
		return true
	}
	printMessage(msgBlockedLine)
	tone(sameExitTone, sameExitTime)
	clearKeyboard()
	return false
}

const (
	dungeonCombatKind = 2
	ladderHole        = mapTile134
	climbable         = mapTile76
)

// klimbInCombat climbs a ladder off the map, or a rope onto a ledge.
// C: SJOG_1d6a_CombatKlimb
func klimbInCombat() bool {
	printMessage(msgKlimb)
	tile := int(*tileAt(int(mapX), int(mapY)))
	if combatKind&dungeonCombatKind != 0 && tile == mapTileLadderUp && moongateFoe != 0 {
		printMessage(msgUD)
		for key := 0; key == 0; {
			key = getChar()
			switch key {
			case keyDown, 'D':
				tile = mapTileLadderDown
			case keyUp, 'U':
			default:
				key = 0
			}
		}
	}
	if tile == mapTileLadderUp {
		printMessage(msgUpLine)
		return leaveCombat(int(combatant), exitUp)
	}
	if (tile == ladderHole && combatKind&shipCombat != 0) || tile == mapTileLadderDown {
		printMessage(msgDownLine)
		return leaveCombat(int(combatant), exitDown)
	}
	if selectDirection() {
		x, y := int(mapX)+int(moveX), int(mapY)+int(moveY)
		if *tileAt(x, y) != climbable || topActorAt(x, y, 0) != 0 {
			printMessage(msgWhatLine)
			return false
		}
		f := &combatants[combatant]
		a := &actors[f.actorSlot]
		f.x, a.x = byte(x), byte(x)
		f.y, a.y = byte(y), byte(y)
		blockedInCombat(x, y)
	}
	return true
}

const (
	mirrorRow    = 2
	mirrorOffset = 16
)

// checkMirror absorbs the combatant who steps before the mirror of the endgame.
// C: SJOG_1ea4_CheckMirror
func checkMirror() {
	f := &combatants[combatant]
	if f.flags == 0 || f.flags&fighterDead != 0 || f.y != mirrorRow ||
		actorTiles[int(f.x)+mirrorOffset]&shapeMask != actorMirror {
		return
	}
	exitDirection = exitMirror
	printChar('\n')
	printFighterName(int(combatant))
	printMessage(msgIsAbsorbedLine)
	sweepTone(leaveFrom, leaveTo, 1, leaveTime)
	activeMember = noMember
	showStatus()
	removeFromCombat(-int(combatant) - 1)
	updateFrame()
}

// Answers to a command that does nothing in combat.
const (
	answerWhat   = 1
	answerHere   = 2
	answerFunny  = 3
	answerToneHi = 220
	answerToneLo = 150
	answerTime   = 150
)

// refuseInCombat answers a command not usable in combat.
// C: SJOG_1f26_CombatMiscCmd
func refuseInCombat(name, answer int) bool {
	printMessage(name)
	switch answer {
	case answerWhat:
		printMessage(msgWhat2)
	case answerHere:
		printMessage(msgNotHere)
	case answerFunny:
		printMessage(msgFunnyNoResponse)
	}
	printChar('\n')
	tone(answerToneHi, answerTime)
	tone(answerToneLo, answerTime)
	return true
}

// setActiveInCombat makes member who the one the party's moves follow.
// C: SJOG_1f7a_CombatSetActivePlayer
func setActiveInCombat(who int) bool {
	able := false
	printMessage(msgSetActivePlrLine)
	slot := 0
	for slot = 0; slot < combatantSlots; slot++ {
		flags := combatants[slot].flags
		if flags&fighterPlayer != 0 && !computerControlled(slot) && int(combatants[slot].entity) == who {
			able = flags&(fighterDead|fighterAsleep|fighterFlag4) == 0
			break
		}
	}
	if !able {
		printMessage(msgInvalidLine)
		return false
	}
	activeMember = byte(who)
	printFighterName(slot)
	printChar('\n')
	showStatus()
	return true
}

// endCombatTurn ends a combatant's turn: rings, fading foes and the protection spell.
// C: SJOG_2012
func endCombatTurn() {
	startRings(int(combatant))
	fadeSummoned()
	if protection != 0 && protection != owned {
		protection--
		if protection == 0 {
			spellIcon = 0
			statusChanged = true
		}
	}
}

const pirateRank = 8

// afterCombat puts the map's actors back after a fight with the one in slot.
// C: SJOG_203e
func afterCombat(slot int) {
	if exitDirection == exitMirror {
		endgame()
	}
	actors = savedActors
	if combatKind&combatTrapsKind != 0 {
		return
	}
	if slot < actorSlots {
		a := &actors[slot]
		if battleWon != 0 && a.tile&shapeMask == actorPirate {
			a.tile -= pirateRank
			a.animTile -= pirateRank
			a.hull = maxItem
			a.target = 2
		} else {
			a.tile, a.animTile, a.x, a.y, a.z = 0, 0, 0, 0, 0
		}
	}
	printChar('\n')
	mapID = savedMapID
}

// cannotStand tells whether combatant slot cannot stand at x, y.
// C: SJOG_20d8
func cannotStand(x, y, slot int) bool {
	f := &combatants[slot]
	if x > -1 && x <= combatEdge && y > -1 && y <= combatEdge {
		return !fighterMayStand(int(actors[f.actorSlot].tile), x, y)
	}
	return f.flags&fighterFlag2 == 0
}

// hemmedIn tells whether combatant slot cannot step anywhere.
// C: SJOG_2148
func hemmedIn(slot int) bool {
	blocked := 0
	x, y := int(combatants[slot].x), int(combatants[slot].y)
	for i := range 4 {
		switch i {
		case 0:
			y++
		case 2:
			x -= 2
		case 1, 3:
			y--
			x++
		}
		if cannotStand(x, y, slot) {
			blocked++
		}
	}
	return blocked == 4
}

const (
	passOutFreq     = 3100
	passOutDuration = 30000
	passOutWidth    = 1000
	passOutStep     = 2
	charmItem       = 35
	partyFighters   = 6
)

// passOut ends the charm of the first charmed member; it returns their slot (-1 none).
// C: SJOG_21ce
func passOut() int {
	for i := range partyFighters {
		f := &combatants[i]
		if f.flags&fighterCharmed == 0 || f.flags&fighterPlayer == 0 {
			continue
		}
		f.flags &^= fighterCharmed
		printString(party[f.entity].nameString())
		printMessage(msgPassesOut)
		pulse(passOutFreq, 1, passOutDuration, passOutWidth, passOutStep)
		unequip(int(f.entity), charmItem)
		sleepFighter(i)
		return i
	}
	return -1
}

const (
	summonedShape = actor232
	fadeOdds      = 16
)

// fadeSummoned sends away, now and then, the creatures a spell summoned.
// C: COMSUBS_0056
func fadeSummoned() {
	for i := range actorSlots {
		if actors[i].tile&shapeMask == summonedShape && randomRange(0, lowByteMask) < fadeOdds {
			removeFromCombat(i + 1)
		}
	}
}
