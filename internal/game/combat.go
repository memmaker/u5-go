package game

// The combat overlay: the fighters' turns, the foes' minds, hits and deaths (combat.c).

// Bits of a foe type's flags (D_153c).
const (
	foeMagic     = 32768
	foeFlag4000  = 16384
	foeTeleports = 8192
	foeVanishes  = 4096
	foeDisappear = 2048
	foeSummons   = 1024
	foePoisons   = 512
	foeFlag100   = 256
	foeStrong    = 128 // hits with strength, not dexterity
	foeCharms    = 64
	foeUndead    = 32
	foeDivides   = 16
	foeNoDamage  = 8
	foePoisonous = 4
	foeSteals    = 2
	foeNoCorpse  = 1
)

// The statistics combatantStat looks up.
const (
	statIntelligence = -1
	statDexterity    = -2
	statStrength     = -3
	statDefence      = -4
	statByWeapon     = 0
)

// Foe types the combat code knows by number.
const (
	mimicFoe    = 26
	reaperFoe   = 27
	gazerFoe    = 28
	sandTrapFoe = 30
	daemonFoe   = 38
	corpserFoe  = 45
	shadowFoe   = 47
	savageFoe   = 40 // the saved map whose fighters see the invisible
)

// Items the combat code knows by number.
const (
	daggerItem      = 16
	flamingOilItem  = 19
	oilFlaskItem    = 17
	spearItem       = 21
	throwingAxeItem = 22
	morningStarItem = 25
	bowItem         = 26
	halberdItem     = 34
	chaosSwordItem  = 35
	magicBowItem    = 36
	cannonItem      = 38
	glassSwordItem  = 39
	jeweledSword    = 40
	turningAmulet   = 45
	firstMagicArm   = 42 // the weapons from 42 to 49 and 51 hit by intelligence when cast
	lastMagicArm    = 49
	magicArm51      = 51
	bareHands       = 255
	// The spell attacks: fields of fire, poison, sleep and energy.
	fireFieldSpell   = 53
	poisonFieldSpell = 51
	sleepFieldSpell  = 52
	energyFieldSpell = 54
	charmAttack      = 33
)

const (
	iconNegate  = 'N'
	iconNoMagic = 28 // a second icon under which no magic works
	iconCharm   = 'C'
)

// Combat events (D_58a2 bits).
const (
	eventKilled   = 1
	eventVanished = 2
	eventSlept    = 4
	eventPoisoned = 8
	eventEscaped  = 16
	eventGrazed   = 32
)

const (
	killingBlow         = 99 // damage that always kills
	noDistance          = 99
	turnTimerBase       = 36
	roundsPerMinute     = 10
	maxExperience       = 9999
	stealFood           = 5
	quarterRoll         = 3
	fieldFire           = 100
	fieldPoison         = 50
	fieldSleep          = 150
	fieldFireDamage     = 10
	poisonDamage        = 20
	wakeChance          = 16
	awakeChance         = 16
	teleportRange       = 15
	hitRollMidpoint     = 30
	mapTileFireField    = 188
	mapTileSwamp132     = 132 // with 133: the tiles a fighter loses its turns on
	mapTileCombatBridge = 135
	sandTrapTile        = 76
	triggerCount        = 8
	rowTriggers         = 8
	rowNewTile1         = 9
	rowNewTile2         = 10
	triggerYColumn      = specialColumn + triggerCount
	noTrigger           = 255
)

// Tones of the combat code.
const (
	strikeFrom = 750
	strikeTo   = 400
	strikeStep = 5
	strikeTime = 150
	stealFrom  = 800
	stealTo    = 2000
	stealTime  = 50
	argRate    = 40
	argTime    = 3000
	argLimit   = 500
	gulpRate   = 1
	gulpTime   = 7000
	gulpLimit  = 600
)

// Addresses of the combat tables.
const (
	foeFlagsAddress      = 5436 // C: D_153c
	foeRangeAddress      = 5532 // C: D_159c
	foeMissileAddress    = 5580 // C: D_15cc
	armsDamageAddress    = 5628 // C: D_15fc
	armsRangeAddress     = 5732 // C: D_1664
	armsMissileAddress   = 5788 // C: D_169c
	missileShapesAddress = 8584 // C: D_2188
	missileShapesBytes   = 66   // the shapes (50) and the spark offsets (16) after them
	sparkOffsetsStart    = 50   // C: D_21ba
	fighterNamesAddress  = 6230 // C: D_1856 (pointers)
)

var (
	foeFlags      []uint16 // C: D_153c
	foeRange      []byte   // C: D_159c
	foeMissile    []byte   // C: D_15cc
	armsDamage    []byte   // C: D_15fc
	armsRange     []byte   // C: D_1664
	armsMissile   []byte   // C: D_169c
	missileShapes []int8   // C: D_2188 (and D_21ba after it)
	fighterNames  []string // C: D_1856 (a foe type's name in combat)
)

// loadFightTables reads the foes' and weapons' fighting tables.
func loadFightTables() {
	foeFlags = segment.Words(foeFlagsAddress, foeTypes)
	foeRange = segment.Bytes(foeRangeAddress, foeTypes)
	foeMissile = segment.Bytes(foeMissileAddress, foeTypes)
	armsDamage = segment.Bytes(armsDamageAddress, itemTypes)
	armsRange = segment.Bytes(armsRangeAddress, itemTypes)
	armsMissile = segment.Bytes(armsMissileAddress, itemTypes)
	fighterNames = nil
	for _, address := range segment.Words(fighterNamesAddress, foeTypes) {
		fighterNames = append(fighterNames, segment.String(int(address)))
	}
	missileShapes = nil
	for _, b := range segment.Bytes(missileShapesAddress, missileShapesBytes) {
		missileShapes = append(missileShapes, int8(b))
	}
}

// fighterMayStand returns whether an actor of tile may stand at (x, y) of the combat map;
// off the map counts as free.
// C: COMBAT_0000
func fighterMayStand(tile, x, y int) bool {
	if x > combatEdge || y > combatEdge || x < 0 || y < 0 {
		return true
	}
	if !canEnter(tile, int(*tileAt(x, y))) || *tileAt(x, y) == hiddenTile {
		return false
	}
	for i := range combatantSlots {
		a := &actors[i]
		if int(a.x) == x && int(a.y) == y {
			t := int(a.tile)
			if t == actor235 {
				return false
			}
			if t&^3 == actor232 || t == actorSleep || t == actorDead {
				continue
			}
			if t != 0 && a.animTile != 0 {
				return false
			}
		} else if int(combatants[i].x) == x && int(combatants[i].y) == y {
			flags := combatants[i].flags
			if flags&(fighterDead|fighterFlag4) == 0 && flags == 0 {
				return false
			}
			if a.animTile != 0 && flags == fighterAsleep {
				return false
			}
		}
	}
	return true
}

// shootAtRange lets foe slot fire its missile at target from afar; held is whether the
// target's amulet turned the magic. It returns whether the foe used its turn.
// C: COMBAT_014e
func shootAtRange(slot, target int, held bool) bool {
	kind := int(combatants[slot].entity)
	missile := int(foeMissile[kind])
	if kind != mimicFoe && randomRange(0, 255) >= 128 ||
		foeFlags[kind]&foeMagic != 0 && (spellIcon == iconNegate || spellIcon == iconNoMagic) {
		return false
	}
	sweepTone(strikeFrom, strikeTo, strikeStep, strikeTime)
	hits := false
	if !held {
		hits = rollToHit(target, slot, -int(rangeX), 0)
	}
	struck := flyMissile(slot, int(combatants[target].x), int(combatants[target].y), hits, missile)
	if struck >= 0 {
		hitActor(struck)
		printChar('\n')
		strike(struck, slot)
		reportHit(struck, slot)
	}
	return true
}

// foeAttacks lets computer-run fighter slot attack its nearest enemy. It returns whether
// it used its turn.
// C: COMBAT_0226
func foeAttacks(slot int) bool {
	held := false
	kind := int(combatants[slot].entity)
	actorSlot := int(combatants[slot].actorSlot)
	target := nearestEnemy(slot)
	if target < 0 {
		return false
	}
	targetEntity := int(combatants[target].entity)
	if combatants[slot].flags&fighterCharmed != 0 {
		if fighterDistance(slot, target) != 1 {
			return false
		}
		attackWeapon = charmAttack
		meleeStrike(slot, target, charmAttack)
		return true
	}
	if combatants[target].flags&fighterPlayer != 0 &&
		party[targetEntity].equip[slotAmulet] == turningAmulet && foeFlags[kind]&foeMagic != 0 {
		held = randomRange(0, 255) < 128
	}
	if fighterDistance(slot, target) > int(foeRange[kind]) {
		return false
	}
	if kind == mimicFoe {
		actors[actorSlot].state = monsterShows
	}
	if fighterDistance(slot, target) != 1 {
		return shootAtRange(slot, target, held)
	}
	if kind == corpserFoe {
		actors[actorSlot].state = monsterShows
	}
	sweepTone(strikeFrom, strikeTo, strikeStep, strikeTime)
	combatTargets[target] = byte(slot)
	if !rollToHit(target, slot, -int(rangeX), 0) {
		return true
	}
	if foeFlags[kind]&foeSteals != 0 && randomRange(0, quarterRoll) != 0 && food != 0 {
		printMessage(msgASp)
		printFighterName(slot)
		printMessage(msgStoleSomeFoodLine)
		decreaseInt(&food, stealFood)
		sweepTone(stealFrom, stealTo, 1, stealTime)
		showStatus()
		return true
	}
	hitActor(target)
	printChar('\n')
	strike(target, slot)
	reportHit(target, slot)
	return true
}

// monsterShows is the state of a mimic or corpser that has shown itself.
const monsterShows = 32

// computerTurn plays the turn of the computer-run fighter whose turn it is.
// C: COMBAT_03f4
func computerTurn() {
	f := &combatants[combatant]
	attackWeapon = 0
	actorSlot := int(f.actorSlot)
	if spellIcon == iconTimeStop || spellIcon == iconQuickness && randomRange(0, 1) == 0 || f.flags&fighterFlag4 != 0 {
		return
	}
	if f.flags&fighterAsleep != 0 {
		if randomRange(0, wakeChance) == wakeChance {
			wakeFighter(int(combatant))
		}
		return
	}
	if f.entity == corpserFoe {
		actors[actorSlot].state = 0
	}
	if f.flags&fighterFlag2 != 0 {
		if f.flags&fighterPlayer == 0 && randomRange(0, quarterRoll) == quarterRoll {
			f.hp++
		}
		woundLevel(int(combatant))
	} else {
		if foeMagicTurn(int(combatant)) {
			return
		}
		if foeAttacks(int(combatant)) {
			return
		}
	}
	if foeMoves(int(combatant)) {
		if combatEvent&eventEscaped != 0 {
			printChar('\n')
			sweepTone(escapeFrom, escapeTo, 1, escapeTime)
			printFighterName(int(combatant))
			printMessage(msgEscapesLine)
			removeFromCombat(-int(combatant) - 1)
			if f.flags&fighterPlayer == 0 && f.entity == shadowFoe {
				passOut()
			}
		}
		updateFrame()
	} else if f.flags&fighterFlag2 != 0 {
		foeAttacks(int(combatant))
	}
}

// The commands memberCommand hands on to the map commands.
const (
	actGet = iota
	actJimmy
	actOpen
	actReady
	actSearch
	actUse
)

// memberCommand prints caption and runs command for a member; a foe can't. It returns
// whether the turn goes on.
// C: COMBAT_0544
func memberCommand(caption, command int) bool {
	printMessage(caption)
	if combatants[combatant].flags&fighterPlayer == 0 {
		printMessage(msgCanTLine)
		return true
	}
	switch command {
	case actGet:
		getCommand()
	case actJimmy:
		jimmyCommand()
	case actOpen:
		openCommand()
	case actReady:
		readyCommand()
	case actSearch:
		searchCommand()
	case actUse:
		useCommand()
	}
	return false
}

// listWeapon adds the name of item to the scratch line when it does damage, after a comma
// unless first. It returns whether it did.
// C: COMBAT_05b6
func listWeapon(item int, first bool, line *string) bool {
	if item == noEquipment || armsDamage[item] == 0 {
		return false
	}
	if !first {
		*line += message(msgChar44Space)
	}
	*line += itemNames[item]
	return true
}

// memberTurn asks the member whose turn it is for a command and carries it out.
// C: COMBAT_063e_ProcessCommand
func memberTurn() {
	key := noKey
	f := &combatants[combatant]
	mapX = f.x
	mapY = f.y
	entity := int(f.entity)
	if activeMember != noMember && (f.flags&fighterPlayer == 0 || int(activeMember) != entity) {
		endCombatTurn()
		return
	}
	if f.flags&fighterPlayer != 0 && (party[entity].equip[slotLeftHand] == chaosSwordItem || party[entity].equip[slotRightHand] == chaosSwordItem) {
		f.flags |= fighterCharmed
		activeMember = noMember
		statusChanged = true
		computerTurn()
	} else {
		if f.flags&fighterPlayer != 0 {
			highlightMember(int(f.entity))
		}
		if statusChanged {
			showStatus()
			statusChanged = false
		}
		weapons := 0
		for again := true; again; {
			printChar('\n')
			printFighterName(int(combatant))
			line := ""
			if f.flags&fighterPlayer != 0 {
				printMessage(msgArmedWithSp)
				weapons = 0
				for _, slot := range []int{slotHead, slotLeftHand, slotRightHand} {
					if listWeapon(int(party[entity].equip[slot]), weapons == 0, &line) {
						weapons++
					}
				}
				if weapons == 0 {
					line += message(msgBareHands)
				}
			}
			line += message(msgChar58)
			printString(line)
			again = false
			for done := false; !done; {
				printChar('\n')
				printOpenCaption()
				done, again, key = memberCommandKey(weapons)
			}
		}
		if f.flags&fighterPlayer != 0 {
			highlightMember(int(f.entity))
		}
	}
	if key < '0' || key > '6' {
		endCombatTurn()
		updateFrame()
		checkMirror()
	}
}

// noKey stands for no key read (the C's 255 start).
const noKey = 255

// memberCommandKey reads and carries out one command key of memberTurn. It returns whether
// the key was taken, whether the member asks again, and the key.
func memberCommandKey(weapons int) (done, again bool, key int) {
	defer asking(promptCommand)()
	f := &combatants[combatant]
	if f.flags&fighterFlag4 != 0 {
		printMessage(msgArghLine)
		whiteNoise(argRate, argTime, argLimit)
		regurgitate(int(combatant))
		return true, false, noKey
	}
	if f.flags&fighterAsleep != 0 {
		if randomRange(0, 255) < awakeChance {
			wakeFighter(int(combatant))
		}
		printMessage(msgZzzzzLine)
		return true, false, noKey
	}
	key = getChar()
	done = true
	switch key {
	case keyCtrlB:
		printMessage(msgBufferO)
		keyBuffering = !keyBuffering
		if keyBuffering {
			printMessage(msgFfLine)
		} else {
			printMessage(msgNLine2)
		}
		done = false
	case keyCtrlS:
		printMessage(msgSoundSp)
		if soundOn {
			printMessage(msgOffLine)
		} else {
			printMessage(msgOnLine)
		}
		soundOn = !soundOn
		done = false
	case 'A':
		attackCommand(int(combatant), weapons)
	case 'C':
		printMessage(msgCastLine)
		again = true
		if f.flags&fighterPlayer != 0 {
			if !interfered(int(combatant)) {
				rangeX, rangeY = 1, 1
				again = false
				if spellIcon == iconNegate || crown == 0 && savedMapID == shadowlordStronghold {
					printMessage(msgAbsorbedLine)
					pulse(absorbFreq, 1, absorbTime, absorbWidth, absorbStep)
				} else {
					castCommand()
				}
			}
		} else {
			printMessage(msgCanTLine)
		}
	case 'G':
		again = memberCommand(msgGet, actGet)
	case 'J':
		again = memberCommand(msgJimmy, actJimmy)
	case 'K':
		if !klimbInCombat() {
			done = false
		}
	case 'O':
		again = memberCommand(msgOpen, actOpen)
	case 'P':
		printMessage(msgPush)
		pushCommand()
	case 'R':
		again = memberCommand(msgReadyLine, actReady)
	case 'S':
		again = memberCommand(msgSearch, actSearch)
	case 'U':
		again = memberCommand(msgUseItemLine, actUse)
	case 'Y':
		printMessage(msgYellSp)
		yellCommand()
	case 'Z':
		printMessage(msgZStatsLine)
		zstatsCommand()
	case keyEscape:
		again = escapeCombat() != 0
	case keySpace:
		printMessage(msgPassLine2)
	case '0':
		activeMember = noMember
		printMessage(msgSetActivePlrNoneLine)
		showStatus()
	case '1', '2', '3', '4', '5', '6':
		if !setActiveInCombat(key - '1') {
			again = true
		}
	case keyLeft, keyRight, keyUp, keyDown:
		if !moveInCombat(int(combatant), key) {
			done = false
		}
	case 'B':
		again = refuseInCombat(msgBoard, answerWhat)
	case 'D':
		printMessage(msgDWhatLine)
		again = true
	case 'E':
		again = refuseInCombat(msgEnter, answerHere)
	case 'F':
		again = refuseInCombat(msgFire, answerHere)
	case 'H':
		again = refuseInCombat(msgHoleUp, answerHere)
	case 'I':
		again = refuseInCombat(msgIgniteTorch, answerHere)
	case 'L':
		again = refuseInCombat(msgLook, answerHere)
	case 'M':
		again = refuseInCombat(msgMix, answerHere)
	case 'N':
		again = refuseInCombat(msgNewOrder, answerHere)
	case 'Q':
		again = refuseInCombat(msgQuit, answerHere)
	case 'T':
		again = refuseInCombat(msgTalk, answerFunny)
	case 'V':
		again = refuseInCombat(msgView, answerHere)
	case 'W':
		printMessage(msgWWhatLine)
		again = true
	case 'X':
		again = refuseInCombat(msgXIt, answerWhat)
	default:
		printMessage(msgWhatLine)
		done = false
	}
	return done, again, key
}

// The "Absorbed!" pulse.
const (
	absorbFreq  = 9800
	absorbTime  = 28000
	absorbWidth = 1000
	absorbStep  = 2
)

// shadowlordStronghold is the map whose magic the crown alone lets through.
const shadowlordStronghold = 18

// combatMainLoop gives the fighters their turns until one side is gone. It returns whether
// the party lost (or left).
// C: COMBAT_0b94_MainLoop
func combatMainLoop() bool {
	lost := true
	exitDirection = 0
	updateFrame()
	showStatus()
	clearKeyboard()
	countSides()
	if moveX == 0 {
		battleWon = 1
	} else {
		battleWon = 0
	}
	for over := false; !over; {
		for combatant = 0; int(combatant) < combatantSlots; combatant++ {
			doorOpen = 0
			f := &combatants[combatant]
			flags := f.flags
			if flags&(fighterPlayer|fighterMonster) == 0 || flags&fighterDead != 0 {
				continue
			}
			if flags&fighterPlayer != 0 && party[f.entity].status == statusDead {
				f.flags |= fighterDead
				hitActor(int(combatant))
				damage(int(combatant), killingBlow)
				continue
			}
			if *combatCell(int(f.x), int(f.y))&^1 == mapTileSwamp132 {
				continue
			}
			f.turnTimer--
			if f.turnTimer != 0 {
				continue
			}
			f.turnTimer = byte(turnTimerBase - int(f.dex))
			attackWeapon, rangeY, rangeX, combatEvent, crosshairOn = 0, 0, 0, 0, 0
			combatRounds++
			if combatRounds == roundsPerMinute {
				combatRounds = 0
				passTime(1)
			}
			focusBlink = 1
			if computerControlled(int(combatant)) {
				computerTurn()
			} else {
				memberTurn()
			}
			combatTargets[combatant] = noTarget
			fieldEffects(int(combatant))
			countSides()
			if moveY == 0 {
				if moveX == 0 {
					lost = false
					over = true
					break
				}
				if passOut() == -1 {
					updateFrame()
					if battleWon == 0 {
						printMessage(msgBattleIsLost)
						lost = true
					}
					over = true
					break
				}
				continue
			}
			if moveX == 0 && battleWon == 0 {
				printMessage(msgVictoryLine)
				battleWon = 1
				victoryTune()
				clearKeyboard()
			}
		}
	}
	clearKeyboard()
	flickerStep = flickerStart
	return lost
}

// nearestEnemy finds the nearest fighter of the other side from slot and sets moveX,
// moveY to the step towards it (away when fleeing). It returns its slot or -1.
// C: COMBAT_0d30
func nearestEnemy(slot int) int {
	members := -1
	nearest := -1
	best := noDistance
	targetX, targetY := 0, 0
	f := &combatants[slot]
	side := computerControlled(slot)
	if spellIcon == iconCharm {
		if combatantStat(slot, statIntelligence) < randomDamage() {
			side = false
		}
	}
	for i := combatantSlots - 1; i > -1; i-- {
		c := &combatants[i]
		if i == slot || c.flags == 0 || c.flags&fighterDead != 0 {
			continue
		}
		if side == computerControlled(i) {
			continue
		}
		if (savedMapID == savageFoe || f.entity == shadowFoe || c.flags&fighterInvisible == 0) && c.flags&fighterFlag4 == 0 {
			if i < maxPartySize-1 {
				members++
			}
			d := combatDistance(int(f.x), int(f.y), int(c.x), int(c.y))
			if d < best {
				best = d
				nearest = i
				targetX = int(c.x)
				targetY = int(c.y)
			}
		}
	}
	if members == -1 && nearest == -1 {
		nearest = passOut()
	}
	if nearest == -1 {
		for i := combatantSlots - 1; i > maxPartySize-1; i-- {
			if combatants[i].flags&fighterMonster != 0 {
				combatants[i].hp = 1
				combatants[i].flags |= fighterFlag2
			}
		}
		targetX, targetY = viewCentre, viewCentre
	}
	moveX, moveY = 0, 0
	if int(f.x) > targetX {
		moveX = -1
	}
	if int(f.x) < targetX {
		moveX = 1
	}
	if int(f.y) > targetY {
		moveY = -1
	}
	if int(f.y) < targetY {
		moveY = 1
	}
	if f.flags&fighterFlag2 != 0 {
		moveX = -moveX
		moveY = -moveY
	}
	if nearest == slot {
		nearest = -1
	}
	return nearest
}

// The tries and markers of foeMoves.
const (
	wanderTries   = 4
	steppedAcross = 999
	steppedDown   = 998
	steppedAny    = 991
	steppedMark   = 990
)

// foeMoves moves computer-run fighter slot: it teleports, steps towards its enemy or
// wanders. It returns whether it moved.
// C: COMBAT_0ee4
func foeMoves(slot int) bool {
	moved := 0
	step := 0
	f := &combatants[slot]
	if f.entity == reaperFoe || f.entity == mimicFoe {
		return false
	}
	if f.flags&fighterPlayer == 0 && foeFlags[f.entity]&foeTeleports != 0 && spellIcon != iconNegate && spellIcon != iconNoMagic &&
		(hemmedIn(slot) || random(quarterRoll) != quarterRoll) && randomCell() {
		if fighterMayStand(int(actors[f.actorSlot].tile), int(moveX), int(moveY)) {
			f.x = byte(moveX)
			actors[f.actorSlot].x = f.x
			f.y = byte(moveY)
			actors[f.actorSlot].y = f.y
			printString(fighterNames[f.entity])
			printMessage(msgTeleportsLine)
			moved++
		}
	}
	if moved == 0 && !hemmedIn(slot) {
		nearestEnemy(slot)
		if random(255) > 127 && !cannotStand(int(f.x)+int(moveX), int(f.y), slot) {
			moveY = 0
			step = steppedAcross
		} else if !cannotStand(int(f.x), int(f.y)+int(moveY), slot) {
			moveX = 0
			step = steppedDown
		}
		if step < steppedMark {
			step = 0
			for tries := wanderTries; tries > 0; tries-- {
				step = random(3)
				switch step {
				case 0:
					moveY, moveX = 1, 0
				case 1:
					moveX, moveY = 1, 0
				case 2:
					moveY, moveX = -1, 0
				case 3:
					moveX, moveY = -1, 0
				}
				if !cannotStand(int(f.x)+int(moveX), int(f.y)+int(moveY), slot) {
					step = steppedAny
				}
				if step > steppedMark {
					break
				}
				moveX, moveY = 0, 0
			}
		}
		if step != 0 {
			f.x = byte(int(f.x) + int(moveX))
			actors[f.actorSlot].x = f.x
			f.y = byte(int(f.y) + int(moveY))
			actors[f.actorSlot].y = f.y
			moved = 1
			if !onCombatMap(int(f.x), int(f.y)) {
				combatEvent = eventEscaped
			}
		}
	}
	return moved != 0
}

// blockedInCombat springs the combat map's triggers at (x, y), changing the tiles they
// name. It returns whether one sprang.
// C: COMBAT_111a
func blockedInCombat(x, y int) bool {
	sprang := false
	if combatKind&shipCombat == 0 && combatKind&combatKindTriggers == 0 {
		return false
	}
	for i := range triggerCount {
		tx := combatCell(specialColumn+i, rowTriggers)
		ty := combatCell(triggerYColumn+i, rowTriggers)
		if int(*tx) != x || int(*ty) != y {
			continue
		}
		*tx, *ty = noTrigger, noTrigger
		newTile := *combatCell(specialColumn+i, rowNewTiles)
		x1, y1 := int(*combatCell(specialColumn+i, rowNewTile1)), int(*combatCell(triggerYColumn+i, rowNewTile1))
		if x1 < viewSize && y1 < viewSize {
			*combatCell(x1, y1) = newTile
		}
		x2, y2 := int(*combatCell(specialColumn+i, rowNewTile2)), int(*combatCell(triggerYColumn+i, rowNewTile2))
		if x2 < viewSize && y2 < viewSize {
			*combatCell(x2, y2) = newTile
		}
		sprang = true
	}
	if sprang {
		updateFrame()
	}
	return sprang
}

// combatKindTriggers marks a combat map with triggers (D_58a1 bit 2).
const combatKindTriggers = 2

// randomCell sets moveX, moveY to a random cell, returning whether it is on the map.
// C: COMBAT_120e
func randomCell() bool {
	moveX = int16(random(teleportRange))
	moveY = int16(random(teleportRange))
	return moveX <= combatEdge && moveY <= combatEdge
}

// removeFromCombat clears a fighter: which < 0 clears combatant -which-1 and its actor,
// which > 0 only actor which-1.
// C: COMBAT_1236
func removeFromCombat(which int) {
	var slot int
	if which < 0 {
		which = -which - 1
		slot = int(combatants[which].actorSlot)
		combatants[which] = fighter{}
	} else {
		slot = which - 1
	}
	a := &actors[slot]
	a.y, a.z, a.hull, a.tile, a.animTile, a.x = 0, 0, 0, 0, 0, 0
}

// hitDamage rolls the damage attacker deals target, less the target's defence; 99 kills.
// C: COMBAT_12b0
func hitDamage(attacker, target int) int {
	var power int
	if combatants[attacker].flags&fighterMonster != 0 {
		power = int(foes[combatants[attacker].entity].attack)
	} else {
		power = int(attackWeapon)
		switch power {
		case glassSwordItem:
			printMessage(msgThySwordHathShatteredLine)
			unequip(attacker, power)
			power = killingBlow
		case bareHands:
			power = 1
		case jeweledSword:
			power = 0
		default:
			power = int(armsDamage[power])
			if power > 1 && power != killingBlow {
				power = randomRange(1, power)
			}
		}
	}
	if power == killingBlow {
		return killingBlow
	}
	var defence int
	if combatants[target].flags&fighterMonster != 0 {
		defence = int(foes[combatants[target].entity].defence)
	} else {
		// The original reads the member's inn byte here, not its armour.
		defence = int(party[combatants[target].entity].innSlot)
	}
	if defence != 0 {
		power -= randomRange(1, defence)
	}
	return power
}

// fighterDexterity is slot's dexterity for a hit roll: 1 for the time-stopped, the mimic
// and the sleeping.
// C: COMBAT_139a
func fighterDexterity(slot int) int {
	f := &combatants[slot]
	if spellIcon == iconTimeStop && f.flags&fighterMonster != 0 {
		return 1
	}
	if f.entity == mimicFoe {
		return 1
	}
	if f.flags&fighterAsleep != 0 {
		return 1
	}
	return int(f.dex)
}

// combatantStat returns stat (statIntelligence…, or a weapon to pick strength or
// dexterity, or 0 for a foe's own) of fighter slot.
// C: COMBAT_13e2
func combatantStat(slot, stat int) int {
	f := &combatants[slot]
	if f.flags&fighterMonster != 0 {
		if stat == statByWeapon {
			if foeFlags[f.entity]&foeStrong != 0 {
				stat = statStrength
			} else {
				stat = statDexterity
			}
		}
	} else if stat > 0 {
		// fix: bare hands (255) read past the missile table.
		if stat == bareHands {
			stat = statDexterity
		} else if armsMissile[stat] == missileStrength {
			stat = statStrength
		} else {
			stat = statDexterity
		}
	}
	switch stat {
	case statIntelligence:
		if f.flags&fighterMonster != 0 {
			return int(foes[f.entity].intel)
		}
		return int(party[f.entity].intel)
	case statDexterity:
		return fighterDexterity(slot)
	case statStrength:
		if f.flags&fighterMonster != 0 {
			return int(foes[f.entity].str)
		}
		return int(party[f.entity].str)
	case statDefence:
		if f.flags&fighterMonster != 0 {
			return int(foes[f.entity].defence)
		}
		return memberDefence(int(f.entity))
	}
	return 0
}

// missileStrength marks a weapon that hits by strength.
const missileStrength = 8

// rollToHit returns whether attacker hits target with weapon; casting (rangeX) rolls
// intelligence, some weapons always hit.
// C: COMBAT_14d6
func rollToHit(target, attacker, _ int, weapon int) bool {
	targetStat := statIntelligence
	attackStat := statIntelligence
	hits := false
	if rangeX != 0 {
		if weapon > firstMagicArm-1 && weapon < lastMagicArm+1 || weapon == magicArm51 {
			hits = true
		} else {
			targetStat, attackStat = statIntelligence, statIntelligence
		}
	} else {
		if weapon == glassSwordItem || weapon == chaosSwordItem || weapon == jeweledSword {
			hits = true
		} else {
			targetStat = statDexterity
			attackStat = weapon
		}
	}
	if !hits {
		t := combatantStat(target, targetStat)
		a := combatantStat(attacker, attackStat)
		if (t-a+hitRollMidpoint)/2 <= randomDamage() {
			hits = true
		}
	}
	return hits
}

// damage deals amount to fighter slot (99 kills) and handles its death or division. It
// returns the experience a foe's death gives.
// C: COMBAT_1574
func damage(slot, amount int) int {
	experience := 0
	f := &combatants[slot]
	if amount < 1 {
		combatEvent = eventGrazed
		amount = 0
	}
	if f.flags&fighterPlayer != 0 {
		m := &party[f.entity]
		m.hp -= int16(amount)
		if m.hp < 1 || amount == killingBlow {
			m.hp = 0
			f.flags |= fighterDead
			m.status = statusDead
			actors[f.actorSlot].tile, actors[f.actorSlot].animTile = actorSleep, actorSleep
			if f.entity == activeMember {
				activeMember = noMember
			}
		}
		showStatus()
		return 0
	}
	flags := foeFlags[f.entity]
	if flags&foeUndead != 0 && rangeY == 0 {
		amount /= 2
	}
	if flags&foeNoDamage != 0 {
		amount = 0
	}
	if int(f.hp) < amount {
		f.hp = 0
	} else {
		f.hp -= byte(amount)
	}
	a := &actors[f.actorSlot]
	if f.hp == 0 || amount == killingBlow {
		experience = int(foes[f.entity].maxHP)>>2 + 1
		treasure := int(foes[f.entity].treasure)
		f.flags = fighterDead
		f.hp = 0
		x, y := int(f.x), int(f.y)
		under := int(*combatCell(x, y))
		switch {
		case flags&(foeVanishes|foeNoCorpse) == 0:
			switch {
			case f.entity == gazerFoe:
				a.tile, a.animTile = actorDead, actorDead
				placeFighter(insectFoe, placeFoe, x, y, int(level))
				updateFrame()
			case f.entity == sandTrapFoe:
				*tileAt(x, y) = sandTrapTile
				removeFromCombat(-slot - 1)
			case under == mapTileCombatBridge || under < mapTilePoison:
				removeFromCombat(-slot - 1)
			case treasure >= randomDamage():
				a.tile, a.animTile = actorChest, actorChest
				a.hull = byte(treasure)
				if treasure > randomDamage() {
					a.hull |= trappedChest
				}
			default:
				a.tile, a.animTile = actorDead, actorDead
			}
		case flags&foeVanishes != 0:
			printString(fighterNames[f.entity])
			printMessage(msgVanishes)
			combatEvent = eventVanished
			a.tile, a.animTile = actorCircle, actorCircle
			revealTile(int(*combatCell(int(f.x), int(f.y))), int(f.x), int(f.y))
			removeFromCombat(-(slot + 1))
			passOut()
		default:
			removeFromCombat(-(slot + 1))
		}
	} else if flags&foeDivides != 0 {
		for range divideTries {
			randomNeighbour(int(f.x), int(f.y))
			x, y := int(moveX), int(moveY)
			if fighterMayStand(int(a.tile), x, y) {
				twin := placeFighter((int(a.tile)-actorWizard)>>2, placeFoe, x, y, int(level))
				if twin > -1 {
					combatants[twin].hp = f.hp
					printString(fighterNames[f.entity])
					printMessage(msgDividesLine)
					break
				}
			}
		}
	}
	return experience
}

const (
	insectFoe    = 31 // the swarm a dead gazer leaves
	trappedChest = 128
	divideTries  = 8
)

// poisonOrHurt poisons member slot, or if it is no healthy member hurts it, crediting
// attacker (-1 none) with the experience.
// C: COMBAT_18ba
func poisonOrHurt(slot, attacker int) {
	f := &combatants[slot]
	if f.flags&fighterPlayer != 0 && party[f.entity].status == statusGood {
		m := &party[f.entity]
		m.status = statusPoisoned
		printString(m.nameString())
		printMessage(msgIsPoisonedLine)
		combatEvent = eventPoisoned
		statusChanged = true
		return
	}
	experience := damage(slot, random(poisonDamage))
	if attacker > -1 && combatants[attacker].flags&fighterPlayer != 0 {
		increaseInt(&party[combatants[attacker].entity].exp, experience, maxExperience)
	}
}

// strike carries out attacker's hit on target: poison, a gazer's sleep, a sleep or poison
// field, or damage.
// C: COMBAT_194a
func strike(target, attacker int) {
	foe := combatants[attacker].flags&fighterPlayer == 0
	if foe && foeFlags[combatants[attacker].entity]&(foePoisons|foePoisonous) != 0 && randomRange(0, quarterRoll) != 0 {
		poisonOrHurt(target, attacker)
		return
	}
	if foe && combatants[attacker].entity == gazerFoe && combatants[target].flags&fighterAsleep == 0 {
		sleepFighter(target)
		return
	}
	if rangeY != 0 && attackWeapon == sleepFieldSpell {
		sleepFighter(target)
		return
	}
	if rangeY != 0 && attackWeapon == poisonFieldSpell {
		poisonOrHurt(target, attacker)
		return
	}
	amount := hitDamage(attacker, target)
	if amount < 0 && combatants[target].flags&fighterPlayer != 0 {
		combatEvent = eventGrazed
		return
	}
	experience := damage(target, amount)
	if combatants[attacker].flags&fighterPlayer != 0 {
		increaseInt(&party[combatants[attacker].entity].exp, experience, maxExperience)
	}
}

// Wound levels of a foe.
const (
	woundCritical = 1
	woundHeavy    = 2
	woundLight    = 3
	woundBarely   = 4
	fleeChance    = 251
)

// woundLevel returns how hurt foe slot is (woundCritical…woundBarely) and makes it flee
// when critical (or by chance when heavily hurt); 0 for a member.
// C: COMBAT_1a5c
func woundLevel(slot int) int {
	f := &combatants[slot]
	if f.flags&fighterPlayer != 0 {
		// The original returns no value here.
		return 0
	}
	flee := false
	level := 0
	quarter := int(foes[f.entity].maxHP) / 4
	switch {
	case int(f.hp) < quarter:
		level = woundCritical
		flee = true
	case int(f.hp) < quarter*2:
		level = woundHeavy
		if random(256) > fleeChance {
			flee = true
		}
	case int(f.hp) < quarter*2/2*3:
		level = woundLight
	default:
		level = woundBarely
	}
	if flee {
		f.flags |= fighterFlag2
	} else {
		f.flags &^= fighterFlag2
	}
	return level
}

// fieldEffects hurts, poisons or sleeps fighter slot standing in lava or a field.
// C: COMBAT_1b1e
func fieldEffects(slot int) {
	effect := 0
	f := &combatants[slot]
	under := int(*combatCell(int(f.x), int(f.y)))
	if under == mapTileLava || under == mapTileFireField {
		effect = fieldFire
	}
	if under == mapTilePoison {
		effect = fieldPoison
	}
	if effect == 0 {
		for i := range combatantSlots {
			a := &actors[i]
			if i == int(f.actorSlot) || f.x != a.x || f.y != a.y {
				continue
			}
			switch a.tile {
			case actor234:
				effect = fieldFire
			case actor232:
				effect = fieldPoison
			case actor233:
				effect = fieldSleep
			}
			if effect != 0 {
				break
			}
		}
	}
	switch effect {
	case fieldSleep:
		sleepFighter(slot)
	case fieldFire:
		hitActor(slot)
		damage(slot, random(fieldFireDamage))
		reportHit(slot, noTarget)
		statusChanged = true
	case fieldPoison:
		if actors[f.actorSlot].tile < actor128 {
			poisonOrHurt(slot, -1)
			hitActor(slot)
		}
	}
}

// regurgitate frees fighter slot from the creature that swallowed it, by its dexterity.
// C: COMBAT_1c66
func regurgitate(slot int) {
	f := &combatants[slot]
	if randomDamage() >= int(f.dex) {
		return
	}
	if f.flags&fighterPlayer != 0 {
		printString(party[f.entity].nameString())
	} else {
		printString(fighterNames[f.entity])
	}
	printMessage(msgRegurgitatedLine)
	whiteNoise(gulpRate, gulpTime, gulpLimit)
	f.flags &^= fighterFlag4
	actors[f.actorSlot].animTile = actors[f.actorSlot].tile
}
