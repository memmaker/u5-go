package game

// Combat helpers: names, hits reported, distances, aiming and missiles (comsubs.c).

// resistsCharm returns whether target resists the charm of attacker; some weapons never
// charm.
// C: COMSUBS_0000
func resistsCharm(attacker, target, weapon int) bool {
	if weapon == magicArm48 || weapon == magicArm49 || weapon >= magicArm51 {
		return false
	}
	a := combatantStat(attacker, statIntelligence)
	t := combatantStat(target, statIntelligence)
	return randomDamage() < (t+hitRollMidpoint-a)/2
}

const (
	magicArm48 = 48
	magicArm49 = 49
)

// printFighterName prints the name of fighter slot.
// C: COMSUBS_0094_PrintName
func printFighterName(slot int) {
	if combatants[slot].flags&fighterPlayer != 0 {
		printString(party[combatants[slot].entity].nameString())
	} else {
		printString(fighterNames[combatants[slot].entity])
	}
}

// reportMiss prints that a spell failed or slot was missed.
// C: COMSUBS_00d2
func reportMiss(slot int) {
	if rangeX != 0 {
		printMessage(msgFailedLine)
		return
	}
	printFighterName(slot)
	printMessage(msgMissedLine2)
}

// The "possessed" and daemon gate sounds.
const (
	possessFreq  = 3100
	possessTime  = 30000
	possessWidth = 1000
	possessStep  = 2
	gateFreq     = 2760
	gateTime     = 5000
	gateWidth    = 1000
	gateStep     = 15
	magicChance  = 32
)

// foeMagicTurn lets foe slot charm a member, turn invisible or gate in a daemon. It
// returns whether it did.
// C: COMSUBS_00f4
func foeMagicTurn(slot int) bool {
	kind := int(combatants[slot].entity)
	actorSlot := int(combatants[slot].actorSlot)
	if spellIcon == iconNoMagic || spellIcon == iconNegate || combatants[slot].flags&fighterPlayer != 0 {
		return false
	}
	if foeFlags[kind]&foeCharms != 0 {
		victim := randomRange(0, combatantSlots-1)
		flags := combatants[victim].flags
		if flags&fighterPlayer != 0 &&
			flags&(fighterDead|fighterInvisible|fighterAsleep|fighterFlag4|fighterCharmed) == 0 {
			if !resistsCharm(slot, victim, 0) {
				printChar('\n')
				combatants[victim].flags |= fighterCharmed
				if combatants[victim].entity == activeMember {
					activeMember = noMember
				}
				showStatus()
				printFighterName(victim)
				printMessage(msgPossessedLine)
				pulse(possessFreq, 1, possessTime, possessWidth, possessStep)
				if kind == daemonFoe {
					removeFromCombat(-slot - 1)
				}
			}
			return true
		}
	}
	if foeFlags[kind]&foeDisappear != 0 && randomRange(0, 255) < magicChance {
		printChar('\n')
		printFighterName(slot)
		a := &actors[actorSlot]
		if a.animTile == 0 {
			printMessage(msgReappears)
			combatants[slot].flags &^= fighterInvisible
			a.animTile = a.tile
		} else {
			printMessage(msgDisappears)
			combatants[slot].flags |= fighterInvisible
			a.animTile = 0
		}
		return true
	}
	if foeFlags[kind]&foeSummons != 0 && randomRange(0, 255) < magicChance {
		if randomCell() && fighterMayStand(actor216, int(moveX), int(moveY)) {
			daemon := placeFighter(daemonFoe, placeFoe, int(moveX), int(moveY), int(level))
			if daemon != -1 {
				printChar('\n')
				printFighterName(slot)
				printMessage(msgGatesInADaemonLine)
				pulse(gateFreq, 1, gateTime, gateWidth, gateStep)
				a := &actors[combatants[daemon].actorSlot]
				a.tile, a.animTile = actorCircle, actorCircle
				revealTile(actorTileBase+actor216, int(moveX), int(moveY))
				a.tile, a.animTile = actor216, actor216
				return true
			}
		}
	}
	return false
}

// reportHit prints what attacker's (noTarget none) hit did to slot.
// C: COMSUBS_0312
func reportHit(slot, attacker int) {
	flags := combatants[slot].flags
	combatEvent &^= eventKilled
	if combatEvent&eventGrazed != 0 {
		printFighterName(slot)
		printMessage(msgGrazedLine)
		sweepTone(escapeFrom, escapeTo, 1, escapeTime)
	}
	if combatEvent&(eventGrazed|eventVanished) != 0 {
		combatEvent &^= eventGrazed | eventVanished
		return
	}
	switch {
	case flags == 0 || flags&fighterDead != 0:
		printFighterName(slot)
		printMessage(msgKilledLine)
		combatEvent |= eventKilled
	case combatEvent&eventSlept != 0:
		printFighterName(slot)
		printMessage(msgSleptLine2)
	case combatEvent&eventPoisoned == 0:
		printFighterName(slot)
		if flags&fighterPlayer != 0 {
			if attacker != noTarget && combatants[attacker].entity == corpserFoe {
				printMessage(msgDraggedUnderLine)
				sweepTone(escapeFrom, escapeTo, 1, escapeTime)
				combatants[slot].flags |= fighterFlag4
				actors[combatants[slot].actorSlot].animTile = 0
				sleepFrames(dragFrames)
			} else {
				printMessage(msgHitLine)
			}
		} else {
			switch woundLevel(slot) {
			case woundBarely:
				printMessage(msgBarelyWoundedLine)
			case woundLight:
				printMessage(msgLightlyWoundedLine)
			case woundHeavy:
				printMessage(msgHeavilyWoundedLine)
			case woundCritical:
				printMessage(msgCriticalLine)
			}
		}
	}
	if flags&fighterPlayer != 0 {
		showStatus()
		updateFrame()
	}
	combatEvent &^= eventSlept | eventPoisoned
}

const dragFrames = 4

// combatDistance returns the whole square root of the squared distance from (x1, y1) to (x2, y2),
// counted by odd numbers.
// C: COMSUBS_048a (COMSUBS_0458 squares)
func combatDistance(x1, y1, x2, y2 int) int {
	dx, dy := x1-x2, y1-y2
	rest := dx*dx + dy*dy
	root := 0
	for odd := 1; rest >= odd; odd += 2 {
		rest -= odd
		root++
	}
	return root
}

// fighterDistance returns the distance between fighters a and b.
// C: COMSUBS_04d4
func fighterDistance(a, b int) int {
	return combatDistance(int(combatants[a].x), int(combatants[a].y), int(combatants[b].x), int(combatants[b].y))
}

// aim lets fighter slot move the crosshair within reach of it, starting on its last
// target. It returns the distance aimed at, 0 when cancelled.
// C: COMSUBS_0504
func aim(slot, reach int) int {
	crosshairOn = 1
	target := int(actors[combatants[slot].actorSlot].target)
	if target > combatantSlots-1 ||
		combatants[target].flags&(fighterDead|fighterInvisible) != 0 || combatants[target].flags == 0 ||
		actors[combatants[target].actorSlot].tile == 0 || fighterDistance(slot, target) > reach {
		target = slot
	}
	crosshairX = combatants[target].x
	crosshairY = combatants[target].y
	f := &combatants[slot]
	fired, cancelled := false, false
	d := 0
	for !fired && !cancelled {
		dx, dy := 0, 0
		updateFrame()
		switch getChar() {
		case keyHome:
			dx, dy = -1, -1
		case keyPageDown:
			dx, dy = 1, 1
		case keyPageUp:
			dy, dx = -1, 1
		case keyEnd:
			dy, dx = 1, -1
		case keyUp:
			dy = -1
		case keyDown:
			dy = 1
		case keyRight:
			dx = 1
		case keyLeft:
			dx = -1
		case keySpace:
			if f.x == crosshairX && f.y == crosshairY {
				cancelled = true
			}
			fired = aimAt(f)
		case keyEnter, 'A':
			fired = aimAt(f)
		case keyEscape:
			cancelled = true
		}
		x := dx + int(crosshairX)
		y := dy + int(crosshairY)
		d = combatDistance(x, y, int(f.x), int(f.y))
		if d <= reach && x >= 0 && x < viewSize && y >= 0 && y < viewSize {
			crosshairX = byte(x)
			crosshairY = byte(y)
		}
	}
	crosshairOn = 0
	if cancelled {
		printChar('\n')
		return 0
	}
	return d
}

// Noise of a spell let fly.
const (
	castNoiseRate  = 800
	castNoiseBase  = 8000
	castNoiseStep  = 1600
	castNoiseLimit = 700
)

// aimAt fires when the crosshair is off fighter f, returning whether it did.
func aimAt(f *fighter) bool {
	if f.x == crosshairX && f.y == crosshairY {
		return false
	}
	printChar('\n')
	if rangeX != 0 {
		whiteNoise(castNoiseRate, int(rangeX)*castNoiseStep+castNoiseBase, castNoiseLimit)
	}
	return true
}

// fighterAt returns the living fighter at (x, y) of the combat map, or -1.
// C: COMSUBS_0748
func fighterAt(x, y int) int {
	found := -1
	for i := range combatantSlots {
		f := &combatants[i]
		if int(f.x) != x || int(f.y) != y {
			continue
		}
		if actors[f.actorSlot].animTile != actor244 &&
			f.flags&(fighterPlayer|fighterMonster) != 0 && f.flags&fighterDead == 0 && f.flags&fighterFlag4 == 0 {
			found = i
			break
		}
		found = -1
	}
	return found
}

// randomNeighbour sets moveX, moveY to a random cell next to (x, y) on the map.
// C: COMSUBS_07d4
func randomNeighbour(x, y int) {
	for {
		moveX = int16(randomRange(1, 3) + x - 2)
		moveY = int16(randomRange(1, 3) + y - 2)
		if moveX >= 0 && moveX <= combatEdge && moveY >= 0 && moveY <= combatEdge {
			return
		}
	}
}

// flyMissile flies attacker's missile of kind towards (x, y), astray beside it when it
// misses, springing triggers and leaving a field the weapon makes. It returns the fighter
// struck, or -1; moveX, moveY hold where it landed.
// C: COMSUBS_0822
func flyMissile(attacker, x, y int, hits bool, kind int) int {
	tx, ty := x, y
	fromX := int(combatants[attacker].x)
	fromY := int(combatants[attacker].y)
	if !hits {
		for {
			randomNeighbour(tx, ty)
			if fromX != int(moveX) || fromY != int(moveY) {
				break
			}
		}
		tx, ty = int(moveX), int(moveY)
	}
	arrived := false
	if attackWeapon == morningStarItem || attackWeapon == halberdItem {
		if hits {
			tx, ty = x, y
			arrived = true
		}
	} else {
		arrived = cannonFire(fromX, fromY, tx, ty, kind) != 0
		if arrived {
			blockedInCombat(tx, ty)
		} else {
			tx, ty = int(moveX), int(moveY)
		}
	}
	if arrived {
		struck := fighterAt(tx, ty)
		switch attackWeapon {
		case flamingOilItem, fireFieldSpell:
			placeFighter(actor234, placeBareActor, tx, ty, int(level))
		case poisonFieldSpell:
			placeFighter(actor232, placeBareActor, tx, ty, int(level))
		case sleepFieldSpell:
			placeFighter(actor233, placeBareActor, tx, ty, int(level))
		case energyFieldSpell:
			placeFighter(actor235, placeBareActor, tx, ty, int(level))
		}
		if (rangeX == 0 || hits) && struck >= 0 && struck != int(combatant) {
			moveX, moveY = int16(tx), int16(ty)
			return struck
		}
	}
	moveX, moveY = int16(tx), int16(ty)
	return -1
}

// useAmmunition spends what a shot with weapon took (an arrow, a quarrel, a thrown weapon
// beyond reach 1).
// C: COMSUBS_097c
func useAmmunition(weapon, reach int) {
	if rangeX != 0 {
		return
	}
	switch weapon {
	case bowItem, magicBowItem:
		equipment[arrowsItem]--
		if equipment[arrowsItem] == 0 {
			equipment[weapon] += byte(unequipAll(weapon))
		}
	case crossbowItem:
		equipment[quarrelsItem]--
		if equipment[quarrelsItem] == 0 {
			equipment[weapon] += byte(unequipAll(weapon))
		}
	case daggerItem, spearItem, throwingAxeItem:
		if reach <= 1 {
			break
		}
		if equipment[weapon] != 0 {
			equipment[weapon]--
			break
		}
		unequip(int(combatants[combatant].entity), weapon)
	}
}

// interfered returns whether an enemy next to slot that last fought it keeps it from
// shooting or casting.
// C: COMSUBS_09fc
func interfered(slot int) bool {
	other := int(combatTargets[slot])
	if other != noTarget && combatants[other].flags != 0 && computerControlled(other) &&
		combatants[other].flags&(fighterAsleep|fighterFlag4) == 0 && spellIcon != iconTimeStop && fighterDistance(slot, other) == 1 {
		printChar('\n')
		printFighterName(other)
		printMessage(msgInterferesLine)
		return true
	}
	return false
}

// Shot tones.
const (
	shotFrom = 1300
	shotTo   = 300
	shotStep = 5
	shotTime = 100
)

// shoot lets slot aim weapon (whose missile is kind) and fire.
// C: COMSUBS_0a68
func shoot(slot, weapon, kind int) {
	if (weapon == bowItem || weapon == crossbowItem || weapon == magicBowItem || weapon == flamingOilItem || weapon == oilFlaskItem) &&
		interfered(slot) {
		return
	}
	printMessage(msgAimSp)
	reach := aim(slot, int(armsRange[weapon]))
	if reach == 0 {
		return
	}
	sweepTone(shotFrom, shotTo, shotStep, shotTime)
	if weapon == flamingOilItem {
		if equipment[weapon] != 0 {
			equipment[weapon]--
		} else {
			unequip(int(combatants[combatant].entity), weapon)
		}
	}
	a := &actors[combatants[slot].actorSlot]
	a.target = noTarget
	aimed := fighterAt(int(crosshairX), int(crosshairY))
	struck := aimed
	if struck > -1 {
		a.target = byte(struck)
		useAmmunition(weapon, reach)
		hits := rollToHit(struck, slot, -int(rangeX), weapon)
		struck = flyMissile(slot, int(crosshairX), int(crosshairY), hits, kind)
	} else {
		struck = flyMissile(slot, int(crosshairX), int(crosshairY), true, kind)
	}
	landX, landY := int(moveX), int(moveY)
	if struck > -1 {
		hitActor(struck)
		printChar('\n')
		strike(struck, slot)
		reportHit(struck, slot)
	} else if aimed > -1 {
		printChar('\n')
		reportMiss(aimed)
	}
	if weapon == cannonItem {
		cannonFire(landX, landY, int(combatants[slot].x), int(combatants[slot].y), kind)
	}
}

// meleeStrike carries out attacker's blow with weapon on target.
// C: COMSUBS_0bf8
func meleeStrike(attacker, target, weapon int) {
	sweepTone(strikeTo, strikeFrom, strikeStep, strikeTime)
	printChar('\n')
	if rollToHit(target, attacker, -int(rangeX), weapon) {
		hitActor(target)
		strike(target, attacker)
		reportHit(target, attacker)
	} else {
		reportMiss(target)
	}
}

// firstSpellArm is the first of the weapons that are spells (they count as cast).
const firstSpellArm = 35

// attackWith lets slot attack with weapon: hand to hand next to it, else by aiming.
// C: COMSUBS_0c52
func attackWith(slot, weapon int) {
	if weapon >= firstSpellArm {
		rangeY = 1
	}
	var reach, kind int
	if combatants[slot].flags&fighterPlayer == 0 {
		reach = int(foeRange[combatants[slot].entity])
		if reach == 1 {
			reach = 0
		}
		kind = int(foeMissile[combatants[slot].entity])
	} else {
		if weapon != bareHands {
			reach = int(armsRange[weapon])
			// fix: bare hands (255) read past the missile table.
			kind = int(armsMissile[weapon])
		}
	}
	if reach != 0 {
		shoot(slot, weapon, kind)
		return
	}
	printMessage(msgAimSp)
	aimed := aim(slot, 1)
	target := fighterAt(int(crosshairX), int(crosshairY))
	if aimed == 0 || target == -1 {
		printMessage(msgNothingLine3)
		return
	}
	actors[combatants[slot].actorSlot].target = byte(target)
	meleeStrike(slot, target, weapon)
}

// attackWithArm attacks with the member's arm item when it is a weapon, naming it when the
// member holds more than one.
// C: COMSUBS_0d3c
func attackWithArm(item, weapons int) {
	if item == noEquipment || armsDamage[item] == 0 {
		return
	}
	if weapons > 1 {
		rangeY, rangeX = 0, 0
		printChar('\n')
		printString(itemNames[item])
		printMessage(msgChar58Newline)
	}
	attackWeapon = byte(item)
	printMessage(msgAttack)
	attackWith(int(combatant), item)
}

// attackCommand is the combat Attack command of slot, holding weapons weapons.
// C: COMSUBS_0d96
func attackCommand(slot, weapons int) {
	if combatants[slot].flags&fighterPlayer == 0 || weapons == 0 {
		attackWeapon = bareHands
		printMessage(msgAttack)
		attackWith(slot, bareHands)
		return
	}
	m := &party[combatants[slot].entity]
	if weapons > 1 {
		setTextPosition(0, textY())
		printChar(' ')
	}
	// The original tries equipment slots 0, 2 and 3.
	attackWithArm(int(m.equip[slotHead]), weapons)
	attackWithArm(int(m.equip[slotLeftHand]), weapons)
	attackWithArm(int(m.equip[slotRightHand]), weapons)
}

// Missile path sizes: the path buffers and how far a missile may fly in them.
const (
	pathEnd       = 255
	pathCells     = 330 // the distance from the first buffer to the second (D_a872-D_a728)
	pathCleared   = 256
	pathLimit     = 328
	slopeScale    = 100
	verticalSlope = 19200
)

var (
	pathX [pathCells + 1]byte // C: D_a728
	pathY [pathCells + 1]byte // C: D_a872
)

// traceMissilePath fills pathX, pathY with the pixels from (x1, y1) to (x2, y2), ending with
// pathEnd.
// C: COMSUBS_0e26
func traceMissilePath(x1, y1, x2, y2 int) {
	stepX, stepY := 1, 1
	left := pathLimit
	slope := verticalSlope
	if x2 != x1 {
		slope = (y2 - y1) * slopeScale / (x2 - x1)
	}
	if slope < 0 {
		slope = -slope
	}
	if x2 < x1 {
		stepX = -1
	}
	if y2 < y1 {
		stepY = -1
	}
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	acc := slope
	n := 0
	pathX[n], pathY[n] = byte(x1), byte(y1)
	n++
	for dx > 0 || dy > 0 {
		for acc > 0 && dy > 0 {
			acc -= slopeScale
			y1 += stepY
			pathY[n] = byte(y1)
			dy--
			pathX[n] = byte(x1)
			n++
			left--
			if left == 0 {
				break
			}
		}
		if left == 0 {
			break
		}
		dx--
		acc += slope
		x1 += stepX
		pathY[n], pathX[n] = byte(y1), byte(x1)
		n++
		left--
	}
	pathX[n], pathY[n] = pathEnd, pathEnd
}

// Missile kinds.
const (
	missileTrail  = 0
	missileBall   = 1
	missileArrow  = 2
	missileRed    = 3
	missileBlue   = 4
	missileGreen  = 5
	missileViolet = 6
	missileBolt   = 7
	trailPixels   = 9
	trailBehind   = 4
	sparkCount    = 48
	sparkOffsets  = 16
	shapeSecond   = 16
	highColours   = 8
	arrowLines    = 4
	boltLines     = 7
	boltDots      = 3
)

// drawMissile draws the missile of kind at pixel (x, y) in turn (1–4) of its flight; at
// is its place in the path.
// C: COMSUBS_0f4a
func drawMissile(x, y, turn, kind, at int) {
	signX, signY := 1, 1
	colour := -1
	xs, ys := 0, 0 // indexes into missileShapes
	if kind == missileArrow {
		xs, ys = 0, 0
	}
	if kind > missileArrow {
		xs, ys = shapeSecond, shapeSecond
	}
	switch turn {
	case 1:
		ys++
	case 2:
		xs++
		signX = -1
	case 3:
		ys++
		signX, signY = -1, -1
	case 4:
		xs++
		signY = -1
	}
	setPenColour(brightWhiteColour)
	shapeLine := func() {
		x1 := int(missileShapes[xs])*signX + x
		xs += 2
		y1 := int(missileShapes[ys])*signY + y
		ys += 2
		x2 := int(missileShapes[xs])*signX + x
		xs += 2
		y2 := int(missileShapes[ys])*signY + y
		ys += 2
		line(x1, y1, x2, y2)
	}
	switch kind {
	case missileTrail:
		for i := range trailPixels {
			p := at - trailBehind + i
			if pathX[p] == pathEnd {
				break
			}
			line(int(pathX[p]), int(pathY[p]), int(pathX[p])+1, int(pathY[p]))
		}
	case missileBall:
		line(x-1, y-2, x+1, y-2)
		line(x-2, y-1, x+2, y-1)
		line(x-2, y, x+2, y)
		line(x-2, y+1, x+2, y+1)
		line(x-1, y+2, x+1, y+2)
	case missileArrow:
		for range arrowLines {
			shapeLine()
		}
	case missileRed:
		colour = redColour + highColours
	case missileBlue:
		colour = blueColour + highColours
	case missileGreen:
		colour = greenColour + highColours
	case missileViolet:
		colour = magentaColour + highColours
	case missileBolt:
		for range boltLines {
			shapeLine()
		}
		for range boltDots {
			plot(int(missileShapes[xs])*signX+x, int(missileShapes[ys])*signY+y)
			xs += 2
			ys += 2
		}
	}
	if colour > -1 {
		signX, signY = -1, 1
		for range sparkCount {
			sx := sparkOffsetsStart + randomRange(0, sparkOffsets-1)
			sy := sparkOffsetsStart + randomRange(0, sparkOffsets-1)
			if randomRange(0, 1) != 0 {
				signX = -signX
			}
			if randomRange(0, 1) != 0 {
				signY = -signY
			}
			setPenColour(colour)
			plot(int(missileShapes[sx])*signX+x, int(missileShapes[sy])*signY+y)
		}
	}
}

// Missile flight.
const (
	pixelsPerCell  = 16
	flightStep     = 13
	flightStepTown = 6
	flightStepBolt = 8
	flightDelay    = 40
	missileHalf    = 8
	missileFar     = 15
	byteAlign      = 248
	lastSettlement = 32
)

// cannonFire flies a missile of kind from cell (x, y) to (tx, ty) of the view, drawing it
// on the screen. It returns 0 when something on the way stopped it (moveX, moveY hold
// where), 1 when it arrived.
// C: COMSUBS_12de
func cannonFire(x, y, tx, ty, kind int) int {
	step := flightStep
	turn := 0
	copyPage(0, 1)
	for i := range pathCleared {
		pathX[i], pathY[i] = pathEnd, pathEnd
	}
	traceMissilePath(x*pixelsPerCell+pixelsPerCell, y*pixelsPerCell+pixelsPerCell,
		tx*pixelsPerCell+pixelsPerCell, ty*pixelsPerCell+pixelsPerCell)
	at := 0
	if mapID != 0 && int(mapID) < lastSettlement+1 {
		step = flightStepTown
	} else if kind == missileBolt {
		step = flightStepBolt
	}
	if kind == missileTrail {
		at += trailBehind
	}
	for pathX[at] != pathEnd {
		turn = turn&3 + 1
		pixelToView(int(pathX[at]), int(pathY[at]))
		if moveX == -1 {
			break
		}
		drawMissile(int(pathX[at]), int(pathY[at]), turn, kind, at)
		busyWait(flightDelay)
		px := int(pathX[at]) & byteAlign
		py := int(pathY[at])
		copyArea(1, 0, px-missileHalf, py-missileHalf, px+missileFar, py+missileFar)
		at += step
		if at > pathCells {
			break
		}
		if pathX[at] == pathEnd {
			break
		}
		if !isReachable(int(moveX), int(moveY)) && (x != int(moveX) || y != int(moveY)) {
			return 0
		}
	}
	return 1
}
