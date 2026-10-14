package game

// Transferring a character from Ultima IV: the Avatar's PARTY.SAV becomes the new game's
// hero, its numbers scaled to Ultima V's (intro.c).

import "encoding/binary"

// Addresses of the tables in the data segment.
const (
	sheetRowsAddress   = 13910 // C: D_3656 (the row of each line of the sheet)
	sheetLabelsAddress = 13926 // C: D_3666
)

const (
	sheetLines  = 8
	transferMap = journeyMap + 1 // C: 66, the map the GOG release leaves before a transfer
)

var (
	sheetRows   []uint16 // C: D_3656
	sheetLabels []string // C: D_3666
)

// loadTransferTables reads the tables of the character sheet.
func loadTransferTables() {
	sheetRows = segment.Words(sheetRowsAddress, sheetLines)
	sheetLabels = messages(sheetLabelsAddress, sheetLines)
}

// The Ultima IV save file and its records.
const (
	u4SaveFile      = "PARTY.SAV"
	u4HeroAt        = 8   // the first member's record (S_bc88_U4Party)
	u4HeroBytes     = 40  // C: 0x28
	u4VirtuesAt     = 320 // the virtues (S_bb1c)
	u4VirtuesBytes  = 182 // C: 0xb6
	u4VirtueSkip    = 6
	u4Virtues       = 8
	u4NameAt        = 20
	u4NameBytes     = 8
	u4GenderAt      = 36
	u4ClassAt       = 37
	u4MaxStat       = 70
	u4MaxPoints     = 9999
	u4MaxClass      = 7
	firstPrintable  = ' '
	u4Hp            = 0 // word indexes of the record
	u4MaxHp         = 1
	u4Exp           = 2
	u4Str           = 3
	u4Dex           = 4
	u4Int           = 5
	u4Mag           = 6
	u4HpPerLevel    = 100
	u5ExpDivisor    = 10
	u5ExpPerLevel   = 100
	u5HpPerLevel    = 30
	u5MinStrength   = 20
	statsKeptBelow  = 10 // stats below this stay as they are
	statsHalvedFrom = 9
	statsQuartered  = 30 // stats from this are quartered above twenty
	statsHalfBase   = 10
	statsTopBase    = 20
)

// The classes of Ultima IV, in its order.
var u4Classes = []byte{classMage, classBard, classFighter, classDruid, classTinker, classPaladin, classRanger, classShepherd}

// u4IsAvatar tells whether the transferred hero had reached Avatarhood (all virtues at 0).
var u4IsAvatar bool // C: D_3304

// Layout of the sheet's frames.
const (
	sheetWindowRight = 19
	sheetRightLeft   = 21
	sheetBottom      = 18
	sheetPromptLeft  = 3
	sheetPromptRight = 37
	sheetPromptRow   = 21
	rightSheetX      = 168
	framePromptRow   = 19
	frameWidth       = 38
	frameSides       = 3
	frameEdge        = 39
	frameLineTop     = 159
	frameLineRight   = 312
	frameLineBottom  = 184
	frameLineLeft    = 7
	sheetInnerRight  = 6
	sheetPanelRight  = 143
	sheetPanelEdge   = 151
	sheetPanelTop    = 137
	sheetPanelInner  = 150
	sheetPanelBottom = 143
	sheetBottomRow   = 17
	sheetRightCorner = 18
	sheetTitleLine   = 24
	sheetLineY       = 7
	sheetLineBottom  = 136
	sheetTitleEnd    = 128
	finalWindowLeft  = 2
	finalWindowTop   = 21
	finalWindowRight = 37
	finalWindowBot   = 22
	glyphFrameTL     = 123
	glyphFrameTR     = 124
	glyphFrameBL     = 125
	glyphFrameBR     = 126
	glyphFrameBar    = 127
)

// convertU4Save reads the hero of the Ultima IV save into the first party member, telling
// whether the save held bad data.
// C: INTRO_1016_ConvertU4Savegame
func convertU4Save() bool {
	record := make([]byte, u4HeroBytes)
	bad := !disk.Exists(u4SaveFile) // fix: the original reads whatever is on the disk
	if !bad {
		readFile(u4SaveFile, record, u4HeroAt)
	}
	word := func(i int) int { return int(binary.LittleEndian.Uint16(record[i*2:])) }
	if word(u4Str) > u4MaxStat || word(u4Dex) > u4MaxStat || word(u4Int) > u4MaxStat ||
		word(u4Exp) > u4MaxPoints || word(u4Hp) > u4MaxPoints || word(u4MaxHp) > u4MaxPoints ||
		int(record[u4ClassAt]) > u4MaxClass {
		bad = true
	}
	name := record[u4NameAt : u4NameAt+u4NameBytes]
	for _, c := range name {
		if c != 0 && c < firstPrintable {
			bad = true
			break
		}
	}
	if bad {
		printChar(charClearWindow)
		printChar(charCentre)
		setTextPosition(0, 5)
		printMessage(msgErrorYourUltimaIvGameContains)
		printMessage(msgUnableToContinueTransferLine)
		printMessage(msgPressAnyKeyToReturnTo)
		printChar(charUncentre)
		waitForKey(0)
		return true
	}
	hero := &party[0]
	n := 0
	for ; n < u4NameBytes && name[n] != 0; n++ {
		hero.name[n] = name[n]
	}
	hero.name[n] = 0
	if record[u4GenderAt] == maleGender {
		hero.gender = maleGender
	} else {
		hero.gender = femaleGend
	}
	hero.class = u4Classes[record[u4ClassAt]]
	hero.status = statusGood
	hero.str = byte(word(u4Str))
	hero.dex = byte(word(u4Dex))
	hero.intel = byte(word(u4Int))
	hero.magic = byte(word(u4Mag))
	hero.hp = int16(word(u4Hp))
	hero.maxHP = int16(word(u4MaxHp))
	hero.exp = int16(word(u4Exp))
	hero.level = byte(word(u4MaxHp) / u4HpPerLevel)
	virtues := make([]byte, u4VirtuesBytes)
	readFile(u4SaveFile, virtues, u4VirtuesAt)
	avatar := true
	for i := range u4Virtues {
		if binary.LittleEndian.Uint16(virtues[u4VirtueSkip+i*2:]) != 0 {
			avatar = false
		}
	}
	if avatar {
		u4IsAvatar = true
	}
	return false
}

// printU4Class prints the hero's class.
// C: INTRO_1278_PrintU4Class
func printU4Class() {
	switch party[0].class {
	case classMage:
		printMessage(msgMageLine)
	case classBard:
		printMessage(msgBardLine)
	case classFighter:
		printMessage(msgFighterLine)
	case classDruid:
		printMessage(msgDruidLine)
	case classTinker:
		printMessage(msgTinkerLine)
	case classPaladin:
		printMessage(msgPaladinLine)
	case classRanger:
		printMessage(msgRangerLine)
	case classShepherd:
		printMessage(msgShepherdLine)
	}
}

// scaleStat scales an Ultima IV stat (to 50) to Ultima V's (to 30).
// C: INTRO_12ea
func scaleStat(v int) int {
	if v < statsKeptBelow {
		return v
	}
	if v >= statsQuartered {
		return (v-statsQuartered)/4 + statsTopBase
	}
	return (v-statsHalvedFrom)/2 + statsHalfBase
}

// drawSheetLabel prints the label of line on both sheets.
// C: INTRO_1e22
func drawSheetLabel(line int) {
	for _, w := range []int{0, 1} {
		selectWindow(w)
		setTextPosition(3, int(sheetRows[line]))
		printString(sheetLabels[line])
	}
}

// drawPromptFrame draws the frame of the prompt under the sheets.
// C: INTRO_1e62
func drawPromptFrame() {
	setTextColour(blueColour)
	setTextPosition(0, framePromptRow)
	printChar(glyphFrameTL)
	for range frameWidth {
		printChar(glyphFrameBar)
	}
	printChar(glyphFrameTR)
	for range frameSides {
		printChar(glyphFrameBar)
		setTextPosition(frameEdge, textY())
		printChar(glyphFrameBar)
	}
	printChar(glyphFrameBL)
	for range frameWidth {
		printChar(glyphFrameBar)
	}
	printChar(glyphFrameBR)
	setPenColour(brightWhiteColour)
	line(frameLineLeft, frameLineTop, frameLineRight, frameLineTop)
	lineTo(frameLineRight, frameLineBottom)
	lineTo(frameLineLeft, frameLineBottom)
	lineTo(frameLineLeft, frameLineTop)
	setTextColour(brightWhiteColour)
}

// drawSheetFrame draws the frame of a sheet from x, titled "Ultima IV".
// C: INTRO_1f26
func drawSheetFrame(x int) {
	setPenColour(blueColour)
	fillRect(x, 0, x+sheetInnerRight, sheetPanelBottom)
	fillRect(x+sheetPanelRight, 0, x+sheetPanelEdge, sheetPanelTop)
	fillRect(x+sheetLineY, sheetPanelTop, x+sheetPanelInner, sheetPanelBottom)
	setTextColour(blueColour)
	setTextPosition(0, 0)
	printString("\x7b\x7f\x7f")
	printOpenCaption()
	setTextColour(brightWhiteColour)
	printMessage(msgUltimaIvSp)
	setTextColour(blueColour)
	printCloseCaption()
	printString("\x7f\x7f")
	printChar(glyphFrameTR)
	setTextPosition(0, sheetBottomRow)
	printChar(glyphFrameBL)
	setTextPosition(sheetRightCorner, sheetBottomRow)
	printChar(glyphFrameBR)
	setPenColour(brightWhiteColour)
	line(x+sheetTitleLine, sheetLineY, x+sheetLineY, sheetLineY)
	lineTo(x+sheetLineY, sheetLineBottom)
	lineTo(x+sheetPanelRight, sheetLineBottom)
	lineTo(x+sheetPanelRight, sheetLineY)
	lineTo(x+sheetTitleEnd, sheetLineY)
	setTextColour(brightWhiteColour)
}

// drawSheets sets up the two sheets, Ultima IV's and Ultima V's, and the prompt.
// C: INTRO_2024
func drawSheets() {
	drawPromptFrame()
	setWindowSize(0, 0, 0, sheetWindowRight, sheetBottom)
	setWindowSize(1, sheetRightLeft, 0, textColumns-1, sheetBottom)
	setWindowSize(2, sheetPromptLeft, sheetPromptRow, sheetPromptRight, sheetPromptRow)
	selectWindow(0)
	drawSheetFrame(0)
	selectWindow(1)
	drawSheetFrame(rightSheetX)
	setTextPosition(12, 0)
	printChar(' ')
}

// labelAt prints a sheet label at (x, y), highlighted when lit.
func labelAt(x, y int, label int, lit bool) {
	setTextPosition(x, y)
	if lit {
		printChar(charToggleInverse)
	}
	printMessage(label)
	if lit {
		printChar(charToggleInverse)
	}
}

// moveHighlight unlights the label at (x1, y1) and lights the one at (x2, y2) in window w.
func moveHighlight(w, x1, y1, from, x2, y2, to int) {
	selectWindow(w)
	labelAt(x1, y1, from, false)
	labelAt(x2, y2, to, true)
}

// promptLine shows text in the prompt, centred.
func promptLine(text int) {
	selectWindow(2)
	printChar(charClearWindow)
	printChar(charCentre)
	printMessage(text)
	printChar(charUncentre)
}

// readYN waits for Y or N.
func readYN() int {
	for {
		key := toUpper(waitForKey(0))
		if key == 'Y' || key == 'N' {
			return key
		}
	}
}

// printClassOrAvatar prints "Avatar" for an Avatar, else the class.
func printClassOrAvatar(avatar int) {
	if u4IsAvatar {
		printMessage(avatar)
	} else {
		printU4Class()
	}
}

// Rows of the sheet's lines.
const (
	rowName  = 2
	rowSex   = 5
	rowClass = 6
	rowExp   = 8
	rowLevel = 9
	rowStr   = 11
	rowDex   = 12
	rowInt   = 13
	rowKind  = 15
	valueX   = 10
)

// transferFromUltima4 makes the hero of an Ultima IV game the hero of a new game, showing
// the conversion on two sheets.
// C: INTRO_132a_TransferFromU4
func transferFromUltima4() {
	// the GOG release reads the Ultima IV save from the game's own drive without asking
	if mapID == transferMap {
		mapID = journeyMap
	}
	keptMap := mapID
	contents := make([]byte, saveSize)
	readFile("INIT.GAM", contents, 0)
	if err := loadSave(contents); err != nil {
		diskError(err)
	}
	ool := make([]byte, 2*oolSize)
	readFile("INIT.OOL", ool[oolSize:], 0)
	setWindowSize(0, 0, 0, textColumns-1, textRows-1)
	selectWindow(0)
	printChar(charClearWindow)
	if convertU4Save() {
		mapID = keptMap
		return
	}
	hero := &party[0]
	if u4IsAvatar {
		hero.class = classAvatar
	}
	printChar(charClearWindow)
	setTextPosition(0, 11)
	printChar(charCentre)
	printMessage(msgFoundLine)
	setTextPosition(0, 12)
	printString(hero.nameString())
	printChar(charUncentre)
	setTextPosition(12, 13)
	printMessage(msgALevelSp)
	printNumber(int(hero.level), 1, ' ')
	if hero.gender == maleGender {
		printMessage(msgMaleSp)
	} else {
		printMessage(msgFemaleSp)
	}
	printU4Class()
	for i, stat := range []struct {
		label int
		value byte
	}{{msgStrSp, hero.str}, {msgDexSp, hero.dex}, {msgIntSp, hero.intel}} {
		setTextPosition(17, 15+i)
		printMessage(stat.label)
		printNumber(int(stat.value), 1, ' ')
	}
	setTextPosition(10, 20)
	printString(hero.nameString())
	printMessage(msgIsSp)
	if u4IsAvatar {
		printMessage(msgAnAvatar)
	} else {
		printMessage(msgNotAnAvatar)
	}
	waitForKey(0)
	printChar(charClearWindow)

	drawSheets()
	for i := range sheetLines {
		drawSheetLabel(i)
	}
	selectWindow(0)
	setTextPosition(0, 3)
	printChar(charCentre)
	printString(hero.nameString())
	printChar(charUncentre)
	setTextPosition(valueX, rowSex)
	if hero.gender == maleGender {
		printMessage(msgMale)
	} else {
		printMessage(msgFemale)
	}
	setTextPosition(valueX, rowClass)
	printClassOrAvatar(msgAvatar)
	for _, v := range []struct{ row, value int }{
		{rowExp, int(hero.exp)}, {rowLevel, int(hero.level)}, {rowStr, int(hero.str)},
		{rowDex, int(hero.dex)}, {rowInt, int(hero.intel)},
	} {
		setTextPosition(valueX, v.row)
		printNumber(v.value, 1, ' ')
	}
	setTextPosition(0, rowKind)
	printChar(charCentre)
	if u4IsAvatar {
		printMessage(msgAvatar)
	} else {
		printMessage(msgNonAvatar)
	}
	printChar(charUncentre)
	labelAt(7, rowName, msgName, true)
	selectWindow(1)
	labelAt(7, rowName, msgName, true)

	selectWindow(2)
	printChar(charClearWindow)
	setTextPosition(10, 0)
	printMessage(msgKeepThisName)
	answer := readYN()
	printChar(charClearWindow)
	if answer == 'N' {
		for {
			setTextPosition(1, 0)
			printMessage(msgEnterNewNameSp)
			name := readLine(u4NameBytes)
			hero.name = [nameBytes]byte{}
			copy(hero.name[:], name)
			if hero.name[0] != 0 {
				break
			}
		}
	}
	selectWindow(1)
	setTextPosition(0, 3)
	printChar(charCentre)
	printString(hero.nameString())
	printChar(charUncentre)
	moveHighlight(1, 7, rowName, msgName, 5, rowSex, msgSex)
	moveHighlight(0, 7, rowName, msgName, 5, rowSex, msgSex)

	promptLine(msgKeepSameSex)
	answer = readYN()
	selectWindow(1)
	setTextPosition(valueX, rowSex)
	if (answer == 'Y' && hero.gender == maleGender) || (answer == 'N' && hero.gender == femaleGend) {
		printMessage(msgMale)
		hero.gender = maleGender
	} else {
		printMessage(msgFemale)
		hero.gender = femaleGend
	}
	labelAt(5, rowSex, msgSex, false)
	labelAt(3, rowClass, msgClass, true)
	setTextPosition(valueX, rowClass)
	printClassOrAvatar(msgAvatar)
	moveHighlight(0, 5, rowSex, msgSex, 3, rowClass, msgClass)
	selectWindow(2)
	printChar(charClearWindow)
	setTextPosition(2, 0)
	if u4IsAvatar {
		printMessage(msgThouArtNowAnAvatar)
	} else {
		printMessage(msgClassRemainsIntact)
	}
	waitForKey(0)

	moveHighlight(0, 3, rowClass, msgClass, 5, rowExp, msgExp)
	moveHighlight(1, 3, rowClass, msgClass, 5, rowExp, msgExp)
	setTextPosition(valueX, rowExp)
	hero.exp /= u5ExpDivisor
	printNumber(int(hero.exp), 1, ' ')
	promptLine(msgExperienceHasBeenConverted)
	waitForKey(0)

	moveHighlight(0, 5, rowExp, msgExp, 3, rowLevel, msgLevel)
	moveHighlight(1, 5, rowExp, msgExp, 3, rowLevel, msgLevel)
	setTextPosition(valueX, rowLevel)
	hero.level = 1
	for v := int(hero.exp) / u5ExpPerLevel; v > 0; v >>= 1 {
		hero.level++
	}
	hero.maxHP = int16(hero.level) * u5HpPerLevel
	hero.hp = hero.maxHP
	printNumber(int(hero.level), 1, ' ')
	promptLine(msgLevelHasBeenConverted)
	waitForKey(0)

	moveHighlight(0, 3, rowLevel, msgLevel, 5, rowStr, msgStr)
	moveHighlight(1, 3, rowLevel, msgLevel, 5, rowStr, msgStr)
	setTextPosition(valueX, rowStr)
	was := int(hero.str)
	hero.str = byte(max(scaleStat(was), u5MinStrength))
	printNumber(int(hero.str), 1, ' ')
	showConversion(msgStrengthWasSp, was, int(hero.str))

	moveHighlight(0, 5, rowStr, msgStr, 5, rowDex, msgDex)
	moveHighlight(1, 5, rowStr, msgStr, 5, rowDex, msgDex)
	setTextPosition(valueX, rowDex)
	was = int(hero.dex)
	hero.dex = byte(scaleStat(was))
	printNumber(int(hero.dex), 1, ' ')
	showConversion(msgDexterityWasSp, was, int(hero.dex))

	moveHighlight(0, 5, rowDex, msgDex, 5, rowInt, msgInt)
	moveHighlight(1, 5, rowDex, msgDex, 5, rowInt, msgInt)
	setTextPosition(valueX, rowInt)
	was = int(hero.intel)
	hero.intel = byte(scaleStat(was))
	hero.magic = hero.intel
	printNumber(int(hero.magic), 1, ' ')
	showConversion(msgIntellectWasSp, was, int(hero.intel))

	selectWindow(0)
	labelAt(5, rowInt, msgInt, false)
	selectWindow(1)
	labelAt(5, rowInt, msgInt, false)
	setTextPosition(0, rowKind)
	printChar(charCentre)
	if u4IsAvatar {
		printMessage(msgAvatar)
	} else {
		printMessage(msgNonAvatar)
	}
	printChar(charUncentre)

	selectWindow(2)
	setWindowSize(2, finalWindowLeft, finalWindowTop, finalWindowRight, finalWindowBot)
	printChar(charClearWindow)
	printMessage(msgConversionCompleteSavingLine)
	writeFile("SAVED.OOL", ool)
	writeFile("SAVED.GAM", encodeSave())
	mapID = keptMap
}

// showConversion tells how a stat was converted and waits for a key.
func showConversion(label, was, now int) {
	selectWindow(2)
	printChar(charClearWindow)
	setTextPosition(1, 0)
	printMessage(label)
	printNumber(was, 1, ' ')
	printMessage(msg50NowSp)
	printNumber(now, 1, ' ')
	printMessage(msg30)
	waitForKey(0)
}
