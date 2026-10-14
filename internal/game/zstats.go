package game

// The party's statistics and equipment (zstats.c): the Z-stats pages, the owned-item
// lists, and readying arms.

// Addresses of the overlay's tables in the data segment.
const (
	ownedItemNamesAddress = 6422  // C: D_1916
	armsNamesAddress      = 6498  // C: D_1962
	spellNamesAddress     = 6626  // C: D_19e2
	classNamesAddress     = 6724  // C: D_1a44
	classIndentsAddress   = 6744  // C: D_1a58
	healthNamesAddress    = 6762  // C: D_1a6a
	armsSlotsAddress      = 6782  // C: D_1a7e
	armsWeightsAddress    = 6830  // C: D_1aae
	armsGlyphsAddress     = 6888  // C: D_1ae8
	statusLettersAddress  = 38608 // the statuses' letters, "GPDSC"
)

const (
	ownedItemKinds = 38
	armsNameCount  = 56
	healthCount    = 5
)

// The overlay's tables.
var (
	ownedItemNames []string // C: D_1916
	armsNames      []string // C: D_1962
	spellNames     []string // C: D_19e2
	classNames     []string // C: D_1a44
	classIndents   []int    // C: D_1a58
	healthNames    []string // C: D_1a6a
	armsSlots      []byte   // C: D_1a7e
	armsWeights    []byte   // C: D_1aae
	armsGlyphs     []byte   // C: D_1ae8
	statusLetters  string
)

// ownedItems are the counts of the items listed, gathered from the save (D_b9ee).
var ownedItems [ownedItemKinds]byte // C: D_b9ee

// loadStatsTables reads the overlay's tables.
func loadStatsTables() {
	ownedItemNames = messages(ownedItemNamesAddress, ownedItemKinds)
	armsNames = messages(armsNamesAddress, armsNameCount)
	spellNames = messages(spellNamesAddress, spellCount)
	classNames = messages(classNamesAddress, classCount)
	classIndents = make([]int, classCount)
	for i := range classIndents {
		classIndents[i] = int(int16(segment.Word(classIndentsAddress + i*2)))
	}
	healthNames = messages(healthNamesAddress, healthCount)
	armsSlots = segment.Bytes(armsSlotsAddress, itemTypes)
	armsWeights = segment.Bytes(armsWeightsAddress, itemTypes)
	armsGlyphs = segment.Bytes(armsGlyphsAddress, itemTypes)
	statusLetters = message(statusLettersAddress)
}

// The status panel's window while it shows a page.
const (
	panelWindow    = 1
	panelLeft      = 24
	panelTop       = 1
	panelRight     = 38
	panelFullRight = 39
	panelBottom    = 9
	panelLastRow   = 8
	panelWidth     = 14
	wholeParty     = -2
	noneChosen     = -1
)

// choosePlayer picks the member a command is for: the combatant in combat, else asked.
// C: ZSTATS_0000
func choosePlayer(everyone bool) int {
	who := noneChosen
	if mapID > firstMapCombat && combatants[combatant].flags&fighterPlayer != 0 {
		who = int(combatants[combatant].entity)
	} else {
		printMessage(msgPlayerSp)
		who = selectMember(everyone)
		if who >= 0 {
			printString(party[who].nameString())
			if textX() != 0 {
				printChar('\n')
			}
		}
	}
	switch who {
	case noneChosen:
		printMessage(msgNoneLine)
	case wholeParty:
		printChar('\n')
	}
	return who
}

// clearPanel clears the panel window and sets it to its full width.
func clearPanel() {
	setWindowSize(panelWindow, panelLeft, panelTop, panelRight, panelBottom)
	printChar(charClearWindow)
	setWindowSize(panelWindow, panelLeft, panelTop, panelFullRight, panelBottom)
}

// showCharacter shows a member's statistics.
// C: ZSTATS_0082
func showCharacter(row int) {
	clearMiddleBar()
	m := &party[row]
	showCaptionText(m.nameString())
	clearPanel()
	class := indexOf(classLetters, m.class)
	health := indexOf(statusLetters, m.status)
	for range classIndents[class] {
		printChar(' ')
	}
	printChar(int(m.gender))
	printMessage(msgLv)
	printNumber(int(m.level), 1, ' ')
	printChar(' ')
	printString(classNames[class])
	setWindowSize(panelWindow, panelLeft, panelTop, panelRight, panelBottom)
	setTextPosition(0, 1)
	printChar(charCentre)
	printString(healthNames[health])
	setWindowSize(panelWindow, panelLeft, panelTop, panelFullRight, panelBottom)
	setTextPosition(0, 3)
	printChar(charUncentre)
	printMessage(msgStr2)
	printNumber(int(m.str), 2, '0')
	printMessage(msgHp)
	printNumber(int(m.hp), 4, ' ')
	printMessage(msgInt2)
	printNumber(int(m.intel), 2, '0')
	printMessage(msgHm)
	printNumber(int(m.maxHP), 4, ' ')
	printMessage(msgDex2)
	printNumber(int(m.dex), 2, '0')
	printMessage(msgEx)
	printNumber(int(m.exp), 4, ' ')
	printMessage(msgMagic)
	printNumber(int(m.magic), 2, ' ')
}

// printArm prints a readied arm's name; it tells whether the slot held one.
// C: ZSTATS_0278
func printArm(item byte) int {
	if item == noEquipment {
		return 0
	}
	printChar(' ')
	printString(armsNames[item])
	printChar('\n')
	return 1
}

const noneReadyRow = 4

// showArms shows the arms a member has ready.
// C: ZSTATS_02a8
func showArms(row int) {
	clearMiddleBar()
	m := &party[row]
	showCaptionText(m.nameString())
	setWindowSize(panelWindow, panelLeft, panelTop, panelRight, panelBottom)
	printChar(charClearWindow)
	printChar(charCentre)
	printChar(charToggleUnderline)
	printMessage(msgArmsLine)
	printChar(charUncentre)
	printChar(charToggleUnderline)
	count := 0
	for _, item := range m.equip {
		count += printArm(item)
	}
	if count == 0 {
		setTextPosition(0, noneReadyRow)
		printChar(charCentre)
		printMessage(msgNoneReady)
		printChar(charUncentre)
	}
	setWindowSize(panelWindow, panelLeft, panelTop, panelFullRight, panelBottom)
}

// showSupplies shows the party's food, gold and tools.
// C: ZSTATS_039c
func showSupplies() {
	clearMiddleBar()
	showCaption(msgEquipment)
	clearPanel()
	printMessage(msgFoodSp)
	printNumber(int(food), 4, ' ')
	printMessage(msgGoldSp)
	printNumber(int(gold), 4, ' ')
	printMessage(msgKeys)
	printNumber(int(keys), 2, ' ')
	printMessage(msgGems)
	printNumber(int(gems), 2, ' ')
	printMessage(msgTorches)
	printNumber(int(torches), 2, ' ')
	if grapple != 0 {
		printMessage(msgGrapple)
	}
}

// Glyphs of the list frame (the alternate font's box lines).
const (
	frameTopLeft     = 16
	frameTop         = 17
	frameTopRight    = 19
	frameSide        = 23
	frameBottomLeft  = 20
	frameBottom      = 21
	frameBottomRight = 22
	frameInnerWidth  = 13
	frameRightColumn = 14
)

// drawListFrame frames a list rows high in the panel.
// C: ZSTATS_045e
func drawListFrame(rows int) {
	setWindowSize(panelWindow, panelLeft, panelTop, panelRight, rows+1)
	printChar(charClearWindow)
	setWindowSize(panelWindow, panelLeft, panelTop, panelFullRight, panelBottom)
	printChar(frameTopLeft)
	for range frameInnerWidth {
		printChar(frameTop)
	}
	printChar(frameTopRight)
	for row := 1; row != rows; row++ {
		setTextPosition(0, row)
		printChar(frameSide)
		setTextPosition(frameRightColumn, row)
		printChar(frameSide)
	}
	printChar('\n')
	printChar(frameBottomLeft)
	for range frameInnerWidth {
		printChar(frameBottom)
	}
	printChar(frameBottomRight)
}

// hasReady tells whether member row has item ready.
// C: ZSTATS_0518
func hasReady(row, item int) bool {
	for _, e := range party[row].equip {
		if int(e) == item {
			return true
		}
	}
	return false
}

// noOwner lists only what is owned, not what someone has ready.
const noOwner = 255

// previousListed returns the listed entry before at (-1 none): owned, or ready by owner.
// C: ZSTATS_056c
func previousListed(at int, counts []byte, owner int) int {
	for at--; at >= 0; at-- {
		if counts[at] != 0 || (owner != noOwner && hasReady(owner, at)) {
			return at
		}
	}
	return -1
}

// nextListed returns the listed entry after at, below end (-1 none).
// C: ZSTATS_05a4
func nextListed(at, end int, counts []byte, owner int) int {
	for at++; at < end; at++ {
		if counts[at] != 0 || (owner != noOwner && hasReady(owner, at)) {
			return at
		}
	}
	return -1
}

// List entries' marks.
const (
	countHidden   = 255
	nameRune      = '*'
	nameSpell     = '!'
	nameMoonstone = '('
	runeGlyph     = 28
	spellGlyph    = 29
	ringNames     = 40
	plainMarker   = ' '
	listMarker    = '-'
	altFontBelow  = 32
)

// printListEntry prints an entry of a list: its count, marker and name.
// C: ZSTATS_05e2
func printListEntry(at int, counts []byte, names []string, marker int) {
	if count := counts[at]; count != countHidden {
		if count != 0 {
			printNumber(int(count), 2, ' ')
		} else {
			printMessage(msgChar45Char45)
		}
		if marker < altFontBelow {
			selectFont(1)
		}
		printChar(marker)
		selectFont(0)
	}
	name := names[at]
	switch {
	case len(name) > 0 && name[0] == nameRune:
		selectFont(1)
		printMessage(msgChar28SpaceChar43Space)
		printString(name[1:])
		selectFont(0)
	case len(name) > 0 && name[0] == nameSpell:
		selectFont(1)
		printMessage(msgChar29SpaceChar43Space)
		selectFont(0)
		printString(armsNames[ringNames+at])
	case len(name) > 0 && name[0] == nameMoonstone:
		printMessage(msgMoonstoneSp)
		selectFont(1)
		printChar(int(name[1]))
		selectFont(0)
	default:
		printString(name)
	}
	if x := textX(); x < frameRightColumn {
		for range frameRightColumn - x {
			printChar(' ')
		}
	}
	printChar('\n')
}

// Arrows shown in the middle bar for a list that scrolls.
const (
	arrowUpDown = 18
	arrowUp     = 24
	arrowDown   = 25
	pageRows    = 7
	listRows    = 8
)

// showScrollArrows shows which ways a list can scroll.
func showScrollArrows(more int) {
	switch more {
	case 0:
		clearMiddleBar()
	case 1:
		showMiddleCaption(arrowDown)
	case 2:
		showMiddleCaption(arrowUp)
	case 3:
		showMiddleCaption(arrowUpDown)
	}
}

// browseList shows a list of owned things until a key leaves it; it returns the key.
// C: ZSTATS_06e8
func browseList(caption int, end int, counts []byte, names []string) int {
	showCaption(caption)
	drawListFrame(listRows)
	top := nextListed(-1, end, counts, noOwner)
	if top == -1 {
		drawListFrame(listRows)
		setTextPosition(1, noneReadyRow)
		printMessage(msgNoneOwned)
		clearMiddleBar()
		selectWindow(windowScroll)
		return getChar()
	}
	key := 0
	for done := false; !done; {
		selectWindow(panelWindow)
		setTextPosition(1, 1)
		last := top
		for ; last != -1; last = nextListed(last, end, counts, noOwner) {
			printListEntry(last, counts, names, listMarker)
			setTextPosition(1, textY())
			if textY() == panelLastRow {
				break
			}
		}
		more := 0
		if previousListed(top, counts, noOwner) != -1 {
			more = 2
		}
		if last != -1 && nextListed(last, end, counts, noOwner) != -1 {
			more++
		}
		showScrollArrows(more)
		selectWindow(windowScroll)
		key = getChar()
		switch key {
		case keyUp, keyPageUp:
			steps := 1
			if key == keyPageUp {
				steps = pageRows
			}
			for range steps {
				if previousListed(top, counts, noOwner) != -1 {
					top = previousListed(top, counts, noOwner)
				}
			}
		case keyDown, keyPageDown:
			steps := 1
			if key == keyPageDown {
				steps = pageRows
			}
			for range steps {
				if last != -1 && nextListed(last, end, counts, noOwner) != -1 {
					top = nextListed(top, end, counts, noOwner)
					last = nextListed(last, end, counts, noOwner)
				}
			}
		case keyHome:
			top = nextListed(-1, end, counts, noOwner)
		case keyEnd:
			top = previousListed(end, counts, noOwner)
			// the decompile steps to the bottom row's entry; the binary steps up the list
			for i := 0; i < pageRows-1; i++ {
				previous := previousListed(top, counts, noOwner)
				if previous == -1 {
					break
				}
				top = previous
			}
		case 0:
		case keyLeft, keyRight, keyEscape, keySpace, '0', '1', '2', '3', '4', '5', '6':
			done = true
		}
	}
	return key
}

// Places in the owned-items list.
const (
	ownedScrolls    = 0
	ownedPotions    = 8
	ownedCarpets    = 16
	ownedSkullKeys  = 17
	ownedAmulet     = 18
	ownedCrown      = 19
	ownedSceptre    = 20
	ownedMoonstones = 21
	ownedShards     = 29
	ownedSpyglasses = 32
	ownedCapePlans  = 33
	ownedSextants   = 34
	ownedWatch      = 35
	ownedBadge      = 36
	ownedBox        = 37
)

// gatherItems fills the owned-items list from the save.
// C: ZSTATS_099a
func gatherItems() {
	for i := range moonstones {
		ownedItems[ownedScrolls+i] = scrolls[i]
		ownedItems[ownedPotions+i] = potions[i]
		if moonstoneState[i] == owned {
			ownedItems[ownedMoonstones+i] = owned
		} else {
			ownedItems[ownedMoonstones+i] = 0
		}
	}
	ownedItems[ownedCarpets] = magicCarpets
	ownedItems[ownedSkullKeys] = skullKeys
	ownedItems[ownedAmulet] = amulet
	ownedItems[ownedCrown] = crown
	ownedItems[ownedSceptre] = sceptre
	for i := range shardCount {
		ownedItems[ownedShards+i] = shards[i]
	}
	ownedItems[ownedSpyglasses] = spyglasses
	if hmsCapePlans != 0 {
		ownedItems[ownedCapePlans] = owned
	} else {
		ownedItems[ownedCapePlans] = 0
	}
	ownedItems[ownedSextants] = sextants
	ownedItems[ownedWatch] = pocketWatch
	ownedItems[ownedBadge] = blackBadge
	ownedItems[ownedBox] = sandalwoodBox
}

// Pages of the Z-stats display: two for each member, then these.
const (
	pageSupplies  = 12
	pageReagents  = 13
	pageSpells    = 14
	pageItems     = 15
	pageArms      = 16
	wholePartyRow = 6
)

// zstatsCommand shows the party's statistics page by page.
// C: ZSTATS_0a3a_ZstatsCmd
func zstatsCommand() {
	who := choosePlayer(true)
	if who == wholeParty {
		who = wholePartyRow
	} else if who < 0 {
		return
	}
	printMessage(msgStatusSp)
	gatherItems()
	clearIconRow()
	page := who << 1
	for key := 0; key != keySpace && key != keyEscape; {
		switch key {
		case keyUp, keyLeft:
			switch {
			case page == pageSupplies:
				page = int(partySize)*2 - 1
			case page > 0:
				page--
			default:
				page = pageArms
			}
		case keyRight, keyDown:
			switch {
			case int(partySize)*2-1 == page:
				page = pageSupplies
			case page < pageArms:
				page++
			default:
				page = 0
			}
		case '1', '2', '3', '4', '5', '6':
			if key-'1' < int(partySize) {
				page = (key - '1') * 2
			}
		}
		selectWindow(panelWindow)
		switch {
		case page < pageSupplies:
			if page&1 != 0 {
				showArms(page >> 1)
			} else {
				showCharacter(page >> 1)
			}
			selectWindow(windowScroll)
			key = getChar()
		case page == pageSupplies:
			showSupplies()
			selectWindow(windowScroll)
			key = getChar()
		case page == pageReagents:
			key = browseList(msgReagents2, reagentTypes, reagents[:], reagentNames)
		case page == pageSpells:
			key = browseList(msgSpells, spellCount, spellMixtures[:], spellNames)
		case page == pageItems:
			key = browseList(msgItems2, ownedItemKinds, ownedItems[:], ownedItemNames)
		case page == pageArms:
			key = browseList(msgArmaments, itemTypes, equipment[:], armsNames)
		}
	}
	clearMiddleBar()
	clearCaption()
	clearIconBar()
	showStatus()
	printMessage(msgDoneLine3)
}

// refuseArm explains why an arm cannot be readied.
// C: ZSTATS_0bee
func refuseArm(reason int) {
	printMessage(msgNewlines2)
	printMessage(reason)
	printMessage(msgItemSp2)
}

// Results of freeHand.
const (
	firstHand = 0
	otherHand = 1
	bothHands = 2
	noHand    = 255
)

// freeHand tells which hand of member row is free.
// C: ZSTATS_0c0a
func freeHand(row int) int {
	m := &party[row]
	switch {
	case m.equip[slotLeftHand] == noEquipment && m.equip[slotRightHand] == noEquipment:
		return bothHands
	case m.equip[slotLeftHand] == noEquipment:
		return firstHand
	case m.equip[slotRightHand] == noEquipment && armsSlots[m.equip[slotLeftHand]] != slotTwoHanded:
		return otherHand
	}
	return noHand
}

// Where an arm is worn (armsSlots).
const (
	slotHelmKind  = 128
	slotArmour    = 64
	slotOneHanded = 32
	slotTwoHanded = 48
	slotAmuletK   = 4
	slotRingK     = 2
)

// Arms with rules of their own.
const (
	firstArmour   = 9
	lastArmour    = 15
	slingItem     = 26
	crossbowItem  = 28
	magicAxeItem  = 36
	vanishOdds    = 15
	invisibleTile = actor29
)

// readyArm readies or (when ready) removes an arm for member row; it tells whether the
// list must close.
// C: ZSTATS_0c5c_EquipItem
func readyArm(row, item int) bool {
	var me *actor
	if mapID >= firstMapCombat {
		me = &actors[combatants[combatant].actorSlot]
	}
	if item == arrowsItem || item == quarrelsItem {
		return false
	}
	if item >= firstArmour && item <= lastArmour && mapID >= firstMapCombat && battleWon == 0 {
		refuseArm(msgThouCanstNotChangeArmourIn)
		return false
	}
	if hasReady(row, item) {
		unequip(row, item)
		if equipment[item] < maxItem {
			equipment[item]++
		}
		if mapID < firstMapCombat || item != invisibilityRing {
			return false
		}
		class := indexOf(classLetters, party[row].class)
		me.tile = classTiles[class]
		me.animTile = me.tile
		return false
	}
	if ((item == slingItem || item == magicAxeItem) && equipment[arrowsItem] == 0) ||
		(item == crossbowItem && equipment[quarrelsItem] == 0) {
		refuseArm(msgThouHastNoAmmunitionForThat)
		return false
	}
	load := 0
	for _, e := range party[row].equip {
		if e != noEquipment {
			load += int(armsWeights[e])
		}
	}
	strong := int(armsWeights[item])+load <= int(party[row].str)
	m := &party[row]
	var slot *byte
	switch armsSlots[item] {
	case slotHelmKind:
		if m.equip[slotHead] != noEquipment {
			refuseArm(msgRemoveFirstThyPresentHelm)
			return false
		}
		slot = &m.equip[slotHead]
	case slotArmour:
		if m.equip[slotBody] != noEquipment {
			refuseArm(msgThouMustFirstRemoveThineOther)
			return false
		}
		slot = &m.equip[slotBody]
	case slotOneHanded:
		hand := freeHand(row)
		if hand == noHand {
			refuseArm(msgThouMustFreeOneOfThy)
			return false
		}
		if hand == bothHands {
			hand = firstHand
		}
		slot = &m.equip[slotLeftHand+hand]
	case slotTwoHanded:
		if freeHand(row) != bothHands {
			refuseArm(msgBothHandsMustBeFreeBefore)
			return false
		}
		slot = &m.equip[slotLeftHand]
	case slotAmuletK:
		if m.equip[slotAmulet] != noEquipment {
			refuseArm(msgThouMustRemoveThineOtherAmulet)
			return false
		}
		slot = &m.equip[slotAmulet]
	case slotRingK:
		if m.equip[slotRing] != noEquipment {
			refuseArm(msgOnlyOneMagicRingMayBe)
			return false
		}
		slot = &m.equip[slotRing]
	}
	if !strong {
		refuseArm(msgThouArtNotStrongEnough)
		return false
	}
	*slot = byte(item)
	equipment[item]--
	// the decompile tests the count left; the binary tests the ring
	if (item == invisibilityRing || item == regenerationRing) && randomRange(0, vanishOdds) == 0 {
		printMessage(msgRingVanishesLine)
		m.equip[slotRing] = noEquipment
		sweepTone(leaveFrom, leaveTo, 1, leaveTime)
		return true
	}
	if item != invisibilityRing || mapID < firstMapCombat {
		return false
	}
	me.tile, me.animTile = invisibleTile, invisibleTile
	return false
}

// Modes of chooseFromList.
const (
	listReady = 'R'
	listUse   = 'U'
)

// chooseFromList lets the player move through arms (to ready) or items (to use), the
// first shown at top; it returns the entry chosen (-1 none).
// C: ZSTATS_0f2e
func chooseFromList(top, owner, mode int) int {
	chosen := 0
	done := false
	cursor := 1
	counts, end := ownedItems[:], ownedItemKinds
	if mode == listReady {
		counts, end = equipment[:], itemTypes
	}
	for !done {
		selectWindow(panelWindow)
		setTextPosition(1, 1)
		shown := 0
		last := top
		for last != -1 {
			shown++
			if textY() == cursor {
				printChar(charToggleInverse)
				chosen = last
			}
			if mode == listReady {
				marker := int(plainMarker)
				if hasReady(owner, last) {
					marker = int(armsGlyphs[last])
				}
				printListEntry(last, counts, armsNames, marker)
			} else {
				printListEntry(last, counts, ownedItemNames, plainMarker)
			}
			setTextPosition(1, textY())
			if textY()-1 == cursor {
				printChar(charToggleInverse)
			}
			if textY() == listRows {
				break
			}
			last = nextListed(last, end, counts, owner)
		}
		more := 0
		if previousListed(top, counts, owner) != -1 {
			more = 2
		}
		if last != -1 && nextListed(last, end, counts, owner) != -1 {
			more++
		}
		showScrollArrows(more)
		selectWindow(windowScroll)
		key := getChar()
		switch key {
		case keyLeft, keyUp, keyPageUp:
			steps := 1
			if key == keyPageUp {
				steps = pageRows
			}
			for range steps {
				if cursor == 4 || cursor == 1 {
					if previousListed(top, counts, owner) != -1 {
						top = previousListed(top, counts, owner)
					} else if cursor == 4 {
						cursor--
					}
				} else {
					cursor--
				}
			}
		case keyRight, keyDown, keyPageDown:
			steps := 1
			if key == keyPageDown {
				steps = pageRows
			}
			for range steps {
				if cursor == 4 || cursor == pageRows || cursor >= shown {
					if last != -1 && nextListed(last, end, counts, owner) != -1 {
						top = nextListed(top, end, counts, owner)
						last = nextListed(last, end, counts, owner)
					} else if cursor == 4 && shown > 4 {
						cursor++
					}
				} else {
					cursor++
				}
			}
		case keyEnter, keySpace:
			if mode == listReady {
				done = readyArm(owner, chosen)
			} else {
				done = true
			}
		case keyEscape:
			if mode == listReady {
				printMessage(msgDoneLine3)
			} else {
				printMessage(msgNoneLine)
			}
			done = true
			chosen = -1
		case keyHome:
			top = nextListed(-1, end, counts, owner)
			cursor = 1
		case keyEnd:
			top = previousListed(end, counts, owner)
			for cursor = 1; cursor < pageRows; cursor++ {
				chosen = previousListed(top, counts, owner)
				if chosen == -1 {
					break
				}
				top = chosen
			}
		}
	}
	return chosen
}

// readyCommand readies or removes a member's arms.
// C: ZSTATS_1296_ReadyCmd
func readyCommand() {
	who := choosePlayer(false)
	if who < 0 {
		return
	}
	first := nextListed(-1, itemTypes, equipment[:], who)
	if first == -1 {
		printMessage(msgThouArtEmptyHandedLine)
		return
	}
	printMessage(msgItemSp)
	clearIconRow()
	selectWindow(panelWindow)
	showCaptionText(party[who].nameString())
	drawListFrame(listRows)
	chooseFromList(first, who, listReady)
	clearMiddleBar()
	clearCaption()
	clearIconBar()
	showStatus()
}
