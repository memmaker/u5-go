package game

// Segment 3000: the command keys, the screen shake, explosions, picking a direction,
// finding actors, typing numbers, holing up and the counters' helpers.

const (
	shakeRounds    = 8
	shakeTop       = 8
	shakeBottom    = 179
	shakeStep      = 3
	shakeBand      = 4
	shakeLeft      = 8
	shakeRight     = 183
	shakeScroll    = 2
	shakeNoiseRate = 19
	shakeNoiseTop  = 150
)

// shakeScreen shakes the map view, as earthquakes and cannon fire do. The original runs
// as fast as the machine; the speaker noise between the passes keeps the u5d timing.
// C: ULTIMA_3072_ShakeScreen
func shakeScreen() {
	copyPage(0, 1)
	selectPage(1)
	for range shakeRounds {
		for y := shakeTop; y <= shakeBottom; y += shakeStep {
			scrollWindowUp(shakeLeft, y, shakeRight, y+shakeBand, shakeScroll)
		}
		whiteNoise(shakeNoiseRate, 700, shakeNoiseTop)
		for y := shakeTop; y <= shakeBottom; y += shakeStep {
			copyArea(1, 0, shakeLeft, y, shakeRight, y+shakeBand)
		}
		whiteNoise(shakeNoiseRate, 300, shakeNoiseTop)
		for y := shakeBottom; y >= shakeTop; y -= shakeStep {
			scrollWindowDown(shakeLeft, y, shakeRight, y+shakeBand, shakeScroll)
		}
		whiteNoise(shakeNoiseRate, 700, shakeNoiseTop)
		for y := shakeBottom; y >= shakeTop; y -= shakeStep {
			copyArea(1, 0, shakeLeft, y, shakeRight, y+shakeBand)
		}
		whiteNoise(shakeNoiseRate, 300, shakeNoiseTop)
	}
	selectPage(0)
}

// Results of a command.
const (
	commandNoTurn = 0 // the command took no time
	commandTurn   = 1
	commandTalked = 2
)

const lastDungeonMap = 40

// inDungeon says whether the party is in one of the eight dungeons.
func inDungeon() bool { return mapID >= firstDungeonMap && mapID <= lastDungeonMap }

// doCommand runs the command of key. It returns commandNoTurn when no time passes.
// C: ULTIMA_3178_ProcessCommand
func doCommand(key int) int {
	result := commandTurn
	switch key {
	case keyCtrlB:
		printMessage(msgBufferO)
		keyBuffering = !keyBuffering
		if keyBuffering {
			printMessage(msgFfLine)
		} else {
			printMessage(msgNLine2)
		}
		result = commandNoTurn
	case keySpace:
		if mapID == 0 && sailHeading != 0 {
			printMessage(msgSheetsInIronsLine)
			sailHeading = 0
		} else {
			printMessage(msgPassLine2)
		}
	case 'A':
		switch {
		case mapID == 0:
			result = attackOutdoors()
		case mapID < firstDungeonMap:
			result = attackInTown()
		default:
			result = attackInDungeon()
		}
	case 'B':
		printMessage(msgBoardSp)
		result = boardCommand()
	case 'C':
		printMessage(msgCastLine)
		result = castCommand()
	case 'D':
		printMessage(msgDWhatLine)
		result = commandNoTurn
	case 'E':
		if mapID == 0 {
			result = enterCommand()
			break
		}
		printMessage(msgEnterWhatLine)
	case 'F':
		printMessage(msgFire2)
		fireCommand()
	case 'G':
		if mapID < firstDungeonMap {
			printMessage(msgGet)
		}
		getCommand()
	case 'H':
		if mapID == 0 || mapID >= firstDungeonMap {
			holeUpOutside()
			break
		}
		tile := *tileAt(int(mapX), int(mapY))
		printMessage(msgHoleUpSp)
		if tile != mapTileBed {
			printMessage(msgOnlyInBedLine)
		} else {
			sleepInBed()
		}
	case 'I':
		printMessage(msgIgniteTorchLine)
		igniteTorch()
	case 'J':
		printMessage(msgJimmy)
		jimmyCommand()
	case 'K':
		switch {
		case mapID == 0:
			printMessage(msgKlimb)
			klimbOutdoors()
		case mapID < firstDungeonMap:
			result = klimbInTown()
		default:
			result = klimbInDungeon()
		}
	case 'L':
		printMessage(msgLook)
		if inDungeon() {
			printMessage(msgChar46Char46Char46Newline)
			lookInDungeon()
		} else {
			printChar('-')
			lookCommand()
		}
	case 'M':
		printMessage(msgMixReagentsLine)
		mixCommand()
	case 'N':
		printMessage(msgNewOrder2)
		newOrderCommand()
	case 'O':
		printMessage(msgOpen)
		openCommand()
	case 'P':
		if inDungeon() {
			printMessage(msgPushNotHereLine)
			result = commandNoTurn
		} else {
			printMessage(msgPush)
			pushCommand()
		}
	case 'Q':
		printMessage(msgQuit2)
		saveGame()
		result = commandNoTurn
	case 'R':
		printMessage(msgReadyLine)
		readyCommand()
	case 'S':
		if mapID < firstDungeonMap {
			printMessage(msgSearch)
		} else {
			printMessage(msgSearchLine)
		}
		searchCommand()
	case 'T':
		switch {
		case mapID == 0:
			printMessage(msgTalk2)
			if !selectDirection() {
				break
			}
			printMessage(msgFunnyNoResponseLine)
		case mapID >= firstDungeonMap:
			printMessage(msgTalkFunnyNoResponseLine)
		default:
			printMessage(msgTalk2)
			if talkCommand() != 0 {
				result = commandTalked
			}
		}
	case 'U':
		printMessage(msgUseItemLine)
		useCommand()
	case 'V':
		printMessage(msgViewAGemLine)
		if gems != 0 {
			gems--
			if mapID < firstDungeonMap {
				viewGem(int(mapX), int(mapY))
			} else {
				viewGemInDungeon()
			}
			break
		}
		printMessage(msgYouHaveNoneLine)
	case 'W':
		printMessage(msgWWhatLine)
		result = commandNoTurn
	case 'X':
		printMessage(msgXItSp)
		xitCommand()
	case 'Y':
		printMessage(msgYellSp)
		result = yellCommand()
	case 'Z':
		printMessage(msgZStatsLine)
		zstatsCommand()
	default:
		printMessage(msgWhatLine)
		result = commandNoTurn
	}
	return result
}

// explodeAt shows an explosion at map position x, y (view position in combat).
// C: ULTIMA_3522
func explodeAt(x, y int) {
	if mapID < firstMapCombat {
		x -= int(mapX) - viewCentre
		y -= int(mapY) - viewCentre
	}
	putTile(0, x, y)
	whiteNoise(10, 3000, 2000)
	updateFrame()
}

// hitActor shows a hit on actor slot (in combat, on combatant slot), flashing the status
// row of a member hit.
// C: ULTIMA_3564
func hitActor(slot int) {
	var hit *fighter
	if mapID >= firstMapCombat {
		hit = &combatants[slot]
		slot = int(hit.actorSlot)
	}
	putTile(0, int(actors[slot].x), int(actors[slot].y))
	if hit != nil && hit.flags&fighterPlayer != 0 {
		highlightMember(int(hit.entity))
		whiteNoise(40, 3000, 500)
		highlightMember(int(hit.entity))
	} else {
		whiteNoise(10, 3000, 2000)
	}
	updateFrame()
}

// selectDirection reads a direction into moveX, moveY; space and escape pass.
// C: ULTIMA_35ec_SelectDirection
func selectDirection() bool {
	defer asking(promptDir)()
	moveX, moveY = 0, 0
	var key int
	for {
		key = getChar() & lowByteMask
		if key == keySpace || key == keyUp || key == keyDown || key == keyLeft || key == keyRight {
			break
		}
	}
	switch key {
	case keyUp:
		printMessage(msgNorthLine)
		moveY--
	case keyDown:
		printMessage(msgSouthLine)
		moveY++
	case keyLeft:
		printMessage(msgWestLine)
		moveX--
	case keyRight:
		printMessage(msgEastLine)
		moveX++
	default:
		printMessage(msgPassLine2)
		return false
	}
	return true
}

// lowByteMask keeps a key's low byte, as the original's byte variable does.
const lowByteMask = 255

// actorAt returns the tile of the first actor (from slot 1) at x, y on level z; its slot
// is left in moveX, as the original does.
// C: ULTIMA_368e_FindActorTileAtPos
func actorAt(x, y, z int) int {
	for moveX = 1; moveX < actorSlots; moveX++ {
		a := &actors[moveX]
		if int(a.x) == x && int(a.y) == y && (mapID >= firstMapCombat || int(a.z) == z) {
			return int(a.tile)
		}
	}
	return 0
}

// topActorAt is actorAt searching from the last slot down, the party included.
// C: ULTIMA_3702
func topActorAt(x, y, z int) int {
	for moveX = actorSlots - 1; moveX >= 0; moveX-- {
		a := &actors[moveX]
		if int(a.x) == x && int(a.y) == y && (mapID >= firstMapCombat || int(a.z) == z) {
			return int(a.tile)
		}
	}
	return 0
}

const (
	firstObjectSlots = 24 // the actor slots objects and monsters can be dropped from
	viewEdge         = 10
)

// replaceableActor returns the first actor slot from 1 with a tile in low..high (not the
// crown), only ones out of sight when outOfSight, or 0.
// C: ULTIMA_3868
func replaceableActor(low, high int, outOfSight bool) int {
	for slot := 1; slot < firstObjectSlots; slot++ {
		a := &actors[slot]
		if int(a.tile) < low || int(a.tile) > high || a.tile == actorCrown {
			continue
		}
		if !outOfSight {
			return slot
		}
		dx := a.x - mapX + viewCentre
		dy := a.y - mapY + viewCentre
		if dx > viewEdge || dy > viewEdge {
			return slot
		}
	}
	return 0
}

// The tile ranges freeActorSlot gives up first.
const (
	actorItemsFirst    = 1
	actorItemsLast     = 15
	actorMonstersFirst = 128
	actorLast          = 255
	actorHorsesFirst   = 16
	actorHorsesLast    = 17
	actorPeopleFirst   = 48
	actorPeopleLast    = 127
)

// freeActorSlot finds a slot for a new actor: an empty one, else one holding an item, a
// monster, a horse or a person out of sight, else in sight.
// C: ULTIMA_38e4
func freeActorSlot() int {
	tries := []struct {
		low, high  int
		outOfSight bool
	}{
		{0, 0, false},
		{actorItemsFirst, actorItemsLast, true},
		{actorMonstersFirst, actorLast, true},
		{actorHorsesFirst, actorHorsesLast, true},
		{actorPeopleFirst, actorPeopleLast, true},
		{actorItemsFirst, actorItemsLast, false},
		{actorMonstersFirst, actorLast, false},
		{actorHorsesFirst, actorHorsesLast, false},
		{actorPeopleFirst, actorPeopleLast, false},
		{0, actorLast, false},
	}
	for _, t := range tries {
		if slot := replaceableActor(t.low, t.high, t.outOfSight); slot != 0 {
			return slot
		}
	}
	return 0
}

// setTile changes a settlement's map tile.
// C: ULTIMA_39cc_SetTile
func setTile(tile, x, y byte) {
	if mapID != 0 && mapID < firstDungeonMap && tile != 0 {
		*tileAt(int(x), int(y)) = tile
	}
}

// Results of firstAbleMember.
const (
	someoneAble  = 0
	onlySleepers = 1
	noneAble     = -1
)

// firstAbleMember finds the first member able to act and leaves them in moveX.
// C: ULTIMA_39fc_GetFirstActivePartyMember
func firstAbleMember() int {
	sleeping := 0
	for i := range int(partySize) {
		switch party[i].status {
		case statusGood, statusPoisoned:
			moveX = int16(i)
			return someoneAble
		case statusSleeping:
			sleeping++
		}
	}
	if sleeping != 0 {
		return onlySleepers
	}
	return noneAble
}

// random returns a number in 0..high.
// C: ULTIMA_3aae_Random
func random(high int) int { return randomRange(0, high) }

// randomDamage returns 1..30.
// C: ULTIMA_3abe
func randomDamage() int {
	damage := random(60) / 2
	if damage == 0 {
		damage++
	}
	return damage
}

// sleepFrames updates the frame each tick for ticks, when the map is shown.
// C: ULTIMA_3ae6
func sleepFrames(ticks int) {
	if mapShown == 0 {
		return
	}
	for range ticks {
		updateFrame()
		waitTicks(1)
	}
}

// readString reads a line of up to maxLen printable characters, without key buffering.
// C: ULTIMA_3b1c_GetString
func readString(maxLen int) string {
	if prompt != promptWords {
		defer asking(promptText)()
	}
	buffering := keyBuffering
	keyBuffering = false
	var typed []byte
	for {
		key := getChar()
		switch {
		case (key == keyBackspace || key == keyLeft) && len(typed) != 0:
			backspace(1)
			typed = typed[:len(typed)-1]
		case key == keyEscape && len(typed) != 0:
			backspace(len(typed))
			typed = typed[:0]
		case key >= ' ' && key <= lastASCII && len(typed) < maxLen:
			typed = append(typed, byte(key))
			printChar(key)
		}
		if key == keyEnter {
			break
		}
	}
	keyBuffering = buffering
	return string(typed)
}

const maxNumberDigits = 5

// readNumber reads a signed number of up to maxLen characters.
// C: ULTIMA_3b9e
func readNumber(maxLen int) int {
	defer asking(promptText)()
	maxLen = min(maxLen, maxNumberDigits)
	var typed []byte
	for {
		key := getChar()
		switch {
		case key >= '0' && key <= '9':
			if len(typed) < maxLen {
				typed = append(typed, byte(key))
				printChar(key)
			}
		case (key == '-' || key == '+') && len(typed) == 0:
			typed = append(typed, byte(key))
			printChar(key)
		case (key == keyBackspace || key == keyLeft) && len(typed) != 0:
			typed = typed[:len(typed)-1]
			backspace(1)
		case key == keyEscape && len(typed) != 0:
			backspace(len(typed))
			typed = typed[:0]
		}
		if key == keyEnter {
			break
		}
	}
	// fix: the original reads a stale digit when nothing was typed.
	if len(typed) == 0 {
		return 0
	}
	value, place := 0, 1
	for i := len(typed) - 1; i > 0; i-- {
		value += int(typed[i]-'0') * place
		place *= 10
	}
	if typed[0] == '-' {
		value = -value
	} else if typed[0] >= '0' {
		value += place * int(typed[0]-'0')
	}
	return value
}

const (
	repairHours   = 5
	maxHull       = 99
	seaworthyHull = 10
	hoursMinutes  = 5 // the minutes passTime adds for each hour of repairs
	campTurnKind  = 6 // the special map of camping in a dungeon
)

// holeUpOutside makes camp or, on a ship with its sails furled, repairs the hull.
// C: ULTIMA_3c9a_HoleUpCmd
func holeUpOutside() {
	shown := actors[0].animTile
	printMessage(msgHoleUpSp2)
	if shown&^7 == actorFrigate32 {
		printMessage(msgRepairLine)
		if actors[0].animTile < actorFrigate36 {
			printMessage(msgSailsMustBeLine)
			printMessage(msgLoweredLine)
		} else {
			for range repairHours {
				pauseNearMonster()
				moveMonstersOutdoors()
				if partyTile&^3 != actorFrigate36 {
					return
				}
				passTime(hoursMinutes)
			}
			ship := &actors[0]
			for {
				ship.hull += byte(randomRange(1, 3))
				if ship.hull > maxHull {
					ship.hull = maxHull
				}
				if ship.hull >= seaworthyHull {
					break
				}
			}
			printMessage(msgHullNowSp)
			printNumber(int(ship.hull), 2, ' ')
			printMessage(msgChar33Newlines2)
			statusChanged = true
		}
		viewChanged = 1
		return
	}
	printMessage(msgCampLine)
	if mapID < firstDungeonMap {
		shown = *tileAt(int(mapX), int(mapY))
	}
	switch {
	case mapID < firstDungeonMap && shown != 0 && shown < 4:
		printMessage(msgOnLandOrShipLine)
	case mapID < firstDungeonMap && actors[0].animTile != actorAvatar:
		printMessage(msgOnFootLine)
	default:
		printMessage(msgForHowManyHours19)
		var key int
		for {
			key = getChar() & lowByteMask
			if key == keySpace || (key >= '0' && key <= '9') {
				break
			}
		}
		printChar(key)
		printChar('\n')
		if key == keySpace || key == '0' {
			return
		}
		hours := key - '0'
		awake := 0
		for i := 0; i != int(partySize); i++ {
			if party[i].status == statusGood || party[i].status == statusPoisoned {
				awake++
			}
		}
		guard := -1
		if awake > 1 {
			printMessage(msgWiltThouSetAWatchSp)
			for {
				key = getChar() & lowByteMask
				if key == 'Y' || key == 'N' {
					break
				}
			}
			if key == 'N' {
				printMessage(msgNoLine3)
			} else {
				printMessage(msgYesLine2)
				printMessage(msgWhoWillStandGuardSp)
				guard = selectOne()
				printChar('\n')
				if guard == -1 || party[guard].status != statusGood {
					guard = -1
					printMessage(msgNonePostedLine)
				}
			}
		}
		if mapID >= firstDungeonMap {
			combatKind = campTurnKind
			hideDungeonView()
			drawDungeonFrame()
			specialMap(int(combatKind), guard, hours)
			// fix: the original calls this without its argument, passing whatever is on the
			// stack; the wanderer is put back just below either way.
			restoreDungeonLook(0)
			actors[1] = savedActors[1]
			resumeDungeon(0)
		} else {
			camp(guard, hours)
		}
	}
	viewChanged = 1
}

// increaseByte adds amount to v, up to limit.
// C: ULTIMA_3ef0_IncreaseByte
func increaseByte(v *byte, amount, limit int) {
	if int(*v)+amount < limit {
		*v += byte(amount)
	} else {
		*v = byte(limit)
	}
}

// increaseInt adds amount to v, up to limit.
// C: ULTIMA_3f14_IncreaseInt
func increaseInt(v *int16, amount, limit int) {
	if int(*v)+amount < limit {
		*v += int16(amount)
	} else {
		*v = int16(limit)
	}
}

// decreaseByte takes amount from v, down to 0.
// C: ULTIMA_3f36_DecreaseByte
func decreaseByte(v *byte, amount int) {
	if int(*v) > amount {
		*v -= byte(amount)
	} else {
		*v = 0
	}
}

// decreaseInt takes amount from v, down to 0.
// C: ULTIMA_3f54_DecreaseInt
func decreaseInt(v *int16, amount int) {
	if int(*v) > amount {
		*v -= int16(amount)
	} else {
		*v = 0
	}
}

// Address of the table of which view tiles a ranged weapon can reach, a bit each.
const reachableTilesAddress = 27156 // C: D_6a14

var reachableTiles []byte // C: D_6a14

// isReachable says whether view position (x, y) can be fired at.
// C: ULTIMA_3f6e
func isReachable(x, y int) bool {
	tile := int(viewTile(x, y))
	return reachableTiles[tile/bitsPerByte]&(128>>(tile%bitsPerByte)) != 0
}

const (
	viewLeft   = 8
	viewRight  = 183
	tilePixels = 16
)

// pixelToView sets moveX, moveY to the view cell of screen position (x, y), -1 outside.
// C: ULTIMA_3fb4
func pixelToView(x, y int) {
	if x < viewLeft || x > viewRight || y < viewLeft || y > viewRight {
		moveX, moveY = -1, -1
		return
	}
	moveX = int16((x - viewLeft) / tilePixels)
	moveY = int16((y - viewLeft) / tilePixels)
}
