package game

// The modern layout's wider map (no part of the original): the original's 11x11 view in the
// middle, as drawn, and around it what the party would see of the loaded map by day, worked
// out by a line of sight through the tiles that let sight through.

import (
	"github.com/memmaker/u5-go/internal/ega"
	"github.com/memmaker/u5-go/internal/host"
)

func init() { host.WideView = wideView }

// wideView returns cols x rows tiles of map around the party as palette indexes, nil where
// the original view must do: in combat, below ground, and when it is not full day.
func wideView(cols, rows int) []byte {
	if mapShown == 0 || mapID >= firstMapCombat || inDungeon() || int(light) < dayLight || host.ShowFrame == nil {
		return nil
	}
	frame := host.ShowFrame()
	w, h := cols*ega.TileSize, rows*ega.TileSize
	out := make([]byte, w*h)
	actorsAt := map[[2]int]int{}
	for slot := 1; slot < actorSlots; slot++ {
		a := &actors[slot]
		if a.tile != 0 && a.animTile != 0 && a.z == level {
			actorsAt[[2]int{int(a.x), int(a.y)}] = int(a.animTile)
		}
	}
	for row := range rows {
		for col := range cols {
			dx, dy := col-cols/2, row-rows/2
			left, top := col*ega.TileSize, row*ega.TileSize
			if abs(dx) <= viewCentre && abs(dy) <= viewCentre {
				sx := mapWindowLeft + (dx+viewCentre)*ega.TileSize
				sy := mapWindowTop + (dy+viewCentre)*ega.TileSize
				for y := range ega.TileSize {
					copy(out[(top+y)*w+left:(top+y)*w+left+ega.TileSize], frame.Pixels[(sy+y)*ega.Width+sx:])
				}
				continue
			}
			x, y := int(mapX)+dx, int(mapY)+dy
			if !loaded(x, y) || !inSight(dx, dy) {
				continue
			}
			tile := int(tileShown[*tileAt(x, y)])
			if a, ok := actorsAt[[2]int{x & worldMask(), y & worldMask()}]; ok {
				tile = actorTileBase + a
			}
			for py := range ega.TileSize {
				for px := range ega.TileSize {
					out[(top+py)*w+left+px] = screen.TilePixel(tile, px, py)
				}
			}
		}
	}
	return out
}

// worldMask wraps overworld coordinates; settlements do not wrap.
func worldMask() int {
	if mapID == 0 {
		return 0xff
	}
	return -1
}

// loaded says whether map position (x, y) is in the map held in memory.
func loaded(x, y int) bool {
	if mapID != 0 {
		return x >= 0 && y >= 0 && x < settlementSize && y < settlementSize
	}
	rx, ry := (x-int(chunkX))&0xff, (y-int(chunkY))&0xff
	return rx < settlementSize && ry < settlementSize
}

// inSight walks a line from the party to (dx, dy) and says whether every tile before it
// lets sight through.
func inSight(dx, dy int) bool {
	steps := max(abs(dx), abs(dy))
	for i := 1; i < steps; i++ {
		x := int(mapX) + floorDiv(dx*i*2+steps, 2*steps)
		y := int(mapY) + floorDiv(dy*i*2+steps, 2*steps)
		if !loaded(x, y) || !letsSightThrough(*tileAt(x, y), i) {
			return false
		}
	}
	return true
}

// floorDiv divides rounding down.
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}
