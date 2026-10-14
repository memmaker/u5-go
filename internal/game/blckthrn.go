package game

// Death and Lord British's resurrection, and capture by Blackthorn's guards with his
// interrogation (blckthrn.c).

import "github.com/memmaker/u5-go/internal/music"
import "strings"

// Addresses of the tables in the data segment.
const (
	cellPlacesAddress    = 7954  // C: D_1f12 (by party size: the cells the members are chained in)
	cellXAddress         = 7998  // C: D_1f42
	cellYAddress         = 8008  // C: D_1f48
	dragAwayAddress      = 13982 // C: D_369e (the scripts of the scene)
	executionAddress     = 14042 // C: D_36da
	guardsComeAddress    = 14082 // C: D_3702
	releaseAddress       = 14094 // C: D_370e
	blackthornGoAddress  = 14102 // C: D_3716
	summonFreqsAddress   = 14112 // C: D_3720 (the notes as the dead are summoned)
	summonLengthsAddress = 14124 // C: D_372c
	summonWidthsAddress  = 14136 // C: D_3738
	summonStepsAddress   = 14148 // C: D_3744
)

const (
	cellPlaceSlots   = 8
	cellCount        = 6
	dragAwayLength   = 60
	executionLength  = 39
	guardsComeLength = 12
	releaseLength    = 8
	blackthornLength = 10
	summonNotes      = 6
)

var (
	cellPlaces      []byte // C: D_1f12
	cellX           []byte // C: D_1f42
	cellY           []byte // C: D_1f48
	dragAwayScript  []byte // C: D_369e
	executionScript []byte // C: D_36da
	guardsScript    []byte // C: D_3702
	releaseScript   []byte // C: D_370e
	blackthornGo    []byte // C: D_3716
	summonFreqs     []uint16
	summonLengths   []uint16
	summonWidths    []uint16
	summonSteps     []int
)

// loadBlackthornTables reads the tables of the death and capture scenes.
func loadBlackthornTables() {
	cellPlaces = segment.Bytes(cellPlacesAddress, maxPartySize*cellPlaceSlots)
	cellX = segment.Bytes(cellXAddress, cellCount)
	cellY = segment.Bytes(cellYAddress, cellCount)
	dragAwayScript = segment.Bytes(dragAwayAddress, dragAwayLength)
	executionScript = segment.Bytes(executionAddress, executionLength)
	guardsScript = segment.Bytes(guardsComeAddress, guardsComeLength)
	releaseScript = segment.Bytes(releaseAddress, releaseLength)
	blackthornGo = segment.Bytes(blackthornGoAddress, blackthornLength)
	summonFreqs = segment.Words(summonFreqsAddress, summonNotes)
	summonLengths = segment.Words(summonLengthsAddress, summonNotes)
	summonWidths = segment.Words(summonWidthsAddress, summonNotes)
	summonSteps = segment.SignedWords(summonStepsAddress, summonNotes)
}

// Commands of a scene script; bytes from scriptMoveFirst on move an actor a step.
const (
	scriptTogether   = 1 // the next move moves two actors
	scriptRepeat     = 2 // the next move happens n times
	scriptStepsOn    = 3 // footsteps sound with each move
	scriptStepsOff   = 4
	scriptSleep      = 5 // waits n frames
	scriptSetTile    = 6 // sets a tile of the map: tile, x, y
	scriptShow       = 7 // shows the frame
	scriptFootsteps  = 8 // sounds n footsteps (the repeat count)
	scriptRemove     = 9 // takes an actor off the map
	scriptMoveFirst  = 16
	scriptActorMask  = 252
	scriptDirections = 3
)

// The actor each move byte moves (by its bits above the direction).
var scriptActors = map[int]int{16: 6, 20: 7, 24: 8, 28: 0, 32: 1}

// Places, tiles and figures of the scenes.
const (
	stepFrames           = 2
	jailMapID            = shadowlordStronghold
	cellWakeX, cellWakeY = 10, 7 // where the party wakes in the jail
	refugeMapID          = 17    // Lord British's castle
	refugeLevel          = 1
	refugeXY             = 10
	refugeHour           = 6
	refugeMinKarma       = 75
	refugeFood           = 63
	interrogationBytes   = 1000
	jailMapBytes         = 176
	jailMapStride        = 16
	draggingSteps        = 5
	draggingStepsSound   = 18
	draggingStepsQuiet   = 3
	deathAnswerTries     = 4
	mantraMaxLength      = 14
	slicedPitchFrom      = 2000
	slicedPitchTo        = 25000
	slicedPitchStep      = 50
	slicedFreq           = 2640
	slicedDuration       = 200
	blackthornFreq       = 2800
	blackthornDuration   = 13000
	blackthornWidth      = 100
	blackthornStep       = 5
	interrogationKarma   = 5
	revivalBase          = 36400
	revivalOffset        = 7
	revivalDuration      = 30000
	revivalWidth         = 2000
	revivalStep          = 2
	answeredVirtue       = 255
	leftInJail           = 127 // the location of a member Blackthorn killed
	executionTile        = 128
	bloodTileA           = 234
	bloodTileB           = 235
	bloodTileC           = 232
	gateLeftX            = 2
	gateRightX           = 8
	gateY                = 7
	stairsY              = 2
)

// Offsets of the interrogation's texts in MISCMSG.DAT (as read to scratch).
const (
	askFirstAt      = 0
	askSecondAt     = 45
	askThirdAt      = 82
	askLastAt       = 172
	executionerAt   = 241
	readyAt         = 301
	slicedAt        = 416
	threatAt        = 553
	condemnAt       = 607
	answeredAloneAt = 668
	refusedAloneAt  = 732
	interrogateAt   = 812
)

// sceneFootsteps sounds count footsteps.
// C: BLCKTHRN_0000
func sceneFootsteps(count int) {
	for range count {
		footstep()
		sleepFrames(stepFrames)
	}
}

// scriptMove decodes a move byte: it returns the actor it moves and sets moveX, moveY to
// the step.
// C: BLCKTHRN_002e
func scriptMove(code int) int {
	slot := scriptActors[code&scriptActorMask]
	moveX, moveY = 0, 0
	switch code & scriptDirections {
	case faceNorth:
		moveY--
	case faceSouth:
		moveY++
	case faceEast:
		moveX++
	case faceWest:
		moveX--
	}
	return slot
}

// playScript plays a scene script.
// C: BLCKTHRN_00be
func playScript(script []byte) {
	steps := true
	count := 1
	together := false
	at := 0
	next := func() int {
		b := int(script[at])
		at++
		return b
	}
	for {
		code := next()
		if code == 0 {
			return
		}
		if code < scriptMoveFirst {
			switch code {
			case scriptTogether:
				together = true
			case scriptRepeat:
				count = next()
			case scriptStepsOn:
				steps = true
			case scriptStepsOff:
				steps = false
			case scriptSleep:
				sleepFrames(next())
			case scriptSetTile:
				tile, x, y := next(), next(), next()
				*combatCell(x, y) = byte(tile)
			case scriptShow:
				updateFrame()
			case scriptFootsteps:
				sceneFootsteps(count)
				count = 1
			case scriptRemove:
				slot := next()
				actors[slot].tile, actors[slot].animTile = 0, 0
			}
			continue
		}
		first := scriptMove(code)
		dx, dy := moveX, moveY
		second, dx2, dy2 := 0, int16(0), int16(0)
		if together {
			second = scriptMove(next())
			dx2, dy2 = moveX, moveY
		}
		for ; count > 0; count-- {
			actors[first].x += byte(dx)
			actors[first].y += byte(dy)
			if together {
				actors[second].x += byte(dx2)
				actors[second].y += byte(dy2)
			}
			if steps {
				sceneFootsteps(1)
			}
		}
		count = 1
		together = false
	}
}

// askAboutShrine has Blackthorn ask, the try'th time, for the mantra of shrine.
// C: BLCKTHRN_0278
func askAboutShrine(try, shrine int) {
	switch try {
	case 0:
		printString(scratchText(askFirstAt))
		printString(shrineNames[shrine])
		printMessage(msgChar63Char34)
	case 1:
		printString(scratchText(askSecondAt))
		printString(shrineNames[shrine])
		printMessage(msgChar63Char34)
	case 2:
		printString(scratchText(askThirdAt))
		printString(shrineNames[shrine])
		printMessage(msgChar63Char34)
	case 3:
		printString(scratchText(askLastAt))
	}
}

// answersMantra reads the player's answer and tells whether it holds the shrine's mantra.
// C: BLCKTHRN_02ea
func answersMantra(shrine int) bool {
	printMessage(msgYourResponse)
	answer := readString(mantraMaxLength)
	printMessage(msgNewlines2)
	return strings.Contains(strings.ToUpper(answer), strings.ToUpper(mantras[shrine]))
}

// executeCompanion has the executioner slay the party's second living member, sliced
// when sliced is set.
// C: BLCKTHRN_03ae
func executeCompanion(sliced bool) {
	if sliced {
		printString(scratchText(executionerAt))
	} else {
		printString(scratchText(readyAt))
	}
	sleepFrames(10)
	for width := slicedPitchFrom; width < slicedPitchTo; width += slicedPitchStep {
		pulse(slicedFreq, 1, slicedDuration, width, 0)
	}
	for width := slicedPitchTo; slicedPitchFrom < width; width -= slicedPitchStep {
		pulse(slicedFreq, 1, slicedDuration, width, 0)
	}
	explodeAt(int(actors[1].x), int(actors[1].y))
	actors[1].tile, actors[1].animTile = 0, 0
	*combatCell(5, 7) = executionTile
	living := 0
	row := 0
	for ; row < int(partySize); row++ {
		if party[row].status != statusDead {
			living++
		}
		if living == 2 {
			break
		}
	}
	victim := party[row]
	copy(party[row:partySlots-1], party[row+1:])
	party[partySlots-1] = victim
	party[partySlots-1].location = leftInJail
	partySize--
	showStatus()
	if sliced {
		printMessage(msgNewlines2)
		printString(party[partySlots-1].nameString())
		printMessage(msgIsSlicedInHalfSp)
		getChar()
		printString(scratchText(slicedAt))
	}
	printChar('\n')
}

// dragAway waits for a key and has the guards drag the party back to its cell.
// C: BLCKTHRN_0510
func dragAway() {
	getChar()
	playScript(dragAwayScript)
}

// threatenCompanion brings the executioner to the party's companion.
// C: BLCKTHRN_051c
func threatenCompanion() {
	printString(scratchText(threatAt))
	playScript(executionScript)
	printString(scratchText(condemnAt))
	printString(party[1].nameString())
	printMessage(msgDieSp)
	getChar()
	printMessage(msgNewlines2)
}

// interrogate has Blackthorn demand the mantra of shrine of a party of living members,
// killing a companion for each refusal.
// C: BLCKTHRN_054a
func interrogate(shrine, living int) {
	threatened := false
	printMessage(msgNewlines2)
	for try := range deathAnswerTries {
		askAboutShrine(try, shrine)
		if answersMantra(shrine) {
			shrineRunes[shrine] = answeredVirtue
			decreaseByte(&karma, interrogationKarma)
			if living > 1 {
				executeCompanion(false)
			} else {
				printString(scratchText(answeredAloneAt))
			}
			dragAway()
			break
		}
		if living < 2 {
			printString(scratchText(refusedAloneAt))
			dragAway()
			break
		}
		if !threatened {
			threatened = true
			threatenCompanion()
			continue
		}
		passTime(2)
		showStatus()
		switch try {
		case 0:
			*combatCell(5, 9) = bloodTileA
		case 1:
			*combatCell(5, 9) = bloodTileB
		case 2:
			*combatCell(5, 9) = bloodTileC
		case 3:
			executeCompanion(true)
		}
	}
}

// captureByBlackthorn has the party taken to Blackthorn's dungeon and questioned, then left
// in his jail.
// C: BLCKTHRN_060e_Capture
func captureByBlackthorn() {
	playSong(music.Stop)
	defer followMap()
	living := 0
	for row := range int(partySize) {
		if party[row].status != statusDead {
			living++
		}
	}
	partyTile = actorAvatar
	updateFrame()
	printMessage(msgThouArtSubduedAndBlindfolded)
	shrine := 0
	for ; shrine < shrineCount; shrine++ {
		if shrineRunes[shrine] == 0 {
			break
		}
	}
	if shrine < shrineCount {
		sleepFrames(2)
		setPenColour(0)
		fillRect(viewLeft, viewTop, viewRight, viewBottom)
		level = underworldLevel
		showSky()
		setWind(-1)
		for range draggingSteps {
			waitTicks(5)
			footstep()
		}
		printMessage(msgStrongGuardsDragTheeAway)
		steps := draggingStepsQuiet
		if soundOn {
			steps = draggingStepsSound
		}
		for range steps {
			waitTicks(5)
			footstep()
		}
		for slot := range actorSlots {
			setActor(0, 0, 0, 0, 0, 0, slot)
		}
		mapID = combatMapID
		combatant = noCombatant
		readFile("MISCMAPS.DAT", actorTiles[:jailMapBytes], 0)
		readFile("MISCMSG.DAT", scratch[:interrogationBytes], 0)
		for y := range viewSize {
			for x := range viewSize {
				*combatCell(x, y) = actorTiles[y*jailMapStride+x]
			}
		}
		for row := range living {
			a := &actors[row]
			place := cellPlaces[(living-1)*cellPlaceSlots+row]
			a.x, a.y = cellX[place], cellY[place]
			tile := classTiles[indexOf(classLetters, party[row].class)]
			a.tile, a.animTile = tile, tile
		}
		sleepFrames(16)
		printMessage(msgThouHastBeenChainedAndManacled)
		sleepFrames(50)
		printMessage(msgFootsteps)
		sceneFootsteps(8)
		setActor(actorGuard, actorGuard, 4, 10, 0, 0, 6)
		setActor(actorGuard, actorGuard, 6, 10, 0, 0, 7)
		playScript(guardsScript)
		pulse(blackthornFreq, 1, blackthornDuration, blackthornWidth, blackthornStep)
		setActor(actorCircle, actorCircle, 5, 5, 0, 0, 8)
		revealTile(actorTileBase+actorBlackthorn, 5, 5)
		setActor(actorBlackthorn, actorBlackthorn, 5, 5, 0, 0, 8)
		sleepFrames(8)
		printMessage(msgBlackthornSaysAhSp)
		printString(party[0].nameString())
		printMessage(msgTisIndeedAnHonourToMeet)
		getChar()
		printMessage(msgGuardReleaseThisGood)
		switch party[0].gender {
		case femaleGend:
			printMessage(msgLadySp)
		case maleGender:
			printMessage(msgManSp)
		}
		printMessage(msgAtOnce)
		playScript(releaseScript)
		printString(scratchText(interrogateAt))
		getChar()
		interrogate(shrine, living)
		if actors[8].tile != 0 {
			playScript(blackthornGo)
		}
	}
	level = underworldLevel
	showSky()
	setWind(-1)
	mapX, mapY = cellWakeX, cellWakeY
	keys = 0
	mapID = jailMapID
}

// partyDied plays the party's death and Lord British's resurrection of it in his castle.
// C: BLCKTHRN_0910_Death
func partyDied() {
	playSong(music.Stop)
	defer followMap()
	// the original asks for the Britannia disk here
	setLightSpell(int(lightSpell))
	partyTile = actorSleep
	if mapID < firstDungeonMap {
		updateFrame()
	}
	waitTicks(10)
	diedIn := int(mapID)
	mapID = combatMapID
	combatant = noCombatant
	printMessage(msgAnUnendingDarknessEngulfsThee)
	selectPage(1)
	setPenColour(0)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	selectPage(0)
	reveal(viewLeft, viewTop, viewRight, viewBottom)
	for y := range viewSize {
		for x := range viewSize {
			*combatCell(x, y) = hiddenTile
		}
	}
	for slot := range actorSlots {
		setActor(0, 0, 0, 0, 0, 0, slot)
	}
	updateFrame()
	printMessage(msgThouHastFoundRefuge)
	waitTicks(14)
	printMessage(msgNoEvilLivesHereOnlyPeace)
	waitTicks(28)
	printMessage(msgButThySlumberIsDisturbed)
	*combatCell(5, 5) = 0
	actors[0].tile, actors[0].animTile = actorAvatar, actorAvatar
	actors[0].x, actors[0].y = 5, 5
	updateFrame()
	for i := range summonNotes {
		pulse(int(summonFreqs[i]), 1, int(summonLengths[i]), int(summonWidths[i]), summonSteps[i])
	}
	printMessage(msgSomeoneShoutsFortisFortunaAventari)
	waitTicks(6)
	gate := &actors[1]
	gate.tile, gate.animTile = actorCircle, actorCircle
	gate.x, gate.y = gateLeftX, gateY
	*combatCell(gateLeftX, gateY) = 0
	revealTile(mapTile94, gateLeftX, gateY)
	*combatCell(gateLeftX, gateY) = mapTile94
	gate.tile = 0
	updateFrame()
	waitTicks(4)
	gate.tile = actorCircle
	gate.x = gateRightX
	*combatCell(gateRightX, gateY) = 0
	revealTile(mapTile95, gateRightX, gateY)
	*combatCell(gateRightX, gateY) = mapTile95
	gate.tile = 0
	updateFrame()
	waitTicks(4)
	printMessage(msgThereIsAPealOfThunder)
	shakeScreen()
	shakeScreen()
	*combatCell(5, stairsY) = 0
	gate.tile, gate.animTile = actorCircle, actorCircle
	gate.x, gate.y = 5, stairsY
	revealTile(actorTileBase+actor116, 5, stairsY)
	gate.tile, gate.animTile = actor116, actor116
	updateFrame()
	printMessage(msgNewlineChar34)
	readFile("KARMA.DAT", scratch[:karmaTextBytes], int(karmaTexts[int(karma)/karmaPerText]))
	printString(scratchText(0))
	printChar('"')
	getChar()
	printMessage(msgStrangeWordsAreIntoned)
	waitTicks(4)
	for row := range int(partySize) {
		pulse(revivalBase/(row+revivalOffset), 1, revivalDuration, revivalWidth, revivalStep)
		resurrect(row, answeredVirtue)
		party[row].hp = party[row].maxHP
		showStatus()
	}
	printMessage(msgVertigoLine)
	waitTicks(4)
	selectPage(1)
	setPenColour(0)
	fillRect(viewLeft, viewTop, viewRight, viewBottom)
	putTile(actorTileBase+actorAvatar, 5, 5)
	selectPage(0)
	reveal(viewLeft, viewTop, viewRight, viewBottom)
	if karma < refugeMinKarma {
		karma = refugeMinKarma
	}
	mapID = refugeMapID
	level = refugeLevel
	partyTile = actorAvatar
	mapX, mapY = refugeXY, refugeXY
	spellIcon, protection = 0, 0
	for hour != refugeHour {
		passTime(9)
	}
	torchLight, lightSpell = 0, 0
	if food == 0 {
		food = refugeFood
	}
	if diedIn != 0 && diedIn < firstDungeonMap {
		enterSettlement(true)
	}
}
