package game

// The inn: a night's rest, and leaving companions to wait or picking them up (shoppes3.c).

// Addresses of the inn tables.
const (
	roomPricesAddress = 19838 // C: D_4d7e (by inn)
	roomsAddress      = 19908 // C: D_4dc4 (the guests each inn takes)
	roomWordsAddress  = 20078 // C: D_4e6e (SHOPPE.DAT offsets)
	bedXAddress       = 20090 // C: D_4e7a (where the party sleeps)
	bedYAddress       = 20096 // C: D_4e80
	innCount          = 8
	innsWithRooms     = 6
)

var (
	roomPrices []byte   // C: D_4d7e
	rooms      []byte   // C: D_4dc4
	roomWords  []uint16 // C: D_4e6e
	bedX       []byte   // C: D_4e7a
	bedY       []byte   // C: D_4e80
)

// loadInnTables reads the inns' tables.
func loadInnTables() {
	roomPrices = segment.Bytes(roomPricesAddress, innCount)
	rooms = segment.Bytes(roomsAddress, innsWithRooms)
	roomWords = segment.Words(roomWordsAddress, innsWithRooms)
	bedX = segment.Bytes(bedXAddress, innsWithRooms)
	bedY = segment.Bytes(bedYAddress, innsWithRooms)
}

// SHOPPE.DAT offsets of the innkeeper's lines, and the inn's figures.
const (
	travellingAloneAt = 9960
	deadGuestAt       = 10019
	unpaidAt          = 10074
	monthlyRate       = 10
	innWakeHour       = 6
	lampsOutHour      = 5
	lampsOnHour       = 20
	nightSteps        = 12
	nightStepMinutes  = 5
	dozeMinutes       = 9
	registerLeft      = 198
	registerRight     = 305
	registerTop       = 40
	registerRow       = 8
	registerRows      = 8
	registerName      = 4
)

// innPrice lowers a price by the intelligence of the one paying.
func innPrice(price, intel int) int { return price + price*(100-intel*3)/100 }

// guestsHere counts the characters waiting in this town.
// C: SHOPPES3_0000
func guestsHere() int {
	count := 0
	for row := range partySlots {
		if party[row].location == mapID {
			count++
		}
	}
	return count
}

// roomFree tells whether the inn has a room; if not the innkeeper says so.
// C: SHOPPES3_002c
func roomFree(gender int) bool {
	printMessage(msgNewlines2)
	if guestsHere() < int(rooms[shopIndex]) {
		return true
	}
	printMessage(msgIAmSorryLine)
	if gender == femaleGend {
		printMessage(msgMilady)
	} else {
		printMessage(msgSir)
	}
	printMessage(msgButWeHaveNoRoomAvailable)
	return false
}

// restoreMagic gives a member back its magic points after a rest.
func restoreMagic(m *member) {
	switch m.class {
	case classAvatar, classMage:
		m.magic = m.intel
	case classBard:
		m.magic = m.intel >> 1
	}
}

// restAtInn lets the party sleep until morning. It returns -2 when there was no room.
// C: SHOPPES3_0072
func restAtInn(intel, gender, result int) int {
	printChar('R')
	if !roomFree(gender) {
		return -2
	}
	printChar('"')
	shopPrice = innPrice(int(roomPrices[shopIndex])*int(partySize), intel)
	sayShopFile(int(roomWords[shopIndex]))
	printMessage(msgWiltThouTakeItSp)
	answer := askYesNo()
	printMessage(msgNewlines2)
	if answer != 'Y' {
		return result
	}
	if int(gold) < shopPrice {
		printMessage(msgHighwaymenCheapAtThatOutSp)
		sayShopMessage(msgScreamsLine)
		return result
	}
	gold -= int16(shopPrice)
	shadowlordTax()
	showStatus()
	printMessage(msgHaveAPleasantNightSp)
	if gender == femaleGend {
		printMessage(msgMilady)
	} else {
		printMessage(msgSir)
	}
	sayShopMessage(msgSaysLine)
	sleepFrames(10)
	setTile(doorOpen, doorX, doorY)
	mapX, mapY = bedX[shopIndex], bedY[shopIndex]
	viewChanged = 1
	active := activeMember
	for row := range int(partySize) {
		if party[row].status == statusGood {
			party[row].status = statusSleeping
		}
	}
	showStatus()
	updateFrame()
	printMessage(msgZzzzzzLine3)
	for range nightSteps {
		passTime(nightStepMinutes)
	}
	for hour != innWakeHour {
		sleepFrames(1)
		regenerate()
		showStatus()
		passTime(dozeMinutes)
		if hour == lampsOnHour || hour == lampsOutHour {
			switchLamps()
		}
	}
	printMessage(msgMorningLine)
	placeAllPeople()
	for row := range int(partySize) {
		m := &party[row]
		if m.status == statusDead {
			continue
		}
		m.hp = m.maxHP
		restoreMagic(m)
		switch m.status {
		case statusPoisoned:
			m.status = statusDead
			m.hp = 0
			printChar('\n')
			printString(m.nameString())
			printMessage(msgHasPassedAwayLine)
		case statusSleeping:
			m.status = statusGood
		}
	}
	activeMember = active
	mapX++
	viewChanged = 1
	showStatus()
	return -1
}

// leaveCompanion leaves a member to wait at the inn.
// C: SHOPPES3_02ae
func leaveCompanion(intel, gender, result int) int {
	printChar('L')
	if !roomFree(gender) {
		return result
	}
	if partySize == 1 {
		sayShopFile(travellingAloneAt)
		return -1
	}
	for asked := false; !asked; {
		asked = true
		sayShopMessage(msgAsksWhoWillStaySp)
		who := selectOne()
		if who == -1 {
			printMessage(msgNobodyLine2)
			result = -2
			continue
		}
		printMessage(msgNewlines2)
		switch {
		case who == 0:
			printMessage(msgThyFriend)
			if partySize > 2 {
				printChar('s')
			}
			printMessage(msgWillNotLeaveTheeLine)
			asked = false
		case party[who].status == statusDead:
			sayShopFile(deadGuestAt)
			result = -1
		default:
			shopPrice = innPrice(int(roomPrices[shopIndex])*monthlyRate, intel)
			printMessage(msgTheRateForOurMostComfortable)
			sayShopMessage(msgGoldPerMonthDueAtCheck)
			printMessage(msgWiltThouTakeItSp)
			answer := askYesNo()
			printMessage(msgNewlines2)
			if answer != 'Y' {
				break
			}
			result = 1
			if int(activeMember) == who {
				activeMember = noMember
			} else if who < int(activeMember) && activeMember != noMember {
				activeMember--
			}
			party[who].location = mapID
			party[who].monthsInInn = 0
			guest := party[who]
			copy(party[who:partySlots-1], party[who+1:])
			party[partySlots-1] = guest
			partySize--
			showStatus()
			sayShopMessage(msgIThankTheeSaysLine)
		}
	}
	return result
}

// previousGuest returns the guest waiting here before row (0 none).
// C: SHOPPES3_0494
func previousGuest(row int) int {
	for row--; row != 0; row-- {
		if party[row].location == mapID {
			break
		}
	}
	return row
}

// nextGuest returns the guest waiting here after row (0 none).
// C: SHOPPES3_04b6
func nextGuest(row int) int {
	for row++; row < partySlots; row++ {
		if party[row].location == mapID {
			return row
		}
	}
	return 0
}

// pickUpCompanion has a waiting companion check out and join the party again.
// C: SHOPPES3_04e6
func pickUpCompanion(intel, result int) int {
	waiting := guestsHere()
	switch {
	case partySize == maxPartySize:
		printMessage(msgOneMustFirstBeLeftBehind)
		return result
	case waiting == 0:
		sayShopMessage(msgNoOneHereIsFromThy)
		return result
	}
	guest := nextGuest(0)
	if waiting > 1 {
		guest = chooseGuest()
	}
	if guest == 0 {
		return result
	}
	months := int(party[guest].monthsInInn)
	if months == 0 {
		months = 1
	}
	shopPrice = months * innPrice(int(roomPrices[shopIndex])*monthlyRate, intel)
	sayShopMessage(msgThatWillBeGoldPlease)
	if int(gold) < shopPrice {
		sayShopFile(unpaidAt)
		raiseAlarm()
		return -1
	}
	gold -= int16(shopPrice)
	shadowlordTax()
	showStatus()
	joining := party[guest]
	for row := guest; row > int(partySize); row-- {
		party[row] = party[row-1]
	}
	party[partySize] = joining
	m := &party[partySize]
	partySize++
	m.location = 0
	if m.status == statusPoisoned {
		m.status = statusDead
		m.hp = 0
		printMessage(msgThyFriendHasDiedByThe)
	} else {
		m.hp = m.maxHP
		restoreMagic(m)
		printMessage(msgIHopeThouHastFoundThy)
	}
	sayShopMessage(msgSaysLine2)
	showStatus()
	return 1
}

// chooseGuest shows the guest register in the panel and lets the player pick a guest
// (0 none).
func chooseGuest() int {
	printMessage(msgWhoWillCheckOutSp)
	selectWindow(panelWindow)
	clearIconRow()
	drawListFrame(registerRows)
	setTextPosition(1, 1)
	printMessage(msgGuest)
	setTextPosition(1, 2)
	printMessage(msgRegisterLine)
	for row := 1; row < partySlots; row++ {
		if party[row].location == mapID {
			setTextPosition(registerName, textY())
			printString(party[row].nameString())
			printChar('\n')
		}
	}
	selectWindow(2)
	guest := nextGuest(0)
	y := registerTop
	setPenColour(brightWhiteColour)
	for done := false; !done; {
		xorFillRect(registerLeft, y, registerRight, y+7)
		for moved := false; !moved; {
			moved = true
			switch getChar() {
			case keyLeft, keyUp:
				if p := previousGuest(guest); p != 0 {
					xorFillRect(registerLeft, y, registerRight, y+7)
					guest = p
					y -= registerRow
				}
			case keyRight, keyDown:
				if n := nextGuest(guest); n != 0 {
					xorFillRect(registerLeft, y, registerRight, y+7)
					guest = n
					y += registerRow
				}
			case keyEscape:
				printMessage(msgNoOneLine)
				guest = 0
				done = true
			case keySpace, keyEnter:
				done = true
			default:
				moved = false
			}
		}
	}
	clearIconBar()
	showStatus()
	return guest
}

// inn is the inn: rest, leave a companion, or pick one up.
// C: SHOPPES3_08b4
func inn(buyer int) {
	result := 0
	intel := int(party[buyer].intel)
	gender := int(party[buyer].gender)
	spellIcon, protection = 0, 0
	showStatus()
	shopGreeting()
	for done := false; !done; {
		c := 0
		for c != 'N' && c != 'Y' {
			c = getChar()
			if c == keySpace {
				c = 'N'
			}
			switch c {
			case 'N':
				printMessage(msgNo)
			case 'Y':
				printMessage(msgYes)
			}
		}
		if c == 'N' {
			break
		}
		sayShopMessage(msgAsksArtThouHereToPick)
		printMessage(msgLeaveACompanionOrToRest)
		for chosen := false; !chosen; {
			chosen = true
			switch getChar() {
			case 'R':
				if r := restAtInn(intel, gender, result); r > -2 {
					result = r
					done = true
				}
			case keySpace:
				done = true
			case 'L':
				if r := leaveCompanion(intel, gender, result); r > -2 {
					result = r
				}
				done = result == -1
			case 'P':
				result = pickUpCompanion(intel, result)
				done = result == -1
			default:
				chosen = false
			}
		}
		if !done {
			printMessage(msgIsThereAnythingMoreICan)
		}
	}
	shopFarewell(result)
}
