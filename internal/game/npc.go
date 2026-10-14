package game

// The settlements' people (npc.c): their files, their daily schedules, and their walking,
// with a breadth-first search for a path to the place they keep.

// The people files, by settlement kind.
var peopleFiles = [settlementKinds]string{"TOWNE.NPC", "DWELLING.NPC", "CASTLE.NPC", "KEEP.NPC"}

const (
	peopleRecord    = 576 // a settlement's bytes in its people file
	schedulesBytes  = 512
	typesOffset     = 512
	moodsOffset     = 544
	scheduleRecSize = 16
)

// loadPeople reads the settlement's people: their schedules, tiles and moods.
// C: NPC_0000_LoadNpcFile
func loadPeople() {
	index := int(mapID) - 1
	kind := index >> 3
	at := (index - kind<<3) * peopleRecord
	buffer := make([]byte, schedulesBytes)
	readFile(peopleFiles[kind], buffer, at)
	for who := range schedules {
		c := saveCursor{buf: buffer, at: who * scheduleRecSize}
		c.schedule(&schedules[who])
	}
	readFile(peopleFiles[kind], npcTypes[:], at+typesOffset)
	var moods [npcSlots]byte
	readFile(peopleFiles[kind], moods[:], at+moodsOffset)
	for who := range npcSlots {
		npcs[who].mood = uint16(moods[who])
	}
}

// placePeopleAt sets every person at the schedule's place for hour.
// C: NPC_00d6
func placePeopleAt(hour int) {
	for who := 1; who < npcSlots; who++ {
		n := &npcs[who]
		if npcTypes[who] == 0 {
			n.moveState = 0
			continue
		}
		place := schedulePlace(who, hour)
		s := &schedules[who]
		n.moveState = 1
		n.x, n.y, n.z = int16(s.x[place]), int16(s.y[place]), uint16(s.z[place])
		n.shape = uint16(npcTypes[who])
		n.place = uint16(place)
		n.actorAt = 0
		movementPos[who] = -1
		movements[who][0] = 0
		npcWaits[who] = 0
	}
}

// The path grid: each cell's high nibble is the way it was reached from, the low one marks
// the goal (5) and the start (6); a cell of 144 is blocked.
const (
	pathFree      = 0
	pathGoal      = 5
	pathStart     = 6
	pathBlocked   = 144
	pathStartCell = 70 // the start, reached "from" way 4
	pathQueueSize = 32
	nibble        = 4
	nibbleMask    = 15
	nearPerson    = 4
	pathLadderUp  = -1
	pathLadderDn  = -2
	noWay         = 99
)

// pathGrid is the search's grid; it shares the scratch buffer as the original does.
// C: D_b11c
func pathGrid() []byte { return scratch[:] }

// findLadder looks for the nearest ladder up (or down) from x, y.
// C: NPC_01a0
func findLadder(x, y int, up bool, who int) int {
	goal := pathLadderDn
	if up {
		goal = pathLadderUp
	}
	return findPath(x, y, 0, 0, goal, who)
}

// markPathGrid sets up the grid for person who's search from x, y to goalX, goalY (or a
// ladder when goal is negative).
// C: NPC_01d2
func markPathGrid(x, y, goalX, goalY, goal, who int) {
	grid := pathGrid()
	var ladder byte
	if goal == pathLadderUp {
		ladder = mapTileLadderUp
	}
	if goal == pathLadderDn {
		ladder = mapTileLadderDown
	}
	place := schedulePlace(who, int(hour))
	for row := range settlementSize {
		for column := range settlementSize {
			cell := row*settlementSize + column
			if personMayStand(column, row, int(level), who, place) != 0 {
				grid[cell] = pathFree
			} else {
				grid[cell] = pathBlocked
			}
			if goal < 0 && mapTiles[cell] == ladder {
				grid[cell] = pathGoal
			}
		}
	}
	n := &npcs[who]
	for slot := actorSlots - 1; slot > 0; slot-- {
		a := &actors[slot]
		if a.tile > actorWizard-1 && distance(int(a.x), int(a.y), int(n.x), int(n.y)) < nearPerson {
			grid[int(a.y)*settlementSize+int(a.x)] = pathBlocked
		}
	}
	a := &actors[0]
	if distance(int(a.x), int(a.y), int(n.x), int(n.y)) < nearPerson {
		grid[int(a.y)*settlementSize+int(a.x)] = pathBlocked
	}
	if goal > -1 {
		grid[goalY*settlementSize+goalX] = pathGoal
	}
	grid[y*settlementSize+x] = pathStartCell
}

// findPath searches the grid breadth first from x, y; on success it leaves the goal
// reached in moveX, moveY.
// C: NPC_032c
func findPath(x, y, goalX, goalY, goal, who int) int {
	var queueX, queueY [pathQueueSize]byte
	grid := pathGrid()
	found, head, tail := 0, 0, 1
	markPathGrid(x, y, goalX, goalY, goal, who)
	queueX[0], queueY[0] = byte(x), byte(y)
	for {
		cx, cy := int(queueX[head]), int(queueY[head])
		way := int(grid[cy*settlementSize+cx] >> nibble)
		for range 4 {
			switch way {
			case 1:
				cx--
				if cx < 0 {
					found = 2
				}
			case 2:
				cy++
				if cy > settlementSize {
					found = 2
				}
			case 3:
				cx++
				if cx > settlementSize {
					found = 2
				}
			case 4:
				cy--
				if cy < 0 {
					found = 2
				}
			}
			if found != 2 {
				cell := cy*settlementSize + cx
				if grid[cell] < 1<<nibble {
					found = int(grid[cell] & nibbleMask)
					grid[cell] = byte(way << nibble)
					if found == pathGoal {
						found = 1
						moveX, moveY = int16(cx), int16(cy)
						break
					}
					if tail != head {
						queueX[tail], queueY[tail] = byte(cx), byte(cy)
						tail++
					}
					if tail >= pathQueueSize {
						tail = 0
					}
				}
			}
			found = 0
			way = way&3 + 1
			cx, cy = int(queueX[head]), int(queueY[head])
		}
		if found == 1 {
			return 1
		}
		head++
		if head == pathQueueSize {
			head = 0
		}
		if head == tail {
			return found
		}
	}
}

// tracePath writes the path found to x, y into person who's movement queue as pairs of
// steps and way, from the start.
// C: NPC_04ac
func tracePath(who, goal, x, y int) int {
	grid := pathGrid()
	path := &movements[who]
	steps, at := 0, 0
	movementPos[who] = 0
	mark := int(grid[y*settlementSize+x] & nibbleMask)
	way := int(grid[y*settlementSize+x] >> nibble)
	last := way
	for {
		switch way {
		case 1:
			x++
		case 2:
			y--
		case 3:
			x--
		case 4:
			y++
		}
		if last == way && mark != pathStart {
			steps++
		}
		if last != way || mark == pathStart {
			path[at] = byte(steps)
			at++
			path[at] = byte(last)
			at++
			if mark == pathStart {
				break
			}
			last = way
			steps = 1
		}
		way = int(grid[y*settlementSize+x] >> nibble)
		mark = int(grid[y*settlementSize+x] & nibbleMask)
		if at >= movementBytes {
			break
		}
	}
	back := at - 2
	if back >= 0 {
		front := 0
		for {
			path[front], path[back] = path[back], path[front]
			back++
			front++
			swap := (path[front]+1)&3 + 1
			path[front] = (path[back]+1)&3 + 1
			path[back] = swap
			front++
			back -= 3
			if back < front {
				break
			}
		}
	}
	return at
}

// stepWay moves moveX, moveY one step in way (1 east, 2 north, 3 west, 4 south), with
// the original's odd clamps.
// C: NPC_0632
func stepWay(way int) {
	switch way {
	case 1:
		moveX++
		if moveY > settlementSize {
			moveX = settlementSize
		}
	case 2:
		moveY--
		if moveX < 0 {
			moveY = 0
		}
	case 3:
		moveX--
		if moveY < 0 {
			moveX = 0
		}
	case 4:
		moveY++
		if moveX > settlementSize {
			moveY = settlementSize
		}
	}
}

// distance returns the steps between two places.
// C: NPC_06a0
func distance(x1, y1, x2, y2 int) int {
	dx, dy := x1-x2, y1-y2
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Schedule kinds.
const (
	kindStill   = 0
	kindWander3 = 1
	kindWander  = 2
	kindFlee    = 3
	kindWait    = 4
	kindTalk    = 5
)

// approachParty moves person who toward the party (or away when fleeing), and has them
// talk or attack when beside it.
// C: NPC_06e4
func approachParty(who, place int) {
	party0 := &actors[0]
	n := &npcs[who]
	kind := int(schedules[who].kind[place])
	near := distance(int(party0.x), int(party0.y), int(n.x), int(n.y))
	if near == 1 && kind > kindFlee {
		if kind == kindWait || kind == kindTalk {
			if n.mood != 0 {
				townMode = townModeTalk
				npcTurn = byte(who)
			}
		} else {
			townMode = townModeAttack
			npcTurn = byte(who)
		}
		return
	}
	var away [4]int
	for way := 1; way < 5; way++ {
		moveX, moveY = n.x, n.y
		stepWay(way)
		x, y := int(moveX), int(moveY)
		if personMayStep(x, y, who, -1) != 0 {
			away[way-1] = distance(int(party0.x), int(party0.y), x, y)
		} else {
			away[way-1] = noWay
		}
	}
	chosen := -1
	near = distance(int(party0.x), int(party0.y), int(n.x), int(n.y))
	for way := 1; way < 8; way++ {
		if way < 5 && away[way-1] != noWay {
			if kind == kindFlee {
				if near < away[way-1] {
					if chosen == -1 || randomRange(0, 1) != 0 {
						chosen = way
					}
					if way == 4 {
						break
					}
				}
			} else if away[way-1] < near {
				chosen = way
				break
			}
		}
		// fix: the original tests away[way-1] here, past the array's end.
		if way > 4 && away[way-5] != noWay && away[way-5] == near {
			chosen = way
			break
		}
	}
	if (kind == kindTalk || kind == scheduleHunt) && randomRange(0, 63) < 16 {
		other := chosen
		for way := 1; way < 5; way++ {
			if chosen != way && away[way-1] != noWay && (other == chosen || randomRange(0, 63) < 16) {
				other = way
			}
		}
		chosen = other
	}
	if chosen > -1 {
		moveX, moveY = n.x, n.y
		stepWay(chosen)
		n.x, n.y = moveX, moveY
		actors[n.actorAt].x = byte(moveX)
		actors[n.actorAt].y = byte(moveY)
		viewChanged |= 2
	}
}

// Movement states (npc.moveState).
const (
	stateAbsent      = 0
	stateThere       = 1
	stateWalking     = 2
	stateToStairs    = 3
	stateFetchUp     = 4
	stateFetchDown   = 5
	stateLeaveUp     = 6
	stateLeaveDown   = 7
	stateOtherLevels = 8
)

// checkSchedule starts person who on the way to the next place when hour is one of their
// schedule's times; it returns the place plus one, or 0.
// C: NPC_0938
func checkSchedule(who, hour int) int {
	n := &npcs[who]
	s := &schedules[who]
	place := -1
	for i := range scheduleTimes {
		if int(s.time[i]) != hour {
			continue
		}
		place = schedulePlace(who, hour)
		if place == int(n.place) {
			n.moveState = stateThere
			break
		}
		switch {
		case n.z != uint16(level) && s.z[place] != level:
			n.moveState = stateOtherLevels
		case n.z == uint16(level):
			switch {
			case s.z[place] == level:
				n.moveState = stateWalking
			case int8(level) < int8(s.z[place]):
				n.moveState = stateLeaveUp
			default:
				n.moveState = stateLeaveDown
			}
		case int8(level) < int8(n.z):
			n.moveState = stateFetchUp
		default:
			n.moveState = stateFetchDown
		}
		break
	}
	if place > -1 && n.x == int16(s.x[place]) && n.y == int16(s.y[place]) && n.z == uint16(s.z[place]) {
		place = 0
		n.moveState = stateThere
	}
	if place == -1 {
		return 0
	}
	return place + 1
}

const stairMask = 244

// onLevelChange tells whether person who stands on a ladder or stairs going their way.
// C: NPC_0a4a
func onLevelChange(who, place int) bool {
	n := &npcs[who]
	a := &actors[n.actorAt]
	tile := *tileAt(int(a.x), int(a.y))
	if int8(schedules[who].z[place]) < int8(level) {
		return tile == mapTileLadderDown || tile&stairMask == mapTileStair
	}
	return tile == mapTileLadderUp || tile&stairMask == mapTileStair
}

// personMayStand tells whether person who can stand at x, y on level z: 2 at the place
// they keep, 1 on open ground.
// C: NPC_0adc
func personMayStand(x, y, z, who, place int) int {
	s := &schedules[who]
	if place > -1 && int(s.x[place]) == x && int(s.y[place]) == y && int(s.z[place]) == z {
		return 2
	}
	var tile byte
	if x < 0 || y < 0 || x >= settlementSize || y >= settlementSize {
		tile = mapTiles[lastMapByte]
	} else {
		tile = mapTiles[y*settlementSize+x]
	}
	if npcs[who].moveState == stateToStairs && (tile == mapTileLadderUp || tile == mapTileLadderDown) {
		return 1
	}
	if personBlocks[tile>>3]&(128>>(tile&7)) == 0 {
		return 1
	}
	return 0
}

// personMayStep tells whether person who can step onto x, y.
// C: NPC_0b9e
func personMayStep(x, y, who, place int) int {
	if x < 0 || x >= settlementSize || y < 0 || y >= settlementSize {
		return 0
	}
	tile := int(*tileAt(x, y) & shapeMask)
	free := 0
	switch {
	case npcTypes[who] == shadowlordKind:
		free = 1
	case tile == mapTile48:
		free = 1
	case tile == mapTileChair144 && npcs[who].moveState != stateWalking:
		free = 0
	default:
		free = personMayStand(x, y, int(level), who, place)
	}
	if free == 0 || topActorAt(x, y, int(level)) != 0 {
		return 0
	}
	return free
}

// wanderPerson moves a person a random step now and then, within reach of their place.
// C: NPC_0c50
func wanderPerson(n *npc, reach, place, who int, s *schedule) {
	if randomRange(0, 255)&8 == 0 {
		return
	}
	way := random(64)&3 + 1
	moveX, moveY = n.x, n.y
	stepWay(way)
	x, y := moveX, moveY
	if (reach == 0 || distance(int(s.x[place]), int(s.y[place]), int(moveX), int(moveY)) <= reach) &&
		personMayStep(int(moveX), int(moveY), who, place) != 0 {
		n.x, n.y = x, y
		actors[n.actorAt].x = byte(x)
		actors[n.actorAt].y = byte(y)
		viewChanged |= 2
	}
}

// keepSchedule does what person who's schedule asks at their place.
// C: NPC_0d00
func keepSchedule(who, place int) {
	n := &npcs[who]
	s := &schedules[who]
	switch s.kind[place] {
	case kindStill:
	case kindWander:
		wanderPerson(n, 0, place, who, s)
	case kindWait:
		if distance(int(mapX), int(mapY), int(s.x[place]), int(s.y[place])) < nearPerson {
			approachParty(who, place)
		}
	case kindWander3:
		wanderPerson(n, 3, place, who, s)
	case kindFlee, scheduleGuard:
		if distance(int(mapX), int(mapY), int(n.x), int(n.y)) < nearPerson {
			approachParty(who, place)
		}
	case kindTalk, scheduleHunt:
		approachParty(who, place)
	}
}

const (
	stuckLimit   = 3
	giveUpWait   = 200
	giveUpResume = 204
)

// movePeople moves the settlement's people for hour.
// C: NPC_0db4
func movePeople(hour int) {
	searched := 0
	townMode, npcTurn = 0, 0
	for who := 1; who < npcSlots; who++ {
		if npcTypes[who] == 0 {
			continue
		}
		place := schedulePlace(who, hour)
		n := &npcs[who]
		s := &schedules[who]
		switch {
		case n.moveState <= stateThere && checkSchedule(who, hour) == 0:
			if n.z == uint16(level) {
				keepSchedule(who, place)
			}
		case n.moveState > stateToStairs:
			switch n.moveState {
			case stateFetchUp, stateFetchDown:
				if searched == 1 {
					break
				}
				searched = 1
				goal := 4
				if n.moveState == stateFetchUp {
					goal = 3
				}
				var lx, ly int
				found := findLadder(int(s.x[place]), int(s.y[place]), goal == 3, who)
				if found != 0 {
					lx, ly = int(moveX), int(moveY)
					found = findPath(lx, ly, int(s.x[place]), int(s.y[place]), goal, who)
				}
				if found != 0 {
					tracePath(who, goal, int(s.x[place]), int(s.y[place]))
					tile := mapTiles[ly*settlementSize+lx]
					if (goal == 3 && tile == mapTileLadderUp) || (goal == 4 && tile == mapTileLadderDown) || tile&shapeMask == mapTileStair {
						placePerson(who, lx, ly, int(level))
					}
					n.moveState = stateWalking
				}
			case stateLeaveUp, stateLeaveDown:
				if onLevelChange(who, place) {
					placePerson(who, int(s.x[place]), int(s.y[place]), int(s.z[place]))
					n.place = uint16(place)
					movementPos[who] = -1
					n.moveState = stateThere
					break
				}
				if n.moveState == stateWalking || searched == 1 {
					break
				}
				searched = 1
				goal := 2
				if n.moveState == stateLeaveUp {
					goal = 1
				}
				var lx, ly int
				found := findLadder(int(n.x), int(n.y), goal == 1, who)
				if found != 0 {
					lx, ly = int(moveX), int(moveY)
					found = findPath(int(n.x), int(n.y), lx, ly, goal, who)
				}
				if found != 0 {
					tracePath(who, goal, lx, ly)
					n.moveState = stateToStairs
				}
			}
		case movementPos[who] > -1 && movements[who][movementPos[who]] != 0:
			path := &movements[who]
			moveX, moveY = n.x, n.y
			stepWay(int(path[movementPos[who]+1]))
			x, y := int(moveX), int(moveY)
			free := personMayStep(x, y, who, place)
			if free == 0 {
				npcWaits[who]++
				wanderPerson(n, 0, place, who, s)
				if npcWaits[who] > stuckLimit {
					movementPos[who] = -1
					npcWaits[who] = 0
				}
				break
			}
			actors[n.actorAt].x = byte(x)
			n.x = int16(x & lowByteMask)
			actors[n.actorAt].y = byte(y)
			n.y = int16(y & lowByteMask)
			viewChanged |= 2
			path[movementPos[who]]--
			npcWaits[who] = 0
			if path[movementPos[who]] != 0 {
				break
			}
			movementPos[who]++
			path[movementPos[who]] = 0
			movementPos[who]++
			if movementPos[who] >= movementBytes || path[movementPos[who]] < 1 {
				movementPos[who] = -1
			}
			if free == 2 {
				n.place = uint16(place)
				n.moveState = stateThere
				movementPos[who] = -1
			}
		default:
			if movementPos[who] == -1 && n.moveState == stateToStairs {
				place = schedulePlace(who, hour)
				if level < s.z[place] {
					n.moveState = stateLeaveUp
				} else {
					n.moveState = stateLeaveDown
				}
				return
			}
			switch {
			case searched < 1:
				if n.moveState == stateThere {
					break
				}
				if npcWaits[who] < giveUpWait && (npcWaits[who] == 0 || randomRange(0, 2) == 1) {
					if movementPos[who] == -1 {
						searched++
						if findPath(int(n.x), int(n.y), int(s.x[place]), int(s.y[place]), 0, who) != 0 {
							tracePath(who, 0, int(s.x[place]), int(s.y[place]))
							npcWaits[who] = 0
							continue
						}
						npcWaits[who] = giveUpWait
					}
					wanderPerson(n, 0, place, who, s)
				} else {
					if npcWaits[who] >= giveUpWait {
						npcWaits[who]++
					}
					if npcWaits[who] > giveUpResume {
						npcWaits[who] = 0
					}
				}
			default:
				keepSchedule(who, int(n.place))
			}
		}
	}
}

// schedulePlace returns the schedule entry person who keeps at hour: the one whose time
// passed last (the fourth time goes back to the second place).
// C: NPC_12e0
func schedulePlace(who, hour int) int {
	var since [scheduleTimes]byte
	for i := range scheduleTimes {
		since[i] = byte(hour) - schedules[who].time[i]
	}
	place := 0
	if since[1] < since[0] {
		since[0] = since[1]
		place = 1
	}
	if since[2] < since[0] {
		since[0] = since[2]
		place = 2
	}
	if since[3] < since[0] {
		place = 1
	}
	return place
}
