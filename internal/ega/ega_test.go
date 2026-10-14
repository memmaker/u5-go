package ega

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/memmaker/u5-go/internal/data"
)

func gameDir(t *testing.T) string {
	dir := os.Getenv("U5DATA")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Projects/u5-re/gog")
	}
	if _, err := os.Stat(filepath.Join(dir, "T1K.DRV")); err != nil {
		t.Skip("no game data")
	}
	return dir
}

func newDisplay(t *testing.T) (*Display, *data.Disk) {
	disk := data.NewDisk(gameDir(t))
	drv, err := disk.ReadFile("T1K.DRV")
	if err != nil {
		t.Fatal(err)
	}
	d, err := New(drv)
	if err != nil {
		t.Fatal(err)
	}
	ovl, _ := disk.ReadFile("DATA.OVL")
	const compressedFileTable, compressedFiles = 9706, 29
	disk.SetCompressed(data.NewSegment(ovl).Strings(compressedFileTable, compressedFiles))
	return d, disk
}

// TestTileSheet draws the tile set; with U5_TILESHEET set it writes it there as a PNG
// (32 tiles a row), for looking the tiles up.
func TestTileSheet(t *testing.T) {
	d, disk := newDisplay(t)
	tiles, err := disk.LoadResource("TILES.16")
	if err != nil {
		t.Fatal(err)
	}
	d.LoadTiles(tiles)
	d.AnimateTiles()
	const perRow, count = 32, 512
	sheet := image.NewPaletted(image.Rect(0, 0, perRow*TileSize*2, count/perRow*TileSize*2), nil)
	for c, rgb := range egaPalette {
		sheet.Palette = append(sheet.Palette, color.RGBA{rgb[0] * 4, rgb[1] * 4, rgb[2] * 4, 255})
		_ = c
	}
	for tile := range count {
		clear(d.pages[0])
		d.PutTile(0, 0, tile, 0, 0)
		for y := range TileSize * 2 {
			for x := range TileSize * 2 {
				sheet.SetColorIndex(tile%perRow*TileSize*2+x, tile/perRow*TileSize*2+y, d.pages[0][y/2*Width+x/2])
			}
		}
	}
	if out := os.Getenv("U5_TILESHEET"); out != "" {
		f, err := os.Create(out)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		png.Encode(f, sheet)
	}
}
