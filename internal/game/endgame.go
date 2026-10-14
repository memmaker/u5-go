package game

// The end of the game: the party before Lord British in the Underworld, then the pictures of
// the ending and the certificate (endgame.c).

import "github.com/memmaker/u5-go/internal/music"
import "strconv"

// Addresses of the tables in the data segment.
const (
	endingLeftAddress     = 15782 // C: D_3da6 (x, y bytes for each page)
	endingRightAddress    = 15794 // C: D_3db2 (x, y words)
	endingTextAddress     = 15818 // C: D_3dca (END.DAT offsets)
	endingNarrowAddress   = 15830 // C: D_3dd6
	endingWideAddress     = 15836 // C: D_3ddc
	endingPenXAddress     = 15842 // C: D_3de2
	endingPenYAddress     = 15848 // C: D_3de8
	endingPictureAddress  = 15854 // C: D_3dee
	endingFileAddress     = 15860 // C: D_3df4 (which END*.16)
	endingPictureXAddress = 15866 // C: D_3dfa
	endingPictureYAddress = 15872 // C: D_3e00
	endingFrameAddress    = 15878 // C: D_3e06 (whether the page has the text frame)
	numberWordsAddress    = 15884 // C: D_3e0c (One to Nineteen)
	tensWordsAddress      = 15922 // C: D_3e32 (Twenty to Ninety)
	ordinalWordsAddress   = 15938 // C: D_3e42 (First to Twelfth)
	endingPlaceXAddress   = 15962 // C: D_3e5a (where each member stands)
	endingPlaceYAddress   = 15968 // C: D_3e60
)

const (
	endingPages   = 6
	numberWords   = 19
	tensWords     = 8
	ordinalWords  = 12
	certificateW  = 39 // the widest line of the certificate
	endingTextMax = 2000
)

var (
	endingLeft     []byte // C: D_3da6
	endingRight    []uint16
	endingText     []uint16
	endingNarrow   []byte
	endingWide     []byte
	endingPenX     []byte
	endingPenY     []byte
	endingPicture  []byte
	endingFile     []byte
	endingPictureX []byte
	endingPictureY []byte
	endingFrame    []byte
	numberNames    []string // C: D_3e0c
	tensNames      []string // C: D_3e32
	ordinalNames   []string // C: D_3e42
	endingPlaceX   []byte
	endingPlaceY   []byte
	certificate    []byte // C: D_bcb2, D_bcda (the line being set)
)

// loadEndgameTables reads the tables of the ending.
func loadEndgameTables() {
	endingLeft = segment.Bytes(endingLeftAddress, endingPages*2)
	endingRight = segment.Words(endingRightAddress, endingPages*2)
	endingText = segment.Words(endingTextAddress, endingPages)
	endingNarrow = segment.Bytes(endingNarrowAddress, endingPages)
	endingWide = segment.Bytes(endingWideAddress, endingPages)
	endingPenX = segment.Bytes(endingPenXAddress, endingPages)
	endingPenY = segment.Bytes(endingPenYAddress, endingPages)
	endingPicture = segment.Bytes(endingPictureAddress, endingPages)
	endingFile = segment.Bytes(endingFileAddress, endingPages)
	endingPictureX = segment.Bytes(endingPictureXAddress, endingPages)
	endingPictureY = segment.Bytes(endingPictureYAddress, endingPages)
	endingFrame = segment.Bytes(endingFrameAddress, endingPages)
	numberNames = messages(numberWordsAddress, numberWords)
	tensNames = messages(tensWordsAddress, tensWords)
	ordinalNames = messages(ordinalWordsAddress, ordinalWords)
	endingPlaceX = segment.Bytes(endingPlaceXAddress, endingPages)
	endingPlaceY = segment.Bytes(endingPlaceYAddress, endingPages)
}

// Pictures, tiles and places of the ending.
const (
	endPicturesFirst   = 24 // resourceFiles index of END1.16
	endScreenFile      = 28 // ENDSC.16
	noEndFile          = -1 // forces the first page to load its file
	framedFirstPage    = 0
	framedSecondPage   = 3
	frameTopX          = 216
	frameTitleIndex    = 4
	frameTitleX        = 152
	frameTitleY        = 28
	frameSideIndex     = 5
	frameSideX         = 224
	frameBottomX       = 176
	endScreenX         = 40
	endingMapOffset    = 528 // the throne room in MISCMAPS.DAT
	endingMapBytes     = 176
	endingMapStride    = 16
	endingMessageBytes = 1000
	britishSlot        = 31
	britishTile        = 124 // C: TILE_ACTOR_7C
	britishStartY      = 8
	throneY            = 3
	boxX, boxY         = 5, 4
	memberStartY       = 9
	floorTile          = mapTile68 // C: TILE_MAP_44
	codexTile          = mapTileMoongate
	boxSlot            = 6
	revivalFreq        = 8800
	revivalLength      = 40000
	revivalWidthStart  = 5000
	codexFreq          = 5200
	codexLength        = 50000
	codexWidth         = 10000
	viewInnerEdge      = 183
	viewStart          = 8
	answerYes          = 'Y'
	answerNo           = 'N'
	startYear          = 139 // the year, month and day the quest began
	startMonth         = 4
	startDay           = 5
	hundred            = 100
	ten                = 10
	twenty             = 20
	firstTwenty        = 21
	firstTens          = 2
)

// Offsets of the texts in ENDMSG.DAT.
const (
	askAt       = 33
	askAgainAt  = 73
	boxAt       = 171
	speech1At   = 211
	speech2At   = 296
	speech3At   = 359
	speech4At   = 457
	speech5At   = 529
	speech6At   = 587
	badEndingAt = 725
)

// showEnding shows the pages of the ending, each picture with its text, and the last
// screen.
// C: ENDGAME_0000
func showEnding() {
	leaveWorld()
	// the original asks for the second disk here
	font := loadPictures(proportionalFont)
	frame := loadPictures(textPictures)
	pictures := loadPictures(resourceFiles[endPicturesFirst])
	reveal(0, 0, screenRight, screenBottom)
	selectWindow(0)
	file := noEndFile
	for page := range endingPages {
		playSong(songForEnding(page))
		if int(endingFile[page]) != file {
			file = int(endingFile[page])
			pictures = loadPictures(resourceFiles[endPicturesFirst+file])
		}
		selectPage(1)
		printChar(charClearWindow)
		if endingFrame[page] == 1 {
			switch page {
			case framedFirstPage:
				putPicture(frame, 0, frameTopX, 0, 0)
				putPicture(frame, frameTitleIndex, frameTitleX, frameTitleY, 0)
			case framedSecondPage:
				putPicture(frame, frameSideIndex, frameSideX, 0, 0)
				putPicture(frame, 0, frameBottomX, 0, 0)
			}
		}
		putPicture(pictures, int(endingPicture[page]), int(endingPictureX[page]), int(endingPictureY[page]), 0)
		storyLeft = [2]int{int(endingLeft[page*2]), int(endingLeft[page*2+1])}
		storyRight = [2]int{int(endingRight[page*2]), int(endingRight[page*2+1])}
		storyNarrowY = int(endingNarrow[page])
		storyWideY = int(endingWide[page])
		storyX = int(endingPenX[page])
		storyY = int(endingPenY[page])
		readFile("END.DAT", scratch[:endingTextMax], int(endingText[page]))
		printStory(font, scratch[:endingTextMax])
		if page != 0 {
			for pollKey() == 0 {
			}
		}
		copyPage(1, 0)
	}
	for pollKey() == 0 {
	}
	selectPage(1)
	printChar(charClearWindow)
	last := loadPictures(resourceFiles[endScreenFile])
	putPicture(last, 0, endScreenX, 0, 0)
	copyPage(1, 0)
	selectPage(0)
}

// addToCertificate adds text to the line being set, printing the line when it ends.
// C: ENDGAME_023a
func addToCertificate(text string) {
	for i := 0; i < len(text) && len(certificate) < certificateW; i++ {
		certificate = append(certificate, text[i])
	}
	if len(certificate) > 0 && certificate[len(certificate)-1] == '\n' {
		printString(string(certificate))
		certificate = certificate[:0]
	}
}

// addNumber adds a number below a hundred in words.
// C: ENDGAME_028c
func addNumber(n int) {
	if n < firstTwenty {
		addToCertificate(numberNames[n-1])
		return
	}
	addToCertificate(tensNames[n/ten-firstTens])
	if n%ten != 0 {
		addToCertificate(message(msgChar45))
		addToCertificate(numberNames[n%ten-1])
	}
}

// addOrdinal adds an ordinal number to twenty-nine in words.
// C: ENDGAME_02d6
func addOrdinal(n int) {
	switch {
	case n <= ordinalWords:
		addToCertificate(ordinalNames[n-1])
	case n < twenty:
		addNumber(n)
		addToCertificate(message(msgTh))
	default:
		addToCertificate(message(msgTwent))
		if n == twenty {
			addToCertificate(message(msgIeth))
		} else {
			addToCertificate(message(msgY))
			addToCertificate(ordinalNames[n-firstTwenty])
		}
	}
}

// addCount adds n units, as "n unit(s)".
func addCount(n int, unit int) {
	addToCertificate(strconv.Itoa(n))
	addToCertificate(message(unit))
	if n > 1 {
		addToCertificate(message(msgS))
	}
}

// showCertificate shows the certificate of the quest and the time it took, and stays.
// C: ENDGAME_0326
func showCertificate() {
	setTextPosition(0, 1)
	printChar(charToggleInverse)
	printChar(charCentre)
	printMessage(msgBeItKnownThatOnLine)
	certificate = certificate[:0]
	addToCertificate(message(msgTheSp))
	addOrdinal(int(day))
	addToCertificate(message(msgDayOfLine))
	addToCertificate(message(msgTheSp))
	addOrdinal(int(month))
	addToCertificate(message(msgMonthLine))
	printMessage(msgOfTheYearLine)
	addNumber(int(int16(year)) / hundred)
	addToCertificate(message(msgHundredLine))
	addNumber(int(int16(year)) % hundred)
	addToCertificate("\n\n")
	addToCertificate(party[0].nameString())
	addToCertificate(message(msgTheAvatarLine))
	printMessage(msgSavedTheLifeLine)
	printMessage(msgOfOurSovereignLine)
	printMessage(msgLordBritishTherebyLine)
	printMessage(msgSavingOurPeopleLine)
	printMessage(msgAndOurLandLine)
	selectFont(1)
	printMessage(msgEQueOfEAvatarLine)
	printMessage(msgIsForeverLine)
	selectFont(0)
	printChar(charToggleInverse)
	printMessage(msgReportNowThyQuestCompleatIn)
	years := int(year) - startYear
	months := int(month) - startMonth
	days := int(day) - startDay
	if days < 0 {
		days += daysPerMonth
		months--
	}
	if months < 0 {
		months += monthsPerYear
		years--
	}
	if years != 0 {
		addCount(years, msgYear)
		if months != 0 || days != 0 {
			addToCertificate(", ")
		}
	}
	if months != 0 {
		addCount(months, msgMonth)
		if days != 0 {
			addToCertificate(", ")
		}
	}
	if days != 0 {
		addCount(days, msgDay)
	}
	addToCertificate("\n")
	printMessage(msgToLordBritishAtOriginSystems)
	for {
		waitTicks(1)
	}
}

// endingStep sounds a footstep between frames.
// C: ENDGAME_04fe
func endingStep() {
	sleepFrames(2)
	footstep()
	sleepFrames(3)
}

// walkToward moves actor slot a step toward (x, y), telling whether it moved.
// C: ENDGAME_0510
func walkToward(slot, x, y int) bool {
	a := &actors[slot]
	if a.tile == 0 || (int(a.x) == x && int(a.y) == y) {
		return false
	}
	dx := abs(int(a.x) - x)
	dy := abs(int(a.y) - y)
	switch {
	case dx < dy && int(a.y) > y:
		a.y--
	case dx < dy:
		a.y++
	case int(a.x) > x:
		a.x--
	default:
		a.x++
	}
	endingStep()
	return true
}

// wanderAtEnd moves actor slot a random step onto the floor, half the time.
// C: ENDGAME_05a2
func wanderAtEnd(slot int) {
	a := &actors[slot]
	if a.tile != 0 && randomRange(0, 1) != 0 {
		for range 8 {
			x, y := int(a.x), int(a.y)
			switch randomRange(0, 3) {
			case 0:
				x++
			case 1:
				x--
			case 2:
				y++
			case 3:
				y--
			}
			if *combatCell(x, y) == floorTile {
				a.x, a.y = byte(x), byte(y)
				break
			}
		}
	}
	sleepFrames(1)
}

// readYesNo reads Y or N, echoing the answer.
func readYesNo() int {
	defer asking(promptYesNo)()
	for {
		answer := getChar()
		switch answer {
		case answerYes:
			printString("Yes\n\n")
			return answer
		case answerNo:
			printString("No\n\n")
			return answer
		}
	}
}

// endgame brings the party before Lord British, who asks for the Codex to be destroyed:
// with the Sandalwood Box the quest ends, without it he stays a tyrant.
// C: ENDGAME_0648_EndgameMain
func endgame() {
	leaveWorld()
	musicFollowsMap = false
	music.PlayThen(music.JoyousReunion, music.RuleBritannia)
	combatant = noCombatant
	showStatus()
	changeTileSet(1)
	readFile("MISCMAPS.DAT", actorTiles[:endingMapBytes], endingMapOffset)
	readFile("ENDMSG.DAT", scratch[:endingMessageBytes], 0)
	for y := range viewSize {
		for x := range viewSize {
			*combatCell(x, y) = actorTiles[y*endingMapStride+x]
		}
	}
	for slot := range actorSlots {
		actors[slot].tile, actors[slot].animTile = 0, 0
	}
	british := &actors[britishSlot]
	british.tile, british.animTile = britishTile, britishTile
	british.x, british.y = boxX, britishStartY
	british.hull = 0
	sleepFrames(40)
	for walkToward(britishSlot, boxX, throneY) {
	}
	for row := range int(partySize) {
		m := &party[row]
		if m.status == statusDead {
			printChar('\n')
			printString(m.nameString())
			printMessage(msgLivesLine)
			m.status = statusGood
			m.hp = m.maxHP
			setPenColour(brightWhiteColour)
			xorFillRect(viewStart, viewStart, viewInnerEdge, viewInnerEdge)
			pulse(revivalFreq, 1, revivalLength, revivalWidthStart, 1)
			showStatus()
		}
		a := &actors[row]
		tile := classTiles[indexOf(classLetters, m.class)]
		a.tile, a.animTile = tile, tile
		a.x, a.y = boxX, memberStartY
		a.hull = 0
		sleepFrames(1)
		for walkToward(row, int(endingPlaceX[row]), int(endingPlaceY[row])) {
		}
	}
	sleepFrames(40)
	printString(scratchText(0))
	printString(party[0].nameString())
	printMessage(msgChar33Char34Newlines2)
	getChar()
	printString(scratchText(askAt))
	answer := readYesNo()
	if answer == answerNo {
		printString(scratchText(askAgainAt))
		answer = readYesNo()
	}
	if answer == answerYes && sandalwoodBox != 0 {
		sleepFrames(8)
		for walkToward(0, boxX, boxY) {
		}
		for walkToward(0, boxX, boxY+1) {
		}
		sleepFrames(4)
		box := &actors[boxSlot]
		box.tile, box.animTile = actorSandalwoodBox, actorSandalwoodBox
		box.x, box.y = boxX, boxY
		box.hull = 0
		printString(scratchText(boxAt))
		sleepFrames(40)
		printMessage(msgHeSaysLine)
		for _, at := range []int{speech1At, speech2At, speech3At, speech4At, speech5At} {
			getChar()
			printString(scratchText(at))
		}
		getChar()
		printString(scratchText(speech6At))
		box.tile, box.animTile = actorGem, actorGem
		getChar()
		pulse(codexFreq, 1, codexLength, codexWidth, 1)
		box.tile, box.animTile = 0, 0
		*combatCell(boxX, boxY) = codexTile
		for moongateRise = 1; moongateRise < moongateHeight; moongateRise++ {
			sleepFrames(1)
		}
		sleepFrames(4)
		for walkToward(britishSlot, boxX, boxY) {
		}
		british.tile, british.animTile = 0, 0
		sleepFrames(1)
		for row := range int(partySize) {
			for walkToward(row, boxX, boxY) {
			}
			actors[row].tile, actors[row].animTile = 0, 0
			sleepFrames(1)
		}
		for moongateRise = moongateHeight - 1; moongateRise > 0; moongateRise-- {
			sleepFrames(1)
		}
		putTile(floorTile, boxX, boxY)
		selectPage(1)
		setPenColour(0)
		// the original frees the tile set here (ULTIMA_102e_UnloadTileset) to make room
		fillRect(0, 0, screenRight, screenBottom)
		selectPage(0)
		showEnding()
		showCertificate()
		return
	}
	// without the Box, Lord British turns away
	printMessage(msgISeeLine)
	sleepFrames(40)
	printString(scratchText(badEndingAt))
	actors[0].y--
	endingStep()
	for {
		moved := walkToward(2, 8, 6)
		if walkToward(britishSlot, 4, 1) {
			moved = true
		}
		if walkToward(0, 8, 4) {
			moved = true
		}
		if !moved {
			break
		}
	}
	for {
		wanderAtEnd(1)
		wanderAtEnd(3)
		wanderAtEnd(4)
		wanderAtEnd(5)
	}
}
