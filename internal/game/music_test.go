package game

import (
	"testing"

	"github.com/memmaker/u5-go/internal/music"
)

func TestSongForMap(t *testing.T) {
	defer func(m, l, p, b byte) { mapID, level, partyTile, battleWon = m, l, p, b }(mapID, level, partyTile, battleWon)
	for _, c := range []struct {
		mapID, level, tile, won byte
		want                    int
	}{
		{0, 0, actorAvatar, 0, music.BritannicLands},
		{0, 1, actorAvatar, 0, music.WorldsBelow},
		{0, 0, frigateTiles + 2, 0, music.Hornpipe},
		{combatMapID, 0, actorAvatar, 0, music.EngagementAndMelee},
		{combatMapID, 0, actorAvatar, 1, music.UltimaTheme},
		{1, 0, actorAvatar, 0, music.VillagerTarantella},
		{britishCastleMap, 0, actorAvatar, 0, music.MissingMonarch},
		{blackthornMap, 0, actorAvatar, 0, music.LordBlackthorn},
		{lastCastleMap, 0, actorAvatar, 0, music.Fanfare},
		{firstDungeonMap, 0, actorAvatar, 0, music.HallsOfDoom},
	} {
		mapID, level, partyTile, battleWon = c.mapID, c.level, c.tile, c.won
		if got := songForMap(); got != c.want {
			t.Errorf("map %d level %d tile %#x won %d: song %d, want %d", c.mapID, c.level, c.tile, c.won, got, c.want)
		}
	}
}
