package game

import (
	"github.com/memmaker/u5-go/internal/host"
	"github.com/memmaker/u5-go/internal/music"
)

// The music is no part of the PC original: the songs are chosen as the Exodus Project's
// Ultima V Upgrade chooses them (its MIDI driver, midi.asm), by the screen showing and, in
// the game, by the map, the level, the party's ship and the battle.

// Maps by kind, as the upgrade groups them.
const (
	lastCityMap       = 8
	lastLighthouseMap = 12
	lastHutMap        = 16
	britishCastleMap  = 17
	blackthornMap     = 18
	lastTownMap       = 24
	lastKeepMap       = 29
	lastCastleMap     = 32
)

// The party's ship tiles, without their facing.
const (
	shipFacingMask = 0xf8
	frigateTiles   = 0x20
)

// The story pages and ending pages by the upgrade's song changes.
const (
	lastSummoningPage = 7
	lastArrivalPage   = 14
	lastTalePage      = 21
	lastStonesEnding  = 3
)

// musicFollowsMap says whether the game's place picks the song (the upgrade's MAPDETECT).
var musicFollowsMap bool

func init() {
	host.AddTimer(func() {
		if musicFollowsMap {
			music.Play(songForMap())
		}
	})
}

// playSong plays song n for the screen showing, until the map picks again.
func playSong(n int) {
	musicFollowsMap = false
	music.Play(n)
}

// followMap lets the place the party is in pick the song.
func followMap() { musicFollowsMap = true }

// songForMap is the song for where the party is.
func songForMap() int {
	switch {
	case mapID == combatMapID:
		if battleWon != 0 {
			return music.UltimaTheme
		}
		return music.EngagementAndMelee
	case partyTile&shipFacingMask == frigateTiles:
		return music.Hornpipe
	case mapID == 0:
		if level == 0 {
			return music.BritannicLands
		}
		return music.WorldsBelow
	case mapID <= lastCityMap:
		return music.VillagerTarantella
	case mapID <= lastLighthouseMap:
		return music.DreamOfLadyNan
	case mapID <= lastHutMap:
		return music.GreysonsTale
	case mapID == britishCastleMap:
		return music.MissingMonarch
	case mapID == blackthornMap:
		return music.LordBlackthorn
	case mapID <= lastTownMap:
		return music.GreysonsTale
	case mapID <= lastKeepMap:
		return music.DreamOfLadyNan
	case mapID <= lastCastleMap:
		return music.Fanfare
	}
	return music.HallsOfDoom
}

// songForStory is the song of introduction page n.
func songForStory(n int) int {
	switch {
	case n <= lastSummoningPage:
		return music.Stones
	case n <= lastArrivalPage:
		return music.HallsOfDoom
	case n <= lastTalePage:
		return music.GreysonsTale
	}
	return music.Stop
}

// songForEnding is the song of ending page n.
func songForEnding(n int) int {
	if n <= lastStonesEnding {
		return music.Stones
	}
	return music.DreamOfLadyNan
}
