package game

import "github.com/memmaker/u5-go/internal/music"
import "github.com/memmaker/u5-go/internal/ega"

// The FONT overlay: the proportional story text, the "Return to the View" demo that plays
// on the title screen and the character creation.

// Addresses of the overlay's tables in the data segment.
const (
	glyphWidthsAddress    = 20715 // C: D_50eb (the width of each glyph from '!' on)
	sceneNamesAddress     = 40976 // C: D_515c (the strings of the demo's four scenes)
	intBonusAddress       = 20836 // C: D_5164
	dexBonusAddress       = 20844 // C: D_516c
	strBonusAddress       = 20852 // C: D_5174
	questionOffsetAddress = 20860 // C: D_517c (offset in QUESTION.DAT of each pair of virtues)
	virtueXAddress        = 20988 // C: D_51fc
	virtueYAddress        = 20996 // C: D_5204
	originWidthsAddress   = 21254 // C: D_5306
	originHeightsAddress  = 21268 // C: D_5314
)

const (
	glyphCount    = 90
	originLines   = 7
	sceneCount    = 4
	firstGlyph    = '!'
	wideGlyph     = '{' // a 15 pixel gap
	wideGlyphGap  = 15
	newlineGlyph  = '_' // drawn as glyph 13 where a line ends
	newlineShape  = 13
	glyphOffset   = ' ' // glyph n of the font is the character n+32
	storyLineStep = 9
	storyBottom   = 192
)

// Overlay tables read from DATA.OVL.
var (
	glyphWidths      []byte // C: D_50eb
	sceneNames       [sceneCount]int
	intBonus         []byte     // C: D_5164
	dexBonus         []byte     // C: D_516c
	strBonus         []byte     // C: D_5174
	questionOffsets  [][]uint16 // C: D_517c
	virtueX, virtueY []byte     // C: D_51fc, D_5204
	originWidths     []int      // C: D_5306
	originHeights    []int      // C: D_5314
)

// loadFontTables reads the FONT overlay's tables.
func loadFontTables() {
	glyphWidths = segment.Bytes(glyphWidthsAddress, glyphCount)
	sceneNames = [sceneCount]int{sceneNamesAddress}
	for i := 1; i < sceneCount; i++ {
		sceneNames[i] = sceneNames[i-1] + len(message(sceneNames[i-1])) + 1
	}
	intBonus = segment.Bytes(intBonusAddress, virtueCount)
	dexBonus = segment.Bytes(dexBonusAddress, virtueCount)
	strBonus = segment.Bytes(strBonusAddress, virtueCount)
	questionOffsets = make([][]uint16, virtueCount)
	for i := range questionOffsets {
		questionOffsets[i] = segment.Words(questionOffsetAddress+i*virtueCount*2, virtueCount)
	}
	virtueX = segment.Bytes(virtueXAddress, virtueCount)
	virtueY = segment.Bytes(virtueYAddress, virtueCount)
	originWidths = segment.SignedWords(originWidthsAddress, originLines)
	originHeights = segment.SignedWords(originHeightsAddress, originLines)
}

// The layout of the story text: two column ranges, the second used between two heights
// (to flow around a picture), the pen position and the width of a space.
var (
	storyLeft     = [2]int{0, 0}                 // C: D_5146
	storyRight    = [2]int{ega.Width, ega.Width} // C: D_514c
	storyNarrowY  = 200                          // C: D_5150
	storyWideY    = 200                          // C: D_5152
	storySpace    = 5                            // C: D_5154
	storyX        = 0                            // C: D_5156
	storyY        = 60                           // C: D_5158
	demoNoiseStep = 0                            // C: D_515a
	demoLeft      byte                           // C: D_bd26
	demoRight     byte                           // C: D_bd27
	demoGrow      bool                           // C: D_bd28
	demoScene     byte                           // C: D_bd29
	chosenVirtue  [virtueCount]bool              // C: D_bd2a
	lostVirtue    [virtueCount]bool              // C: D_bd32
	newInt        byte                           // C: D_bd3c
	newDex        byte                           // C: D_bd3d
	newStr        byte                           // C: D_bd3e
	scratch       [scratchSize]byte              // C: D_b21e
)

// scratchSize is the size of the scratch buffer files are read into.
const scratchSize = 4096

// markWidth is the advance of the paragraph mark.
func markWidth() int { return int(glyphWidths[newlineShape-1]) + 1 }

// glyphWidth is the advance of a printable character.
func glyphWidth(c byte) int { return int(glyphWidths[c-firstGlyph]) + 1 }

// storyColumn picks the column range for the pen's height.
func storyColumn() int {
	if storyY > storyNarrowY && storyY < storyWideY {
		return 1
	}
	return 0
}

// printStory sets text in the proportional font of pictures on the back buffer,
// justified, wrapping at the column's edge; '_' ends a paragraph with a mark.
// C: FONT_0000
func printStory(font []byte, text []byte) {
	selectPage(1)
	at := func(i int) byte {
		if i < len(text) {
			return text[i]
		}
		return 0
	}
	column := storyColumn()
	width := storyRight[column] - storyLeft[column]
	start := 0
	used := storyX - storyLeft[column]
	for at(start) != 0 {
		gaps := 0
		end := start
		for at(end) != 0 && used < width && at(end) != '\n' {
			switch c := at(end); {
			case c <= ' ':
				used += storySpace
				gaps++
			case c == wideGlyph:
				used += wideGlyphGap
			case c != newlineGlyph:
				used += glyphWidth(c)
			}
			end++
		}
		// Back up to the last space that lets the line fit.
		for {
			if at(end) == 0 && used < width {
				break
			}
			if at(end) == '\n' {
				break
			}
			end--
			if end <= start {
				break
			}
			c := at(end)
			if c == ' ' {
				used -= storySpace
				gaps--
				break
			}
			// The paragraph mark is as wide as glyph 12 (D_50f7).
			if c == newlineGlyph && used+markWidth() < width {
				used += markWidth()
				break
			}
			if c != newlineGlyph && c != wideGlyph {
				used -= glyphWidth(c)
			}
		}
		slack := width - used
		lineEnd := end
		for i := start; i < lineEnd; i++ {
			switch c := at(i); {
			case c <= ' ':
				storyX += storySpace
				if gaps != 0 && at(lineEnd) != 0 && at(lineEnd) != '\n' {
					storyX += slack / gaps
					slack -= slack / gaps
					gaps--
				}
			case c == wideGlyph:
				storyX += wideGlyphGap
			case c != newlineGlyph:
				if storyY < storyBottom {
					putBitPicture(font, int(c)-glyphOffset, storyX, storyY)
				}
				storyX += glyphWidth(c)
			}
		}
		start = lineEnd
		if at(start) != 0 {
			if at(start) == newlineGlyph && storyY < storyBottom {
				putBitPicture(font, newlineShape, storyX, storyY)
			}
			used = 0
			storyY += storyLineStep
			column = storyColumn()
			storyX = storyLeft[column]
			width = storyRight[column] - storyX
			start++
		}
	}
	selectPage(0)
}

// The "Return to the View" demo plays on map 66 with the map's top four rows as the
// scenery and the next four as the actors drawn over it.
const (
	demoRows       = 4
	demoColumns    = 19
	demoActorLayer = demoRows * settlementSize
	demoTop        = 7 // the view row the scenery starts at
	demoMiddle     = 9
	firstActorTile = 256
	// demoGap marks a scenery cell left undrawn.
	demoGap        = mapTile254
	demoScriptAt   = 704 // offset of the demo in MISCMAPS.DAT
	demoScriptSize = 2000
	demoScriptBody = 512 // the scripts start after the four scenes
	sceneBytes     = demoRows * settlementSize
	sceneStorm     = 2
	sceneChimes    = 3
	chimeSteps     = 8
	// moongateRows is how many steps a moongate takes to rise or sink.
	moongateRows = 16
)

// demoDeltaX, demoDeltaY move a demo actor north, east, south or west.
var (
	demoDeltaX = [4]int{0, 1, 0, -1} // C: D_24d6
	demoDeltaY = [4]int{-1, 0, 1, 0} // C: D_24de
)

// demoMap and demoBackup are a cell of the demo's scenery and its saved copy.
func demoMap(x, y int) *byte    { return &mapTiles[y*settlementSize+x] }
func demoBackup(x, y int) *byte { return &mapBackup[y*settlementSize+x] }

// drawDemoCell draws one cell of the demo: the scenery or, where an actor stands, the
// actor.
// C: FONT_02a2
func drawDemoCell(x, y int) {
	if tile := *demoMap(x, y); tile != 0 {
		putTile(int(tileShown[tile]), x, y+demoTop)
		return
	}
	tile := mapTiles[x+y*settlementSize+demoActorLayer]
	if tile == actorCircle {
		return
	}
	putTile(firstActorTile+int(tile), x, y+demoTop)
}

// animateTitle runs frames of the demo: it animates the actors and the banner, widens
// the drawn part of the scene every other frame and makes the scene's noises. It returns
// 1 when a key cut it short.
// C: FONT_02fc
func animateTitle(frames int) int {
	for {
		animateActors()
		animateBanner()
		for i := range actors {
			a := &actors[i]
			if a.tile != 0 {
				at := int(a.x) + int(a.y)*settlementSize
				mapTiles[at] = 0
				mapTiles[at+demoActorLayer] = a.animTile
			}
		}
		for x := int(demoLeft); x <= int(demoRight); x++ {
			for y := range demoRows {
				if *demoMap(x, y) != demoGap {
					drawDemoCell(x, y)
				}
			}
		}
		if demoLeft != 0 && demoGrow {
			demoLeft--
			demoRight++
		}
		demoGrow = !demoGrow
		if pollKey() != 0 {
			return 1
		}
		if frames != 0 {
			waitTicks(1)
		}
		switch demoScene {
		case sceneStorm:
			whiteNoise(20, 60, 10000)
		case sceneChimes:
			demoNoiseStep++
			if demoNoiseStep >= chimeSteps {
				demoNoiseStep = 0
			}
			switch demoNoiseStep {
			case 0:
				tone(3000, 3)
			case chimeSteps / 2:
				tone(2000, 3)
			}
		}
		frames--
		if frames == 0 {
			return 0
		}
	}
}

// showScene names a scene of the demo and lays out its scenery from the scratch buffer.
// C: FONT_0418
func showScene(scene int) {
	showTitleCaption(message(sceneNames[scene]))
	for x := range demoColumns {
		for y := range demoRows {
			tile := scratch[scene*sceneBytes+x+y*settlementSize]
			*demoMap(x, y) = tile
			*demoBackup(x, y) = tile
		}
	}
	demoLeft, demoRight = demoMiddle, demoMiddle
	demoGrow = false
	demoScene = byte(scene)
}

// The demo script's opcodes.
const (
	demoPlace = iota
	demoRemove
	demoStep
	demoWait
	demoRaiseGate
	demoLowerGate
	demoShowScene
	demoFadeIn
	demoFadeOut
	demoRestart
	demoSetTile
	demoBolt
	demoClear
	demoWalk
	demoRepeat
	demoLoop
)

// returnToView plays the demo until a key is pressed.
// C: FONT_04a4
func returnToView() {
	video.tileYOffset = 16
	demoLeft, demoRight = demoMiddle, demoMiddle
	demoGrow = false
	setPenColour(blueColour)
	fillRect(120, 120, 199, 126)
	setPenColour(brightWhiteColour)
	line(120, 127, 199, 127)
	if mapID != titleMap {
		readFile("MISCMAPS.DAT", scratch[:demoScriptSize], demoScriptAt)
		if mapID == introMap {
			for !loadTiles(tileFile) {
			}
		}
	}
	mapID = titleMap
	var gateX, gateY, repeats, repeatAt int
	next := func(at *int) int {
		*at++
		return int(scratch[*at])
	}
	restore := func(a *actor) {
		*demoMap(int(a.x), int(a.y)) = *demoBackup(int(a.x), int(a.y))
	}
	for at := demoScriptBody; at < demoScriptSize; at++ {
		switch scratch[at] {
		case demoPlace:
			a := &actors[next(&at)]
			a.tile = byte(next(&at))
			a.animTile = a.tile
			a.x = byte(next(&at))
			a.y = byte(next(&at))
			a.state = 0
		case demoRemove:
			a := &actors[next(&at)]
			a.tile, a.animTile, a.state = 0, 0, 0
			*demoMap(int(a.x), int(a.y)) = *demoBackup(int(a.x), int(a.y))
		case demoStep, demoWalk:
			op := scratch[at]
			a := &actors[next(&at)]
			direction := next(&at)
			restore(a)
			a.x = byte(int(a.x) + demoDeltaX[direction])
			a.y = byte(int(a.y) + demoDeltaY[direction])
			if op == demoWalk && animateTitle(7) != 0 {
				return
			}
		case demoWait:
			if animateTitle(next(&at)) != 0 {
				return
			}
		case demoRaiseGate:
			gateX = next(&at)
			gateY = next(&at)
			*demoMap(gateX, gateY), *demoBackup(gateX, gateY) = demoGap, demoGap
			for rows := 1; rows < moongateRows; rows++ {
				putMoongate(rows, gateX, gateY+demoTop)
				if animateTitle(1) != 0 {
					return
				}
			}
			*demoMap(gateX, gateY), *demoBackup(gateX, gateY) = mapTileMoongate, mapTileMoongate
			if animateTitle(2) != 0 {
				return
			}
		case demoLowerGate:
			*demoMap(gateX, gateY), *demoBackup(gateX, gateY) = demoGap, demoGap
			for rows := moongateRows - 1; rows > 0; rows-- {
				putMoongate(rows, gateX, gateY+demoTop)
				if animateTitle(1) != 0 {
					return
				}
			}
			*demoMap(gateX, gateY), *demoBackup(gateX, gateY) = mapTileGrass, mapTileGrass
			if animateTitle(2) != 0 {
				return
			}
		case demoShowScene:
			showScene(next(&at))
		case demoFadeIn:
			a := &actors[next(&at)]
			tile := a.tile
			a.tile, a.animTile = actorCircle, actorCircle
			if revealTile(firstActorTile+int(tile), int(a.x), int(a.y)+demoTop) != 0 {
				return
			}
			a.tile, a.animTile = tile, tile
		case demoFadeOut:
			a := &actors[next(&at)]
			x, y, tile := int(a.x), int(a.y), a.tile
			a.tile, a.animTile = actorCircle, actorCircle
			if revealTile(int(*demoBackup(x, y)), x, y+demoTop) != 0 {
				return
			}
			a.tile, a.animTile = tile, tile
		case demoRestart:
			at = demoScriptBody - 1
		case demoSetTile:
			tile := byte(next(&at))
			x := next(&at)
			y := next(&at)
			*demoBackup(x, y), *demoMap(x, y) = tile, tile
		case demoBolt:
			for step := range 5 {
				if animateTitle(1) != 0 {
					return
				}
				x, y := step*9, step*3
				setPenColour(brightWhiteColour)
				line(x+128, y+152, x+137, y+155)
				line(x+128, y+153, x+137, y+156)
			}
			at += 2
			a := &actors[scratch[at]]
			putTile(0, int(a.x), int(a.y)+demoTop)
			whiteNoise(1, 1200, 4000)
			if animateTitle(3) != 0 {
				return
			}
		case demoClear:
			for i := range actors {
				actors[i].animTile, actors[i].tile, actors[i].state = 0, 0, 0
			}
		case demoRepeat:
			repeats = next(&at)
			repeatAt = at
		case demoLoop:
			repeats--
			if repeats != 0 {
				at = repeatAt
			}
		}
		if pollKey() != 0 {
			return
		}
	}
}

// pickVirtue picks a virtue neither asked about in this round nor given up.
// C: FONT_0998
func pickVirtue() int {
	for {
		v := randomRange(0, virtueCount-1)
		if !chosenVirtue[v] && !lostVirtue[v] {
			chosenVirtue[v] = true
			return v
		}
	}
}

// askQuestion lets the gypsy ask about two virtues and adds the bonus of the one chosen.
// C: FONT_09c8
func askQuestion(font, pictures []byte) {
	selectPage(1)
	printChar(charClearWindow)
	putPicture(pictures, 1, 16, 0, 0)
	putPicture(pictures, 1, 200, 0, 0)
	storyX, storyY = 0, 152
	first := pickVirtue()
	second := pickVirtue()
	left, right, leftKey := first, second, byte('A')
	if first > second {
		left, right, leftKey = second, first, 'B'
	}
	putPicture(pictures, left+2, int(virtueX[left]), int(virtueY[left]), 0)
	putPicture(pictures, right+2, int(virtueX[right])+184, int(virtueY[right]), 0)
	readFile("QUESTION.DAT", scratch[:demoScriptSize], int(questionOffsets[first][second]))
	printStory(font, scratch[:demoScriptSize])
	copyPage(1, 0)
	var key byte
	for {
		key = byte(toUpper(pollKey()))
		if key == 'A' || key == 'B' {
			break
		}
	}
	if key != leftKey {
		first, second = second, first
	}
	newInt += intBonus[first]
	newDex += dexBonus[first]
	newStr += strBonus[first]
	lostVirtue[second] = true
}

// Character creation's layout.
const (
	nameLength  = 8
	minStrength = 20
	maleGender  = 11
	femaleGend  = 12
	oolSize     = 256
	questionEnd = 802 // offset of the closing text in QUESTION.DAT
)

// createCharacter asks the player's name and sex, lets the gypsy ask her questions and
// writes the new game.
// C: FONT_0b0a
func createCharacter() {
	playSong(music.AmigaTheme)
	if mapID != introMap {
		screen.UnloadTiles()
	}
	font := loadPictures(proportionalFont)
	pictures := loadPictures("CREATE.16")
	if contents, err := disk.ReadFile("INIT.GAM"); err != nil || loadSave(contents) != nil {
		diskError(err)
	}
	setPenColour(blueColour)
	fillRect(120, 120, 200, 126)
	setPenColour(brightWhiteColour)
	line(120, 127, 200, 127)
	setPenColour(0)
	fillRect(8, 128, 311, 191)
	setTextPosition(3, 17)
	printMessage(msgByWhatNameShaltThouBe)
	setTextPosition(14, 19)
	printChar(':')
	name := readLine(nameLength)
	hero := &party[0]
	clear(hero.name[:])
	copy(hero.name[:], name)
	if hero.name[0] != 0 {
		setTextPosition(8, 21)
		printMessage(msgArtThouMaleOrFemaleSp)
		var key int
		for {
			key = toUpper(waitForKey(0))
			if key == 'M' || key == 'F' {
				break
			}
		}
		printChar(key)
		hero.gender = femaleGend
		if key == 'M' {
			hero.gender = maleGender
		}
		storyX, storyY = 0, 9
		storyLeft = [2]int{0, 175}
		storyRight = [2]int{ega.Width, ega.Width}
		storyNarrowY, storyWideY = 89, 200
		readFile("QUESTION.DAT", scratch[:demoScriptSize], 0)
		selectPage(1)
		printChar(charClearWindow)
		putPicture(pictures, 0, 0, 96, 0)
		printStory(font, scratch[:demoScriptSize])
		copyPage(1, 0)
		storyNarrowY = 200
		for pollKey() == 0 {
		}
		newInt, newDex, newStr = hero.intel, hero.dex, hero.str
		chosenVirtue = [virtueCount]bool{}
		lostVirtue = [virtueCount]bool{}
		// Four questions, then two among the winners, then the final one.
		for _, round := range []int{4, 2, 1} {
			for range round {
				askQuestion(font, pictures)
			}
			chosenVirtue = [virtueCount]bool{}
		}
		storyX, storyY = 0, 0
		storyNarrowY = 90
		storyLeft[1], storyRight[1] = 0, 166
		storySpace = 4
		readFile("QUESTION.DAT", scratch[:demoScriptSize], questionEnd)
		selectPage(1)
		printChar(charClearWindow)
		putPicture(pictures, 10, 168, 100, 0)
		printStory(font, scratch[:demoScriptSize])
		storySpace = 5
		copyPage(1, 0)
		for pollKey() == 0 {
		}
		selectPage(1)
		printChar(charClearWindow)
		selectPage(0)
		hero.intel, hero.magic = newInt, newInt
		hero.dex = newDex
		hero.str = max(newStr, minStrength)
		ool := make([]byte, 2*oolSize)
		readFile("INIT.OOL", ool[oolSize:], 0)
		printChar(charClearWindow)
		setTextPosition(0, 10)
		writeFile("SAVED.OOL", ool)
		writeFile("SAVED.GAM", encodeSave())
	}
	mapID = introMap
}
