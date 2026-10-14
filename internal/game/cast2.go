package game

// Spell helpers, the shrines, the Codex and saving the game (cast2.c).

// Addresses of the cast2 tables.
const (
	castFreqAddress      = 19190 // C: D_4af6
	castWidthAddress     = 19208 // C: D_4b08
	castWidthDownAddress = 19226 // C: D_4b1a
	castStepAddress      = 19244 // C: D_4b2c
	virtueWordsAddress   = 19262 // C: D_4b3e (pointers)
	questTextAddress     = 19294 // C: D_4b5e (offsets in MISCMSG.DAT)
	codexTextAddress     = 19310 // C: D_4b6e (offsets in MISCMSG.DAT)
	strengthGiftAddress  = 19326 // C: D_4b7e
	dexterityGiftAddress = 19334 // C: D_4b86
	intellectGiftAddress = 19342 // C: D_4b8e
	urnStartAddress      = 19350 // C: D_4b96
	urnXAddress          = 19358 // C: D_4b9e
	urnYAddress          = 19394 // C: D_4bc2
	questFreqAddress     = 19430 // C: D_4be6
	questTimeAddress     = 19444 // C: D_4bf4
	questWidthAddress    = 19458 // C: D_4c02
	questStepAddress     = 19472 // C: D_4c10
	syllablesAddress     = 7164  // C: D_1bfc (pointers)
	spellLettersAddress  = 7216  // C: D_1c30 (pointers)
	castEffects          = 9
	urnPlaces            = 36
	questTunes           = 7
	syllableCount        = 26
	spellLetterLimit     = 4
	spellLineWidth       = 13
	shrineVirtueFallback = 6
	shrineInputLen       = 12
	mantraRepeats        = 3
	statLimit            = 30
	karmaLimit           = 99
	karmaGift            = 3
	humilityShrine       = 7
	codexLines           = 4
	shrineFootsteps      = 4
	shrineAltarRow       = 10
	shrineAltarColumn    = 5
	shrineWalk           = 4
	codexWalk            = 7
	shrineMinutes        = 16
	goneMember           = 127
	hundredweight        = 100
	scaledKarma          = 98
	experiencePerLevel   = 100
	allQuests            = 255
	shrineTile           = actorBeggar // the Avatar kneeling
)

var (
	castFreq      []uint16 // C: D_4af6
	castWidth     []uint16 // C: D_4b08
	castWidthDown []uint16 // C: D_4b1a
	castStep      []uint16 // C: D_4b2c
	virtueWords   []string // C: D_4b3e
	questText     []uint16 // C: D_4b5e
	codexText     []uint16 // C: D_4b6e
	strengthGift  []byte   // C: D_4b7e
	dexterityGift []byte   // C: D_4b86
	intellectGift []byte   // C: D_4b8e
	urnStart      []byte   // C: D_4b96
	urnX          []byte   // C: D_4b9e
	urnY          []byte   // C: D_4bc2
	questFreq     []uint16 // C: D_4be6
	questTime     []uint16 // C: D_4bf4
	questWidth    []uint16 // C: D_4c02
	questStep     []uint16 // C: D_4c10
	syllables     []string // C: D_1bfc
	spellLetters  []string // C: D_1c30
)

// loadCastTables reads the spell and shrine tables.
func loadCastTables() {
	castFreq = segment.Words(castFreqAddress, castEffects)
	castWidth = segment.Words(castWidthAddress, castEffects)
	castWidthDown = segment.Words(castWidthDownAddress, castEffects)
	castStep = segment.Words(castStepAddress, castEffects)
	virtueWords = messages(virtueWordsAddress, virtueCount)
	questText = segment.Words(questTextAddress, virtueCount)
	codexText = segment.Words(codexTextAddress, virtueCount)
	strengthGift = segment.Bytes(strengthGiftAddress, virtueCount)
	dexterityGift = segment.Bytes(dexterityGiftAddress, virtueCount)
	intellectGift = segment.Bytes(intellectGiftAddress, virtueCount)
	urnStart = segment.Bytes(urnStartAddress, virtueCount)
	urnX = segment.Bytes(urnXAddress, urnPlaces)
	urnY = segment.Bytes(urnYAddress, urnPlaces)
	questFreq = segment.Words(questFreqAddress, questTunes)
	questTime = segment.Words(questTimeAddress, questTunes)
	questWidth = segment.Words(questWidthAddress, questTunes)
	questStep = segment.Words(questStepAddress, questTunes)
	syllables = messages(syllablesAddress, syllableCount)
	spellLetters = messages(spellLettersAddress, spellCount)
}

// The screen flash of a spell.
const (
	flashLeft    = 8
	flashRight   = 183
	noiseRate    = 800
	noiseBase    = 8000
	noiseStep    = 1600
	noiseLimit   = 700
	effectBase   = 10000
	effectStep   = 4000
	fizzleFreq   = 4480
	fizzleTime   = 65000
	fizzleWidth  = 300
	effectFizzle = 9
	effectWind   = 1
	effectCalm   = 2
	effectDispel = 4
	effectSummon = 5
	effectRaise  = 6
	effectFail   = 8
)

// castEffect flashes the view and sounds spell effect kind (9 a fizzle).
// C: CAST2_0000_CastEffect
func castEffect(kind int) {
	if kind >= effectFizzle {
		pulse(fizzleFreq, 1, fizzleTime, fizzleWidth, 1)
		return
	}
	whiteNoise(noiseRate, kind*noiseStep+noiseBase, noiseLimit)
	setPenColour(brightWhiteColour)
	xorFillRect(flashLeft, flashLeft, flashRight, flashRight)
	pulse(int(castFreq[kind]), 1, kind*effectStep+effectBase, int(castWidth[kind]), int(castStep[kind]))
	pulse(int(castFreq[kind]), 1, kind*effectStep+effectBase, int(castWidthDown[kind]), -int(castStep[kind]))
	xorFillRect(flashLeft, flashLeft, flashRight, flashRight)
}

// onWho asks for the member a spell is for, returning its row or -1.
// C: CAST2_009e_OnWho
func onWho() int {
	printMessage(msgOnWhoSp)
	who := selectOne()
	if who < 0 {
		printMessage(msgNone)
	} else {
		printString(party[who].nameString())
	}
	if textX() != 0 {
		printChar('\n')
	}
	return who
}

// selectSpell reads the syllables of a spell. It returns the spell, -1 when none was
// given and -2 when the words are no spell.
// C: CAST2_00de_SelectSpell
func selectSpell() int {
	var letters [spellLetterLimit]byte
	var columns [spellLetterLimit]int
	count := 0
	spell := -1
	for done := false; !done; {
		c := toUpper(getChar())
		switch {
		case c == 'J' || c == 'O':
		case c >= 'A' && c <= 'Z':
			if count < spellLetterLimit {
				letters[count] = byte(c)
				if textX()+len(syllables[c-'A']) > spellLineWidth {
					printChar('\n')
				}
				columns[count] = textX()
				printString(syllables[c-'A'])
				printChar(' ')
				count++
			}
		case c == keyBackspace:
			if count != 0 {
				count--
				if textX() == 0 {
					setTextPosition(columns[count], textY()-1)
				} else {
					setTextPosition(columns[count], textY())
				}
				for range len(syllables[letters[count]-'A']) {
					printChar(' ')
				}
				setTextPosition(columns[count], textY())
			}
		case c == keyEscape:
			count = 0
			done = true
		case c == keyEnter || c == keySpace:
			done = true
		}
	}
	for sorted := false; !sorted; {
		sorted = true
		for i := 0; i < count-1; i++ {
			if letters[i+1] < letters[i] {
				letters[i], letters[i+1] = letters[i+1], letters[i]
				sorted = false
			}
		}
	}
	for i := range spellCount {
		match := true
		at := 0
		word := spellLetters[i] + "\x00"
		for j := 0; j < count; j++ {
			if word[at] != letters[j] {
				match = false
				break
			}
			for j < count-1 && letters[j] == letters[j+1] {
				j++
			}
			at++
		}
		if match && word[at] != 0 {
			match = false
		}
		if match {
			spell = i
			break
		}
	}
	if count == 0 {
		return -1
	}
	printChar('\n')
	if spell < 0 {
		spell = -2
	}
	return spell
}

// castDirection asks a direction from the caster, leaving the cell it points at in
// moveX, moveY. It returns the key (0 for a pass).
// C: CAST2_0306
func castDirection() int {
	defer asking(promptDir)()
	if mapID >= firstMapCombat {
		moveX = int16(combatants[combatant].x)
		moveY = int16(combatants[combatant].y)
	} else {
		moveX = int16(mapX)
		moveY = int16(mapY)
	}
	printMessage(msgDirection)
	for {
		key := getChar()
		switch key {
		case keyUp:
			printMessage(msgNorthLine)
			moveY--
		case keyRight:
			printMessage(msgEastLine)
			moveX++
		case keyDown:
			printMessage(msgSouthLine)
			moveY++
		case keyLeft:
			printMessage(msgWestLine)
			moveX--
		case keySpace:
			printMessage(msgPassLine2)
			return 0
		default:
			continue
		}
		return key
	}
}

// healMember heals member row by a random amount, unless dead.
// C: CAST2_03c2
func healMember(row int) bool {
	m := &party[row]
	if m.status == statusDead {
		return false
	}
	increaseInt(&m.hp, randomDamage(), int(m.maxHP))
	statusChanged = true
	return true
}

// changeWind turns the wind to way (calm for 0); calm marks the calming spell.
// C: CAST2_040a_ChangeWindDirection
func changeWind(way int, calm bool) {
	if way == 0 && !calm {
		return
	}
	if calm {
		castEffect(effectWind)
	} else {
		castEffect(effectCalm)
	}
	switch way {
	case dirUp:
		setWind(windNorth)
	case dirDown:
		setWind(windSouth)
	case dirRight:
		setWind(windEast)
	case dirLeft:
		setWind(windWest)
	case 0:
		setWind(windCalm)
	}
}

// Steps of the light spell's view refresh.
const (
	lightRefreshSteps = 20
)

// refreshLitView redraws the view a while after a light spell.
// C: CAST2_046c
func refreshLitView() {
	computeView(-1, int(mapX)-int(chunkX), int(mapY)-int(chunkY))
	for range lightRefreshSteps {
		if spellIcon != iconTimeStop {
			animateActors()
		}
		placeActors()
		drawView()
		waitTicks(1)
	}
	updateFrame()
}

// The daemon's arrival pulse.
const (
	daemonFreq  = 2760
	daemonTime  = 12000
	daemonWidth = 500
	daemonStep  = 5
	daemonTries = 8
)

// summonDaemon gates a daemon onto the combat map, charmed to the party when friendly or
// when the caster's intelligence holds it. It returns 1 when it serves, -1 when it turned,
// 0 when there was no room.
// C: CAST2_04c2
func summonDaemon(friendly bool) int {
	result := 0
	if friendly {
		castEffect(effectSummon)
	} else {
		castEffect(effectFail)
	}
	for range daemonTries {
		if !randomCell() || !fighterMayStand(actor216, int(moveX), int(moveY)) || *tileAt(int(moveX), int(moveY)) == hiddenTile {
			continue
		}
		slot := placeFighter(daemonFoe, placeFoe, int(moveX), int(moveY), int(level))
		if slot >= 0 {
			pulse(daemonFreq, 1, daemonTime, daemonWidth, daemonStep)
			a := &actors[combatants[slot].actorSlot]
			a.tile, a.animTile = actorCircle, actorCircle
			revealTile(actorTileBase+actor216, int(moveX), int(moveY))
			a.tile, a.animTile = actor216, actor216
			if !friendly && randomDamage() >= combatantStat(int(combatant), statIntelligence) {
				printMessage(msgOopsLine)
				result = -1
			} else {
				combatants[slot].flags |= fighterCharmed
				result = 1
			}
		}
		break
	}
	return result
}

// Who brings a member back.
const (
	raiseQuiet  = -1
	raiseFailed = 0
	raiseSpell  = 1
)

// resurrect brings member row back to life, costing experience by karma. It returns 1
// when raised, 0 when not dead, -1 for no one.
// C: CAST2_05e0_Resurrect
func resurrect(row, by int) int {
	if row < 0 {
		return -1
	}
	m := &party[row]
	if m.status != statusDead {
		if by != 0 {
			printMessage(msgNotDeadLine)
		}
		return 0
	}
	m.status = statusGood
	m.hp = 1
	switch m.class {
	case classAvatar, classMage:
		m.magic = m.intel
	case classBard:
		m.magic = m.intel >> 1
	}
	if karma < scaledKarma {
		m.exp = int16(int32(m.exp) * int32(karma) / hundredweight)
	}
	level := 1
	for e := int(m.exp) / experiencePerLevel; e > 0; e >>= 1 {
		level++
	}
	m.level = byte(level)
	m.maxHP = int16(level * hpPerLevel)
	if by == raiseSpell {
		castEffect(effectRaise)
	} else if by == raiseFailed {
		castEffect(effectFail)
	}
	statusChanged = true
	return 1
}

// printLocation prints the party's place in runes (latitude, longitude).
// C: CAST2_06ec
func printLocation() {
	selectFont(runicFont)
	printChar('\n')
	printChar(int(mapY)>>4 + 'A')
	printChar('\'')
	printChar(int(mapY)&15 + 'A')
	printMessage(msgChar34Char44Space)
	printChar(int(mapX)>>4 + 'A')
	printChar('\'')
	printChar(int(mapX)&15 + 'A')
	printChar('"')
	selectFont(0)
	printChar('\n')
}

// runicFont is the font with the runes.
const runicFont = 1

// unlockMagically opens a magically locked door in a direction. It returns 1 when it did,
// 0 when there was none, -1 when no direction was given.
// C: CAST2_0768
func unlockMagically() int {
	if castDirection() == 0 {
		return -1
	}
	tile := tileAt(int(moveX), int(moveY))
	switch *tile {
	case mapTile151:
		*tile = mapTileDoor184
	case mapTile152:
		*tile = mapTileDoor186
	default:
		return 0
	}
	viewChanged |= 2
	return 1
}

// Field cells of a dungeon.
const (
	dungeonFieldMask = 240
	dungeonField     = 128
	dungeonKeepBit   = 8
	dungeonAxisMask  = 7
)

// dispelField destroys a field: in a dungeon the one the party stands in or faces, else the
// summoned one in a direction. It returns 1 (or -1 in a dungeon) when one went, 0 when
// none, -1 when no direction was given.
// C: CAST2_07bc
func dispelField(withEffect bool) int {
	if mapID < firstMapCombat {
		if withEffect {
			castEffect(effectDispel)
		}
		dx := demoDeltaX[facing]
		dy := demoDeltaY[facing]
		cell := dungeonCellAt(int(mapX), int(mapY), int(level))
		if *cell&dungeonFieldMask != dungeonField {
			cell = dungeonCellAt((int(mapX)+dx)&dungeonAxisMask, (int(mapY)+dy)&dungeonAxisMask, int(level))
		}
		if *cell&dungeonFieldMask == dungeonField {
			*cell &= dungeonKeepBit
			printMessage(msgFieldDestroyedLine)
			return -1
		}
		return 0
	}
	if castDirection() == 0 {
		return -1
	}
	if withEffect {
		castEffect(effectDispel)
	}
	for i := range combatantSlots {
		a := &actors[i]
		if a.tile&^3 == actor232 && int(moveX) == int(a.x) && int(moveY) == int(a.y) {
			setActor(0, 0, 0, 0, 0, 0, i)
			return 1
		}
	}
	return 0
}

// dungeonCellAt returns the dungeon map cell at (x, y) of level.
func dungeonCellAt(x, y, level int) *byte {
	return &dungeonMap[level*dungeonSize*dungeonSize+y*dungeonSize+x]
}

// setLightSpell sets the turns of magic light.
// C: CAST2_08ea
func setLightSpell(turns int) { lightSpell = byte(turns) }

// setSpellIcon shows icon for a spell lasting turns, with effect.
// C: CAST2_08f8
func setSpellIcon(icon, turns, effect int) {
	spellIcon = byte(icon)
	protection = byte(turns)
	castEffect(effect)
	showStatus()
}

// placeUrns puts an urn on the Codex's floor for each member gone to it.
// C: CAST2_0914
func placeUrns(count int) {
	place := int(urnStart[count-1])
	for i := 0; i != count; i++ {
		setActor(actorPotion, actorPotion, urnX[place], urnY[place], level, 0, i+1)
		place++
	}
}

// Offsets of the shrine texts in MISCMSG.DAT (as read to scratch).
const (
	kneelText          = 877
	meditateText       = 920
	unfocusedText      = 960
	questOrdainedFirst = 993
	questOrdainedText  = 1038
	questReturnText    = 1099
	offerGoldText      = 1140
	notThatMuchText    = 1179
	wellDoneText       = 1209
	codexOpensText     = 1253
	codexWisdomText    = 1301
	codexEmptyText     = 1342
	codexEndText       = 1365
	codexRunes1Text    = 1407
	codexRunes2Text    = 1489
	codexRunes3Text    = 1548
	codexRunes4Text    = 1664
	approachShrineText = 1710
	approachCodexText  = 1755
	miscMessageBytes   = 2000
	miscMessageAt      = 939
	shrineMapBytes     = 176
	shrineMapAt        = 176
	codexMapAt         = 352
)

// scratchText returns the text at offset of the scratch buffer.
func scratchText(offset int) string { return cString(scratch[offset:]) }

// The sounds of the shrine.
const (
	alakazamFreq = 2700
	wellDoneFreq = 3100
	sweepLow     = 2000
	sweepHigh    = 25000
	sweepStep    = 50
	alakazamTime = 200
	wellDoneTime = 150
	kneelFrames  = 10
	answerFrames = 6
	mantraFrames = 12
)

// shrineOfVirtue asks the virtue and mantras at the altar and gives the quest, the
// blessing for gold, or the reward of a quest done.
// C: CAST2_0966_ShrineOfVirtue
func shrineOfVirtue() {
	focused := true
	virtue := 0
	for ; virtue < virtueCount; virtue++ {
		if mapX == shrineX[virtue] && mapY == shrineY[virtue] {
			break
		}
	}
	if virtue == virtueCount {
		virtue = shrineVirtueFallback
	}
	actors[0].tile, actors[0].animTile = shrineTile, shrineTile
	updateFrame()
	printString(scratchText(kneelText))
	sleepFrames(kneelFrames)
	printString(scratchText(meditateText))
	var answer [shrineInputLen]byte
	copy(answer[:], readString(shrineInputLen))
	if answer[0] == 0 {
		return
	}
	if findWord([]byte(virtueWords[virtue]+"\x00"), answer[:]) == -1 {
		focused = false
	}
	sleepFrames(answerFrames)
	printChar('\n')
	for range mantraRepeats {
		printMessage(msgMantra)
		answer = [shrineInputLen]byte{}
		copy(answer[:], readString(shrineInputLen))
		if answer[0] == 0 {
			return
		}
		if findWord([]byte(mantras[virtue]+"\x00"), answer[:]) == -1 {
			focused = false
		}
		sleepFrames(mantraFrames)
	}
	bit := uint16(1) << virtue
	switch {
	case !focused:
		printString(scratchText(unfocusedText))
	case questsDone&bit == 0:
		questsActive |= bit
		printString(scratchText(questOrdainedFirst))
		actors[0].tile, actors[0].animTile = actorAvatar, actorAvatar
		getChar()
		printString(scratchText(questOrdainedText))
		printString(scratchText(int(questText[virtue])))
		printMessage(msgChar34Newline)
		getChar()
		printString(scratchText(questReturnText))
		for i := range questTunes {
			pulse(int(questFreq[i]), 1, int(questTime[i]), int(questWidth[i]), int(questStep[i]))
		}
	case questsActive&bit == 0:
		offering := 0
		for paid := false; !paid; {
			printString(scratchText(offerGoldText))
			key := getChar()
			for key < '0' || key > '9' {
				key = getChar()
			}
			printChar(key)
			offering = key - '0'
			if offering == 0 {
				printMessage(msgGpLine)
				return
			}
			printMessage(msg00GpLine)
			if int(gold) < offering*hundredweight {
				printString(scratchText(notThatMuchText))
			} else {
				paid = true
			}
		}
		gold -= int16(offering * hundredweight)
		showStatus()
		addKarma(offering)
		selectFont(runicFont)
		printMessage(msgAlakazam)
		selectFont(0)
		printMessage(msgChar33Newline)
		setPenColour(brightWhiteColour)
		xorFillRect(flashLeft, flashLeft, flashRight, flashRight)
		for w := sweepLow; w < sweepHigh; w += sweepStep {
			pulse(alakazamFreq, 1, alakazamTime, w, 0)
		}
		for w := sweepHigh; w > sweepLow; w -= sweepStep {
			pulse(alakazamFreq, 1, alakazamTime, w, 0)
		}
	default:
		// The original clears the bit in the low byte only.
		low := byte(questsActive)
		low &= (byte(bit) & low) ^ allQuests
		questsActive = questsActive&^allQuests | uint16(low)
		printString(scratchText(wellDoneText))
		setPenColour(brightWhiteColour)
		xorFillRect(flashLeft, flashLeft, flashRight, flashRight)
		for w := sweepLow; w < sweepHigh; w += sweepStep {
			pulse(wellDoneFreq, 1, wellDoneTime, w, 0)
		}
		for w := sweepHigh; w > sweepLow; w -= sweepStep {
			pulse(wellDoneFreq, 1, wellDoneTime, w, 0)
		}
		shakeScreen()
		addKarma(karmaGift)
		avatar := &party[0]
		if strengthGift[virtue] != 0 {
			avatar.str = min(avatar.str+1, statLimit)
			printMessage(msgStrength1Line)
		}
		if dexterityGift[virtue] != 0 {
			avatar.dex = min(avatar.dex+1, statLimit)
			printMessage(msgDexterity1Line)
		}
		if intellectGift[virtue] != 0 {
			avatar.intel = min(avatar.intel+1, statLimit)
			printMessage(msgIntelligence1Line)
		}
		if virtue == humilityShrine {
			addKarma(karmaGift)
		}
	}
	sleepFrames(kneelFrames)
}

// addKarma raises karma by amount, up to 99.
func addKarma(amount int) {
	karma = byte(min(int(karma)+amount, karmaLimit))
}

// codex reads the Codex: the wisdom of the first quest begun, and when all are done the
// runes of the end.
// C: CAST2_0d24_Codex
func codex() {
	getChar()
	printString(scratchText(codexOpensText))
	getChar()
	printString(scratchText(codexWisdomText))
	getChar()
	virtue := 0
	for ; virtue < virtueCount; virtue++ {
		if byte(questsActive)&(1<<virtue) != 0 {
			break
		}
	}
	if virtue == virtueCount {
		printString(scratchText(codexEmptyText))
		return
	}
	questsDone |= 1 << virtue
	printChar('"')
	printString(scratchText(int(codexText[virtue])))
	printMessage(msgChar34Newlines2)
	getChar()
	if byte(questsDone) != allQuests {
		return
	}
	for _, colour := range []int{redColour, brightWhiteColour, redColour} {
		setPenColour(colour)
		xorFillRect(flashLeft, flashLeft, flashRight, flashRight)
		shakeScreen()
	}
	printString(scratchText(codexEndText))
	getChar()
	printMessage(msgThouDostReadLine)
	for _, text := range []int{codexRunes1Text, codexRunes2Text, codexRunes3Text, codexRunes4Text} {
		selectFont(runicFont)
		printString(scratchText(text))
		selectFont(0)
		getChar()
	}
}

// shrineStep pauses for a footstep.
// C: CAST2_0e64
func shrineStep() {
	sleepFrames(1)
	footstep()
	sleepFrames(4)
}

// shrine enters the shrine or the Codex the party stands on.
// C: CAST2_0e76_Shrine
func shrine() {
	urns := 0
	place := int(*tileAt(int(mapX), int(mapY)))
	viewChanged = 1
	from := mapID
	mapID = combatMapID
	combatant = noCombatant
	savedActors = actors
	for i := range actors {
		actors[i].tile = 0
	}
	at := shrineMapAt
	if place == mapTileCodex {
		at = codexMapAt
	}
	readFile("MISCMAPS.DAT", actorTiles[:shrineMapBytes], at)
	readFile("MISCMSG.DAT", scratch[:miscMessageBytes], miscMessageAt)
	for y := range viewSize {
		for x := range viewSize {
			*combatCell(x, y) = actorTiles[y*actorMapStride+x]
		}
	}
	updateFrame()
	for i := 1; i < partySlots; i++ {
		if party[i].location == goneMember {
			urns++
		}
	}
	if urns > 0 {
		placeUrns(urns)
	}
	if place != mapTileCodex {
		printString(scratchText(approachShrineText))
	} else {
		printString(scratchText(approachCodexText))
	}
	for range shrineFootsteps {
		shrineStep()
	}
	avatar := &actors[0]
	avatar.tile, avatar.animTile = actorAvatar, actorAvatar
	avatar.x, avatar.y, avatar.z = shrineAltarColumn, shrineAltarRow, level
	shrineStep()
	if place == mapTileCodex && urns > 0 {
		printMessage(msgThouDostSeeLine2)
		if urns == 1 {
			printMessage(msgAnUrnMarkedLine)
		} else {
			printMessage(msgUrnsMarkedLine)
		}
		for i := 1; i < partySlots; i++ {
			if party[i].location == goneMember {
				for _, c := range []byte(party[i].nameString()) {
					printChar(toUpper(int(c)))
				}
				printChar('\n')
			}
		}
	}
	walk := shrineWalk
	if place == mapTileCodex {
		walk = codexWalk
	}
	for range walk {
		avatar.y--
		shrineStep()
	}
	if place != mapTileCodex {
		shrineOfVirtue()
	} else {
		codex()
	}
	avatar.tile, avatar.animTile = actorAvatar, actorAvatar
	sleepFrames(1)
	for y := int(avatar.y); y < shrineAltarRow; y++ {
		avatar.y++
		shrineStep()
	}
	avatar.tile, avatar.x, avatar.y = 0, 0, 0
	for range shrineFootsteps {
		shrineStep()
	}
	sleepFrames(kneelFrames)
	actors = savedActors
	mapID = from
	passTime(shrineMinutes)
	updateVisibility()
}

// The size of the outdoor actor lists in SAVED.OOL.
const actorListBytes = 256

// saveGame asks and saves the game to SAVED.GAM and the outdoor actor lists to SAVED.OOL.
// C: CAST2_10fe_SaveGame
func saveGame() {
	printMessage(msgSaveGameSp)
	key := getChar()
	for key != 'Y' && key != 'N' {
		key = getChar()
	}
	if key == 'N' {
		printMessage(msgNoLine)
		return
	}
	printMessage(msgYesSavingLine)
	lists := make([]byte, 2*actorListBytes)
	readFile("UNDER.OOL", lists[actorListBytes:], 0)
	readFile("BRIT.OOL", lists[:actorListBytes], 0)
	writeFile("SAVED.GAM", encodeSave())
	writeFile("SAVED.OOL", lists)
	printMessage(msgDoneLine2)
}
