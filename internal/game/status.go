package game

// The status panel at the right of the screen (segment 2000): the party's names, hit points
// and states, food, gold, the date, the active spell's letter and the winds; damage to the
// party; and what each hour costs.

// Colours the program draws its frame in (VGA/EGA values; set by the driver selection).
var (
	redColour         = 4  // C: D_13ae
	brightWhiteColour = 15 // C: D_13b0
	blueColour        = 1  // C: D_13b2
	greenColour       = 2  // C: D_13b4
	magentaColour     = 5  // C: D_13b6
	brightYellow      = 14 // C: D_13b8
	lightGrayColour   = 7  // C: D_13ba
)

// The text windows the program uses.
const (
	windowFrame  = 0 // the whole screen, for the frame's captions
	windowStatus = 1 // the status panel
	windowScroll = 2 // the message scroll
	windowView   = 3
)

const (
	statusNameColumn = 9  // where a member's hit points start
	statusWidth      = 15 // the status panel's columns
	statusFoodRow    = 7
	goldColumn       = 8
	iconColumn       = 6
	iconRow          = 6
	charArrow        = 26 // the pointer before the active member
	firstMapCombat   = 128
	noMember         = 255
	yearDigits       = 3
)

// printStatusLine prints member row of the status panel, or clears it.
// C: ULTIMA_2726
func printStatusLine(row int) {
	setTextPosition(0, row)
	if row >= int(partySize) {
		for range statusWidth {
			printChar(' ')
		}
		return
	}
	inCombat := mapID >= firstMapCombat
	// The original reads the fighter before checking there is one.
	if inCombat && combatant != noCombatant && combatants[combatant].flags&fighterPlayer != 0 && int(combatants[combatant].entity) == row {
		printChar(charToggleInverse)
	}
	m := &party[row]
	printString(m.nameString())
	for i := len(m.nameString()); i < statusNameColumn; i++ {
		printChar(' ')
	}
	if row == int(activeMember) {
		if m.status != statusDead && m.status != statusSleeping {
			printChar(charArrow)
		} else {
			printChar(' ')
			activeMember = noMember
		}
	} else {
		printChar(' ')
	}
	printNumber(int(m.hp), 4, ' ')
	if inCombat {
		const kinds = fighterPlayer | fighterMonster | fighterDead | fighterCharmed
		own := &combatants[row]
		if own.flags&kinds == fighterPlayer|fighterCharmed && int(own.entity) == row {
			printChar(statusCharmed)
		} else {
			printChar(int(m.status))
		}
		if combatant != noCombatant && combatants[combatant].flags&fighterPlayer != 0 && int(combatants[combatant].entity) == row {
			printChar(charToggleInverse)
		}
		return
	}
	printChar(int(m.status))
}

// showGold prints the gold in the status panel.
// C: ULTIMA_2884_UpdateGoldDisplay
func showGold() {
	previous := windowIndex
	selectWindow(windowStatus)
	setTextPosition(goldColumn, statusFoodRow)
	for _, limit := range []int16{1000, 100, 10} {
		if gold < limit {
			printChar(' ')
		}
	}
	printMessage(msgG)
	printNumber(int(gold), 1, ' ')
	for textX() < statusWidth {
		printChar(' ')
	}
	selectWindow(previous)
}

// showStatus redraws the status panel.
// C: ULTIMA_2900_UpdateVitalsDisplay
func showStatus() {
	selectWindow(windowStatus)
	for row := range maxPartySize {
		printStatusLine(row)
	}
	setTextPosition(0, statusFoodRow)
	printMessage(msgF)
	printNumber(int(food), 1, ' ')
	for textX() < goldColumn {
		printChar(' ')
	}
	if mapID < firstMapCombat && partyTile&^7 == actorFrigate32 {
		printMessage(msgShip)
		printNumber(int(actors[0].hull), 1, ' ')
		if actors[0].hull < 10 {
			printChar(' ')
		}
	} else {
		showGold()
	}
	printMessage(msgNewlineSpaces3)
	if month < 10 && day < 10 {
		printChar(' ')
	}
	printNumber(int(month), 1, ' ')
	printChar('-')
	printNumber(int(day), 1, ' ')
	printChar('-')
	printNumber(int(year), yearDigits, '0')
	for textX() < statusWidth {
		printChar(' ')
	}
	setTextPosition(iconColumn, iconRow)
	if spellIcon != 0 {
		printOpenCaption()
		printChar(int(spellIcon))
		printCloseCaption()
	} else {
		clearIconBar()
	}
	selectWindow(windowScroll)
}

// highlightMember inverts member's row of the status panel.
// C: ULTIMA_2a28
func highlightMember(row int) {
	setPenColour(brightWhiteColour)
	xorFillRect(192, row*cellSize+cellSize, 311, row*cellSize+cellSize+7)
}

// hurtMember takes damage from member, flashing their row; at no hit points they die.
// C: ULTIMA_2a52
func hurtMember(row, damage int) {
	highlightMember(row)
	whiteNoise(10, 1600, 2000)
	highlightMember(row)
	m := &party[row]
	m.hp -= int16(damage)
	if m.hp <= 0 {
		m.hp = 0
		m.status = statusDead
		if row == int(activeMember) {
			activeMember = noMember
		}
	}
	showStatus()
}

// hurtParty does 1-8 damage to each living member.
// C: ULTIMA_2aa8
func hurtParty() {
	for i := range maxPartySize {
		if i < int(partySize) && party[i].status != statusDead {
			hurtMember(i, randomRange(1, 8))
		}
	}
}

// Hours the party eats at.
var mealHours = [...]byte{6, 12, 18}

// passTurn does what each turn costs: poison, food at meal times, starving, and the
// protection spell running out.
// C: ULTIMA_2ae8
func passTurn() {
	awake := 0
	for i := range int(partySize) {
		status := party[i].status
		if status == statusDead && i == int(activeMember) {
			activeMember = noMember
		}
		if status != statusDead && status != statusSleeping {
			if status == statusPoisoned {
				hurtMember(i, 1)
			}
			awake++
		}
	}
	if hour != hourShown {
		if food == 0 {
			printMessage(msgStarvingLine)
			hurtParty()
		} else if hour == mealHours[0] || hour == mealHours[1] || hour == mealHours[2] {
			decreaseInt(&food, awake)
		}
		hourShown = hour
	}
	increaseByte(&turnCount, 1, 255)
	if protection != 0 && protection != 255 {
		protection--
		if protection == 0 {
			spellIcon = 0
			showStatus()
		}
	}
	regenerate()
}

// Movement kinds of actor tiles (movementKinds).
const (
	movesWalking = iota
	movesSwimming
	movesFlying
	movesRiding
	movesEthereal
	movesSkiff
	movesSailing
	movesRotworm
	movesCorpser
	movesWhirlpool
	movesSandTrap
)

// walkableBy says whether a walker can enter target: it isn't marked blocking, and only the
// Avatar and wizards sit on chairs.
// C: ULTIMA_2bd4
func walkableBy(actorTile byte, target int) bool {
	ok := blockedTiles[target/bitsPerByte]&(128>>(target%bitsPerByte)) == 0
	if actorTile&^1 != actorAvatar && actorTile&^15 != actorWizard && byte(target)&^3 == mapTileChair144 {
		ok = false
	}
	return ok
}

// isWater says whether tile is deep water or a river/shoreline tile.
// C: ULTIMA_2c2e
func isWater(tile int) bool { return tile < 4 || byte(tile)&^15 == mapTile96 }

// canEnter says whether an actor of actorTile can move onto target.
// C: ULTIMA_2c4c_IsWalkableTile
func canEnter(actorTile, target int) bool {
	switch movementKinds[actorTile>>2] {
	case movesWalking:
		return walkableBy(byte(actorTile), target)
	case movesSwimming:
		return isWater(target)
	case movesFlying:
		return byte(target)&^15 == mapTile96 || isWater(target) || walkableBy(byte(actorTile), target)
	case movesRiding:
		return walkableBy(byte(actorTile), target) && target != mapTileLava && target != mapTilePoison
	case movesEthereal:
		return !isWater(target)
	case movesSkiff:
		direction := byte(8 >> (actorTile & 3))
		if byte(target)&^3 == mapTile52 {
			return direction&skiffBridge[target-mapTile52] != 0
		}
		if !isWater(target) {
			return false
		}
		if target < mapTile96 {
			return true
		}
		return skiffWater[target-mapTile96]&direction != 0
	case movesSailing:
		return target <= 2
	case movesRotworm:
		return target == mapTilePoison
	case movesCorpser:
		return target == mapTileGrass
	case movesWhirlpool:
		return target == mapTileWater1
	case movesSandTrap:
		return target == mapTile7
	}
	return false
}

// Results of selectMember besides a member's row.
const (
	selectCancelled = -1
	selectEveryone  = -2
)

// selectMember lets the player pick a member with the digits or the arrows; with everyone
// set, 0 picks the whole party.
// C: ULTIMA_2d7a
func selectMember(everyone bool) int {
	promptEveryone = everyone
	defer asking(promptMember)()
	choice, shown, done := int16(0), int16(-1), false
	if mapID >= firstMapCombat && combatants[combatant].flags&fighterPlayer != 0 {
		highlightMember(int(combatants[combatant].entity))
	}
	showCaption(msgSelect)
	for !done {
		if shown != choice {
			highlightMember(int(choice))
			shown = choice
		}
		var key int
		for key = getChar(); key > '7'; key = getChar() {
		}
		if key > '0' && key < '7' && key-'1' < int(partySize) {
			choice = int16(key - '1')
		} else {
			switch key {
			case keyLeft, keyUp:
				decreaseInt(&choice, 1)
			case keyRight, keyDown:
				increaseInt(&choice, 1, int(partySize)-1)
			case '0':
				if everyone {
					choice, done = selectEveryone, true
				}
			case keyEscape:
				choice, done = selectCancelled, true
			case keyEnter, keySpace:
				done = true
			}
		}
		if shown != choice {
			highlightMember(int(shown))
		}
	}
	if choice >= 0 {
		highlightMember(int(choice))
	}
	clearCaption()
	return int(choice)
}

// selectOne lets the player pick one member.
// C: ULTIMA_2e8e
func selectOne() int { return selectMember(false) }

// Wind directions.
const (
	windCalm = iota
	windNorth
	windSouth
	windEast
	windWest
	windKeep = -1
)

const (
	firstDungeonMap   = 33
	mapUnderworldCave = 25 // C: map 0x19
	underworldLevel   = 128
	windColumn        = 6
	windRow           = 23
)

var windNames = [...]int{msgCalmSp, msgNorth, msgSouth, msgEastSp, msgWestSp}

// setWind sets the wind (unless direction is windKeep) and shows it under the map.
// C: ULTIMA_2e96_SetWindDirection
func setWind(direction int) {
	if direction != windKeep {
		windDirection = byte(direction)
		stepCounter = 0
	}
	if mapID >= firstDungeonMap || mapID == mapUnderworldCave {
		return
	}
	if level < underworldLevel {
		selectWindow(windowFrame)
		setTextPosition(windColumn, windRow)
		printOpenCaption()
		if int(windDirection) < len(windNames) {
			printMessage(windNames[windDirection])
		}
		printMessage(msgWinds)
		printCloseCaption()
		selectWindow(windowScroll)
	} else {
		setPenColour(brightWhiteColour)
		line(48, 184, 152, 184)
		setPenColour(blueColour)
		fillRect(48, 185, 152, 191)
	}
}

// changeWindSometimes turns the wind now and then, favouring calm.
// C: ULTIMA_2f62_UpdateWindDirectionRandomly
func changeWindSometimes() {
	if randomRange(0, 63) != 0 {
		return
	}
	direction := randomRange(0, 4)
	for direction == windCalm && randomRange(0, 255) < 192 {
		direction = randomRange(0, 4)
	}
	setWind(direction)
}

// poisonMember poisons a living member.
// C: ULTIMA_2fa6
func poisonMember(row int) {
	if row < int(partySize) && party[row].status != statusDead {
		party[row].status = statusPoisoned
		showStatus()
	}
}

// Trap kinds.
const (
	trapAcid = iota
	trapPoison
	trapBomb
	trapGas
)

// springTrap springs a trap on member.
// C: ULTIMA_2fd0
func springTrap(row int) {
	whiteNoise(40, 3000, 500)
	var kind int
	if mapID >= firstMapCombat {
		kind = randomRange(0, 1)
	} else {
		kind = int(trapKinds[randomRange(0, 7)])
	}
	switch kind {
	case trapAcid:
		printMessage(msgAcidLine)
		hurtMember(row, randomDamage())
	case trapPoison:
		printMessage(msgPoisonLine2)
		poisonMember(row)
	case trapBomb:
		printMessage(msgBombLine)
		hurtParty()
	case trapGas:
		printMessage(msgGasLine)
		for i := range maxPartySize {
			poisonMember(i)
		}
	}
}
