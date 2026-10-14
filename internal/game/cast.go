package game

// The Cast and Use commands: the forty-eight spells, scrolls, potions and the party's
// special items (cast.c).

// Addresses of the cast tables.
const (
	spellPlacesAddress    = 7312  // C: D_1c90
	boltSlopesAddress     = 7408  // C: D_1cf0
	fieldWeaponsAddress   = 17810 // C: D_4592
	dungeonFieldsAddress  = 17814 // C: D_4596
	flameXAddress         = 18562 // C: D_4882
	flameYAddress         = 18566 // C: D_4886
	flameMapAddress       = 18570 // C: D_488a
	flameLevelAddress     = 18574 // C: D_488e
	shadowlordBitsAddress = 18578 // C: D_4892
	fieldKinds            = 4
	boltRays              = 21
)

// Where a spell works: bits of spellPlaces.
const (
	spellInCombat     = 1
	spellInDungeon    = 2
	spellInSettlement = 4
	spellOutdoors     = 8
)

var (
	spellPlaces    []byte   // C: D_1c90
	boltSlopes     []uint16 // C: D_1cf0
	fieldWeapons   []byte   // C: D_4592 (the field a combat spell throws)
	dungeonFields  []byte   // C: D_4596 (the field a dungeon spell lays)
	flameX         []byte   // C: D_4882 (the flames the shards are cast into)
	flameY         []byte   // C: D_4886
	flameMap       []byte   // C: D_488a
	flameLevel     []byte   // C: D_488e
	shadowlordBits []byte   // C: D_4892 (a shadowlord's bit in the killed people)

	// boltCounters step each ray of a bolt along its slope (D_a9d0).
	boltCounters [boltRays]uint16 // C: D_a9d0
)

// loadSpellCastTables reads the tables of the Cast and Use commands.
func loadSpellCastTables() {
	spellPlaces = segment.Bytes(spellPlacesAddress, spellCount)
	boltSlopes = segment.Words(boltSlopesAddress, boltRays)
	fieldWeapons = segment.Bytes(fieldWeaponsAddress, fieldKinds)
	dungeonFields = segment.Bytes(dungeonFieldsAddress, fieldKinds)
	flameX = segment.Bytes(flameXAddress, shadowlordCount)
	flameY = segment.Bytes(flameYAddress, shadowlordCount)
	flameMap = segment.Bytes(flameMapAddress, shadowlordCount)
	flameLevel = segment.Bytes(flameLevelAddress, shadowlordCount)
	shadowlordBits = segment.Bytes(shadowlordBitsAddress, shadowlordCount)
}

// Places and things the spells care about.
const (
	mapDoom             = 40
	dungeonCellContent  = 247 // all of a dungeon cell but its "seen" bit
	dungeonCellSeen     = 8
	dungeonCellKind     = 240
	dungeonChest        = 64
	dungeonChestTrapped = 1
	dungeonOpenChest    = 112
	foeMageType         = 47 // foes no spell of this kind touches
	foeShadowlordType   = 14
	foeShadowlordType2  = 15
	summonedFoes        = 4
	// the foes Kal Xen calls: rats, bats, spiders, snakes
	kalXenRat    = 20
	kalXenBat    = 22
	kalXenSpider = 21
	kalXenSnake  = 34
	kalXenTries  = 8
	inPorTries   = 7
	turnTimerHit = 128
	chestLocked  = 128 // a chest actor's locked bit
	removedIcon  = 0
	amuletIcon   = 14
	crownIcon    = 28
	plansDouble  = 128
	sweepFrom    = 800
	sweepTo      = 2000
	sweepTime    = 50
)

// The sounds of the spells.
const (
	poofFrom      = 1200
	poofTo        = 2000
	poofTime      = 40
	shardFreq     = 2640
	shardTime     = 200
	shardLow      = 2000
	shardHigh     = 25000
	shardStep     = 50
	sceptreFreq   = 5200
	sceptreTime   = 50000
	sceptreWidth  = 5000
	sceptreNoise  = 10
	sceptreLength = 3000
	sceptreLimit  = 2000
	boltNoiseRate = 800
	boltNoiseTime = 20800
	sleepBoltTime = 16000
	poisonBolt    = 19200
	boltNoiseTop  = 700
)

// The kinds of bolt: sleep, poison, fire and death.
const (
	boltSleep = 1 + iota
	boltPoison
	boltFire
	boltDeath
)

// spellProof tells the foes no charm, sleep or death spell touches: mages and the
// shadowlords.
// C: CAST_0000
func spellProof(slot int) bool {
	kind := combatants[slot].entity
	return kind == foeMageType || kind == foeShadowlordType || kind == foeShadowlordType2
}

// throwField attacks with a field weapon.
// C: CAST_0032
func throwField(weapon int) {
	attackWeapon = byte(weapon)
	attackWith(int(combatant), weapon)
}

// layField lays a field ahead of the party in a dungeon, or throws one in combat.
// C: CAST_004c
func layField(kind int) int {
	if mapID < firstMapCombat {
		if kind == 3 {
			castEffect(4)
		} else {
			castEffect(3)
		}
		x := (demoDeltaX[facing] + int(mapX)) & (dungeonSize - 1)
		y := (demoDeltaY[facing] + int(mapY)) & (dungeonSize - 1)
		cell := dungeonCellAt(x, y, int(level))
		if *cell&dungeonCellContent != 0 {
			return 0
		}
		*cell = *cell&dungeonCellSeen | dungeonFields[kind]
		return -1
	}
	attackWeapon = fieldWeapons[kind]
	attackWith(int(combatant), int(attackWeapon))
	return -1
}

// anZu awakens a member.
// C: CAST_0114_AnZu
func anZu() int {
	who := onWho()
	if who < 0 {
		return -1
	}
	if party[who].status != statusSleeping {
		return 0
	}
	party[who].status = statusGood
	if mapID >= firstMapCombat {
		for slot := range combatantSlots {
			f := &combatants[slot]
			if f.flags&(fighterPlayer|fighterMonster|fighterDead|fighterAsleep) == fighterPlayer|fighterAsleep &&
				int(f.entity) == who {
				wakeFighter(slot)
				break
			}
		}
	}
	castEffect(1)
	statusChanged = true
	return -1
}

// anNox cures poison.
// C: CAST_01ae_AnNox
func anNox() int {
	who := onWho()
	if who < 0 {
		return -1
	}
	if party[who].status != statusPoisoned {
		return 0
	}
	party[who].status = statusGood
	castEffect(1)
	statusChanged = true
	return 1
}

// mani heals a little.
// C: CAST_01fa_Mani
func mani() int {
	result := -1
	if who := onWho(); who >= 0 {
		result = 0
		if healMember(who) {
			result = 1
		}
	}
	if result != 0 {
		castEffect(1)
	}
	return result
}

// anYlem makes a piece of furniture vanish.
// C: CAST_0230_AnYlem
func anYlem() int {
	if castDirection() == 0 {
		return -1
	}
	castEffect(1)
	at := tileAt(int(moveX), int(moveY))
	switch *at {
	case mapTile91, mapTileChair144, mapTileChair145, mapTileChair146, mapTileChair147, mapTileMirror,
		mapTileDesk, mapTileBarrel, mapTileVanity, mapTile169, mapTileDresser, mapTile174, mapTileTrunk:
		*at = mapTile68
		printMessage(msgPoofLine)
		viewChanged |= 2
		updateFrame()
		sweepTone(poofFrom, poofTo, 1, poofTime)
		return -1
	}
	return 0
}

// anSanct opens a chest or unlocks a door.
// C: CAST_02d2_AnSanct
func anSanct() int {
	if mapID > settlements && mapID < firstMapCombat {
		castEffect(2)
		cell := dungeonCellAt(int(mapX), int(mapY), int(level))
		if *cell&dungeonCellKind != dungeonChest {
			cell = dungeonCellAt((demoDeltaX[facing]+int(mapX))&(dungeonSize-1),
				(demoDeltaY[facing]+int(mapY))&(dungeonSize-1), int(level))
		}
		if *cell&dungeonCellKind != dungeonChest {
			return 0
		}
		if *cell&dungeonChestTrapped != 0 {
			printMessage(msgDisarmedLine)
		}
		*cell = *cell&dungeonCellSeen | dungeonOpenChest
		printMessage(msgChestOpenedLine)
		return -1
	}
	if castDirection() == 0 {
		return -1
	}
	at := tileAt(int(moveX), int(moveY))
	if *at == mapTileDoor185 || *at == mapTileDoor187 {
		*at--
		viewChanged |= 2
		castEffect(2)
		return 1
	}
	for i := range actorSlots {
		a := &actors[i]
		if a.tile == actorChest && int(a.x) == int(moveX) && int(a.y) == int(moveY) &&
			(mapID >= firstMapCombat || a.z == level) {
			a.hull &^= chestLocked
			castEffect(2)
			return 1
		}
	}
	return 0
}

// anCorpXen strikes the undead down to one hit point.
// C: CAST_043e_AnCorpXen
func anCorpXen() {
	castEffect(2)
	for slot := range combatantSlots {
		f := &combatants[slot]
		if !spellProof(slot) && f.flags&(fighterPlayer|fighterMonster) == fighterMonster &&
			foeFlags[f.entity]&foeUndead != 0 {
			if !resistsCharm(int(combatant), slot, 0) {
				f.hp = 1
				f.flags |= fighterFlag2
			}
		}
	}
}

// inWis tells where the party is.
// C: CAST_04a4_InWis
func inWis() {
	castEffect(2)
	printLocation()
}

// kalXen calls a small creature to fight for the party.
// C: CAST_04b0_KalXen
func kalXen() int {
	castEffect(2)
	roll := randomRange(0, 15)
	var kind int
	switch {
	case roll < 6:
		kind = kalXenRat
	case roll < 11:
		kind = kalXenBat
	case roll < 14:
		kind = kalXenSpider
	default:
		kind = kalXenSnake
	}
	tile := byte(kind*4 + 64)
	shape := kind*4 + firstActorShape + 64
	for tries := 0; ; {
		if !randomCell() || !fighterMayStand(actor144, int(moveX), int(moveY)) {
			if tries++; tries >= kalXenTries {
				return 0
			}
			continue
		}
		slot := placeFighter(kind, 0, int(moveX), int(moveY), int(level))
		if slot < 0 {
			return 0
		}
		a := &actors[combatants[slot].actorSlot]
		a.tile, a.animTile = actorCircle, actorCircle
		revealTile(shape, int(moveX), int(moveY))
		a.tile, a.animTile = tile, tile
		combatants[slot].flags |= fighterCharmed
		return 1
	}
}

// inManiXen makes food.
// C: CAST_05b4_InManiXen
func inManiXen() int {
	castEffect(2)
	increaseInt(&food, randomRange(1, 3), maxFood)
	statusChanged = true
	return 1
}

// inPor blinks: in combat to a random place, elsewhere a way over the land.
// C: CAST_05dc_InPor
func inPor() int {
	if mapID >= firstMapCombat {
		a := &actors[combatants[combatant].actorSlot]
		if combatKind&2 != 0 {
			return 0
		}
		castEffect(3)
		for range inPorTries {
			if !randomCell() || !fighterMayStand(int(a.tile), int(moveX), int(moveY)) {
				continue
			}
			combatants[combatant].x, combatants[combatant].y = byte(moveX), byte(moveY)
			a.x, a.y = byte(moveX), byte(moveY)
			return 1
		}
		return 0
	}
	if castDirection() == 0 {
		return -1
	}
	castEffect(3)
	stepX := int(moveX) - int(mapX)
	stepY := int(moveY) - int(mapY)
	right := min(int(chunkX)+chunkSpan, mapSpan)
	bottom := min(int(chunkY)+chunkSpan, mapSpan)
	result := 0
	x := int(byte(moveX))
	for y := int(byte(moveY)); int(chunkX) <= x && x < right && int(chunkY) <= y && y < bottom; y += stepY {
		if *tileAt(x, y) == mapTileGrass {
			result = -1
			mapX, mapY = byte(x), byte(y)
			viewChanged = 1
		}
		x += stepX
	}
	if result != 0 {
		redrawWorld()
	}
	return result
}

// The span of the map loaded around the party, and of the whole map.
const (
	chunkSpan = 32
	mapSpan   = 256
)

// quasWis shows the invisible foes.
// C: CAST_074c_QuasWis
func quasWis() {
	castEffect(4)
	for slot := range combatantSlots {
		f := &combatants[slot]
		if f.flags != 0 && f.flags&fighterPlayer == 0 && f.flags&fighterInvisible != 0 {
			f.flags &^= fighterInvisible
			a := &actors[f.actorSlot]
			a.animTile = a.tile
			placeActors()
			drawView()
		}
	}
}

// betInXen calls a swarm of insects to fight for the party.
// C: CAST_07b4_BetInXen
func betInXen() int {
	castEffect(5)
	found := false
	for range kalXenTries {
		if randomCell() && fighterMayStand(actor188, int(moveX), int(moveY)) {
			found = true
			break
		}
	}
	if !found {
		return 0
	}
	result := 0
	for range summonedFoes {
		slot := placeFighter(insectFoe, 0, int(moveX), int(moveY), int(level))
		if slot < 0 {
			break
		}
		combatants[slot].flags |= fighterCharmed
		result = 1
	}
	return result
}

// anExPor locks a door magically.
// C: CAST_0846_AnExPor
func anExPor() int {
	if castDirection() == 0 {
		return -1
	}
	castEffect(5)
	at := tileAt(int(moveX), int(moveY))
	switch *at {
	case mapTileDoor184, mapTileDoor185:
		*at = mapTile151
	case mapTileDoor186, mapTileDoor187:
		*at = mapTile152
	default:
		return 0
	}
	viewChanged |= 2
	return 1
}

// maniVas heals a member fully.
// C: CAST_08ac_ManiVas
func maniVas() int {
	who := onWho()
	if who < 0 {
		return -1
	}
	if party[who].status == statusDead || (mapID >= firstMapCombat && battleWon == 0) {
		return 0
	}
	party[who].hp = party[who].maxHP
	castEffect(5)
	statusChanged = true
	return 1
}

// inPorVasYlem shakes the earth under the foes.
// C: CAST_091e_InPorVasYlem
func inPorVasYlem(caster int) {
	castEffect(6)
	shakeScreen()
	for slot := range combatantSlots {
		if combatants[slot].flags == 0 || !computerControlled(slot) {
			continue
		}
		if combatantStat(slot, statDexterity) <= randomDamage() {
			hitActor(slot)
			increaseInt(&party[caster].exp, damage(slot, randomRange(1, 20)), maxExperience)
			reportHit(slot, int(combatant))
		}
	}
}

// anExXen charms a creature, or frees one charmed.
// C: CAST_09a0_AnExXen
func anExXen() int {
	printMessage(msgCreatureSp)
	if aim(int(combatant), aimAnywhere) == 0 {
		return -1
	}
	castEffect(6)
	if *tileAt(int(crosshairX), int(crosshairY))&^1 == mapTile132 {
		return 0
	}
	slot := fighterAt(int(crosshairX), int(crosshairY))
	if slot >= 0 && !spellProof(slot) && computerControlled(slot) && !resistsCharm(int(combatant), slot, 0) {
		f := &combatants[slot]
		f.flags ^= fighterCharmed
		if f.flags&fighterPlayer != 0 {
			party[f.entity].status = statusGood
			showStatus()
		}
		printFighterName(slot)
		printMessage(msgCharmedLine)
		return -1
	}
	return 0
}

// aimAnywhere lets the crosshair go anywhere on the combat map.
const aimAnywhere = 15

// betRelXen turns a creature into a rat.
// C: CAST_0a5c_BetRelXen
func betRelXen() int {
	printMessage(msgCreatureSp)
	if aim(int(combatant), aimAnywhere) == 0 {
		return -1
	}
	castEffect(6)
	slot := fighterAt(int(crosshairX), int(crosshairY))
	if slot < 0 {
		return 0
	}
	if !spellProof(slot) && !resistsCharm(int(combatant), slot, 0) {
		x, y := int(combatants[slot].x), int(combatants[slot].y)
		removeFromCombat(-slot - 1)
		placeFighter(kalXenRat, 0, x, y, int(level))
		return 1
	}
	return 0
}

// lorSanct makes the caster invisible.
// C: CAST_0afe_LorSanct
func lorSanct() int {
	actors[combatants[combatant].actorSlot].animTile = actor29
	combatants[combatant].flags |= fighterInvisible
	castEffect(7)
	return 1
}

// inQuasXen clones a creature.
// C: CAST_0b28_InQuasXen
func inQuasXen() int {
	printMessage(msgCreatureSp)
	if aim(int(combatant), aimAnywhere) == 0 {
		return -1
	}
	castEffect(7)
	source := fighterAt(int(crosshairX), int(crosshairY))
	if source < 0 {
		return 0
	}
	freeActor, freeFighter := -1, -1
	for i := range actorSlots {
		if actors[i].tile == 0 {
			freeActor = i
			break
		}
	}
	if freeActor < 0 {
		return 0
	}
	for i := range combatantSlots {
		if combatants[i].flags == 0 {
			freeFighter = i
			break
		}
	}
	if freeFighter < 0 {
		return 0
	}
	combatants[freeFighter] = combatants[source]
	combatants[freeFighter].actorSlot = byte(freeActor)
	actors[freeActor] = actors[combatants[source].actorSlot]
	// The original tries random cells until one is free.
	for !randomCell() || !fighterMayStand(actorAvatar, int(moveX), int(moveY)) {
	}
	combatants[freeFighter].x, combatants[freeFighter].y = byte(moveX), byte(moveY)
	actors[freeActor].x, actors[freeActor].y = byte(moveX), byte(moveY)
	return 1
}

// corpInQuas strikes all foes down to one hit point.
// C: CAST_0c98_CorpInQuas
func corpInQuas() {
	castEffect(7)
	for slot := range combatantSlots {
		f := &combatants[slot]
		if f.flags&(fighterPlayer|fighterMonster) == fighterMonster && !spellProof(slot) &&
			!resistsCharm(int(combatant), slot, 0) {
			f.hp = 1
			f.flags |= fighterFlag2
		}
	}
}

// porRelVas sends the party to a moonstone's place.
// C: CAST_0cf0_PorRelVas
func porRelVas() int {
	if partyTile&frigateMask == actorFrigate32 {
		return 0
	}
	printMessage(msgToPhaseSp)
	c := getChar()
	if c >= ' ' {
		printChar(c)
	}
	printChar('\n')
	if c >= '1' && c <= '8' {
		castEffect(8)
		if travelToMoonstone(c-'1') != 0 {
			return -1
		}
	}
	return 0
}

// frigateMask keeps the ship tiles of the party tile.
const frigateMask = 240

// anTym stops time, unless a shadowlord is near.
// C: CAST_0d4c_AnTym
func anTym() int {
	castEffect(8)
	for i := range actorSlots {
		if actors[i].tile == actorShadowlord {
			printMessage(msgMagicAbsorbedLine)
			pulse(absorbFreq, 1, absorbTime, absorbWidth, absorbStep)
			return 0
		}
	}
	spellIcon = iconTimeStop
	protection = 10
	showStatus()
	return -1
}

// castCommand casts a spell. It returns 0 when the party left by the spell.
// C: CAST_0dba_CastSpellCmd
func castCommand() int {
	keep := 1
	rangeX, rangeY = 1, 1
	caster := chooseMember()
	if caster < 0 {
		return caster
	}
	printMessage(msgSpellName)
	spell := selectSpell()
	switch spell {
	case -1:
		printMessage(msgNoneLine)
		return keep
	case -2:
		printMessage(msgNoEffectLine3)
		return keep
	}
	circle := spell/6 + 1
	rangeX = byte(circle)
	allowed := true
	switch {
	case mapID == 0:
		allowed = spellPlaces[spell]&spellOutdoors != 0
	case mapID >= firstMapCombat:
		allowed = spellPlaces[spell]&spellInCombat != 0
	default:
		if (mapID == shadowlordStronghold && crown == 0) || mapID == mapStonegate {
			printMessage(msgAbsorbedLine)
			pulse(absorbFreq, 1, absorbTime, absorbWidth, absorbStep)
			return keep
		}
		if mapID <= settlements {
			allowed = spellPlaces[spell]&spellInSettlement != 0
		} else {
			allowed = spellPlaces[spell]&spellInDungeon != 0
		}
	}
	if !allowed {
		printMessage(msgNotHereLine3)
		sweepTone(sweepFrom, sweepTo, 1, sweepTime)
		return keep
	}
	if spellMixtures[spell] == 0 {
		printMessage(msgNoneMixedLine)
		return keep
	}
	spellMixtures[spell]--
	result := -1
	m := &party[caster]
	switch {
	case int(m.magic) < circle:
		printMessage(msgMPTooLowLine)
		result = 0
	case int(m.level) < circle:
		m.magic -= byte(circle)
		result = 0
	default:
		m.magic -= byte(circle)
		result = castSpell(spell, caster, &keep)
	}
	switch result {
	case 1:
		printMessage(msgSuccessLine)
	case 0:
		printMessage(msgFailedLine)
		sweepTone(sweepFrom, sweepTo, 1, sweepTime)
	}
	return keep
}

// castSpell works one spell, returning 1 success, 0 failure, -1 nothing to say.
// It clears *keep when the party left the map.
func castSpell(spell, caster int, keep *int) int {
	result := -1
	switch spell {
	case 0: // In Lor
		setLightSpell(100)
		castEffect(1)
	case 1: // Grav Por
		throwField(48)
	case 2:
		result = anZu()
	case 3:
		result = anNox()
	case 4:
		result = mani()
	case 5:
		result = anYlem()
	case 6:
		result = anSanct()
	case 7:
		anCorpXen()
	case 8: // Rel Hur
		changeWind(castDirection(), false)
	case 9:
		inWis()
	case 10:
		result = kalXen()
	case 11:
		result = inManiXen()
	case 12: // Vas Lor
		setLightSpell(255)
		castEffect(3)
	case 13: // Vas Flam
		throwField(49)
	case 14: // In Flam Grav
		result = layField(0)
	case 15: // In Nox Grav
		result = layField(1)
	case 16: // In Zu Grav
		result = layField(2)
	case 17:
		result = inPor()
	case 18: // An Grav
		result = dispelField(true)
	case 19: // In Sanct
		setSpellIcon('P', 20, 4)
	case 20: // In Sanct Grav
		result = layField(3)
	case 21, 22: // Uus Por, Des Por
		if mapID == mapDoom {
			return 0
		}
		castEffect(4)
		way := -1
		if spell == 22 {
			way = 1
		}
		if dungeonKlimb(way, 1) != 0 {
			leaveDungeon()
		}
	case 23:
		quasWis()
	case 24:
		result = betInXen()
	case 25:
		result = anExPor()
	case 26: // In Ex Por
		result = unlockMagically()
		if result != -1 {
			castEffect(5)
		}
	case 27:
		result = maniVas()
	case 28: // In Zu
		castBolt(int(combatant), boltSleep, magentaColour)
	case 29: // Rel Tym
		setSpellIcon('Q', 30, 5)
	case 30:
		inPorVasYlem(caster)
	case 31: // An Quas Wis
		setSpellIcon('C', 20, 6)
	case 32: // negate magic
		setSpellIcon('N', 10, 6)
	case 33: // An Wis Ylem
		castEffect(6)
		refreshLitView()
	case 34:
		result = anExXen()
	case 35:
		result = betRelXen()
	case 36:
		result = lorSanct()
	case 37: // Corp Xen
		throwField(50)
	case 38:
		result = inQuasXen()
	case 39: // In Quas Wis
		castEffect(7)
		if mapID <= settlements {
			viewGem(int(mapX), int(mapY))
		} else {
			viewGemInDungeon()
		}
	case 40: // In Nox Hur
		castBolt(int(combatant), boltPoison, greenColour)
	case 41:
		corpInQuas()
	case 42: // In Mani Corp
		result = resurrect(onWho(), 0)
		showStatus()
	case 43: // Kal Xen Corp
		result = summonDaemon(false)
	case 44: // In Corp Hur
		castBolt(int(combatant), boltDeath, blueColour)
	case 45: // In Flam Hur
		castBolt(int(combatant), boltFire, redColour)
	case 46:
		result = porRelVas()
		if result != 0 {
			*keep = 0
		}
	case 47:
		result = anTym()
	}
	return result
}

// useScroll reads a scroll.
// C: CAST_11de_UseScroll
func useScroll(kind int) int {
	result := 1
	scrolls[kind]--
	printMessage(msgScrollLine)
	switch kind {
	case 0:
		setLightSpell(240)
		printMessage(msgLightLine)
		castEffect(0)
	case 1:
		printMessage(msgWindChangeLine)
		way := castDirection()
		if mapID <= settlements {
			changeWind(way, true)
		} else {
			result = 0
		}
	case 2:
		printMessage(msgProtectionLine)
		setSpellIcon('P', 100, 2)
	case 3:
		printMessage(msgNegateMagicLine)
		setSpellIcon('N', 20, 3)
	case 4:
		printMessage(msgViewLine)
		if mapID >= firstMapCombat {
			printMessage(msgNotHereLine3)
			break
		}
		castEffect(4)
		if mapID <= settlements {
			viewGem(int(mapX), int(mapY))
		} else {
			viewGemInDungeon()
		}
	case 5:
		printMessage(msgSummonDaemonLine)
		if mapID >= firstMapCombat {
			result = summonDaemon(true)
		} else {
			printMessage(msgNotHereLine3)
		}
	case 6:
		printMessage(msgResurrectionLine)
		if mapID < firstMapCombat {
			result = resurrect(onWho(), 1)
			showStatus()
		} else {
			printMessage(msgNotHereLine3)
		}
	case 7:
		if mapID == mapStonegate || mapID == mapDoom {
			printMessage(msgNoEffectLine3)
			sweepTone(sweepFrom, sweepTo, 1, sweepTime)
			break
		}
		printMessage(msgNegateTimeLine)
		setSpellIcon(iconTimeStop, 20, 7)
	}
	return result
}

// usePotion drinks a potion. One in sixteen is sleep instead, one in sixteen any.
// C: CAST_135a_UsePotion
func usePotion(kind int) int {
	result := 1
	potions[kind]--
	printMessage(msgPotionLine)
	var who int
	if mapID >= firstMapCombat {
		who = int(combatants[combatant].entity)
	} else {
		who = onWho()
	}
	if who < 0 {
		return who
	}
	castEffect(kind)
	switch randomRange(0, 15) {
	case 0:
		kind = 4
	case 1:
		kind = randomRange(0, 7)
	}
	m := &party[who]
	switch kind {
	case 0: // awaken
		if m.status != statusSleeping {
			return 0
		}
		m.status = statusGood
		if mapID < firstMapCombat {
			showStatus()
			break
		}
		f := &combatants[combatant]
		if who == int(f.entity) &&
			f.flags&(fighterPlayer|fighterMonster|fighterDead|fighterAsleep) == fighterPlayer|fighterAsleep {
			wakeFighter(int(combatant))
		} else {
			result = 0
		}
	case 1: // heal
		if !healMember(who) {
			return 0
		}
		printMessage(msgHealedLine)
		showStatus()
	case 2: // cure
		if m.status != statusPoisoned {
			return 0
		}
		m.status = statusGood
		printMessage(msgPoisonCuredLine)
		showStatus()
	case 3: // poison
		if m.status != statusGood {
			return 0
		}
		m.status = statusPoisoned
		printMessage(msgPoisonedLine2)
		showStatus()
	case 4: // sleep
		if m.status != statusGood {
			return 0
		}
		if mapID < firstMapCombat {
			m.status = statusSleeping
		} else {
			sleepFighter(who)
		}
		printMessage(msgSleptLine)
		showStatus()
	case 5: // transform
		if mapID < firstMapCombat {
			printMessage(msgNoNoticeableEffectNowLine)
			break
		}
		printMessage(msgPoofLine2)
		a := &actors[combatants[combatant].actorSlot]
		a.tile, a.animTile = actor144, actor144
	case 6: // invisibility
		if mapID < firstMapCombat {
			printMessage(msgNoNoticeableEffectNowLine)
			break
		}
		combatants[combatant].flags |= fighterInvisible
		a := &actors[combatants[combatant].actorSlot]
		printMessage(msgInvisibleLine)
		a.tile, a.animTile = actor29, actor29
	case 7: // x-ray
		if mapID <= settlements {
			refreshLitView()
		} else {
			printMessage(msgNoNoticeableEffectNowLine)
		}
	}
	return result
}

// useMoonstone buries a moonstone where the party stands.
// C: CAST_153c_UseMoonstone
func useMoonstone(stone int) {
	tile := int(*tileAt(int(mapX), int(mapY)))
	printMessage(msgMoonstoneSp)
	if mapID <= settlements && (tile == mapTileCropsPicked || tile == mapTileCrops ||
		(tile > mapTileWater3 && tile < mapTile11)) {
		printMessage(msgBuriedLine)
		moonstoneX[stone] = mapX
		moonstoneY[stone] = mapY
		moonstoneState[stone] = mapID
		moonstoneLevel[stone] = level
	} else {
		printMessage(msgCannotBeBuriedHereLine)
	}
}

// useGemShard casts a shard into its flame, dooming its shadowlord when it stands there.
// C: CAST_15b4_UseGemShard
func useGemShard(shard int) {
	printMessage(msgGemShardThouDostHoldAbove)
	printMessage([]int{msgFalsehood, msgHatred, msgCowardice}[shard])
	for w := shardLow; w < shardHigh; w += shardStep {
		pulse(shardFreq, 1, shardTime, w, 0)
	}
	for w := shardHigh; w > shardLow; w -= shardStep {
		pulse(shardFreq, 1, shardTime, w, 0)
	}
	if flameX[shard] != mapX || flameY[shard] != mapY || flameMap[shard] != mapID || flameLevel[shard] != level {
		printMessage(msgNoEffectLine4)
		sweepTone(sweepFrom, sweepTo, 1, sweepTime)
		return
	}
	sleepFrames(7)
	printMessage(msgAndCastItIntoTheFlame)
	printMessage([]int{msgTruthLine, msgLoveLine, msgCourageLine}[shard])
	shakeScreen()
	shakeScreen()
	shakeScreen()
	sleepFrames(3)
	if actorAt(int(mapX), int(mapY)-1, int(level)) != actorShadowlord {
		return
	}
	if shard != int(lastSpell) {
		return
	}
	removePerson(personShownBy(int(moveX)))
	for range 7 {
		explodeAt(int(mapX), int(mapY)-1)
	}
	shrineVisits[shard] = 255
	shards[shard] = 0
	npcKilled[shadowlordKilledByte] |= shadowlordBits[shard]
	printMessage(msgTheDoomOfTheShadowlordSp)
	printMessage([]int{msgFaulinei, msgAstaroth, msgNosfentor}[shard])
	printMessage(msgIsWroughtLine)
	victoryTune()
}

// shadowlordKilledByte holds the shadowlords' bits among the killed people.
const shadowlordKilledByte = 112

// takeOff removes a worn item when it is the one in effect, returning false then.
// C: CAST_1764
func takeOff(icon int) bool {
	if icon != int(spellIcon) {
		return true
	}
	printMessage(msgRemovedLine)
	spellIcon, protection = removedIcon, 0
	statusChanged = true
	return false
}

// The items of the Use list.
const (
	useScrolls     = 0
	usePotions     = 8
	useCarpet      = 16
	useSkullKey    = 17
	useAmulet      = 18
	useCrown       = 19
	useSceptre     = 20
	useMoonstones  = 21
	useShards      = 29
	useSpyglass    = 32
	usePlans       = 33
	useSextant     = 34
	useWatch       = 35
	useBadge       = 36
	useBox         = 37
	usableItems    = 38
	useListMode    = 85
	useListRows    = 8
	noUsableItem   = -1
	dawn           = 6
	dusk           = 18
	fieldTileMask  = 240
	fieldTiles     = 112
	sceptreDoom    = 40
	itemOwnerParty = 255
)

// useCommand uses an item.
// C: CAST_1792_UseCmd
func useCommand() {
	result := 1
	gatherItems()
	first := nextListed(-1, usableItems, ownedItems[:], itemOwnerParty)
	if first == noUsableItem {
		printMessage(msgNoUsableItemsLine)
		return
	}
	printMessage(msgItemSp)
	clearIconRow()
	selectWindow(1)
	showCaption(msgItems)
	drawListFrame(useListRows)
	item := chooseFromList(first, itemOwnerParty, useListMode)
	clearMiddleBar()
	clearCaption()
	clearIconBar()
	showStatus()
	if item < 0 {
		return
	}
	switch {
	case item < usePotions:
		result = useScroll(item)
	case item < useCarpet:
		result = usePotion(item - usePotions)
	case item >= useMoonstones && item < useShards:
		useMoonstone(item - useMoonstones)
	case item == useCarpet:
		printMessage(msgCarpetLine3)
		switch {
		case mapID > settlements || *tileAt(int(mapX), int(mapY)) == mapTile12:
			printMessage(msgNotHereLine3)
		case partyTile == actorAvatar:
			printMessage(msgBoardedLine)
			partyTile = byte(randomRange(0, 1) + actorFlyingCarpet)
			magicCarpets--
		case partyTile&monsterMask == actorFrigate32:
			printMessage(msgXItShipFirstLine)
		default:
			printMessage(msgOnlyOnFootLine)
		}
	case item == useSkullKey:
		skullKeys--
		printMessage(msgSkullKeyLine)
		if mapID <= settlements || mapID >= firstMapCombat {
			result = unlockMagically()
			if result != 0 && mapID < firstMapCombat {
				explodeAt(int(moveX), int(moveY))
			}
		} else {
			printMessage(msgNotHereLine3)
		}
	case item == useAmulet:
		printMessage(msgAmuletLine)
		if takeOff(amuletIcon) {
			printMessage(msgWearingTheAmulet)
			printMessage(msgOfLordBritishLine)
			setSpellIcon(amuletIcon, 255, 9)
		}
	case item == useCrown:
		printMessage(msgCrownLine)
		if takeOff(crownIcon) {
			printMessage(msgThouDostDonTheCrown)
			printMessage(msgOfLordBritishLine)
			setSpellIcon(crownIcon, 255, 9)
		}
	case item == useSceptre:
		useSceptreItem()
	case item >= useShards && item < useSpyglass:
		useGemShard(item - useShards)
	case item == useSpyglass:
		printMessage(msgSpyglassLine)
		switch {
		case mapID > settlements || level >= underworldLevel:
			printMessage(msgNotHereLine3)
		case hour < dawn || hour > dusk:
			printMessage(msgLookingLine)
			lookAtSky()
		default:
			printMessage(msgNoStarsLine)
		}
	case item == usePlans:
		printMessage(msgPlansLine)
		if partyTile&monsterMask == actorFrigate32 {
			hmsCapePlans |= plansDouble
			printMessage(msgShipRiggedForDoubleSpeedLine)
		} else {
			printMessage(msgOnlyUsableOnShipboardLine)
		}
	case item == useSextant:
		printMessage(msgSextantLine)
		switch {
		case level >= underworldLevel || mapID != 0:
			printMessage(msgOnlyOutdoorsLine)
		case hour >= dawn && hour <= dusk:
			printMessage(msgOnlyAtNightLine)
		default:
			printMessage(msgPosition)
			printLocation()
		}
	case item == useWatch:
		printMessage(msgWatchThePocketWatchReadsSp)
		shown := int(hour) % noonHour
		if shown == 0 {
			shown = noonHour
		}
		printNumber(shown, 1, ' ')
		printChar(':')
		printNumber(int(minute), 2, '0')
		if hour >= noonHour {
			printMessage(msgPmLine)
		} else {
			printMessage(msgAmLine)
		}
	case item == useBadge:
		printMessage(msgBadgeLine)
		if takeOff(badgeIcon) {
			printMessage(msgBadgeWornLine)
			spellIcon = badgeIcon
			protection = 255
			statusChanged = true
		}
	case item == useBox:
		printMessage(msgBoxHowLine)
	}
	if result == 0 {
		printMessage(msgFailedLine)
		sweepTone(sweepFrom, sweepTo, 1, sweepTime)
	}
}

// useSceptreItem wields the sceptre: it burns away the fields around the party, or
// dissolves a field ahead.
func useSceptreItem() {
	printMessage(msgSceptreLine)
	printMessage(msgWieldingTheSceptre)
	printMessage(msgOfLordBritishLine)
	pulse(sceptreFreq, 1, sceptreTime, sceptreWidth, 1)
	burnt := 0
	if mapID <= settlements || mapID > sceptreDoom {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				at := tileAt(dx+int(mapX), dy+int(mapY))
				if *at&fieldTileMask == fieldTiles {
					*at = mapTileGrass
					updateFrame()
					whiteNoise(sceptreNoise, sceptreLength, sceptreLimit)
					burnt++
				}
			}
		}
	}
	if burnt != 0 {
		return
	}
	switch dispelField(false) {
	case 1:
		printMessage(msgFieldDissolvedLine)
	case 0:
		printMessage(msgNoEffectLine3)
	}
}

// The edges of the view a bolt flies in, in pixels.
const (
	boltLeft  = 8
	boltRight = 182
)

// plotBolt draws a point of a bolt and tells whether it flies on: false when it left the
// view or hit something that stops it.
// C: CAST_1bb0
func plotBolt(x, y, colour int) bool {
	if x < boltLeft || x > boltRight || y < boltLeft || y > boltRight {
		moveX, moveY = -1, -1
		return false
	}
	setPenColour(colour)
	plot(x, y)
	plot(x+1, y)
	speakerOn(randomRange(100, 10000))
	if y&1 == 0 {
		return true
	}
	pixelToView(x, y)
	if moveX >= 0 && moveY >= 0 {
		return isReachable(int(moveX), int(moveY))
	}
	return false
}

// maxBoltCells is the number of cells a bolt can strike.
const maxBoltCells = 63

// flyBolt fans the rays of a bolt out from the caster in a direction, noting the view
// cells they cross in xs, ys (from index 1). It returns the cells noted.
// C: CAST_1c36
func flyBolt(xs, ys []int, colour, caster, way int) int {
	var rayX, rayY [boltRays]int
	var stopped [boltRays]bool
	cells := 0
	startX := int(combatants[caster].x)*tilePixels + tilePixels/2
	startY := int(combatants[caster].y)*tilePixels + tilePixels/2
	switch way {
	case dirUp:
		startX += tilePixels / 2
	case dirRight:
		startX += tilePixels
		startY += tilePixels / 2
	case dirLeft:
		startY += tilePixels / 2
	case dirDown:
		startX += tilePixels / 2
		startY += tilePixels
	}
	colour += 8
	for i := range boltRays {
		boltCounters[i] = boltSlopes[i]
		rayX[i], rayY[i] = startX, startY
	}
	for done := 0; done < boltRays; {
		done = 0
		for i := range boltRays {
			if stopped[i] {
				done++
				continue
			}
			steps := random(15)
			for range steps {
				if !plotBolt(rayX[i], rayY[i], colour) {
					stopped[i] = true
				} else if cells < maxBoltCells && moveX > -1 && moveY > -1 && moveX < viewSize && moveY < viewSize &&
					viewTiles[int(moveY)*viewStride+int(moveX)] == 0 {
					cells++
					xs[cells], ys[cells] = int(moveX), int(moveY)
					viewTiles[int(moveY)*viewStride+int(moveX)] = hiddenTile
				}
				if !stopped[i] {
					// The counters are unsigned as in the original: only those
					// reaching exactly zero turn the ray aside.
					boltCounters[i] -= 10
					switch way {
					case dirLeft:
						rayX[i]--
					case dirRight:
						rayX[i]++
					case dirUp:
						rayY[i]--
					case dirDown:
						rayY[i]++
					}
					if boltCounters[i] < 1 {
						side := 1
						if i < boltRays/2 {
							side = -1
						}
						switch way {
						case dirLeft:
							rayY[i] -= side
						case dirRight:
							rayY[i] += side
						case dirUp:
							rayX[i] += side
						case dirDown:
							rayX[i] -= side
						}
						boltCounters[i] += boltSlopes[i]
					}
				}
				if stopped[i] {
					break
				}
			}
		}
	}
	speakerOff()
	return cells
}

// castBolt throws a bolt of sleep, poison, fire or death.
// C: CAST_1f60
func castBolt(caster, kind, colour int) {
	m := &party[combatants[caster].entity]
	way := castDirection()
	// fix: the original flies the bolt even when no direction was given.
	if way == 0 {
		return
	}
	noise := boltNoiseTime
	switch kind {
	case boltSleep:
		noise = sleepBoltTime
	case boltPoison:
		noise = poisonBolt
	}
	whiteNoise(boltNoiseRate, noise, boltNoiseTop)
	var xs, ys [maxBoltCells + 1]int
	cells := flyBolt(xs[:], ys[:], colour, caster, way)
	for c := 1; c <= cells; c++ {
		for slot := combatantSlots - 1; slot >= 0; slot-- {
			f := &combatants[slot]
			if int(f.x) != xs[c] || int(f.y) != ys[c] || f.turnTimer&turnTimerHit != 0 ||
				f.flags&fighterDead != 0 || f.flags == 0 {
				continue
			}
			f.turnTimer |= turnTimerHit
			switch kind {
			case boltSleep:
				if !resistsCharm(caster, slot, 0) && !spellProof(slot) {
					hitActor(slot)
					sleepFighter(slot)
					reportHit(slot, caster)
				}
			case boltPoison:
				if combatantStat(slot, statDexterity) <= randomDamage() {
					hitActor(slot)
					poisonOrHurt(slot, caster)
					reportHit(slot, caster)
				}
			case boltFire:
				hitActor(slot)
				increaseInt(&m.exp, damage(slot, random(30)), maxExperience)
				reportHit(slot, caster)
			case boltDeath:
				if !resistsCharm(caster, slot, 0) && !spellProof(slot) {
					hitActor(slot)
					increaseInt(&m.exp, damage(slot, 99), maxExperience)
					reportHit(slot, caster)
				}
			}
			break
		}
	}
	for slot := range combatantSlots {
		combatants[slot].turnTimer &^= turnTimerHit
	}
}
