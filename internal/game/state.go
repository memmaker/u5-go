package game

// The game's saved state: the part of the data segment the original reads from INIT.GAM
// and writes to SAVED.GAM as one block (savegame.go reads and writes it field by field).

const (
	partySlots   = 16
	maxPartySize = 6
	nameBytes    = 9
	equipSlots   = 6
	actorSlots   = 32
	npcSlots     = 32
	settlements  = 32
	spellCount   = 48
	itemTypes    = 48
	scrollTypes  = 8
	potionTypes  = 8
	reagentTypes = 8
	moonstones   = 8
	shardCount   = 4
	// the dungeon map buffer holds 8 levels of 8x8 cells
	dungeonLevels = 8
	dungeonSize   = 8
)

// Party member statuses.
const (
	statusGood     = 'G'
	statusPoisoned = 'P'
	statusDead     = 'D'
	statusSleeping = 'S'
	statusCharmed  = 'C'
)

// Classes.
const (
	classAvatar   = 'A'
	classMage     = 'M'
	classBard     = 'B'
	classFighter  = 'F'
	classDruid    = 'D'
	classTinker   = 'T'
	classPaladin  = 'P'
	classRanger   = 'R'
	classShepherd = 'S'
)

// Equipment slots of a party member.
const (
	slotHead = iota
	slotBody
	slotLeftHand
	slotRightHand
	slotRing
	slotAmulet
)

// noEquipment marks an empty equipment slot.
const noEquipment = 255

// member is a party member or a character who can join (S_55a8).
type member struct {
	name   [nameBytes]byte
	gender byte
	class  byte
	status byte
	str    byte
	dex    byte
	intel  byte
	magic  byte
	hp     int16
	maxHP  int16
	exp    int16
	level  byte
	// monthsInInn counts the months the member has waited at an inn (_17).
	monthsInInn byte
	// innSlot is the inn the member waits at (_18).
	innSlot byte
	equip   [equipSlots]byte
	// location is the settlement the member waits in, 0 in the party, 127 gone (mapId).
	location byte
}

// nameString returns the member's name.
func (m *member) nameString() string { return cString(m.name[:]) }

// cString returns the bytes of b up to its first zero.
func cString(b []byte) string {
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// actor is a thing shown on the map: the party, a ship, a monster, a person (ActorFmt).
type actor struct {
	tile     byte // the base tile (shape)
	animTile byte // the tile shown this frame
	x, y, z  byte
	// hull is a ship's hull strength, a monster's hit points in combat, a combatant's
	// entity (_5).
	hull byte
	// state is a monster's movement or sleep state (_6).
	state byte
	// target is the combatant this one is fighting, 255 none (_7).
	target byte
}

// npc is the state of a settlement's person (NpcFmt).
type npc struct {
	// moveState is how the person moves: 0 absent, 1 at the schedule's place, 2 walking
	// there, 3 walking to stairs, 4-8 changing level (_0).
	moveState uint16
	x, y      int16
	z         uint16
	shape     uint16 // the person's tile (_8)
	// mood is 0 for one who will not talk, 253 fleeing, 254 hostile (_a).
	mood    uint16
	actorAt uint16 // the actor slot showing the person (_c)
	place   uint16 // the schedule entry the person keeps (_e)
}

// schedulePoints is the number of places of a person's daily schedule; scheduleTimes the
// times it changes.
const (
	schedulePoints = 3
	scheduleTimes  = 4
)

// schedule is a person's day (NpcScheduleFmt).
type schedule struct {
	kind    [schedulePoints]byte
	x, y, z [schedulePoints]byte
	time    [scheduleTimes]byte
}

// movementBytes is the size of a person's queued path.
const movementBytes = 32

// The saved state (C names in the comments).
var (
	saveTag uint16             // C: D_55a6
	party   [partySlots]member // C: D_55a8

	food           int16              // C: D_57a8
	gold           int16              // C: D_57aa
	keys           byte               // C: D_57ac
	gems           byte               // C: D_57ad
	torches        byte               // C: D_57ae
	grapple        byte               // C: D_57af
	magicCarpets   byte               // C: D_57b0
	skullKeys      byte               // C: D_57b1
	skullKeyDay    byte               // C: D_57b2 (the day Minoc's skull keys were taken)
	amulet         byte               // C: D_57b3
	crown          byte               // C: D_57b4
	sceptre        byte               // C: D_57b5
	shards         [shardCount]byte   // C: D_57b6
	spyglasses     byte               // C: D_57ba
	hmsCapePlans   byte               // C: D_57bb
	sextants       byte               // C: D_57bc
	pocketWatch    byte               // C: D_57bd
	blackBadge     byte               // C: D_57be
	sandalwoodBox  byte               // C: D_57bf
	equipment      [itemTypes]byte    // C: D_57c0 (counts owned)
	spellMixtures  [spellCount]byte   // C: D_57f0
	scrolls        [scrollTypes]byte  // C: D_5820
	potions        [potionTypes]byte  // C: D_5828
	moonstoneX     [moonstones]byte   // C: D_5830
	moonstoneY     [moonstones]byte   // C: D_5838
	moonstoneState [moonstones]byte   // C: D_5840 (0 buried, 255 carried)
	moonstoneLevel [moonstones]byte   // C: D_5848 (0 Britannia, 255 Underworld)
	reagents       [reagentTypes]byte // C: D_5850
	harvestDay     [3]byte            // C: D_5858
	partySize      byte               // C: D_585b
	eventFlags     [24]byte           // C: D_585c
	year           uint16             // C: D_5874
	// moveX and moveY are the direction of the command in progress.
	moveX int16 // C: D_5876
	moveY int16 // C: D_5878
	// spellIcon is the letter of the spell in effect shown in the status area.
	spellIcon       byte // C: D_587a
	activeMember    byte // C: D_587b (255 none)
	partyTile       byte // C: D_587c
	month           byte // C: D_587d
	day             byte // C: D_587e
	hour            byte // C: D_587f
	hourShown       byte // C: D_5880
	minute          byte // C: D_5881
	combatRounds    byte // C: D_5882
	stepCounter     byte // C: D_5883
	chimesLeft      byte // C: D_5884 (strokes of the clock still to sound)
	trammel         byte // C: D_5885
	felucca         byte // C: D_5886
	moongateRise    byte // C: D_5887 (the rising moongate's height in pixels)
	karma           byte // C: D_5888
	spare5889       byte // C: D_5889 (unused)
	spare588a       byte // C: D_588a (unused)
	turnCount       byte // C: D_588b
	restCooldown    byte // C: D_588c (hours before resting helps again)
	apparitionMonth byte // C: D_588d (the month of the last apparition)
	protection      byte // C: D_588e (turns of protection left)
	rangeX          byte // C: D_588f
	rangeY          byte // C: D_5890
	// animateFrame makes the next map update move the actors (0 no, 255 no but the wind).
	animateFrame  byte     // C: D_5891
	windDirection byte     // C: D_5892
	mapID         byte     // C: D_5893 (0 Britannia, 1-32 settlements, 33+ dungeons, 255 combat)
	savedMapID    byte     // C: D_5894 (the map a combat was entered from)
	level         byte     // C: D_5895
	mapX          byte     // C: D_5896
	mapY          byte     // C: D_5897
	crosshairOn   byte     // C: D_5898
	crosshairX    byte     // C: D_5899
	crosshairY    byte     // C: D_589a
	chunkX        byte     // C: D_589b (the map's top-left chunk loaded)
	chunkY        byte     // C: D_589c
	attackWeapon  byte     // C: D_589d
	combatant     byte     // C: D_589e (the combatant whose turn it is, 255 none)
	focusBlink    byte     // C: D_589f
	exitDirection byte     // C: D_58a0 (where the party left a combat map)
	combatKind    byte     // C: D_58a1
	combatEvent   byte     // C: D_58a2
	battleWon     byte     // C: D_58a3
	mapShown      byte     // C: D_58a4 (the map view is drawn)
	light         byte     // C: D_58a5
	lightSpell    byte     // C: D_58a6 (turns of magic light)
	torchLight    byte     // C: D_58a7 (turns of torch light)
	combatTargets [32]byte // C: D_58a8
	shrineVisits  [3]byte  // C: D_58c8
	lastSpell     byte     // C: D_58cb
	questsActive  uint16   // C: D_58cc (shrine quests begun)
	questsDone    uint16   // C: D_58ce (shrine quests completed)
	shrineStates  [8]byte  // C: D_58d0
	shrineRunes   [8]byte  // C: D_58d8
	dungeonSeen   [14]byte // C: D_58e0
	// the tiles a settlement had changed (the party's doings), to restore: their places
	// and old tiles
	changedX       [32]byte                                        // C: D_58ee
	changedY       [32]byte                                        // C: D_590e
	changedTile    [32]byte                                        // C: D_592e
	changedCount   byte                                            // C: D_594e
	doorOpen       byte                                            // C: D_594f
	doorX          byte                                            // C: D_5950
	doorY          byte                                            // C: D_5951
	doorTurns      byte                                            // C: D_5952
	shipX          byte                                            // C: D_5953
	shipY          byte                                            // C: D_5954
	sailHeading    byte                                            // C: D_5955 (the way the ship sails on its own, 0 none)
	viewDirty      byte                                            // C: D_5956
	drunkTurns     byte                                            // C: D_5957
	shadowlordHere byte                                            // C: D_5958 (the shadowlord haunting this settlement, 255 none)
	ambushed       byte                                            // C: D_5959
	dungeonMap     [dungeonLevels * dungeonSize * dungeonSize]byte // C: D_595a
	npcKilled      [settlements * 4]byte                           // C: D_5b5a (32 bits a settlement)
	npcMet         [settlements * 4]byte                           // C: D_5bda
	actors         [actorSlots]actor                               // C: D_5c5a
	spare5d5a      uint16                                          // C: D_5d5a (unused)
	spare5d5c      uint16                                          // C: D_5d5c (unused)
	schedules      [npcSlots]schedule                              // C: D_5d5e
	npcs           [npcSlots]npc                                   // C: D_5f5e
	movements      [npcSlots][movementBytes]byte                   // C: D_615e
	movementPos    [npcSlots]int16                                 // C: D_655e
	npcTypes       [npcSlots]byte                                  // C: D_659e
	townMode       byte                                            // C: D_65be
	npcTurn        byte                                            // C: D_65bf
	spare65c0      uint16                                          // C: D_65c0 (unused)
	npcWaits       [npcSlots]int16                                 // C: D_65c2
	dungeonMode    byte                                            // C: D_6602
	facing         byte                                            // C: D_6603 (the way the party faces in a dungeon)
	dungeonLook    byte                                            // C: D_6604 (the current dungeon's tile set)
	boughtShip     byte                                            // C: D_6605
)
