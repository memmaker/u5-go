package game

import "strconv"

// The shops: the merchants' words, the guild, the herbalist, the stables, the armoury and
// the healer (shoppes.c).

// Addresses of the shop tables.
const (
	reagentPricesAddress    = 14898 // C: D_3a32 (by shop, by reagent; 0 not sold)
	reagentAmountsAddress   = 14938 // C: D_3a5a
	armPricesAddress        = 14978 // C: D_3a82 (by item)
	armouryStockAddress     = 15074 // C: D_3ae2 (by shop, 255 ends)
	greetingsAddress        = 15146 // C: D_3b2a (SHOPPE.DAT offsets by kind, 4 each)
	thanksAddress           = 15210 // C: D_3b6a
	refusalsAddress         = 15274 // C: D_3baa
	supplyPricesAddress     = 15338 // C: D_3bea (by guild, by supply)
	supplyWordsAddress      = 15370 // C: D_3c0a (SHOPPE.DAT offsets)
	reagentWordsAddress     = 15376 // C: D_3c10 (SHOPPE.DAT offsets)
	shopReagentNamesAddress = 15392 // C: D_3c20 (pointers)
	horsePricesAddress      = 15408 // C: D_3c30
	stableDeltaXAddress     = 15416 // C: D_3c38
	stableDeltaYAddress     = 15424 // C: D_3c40
	armWordsAddress         = 15432 // C: D_3c48 (SHOPPE.DAT offsets by item)
	armPromptsAddress       = 15526 // C: D_3ca6 (pointers)
	armTooDearAddress       = 15534 // C: D_3cae (pointers)
	armMenuAddress          = 15542 // C: D_3cb6 (pointers)
	armOffersAddress        = 15550 // C: D_3cbe (SHOPPE.DAT offsets)
	armSaleNamesAddress     = 15566 // C: D_3cce (pointers, 0 for the item name)
	sellPromptsAddress      = 15662 // C: D_3d2e (pointers)
	sellDoneAddress         = 15670 // C: D_3d36 (pointers)
	sellMoreAddress         = 15678 // C: D_3d3e (pointers)
	armouryAsksAddress      = 15686 // C: D_3d46 (pointers)
	armouryBuyAddress       = 15690 // C: D_3d4a (pointers)
	armouryBuy2Address      = 15698 // C: D_3d52 (pointers)
	noNeedAddress           = 15706 // C: D_3d5a (pointer)
	healPricesAddress       = 15750 // C: D_3d86 (by healer)
	curePricesAddress       = 15758 // C: D_3d8e
	raisePricesAddress      = 15766 // C: D_3d96
)

// Sizes of the shop tables.
const (
	herbalists      = 5
	guilds          = 4
	stableCount     = 4
	healers         = 8
	shopLines       = 4 // the merchants' lines of each kind
	armouryRows     = 9
	armouryStockMax = 8
	supplyKinds     = 3
	armSaleOffers   = 8
	shopFileBytes   = 1500
	stockEnd        = 255
)

var (
	reagentPrices    [herbalists][reagentTypes]byte // C: D_3a32
	reagentAmounts   [herbalists][reagentTypes]byte // C: D_3a5a
	armPrices        []uint16                       // C: D_3a82
	armouryStock     [armouryRows][armouryStockMax]byte
	greetings        [shopKinds][shopLines]uint16 // C: D_3b2a
	thanks           [shopKinds][shopLines]uint16 // C: D_3b6a
	refusals         [shopKinds][shopLines]uint16 // C: D_3baa
	supplyPrices     [guilds][shopLines]uint16    // C: D_3bea
	supplyWords      []uint16                     // C: D_3c0a
	reagentWords     []uint16                     // C: D_3c10
	shopReagentNames []string                     // C: D_3c20
	horsePrices      []uint16                     // C: D_3c30
	stableDeltaX     []int                        // C: D_3c38
	stableDeltaY     []int                        // C: D_3c40
	armWords         []uint16                     // C: D_3c48
	armPrompts       []string                     // C: D_3ca6
	armTooDear       []string                     // C: D_3cae
	armMenu          []string                     // C: D_3cb6
	armOffers        []uint16                     // C: D_3cbe
	armSaleNames     []string                     // C: D_3cce
	sellPrompts      []string                     // C: D_3d2e
	sellDone         []string                     // C: D_3d36
	sellMore         []string                     // C: D_3d3e
	armouryAsks      []string                     // C: D_3d46
	armouryBuy       []string                     // C: D_3d4a
	armouryBuy2      []string                     // C: D_3d52
	noNeed           string                       // C: D_3d5a
	healPrices       []byte                       // C: D_3d86
	curePrices       []byte                       // C: D_3d8e
	raisePrices      []uint16                     // C: D_3d96

	// The values the merchants' words fill in.
	shopPrice    int    // C: D_b118 ('%')
	shopQuantity int    // C: D_b11a ('^')
	shopThing    string // C: D_ab00 ('&')
	shopGuest    string // C: D_ac62 ('*')
)

// loadShopTables reads the shops' tables.
func loadShopTables() {
	for shop := range herbalists {
		copy(reagentPrices[shop][:], segment.Bytes(reagentPricesAddress+shop*reagentTypes, reagentTypes))
		copy(reagentAmounts[shop][:], segment.Bytes(reagentAmountsAddress+shop*reagentTypes, reagentTypes))
	}
	armPrices = segment.Words(armPricesAddress, itemTypes)
	for row := range armouryRows {
		copy(armouryStock[row][:], segment.Bytes(armouryStockAddress+row*armouryStockMax, armouryStockMax))
	}
	for kind := range shopKinds {
		at := kind * shopLines * 2
		copy(greetings[kind][:], segment.Words(greetingsAddress+at, shopLines))
		copy(thanks[kind][:], segment.Words(thanksAddress+at, shopLines))
		copy(refusals[kind][:], segment.Words(refusalsAddress+at, shopLines))
	}
	for guild := range guilds {
		copy(supplyPrices[guild][:], segment.Words(supplyPricesAddress+guild*shopLines*2, shopLines))
	}
	supplyWords = segment.Words(supplyWordsAddress, supplyKinds)
	reagentWords = segment.Words(reagentWordsAddress, reagentTypes)
	shopReagentNames = messages(shopReagentNamesAddress, reagentTypes)
	horsePrices = segment.Words(horsePricesAddress, stableCount)
	stableDeltaX = segment.SignedWords(stableDeltaXAddress, stableCount)
	stableDeltaY = segment.SignedWords(stableDeltaYAddress, stableCount)
	armWords = segment.Words(armWordsAddress, itemTypes)
	armPrompts = messages(armPromptsAddress, shopLines)
	armTooDear = messages(armTooDearAddress, shopLines)
	armMenu = messages(armMenuAddress, shopLines)
	armOffers = segment.Words(armOffersAddress, armSaleOffers)
	armSaleNames = messages(armSaleNamesAddress, itemTypes)
	sellPrompts = messages(sellPromptsAddress, shopLines)
	sellDone = messages(sellDoneAddress, shopLines)
	sellMore = messages(sellMoreAddress, shopLines)
	armouryAsks = messages(armouryAsksAddress, 2)
	armouryBuy = messages(armouryBuyAddress, shopLines)
	armouryBuy2 = messages(armouryBuy2Address, shopLines)
	noNeed = messages(noNeedAddress, 1)[0]
	healPrices = segment.Bytes(healPricesAddress, healers)
	curePrices = segment.Bytes(curePricesAddress, healers)
	raisePrices = segment.Words(raisePricesAddress, healers)
}

// The marks the merchants' words fill in.
const (
	markShopName = '#'
	markOwner    = '$'
	markDayTime  = '@'
	markPrice    = '%'
	markQuantity = '^'
	markThing    = '&'
	markGuest    = '*'
	firstWord    = 128
)

// sayShopText prints a merchant's words, filling in the marks and the dictionary words.
// C: SHOPPES_0026
func sayShopText(text string) {
	out := make([]byte, 0, len(text))
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case c >= firstWord:
			out = append(out, ' ')
			out = append(out, dictionary[c-firstWord]...)
			if i+1 < len(text) && text[i+1] < firstWord {
				out = append(out, ' ')
			}
		case c == markShopName:
			out = append(out, shopName...)
		case c == markOwner:
			out = append(out, shopOwner...)
		case c == markDayTime:
			switch {
			case hour < noonHour:
				out = append(out, message(msgMorning)...)
			case hour < eveningHour:
				out = append(out, message(msgAfternoon)...)
			default:
				out = append(out, message(msgEvening)...)
			}
		case c == markPrice:
			out = strconv.AppendInt(out, int64(int16(shopPrice)), 10)
		case c == markQuantity:
			out = strconv.AppendInt(out, int64(int16(shopQuantity)), 10)
		case c == markThing:
			out = append(out, shopThing...)
		case c == markGuest:
			out = append(out, shopGuest...)
		default:
			out = append(out, c)
		}
	}
	printString(string(out))
}

// eveningHour is when the merchants start to say "evening".
const eveningHour = 18

// sayShopMessage prints the merchant's words at a message address.
func sayShopMessage(address int) { sayShopText(message(address)) }

// sayShopFile prints the merchant's words at offset in SHOPPE.DAT.
// C: SHOPPES_017a
func sayShopFile(offset int) {
	loadShopFile(offset)
	sayShopText(shopFileText(0))
}

// loadShopFile reads SHOPPE.DAT from offset into the scratch buffer.
func loadShopFile(offset int) { readFile("SHOPPE.DAT", scratch[:shopFileBytes], offset) }

// shopFileText returns the words at at in the part of SHOPPE.DAT read.
func shopFileText(at int) string { return cString(scratch[at:shopFileBytes]) }

// shadowlordTax: the shadowlord of falsehood haunting the town makes the party lose some
// gold on every purchase.
// C: SHOPPES_019a
func shadowlordTax() {
	if shadowlordHere == 0 {
		decreaseInt(&gold, randomRange(1, 64))
	}
}

// shopGreeting has the merchant greet the party with one of its lines.
// C: SHOPPES_01b6
func shopGreeting() {
	printChar('"')
	sayShopFile(int(greetings[shopKind][randomRange(0, 3)]))
	switch x := textX(); {
	case x > 11:
		printMessage(msgNewlines2Char58)
	case x == 0:
		printMessage(msgNewlineChar58)
	default:
		printChar(' ')
	}
}

// shopFarewell has the merchant thank the party (1) or regret it bought nothing (0).
// C: SHOPPES_0202
func shopFarewell(result int) {
	switch result {
	case 0:
		printMessage(msgNewlines2Char34)
		sayShopFile(int(refusals[shopKind][randomRange(0, 3)]))
	case 1:
		printMessage(msgNewlines2Char34)
		sayShopFile(int(thanks[shopKind][randomRange(0, 3)]))
	default:
		return
	}
	if textX() != 0 {
		printChar('\n')
	}
	sayShopMessage(msgSaysLine3)
}

// askYesNo waits for Y or N and echoes it.
// C: SHOPPES_0280
func askYesNo() int {
	defer asking(promptYesNo)()
	for {
		c := getChar()
		switch c {
		case 'N':
			printMessage(msgNo)
			return c
		case 'Y':
			printMessage(msgYes)
			return c
		}
	}
}

// discounted returns a price lowered by the buyer's intelligence.
func discounted(price, buyer int) int {
	return price + price*(100-int(party[buyer].intel)*3)/100
}

// sirOrLady prints how a merchant addresses the buyer.
func sirOrLady(buyer int, lady, lord int) {
	if party[buyer].gender == femaleGend {
		printMessage(lady)
	} else {
		printMessage(lord)
	}
}

// The guild's supplies.
const (
	supplyKeys    = 0
	supplyGems    = 1
	supplyTorches = 2
	keysBundle    = 3
	gemsBundle    = 4
	torchesBundle = 5
	maxSupply     = 99
)

// SHOPPE.DAT offsets of a few lines said outside the tables.
const (
	guildTooPoor   = 8678
	guildAboutWare = 8725
	healerTooPoor  = 9131
	horseOffer     = 5699
	reagentsAt     = 6759
	reagentTooPoor = 1220 // in the reagents' part of the file
	armSaleAt      = 3008
	charityHealer  = 5 // Cove's healer heals for nothing
	charityPrice   = 100
)

// buySupply sells the party keys, gems or torches.
// C: SHOPPES_02ba
func buySupply(buyer, kind, result int) int {
	shopPrice = discounted(int(supplyPrices[shopIndex][kind]), buyer)
	printMessage(msgNewlines2Char34)
	sayShopFile(int(supplyWords[kind]))
	printMessage(msgInterestedSp)
	for {
		switch getChar() {
		case 'N':
			printMessage(msgNoWhatElseThenLine)
			return result
		case 'Y':
			printMessage(msgYesLine)
			if int(gold) < shopPrice {
				sayShopFile(guildTooPoor)
				return -1
			}
			gold -= int16(shopPrice)
			shadowlordTax()
			showStatus()
			switch kind {
			case supplyKeys:
				increaseByte(&keys, keysBundle, maxSupply)
			case supplyGems:
				increaseByte(&gems, gemsBundle, maxSupply)
			case supplyTorches:
				increaseByte(&torches, torchesBundle, maxSupply)
			}
			sayShopMessage(msgSoldSaysWhatElseLine)
			sirOrLady(buyer, msgMLady, msgMLord)
			printMessage(msgChar63Newlines2)
			return 1
		}
	}
}

// supplyMenu lists the guild's supplies until the party is done.
// C: SHOPPES_03f6
func supplyMenu(buyer int) int {
	result := 0
	for done := false; !done; {
		printMessage(msgAKeysLine)
		printMessage(msgBGemsLine)
		printMessage(msgCTorchesLine)
		printMessage(msgThyConcernSp)
		for chosen := false; !chosen && !done; {
			c := getChar()
			chosen = true
			switch {
			case c >= 'A' && c <= 'C':
				printChar(c | lowerCaseBit)
				result = buySupply(buyer, c-'A', result)
				if result == -1 {
					done = true
				}
			case c == keyEscape || c == keySpace:
				done = true
			case c == 'D':
				sayShopFile(guildAboutWare)
			default:
				chosen = false
			}
		}
	}
	return result
}

// lowerCaseBit turns a capital letter into a small one.
const lowerCaseBit = 32

// guild is the guild's shop.
// C: SHOPPES_04a2
func guild(buyer int) {
	result := 0
	shopGreeting()
	for done := false; !done; {
		switch getChar() {
		case 'Y':
			printMessage(msgYesWeSellLine)
			result = supplyMenu(buyer)
			done = true
		case keySpace, 'N':
			printMessage(msgNo)
			done = true
		}
	}
	shopFarewell(result)
}

// buyReagent sells the party a reagent: the choice-th the herbalist sells. It returns -2
// for no such reagent.
// C: SHOPPES_0502
func buyReagent(buyer, choice, result int) int {
	reagent, listed := 0, 0
	for ; reagent < reagentTypes; reagent++ {
		if reagentPrices[shopIndex][reagent] != 0 {
			if listed == choice {
				break
			}
			listed++
		}
	}
	if reagent == reagentTypes {
		return -2
	}
	if reagents[reagent] == maxSupply {
		printMessage(msgThouCanstNotCarryAnyMore)
		getChar()
		return result
	}
	shopPrice = discounted(int(reagentPrices[shopIndex][reagent]), buyer)
	shopQuantity = int(reagentAmounts[shopIndex][reagent])
	printMessage(msgNewlines2Char34)
	sayShopText(shopFileText(int(reagentWords[reagent]) - reagentsAt))
	printMessage(msgIsThisThyNeedSp)
	if textX() > 12 {
		printMessage(msgNewlineChar58)
	}
	for {
		switch getChar() {
		case 'N':
			printMessage(msgNoWhatElseLine)
			return result
		case 'Y':
			printMessage(msgYesLine)
			if int(gold) < shopPrice {
				sayShopText(shopFileText(reagentTooPoor))
				return -1
			}
			gold -= int16(shopPrice)
			shadowlordTax()
			showStatus()
			increaseByte(&reagents[reagent], shopQuantity, maxSupply)
			sayShopMessage(msgIThankTheeSaysLine2)
			printMessage(msgAnythingElseLine)
			return 1
		}
	}
}

// reagentMenu lists the herbalist's reagents until the party is done.
// C: SHOPPES_0666
func reagentMenu(buyer int) int {
	result := 0
	loadShopFile(reagentsAt)
	for done := false; !done; {
		letter := 'A'
		for reagent := range reagentTypes {
			if reagentPrices[shopIndex][reagent] != 0 {
				printChar(int(letter))
				letter++
				printMessage(msgChar46Char46Char46)
				printString(shopReagentNames[reagent])
				printChar('\n')
			}
		}
		printMessage(msgThyInterestSp)
		for chosen := false; !chosen && !done; {
			c := getChar()
			chosen = true
			switch {
			case c >= 'A' && c <= 'E':
				switch r := buyReagent(buyer, c-'A', result); r {
				case -1:
					result = -1
					done = true
				case -2:
					chosen = false
				default:
					result = r
				}
			case c == keyEnter || c == keyEscape || c == keySpace:
				done = true
			default:
				chosen = false
			}
		}
	}
	return result
}

// herbalist is the herbalist's shop.
// C: SHOPPES_075e
func herbalist(buyer int) {
	result := 0
	shopGreeting()
	for done := false; !done; {
		switch getChar() {
		case 'Y':
			printMessage(msgYesFineWeSellLine)
			result = reagentMenu(buyer)
			done = true
		case keySpace, 'N':
			printMessage(msgNo)
			done = true
		}
	}
	shopFarewell(result)
}

// stables sells a horse, set down beside the merchant.
// C: SHOPPES_07be
func stables(buyer int) {
	slot := 0
	for ; slot < actorSlots; slot++ {
		if actors[slot].tile == 0 {
			break
		}
	}
	var x, y, side int
	for ; side < stableCount; side++ {
		x = stableDeltaX[side] + int(mapX)
		y = stableDeltaY[side] + int(mapY)
		tile := *tileAt(x, y)
		if actorAt(x, y, int(level)) == 0 && (tile == mapTile68 || tile == mapTile69 || tile == mapTileGrass) {
			break
		}
	}
	if side == stableCount || slot == actorSlots {
		printMessage(msgTheStablesAreClosedLine)
		return
	}
	result := 0
	shopPrice = discounted(int(horsePrices[shopIndex]), buyer)
	shopGreeting()
	for done := false; !done; {
		switch getChar() {
		case 'Y':
			printMessage(msgYes2)
			sayShopFile(horseOffer)
			printMessage(msgDealSp)
			c := 0
			for c != 'Y' && c != 'N' {
				c = getChar()
			}
			done = true
			if c == 'N' {
				printMessage(msgNo)
				break
			}
			printMessage(msgYes3)
			if int(gold) < shopPrice {
				printMessage(msgThouCouldstNotAffordToSp)
				sayShopMessage(msgFeedItYellsLine)
				result = -1
				break
			}
			gold -= int16(shopPrice)
			shadowlordTax()
			a := &actors[slot]
			a.state, a.target, a.hull = 0, 0, 0
			a.tile, a.animTile = actorHorse, actorHorse
			a.x, a.y, a.z = byte(x), byte(y), level
			updateFrame()
			showStatus()
			result = 1
		case keySpace, 'N':
			printMessage(msgNo)
			done = true
		}
	}
	shopFarewell(result)
}

// Items sold as a bundle the party always has enough of.
const (
	fullStock = 99
)

// buyArm sells the party a weapon or armour.
// C: SHOPPES_09ac
func buyArm(buyer, item, result int) int {
	shopPrice = discounted(int(armPrices[item]), buyer)
	printMessage(msgNewlines2Char34)
	sayShopFile(int(int16(armWords[item])))
	printMessage(msgNewlines2)
	printString(armPrompts[randomRange(0, 3)])
	printMessage(msgChar34Space)
	for answered := false; !answered; {
		switch getChar() {
		case 'N':
			printMessage(msgNoLine3)
			answered = true
		case 'Y':
			printMessage(msgYesLine)
			answered = true
			switch {
			case equipment[item] == fullStock:
				printMessage(msgThouCanstNotCarryAnyMore2)
				sayShopMessage(msgSaysLine2)
				getChar()
			case int(gold) < shopPrice:
				printMessage(msgNewlineChar34)
				printString(armTooDear[randomRange(0, 3)])
				sayShopMessage(msgYellsLine)
				result = -1
			default:
				gold -= int16(shopPrice)
				shadowlordTax()
				showStatus()
				if item == arrowsItem || item == quarrelsItem {
					equipment[item] = fullStock
				} else {
					increaseByte(&equipment[item], 1, fullStock)
				}
				sayShopMessage(msgSoldLine)
				result = 1
			}
		}
	}
	if result != -1 {
		printMessage(msgAnythingElseLine2)
		if result != 0 {
			sirOrLady(buyer, msgMilady2, msgSir2)
		} else {
			printMessage(msgThen)
		}
	}
	return result
}

// shortNameLimit is the longest item name the armoury lists; longer ones use the arms name.
const shortNameLimit = 13

// armsMenu lists the armoury's wares until the party is done.
// C: SHOPPES_0b30
func armsMenu(buyer int) int {
	result := 0
	for done := false; !done; {
		printMessage(msgNewlines2)
		count := 0
		for ; count < armouryStockMax; count++ {
			item := armouryStock[shopIndex][count]
			if item == stockEnd {
				break
			}
			printChar('a' + count)
			printMessage(msgChar46Char46Char46)
			if len(itemNames[item]) < shortNameLimit {
				printString(itemNames[item])
			} else {
				printString(armsNames[item])
			}
			if textX() != 0 {
				printChar('\n')
			}
		}
		printChar('\n')
		printString(armMenu[randomRange(0, 3)])
		printMessage(msgChar34Space)
		for chosen := false; !chosen && !done; {
			c := getChar()
			choice := c - 'A'
			chosen = true
			switch {
			case choice < count && choice >= 0:
				printChar(choice + 'a')
				result = buyArm(buyer, int(armouryStock[shopIndex][choice]), result)
				if result == -1 {
					done = true
				}
			case c == keyEscape || c == keySpace:
				done = true
			default:
				chosen = false
			}
		}
	}
	return result
}

// ownsArms tells whether the party has anything to sell.
// C: SHOPPES_0c58
func ownsArms() bool {
	for item := range itemTypes {
		if equipment[item] != 0 {
			return true
		}
	}
	return false
}

// The sale list in the panel.
const (
	saleListRows   = 5
	saleListBlank  = "             "
	saleArrowX     = 6
	saleArrowY     = 6
	saleListMarker = '-'
)

// showSaleList shows the party's arms in the panel from top with selected marked, mending
// the two when the selected item is sold out. It returns the selected row.
// C: SHOPPES_0c80
func showSaleList(top, selected *int) int {
	if equipment[*selected] == 0 {
		if *top == *selected {
			*selected = previousListed(*selected, equipment[:], noOwner)
			if *selected == -1 {
				*selected = nextListed(-1, itemTypes, equipment[:], noOwner)
			}
			*top = *selected
		} else {
			*selected = nextListed(*selected, itemTypes, equipment[:], noOwner)
			if *selected == -1 {
				*selected = previousListed(itemTypes, equipment[:], noOwner)
				if p := previousListed(*top, equipment[:], noOwner); p > -1 {
					*top = p
				}
			}
		}
	}
	selectWindow(panelWindow)
	setTextPosition(1, 1)
	row := 0
	at := *top
	for at != -1 {
		if *selected == at {
			printChar(charToggleInverse)
			row = textY()
		}
		printListEntry(at, equipment[:], armsNames, saleListMarker)
		if *selected == at {
			printChar(charToggleInverse)
		}
		setTextPosition(1, textY())
		if textY() == saleListRows {
			break
		}
		at = nextListed(at, itemTypes, equipment[:], noOwner)
	}
	for y := textY(); y != saleListRows; y++ {
		setTextPosition(1, y)
		printString(saleListBlank)
	}
	more := 0
	if previousListed(*top, equipment[:], noOwner) != -1 {
		more = 2
	}
	if at != -1 && nextListed(at, itemTypes, equipment[:], noOwner) != -1 {
		more++
	}
	if more != 0 {
		setTextPosition(saleArrowX, saleArrowY)
		printOpenCaption()
	}
	switch more {
	case 0:
		clearIconBar()
	case 1:
		printChar(arrowDown)
		printCloseCaption()
	case 2:
		printChar(arrowUp)
		printCloseCaption()
	case 3:
		printChar(arrowUpDown)
		printCloseCaption()
	}
	selectWindow(2)
	return row
}

// sellArm buys an item from the party. It returns 1 when the merchant refused to deal.
// C: SHOPPES_0e76
func sellArm(seller, item int) int {
	if item == arrowsItem || item == quarrelsItem {
		sayShopMessage(msgWeDonTDealInUsed)
		return 1
	}
	printMessage(msgNewlines2Char34)
	if armPrices[item] == 0 {
		sayShopMessage(msgThatICannotBuyFromThee)
		return 0
	}
	shopPrice = int(party[seller].intel)*int(armPrices[item])*3/100 + 1
	shopThing = armSaleNames[item]
	if shopThing == "" {
		shopThing = itemNames[item]
	}
	sayShopText(shopFileText(int(armOffers[randomRange(0, 7)]) - armSaleAt))
	printMessage(msgDealSp)
	c := 0
	for c != 'N' && c != 'Y' {
		c = getChar()
	}
	if c == 'N' {
		printMessage(msgNo)
	} else {
		sayShopMessage(msgYesDoneSays)
		increaseInt(&gold, shopPrice, maxGold)
		decreaseByte(&equipment[item], 1)
	}
	showGold()
	return 0
}

// sellArms lets the party sell its arms from a list in the panel.
// C: SHOPPES_0f64
func sellArms(seller int) {
	if !ownsArms() {
		sayShopMessage(msgThouHastNothingToSellGrowls)
		return
	}
	printString(sellPrompts[randomRange(0, 3)])
	printMessage(msgChar34Space)
	showCaption(msgArms)
	selectWindow(panelWindow)
	drawListFrame(saleListRows)
	selected := nextListed(-1, itemTypes, equipment[:], noOwner)
	top := selected
	row := showSaleList(&top, &selected)
	refused := 0
	loadShopFile(armSaleAt)
	for done := false; !done; {
		switch key := getChar(); key {
		case keyLeft, keyUp, keyPageUp:
			steps := 1
			if key == keyPageUp {
				steps = 4
			}
			for range steps {
				previous := previousListed(selected, equipment[:], noOwner)
				if selected == top {
					if previous > -1 {
						top = previous
						selected = previous
					}
				} else {
					selected = previous
				}
			}
			row = showSaleList(&top, &selected)
		case keyRight, keyDown, keyPageDown:
			steps := 1
			if key == keyPageDown {
				steps = 4
			}
			for range steps {
				if next := nextListed(selected, itemTypes, equipment[:], noOwner); next > -1 {
					selected = next
					if row++; row > saleListRows-1 {
						top = nextListed(top, itemTypes, equipment[:], noOwner)
					}
				}
			}
			row = showSaleList(&top, &selected)
		case keyHome:
			selected = nextListed(-1, itemTypes, equipment[:], noOwner)
			top = selected
			row = showSaleList(&top, &selected)
		case keyEnd:
			selected = previousListed(itemTypes, equipment[:], noOwner)
			top = selected
			for range 3 {
				p := previousListed(top, equipment[:], noOwner)
				if p == -1 {
					break
				}
				top = p
			}
			row = showSaleList(&top, &selected)
		case keyEnter, keySpace:
			refused = sellArm(seller, selected)
			if refused == 0 && ownsArms() {
				row = showSaleList(&top, &selected)
				printMessage(msgNewlines2Char34)
				printString(sellMore[randomRange(0, 3)])
				printMessage(msgChar34Space)
			} else {
				done = true
			}
		case keyEscape:
			done = true
		}
	}
	showStatus()
	clearCaption()
	if refused != 0 {
		return
	}
	printMessage(msgNewlines2Char34)
	printString(sellDone[randomRange(0, 3)])
	printMessage(msgChar34Newline)
	if !ownsArms() {
		return
	}
	sayShopMessage(msgSaysLine3)
}

// armoury is the armourer's shop: buy or sell.
// C: SHOPPES_12b2
func armoury(buyer int) {
	result := 0
	sayShopMessage(msgGoodAndWelcomeToLine)
	getChar()
	sayShopMessage(msgSays)
	printString(armouryAsks[randomRange(0, 1)])
	printMessage(msgChar34Space)
	for done := false; !done; {
		switch getChar() {
		case 'B':
			printMessage(msgBuy)
			printString(armouryBuy[randomRange(0, 3)])
			printString(armouryBuy2[randomRange(0, 3)])
			result = armsMenu(buyer)
			done = true
		case 'S':
			printMessage(msgSell)
			sellArms(buyer)
			result = -1
			done = true
		case keySpace:
			printMessage(msgNo)
			done = true
		}
	}
	shopFarewell(result)
}

// whoNeedsAid asks which member the healer is to help (-1 none).
// C: SHOPPES_137c
func whoNeedsAid() int {
	if partySize == 1 {
		return 0
	}
	printMessage(msgWhoNeedsMyAidSp)
	who := selectOne()
	if who == -1 {
		printMessage(msgNoOne)
	}
	return who
}

// The healer's flash and chord.
const (
	healFreq1  = 4110
	healTime1  = 22500
	healWidth1 = 5000
	healTop1   = 27500
	healFreq2  = 4530
	healTime2  = 40000
	healFreq3  = 2300
	healTime3  = 18000
	healTop3   = 36000
	flashEdge  = 8
	flashFar   = 183
)

// healingLight flashes the view and sounds the healer's chord.
// C: SHOPPES_13b0
func healingLight() {
	setPenColour(redColour)
	xorFillRect(flashEdge, flashEdge, flashFar, flashFar)
	pulse(healFreq1, 1, healTime1, healWidth1, 1)
	pulse(healFreq1, 1, healTime1, healTop1, -1)
	setPenColour(brightWhiteColour)
	xorFillRect(flashEdge, flashEdge, flashFar, flashFar)
	pulse(healFreq2, 1, healTime2, 1, 1)
	pulse(healFreq2, 1, healTime2, healTime2, -1)
	xorFillRect(flashEdge, flashEdge, flashFar, flashFar)
	pulse(healFreq3, 1, healTime3, 1, 2)
	pulse(healFreq3, 1, healTime3, healTop3, -2)
}

// askFee names the healer's fee and asks the party to pay; it tells whether it did not.
// C: SHOPPES_146a
func askFee() bool {
	sayShopMessage(msgForGoldWiltThouPaySp)
	c := 0
	for c != 'Y' && c != 'N' {
		c = getChar()
		switch c {
		case 'Y':
			printMessage(msgYes)
		case 'N':
			printMessage(msgNo)
		}
	}
	refused := c == 'N'
	if !refused && int(gold) < shopPrice && (shopPrice > charityPrice || mapID != 7) {
		sayShopFile(healerTooPoor)
		refused = true
	}
	if !refused && int(gold) >= shopPrice {
		gold -= int16(shopPrice)
		shadowlordTax()
	}
	return refused
}

// healer is the healer's shop: cure, heal and raise the dead.
// C: SHOPPES_14f8
func healer(int) {
	shopGreeting()
	for done := false; !done; {
		c := 0
		for c != 'N' && c != 'Y' {
			c = getChar()
			switch c {
			case 'N':
				printMessage(msgNo)
			case 'Y':
				printMessage(msgYesLine2)
			}
		}
		if c == 'N' {
			break
		}
		printMessage(msgWeHavePowersToCureHeal)
		sayShopMessage(msgSaysWhatIsTheNatureOf)
		for c != 'C' && c != 'H' && c != 'R' && c != keySpace && c != keyEnter {
			c = getChar()
		}
		switch c {
		case 'C':
			printMessage(msgCuring)
			who := whoNeedsAid()
			if who == -1 {
				break
			}
			if party[who].status != statusPoisoned {
				sayShopText(noNeed)
				break
			}
			printMessage(msgNewlines2Char34)
			if mapID == charityHealer {
				printMessage(msgReceiveNowTheLight)
				healingLight()
				party[who].status = statusGood
			} else {
				printMessage(msgICanCureThyPoisonedBody)
				shopPrice = int(curePrices[shopIndex])
				if !askFee() {
					healingLight()
					party[who].status = statusGood
				}
			}
		case 'H':
			printMessage(msgHealing)
			who := whoNeedsAid()
			if who == -1 {
				break
			}
			m := &party[who]
			if m.status == statusDead || m.hp == m.maxHP {
				sayShopText(noNeed)
				break
			}
			printMessage(msgNewlines2Char34)
			if mapID == charityHealer {
				printMessage(msgReceiveNowTheLight)
				healingLight()
				m.hp = m.maxHP
			} else {
				printMessage(msgICanHealTheeSp)
				shopPrice = int(healPrices[shopIndex])
				if !askFee() {
					healingLight()
					m.hp = m.maxHP
				}
			}
		case 'R':
			printMessage(msgResurrect)
			who := whoNeedsAid()
			if who == -1 {
				break
			}
			if party[who].status != statusDead {
				sayShopText(noNeed)
				break
			}
			printMessage(msgNewlines2Char34)
			printMessage(msgICanRaiseThisUnfortunatePerson)
			printMessage(msgTheDeadSp)
			shopPrice = int(raisePrices[shopIndex])
			if !askFee() {
				healingLight()
				resurrect(who, 255)
				party[who].hp = party[who].maxHP
				showStatus()
			}
		case keyEnter, keySpace:
			printMessage(msgNothing)
			done = true
		}
		if !done {
			showStatus()
			printMessage(msgIsThereAnyOtherWayIn)
			printMessage(msgAidTheeSp)
		}
	}
	shopFarewell(1)
}
