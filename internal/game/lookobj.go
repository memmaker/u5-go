package game

// Looking at things (lookobj.c): descriptions from LOOK2.DAT, wells, fountains, the sky,
// signs from SIGNS.DAT, and the crystal gem's map.

// Addresses of the overlay's tables in the data segment.
const (
	gemPatternsAddress = 7450  // C: D_1d1a (how each tile shows in a gem)
	planetStartAddress = 14160 // C: D_3750
	planetRowAddress   = 14168 // C: D_3758
	planetPathAddress  = 14176 // C: D_3760
	signArtAddress     = 14352 // C: D_3810
	gemFlowersAddress  = 14370 // C: D_3822
	gemWallsAddress    = 14386 // C: D_3832
)

const (
	planets        = 8
	planetPathLong = 22
	signArtLines   = 9
	gemFlowerKinds = 16
	gemWallKinds   = 8
)

// The overlay's tables.
var (
	gemPatterns []byte   // C: D_1d1a
	planetStart []byte   // C: D_3750
	planetRow   []byte   // C: D_3758
	planetPath  []byte   // C: D_3760
	signArt     []string // C: D_3810
	gemFlowers  []byte   // C: D_3822
	gemWalls    []byte   // C: D_3832
)

// loadLookTables reads the overlay's tables.
func loadLookTables() {
	gemPatterns = segment.Bytes(gemPatternsAddress, mapTileCount)
	planetStart = segment.Bytes(planetStartAddress, planets)
	planetRow = segment.Bytes(planetRowAddress, planets)
	planetPath = segment.Bytes(planetPathAddress, planets*planetPathLong)
	signArt = messages(signArtAddress, signArtLines)
	gemFlowers = segment.Bytes(gemFlowersAddress, gemFlowerKinds)
	gemWalls = segment.Bytes(gemWallsAddress, gemWallKinds)
}

const (
	lookFile       = "LOOK2.DAT"
	signsFile      = "SIGNS.DAT"
	lookTextLength = 128
	lookActorsBase = 512
)

// describeTile prints LOOK2.DAT's description of a map tile.
// C: LOOKOBJ_0000
func describeTile(tile int) {
	printLookText(tile * 2)
}

// printLookText prints the description whose offset is at index in LOOK2.DAT.
func printLookText(index int) {
	var offset [2]byte
	readFile(lookFile, offset[:], index)
	var text [lookTextLength]byte
	readFile(lookFile, text[:], int(offset[0])|int(offset[1])<<8)
	printString(cString(text[:]))
}

const (
	wishLength   = 12
	wishTownOne  = 22
	wishTownTwo  = 31
	wishRate     = 10
	wishDuration = 3000
	wishLimit    = 2000
)

// lookAtWell offers to drop a coin into a well and make a wish.
// C: LOOKOBJ_0042
func lookAtWell(x, y, z byte) {
	printMessage(msgAWellDropACoin)
	key := 0
	for key != 'Y' && key != 'N' {
		key = getChar()
	}
	if key == 'N' {
		printMessage(msgNoLine)
		return
	}
	printMessage(msgYesLine)
	if gold == 0 {
		return
	}
	printMessage(msgThyWishLine)
	gold--
	wish := []byte(readString(wishLength) + "\x00")
	if wish[0] == 0 {
		printMessage(msgNothingLine)
		return
	}
	wanted := false
	for _, car := range []int{msgCorvette, msgFerrari, msgLamborghini, msgLotus, msgPorsche, msgHorse} {
		if findWord([]byte(message(car)+"\x00"), wish) > -1 {
			wanted = true
			break
		}
	}
	if !wanted || (mapID != wishTownOne && mapID != wishTownTwo) {
		printMessage(msgNoEffectLine5)
		return
	}
	printMessage(msgPoofLine3)
	whiteNoise(wishRate, wishDuration, wishLimit)
	slot := freeActorSlot()
	setActor(actorHorse, actorHorse, x+1, y, z, 0, slot)
}

// lookAtFountain lets a member drink from a fountain.
// C: LOOKOBJ_0162
func lookAtFountain() {
	printMessage(msgAGurglingFountainLine)
	printMessage(msgWhoWillDrinkLine)
	who := selectOne()
	if who == -1 {
		printMessage(msgNoneLine)
		return
	}
	status := party[who].status
	if status == statusDead || status == statusSleeping {
		printMessage(msgIncapacitatedLine)
	} else {
		printMessage(msgRefreshingLine)
	}
}

const (
	skyRight  = 176
	skyMargin = 173
)

// drawPlanet draws a planet in the night sky at sky cell x, y.
// C: LOOKOBJ_01ac
func drawPlanet(x, y int) {
	setPenColour(brightWhiteColour)
	plot(x*8+6, y*8+8)
	for row := range 3 {
		for column := range 3 {
			if x*8+column <= skyRight {
				plot(x*8+column+7, y*8+row+7)
			}
		}
	}
	if x*8 <= skyMargin {
		plot(x*8+10, y*8+8)
	}
}

// drawFlame marks with a flame a planet whose shrine the party has visited.
// C: LOOKOBJ_024c
func drawFlame(x, y int) {
	px, py := x*8, y*8
	setPenColour(redColour)
	if px > 2 {
		vertLine(px+5, py+10, py+12)
		vertLine(px+6, py+10, py+12)
		vertLine(px+7, py+8, py+12)
	}
	if px <= skyRight-1 {
		vertLine(px+8, py+8, py+12)
	}
	if px <= skyRight-2 {
		vertLine(px+9, py+6, py+10)
	}
	if px <= skyRight-3 {
		vertLine(px+10, py+6, py+10)
	}
	if px <= skyRight-4 {
		vertLine(px+11, py+5, py+8)
	}
	if px <= skyRight-5 {
		vertLine(px+12, py+5, py+7)
	}
}

const (
	dayStart      = 6
	dayEnd        = 18
	skyTile       = mapTile89
	skyStars      = 80
	starLeft      = 9
	starRight     = 182
	starBottom    = 172
	skyEpochYear  = 39
	skyEpochMonth = 4
	skyEpochDay   = 5
	centuries     = 100
	brightOffset  = 8
)

// lookAtSky shows the sun, or the night sky with its planets.
// C: LOOKOBJ_0366
func lookAtSky() {
	if hour >= dayStart && hour < dayEnd {
		printMessage(msgTheSunLine)
		if activeMember == noMember && firstAbleMember() == someoneAble {
			activeMember = byte(moveX)
		}
		hurtMember(int(activeMember), 1)
		showStatus()
		return
	}
	viewChanged = 1
	for y := range viewSize {
		for x := range viewSize {
			viewTiles[y*viewStride+x] = hiddenTile
		}
	}
	viewTiles[(viewSize-1)*viewStride+viewCentre] = skyTile
	drawView()
	setPenColour(blueColour + brightOffset)
	for range skyStars {
		plot(randomRange(starLeft, starRight), randomRange(starLeft, starBottom))
	}
	for planet := range planets {
		yr, mo, dy := int(year)%centuries, int(month), int(day)
		at := int(planetStart[planet])
		for yr > skyEpochYear || (yr == skyEpochYear && mo > skyEpochMonth) ||
			(yr == skyEpochYear && mo == skyEpochMonth && dy > skyEpochDay) {
			dy--
			if dy == 0 {
				dy = daysPerMonth
				mo--
			}
			if mo == 0 {
				mo = monthsPerYear
				yr--
			}
			for {
				at--
				if at < 0 {
					at = planetPathLong - 1
				}
				if planetPath[planet*planetPathLong+at] != 0 {
					break
				}
			}
		}
		x, y := at+1, int(planetRow[planet])
		drawPlanet(x, y)
		x--
		for i := range shadowlordCount {
			if int(shrineVisits[i]) == planet+1 {
				drawFlame(x, y)
			}
		}
	}
	printMessage(msgTheNightSkySp)
	for pollKeyWithCursor() == 0 {
	}
}

const (
	stairsUp      = mapTile224
	stairsRight   = mapTile225
	stairsLeft    = mapTile226
	flameTile     = mapTile222
	dungeonSign   = mapTile223
	clockHours    = 12
	keepOfTruth   = 30
	keepOfLove    = 31
	keepOfCourage = 32
)

// The doors of the dungeons by their x.
const (
	deceitX   = 240
	despiseX  = 91
	destardX  = 72
	wrongX    = 126
	covetousX = 156
	shameX    = 58
	hythlothX = 239
	doomX     = 128
)

// lookAtTile describes the map tile at x, y.
// C: LOOKOBJ_0502
func lookAtTile(tile, x, y int) {
	for tile == stairsUp || tile == stairsRight || tile == stairsLeft {
		switch tile {
		case stairsUp:
			y--
		case stairsRight:
			x++
		case stairsLeft:
			x--
		}
		tile = int(*tileAt(x, y))
	}
	switch {
	case tile == skyTile:
		lookAtSky()
		return
	case tile == mapTileWell:
		lookAtWell(mapX, mapY, level)
		return
	case tile&shapeMask == mapTileFountain:
		lookAtFountain()
		return
	}
	describeTile(tile)
	switch {
	case tile&pairMask == mapTileClock:
		h := int(hour) % clockHours
		if h == 0 {
			h = clockHours
		}
		printNumber(h, 1, ' ')
		printChar(':')
		printNumber(int(minute), 2, '0')
		if hour >= clockHours {
			printMessage(msgPmLine)
		} else {
			printMessage(msgAmLine)
		}
	case tile == flameTile:
		switch mapID {
		case keepOfTruth:
			printMessage(msgTruthLine2)
		case keepOfLove:
			printMessage(msgLoveLine2)
		case keepOfCourage:
			printMessage(msgCourageLine2)
		}
	case tile == dungeonSign:
		switch x {
		case deceitX:
			printMessage(msgDeceitLine)
		case despiseX:
			printMessage(msgDespiseLine)
		case destardX:
			printMessage(msgDestardLine)
		case wrongX:
			printMessage(msgWrongLine)
		case covetousX:
			printMessage(msgCovetousLine)
		case shameX:
			printMessage(msgShameLine)
		case hythlothX:
			printMessage(msgHythlothLine)
		case doomX:
			printMessage(msgDoomLine)
		}
	default:
		if textX() != 0 {
			printChar('\n')
		}
	}
}

// describeActor prints LOOK2.DAT's description of an actor tile.
// C: LOOKOBJ_06a4
func describeActor(tile int) {
	printLookText(tile*2 + lookActorsBase)
	if textX() != 0 {
		printChar('\n')
	}
}

// Characters of a sign's text.
const (
	signNoSign     = -1
	signHeaderLen  = 4
	signSkip       = 6
	signLineA      = 38
	signLineB      = 39
	signBar        = 'l'
	signArtFirst   = 41
	signArtLast    = 49
	signWait       = 13
	signNormalFont = 128
	signCharMask   = 127
)

// printSign prints the sign whose record starts at offset in scratch (-1: the
// sign of the eight laws).
// C: LOOKOBJ_06f8
func printSign(at int) {
	if at == signNoSign {
		selectFont(1)
		printMessage(msgAbbbbbbbbbbbbbbcGGGLiveBy)
		printMessage(msgGEightLawsGGG)
		selectFont(0)
		return
	}
	for at += signHeaderLen; scratch[at] == '\n'; at += signSkip {
	}
	for {
		c := scratch[at]
		if c&signNormalFont != 0 {
			selectFont(0)
		} else {
			selectFont(1)
		}
		switch {
		case c == signLineA || c == signLineB:
			printChar(signBar)
		case c >= signArtFirst && c <= signArtLast:
			selectFont(1)
			for _, ch := range []byte(signArt[c-signArtFirst]) {
				printChar(int(ch))
			}
		case c == signWait:
			selectFont(0)
			getChar()
		default:
			printChar(int(c & signCharMask))
		}
		at++
		if scratch[at] == 0 {
			break
		}
	}
	selectFont(0)
	if textX() != 0 {
		printChar('\n')
	}
}

const (
	wantedTown   = 4
	wantedX      = 17
	wantedY      = 21
	wantedRight  = 14
	wantedNames  = 3
	wantedCentre = 7
	signIndexLen = 66
	signNoData   = 255
)

// readSign reads the sign at x, y on level z.
// C: LOOKOBJ_07e4
func readSign(z, x, y byte) {
	if mapID == wantedTown && z == 0 && x == wantedX && y == wantedY {
		selectFont(1)
		printMessage(msgAbbbbbbbbbbbbbcGSp)
		selectFont(0)
		printMessage(msgWantedSp)
		selectFont(1)
		printMessage(msgGG)
		setTextPosition(wantedRight, textY())
		printMessage(msgGG)
		for i := range wantedNames {
			if i < int(partySize) {
				selectFont(0)
				name := party[i].nameString()
				setTextPosition(wantedCentre-len(name)/2, textY())
				printString(name)
				selectFont(1)
			}
			setTextPosition(wantedRight, textY())
			printMessage(msgGG)
		}
		setTextPosition(wantedRight, textY())
		printMessage(msgGG)
		selectFont(0)
		printMessage(msgDeadOrAlive)
		selectFont(1)
		printMessage(msgGDeeeeeeeeeeeeefLine)
		selectFont(0)
		return
	}
	var index [signIndexLen]byte
	readFile(signsFile, index[:], 0)
	offset := int(index[int(mapID)*2]) | int(index[int(mapID)*2+1])<<8
	for i := range scratch {
		scratch[i] = signNoData
	}
	if offset == 0 {
		printSign(signNoSign)
		return
	}
	readFile(signsFile, scratch[:], offset)
	for at := 0; ; {
		if scratch[at+1] == z && scratch[at+2] == x && scratch[at+3] == y {
			printSign(at)
			return
		}
		at += signHeaderLen
		// fix: the original scans on past the buffer when the file has no end mark
		for at < scratchSize-signHeaderLen && scratch[at] != 0 {
			at++
		}
		at++
		if at >= scratchSize-signHeaderLen || scratch[at] == signNoData {
			printSign(signNoSign)
			return
		}
	}
}

const (
	visionTile  = mapTile41
	visionSkill = 30
)

// lookCommand looks at what lies beside the party.
// C: LOOKOBJ_099c_LookCmd
func lookCommand() {
	if !selectDirection() {
		return
	}
	x, y := int(mapX)+int(moveX), int(mapY)+int(moveY)
	tile := int(*tileAt(x, y))
	someone := actorAt(x, y, int(level))
	if tile == visionTile {
		who := chooseMember()
		if who == -1 {
			return
		}
		if int(party[who].intel) <= randomRange(1, visionSkill) {
			printMessage(msgDeathVisionLine)
			hurtMember(who, 1)
			showStatus()
			return
		}
		printMessage(msgStrangeVisionLine)
		viewGem(int(mapX), int(mapY))
		return
	}
	printMessage(msgThouDostSeeLine)
	if someone != 0 {
		describeActor(someone)
		return
	}
	switch tile {
	case mapTile137, mapTile138, mapTile160, mapTile164, mapTileSign248:
		printMessage(msgNewline)
		readSign(level, byte(x), byte(y))
	default:
		lookAtTile(tile, x, y)
	}
}

const (
	gemOrigin = 32
	gemCell   = 4
	gemCells  = 32
)

// gemCellAt leaves the screen corner of gem cell x, y in moveX, moveY.
// C: LOOKOBJ_0a9c
func gemCellAt(x, y int) {
	moveX = int16(x*gemCell + gemOrigin)
	moveY = int16(y*gemCell + gemOrigin)
}

// gemPlot plots at the current gem cell's offset dx, dy.
func gemPlot(dx, dy int) { plot(int(moveX)+dx, int(moveY)+dy) }

// gemLine draws within the current gem cell.
func gemLine(x1, y1, x2, y2 int) {
	line(int(moveX)+x1, int(moveY)+y1, int(moveX)+x2, int(moveY)+y2)
}

// gemFill fills the current gem cell.
func gemFill(colour int) {
	setPenColour(colour)
	fillRect(int(moveX), int(moveY), int(moveX)+3, int(moveY)+3)
}

// gemTrees draws sparse greenery.
// C: LOOKOBJ_0abe
func gemTrees() {
	setPenColour(greenColour + brightOffset)
	gemPlot(1, 0)
	gemPlot(1, 2)
	gemPlot(3, 1)
	gemPlot(3, 3)
}

// gemForest fills the cell green.
// C: LOOKOBJ_0b04
func gemForest() { gemFill(greenColour + brightOffset) }

// gemColumns draws two white columns.
// C: LOOKOBJ_0b28
func gemColumns() {
	setPenColour(brightWhiteColour)
	gemLine(0, 0, 0, 3)
	gemLine(2, 0, 2, 3)
}

// gemRows draws two white rows.
// C: LOOKOBJ_0b60
func gemRows() {
	setPenColour(brightWhiteColour)
	gemLine(0, 0, 3, 0)
	gemLine(0, 3, 3, 3)
}

// gemBlock draws a small white block.
// C: LOOKOBJ_0b98
func gemBlock() {
	setPenColour(brightWhiteColour)
	gemLine(1, 1, 2, 1)
	gemLine(1, 2, 2, 2)
}

// gemFrame draws a white frame.
// C: LOOKOBJ_0bd0
func gemFrame() {
	setPenColour(brightWhiteColour)
	gemLine(0, 0, 3, 0)
	gemLine(0, 3, 3, 3)
	gemLine(0, 1, 0, 2)
	gemLine(3, 1, 3, 2)
}

// gemChecks draws yellow checks.
// C: LOOKOBJ_0c36
func gemChecks() {
	setPenColour(brightYellow)
	gemLine(0, 0, 1, 0)
	gemLine(0, 1, 1, 1)
	gemLine(2, 2, 3, 2)
	gemLine(2, 3, 3, 3)
}

// gemField draws furrows.
// C: LOOKOBJ_0c9c
func gemField() {
	setPenColour(greenColour + brightOffset)
	gemLine(0, 0, 3, 0)
	gemLine(0, 2, 3, 2)
	gemPlot(2, 1)
	gemPlot(0, 3)
}

// waterColour is the colour of water in the gem for the video driver.
func waterColour() int {
	if videoDriver == 0 || videoDriver == cgaComposite {
		return magentaColour
	}
	return blueColour + brightOffset
}

const (
	coastKind    = 96
	coastSides   = 15
	coastFirst   = 8
	cgaComposite = 3
)

// gemCoast draws a coast tile, green on its land sides.
// C: LOOKOBJ_0cf4
func gemCoast(tile int) {
	var colours [4]int
	bit := coastFirst
	for i := range colours {
		if tile&cellKindMask == coastKind && int(gemFlowers[tile&coastSides])&bit == 0 {
			colours[i] = greenColour + brightOffset
		} else {
			colours[i] = waterColour()
		}
		bit >>= 1
	}
	setPenColour(colours[0])
	gemPlot(1, 0)
	setPenColour(colours[1])
	gemPlot(3, 1)
	setPenColour(colours[2])
	gemPlot(1, 2)
	setPenColour(colours[3])
	gemPlot(3, 3)
}

// gemShallows draws shallow water.
// C: LOOKOBJ_0dda
func gemShallows() {
	setPenColour(waterColour())
	gemPlot(0, 0)
	gemPlot(2, 2)
}

// gemSwamp draws swamp.
// C: LOOKOBJ_0e16
func gemSwamp() {
	setPenColour(greenColour + brightOffset)
	gemPlot(1, 0)
	gemPlot(3, 1)
	setPenColour(waterColour())
	gemPlot(0, 2)
	gemPlot(2, 3)
}

const (
	gemTownFirst = 32
	gemTownNW    = 34
	gemTownNE    = 35
	gemTownSE    = 36
	gemTownSW    = 37
)

// gemTown draws a settlement with its roads.
// C: LOOKOBJ_0e7a
func gemTown(tile int) {
	gemTrees()
	setPenColour(redColour)
	fillRect(int(moveX)+1, int(moveY)+1, int(moveX)+2, int(moveY)+2)
	roads := gemWalls[tile-gemTownFirst]
	if roads&8 != 0 {
		gemLine(1, 0, 2, 0)
	}
	if roads&4 != 0 {
		gemLine(3, 1, 3, 2)
	}
	if roads&2 != 0 {
		gemLine(1, 3, 2, 3)
	}
	if roads&1 != 0 {
		gemLine(0, 1, 0, 2)
	}
	setPenColour(0)
	switch tile {
	case gemTownNW:
		gemPlot(1, 2)
	case gemTownNE:
		gemPlot(1, 1)
	case gemTownSE:
		gemPlot(2, 1)
	case gemTownSW:
		gemPlot(2, 2)
	}
}

// How tiles show in the gem.
const (
	gemNothing   = 0
	gemTreesKind = 1
	gemForestK   = 2
	gemRed       = 3
	gemRowsK     = 4
	gemBlockK    = 5
	gemFrameK    = 6
	gemWhite     = 7
	gemChecksK   = 8
	gemFieldK    = 9
	gemCoastK    = 10
	gemShallowsK = 11
	gemDot       = 12
	gemSwampK    = 13
	gemPillars   = 14
	gemWater     = 15
	gemTownK     = 16
	gemWhiteToo  = 90
)

// drawGemCell draws the tile in gem cell x, y.
// C: LOOKOBJ_0f7e
func drawGemCell(tile, x, y int) {
	gemCellAt(x, y)
	switch gemPatterns[tile] {
	case gemTreesKind:
		gemTrees()
	case gemForestK:
		gemForest()
	case gemRed:
		gemFill(redColour)
	case gemRowsK:
		gemRows()
	case gemBlockK:
		gemBlock()
	case gemFrameK:
		gemFrame()
	case gemWhite, gemWhiteToo:
		gemFill(brightWhiteColour)
	case gemChecksK:
		gemChecks()
	case gemFieldK:
		gemField()
	case gemCoastK:
		gemCoast(tile)
	case gemShallowsK:
		gemShallows()
	case gemDot:
		setPenColour(waterColour())
		gemPlot(2, 2)
	case gemSwampK:
		gemSwamp()
	case gemPillars:
		setPenColour(brightWhiteColour)
		gemLine(1, 0, 1, 3)
		gemLine(2, 0, 2, 3)
	case gemWater:
		gemFill(blueColour + brightOffset)
	case gemTownK:
		gemTown(tile)
	}
}

const (
	gemViewLeft  = 8
	gemViewRight = 183
	gemBlink     = 4
)

// viewGem shows the map around the party as a crystal gem does, the party at x, y
// blinking.
// C: LOOKOBJ_10fc_ViewCmd
func viewGem(x, y int) {
	x = (x - int(chunkX)) & lowByteMask
	y = (y - int(chunkY)) & lowByteMask
	selectPage(1)
	setPenColour(0)
	fillRect(gemViewLeft, gemViewLeft, gemViewRight, gemViewRight)
	for row := range gemCells {
		for column := range gemCells {
			drawGemCell(int(*tileAt(int(chunkX)+column, int(chunkY)+row)), column, row)
		}
	}
	copyArea(1, 0, gemViewLeft, gemViewLeft, gemViewRight, gemViewRight)
	selectPage(0)
	gemCellAt(x, y)
	for blink := 0; pollKeyWithCursor() == 0; {
		blink++
		if blink == gemBlink {
			blink = 0
			setPenColour(brightWhiteColour)
			xorFillRect(int(moveX), int(moveY), int(moveX)+3, int(moveY)+3)
		}
	}
	updateFrame()
}
