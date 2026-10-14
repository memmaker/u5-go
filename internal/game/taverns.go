package game

// The tavern and the shipwright (shoppes2.c).

// Addresses of the tavern and shipwright tables.
const (
	mealKeysAddress      = 19486 // C: D_4c1e (the key of each tavern menu's meal)
	drinkKeysAddress     = 19492 // C: D_4c24
	foodKeysAddress      = 19498 // C: D_4c2a
	loreKeysAddress      = 19504 // C: D_4c30
	mealPricesAddress    = 19510 // C: D_4c36 (by tavern)
	winePricesAddress    = 19528 // C: D_4c48
	rationPricesAddress  = 19540 // C: D_4c54 (by tavern)
	rationWordsAddress   = 19558 // C: D_4c66 (SHOPPE.DAT offsets)
	loreTopicsAddress    = 19572 // C: D_4c74 (pointers)
	loreThingsAddress    = 19624 // C: D_4ca8 (pointers)
	lorePlaceOfAddress   = 19676 // C: D_4cdc
	lorePlacesAddress    = 19702 // C: D_4cf6 (pointers)
	lorePricesAddress    = 19728 // C: D_4d10
	loreWordsAddress     = 19780 // C: D_4d44 (SHOPPE.DAT offsets)
	tavernMenusAddress   = 19788 // C: D_4d4c (the menu of each tavern)
	menuWordsAddress     = 19798 // C: D_4d56 (SHOPPE.DAT offsets by menu)
	moreWordsAddress     = 19806 // C: D_4d5e (SHOPPE.DAT offsets by menu)
	frigatePricesAddress = 19814 // C: D_4d66 (by shipwright)
	skiffPricesAddress   = 19822 // C: D_4d6e
	dockXAddress         = 19830 // C: D_4d76
	dockYAddress         = 19834 // C: D_4d7a
)

// Sizes of the tavern tables.
const (
	tavernMenus     = 5
	tavernsCount    = 9
	wineKinds       = 6
	rationLines     = 7
	loreTopicCount  = 26
	lorePlaceCount  = 13
	tavernCount     = 10
	shipwrightCount = 4
)

var (
	mealKeys      []byte   // C: D_4c1e
	drinkKeys     []byte   // C: D_4c24
	foodKeys      []byte   // C: D_4c2a
	loreKeys      []byte   // C: D_4c30
	mealPrices    []uint16 // C: D_4c36
	winePrices    []uint16 // C: D_4c48
	rationPrices  []uint16 // C: D_4c54
	rationWords   []uint16 // C: D_4c66
	loreTopics    []string // C: D_4c74
	loreThings    []string // C: D_4ca8
	lorePlaceOf   []byte   // C: D_4cdc
	lorePlaces    []string // C: D_4cf6
	lorePrices    []uint16 // C: D_4d10
	loreWords     []uint16 // C: D_4d44
	tavernMenu    []byte   // C: D_4d4c
	menuWords     []uint16 // C: D_4d56
	moreWords     []uint16 // C: D_4d5e
	frigatePrices []uint16 // C: D_4d66
	skiffPrices   []uint16 // C: D_4d6e
	dockX         []byte   // C: D_4d76
	dockY         []byte   // C: D_4d7a

	// The state of a visit to the tavern or the shipwright.
	menu         int  // C: D_bd16
	boughtAny    bool // C: D_bd18
	mealsServed  int  // C: D_bd1a (food the meal gives)
	guests       int  // C: D_bd1c
	servedTable  int  // C: D_bd1e
	drinksServed int  // C: D_bd20
	sentAway     bool // C: D_bd22
	shipBought   int  // C: D_bd24
)

// loadTavernTables reads the tavern's and the shipwright's tables.
func loadTavernTables() {
	mealKeys = segment.Bytes(mealKeysAddress, tavernMenus)
	drinkKeys = segment.Bytes(drinkKeysAddress, tavernMenus)
	foodKeys = segment.Bytes(foodKeysAddress, tavernMenus)
	loreKeys = segment.Bytes(loreKeysAddress, tavernMenus)
	mealPrices = segment.Words(mealPricesAddress, tavernsCount)
	winePrices = segment.Words(winePricesAddress, wineKinds)
	rationPrices = segment.Words(rationPricesAddress, tavernsCount)
	rationWords = segment.Words(rationWordsAddress, rationLines)
	loreTopics = messages(loreTopicsAddress, loreTopicCount)
	loreThings = messages(loreThingsAddress, loreTopicCount)
	lorePlaceOf = segment.Bytes(lorePlaceOfAddress, loreTopicCount)
	lorePlaces = messages(lorePlacesAddress, lorePlaceCount)
	lorePrices = segment.Words(lorePricesAddress, loreTopicCount)
	loreWords = segment.Words(loreWordsAddress, shopLines)
	tavernMenu = segment.Bytes(tavernMenusAddress, tavernCount)
	menuWords = segment.Words(menuWordsAddress, shopLines)
	moreWords = segment.Words(moreWordsAddress, shopLines)
	frigatePrices = segment.Words(frigatePricesAddress, shipwrightCount)
	skiffPrices = segment.Words(skiffPricesAddress, shipwrightCount)
	dockX = segment.Bytes(dockXAddress, shipwrightCount)
	dockY = segment.Bytes(dockYAddress, shipwrightCount)
}

// SHOPPE.DAT offsets of the tavern's and shipwright's lines.
const (
	noWineAt         = 5139
	hungryAt         = 5183
	loreOfferAt      = 4942
	loreTooPoorAt    = 5226
	shipMenuAt       = 6379
	frigateOfferAt   = 6206
	skiffOfferAt     = 6284
	shipPriceAt      = 6736
	ownFrigateAt     = 6642
	shipTooPoorAt    = 6542
	shipDockedAt     = 6571
	anythingElseAt   = 6618
	skiffOnBoardAt   = 6459
	skiffFreeAt      = 6495
	drinksBeforeTalk = 3
	drunkenTurns     = 25
	rationMeals      = 25
	tooMuchMoney     = 10000
	hungryFood       = 3
)

// Bits of boughtShip.
const (
	ownsFrigate = 128
	ownsSkiff   = 64
	ownsShip    = ownsFrigate | ownsSkiff
)

// countGuests counts the living members and the bill at price each.
// C: SHOPPES2_0000
func countGuests(price int) {
	mealsServed, guests, shopPrice = 0, 0, 0
	for row := 0; row != int(partySize); row++ {
		if party[row].status != statusDead {
			guests++
			mealsServed++
			shopPrice += price
		}
	}
	printChar('\n')
	printChar('\n')
}

// printGuests names the number of guests.
// C: SHOPPES2_006a
func printGuests() {
	switch guests {
	case 2:
		printMessage(msgTwo)
	case 3:
		printMessage(msgThree)
	case 4:
		printMessage(msgFour)
	case 5:
		printMessage(msgFive)
	case 6:
		printMessage(msgSix)
	}
}

// sirOrMilady addresses the Avatar.
// C: SHOPPES2_00ac
func sirOrMilady() {
	activeMember = noMember
	if party[0].gender == maleGender {
		printMessage(msgSir)
	} else {
		printMessage(msgMilady)
	}
}

// presentBill asks the bill for the guests and serves them; it tells whether the party
// was thrown out.
// C: SHOPPES2_00dc
func presentBill() bool {
	printChar('"')
	printMessage(msgThatWillBeSp)
	printNumber(shopPrice, 1, ' ')
	printMessage(msgGoldForTheSp)
	printGuests()
	printMessage(msgOfYeLine)
	sirOrMilady()
	printChar('.')
	if shopPrice > int(gold) {
		printMessage(msgCanTPayBeatItYells)
		printString(shopOwner)
		printChar('.')
		printChar('\n')
		return true
	}
	gold -= int16(shopPrice)
	shadowlordTax()
	if mealsServed != 0 {
		increaseInt(&food, mealsServed, maxFood)
		if at := tileAt(int(mapX), int(mapY)-1); *at == mapTileTable149 {
			*at = mapTileTable155
			servedTable = mapTileTable155
			updateFrame()
		} else if at := tileAt(int(mapX), int(mapY)+1); *at == mapTileTable149 {
			*at = mapTileTable154
			servedTable = mapTileTable154
			updateFrame()
		}
	} else {
		drinksServed++
	}
	printMessage(msgEnjoyLine)
	return false
}

// serveMeal serves the party a meal.
// C: SHOPPES2_01d2
func serveMeal() bool {
	printChar(int(mealKeys[menu]))
	countGuests(int(mealPrices[shopIndex]))
	return presentBill()
}

// Results of the tavern's services.
const (
	served    = 0
	thrownOut = 1
	declined  = 2
)

// serveDrink serves the party ale, or wine where the tavern has a wine list.
// C: SHOPPES2_01f4
func serveDrink() int {
	printChar(int(drinkKeys[menu]))
	if drinksServed == drinksBeforeTalk {
		printMessage(msgIBegThyPardonSp)
		sirOrMilady()
		printMessage(msgSaysSp)
		printString(shopOwner)
		printMessage(msgButHavenTYeHadEnough)
		for answered := false; !answered; {
			switch getChar() {
			case 'Y':
				printMessage(msgYesLine2)
				return served
			case 'N':
				printMessage(msgNo2)
				drunkTurns = drunkenTurns
				decreaseByte(&karma, 1)
				answered = true
			}
		}
	}
	countGuests(1)
	if drinkKeys[menu] != 'W' {
		mealsServed = 0
		if presentBill() {
			return thrownOut
		}
		return served
	}
	printMessage(msgOurWineListLine)
	sirOrMilady()
	printMessage(msgChar46Newlines2)
	printMessage(msgARose18Line)
	printMessage(msgBClaret192Line)
	printMessage(msgCSauterne79Line)
	printMessage(msgDMuscatel30Line)
	printMessage(msgEMoselle275Line)
	printMessage(msgFChablis98Line)
	printMessage(msgThyChoiceSp)
	c := 0
	for c < 'A' || c > 'F' {
		c = getChar()
		if c == keySpace {
			printChar('\n')
			printChar('\n')
			sayShopFile(noWineAt)
			return declined
		}
	}
	printChar(c)
	printMessage(msgAhAFineChoiceSp)
	sirOrMilady()
	printChar('.')
	wine := c - 'A'
	if int(gold) < int(int16(winePrices[wine])) {
		printMessage(msgCanTPayBeatItYells)
		printString(shopOwner)
		printChar('.')
		printChar('\n')
		return thrownOut
	}
	gold -= int16(winePrices[wine])
	shadowlordTax()
	drinksServed++
	printMessage(msgEnjoy)
	printChar('\n')
	printChar('\n')
	return served
}

// sellRations sells the party food to take along.
// C: SHOPPES2_0380
func sellRations(buyer int) int {
	printChar(int(foodKeys[menu]))
	printChar('\n')
	printChar('\n')
	shopPrice = discounted(int(rationPrices[shopIndex]), buyer)
	printChar('"')
	sayShopFile(int(rationWords[randomRange(0, 6)]))
	printMessage(msgHowManyWouldstThouLikeSp)
	wanted := readNumber(2)
	if wanted == 0 {
		printMessage(msgHrumph)
		return declined
	}
	for left := wanted; left != 0; left-- {
		if int(gold) >= shopPrice {
			decreaseInt(&gold, shopPrice)
			increaseInt(&food, rationMeals, maxFood)
			if food == maxFood {
				break
			}
			continue
		}
		bought := wanted - left
		printMessage(msgNewlines2)
		if bought != 0 {
			printMessage(msgThouCanstAffordOnlySp)
			printNumber(bought, 1, ' ')
			printMessage(msgChar33Char34Newlines2)
			return served
		}
		if food < hungryFood {
			increaseInt(&food, randomRange(0, 1)+1, maxFood)
			sayShopFile(hungryAt)
			return thrownOut
		}
		printMessage(msgThouHastNeitherGoldNorNeed)
		printMessage(msgYellsSp)
		printString(shopOwner)
		printMessage(msgChar46Newline)
		return thrownOut
	}
	printMessage(msgNewlines2)
	shadowlordTax()
	return served
}

// sellLore tells the party where things are, for a price.
// C: SHOPPES2_0508
func sellLore() bool {
	printChar(int(loreKeys[menu]))
	printMessage(msgNewlines2Char34)
	for {
		printMessage(msgOfWhatWouldstThouHearMy)
		sirOrMilady()
		printMessage(msgYouRespondLine)
		talkInput = [talkInputSize]byte{}
		copy(talkInput[:], readString(talkInputLen))
		printMessage(msgNewlines2)
		if talkInput[0] == 0 {
			return false
		}
		topic := -1
		for t := range loreTopicCount {
			at := findWord([]byte(loreTopics[t]+"\x00"), talkInput[:])
			if at == 0 || (at > 0 && talkInput[at-1] == ' ') {
				topic = t
				break
			}
		}
		if topic == -1 {
			printMessage(msgThatICannotHelpTheeWith)
			continue
		}
		shopPrice = int(lorePrices[topic])
		sayShopFile(loreOfferAt)
		printMessage(msgFairNuffSp)
		c := 0
		for c != 'N' && c != 'Y' {
			c = getChar()
		}
		if c == 'N' {
			printMessage(msgNoLine3)
			return false
		}
		printMessage(msgYesLine2)
		if shopPrice > int(gold) {
			printMessage(msgSorrySp)
			sirOrMilady()
			sayShopFile(loreTooPoorAt)
			return true
		}
		gold -= int16(shopPrice)
		showStatus()
		shopThing = loreThings[topic]
		shopGuest = lorePlaces[lorePlaceOf[topic]]
		sayShopFile(int(loreWords[randomRange(0, 3)]))
		printMessage(msgSaysSp2)
		printString(shopOwner)
		printMessage(msgChar46Newlines2)
		return false
	}
}

// tavernFarewell has the barkeep thank the party when it bought something.
// C: SHOPPES2_0664
func tavernFarewell() {
	result := 0
	if boughtAny {
		result = 1
	}
	shopFarewell(result)
}

// tavern is the tavern: meals, drinks, food to take along and, after a purchase, lore.
// C: SHOPPES2_066c
func tavern(buyer int) {
	menu = int(tavernMenu[shopIndex])
	servedTable, boughtAny = 0, false
	shopGreeting()
	for c := 0; c != 'Y'; {
		c = getChar()
		if c == 'N' || c == keySpace {
			printMessage(msgNo)
			tavernFarewell()
			return
		}
	}
	printMessage(msgYes2)
	sayShopFile(int(menuWords[menu]))
	printChar('"')
	printChar(' ')
	if textX() > 14 {
		printChar('\n')
	}
	for {
		var result int
		for {
			c := getChar()
			result = served
			if c == keySpace || c == keyEscape || c == keyEnter {
				tavernFarewell()
				return
			}
			if int(mealKeys[menu]) == c && serveMeal() {
				return
			}
			if int(drinkKeys[menu]) == c {
				if result = serveDrink(); result == thrownOut {
					return
				}
			}
			if int(foodKeys[menu]) == c {
				if result = sellRations(buyer); result == thrownOut {
					return
				}
			}
			if int(loreKeys[menu]) == c {
				if !boughtAny {
					continue
				}
				if sellLore() {
					return
				}
			}
			if int(mealKeys[menu]) == c || int(drinkKeys[menu]) == c || int(foodKeys[menu]) == c ||
				int(loreKeys[menu]) == c {
				break
			}
		}
		if result != declined {
			boughtAny = true
		}
		showStatus()
		printMessage(msgAnythingElseForTheeSp)
		for c := 0; c != 'Y'; {
			c = getChar()
			if c == 'N' {
				printMessage(msgNo)
				tavernFarewell()
				return
			}
		}
		printMessage(msgYes2)
		sayShopFile(int(moreWords[menu]))
		printChar('"')
		printChar(' ')
	}
}

// canAffordShip tells whether the party has price; if not the shipwright sends it away.
// C: SHOPPES2_07e2
func canAffordShip(price int) bool {
	if price > int(gold) {
		sayShopFile(shipTooPoorAt)
		sayShopMessage(msgYellsLine2)
		sentAway = true
		return false
	}
	return true
}

// deliverShip docks a new ship, takes the money, and asks whether the party wants more;
// it tells whether the party is done.
// C: SHOPPES2_080e
func deliverShip(docked, pay bool, buyer int) bool {
	if docked {
		sayShopFile(shipDockedAt)
		shipX = dockX[shopIndex]
		shipY = dockY[shopIndex]
	}
	if pay {
		gold -= int16(shopPrice)
		shipBought = 1
		shadowlordTax()
	}
	showStatus()
	sayShopFile(anythingElseAt)
	// fix: the original compares the gender with 'F', so it never says milady.
	if party[buyer].gender == femaleGend {
		printMessage(msgMilady)
	} else {
		printMessage(msgSir)
	}
	printMessage(msgChar63Char34Space)
	c := 0
	for c != 'Y' && c != 'N' {
		c = getChar()
	}
	if c == 'Y' {
		printMessage(msgYes)
		return false
	}
	printMessage(msgNo)
	return true
}

// yesOrNo waits for Y or N.
func yesOrNo() int {
	c := 0
	for c != 'Y' && c != 'N' {
		c = getChar()
	}
	return c
}

// shipMenu sells frigates and skiffs.
// C: SHOPPES2_08a8
func shipMenu(buyer int) {
	for {
		sayShopFile(shipMenuAt)
		done := false
		switch getChar() {
		case 'F':
			printMessage(msgFLine)
			done = true
			if boughtShip&ownsShip != 0 {
				sayShopFile(ownFrigateAt)
				if yesOrNo() == 'Y' {
					printMessage(msgYes)
					canAffordShip(tooMuchMoney)
				} else {
					printMessage(msgNo)
					shipMenu(buyer)
				}
				break
			}
			shopPrice = discounted(int(frigatePrices[shopIndex]), buyer)
			sayShopFile(frigateOfferAt)
			sayShopFile(shipPriceAt)
			if yesOrNo() != 'Y' {
				printMessage(msgNo)
				return
			}
			printMessage(msgYesLine2)
			if !canAffordShip(shopPrice) {
				return
			}
			boughtShip = ownsFrigate | 2
			if deliverShip(true, true, buyer) {
				return
			}
			continue
		case 'S':
			printMessage(msgSLine2)
			shopPrice = discounted(int(skiffPrices[shopIndex]), buyer)
			sayShopFile(skiffOfferAt)
			sayShopFile(shipPriceAt)
			c := yesOrNo()
			done = true
			if c != 'Y' {
				printMessage(msgNo)
				return
			}
			printMessage(msgYesLine2)
			if !canAffordShip(shopPrice) {
				return
			}
			switch {
			case boughtShip&ownsShip == 0:
				boughtShip = ownsSkiff
				done = deliverShip(true, true, buyer)
			case boughtShip&ownsFrigate != 0:
				sayShopFile(skiffOnBoardAt)
				boughtShip++
				done = deliverShip(false, true, buyer)
			default:
				sayShopFile(skiffFreeAt)
				done = deliverShip(false, false, buyer)
			}
		case keyEscape, keySpace:
			done = true
		}
		if done {
			return
		}
	}
}

// shipwright is the shipwright's shop.
// C: SHOPPES2_0abc
func shipwright(buyer int) {
	sentAway = false
	shipBought = 0
	shopGreeting()
	for done := false; !done; {
		switch getChar() {
		case 'Y':
			printMessage(msgYes)
			shipMenu(buyer)
			done = true
		case keySpace, 'N':
			printMessage(msgNo)
			done = true
		}
	}
	if !sentAway {
		shopFarewell(shipBought)
	}
}
