package game

import (
	"encoding/binary"
	"errors"
)

// saveSize is the size of INIT.GAM and SAVED.GAM: the block of the data segment the
// original reads and writes whole.
const saveSize = 4192

var errSaveSize = errors.New("saved game has the wrong size")

// saveCursor walks a saved game's bytes, reading into the state or writing it out, so
// that one list of the fields (visitSave) serves both directions.
type saveCursor struct {
	buf     []byte
	at      int
	writing bool
}

func (c *saveCursor) u8(v *byte) {
	if c.writing {
		c.buf[c.at] = *v
	} else {
		*v = c.buf[c.at]
	}
	c.at++
}

func (c *saveCursor) u16(v *uint16) {
	if c.writing {
		binary.LittleEndian.PutUint16(c.buf[c.at:], *v)
	} else {
		*v = binary.LittleEndian.Uint16(c.buf[c.at:])
	}
	c.at += 2
}

func (c *saveCursor) s16(v *int16) {
	u := uint16(*v)
	c.u16(&u)
	*v = int16(u)
}

func (c *saveCursor) bytes(b []byte) {
	if c.writing {
		copy(c.buf[c.at:], b)
	} else {
		copy(b, c.buf[c.at:])
	}
	c.at += len(b)
}

func (c *saveCursor) member(m *member) {
	c.bytes(m.name[:])
	for _, b := range []*byte{&m.gender, &m.class, &m.status, &m.str, &m.dex, &m.intel, &m.magic} {
		c.u8(b)
	}
	c.s16(&m.hp)
	c.s16(&m.maxHP)
	c.s16(&m.exp)
	c.u8(&m.level)
	c.u8(&m.monthsInInn)
	c.u8(&m.innSlot)
	c.bytes(m.equip[:])
	c.u8(&m.location)
}

func (c *saveCursor) actor(a *actor) {
	for _, b := range []*byte{&a.tile, &a.animTile, &a.x, &a.y, &a.z, &a.hull, &a.state, &a.target} {
		c.u8(b)
	}
}

func (c *saveCursor) schedule(s *schedule) {
	c.bytes(s.kind[:])
	c.bytes(s.x[:])
	c.bytes(s.y[:])
	c.bytes(s.z[:])
	c.bytes(s.time[:])
}

func (c *saveCursor) npc(n *npc) {
	c.u16(&n.moveState)
	c.s16(&n.x)
	c.s16(&n.y)
	for _, w := range []*uint16{&n.z, &n.shape, &n.mood, &n.actorAt, &n.place} {
		c.u16(w)
	}
}

// visitSave lists the saved fields in their file order.
// C: FILE_ReadSavegameFile, FILE_WriteSavegameFile
func visitSave(c *saveCursor) {
	c.u16(&saveTag)
	for i := range party {
		c.member(&party[i])
	}
	c.s16(&food)
	c.s16(&gold)
	for _, b := range []*byte{&keys, &gems, &torches, &grapple, &magicCarpets, &skullKeys, &skullKeyDay, &amulet, &crown, &sceptre} {
		c.u8(b)
	}
	c.bytes(shards[:])
	for _, b := range []*byte{&spyglasses, &hmsCapePlans, &sextants, &pocketWatch, &blackBadge, &sandalwoodBox} {
		c.u8(b)
	}
	for _, b := range [][]byte{equipment[:], spellMixtures[:], scrolls[:], potions[:], moonstoneX[:], moonstoneY[:],
		moonstoneState[:], moonstoneLevel[:], reagents[:], harvestDay[:]} {
		c.bytes(b)
	}
	c.u8(&partySize)
	c.bytes(eventFlags[:])
	c.u16(&year)
	c.s16(&moveX)
	c.s16(&moveY)
	for _, b := range []*byte{&spellIcon, &activeMember, &partyTile, &month, &day, &hour, &hourShown, &minute,
		&combatRounds, &stepCounter, &chimesLeft, &trammel, &felucca, &moongateRise, &karma, &spare5889, &spare588a,
		&turnCount, &restCooldown, &apparitionMonth, &protection, &rangeX, &rangeY, &animateFrame, &windDirection,
		&mapID, &savedMapID, &level, &mapX, &mapY, &crosshairOn, &crosshairX, &crosshairY, &chunkX, &chunkY,
		&attackWeapon, &combatant, &focusBlink, &exitDirection, &combatKind, &combatEvent, &battleWon, &mapShown,
		&light, &lightSpell, &torchLight} {
		c.u8(b)
	}
	c.bytes(combatTargets[:])
	c.bytes(shrineVisits[:])
	c.u8(&lastSpell)
	c.u16(&questsActive)
	c.u16(&questsDone)
	for _, b := range [][]byte{shrineStates[:], shrineRunes[:], dungeonSeen[:], changedX[:], changedY[:], changedTile[:]} {
		c.bytes(b)
	}
	for _, b := range []*byte{&changedCount, &doorOpen, &doorX, &doorY, &doorTurns, &shipX, &shipY, &sailHeading,
		&viewDirty, &drunkTurns, &shadowlordHere, &ambushed} {
		c.u8(b)
	}
	c.bytes(dungeonMap[:])
	c.bytes(npcKilled[:])
	c.bytes(npcMet[:])
	for i := range actors {
		c.actor(&actors[i])
	}
	c.u16(&spare5d5a)
	c.u16(&spare5d5c)
	for i := range schedules {
		c.schedule(&schedules[i])
	}
	for i := range npcs {
		c.npc(&npcs[i])
	}
	for i := range movements {
		c.bytes(movements[i][:])
	}
	for i := range movementPos {
		c.s16(&movementPos[i])
	}
	c.bytes(npcTypes[:])
	c.u8(&townMode)
	c.u8(&npcTurn)
	c.u16(&spare65c0)
	for i := range npcWaits {
		c.s16(&npcWaits[i])
	}
	for _, b := range []*byte{&dungeonMode, &facing, &dungeonLook, &boughtShip} {
		c.u8(b)
	}
}

// loadSave sets the saved state from a saved game file's contents.
func loadSave(contents []byte) error {
	if len(contents) < saveSize {
		return errSaveSize
	}
	visitSave(&saveCursor{buf: contents})
	return nil
}

// encodeSave returns the saved state as a saved game file's contents.
func encodeSave() []byte {
	c := &saveCursor{buf: make([]byte, saveSize), writing: true}
	visitSave(c)
	return c.buf
}

// actorRecordSize is the size of one actor in a saved file.
const actorRecordSize = 8

// encodeActors returns the actor table as the original stores it in files (BRIT.OOL and
// UNDER.OOL).
func encodeActors(list *[actorSlots]actor) []byte {
	c := saveCursor{buf: make([]byte, actorSlots*actorRecordSize), writing: true}
	for i := range list {
		c.actor(&list[i])
	}
	return c.buf
}

// decodeActors reads an actor table stored by encodeActors.
func decodeActors(contents []byte, list *[actorSlots]actor) {
	c := saveCursor{buf: contents}
	for i := range list {
		c.actor(&list[i])
	}
}

// actorBytes returns the actor table for writing to a file.
func actorBytes() []byte { return encodeActors(&actors) }

// readActorBytes reads the actor table from file name.
func readActorBytes(name string) {
	buffer := make([]byte, actorSlots*actorRecordSize)
	readFile(name, buffer, 0)
	decodeActors(buffer, &actors)
}
