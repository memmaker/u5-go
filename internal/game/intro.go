package game

import (
	"github.com/memmaker/u5-go/internal/music"
	"time"

	"github.com/memmaker/u5-go/internal/ega"
	"github.com/memmaker/u5-go/internal/host"
)

// The INTRO overlay: the title sequence, the main menu, the introduction story and the
// acknowledgements.

// The maps the title screen runs with.
const (
	introMap   = 64 // the title screen without tiles
	journeyMap = 65 // the tiles loaded, the game about to start
)

// The files of the title sequence (D_25ea).
const (
	proportionalFont = "PROPORT.PCS"
	britishFile      = "BRITISH.BIT"
	titleFile        = "TITLE.BIT"
	tileFile         = "TILES.16"
	textPictures     = "TEXT.16"
	ultimaPictures   = "ULTIMA.16"
	creditPictures   = "STARTSC.16"
	bannerFile       = "WD.BIT"
	britishPathFile  = "BRITISH.PTH"
)

// storyPictures is the picture file of story part n (STORY1.16 to STORY6.16).
func storyPictures(n int) string { return "STORY" + string(rune('1'+n)) + ".16" }

// Addresses of the introduction's tables, one entry per page.
const (
	pageLeftAddress     = 12184 // C: D_2f98 (the left edges of the two text columns)
	pageRightAddress    = 12226 // C: D_2fc2 (the right edges)
	pageTextAddress     = 12310 // C: D_3016 (the offset of the page's text in STORY.DAT)
	pageNarrowYAddress  = 12352 // C: D_3040
	pageWideYAddress    = 12374 // C: D_3056
	pageXAddress        = 12396 // C: D_306c
	pageYAddress        = 12418 // C: D_3082
	pagePictureAddress  = 12440 // C: D_3098
	pageFileAddress     = 12462 // C: D_30ae
	pagePictureXAddress = 12484 // C: D_30c4
	pagePictureYAddress = 12506 // C: D_30da
	pageKindAddress     = 12528 // C: D_30f0
)

const storyPages = 21

// storyPage is the layout of one page of the introduction.
type storyPage struct {
	left, right              [2]int
	textAt                   int
	narrowY, wideY, x, y     int
	picture, file            int
	pictureX, pictureY, kind int
}

// The page kinds.
const (
	pageFramed = 1 // with Lord British's frame (pages 0, 7 and 14)
	pageReveal = 2 // the picture's second part dissolves in afterwards
	pageGate   = 3 // the moongate opens
	pageTwoUp  = 4 // kinds 4 and up add a second picture below the first
)

var storyPagesTable [storyPages]storyPage

// loadIntroTables reads the introduction's page layouts.
func loadIntroTables() {
	for i := range storyPagesTable {
		p := &storyPagesTable[i]
		p.left = [2]int{int(segment.Byte(pageLeftAddress + i*2)), int(segment.Byte(pageLeftAddress + i*2 + 1))}
		p.right = [2]int{int(segment.Word(pageRightAddress + i*4)), int(segment.Word(pageRightAddress + i*4 + 2))}
		p.textAt = int(segment.Word(pageTextAddress + i*2))
		p.narrowY = int(segment.Byte(pageNarrowYAddress + i))
		p.wideY = int(segment.Byte(pageWideYAddress + i))
		p.x = int(segment.Byte(pageXAddress + i))
		p.y = int(segment.Byte(pageYAddress + i))
		p.picture = int(segment.Byte(pagePictureAddress + i))
		p.file = int(segment.Byte(pageFileAddress + i))
		p.pictureX = int(segment.Byte(pagePictureXAddress + i))
		p.pictureY = int(segment.Byte(pagePictureYAddress + i))
		p.kind = int(segment.Byte(pageKindAddress + i))
	}
}

var (
	titlePictures ega.Pictures // C: D_bb1a
	pathAt        int          // C: D_bb18 (where the next Lord British stroke starts)
)

// Banner frames are drawn 50 pixels apart on the back buffer.
const (
	bannerFrames       = 4
	bannerSpacing      = 50
	bannerFrameX       = 16
	screenMiddleY      = 100
	menuTop            = 17
	menuOptions        = 6
	captionRow         = 24
	titleCaptionCentre = 18
	pathStepMask       = 7
	pathSignBit        = 8
	pathNegateX        = 128
	pathMaxStep        = 2
	pathTicksEvery     = 32
)

// drawBannerFrames puts the four frames of the "Warriors of Destiny" banner on the back
// buffer.
// C: INTRO_0010
func drawBannerFrames() {
	selectPage(1)
	printChar(charClearWindow)
	for i := range bannerFrames {
		putPicture(titlePictures, i+1, bannerFrameX, i*bannerSpacing, 0)
	}
}

// drawBritishStroke writes one stroke of Lord British's signature from (x, y): each
// path byte moves the pen up to two pixels, a longer move lifts it. It returns where the
// next stroke starts, 0 when a key cut it short.
// C: INTRO_0050_LordBritishAnim
func drawBritishStroke(x, y int) int {
	pen := brightWhiteColour
	setPenColour(pen)
	at := pathAt
	for ; at < len(scratch); at++ {
		if pen != 0 {
			plot(x, y)
		}
		pen = brightWhiteColour
		step := scratch[at]
		dy := int(step & pathStepMask)
		if dy > pathMaxStep {
			pen = 0
		}
		if step&pathSignBit != 0 {
			dy = -dy
		}
		dx := int(step>>4) & pathStepMask
		if dx > pathMaxStep {
			pen = 0
		}
		if step&pathNegateX != 0 {
			dx = -dx
		}
		x += dx
		y += dy
		if pollKey() != 0 {
			return 0
		}
		if at%pathTicksEvery == 0 {
			waitTicks(1)
		}
		if step == 0 {
			break
		}
	}
	pathAt = at + 1
	return pathAt
}

// introduction tells the story of the Avatar's summons, page by page.
// C: INTRO_014e_Introduction
func introduction() {
	if mapID != introMap {
		screen.UnloadTiles()
	}
	mapID = introMap
	font := loadPictures(proportionalFont)
	printStory(font, nil)
	selectPage(1)
	printChar(charClearWindow)
	selectPage(0)
	printChar(charClearWindow)
	setTextPosition(0, 10)
	frame := loadPictures(textPictures)
	pictures := loadPictures(storyPictures(0))
	file := 0
	for i, page := range storyPagesTable {
		playSong(songForStory(i))
		if file != page.file {
			file = page.file
			pictures = loadPictures(storyPictures(file))
		}
		selectPage(1)
		printChar(charClearWindow)
		if page.kind == pageFramed {
			switch i {
			case 0:
				putPicture(frame, 0, 224, 30, 0)
				putPicture(frame, 1, 168, 58, 0)
			case 7:
				putPicture(frame, 0, 232, 26, 0)
				putPicture(frame, 2, 200, 54, 0)
			case 14:
				putPicture(frame, 0, 184, 0, 0)
				putPicture(frame, 3, 248, 0, 0)
			}
		}
		putPicture(pictures, page.picture, page.pictureX, page.pictureY, 0)
		if page.kind >= pageTwoUp {
			putPicture(pictures, page.kind*2-5, page.pictureX, page.pictureY+55, 0)
		}
		storyLeft, storyRight = page.left, page.right
		storyNarrowY, storyWideY = page.narrowY, page.wideY
		storyX, storyY = page.x, page.y
		if page.kind == pageGate {
			putPicture(pictures, 3, 96, 39, 0)
			printStory(font, []byte(message(msgInstantlyAShimmeringBlueDoorSprings)))
			storyX, storyY = page.x, 180
			printStory(font, []byte(message(msgWithHeartBeatingRapidlyYouStep)))
		} else {
			text := scratch[:demoScriptSize]
			readFile("STORY.DAT", text, page.textAt)
			printStory(font, text)
		}
		if i != 0 {
			clearKeyboard()
			for pollKey() == 0 {
			}
		}
		copyPage(1, 0)
		if page.kind == pageReveal {
			selectPage(1)
			putPicture(pictures, 2, 40, 86, 0)
			reveal(40, 86, 75, 120)
		}
	}
	clearKeyboard()
	for pollKey() == 0 {
	}
	selectPage(1)
	printChar(charClearWindow)
	selectPage(0)
}

// showTitleCaption centres text in the bottom border.
// C: INTRO_043e
func showTitleCaption(text string) {
	left := -(len(text)/2 - titleCaptionCentre)
	right := left + len(text) + 2
	setPenColour(blueColour)
	fillRect(8, 193, left*cellSize, 199)
	fillRect(right*cellSize, 193, 311, 199)
	setPenColour(brightWhiteColour)
	line(8, 192, left*cellSize, 192)
	line(right*cellSize, 192, 311, 192)
	setTextPosition(left, captionRow)
	printOpenCaption()
	printString(text)
	printCloseCaption()
}

// The border glyphs of the character set.
const (
	glyphTopLeft     = 123
	glyphTopRight    = 124
	glyphBottomLeft  = 125
	glyphBottomRight = 126
	glyphBorder      = 127
	borderInside     = 38
)

// drawBoxRows draws a blue text border from row top with rows lines inside.
func drawBoxRows(top, rows int) {
	setTextColour(blueColour)
	setTextPosition(0, top)
	printChar(glyphTopLeft)
	for range borderInside {
		printChar(glyphBorder)
	}
	printChar(glyphTopRight)
	for range rows {
		printChar(glyphBorder)
		setTextPosition(textColumns-1, textY())
		printChar(glyphBorder)
	}
	printChar(glyphBottomLeft)
	for range borderInside {
		printChar(glyphBorder)
	}
}

// drawMenuBorders frames the main menu.
// C: INTRO_04e0_DrawMenuBorders
func drawMenuBorders() {
	drawBoxRows(15, 8)
	advanceCursor = false
	printChar(glyphBottomRight)
	advanceCursor = true
	setPenColour(brightWhiteColour)
	line(7, 127, 312, 127)
	lineTo(312, 192)
	lineTo(7, 192)
	lineTo(7, 127)
	setTextColour(brightWhiteColour)
}

// displayTitle shows the "Ultima" logo and the banner, dissolving them in when slow (a key
// makes it fast).
// C: INTRO_05b0_DisplayTitle
func displayTitle(slow bool) {
	if mapID != introMap {
		screen.UnloadTiles()
		mapID = introMap
	}
	titlePictures = loadPictures(ultimaPictures)
	selectPage(0)
	printChar(charClearWindow)
	selectPage(1)
	putPicture(titlePictures, 0, 0, 0, 0)
	if slow {
		reveal(0, 0, screenRight, screenMiddleY)
		slow = pollKey() == 0
	}
	if !slow {
		copyArea(1, 0, 0, 0, screenRight, screenMiddleY)
	}
	drawBannerFrames()
	if slow {
		screen.AnimateBanner(loadPictures(bannerFile))
	}
	selectPage(0)
	animateBanner()
	drawMenuBorders()
}

// writeMenuOption prints option row of the menu at column x, inverted when selected.
// C: INTRO_0676_WriteMenuOption
func writeMenuOption(selected, row, x, text int) {
	if row == selected {
		printChar(charToggleInverse)
	}
	setTextPosition(x, row+menuTop)
	printChar(' ')
	printMessage(text)
	printChar(' ')
	if row == selected {
		printChar(charToggleInverse)
	}
}

// buildMainMenu prints the menu with option selected highlighted.
// C: INTRO_06bc_BuildMainMenu
func buildMainMenu(selected int) {
	writeMenuOption(selected, 0, 12, msgJourneyOnward)
	writeMenuOption(selected, 1, 9, msgCreateNewCharacter)
	writeMenuOption(selected, 2, 8, msgTransferFromUltimaIv)
	writeMenuOption(selected, 3, 9, msgUltimaVIntroduction)
	writeMenuOption(selected, 4, 11, msgAcknowledgements)
	writeMenuOption(selected, 5, 10, msgReturnToTheView)
}

// waitMilliseconds paces the slides of the acknowledgements, which ran as fast as the
// machine could draw.
func waitMilliseconds(n int) { host.Wait(time.Duration(n) * time.Millisecond) }

// acknowledgements slides the credits in from below and the title out to the sides.
// C: INTRO_072e_Acknowledgements
func acknowledgements() {
	const top = 63
	credits := loadPictures(creditPictures)
	selectPage(1)
	putPicture(credits, 1, 16, top, 0)
	selectPage(0)
	for y := screenBottom; y >= top; y-- {
		putPicture(credits, 0, 144, y, 0)
		putPicture(credits, 2, 160, y, 0)
		// fix: the original drew as fast as the machine could (u5d's speed fix).
		waitMilliseconds((screenBottom - y) / 16)
	}
	for x := 0; x < 144; x += 8 {
		putPicture(credits, 0, 136-x, top, 0)
		copyArea(1, 0, 152-x, top, 159-x, screenBottom)
		putPicture(credits, 2, x+168, top, 0)
		copyArea(1, 0, x+160, top, x+167, screenBottom)
		waitTicks(1)
	}
	titlePictures = loadPictures(ultimaPictures)
	selectPage(1)
	printChar(charClearWindow)
	putPicture(titlePictures, 1, 16, 65, 0)
	drawMenuBorders()
	buildMainMenu(4)
	for pollKey() == 0 {
	}
	selectPage(0)
	for x := 136; x >= 0; x -= 8 {
		putPicture(credits, 0, 144-x, top, 0)
		copyArea(1, 0, 136-x, top, 143-x, screenBottom)
		putPicture(credits, 2, x+160, top, 0)
		copyArea(1, 0, x+176, top, x+183, screenBottom)
		waitTicks(1)
	}
	for y := top; y < screenBottom; y++ {
		putPicture(credits, 0, 144, y+1, 0)
		putPicture(credits, 2, 160, y+1, 0)
		copyArea(1, 0, 144, y, 175, y)
		waitMilliseconds((screenBottom - y) / 16)
	}
	copyArea(1, 0, 144, screenBottom, 175, screenBottom)
	drawBannerFrames()
	selectPage(0)
	clearKeyboard()
}

// pause waits up to ticks for a key and returns it, 0 for none.
// C: INTRO_094e_Pause
func pause(ticks int) int {
	for range ticks {
		if key := pollKey(); key != 0 {
			return key
		}
		waitTicks(1)
	}
	return 0
}

// animateBanner shows the next frame of the banner.
// C: INTRO_2090_AnimateWD
func animateBanner() {
	waitTicks(1)
	screen.ShowNextBannerFrame()
}

// menuIdleTicks is how long the menu waits before the demo plays.
const menuIdleTicks = 200

// titleSequence plays the Origin logo, Lord British's signature and the title.
func titleSequence() {
	slow := true
	logo := loadPictures(titleFile)
	signature := loadPictures(britishFile)
	readFile(britishPathFile, scratch[:], 0)
	if slow {
		slow = !animateOriginLogo(logo)
	}
	// "presents"
	selectPage(1)
	fillRect(0, 140, screenRight, screenBottom)
	putBitPicture(logo, 7, 108, 140)
	copyArea(1, 0, 0, 140, screenRight, screenBottom)
	if slow {
		waitTicks(18)
		slow = pause(20) == 0
	}
	if slow {
		printChar(charClearWindow)
		// "a"
		putBitPicture(logo, 8, 152, 0)
		copyPage(1, 0)
		slow = pause(20) == 0
		if slow {
			selectPage(0)
			pathAt = 0
			slow = false
			if drawBritishStroke(44, 68) != 0 && drawBritishStroke(64, 94) != 0 && drawBritishStroke(143, 78) != 0 {
				slow = drawBritishStroke(167, 105) != 0
			}
			selectPage(1)
		}
		// "Lord British"
		putBitPicture(signature, 0, 24, 66)
		copyPage(1, 0)
		// "production"
		putBitPicture(logo, 9, 104, 160)
		copyPage(1, 0)
		if slow {
			slow = pause(20) == 0
		}
	}
	mapID = introMap
	printChar(charClearWindow)
	selectPage(0)
	displayTitle(slow)
	if slow {
		returnToView()
	}
}

// animateOriginLogo builds the Origin logo on the back buffer and plays its animation;
// it reports whether a key cut it short.
// C: ULTIMA_0d72_AnimateOriginLogo
func animateOriginLogo(logo ega.Pictures) bool {
	y := 0
	selectPage(1)
	setPenColour(0)
	fillRect(0, 0, screenRight, screenBottom)
	for i := range originLines {
		putBitPicture(logo, i, (ega.Width-originWidths[i])/2, y)
		y += originHeights[i]
	}
	selectPage(0)
	return screen.AnimateOrigin()
}

// introMain runs the title screen and its menu until the player journeys onward.
// C: INTRO_0986_Main
func introMain() {
	leaveWorld()
	for i := range tileShown {
		tileShown[i] = byte(i)
	}
	loadFont("ibm.ch", 0)
	loadFont("runes.ch", 1)
	selectFont(0)
	setWindowSize(0, 0, 0, textColumns-1, textRows-1)
	selectWindow(0)
	playSong(music.UltimaTheme)
	choice := toUpper(pollKey())
	if choice == 'J' {
		printChar(charClearWindow)
		printChar(charCentre)
		setTextPosition(0, 11)
		printMessage(msgJourneyOnwardLine)
		printChar(charUncentre)
		for !loadTiles(tileFile) {
		}
		mapID = journeyMap
	} else {
		titleSequence()
		choice = 1
	}
	seedRandom(clockSeed())
	options := message(msgJctuar)
	for {
		playSong(music.UltimaTheme)
		setWindowSize(0, 1, 16, 38, 23)
		selectWindow(0)
		printChar(charClearWindow)
		setWindowSize(0, 0, 0, textColumns-1, textRows-1)
		setWindowSize(1, 24, 1, 39, 9)
		setWindowSize(2, 24, 11, 39, 23)
		var key int
		if choice == 'J' {
			key = 'J'
		} else {
			selected := 0
			showTitleCaption(message(msgCopyright1988LordBritish))
			buildMainMenu(0)
			for key <= ' ' {
				setTextPosition(15, 15)
				printOpenCaption()
				printMessage(msgSelectSp)
				printCloseCaption()
				setTextPosition(23, 15)
				key = 0
				for range menuIdleTicks {
					key = toUpper(pollKeyWithCursor())
					if key != 0 {
						break
					}
					animateBanner()
				}
				switch key {
				case keyLeft, keyUp:
					selected--
					if selected < 0 {
						selected = menuOptions - 1
					}
					buildMainMenu(selected)
				case keyRight, keyDown:
					selected++
					if selected >= menuOptions {
						selected = 0
					}
					buildMainMenu(selected)
				case keyEnter, keySpace:
					key = int(options[selected])
				case 0:
					key = 'R'
				default:
					if i := indexOf(options, byte(key)); i >= 0 {
						buildMainMenu(i)
					} else {
						key = 0
					}
				}
			}
		}
		switch key {
		case 'J':
			drawFrame()
			setWind(0)
			selectWindow(2)
			setTextPosition(0, 12)
			if mapID == introMap {
				for !loadTiles(tileFile) {
				}
				mapID = journeyMap
			}
			contents, err := disk.ReadFile("SAVED.GAM")
			if err == nil {
				err = loadSave(contents)
			}
			if err != nil || party[0].name[0] == 0 {
				printMessage(msgNoActiveGameSp)
				printMessage(msgPleaseCreateACharacterSp)
				printMessage(msgOrTransferOneFromUltimaIv)
				waitForKey(0)
				selectWindow(0)
				printChar(charClearWindow)
				selectPage(1)
				printChar(charClearWindow)
				selectPage(0)
				displayTitle(false)
				choice = 0
				continue
			}
			ool := make([]byte, 2*oolSize)
			readFile("SAVED.OOL", ool, 0)
			writeFile("BRIT.OOL", ool[:oolSize])
			writeFile("UNDER.OOL", ool[oolSize:])
			setWind(-1)
			video.tileYOffset = defaultTileOffset
			return
		case 'T', 'C':
			if key == 'T' {
				printChar(charClearWindow)
				transferFromUltima4()
			} else {
				createCharacter()
			}
			setWindowSize(0, 0, 0, textColumns-1, textRows-1)
			selectWindow(0)
			printChar(charClearWindow)
			selectPage(1)
			printChar(charClearWindow)
			displayTitle(false)
			drawBannerFrames()
			selectPage(0)
		case 'U':
			introduction()
			displayTitle(false)
		case 'A':
			if mapID != introMap {
				screen.UnloadTiles()
				mapID = introMap
			}
			acknowledgements()
		case 'R':
			returnToView()
		}
	}
}
