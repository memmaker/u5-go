package game

// The modern layout's knowledge of the game (no part of the original): what the game is
// waiting for, the menus that answer it, the commands that make sense where the party
// stands, and the font for the host's windows. Menus only type the keys a player would have
// typed: the game's rules are untouched.

import (
	"fmt"
	"strings"

	"github.com/memmaker/u5-go/internal/host"
)

type promptKind int

const (
	promptNone    promptKind = iota
	promptCommand            // a command loop's key (world, town, dungeon or combat)
	promptDir                // a direction
	promptMember             // a party member
	promptYesNo              // yes or no
	promptWords              // a conversation's words
	promptText               // free text
)

var (
	prompt promptKind
	// promptEveryone offers "everyone" (0) at a member prompt.
	promptEveryone bool
)

// ask tells the host what the game is about to wait for; asking promptNone ends it.
func ask(k promptKind) {
	prompt = k
	if k == promptCommand {
		host.InWorld = true
	}
	host.NewPrompt()
}

// asking asks k until the returned function ends it: defer asking(k)().
func asking(k promptKind) func() {
	ask(k)
	return func() { ask(promptNone) }
}

// leaveWorld tells the host the play screen is gone (the title, the ending).
func leaveWorld() { host.InWorld = false }

func init() {
	host.PromptMenu = promptMenu
	host.MapClick = mapClick
	host.Glyphs = glyphs
	host.BackKey = backKey
	host.ScreenClick = screenClick
}

// The host's windows draw text in the IBM font with these colours' indexes.
const (
	glyphOn  = 15
	glyphOff = 0
	ibmFont  = 0
)

var (
	glyphCache  [][64]byte
	glyphSource []byte
)

// glyphs returns the IBM font as 8x8 palette indexes per character.
func glyphs() [][64]byte {
	f := fonts[ibmFont]
	if len(f) == 0 {
		return nil
	}
	if len(glyphSource) > 0 && &glyphSource[0] == &f[0] {
		return glyphCache
	}
	glyphSource = f
	glyphCache = make([][64]byte, len(f)/glyphBytes)
	for c := range glyphCache {
		for row := range cellSize {
			bits := f[c*glyphBytes+row]
			for col := range cellSize {
				if bits&(0x80>>col) != 0 {
					glyphCache[c][row*cellSize+col] = glyphOn
				} else {
					glyphCache[c][row*cellSize+col] = glyphOff
				}
			}
		}
	}
	return glyphCache
}

// backKey is what the gamepad's B types: Enter where an empty line cancels, nothing at a
// command, else Esc.
func backKey() (host.Key, bool) {
	switch prompt {
	case promptWords, promptText:
		return host.KeyEnter, true
	case promptCommand:
		return host.Key{}, false
	}
	return host.KeyEsc, true
}

// The map window on the original screen.
const (
	mapWindowLeft  = 8
	mapWindowTop   = 8
	mapWindowRight = mapWindowLeft + viewSize*tilePixels
)

// screenClick turns a click on the original layout's map window into a mapClick.
func screenClick(x, y int) ([]host.Key, bool) {
	if !host.InWorld || inDungeon() || x < mapWindowLeft || x >= mapWindowRight || y < mapWindowTop || y >= mapWindowRight {
		return nil, false
	}
	keys, ok := mapClick((x-mapWindowLeft)/tilePixels-viewCentre, (y-mapWindowTop)/tilePixels-viewCentre)
	return keys, ok && keys != nil
}

func dirKey(dx, dy int) host.Key {
	switch {
	case dy < 0:
		return host.KeyUp
	case dy > 0:
		return host.KeyDown
	case dx < 0:
		return host.KeyLeft
	}
	return host.KeyRight
}

var dirNames = []struct {
	dx, dy int
	name   string
}{{0, -1, "North"}, {0, 1, "South"}, {1, 0, "East"}, {-1, 0, "West"}}

func key(c byte) host.Key { return host.CharKey(c) }

func item(label string, keys ...host.Key) host.MenuItem {
	return host.MenuItem{Label: label, Keys: keys}
}

func promptMenu(asked bool) *host.Menu {
	switch prompt {
	case promptCommand:
		if !asked {
			return nil
		}
		return &host.Menu{Title: "Command", Items: commandItems(), Pause: true}
	case promptDir:
		m := &host.Menu{Title: "Direction"}
		for _, d := range dirNames {
			m.Items = append(m.Items, item(d.name, dirKey(d.dx, d.dy)))
		}
		return m
	case promptMember:
		m := &host.Menu{Title: "Who"}
		for i := range int(partySize) {
			m.Items = append(m.Items, item(fmt.Sprintf("%d %s", i+1, party[i].nameString()), key(byte('1'+i)), host.KeyEnter))
		}
		if promptEveryone {
			m.Items = append(m.Items, item("0 Everyone", key('0')))
		}
		return m
	case promptYesNo:
		return &host.Menu{Items: []host.MenuItem{item("Yes", key('Y')), item("No", key('N'))}}
	case promptWords:
		m := &host.Menu{Title: "Say (or type)"}
		for _, w := range talkMenuWords() {
			keys := make([]host.Key, 0, len(w)+1)
			for i := range len(w) {
				keys = append(keys, key(w[i]))
			}
			m.Items = append(m.Items, item(w, append(keys, host.KeyEnter)...))
		}
		m.Items = append(m.Items, host.MenuItem{Label: "Type...", Keyboard: true}, item("(nothing)", host.KeyEnter))
		return m
	case promptText:
		if !asked {
			return nil
		}
		return &host.Menu{Items: []host.MenuItem{{Label: "Type...", Keyboard: true}}}
	}
	return nil
}

// talkMenuWords are the words worth saying: the answers to the question asked, else what
// everyone knows (name, job, bye) and the person's own topics. Reading the script leaves
// the conversation where it was.
func talkMenuWords() []string {
	saved := talkAt
	defer func() { talkAt = saved }()
	var words []string
	add := func() {
		var w []byte
		for _, c := range script()[talkAt:] {
			if c == 0 || c > textMask || c < ' ' {
				break
			}
			w = append(w, c)
		}
		if s := strings.TrimSpace(string(w)); s != "" {
			for _, have := range words {
				if strings.EqualFold(have, s) {
					return
				}
			}
			words = append(words, s)
		}
	}
	if talkAnswering != 0 {
		findLabel()
		skipTo(0, scriptLabelLast)
		skipTo(0, scriptLabelLast)
		for range maxTalkWords {
			add()
			if !skipTo(0, scriptLabelEnd) || !skipTo(0, scriptLabelEnd) {
				break
			}
		}
		return words
	}
	words = append(words, "name", "job", "bye")
	for n := 0; n < maxTalkWords && selectKeyword(n); n++ {
		add()
	}
	return words
}

// maxTalkWords bounds the words read from a script.
const maxTalkWords = 64

// commandItems are the commands that apply where the party is: first what it can do with
// its neighbours and the ground it stands on, then what it can do anywhere with what it has.
func commandItems() []host.MenuItem {
	var m []host.MenuItem
	when := func(ok bool, label string, c byte) {
		if ok {
			m = append(m, item(label, key(c)))
		}
	}
	mixtures := false
	for _, n := range spellMixtures {
		mixtures = mixtures || n != 0
	}
	usable := false
	for _, n := range scrolls {
		usable = usable || n != 0
	}
	for _, n := range potions {
		usable = usable || n != 0
	}
	usable = usable || gems != 0 || magicCarpets != 0 || skullKeys != 0 || amulet != 0 || crown != 0 ||
		sceptre != 0 || spyglasses != 0 || hmsCapePlans != 0 || sextants != 0 || pocketWatch != 0 ||
		blackBadge != 0 || sandalwoodBox != 0
	for _, n := range shards {
		usable = usable || n != 0
	}
	switch {
	case mapID >= firstMapCombat:
		m = append(m, item("Attack", key('A')))
		when(mixtures, "Cast", 'C')
		when(true, "Ready", 'R')
		when(usable, "Use item", 'U')
		when(true, "Ztats", 'Z')
	case inDungeon():
		m = dungeonContext()
		when(mixtures, "Cast", 'C')
		when(true, "Hole up & camp", 'H')
		when(torches != 0, "Ignite torch", 'I')
		when(true, "Mix reagents", 'M')
		when(partySize > 1, "New order", 'N')
		when(true, "Quit & save", 'Q')
		when(true, "Ready", 'R')
		when(true, "Search", 'S')
		when(usable, "Use item", 'U')
		when(gems != 0, "View (gem)", 'V')
		when(true, "Ztats", 'Z')
	default:
		m = worldContext()
		when(mixtures, "Cast", 'C')
		when(mapID == 0, "Hole up & camp", 'H')
		when(torches != 0, "Ignite torch", 'I')
		when(true, "Mix reagents", 'M')
		when(partySize > 1, "New order", 'N')
		when(true, "Quit & save", 'Q')
		when(true, "Ready", 'R')
		when(usable, "Use item", 'U')
		when(gems != 0, "View (gem)", 'V')
		when(partyTile&actorShipMask == actorFrigate32 || partyTile&actorShipMask == actorFrigate36, "Yell (sails)", 'Y')
		when(true, "Ztats", 'Z')
	}
	m = append(m, item("Pass", host.KeySpace))
	all := &host.Menu{Title: "All commands", Pause: true}
	for i, name := range allCommands {
		all.Items = append(all.Items, item(name, key(byte('A'+i))))
	}
	return append(m, host.MenuItem{Label: "All commands...", Sub: all})
}

// allCommands are the original's commands, A to Z, for what the context list leaves out.
var allCommands = []string{"Attack", "Board", "Cast", "(D: none)", "Enter", "Fire", "Get",
	"Hole up & camp", "Ignite torch", "Jimmy lock", "Klimb", "Look", "Mix reagents", "New order",
	"Open", "Push", "Quit & save", "Ready", "Search", "Talk", "Use item", "View (gem)",
	"(W: none)", "X-it", "Yell", "Ztats"}

// actorHere is the tile of the first actor at (x, y) on the party's level, from slot 1;
// actorAt without its side effect.
func actorHere(x, y int) int {
	for slot := 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		if a.tile != 0 && int(a.x) == x && int(a.y) == y && int(a.z) == int(level) {
			return int(a.tile)
		}
	}
	return 0
}

// isPerson and isMonster tell actors' tiles apart.
func isPerson(tile int) bool  { return tile >= actorPeopleFirst && tile <= actorPeopleLast }
func isMonster(tile int) bool { return tile >= actorMonstersFirst }

// worldContext lists what the party can do with its neighbours and the tile it is on.
func worldContext() []host.MenuItem {
	var m []host.MenuItem
	inTown := mapID != 0
	for _, d := range dirNames {
		x, y := int(mapX)+d.dx, int(mapY)+d.dy
		dk := dirKey(d.dx, d.dy)
		switch t := actorHere(x, y); {
		case t == 0:
		case inTown && isPerson(t):
			m = append(m, item("Talk "+d.name, key('T'), dk))
		case isMonster(t), !inTown && isPerson(t):
			m = append(m, item("Attack "+d.name, key('A'), dk))
		}
		if !inTown || x < 0 || x >= settlementSize || y < 0 || y >= settlementSize {
			continue
		}
		switch *tileAt(x, y) {
		case mapTileDoor184, mapTileDoor186:
			m = append(m, item("Open door "+d.name, key('O'), dk))
		case lockedDoorLeft, lockedDoorTwo, mapTileDoor185, mapTileDoor187:
			m = append(m, item("Jimmy lock "+d.name, key('J'), dk))
		case mapTileTrunk:
			m = append(m, item("Open trunk "+d.name, key('O'), dk))
		}
	}
	here := actorHere(int(mapX), int(mapY))
	onFoot := partyTile&shapeMask == actorAvatar&shapeMask
	switch {
	case here >= actorItemsFirst && here <= actorItemsLast && inTown:
		m = append(m, item("Get", key('G'), host.KeyUp))
	case onFoot && here != 0 && !isPerson(here) && !isMonster(here):
		m = append(m, item("Board", key('B')))
	}
	if !onFoot {
		m = append(m, item("X-it", key('X')))
		if partyTile&actorShipMask == actorFrigate32 || partyTile&actorShipMask == actorFrigate36 {
			m = append(m, item("Fire cannon", key('F')))
		}
	}
	if !inTown {
		for place := range placeCount {
			if placeX[place] == mapX && placeY[place] == mapY && placeNames[place] != "" {
				m = append(m, item("Enter "+strings.TrimSpace(placeNames[place]), key('E')))
				break
			}
		}
	} else {
		switch *tileAt(int(mapX), int(mapY)) {
		case mapTileLadderUp, mapTileLadderDown:
			m = append(m, item("Klimb ladder", key('K')))
		}
	}
	return m
}

// dungeonContext lists what the party can do on its dungeon square.
func dungeonContext() []host.MenuItem {
	var m []host.MenuItem
	cell := *dungeonCellAt(int(mapX)&(dungeonSize-1), int(mapY)&(dungeonSize-1), int(level))
	switch cell & cellKindMask {
	case cellLadderUp, cellLadderDown, cellLadders:
		m = append(m, item("Klimb ladder", key('K')))
	case cellChest:
		m = append(m, item("Open chest", key('O')), item("Jimmy chest", key('J')))
	}
	return m
}

// mapClick turns a click on the map into keys: a step or a direction toward the tile, or
// the key a "press a key" wait wants. The party's own tile opens the command menu.
func mapClick(dx, dy int) ([]host.Key, bool) {
	if mapID >= firstMapCombat && combatant != noCombatant { // relative to the fighter
		f := &combatants[combatant]
		dx, dy = dx+viewCentre-int(f.x), dy+viewCentre-int(f.y)
	}
	if abs(dx) >= abs(dy) {
		dy = 0
	} else {
		dx = 0
	}
	switch {
	case prompt == promptCommand && dx == 0 && dy == 0:
		return nil, false
	case prompt == promptCommand || prompt == promptDir && (dx != 0 || dy != 0):
		return []host.Key{dirKey(dx, dy)}, true
	case host.KeyPolled && prompt == promptNone:
		return []host.Key{host.KeySpace}, true
	}
	return nil, true
}
