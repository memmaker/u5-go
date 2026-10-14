package game

// fighter is one fighter on a combat map (CombatEntity, D_ba14).
type fighter struct {
	hp        byte
	dex       byte
	flags     byte
	entity    byte // the party member or monster type it is
	actorSlot byte // its actor (in actors)
	turnTimer byte
	x, y      byte
}

// Bits of a combatant's flags.
const (
	fighterPlayer    = 128
	fighterMonster   = 64
	fighterDead      = 32
	fighterInvisible = 16
	fighterAsleep    = 8
	fighterFlag4     = 4
	fighterFlag2     = 2
	fighterCharmed   = 1
)

const combatantSlots = 32

// noCombatant marks that no one's turn it is.
const noCombatant = 255

var combatants [combatantSlots]fighter // C: D_ba14
