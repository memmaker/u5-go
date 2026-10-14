package game

// Segment 4000: time passing, the moons and moongates, the frame's captions, actor
// animation and the ambient sounds of the map.

// Runtime state of segment 4000.
var (
	// viewChanged asks for the view to be worked out again (1) or redrawn (2) (D_24e6).
	viewChanged byte = 1 // C: D_24e6
	// statusChanged asks for the status panel to be redrawn (D_a9fa).
	statusChanged bool // C: D_a9fa
	// ambientBeat counts the eighths of the ambient sounds' bar (D_6a34).
	ambientBeat byte // C: D_6a34
	// songNote is the next note of Iolo's song (D_6a08).
	songNote byte // C: D_6a08
	// mapName is the name shown above the map, padded to 12 (D_a3e2).
	mapName = "            " // C: D_a3e2
)

// Addresses of segment 4000's tables in the data segment.
const (
	actorAnimationsAddress = 6936  // C: D_1b18 (animation scripts, and their choice by tile)
	moonPhasesAddress      = 7898  // C: D_1eda (the moons' phase glyphs for each day)
	songFrequenciesAddress = 27190 // C: D_6a36
	songNotesAddress       = 27208 // C: D_6a48
	dawnLightAddress       = 27264 // C: D_6a80
)

const (
	animationScriptSize = 16
	animatedActorGroups = 51
	songFrequencyCount  = 9
	songLength          = 53
	dawnSteps           = 6
)

var (
	actorAnimations []byte // C: D_1b18
	moonPhases      []byte // C: D_1eda
	songFrequencies []int  // C: D_6a36
	songNotes       []byte // C: D_6a48
	dawnLight       []byte // C: D_6a80
)

// loadClockTables reads segment 4000's tables.
func loadClockTables() {
	actorAnimations = segment.Bytes(actorAnimationsAddress, animationChoiceStart+animatedActorGroups)
	moonPhases = segment.Bytes(moonPhasesAddress, daysPerMonth*2)
	songFrequencies = segment.SignedWords(songFrequenciesAddress, songFrequencyCount)
	songNotes = segment.Bytes(songNotesAddress, songLength)
	dawnLight = segment.Bytes(dawnLightAddress, dawnSteps)
}

// animationChoiceStart is where the animation script number of each tile group starts.
const animationChoiceStart = 176

const (
	daysPerMonth     = 28
	monthsPerYear    = 13
	hoursPerDay      = 24
	minutesPerHour   = 60
	ringSlot         = 4 // the equipment slot rings go in
	regenerationRing = 44
	innMonthsLimit   = 25
)

// regenerate heals each member wearing the ring of regeneration now and then.
// C: ULTIMA_400c
func regenerate() {
	for i := range int(partySize) {
		m := &party[i]
		if m.status != statusDead && m.equip[ringSlot] == regenerationRing && randomRange(0, 7) == 7 {
			increaseInt(&m.hp, 1, int(m.maxHP))
			statusChanged = true
		}
	}
}

// setActiveMember makes the member of the digit key active, or none for 0. It returns 0
// when it did, 1 when the member can't be.
// C: ULTIMA_4080
func setActiveMember(key int) int {
	result := 1
	row := key - '1'
	printMessage(msgSetActivePlrLine2)
	if key == '0' {
		printMessage(msgNoneLine)
		activeMember = noMember
		showStatus()
		return 0
	}
	if int(partySize) > row && party[row].status != statusDead && party[row].status != statusSleeping {
		activeMember = byte(row)
		printString(party[row].nameString())
		printChar('\n')
		showStatus()
		result = 0
	} else {
		printMessage(msgInvalidLine)
	}
	return result
}

// Ambient sounds.
const (
	ambientNone = iota
	ambientClock
	ambientWaterfall
	ambientFountain
	ambientLute
)

const (
	hearingRange = 51 // the squared distance sounds carry
	beatsPerBar  = 8
	chimeBeat    = 4
)

// playAmbientSound plays the nearest clock, waterfall, fountain or lute.
// C: ULTIMA_4102_AudioPlayAmbientSfx
func playAmbientSound() {
	sound, nearest := ambientNone, hearingRange
	cx, cy := viewCentre, viewCentre
	if mapID < firstMapCombat {
		cx, cy = int(mapX), int(mapY)
	}
	for x, column := cx-viewCentre, 0; x < cx+viewCentre+1; x, column = x+1, column+1 {
		for y, row := cy-viewCentre, 0; y < cy+viewCentre+1; y, row = y+1, row+1 {
			distance := (y-cy)*(y-cy) + (x-cx)*(x-cx)
			if distance >= nearest {
				continue
			}
			tile := *tileAt(x, y)
			switch {
			case tile&^1 == mapTileClock:
				sound = ambientClock
			case tile&^3 == mapTileWaterfall:
				sound = ambientWaterfall
			case tile&^3 == mapTileFountain:
				sound = ambientFountain
			case viewTile(column, row) == 0 && actorAtView(column, row)&^3 == actorBard:
				sound = ambientLute
			default:
				distance = nearest
			}
			nearest = distance
		}
	}
	onChimeBeat := ambientBeat == 0 || ambientBeat == chimeBeat
	switch sound {
	case ambientClock:
		if chimesLeft != 0 && onChimeBeat {
			pulse(3116, 1, 2000, 20000, -10)
		} else if ambientBeat == 0 {
			tone(3000, 3)
		} else if ambientBeat == chimeBeat {
			tone(2000, 3)
		}
	case ambientWaterfall:
		whiteNoise(20, 60, 10000)
	case ambientFountain:
		whiteNoise(10, 30, 25000)
	case ambientLute:
		if note := songNotes[songNote]; note != 0 {
			instrumentNote(songFrequencies[note-1], 2000, 20000, -10)
		}
		songNote++
		if songNote >= songLength {
			songNote = 0
		}
	}
	if chimesLeft != 0 && onChimeBeat {
		chimesLeft--
	}
	ambientBeat++
	if ambientBeat >= beatsPerBar {
		ambientBeat = 0
	}
}

// footstep plays a step.
// C: ULTIMA_433e_AudioFootstep
func footstep() {
	whiteNoise(1, 25, 1000)
	busyWait(20)
	whiteNoise(1, 25, 1500)
}

// victoryTune plays the fanfare after a won battle.
// C: ULTIMA_4368_AudioVictoryTune
func victoryTune() {
	for range 3 {
		pulse(4600, 1, 10800, 300, 6)
	}
	pulse(6100, 1, 21600, 300, 3)
}

// Tile animation ranges in tileShown.
const (
	waterfallFirst = mapTileWaterfall
	fountainFirst  = mapTileFountain
	animatedFour   = 4
	flagFirst      = 128
	fireFirst      = 236
	clockFirst     = mapTileClock
	clockLast      = clockFirst + 4
)

// stepTileAnimations moves the waterfalls and fountains on every frame, and the flags,
// fires and clocks every second or fourth.
// C: ULTIMA_44b8_AnimateTiles
func stepTileAnimations() {
	cycle := func(first int) {
		for i := first; i < first+animatedFour; i++ {
			tileShown[i]++
			if int(tileShown[i]) == first+animatedFour {
				tileShown[i] = byte(first)
			}
		}
	}
	cycle(waterfallFirst)
	cycle(fountainFirst)
	if animationStep&1 != 0 {
		for i := flagFirst; i < flagFirst+animatedFour; i++ {
			tileShown[i] ^= 1
		}
		cycle(fireFirst)
		if animationStep&2 != 0 {
			for i := clockFirst; i < clockLast; i++ {
				tileShown[i] ^= 1
			}
		}
	}
	animationStep++
}

// Steps of an animation script.
const (
	scriptRestart     = 0
	scriptFrameLast   = 4
	scriptRest        = 5
	scriptMaybeRepeat = 6
	scriptLoop        = 7
	scriptLoopTo      = 2
	scriptPauseBase   = 128
)

const (
	frameTimerMask = 15
	frozenTimer    = 15
	restTimer      = 6
	firstAnimated  = 52
	actorGargoyleA = 232
	actorGargoyleB = 180
	actorAlwaysA   = actorBard
	actorAlwaysB   = 168
)

// animateActors moves each actor on through its animation script, then the tiles.
// C: ULTIMA_4552_AnimateActors
func animateActors() {
	for i := range actorSlots {
		a := &actors[i]
		if a.tile == 0 {
			continue
		}
		timer := int(a.state & frameTimerMask)
		position := int(a.state >> 4)
		group := int(a.tile &^ 3)
		script := actorAnimations[int(actorAnimations[(group-firstAnimated)>>2+animationChoiceStart])*animationScriptSize:]
		if timer == frozenTimer {
			continue
		}
		if timer != 0 {
			timer--
			a.state = a.state&^frameTimerMask + byte(timer)
			continue
		}
		if a.animTile == 0 || a.animTile == actor29 || a.animTile == actorSleep {
			continue
		}
		if group < firstAnimated || group == actorGargoyleA || group == actorGargoyleB {
			continue
		}
		if group != actorAlwaysA && group != actorAlwaysB && randomRange(0, 255) < 128 {
			continue
		}
		for done := false; !done; {
			switch step := int(script[position]); {
			case step >= 1 && step <= scriptFrameLast:
				a.animTile = byte(group + step - 1)
				position++
				done = true
			case step == scriptRest:
				if randomRange(0, 255) < 64 {
					a.animTile = a.tile
					if group == actorAlwaysA {
						position++
					} else {
						timer = restTimer
					}
					done = true
					break
				}
				position++
			case step == scriptMaybeRepeat:
				if randomRange(0, 255) >= 192 {
					position++
				} else {
					position = 0
				}
			case step == scriptRestart:
				position = 0
			case step == scriptLoop:
				position = scriptLoopTo
			default:
				timer = step + scriptPauseBase
				position++
				done = true
			}
			a.state = byte(position*16 + timer)
		}
	}
	stepTileAnimations()
	screen.AnimateTiles()
}

// moonstoneHere says whether moonstone lies buried on the loaded map.
// C: ULTIMA_4702
func moonstoneHere(stone int) bool {
	if moonstoneState[stone] != mapID || moonstoneLevel[stone] != level {
		return false
	}
	if mapID != 0 {
		return true
	}
	dx := moonstoneX[stone] - chunkX
	dy := moonstoneY[stone] - chunkY
	return dx < settlementSize && dy < settlementSize
}

const (
	nightStarts    = 20
	nightEnds      = 5
	moongateHeight = 16
)

// raiseMoongates raises the moongates over the buried moonstones by night and lowers
// them by day.
// C: ULTIMA_475a
func raiseMoongates() {
	tile := byte(mapTileMoongate)
	if hour >= nightStarts || hour < nightEnds {
		increaseByte(&moongateRise, 1, moongateHeight)
	} else {
		decreaseByte(&moongateRise, 1)
		if moongateRise == 0 {
			tile = mapTileGrass
		}
	}
	for stone := range moonstones {
		if !moonstoneHere(stone) {
			continue
		}
		at := tileAt(int(moonstoneX[stone]), int(moonstoneY[stone]))
		changed := *at != tile
		*at = tile
		viewChanged |= 2
		if changed {
			updateVisibility()
		}
	}
}

// carried marks a moonstone the party carries.
const carried = 255

// travelToMoonstone sends the party to the moongate of moonstone.
// C: ULTIMA_47f4
func travelToMoonstone(stone int) int {
	if moonstoneState[stone] == carried {
		return 0
	}
	if mapID == 0 {
		writeFile(worldSaveFile(), actorBytes())
	}
	from := mapID
	mapID = moonstoneState[stone]
	mapX = moonstoneX[stone]
	mapY = moonstoneY[stone]
	level = moonstoneLevel[stone]
	if mapID != 0 && from != 0 && mapID < firstDungeonMap && from < firstDungeonMap {
		enterSettlement(true)
	} else if mapID == 0 && from == 0 {
		readActorBytes(worldSaveFile())
		redrawWorld()
	}
	return 1
}

// The moongate trip's sound and timing.
const (
	gateHum         = 5900
	gateHumLength   = 30000
	firstMoonPhase  = '0'
	noonHour        = 12
	midnightMinutes = 10
)

// enterMoongate takes the party through the moongate it stands on, to the moonstone of
// the moon phase. It returns 1 when the party stepped in in the first minutes after
// midnight, which the caller handles.
// C: ULTIMA_48a8
func enterMoongate() int {
	result := 0
	if *tileAt(int(mapX), int(mapY)) != mapTileMoongate {
		return 0
	}
	sleepFrames(1)
	pulse(gateHum, 1, gateHumLength, 2000, 2)
	tile := partyTile
	partyTile = actorCircle
	revealTile(mapTileMoongate, viewCentre, viewCentre)
	partyTile = 0
	sleepFrames(1)
	for moongateRise = moongateHeight - 1; moongateRise != 0; moongateRise-- {
		putMoongate(int(moongateRise), viewCentre, viewCentre)
		waitTicks(2)
	}
	*tileAt(int(mapX), int(mapY)) = mapTileGrass
	viewChanged |= 2
	putTile(mapTileGrass, viewCentre, viewCentre)
	if hour == 0 && minute < midnightMinutes {
		result = 1
	} else {
		phase := felucca
		if hour < noonHour {
			phase = trammel
		}
		travelToMoonstone(int(phase) - firstMoonPhase)
	}
	partyTile = tile
	return result
}

// chooseMember picks the member who acts: the combatant whose turn it is, the active
// member, the only one able, or one the player selects. It returns -1 for none.
// C: ULTIMA_4988
func chooseMember() int {
	chosen := -1
	switch {
	case mapID > firstMapCombat:
		chosen = int(combatants[combatant].entity)
	case activeMember != noMember:
		chosen = int(activeMember)
	default:
		able := 0
		for i := 0; i != int(partySize); i++ {
			if party[i].status == statusGood || party[i].status == statusPoisoned {
				chosen = i
				able++
			}
		}
		if able > 1 {
			for done := false; !done; {
				printMessage(msgPlayerSp)
				chosen = selectOne()
				if chosen < 0 {
					done = true
				} else if party[chosen].status == statusGood || party[chosen].status == statusPoisoned {
					printString(party[chosen].nameString())
					if textX() != 0 {
						printChar('\n')
					}
					done = true
				} else {
					printMessage(msgDisabledLine)
				}
			}
		}
	}
	if chosen == selectCancelled {
		printMessage(msgNoneLine)
	} else if chosen == selectEveryone {
		printChar('\n')
	}
	return chosen
}

const (
	skyWidth    = 12
	sunGlyph    = '*'
	sunrise     = 17 // the hour the sun starts at the sky's left
	trammelRise = 8
	feluccaRise = 2
	skyColumn   = 6
	skyFont     = 1
)

// showSky draws the sun and the two moons across the top of the map.
// C: ULTIMA_4a84
func showSky() {
	if mapID >= firstDungeonMap {
		return
	}
	var sky [skyWidth]byte
	for i := range sky {
		sky[i] = ' '
	}
	previous := windowIndex
	selectWindow(windowFrame)
	selectFont(skyFont)
	place := func(at int, glyph byte) {
		if at < skyWidth && at >= 0 {
			sky[at] = glyph
		}
	}
	place(sunrise-int(hour), sunGlyph)
	trammel = moonPhases[(int(day)-1)*2]
	at := trammelRise - int(hour)
	if at < -skyWidth {
		at += hoursPerDay
	}
	place(at, trammel)
	felucca = moonPhases[(int(day)-1)*2+1]
	at = feluccaRise - int(hour)
	if at < -skyWidth {
		at += hoursPerDay
	}
	place(at, felucca)
	setTextPosition(skyColumn, 0)
	if mapID != mapUnderworldCave && level < underworldLevel {
		for _, glyph := range sky {
			if glyph == sunGlyph {
				setTextColour(brightYellow)
			} else {
				setTextColour(lightGrayColour)
			}
			printChar(int(glyph))
		}
		setTextColour(brightWhiteColour)
	} else {
		setPenColour(blueColour)
		fillRect(40, 0, 152, 6)
		setPenColour(brightWhiteColour)
		line(40, 7, 152, 7)
	}
	selectWindow(previous)
	selectFont(0)
}

// showMapName prints the map's name above the map.
// C: ULTIMA_4be8
func showMapName() {
	previous := windowIndex
	selectWindow(windowFrame)
	if mapID != mapUnderworldCave && level < underworldLevel {
		setTextPosition(5, 0)
		printOpenCaption()
		printString(mapName)
		printCloseCaption()
	}
	selectWindow(previous)
}

const (
	glyphOpenCaption  = 2
	glyphCloseCaption = 1
)

// printOpenCaption prints the frame's left caption end, a blue cell with a white '>'.
// C: ULTIMA_4c2a
func printOpenCaption() {
	x, y := (window.left+textX())*cellSize, (window.top+textY())*cellSize
	colours := window.colours
	setTextColour(blueColour)
	setTextBackground(0)
	printChar(glyphOpenCaption)
	setPenColour(brightWhiteColour)
	line(x, y, x+5, y+3)
	line(x+5, y+4, x, y+7)
	setTextColour(int(colours & 15))
	setTextBackground(int(colours >> 4))
}

// printCloseCaption prints the frame's right caption end, a white '<'.
// C: ULTIMA_4cce
func printCloseCaption() {
	x, y := (window.left+textX())*cellSize, (window.top+textY())*cellSize
	colours := window.colours
	setTextColour(blueColour)
	setTextBackground(0)
	printChar(glyphCloseCaption)
	setPenColour(brightWhiteColour)
	line(x+7, y, x+2, y+3)
	line(x+2, y+4, x+7, y+7)
	setTextColour(int(colours & 15))
	setTextBackground(int(colours >> 4))
}

// indexOf returns the position of c in s, or its length.
// C: ULTIMA_4d76_FindIndex
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return len(s)
}

// clearMiddleBar clears the bar between the status panel's parts.
// C: ULTIMA_4daa
func clearMiddleBar() {
	setPenColour(blueColour)
	fillRect(240, 81, 263, 86)
	setPenColour(brightWhiteColour)
	horizLine(240, 80, 263)
	horizLine(240, 87, 263)
}

// showMiddleCaption shows c in the bar between the status panel's parts.
// C: ULTIMA_4dea
func showMiddleCaption(c int) {
	previous := windowIndex
	selectWindow(windowFrame)
	setTextPosition(30, 10)
	printOpenCaption()
	printChar(c)
	printCloseCaption()
	selectWindow(previous)
}

// clearCaption clears the caption above the status panel.
// C: ULTIMA_4e20
func clearCaption() {
	setPenColour(blueColour)
	fillRect(192, 0, 311, 6)
	setPenColour(brightWhiteColour)
	horizLine(192, 7, 311)
}

const captionCentre = 30

// showCaption shows a caption centred above the status panel.
// C: ULTIMA_4e50
func showCaption(address int) { showCaptionText(message(address)) }

// showCaptionText shows the caption text.
func showCaptionText(text string) {
	start := captionCentre - len(text)/2
	end := len(text) + start + 2
	setPenColour(blueColour)
	fillRect(192, 0, start*cellSize, 6)
	fillRect(end*cellSize, 0, 311, 6)
	setPenColour(brightWhiteColour)
	horizLine(192, 7, start*cellSize)
	horizLine(end*cellSize, 7, 311)
	previous := windowIndex
	selectWindow(windowFrame)
	setTextPosition(start, 0)
	printOpenCaption()
	printString(text)
	printCloseCaption()
	selectWindow(previous)
}

// clearIconRow clears the status panel's spell row to black.
// C: ULTIMA_4efc
func clearIconRow() {
	setPenColour(brightWhiteColour)
	vertLine(191, 56, 63)
	vertLine(312, 56, 63)
	setPenColour(0)
	fillRect(192, 56, 311, 63)
}

// clearIconBar draws the status panel's spell row as an empty bar.
// C: ULTIMA_4f3c
func clearIconBar() {
	setPenColour(blueColour)
	fillRect(191, 57, 312, 62)
	setPenColour(brightWhiteColour)
	horizLine(192, 56, 311)
	horizLine(192, 63, 311)
}

// Spell letters shown in the status panel that change time.
const (
	iconQuickness = 'Q'
	iconTimeStop  = 'T'
)

const (
	shrineQuestCount = 3
	shrineVisitDone  = 128
	virtueCount      = 8
	fullLight        = 51
	dayLight         = 50
	darkLight        = 2
	magicLight       = 18
	torchLightLevel  = 10
	duskHour         = 19
	minutesPerDawn   = 10
)

// passTime moves the clock on by minutes (halved under the quickness spell, stopped
// under time stop), with what happens each hour, day and month, then works out the
// light.
// C: ULTIMA_4f7c
func passTime(minutes int) {
	if minutes != 0 {
		if spellIcon == iconQuickness {
			minutes >>= 1
			if minutes == 0 {
				minutes++
			}
		}
		hourShown = hour
		if spellIcon != iconTimeStop {
			minute += byte(minutes)
			decreaseByte(&torchLight, minutes)
			decreaseByte(&lightSpell, minutes)
		}
		if minute >= minutesPerHour {
			minute -= minutesPerHour
			decreaseByte(&restCooldown, 1)
			hour++
			if hour >= hoursPerDay {
				hour = 0
				newDay()
				showStatus()
			}
		}
	}
	if light < fullLight {
		was := light
		switch {
		case mapID == mapUnderworldCave || level >= underworldLevel || hour < nightEnds || hour >= nightStarts:
			light = darkLight
		case hour == nightEnds:
			light = dawnLight[minute/minutesPerDawn]
		case hour == duskHour:
			light = dawnLight[(minutesPerHour-1-int(minute))/minutesPerDawn]
		default:
			light = dayLight
		}
		if lightSpell != 0 && light < magicLight {
			light = magicLight
		}
		if torchLight != 0 && light < torchLightLevel {
			light = torchLightLevel
		}
		if was != light {
			viewChanged = 1
		}
	}
	if hour != hourShown {
		if mapID < firstDungeonMap && level < underworldLevel {
			showSky()
		}
		switch {
		case hour == 0:
			chimesLeft = noonHour
		case hour > noonHour:
			chimesLeft = hour - noonHour
		default:
			chimesLeft = hour
		}
	}
	if mapID != 0 && mapID < firstDungeonMap {
		updateTimeTiles()
	}
}

// newDay moves the shrine quests' destinations, and on to a new month and year.
func newDay() {
	for quest := range shrineQuestCount {
		if shrineVisits[quest] >= shrineVisitDone {
			continue
		}
		var place int
		for {
			place = randomRange(1, virtueCount)
			if int(mapID) == place {
				place = 0
			}
			for other := range shrineQuestCount {
				if int(shrineVisits[other]) == place {
					place = 0
				}
			}
			if place != 0 {
				break
			}
		}
		shrineVisits[quest] = byte(place)
	}
	day++
	if day > daysPerMonth {
		harvestDay = [3]byte{}
		skullKeyDay = 0
		day = 1
		ambushed = 0
		for i := range partySlots {
			if party[i].monthsInInn < innMonthsLimit {
				party[i].monthsInInn++
			}
		}
		month++
		if month > monthsPerYear {
			month = 1
			year++
		}
	}
}
