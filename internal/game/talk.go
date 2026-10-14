package game

// Talking (talk.c): the Talk command, the guards, beggars and merchants, and the
// interpreter of the conversation scripts in the .TLK files.
//
// A script is a run of zero-terminated strings: the name, the description, the greeting,
// the job, the farewell, then keyword and answer pairs, then labelled questions. Letters
// carry their high bit (clear for runes); bytes below 129 are words from a dictionary, and
// bytes from 129 are commands.

// Addresses of the conversation tables in the data segment.
const (
	shopNamesAddress   = 8650  // C: D_21ca (the shops' names, by kind and place)
	shopOwnersAddress  = 8906  // C: D_22ca
	shopPlacesAddress  = 9162  // C: D_23ca (the settlement of each shop)
	dictionaryAddress  = 9450  // C: D_24ea (the words of the scripts)
	talkFilesAddress   = 19104 // C: D_4aa0
	talkKeywordAddress = 19112 // C: D_4aa8 (the keywords everyone answers)
)

const (
	shopKinds      = 8
	shopsOfKind    = 16
	dictionarySize = 128
	keywordCount   = 34
)

// Conversation tables read from the data segment.
var (
	shopNames   [shopKinds][shopsOfKind]string // C: D_21ca
	shopOwners  [shopKinds][shopsOfKind]string // C: D_22ca
	shopPlaces  [shopKinds][]byte              // C: D_23ca
	dictionary  []string                       // C: D_24ea
	talkFiles   []string                       // C: D_4aa0
	commonWords []string                       // C: D_4aa8
)

// loadTalkTables reads the conversation tables.
func loadTalkTables() {
	for kind := range shopKinds {
		for i := range shopsOfKind {
			at := (kind*shopsOfKind + i) * 2
			if p := segment.Word(shopNamesAddress + at); p != 0 {
				shopNames[kind][i] = message(int(p))
			}
			if p := segment.Word(shopOwnersAddress + at); p != 0 {
				shopOwners[kind][i] = message(int(p))
			}
		}
		shopPlaces[kind] = segment.Bytes(shopPlacesAddress+kind*shopsOfKind, shopsOfKind)
	}
	dictionary = make([]string, dictionarySize)
	for i := range dictionary {
		dictionary[i] = message(int(segment.Word(dictionaryAddress + i*2)))
	}
	talkFiles = make([]string, settlementKinds)
	for i := range talkFiles {
		talkFiles[i] = message(int(segment.Word(talkFilesAddress + i*2)))
	}
	commonWords = make([]string, keywordCount)
	for i := range commonWords {
		commonWords[i] = message(int(segment.Word(talkKeywordAddress + i*2)))
	}
}

// Script bytes.
const (
	scriptLetter     = 128 // the high bit of a letter in the normal font
	scriptSpace      = scriptLetter | ' '
	scriptNewline    = scriptLetter | '\n'
	scriptReturn     = scriptLetter | '\r'
	scriptQuote      = scriptLetter | '"'
	scriptAvatarName = 129
	scriptEnd        = 130
	scriptPause      = 131
	scriptJoin       = 132
	scriptPay        = 133
	scriptGive       = 134
	scriptOr         = 135
	scriptAskName    = 136
	scriptKarmaUp    = 137
	scriptKarmaDown  = 138
	scriptCallGuards = 139
	scriptIfKnown    = 140
	scriptRunes      = 142
	scriptWait       = 143
	scriptLabelEnd   = 144 // ends the keyword answers and starts the labels
	scriptLabelFirst = 145
	scriptLabelLast  = 159
	scriptKarmaCheck = 254
	scriptNoLabel    = 255
)

// The fixed strings of a script.
const (
	talkName     = 0
	talkLook     = 1
	talkGreeting = 2
	talkJob      = 3
	talkBye      = 4
	talkKeywords = 5
)

// Conversation state.
var (
	talkPending    byte                              // C: D_4aee (a command waiting for its arguments)
	talkArgCount   byte                              // C: D_4aef
	talkRuneMask   byte                = lowByteMask // C: D_4af0 (clears letters' high bit in runes)
	talkWordLen    byte                              // C: D_4af1
	talkAnswering  byte                              // C: D_4af2 (answering a question, not asked one)
	talkLastX      byte                              // C: D_4af3
	talkLastChar   byte                              // C: D_4af4
	talkAfterWord  byte                              // C: D_4af5
	talkArgs       [3]byte                           // C: D_bce0
	talkWord       [16]byte                          // C: D_bce4
	talkWho        int                               // C: D_bcdc
	talkAt         int                               // C: D_bcde (the reading place in the script)
	talkLabel      byte                              // C: D_bcf4
	talkTopic      int                               // C: D_bcf6
	talkInput      [talkInputSize]byte               // C: D_bcf8
	shopKind       int                               // C: D_b116
	shopIndex      int                               // C: D_b114
	shopName       string                            // C: D_aafc
	shopOwner      string                            // C: D_aafe
	talkPauseTicks = 28
)

const (
	talkInputSize  = 16
	talkInputLen   = 15
	passwordLen    = 14
	passwordPrefix = 4
	nameCompare    = 3
	askedNameLen   = 4
	talkWrapColumn = 18
	textMask       = 127
	tributeEach    = 10
	maxKarma       = 99
	maxItem        = 99
	maxGold        = 9999
	beggarTurns    = 100
	horseDealer    = 131
	firstShopMood  = 129
	charityTown    = 5
)

// script returns the conversation buffer.
func script() []byte { return scratch[:] }

// equalNoCase compares two zero-terminated strings, ignoring case and high bits.
// C: TALK_0000_CompareStringNoCase
func equalNoCase(a, b []byte) bool {
	for i := 0; ; i++ {
		if toUpper(int(a[i]&textMask)) != toUpper(int(b[i]&textMask)) {
			return false
		}
		if a[i+1] == 0 && b[i+1] == 0 {
			return true
		}
	}
}

// behindCounter tells whether the tile at x, y is a counter one can talk across.
// C: TALK_0054
func behindCounter(x, y int) bool {
	switch *tileAt(x, y) {
	case mapTile41, mapTileTable148, mapTileTable149, mapTileTable150, mapTile151, mapTile152, mapTile153,
		mapTileTable154, mapTileTable155, mapTileTable156, mapTileDesk, mapTile174, mapTileDoor186,
		mapTileDoor187, mapTileTable190, mapTile202, mapTile203:
		return true
	}
	return false
}

// askPay asks whether the party pays; it tells whether it refused.
// C: TALK_00ac_AskPay
func askPay() bool {
	printMessage(msgDostThouPay)
	for {
		c := toUpper(getChar())
		if c == 'Y' {
			printMessage(msgYesLine)
			return false
		}
		if c == 'N' {
			break
		}
	}
	printMessage(msgNoLine4)
	return true
}

// talkToMerchant opens a shop of kind (a merchant's mood) for the active member.
// C: TALK_00e6_TalkToMerchant
func talkToMerchant(kind int) {
	if partyTile&pairMask == actorRidingHorse && kind != horseDealer {
		printMessage(msgAMerchantSaysGetThatHorse)
		return
	}
	buyer := int(activeMember)
	if activeMember == noMember {
		firstAbleMember()
		buyer = int(moveX)
	}
	shopKind = kind - firstShopMood
	// fix: the original reads one past the list before testing its end.
	for shopIndex = 0; shopIndex < shopsOfKind-1 && shopPlaces[shopKind][shopIndex] != mapID; shopIndex++ {
	}
	shopName = shopNames[shopKind][shopIndex]
	shopOwner = shopOwners[shopKind][shopIndex]
	switch shopKind {
	case 0:
		armoury(buyer)
	case 1:
		tavern(buyer)
	case 2:
		stables(buyer)
	case 3:
		shipwright(buyer)
	case 4:
		herbalist(buyer)
	case 5:
		guild(buyer)
	case 6:
		healer(buyer)
	case 7:
		inn(buyer)
	}
}

// guardToll has a guard ask the party's toll or password; it tells whether the party
// refused (and is arrested).
// C: TALK_01e2
func guardToll() int {
	if mapID == mapPalace {
		if spellIcon == badgeIcon {
			printChar('"')
			printMessage(msgGiveNowThePasswordBearerOf)
			printChar('"')
			printMessage(msgYourResponseLine)
			var answer [passwordLen + 2]byte
			copy(answer[:], readString(passwordLen))
			printChar('\n')
			answer[passwordPrefix] = 0
			if equalNoCase([]byte(message(msgImpe)+"\x00"), answer[:]) {
				printChar('\n')
				printChar('"')
				printMessage(msgPassFriend)
				printChar('"')
				printChar('\n')
				return 0
			}
		}
		return 1
	}
	if mapID == charityTown {
		printChar('"')
		printMessage(msgThouWiltGiveHalfThyGold)
		printChar('"')
		if askPay() {
			return 1
		}
		gold /= 2
		showStatus()
		return 0
	}
	tribute := 0
	for i := range int(partySize) {
		if party[i].status != statusDead {
			tribute += tributeEach
		}
	}
	printMessage(msgAGuardDemandsASp)
	printNumber(tribute, 2, ' ')
	printMessage(msgGpTributeToBlackthorn)
	if askPay() {
		return 1
	}
	if int(gold) >= tribute {
		gold -= int16(tribute)
		showStatus()
		return 0
	}
	return 1
}

const badgeIcon = 29 // the spell icon the black badge shows

// talkToPerson talks to person who; it tells whether a guard arrests the party.
// C: TALK_031e_TalkToNpc
func talkToPerson(who int) int {
	printChar('\n')
	talkWho = who
	n := &npcs[who]
	s := &schedules[who]
	if s.kind[n.place] == kindWait {
		s.kind[n.place] = kindWander3
	} else if actors[n.actorAt].tile == actorGuard && (n.place&1 == 0 || n.mood == 0) {
		printMessage(msgTheGuardOffersNoResponseLine)
		return 0
	}
	mood := int(n.mood)
	switch {
	case mood == 0:
		printMessage(msgNoResponseLine)
	case mood < scriptLetter:
		startTalk(mood)
	case mood == moodFleeing:
		printChar('"')
		printMessage(msgDonTHurtMePleaseGo)
		printChar('"')
		printChar('\n')
	case mood == moodHostile:
		// fix: the original passes a stale register for the person.
		beginVermin(who)
	case mood == lowByteMask:
		return guardToll()
	case n.place&1 != 0 && schedulePlace(talkWho, int(hour))&1 != 0:
		talkToMerchant(mood)
	default:
		printMessage(msgAMerchantSaysComeSeeMe)
		printMessage(msgWhenItSOpenLine)
	}
	return 0
}

// talkCommand is the Talk command.
// C: TALK_041c_TalkCmd
func talkCommand() int {
	if !selectDirection() {
		return 0
	}
	dx, dy := int(moveX), int(moveY)
	x, y := dx+int(mapX), dy+int(mapY)
	if actorAt(x, y, int(level)) == 0 && behindCounter(x, y) {
		x += dx
		y += dy
	}
	if actorAt(x, y, int(level)) == 0 {
		printMessage(msgNobodySHereLine)
		return 0
	}
	who := personShownBy(int(moveX))
	switch *tileAt(x, y) {
	case mapTileBed:
		printMessage(msgZzzzzzLine4)
	case mapTileMirror:
		printMessage(msgNoResponseLine2)
	default:
		return talkToPerson(who)
	}
	return 0
}

// newParagraph and quote send a line end and a quote mark through the script.
// C: TALK_04d2_PrintCr, TALK_04da_PrintQuoteMark
func newParagraph() { scriptChar(scriptReturn) }
func quote()        { scriptChar(scriptQuote) }

// flushWord prints the word built so far, in runes where its letters lack the high bit.
// C: TALK_04e2_PrintWord
func flushWord() {
	if talkWordLen == 0 {
		return
	}
	for i := 0; i < int(talkWordLen); i++ {
		if talkWord[i]&scriptLetter == 0 {
			selectFont(1)
		} else {
			selectFont(0)
		}
		c := talkWord[i] & textMask
		if textX() == 0 && c == ' ' {
			continue
		}
		if talkLastX < talkWrapColumn-3 || c != '\n' {
			if c != '\n' {
				talkLastX = byte(textX())
			}
			printChar(int(c))
		} else if c == '\n' {
			talkLastX = 0
		}
	}
	talkWordLen = 0
	selectFont(0)
}

// addToWord adds c to the word, printing it at a space or line end, wrapping first when it
// would not fit.
// C: TALK_0574_BuildOrPrintWord
func addToWord(c byte) {
	if int(talkWordLen) != len(talkWord) {
		talkWord[talkWordLen] = c
		talkWordLen++
		if c != scriptNewline && c != scriptSpace {
			return
		}
		if textX()+int(talkWordLen) >= talkWrapColumn {
			printChar('\n')
		}
	}
	flushWord()
}

// payGold takes the gold the script asks; short of it the talk goes on to questions.
// C: TALK_05b6
func payGold() int {
	price := int(talkArgs[0]&textMask)*100 + int(talkArgs[1]&textMask)*10 + int(talkArgs[2]&textMask) - '0'*111
	if int(gold) >= price {
		gold -= int16(price)
		showStatus()
		if actors[npcs[talkWho].actorAt].tile&shapeMask == actorBeggar && turnCount >= beggarTurns {
			turnCount = 0
			increaseByte(&karma, 1, maxKarma)
			if gold == 0 {
				increaseByte(&karma, 2, maxKarma)
			}
		}
		return 0
	}
	talkWordLen = 0
	printChar('"')
	printMessage(msgThouHastNotEnoughGold)
	printChar('"')
	printMessage(msgNewlines2)
	talkPending, talkArgCount = 0, 0
	return askTopics()
}

// Items a script can give besides equipment.
const (
	itemFood         = 65
	itemGold         = 66
	itemKey          = 67
	itemGem          = 68
	itemTorch        = 69
	itemGrapple      = 70
	itemCarpet       = 71
	itemSextant      = 72
	itemSpyglass     = 73
	itemBlackBadge   = 74
	itemSkullKey     = 75
	firstSpecialItem = 64
	maxFood          = 9999
)

// giveItem gives the party one of item.
// C: TALK_0682
func giveItem(item byte) {
	if item < firstSpecialItem {
		// the original indexes past the equipment into the spell mixtures
		if int(item) < itemTypes {
			increaseByte(&equipment[item], 1, maxItem)
		} else {
			increaseByte(&spellMixtures[int(item)-itemTypes], 1, maxItem)
		}
		return
	}
	switch item {
	case itemFood:
		increaseInt(&food, 1, maxFood)
		showStatus()
	case itemGold:
		increaseInt(&gold, 1, maxGold)
		showStatus()
	case itemKey:
		increaseByte(&keys, 1, maxItem)
	case itemGem:
		increaseByte(&gems, 1, maxItem)
	case itemTorch:
		increaseByte(&torches, 1, maxItem)
	case itemGrapple:
		increaseByte(&grapple, 1, maxItem)
	case itemCarpet:
		increaseByte(&magicCarpets, 1, maxItem)
	case itemSextant:
		sextants = lowByteMask
	case itemSpyglass:
		spyglasses = lowByteMask
	case itemBlackBadge:
		blackBadge = lowByteMask
	case itemSkullKey:
		increaseByte(&skullKeys, 1, maxItem)
	}
}

// skipTo moves past the next byte that is end (true) or stop (false).
// C: TALK_0728_SkipOneTalkString
func skipTo(end, stop byte) bool {
	buf := script()
	for {
		c := buf[talkAt]
		talkAt++
		if c == end {
			return true
		}
		if c == stop {
			return false
		}
	}
}

// selectString moves to the nth string of the script.
// C: TALK_075a_SelectNthTalkString
func selectString(n int) {
	talkAt = 0
	for ; n != 0; n-- {
		skipTo(0, scriptLabelEnd)
	}
}

// runString runs the script from the reading place to the string's end; it tells whether
// the talk ended.
// C: TALK_0788_ProcessTalkString
func runString() int {
	buf := script()
	for buf[talkAt] != 0 {
		c := buf[talkAt]
		talkAt++
		if scriptChar(c) != 0 {
			return 1
		}
	}
	return 0
}

// runNthString runs the nth string.
// C: TALK_07aa_ProcessNthTalkString
func runNthString(n int) int {
	selectString(n)
	return runString()
}

// runAlternative runs the string after the current one (an "or" answer).
// C: TALK_07be
func runAlternative() int {
	buf := script()
	talkAt++
	for buf[talkAt] == 0 {
		talkAt++
	}
	for buf[talkAt] != 0 {
		talkAt++
	}
	talkAt++
	return runString()
}

// sayAvatarName prints the Avatar's name through the script.
// C: TALK_07e4_PrintAvatarName
func sayAvatarName() {
	for _, c := range party[0].name {
		if c == 0 {
			break
		}
		scriptChar(c | scriptLetter)
	}
}

const reserveSlots = 15

// joinParty has the person join the party.
// C: TALK_080a_Join
func joinParty() int {
	saved := talkAt
	if partySize == maxPartySize {
		getChar()
		printMessage(msgThouHastNoRoomForMe)
		printMessage(msgSeekMeAgainIfOneOf)
		return 0
	}
	selectString(talkName)
	var wanted, name [nameCompare + 1]byte
	copy(wanted[:nameCompare], script()[talkAt:])
	talkAt += nameCompare
	slot := reserveSlots
	for {
		copy(name[:nameCompare], party[slot].name[:])
		if equalNoCase(wanted[:], name[:]) {
			break
		}
		slot--
		if slot == 0 {
			addToWord('"')
			addToWord('\n')
			printMessage(msgSystemErrorNoMatch)
			talkAt = saved
			return 1
		}
	}
	party[slot].location = 0
	party[slot], party[partySize] = party[partySize], party[slot]
	partySize++
	markKilled(talkWho)
	removePerson(talkWho)
	showStatus()
	flushWord()
	talkAt = saved
	return 1
}

// runLabel runs the label's text.
// C: TALK_093a
func runLabel() int {
	findLabel()
	talkAt++
	return runString()
}

// findLabel moves to the current label.
func findLabel() {
	talkAt = 0
	for script()[talkAt] != talkLabel {
		skipTo(scriptLabelEnd, scriptLabelLast)
	}
}

// runNextString runs the string after the next end.
// C: TALK_0960
func runNextString() int {
	skipTo(0, scriptLabelEnd)
	return runString()
}

// runLabelOtherwise runs the label's answer for any other reply.
// C: TALK_096e
func runLabelOtherwise() int {
	findLabel()
	skipTo(0, scriptLabelLast)
	return runString()
}

// selectKeyword moves to the keyword of topic n; it tells whether there is one.
// C: TALK_099a
func selectKeyword(n int) bool {
	talkAt = 0
	for n = n*2 + talkKeywords; n != 0; n-- {
		if !skipTo(0, scriptLabelEnd) {
			return false
		}
	}
	return true
}

// findTopic finds the topic the player asked of; it tells whether there is one.
// C: TALK_09d8
func findTopic() bool {
	for talkTopic = 0; ; talkTopic++ {
		if !selectKeyword(talkTopic) {
			return false
		}
		if wordAsked(script()[talkAt:]) {
			return true
		}
		talkAt++
	}
}

// wordAsked tells whether word starts a word of the player's input.
func wordAsked(word []byte) bool {
	at := findWord(word, talkInput[:])
	return at != -1 && (at == 0 || talkInput[at-1] == ' ')
}

// readInput reads the player's words.
// C: TALK_0a2c_GetString
func readInput() {
	defer asking(promptWords)()
	flushWord()
	talkInput = [talkInputSize]byte{}
	copy(talkInput[:talkInputLen], readString(talkInputLen))
}

// sayBye says the farewell.
// C: TALK_0a3c
func sayBye() int {
	quote()
	if runNthString(talkBye) == 0 {
		quote()
		newParagraph()
	}
	return 1
}

// answerKeyword answers one of the keywords everyone knows: 0 answered, 1 talk over, 2
// not now.
// C: TALK_0a54
func answerKeyword(keyword int) int {
	switch keyword {
	case 0:
		if talkAnswering != 0 {
			return 2
		}
		printMessage(msgMyNameIsSp)
		if runNthString(talkName) != 0 {
			return 1
		}
	case 1, 2:
		if talkAnswering != 0 {
			return 2
		}
		quote()
		if runNthString(talkJob) != 0 {
			return 1
		}
	case 3, 4:
		if talkAnswering != 0 {
			return 2
		}
		return sayBye()
	default:
		printMessage(msgWithLanguageLikeThatHowDid)
		quote()
		newParagraph()
		newParagraph()
		talkPause()
		return 0
	}
	quote()
	newParagraph()
	newParagraph()
	return 0
}

// talkPause waits a moment or for a key.
func talkPause() {
	for range talkPauseTicks {
		updateFrame()
		if pollKey() != 0 {
			break
		}
		waitTicks(1)
	}
	clearKeyboard()
}

// matchKeywords answers the keywords in the input; it returns answerKeyword's result
// (-1 when none matched).
func matchKeywords() int {
	result := -1
	for k := range keywordCount {
		if !wordAsked([]byte(commonWords[k] + "\x00")) {
			continue
		}
		result = answerKeyword(k)
		if result == 0 || result == 1 {
			return result
		}
	}
	return result
}

// askTopics asks the player's interest until the talk ends.
// C: TALK_0b04
func askTopics() int {
	for {
		talkAnswering = 0
		printMessage(msgYourInterest)
		readInput()
		if talkInput[0] == 0 {
			printMessage(msgByeLine)
			return sayBye()
		}
		newParagraph()
		newParagraph()
		result := matchKeywords()
		if result == 1 {
			return 1
		}
		if result == 0 {
			continue
		}
		if !findTopic() {
			printMessage(msgICannotHelpTheeWithThat)
			quote()
			newParagraph()
			newParagraph()
			continue
		}
		quote()
		if runNthString(talkTopic*2+talkKeywords+1) != 0 {
			return 1
		}
		quote()
		newParagraph()
		newParagraph()
	}
}

// findAnswer looks for the player's reply among the label's answers.
// C: TALK_0bd4
func findAnswer() bool {
	findLabel()
	skipTo(0, scriptLabelLast)
	skipTo(0, scriptLabelLast)
	for {
		if wordAsked(script()[talkAt:]) {
			return true
		}
		if !skipTo(0, scriptLabelEnd) || !skipTo(0, scriptLabelEnd) {
			return false
		}
	}
}

// askLabel runs a label: its question, the player's reply and the answer.
// C: TALK_0c5c
func askLabel() int {
	result := 0
	for result == 0 {
		quote()
		if runLabel() != 0 {
			return 1
		}
		quote()
		for {
			newParagraph()
			newParagraph()
			talkAnswering = lowByteMask
			printMessage(msgYouRespond)
			readInput()
			if talkInput[0] != 0 {
				break
			}
			printMessage(msgWhatDidstThouSay)
		}
		newParagraph()
		newParagraph()
		result = matchKeywords()
		if result == 1 {
			return 1
		}
	}
	if !findAnswer() {
		quote()
		result = runLabelOtherwise()
	} else {
		quote()
		result = runNextString()
	}
	if result != 0 {
		return 1
	}
	quote()
	newParagraph()
	newParagraph()
	return askTopics()
}

// metBits returns this settlement's people the party has been introduced to.
func metBits() uint32 {
	at := (int(mapID) - 1) * 4
	return uint32(npcMet[at]) | uint32(npcMet[at+1])<<8 | uint32(npcMet[at+2])<<16 | uint32(npcMet[at+3])<<24
}

// markMet records that the party knows person who.
// C: TALK_0d42
func markMet(who int) {
	bits := metBits() | 1<<who
	at := (int(mapID) - 1) * 4
	for i := range 4 {
		npcMet[at+i] = byte(bits >> (8 * i))
	}
}

// hasMet tells whether the party knows person who.
// C: TALK_0d7a
func hasMet(who int) bool { return metBits()&(1<<who) != 0 }

// scriptArgument takes an argument of the waiting command.
// C: TALK_0dbe
func scriptArgument(c byte) int {
	switch talkPending {
	case scriptPay:
		talkArgs[talkArgCount] = c
		talkArgCount++
		if talkArgCount != 3 {
			return 0
		}
		if payGold() != 0 {
			return 1
		}
	case scriptGive:
		giveItem(c & textMask)
	case scriptIfKnown:
		talkPending, talkArgCount = 0, 0
		if !hasMet(talkWho) {
			return 0
		}
		if c == scriptNoLabel {
			return askTopics()
		}
		talkLabel = c
		return askLabel()
	case scriptKarmaCheck:
		talkArgs[talkArgCount] = c
		talkArgCount++
		if talkArgCount != 2 {
			return 0
		}
		talkPending, talkArgCount = 0, 0
		if karma < talkArgs[0] {
			return 0
		}
		talkLabel = talkArgs[1]
		return askLabel()
	}
	talkPending, talkArgCount = 0, 0
	return 0
}

// askName asks the player's name; the right name makes the person know the party.
// C: TALK_0e78_AskName
func askName() {
	quote()
	flushWord()
	printMessage(msgWhatIsThyNameLine)
	printMessage(msgYouRespond2)
	readInput()
	talkPending, talkArgCount = 0, 0
	if talkInput[0] == 0 {
		printMessage(msgIfYouSaySo)
		return
	}
	for i := range int(partySize) {
		var name [askedNameLen + 1]byte
		copy(name[:askedNameLen], party[i].name[:])
		if wordAsked(name[:]) {
			markMet(talkWho)
			printMessage(msgAPleasure)
			return
		}
	}
	printMessage(msgIfYouSaySo)
}

// scriptChar runs one script byte; it tells whether the talk ended.
// C: TALK_0f32_ProcessChar
func scriptChar(c byte) int {
	if talkPending != 0 {
		return scriptArgument(c)
	}
	talkArgCount = 0
	if c == scriptQuote && talkLastChar == scriptQuote {
		return 0
	}
	switch {
	case c == scriptAvatarName:
		sayAvatarName()
	case c == scriptEnd:
		return 1
	case c == scriptPause:
		talkPause()
	case c == scriptJoin:
		return joinParty()
	case c == scriptOr:
		saved := talkAt
		if runAlternative() != 0 {
			return 1
		}
		talkAt = saved
	case c == scriptAskName:
		askName()
	case c == scriptKarmaUp:
		increaseByte(&karma, 1, maxKarma)
	case c == scriptKarmaDown:
		decreaseByte(&karma, 1)
	case c == scriptCallGuards:
		raiseAlarm()
	case c == scriptRunes:
		talkRuneMask ^= scriptLetter
	case c == scriptNoLabel:
		flushWord()
		return askTopics()
	case c == scriptWait:
		getChar()
	case c == scriptKarmaCheck || c == scriptPay || c == scriptGive || c == scriptIfKnown:
		talkPending = c
	case c >= scriptLabelFirst && c <= scriptLabelLast:
		talkLabel = c
		return askLabel()
	case c < scriptAvatarName:
		addToWord(scriptSpace)
		word := dictionary[c-1]
		for i := range len(word) {
			addToWord(word[i] | scriptLetter)
		}
		if word == "" {
			addToWord(c)
			return 0
		}
		talkAfterWord = 1
	default:
		c |= scriptLetter
		if c == scriptReturn {
			c = scriptNewline
		}
		c &= talkRuneMask
		talkLastChar = c
		if talkAfterWord != 0 {
			addToWord(scriptSpace)
		}
		addToWord(c)
		talkAfterWord = 0
	}
	return 0
}

// describePerson describes the person and, now and then, their name.
// C: TALK_111c
func describePerson() int {
	printMessage(msgYouSeeSp)
	if runNthString(talkLook) != 0 {
		return 1
	}
	newParagraph()
	newParagraph()
	if hasMet(talkWho) {
		quote()
		if runNthString(talkGreeting) != 0 {
			return 1
		}
		quote()
		newParagraph()
		newParagraph()
		return 0
	}
	seedRandom(clockSeed())
	if randomRange(0, 1) != 0 {
		printMessage(msgIAmCalledSp)
		if runNthString(talkName) != 0 {
			return 1
		}
		quote()
		newParagraph()
		newParagraph()
	}
	return 0
}

// stealSomething has the shadowlord of falsehood's thieves take a thing from the party.
// C: TALK_1180
func stealSomething() {
	if shadowlordHere != 0 {
		return
	}
	printMessage(msgSomethingWasStolenLine)
	sweepTone(theftFrom, theftTo, 1, theftTime)
	seedRandom(clockSeed())
	if keys|gems|torches != 0 {
		for {
			switch randomRange(0, 2) {
			case 0:
				if keys != 0 {
					decreaseByte(&keys, 1)
					return
				}
			case 1:
				if gems != 0 {
					decreaseByte(&gems, 1)
					return
				}
			case 2:
				if torches != 0 {
					decreaseByte(&torches, 1)
					return
				}
			}
		}
	}
	for i := itemTypes - 1; i >= 0; i-- {
		if equipment[i] != 0 {
			decreaseByte(&equipment[i], 1)
			return
		}
	}
	for i := potionTypes - 1; i >= 0; i-- {
		if potions[i] != 0 {
			decreaseByte(&potions[i], 1)
			return
		}
	}
	for i := scrollTypes - 1; i >= 0; i-- {
		if scrolls[i] != 0 {
			decreaseByte(&scrolls[i], 1)
			return
		}
	}
	decreaseInt(&gold, randomRange(1, 15))
	showStatus()
}

const (
	theftFrom = 800
	theftTo   = 2000
	theftTime = 50
	talkIndex = 512
	talkText  = 1024
)

// startTalk reads person script's conversation and holds it.
// C: TALK_127e_StartTalk
func startTalk(person int) {
	file := talkFiles[(mapID-1)>>3]
	buf := script()
	readFile(file, buf[:talkIndex], 0)
	count := int(int16(uint16(buf[0]) | uint16(buf[1])<<8))
	word := func(i int) int { return int(int16(uint16(buf[i*2]) | uint16(buf[i*2+1])<<8)) }
	at := 1
	for i := 1; ; i++ {
		id := word(at)
		at++
		if id == person || i > count {
			break
		}
		at++
	}
	offset := word(at)
	readFile(file, buf[:talkText], offset)
	if describePerson() == 0 && askTopics() == 0 {
		sayBye()
	}
	stealSomething()
}
