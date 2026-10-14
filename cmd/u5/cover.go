package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/memmaker/u5-go/internal/data"
	"github.com/memmaker/u5-go/internal/ega"
)

const (
	// the title: ULTIMA.16's first picture, the "Ultima" logo, and the banner's first frame
	titlePictures = "ULTIMA.16"
	bannerY       = 100
	vgaMaxLevel   = 63
	paletteSize   = 16
)

// writeCover writes dir/COVER.PNG, the picture goplay shows on the game's floppy label: the
// title screen's logo and banner.
func writeCover(dir string) error {
	disk := data.NewDisk(dir)
	disk.SetCompressed([]string{titlePictures})
	driver, err := disk.ReadFile("T1K.DRV")
	if err != nil {
		return err
	}
	screen, err := ega.New(driver)
	if err != nil {
		return err
	}
	pictures, err := disk.LoadResource(titlePictures)
	if err != nil {
		return err
	}
	screen.SetPage(ega.Screen)
	screen.PutPicture(pictures, 0, 0, 0, 0)
	screen.PutPicture(pictures, 1, 16, bannerY, 0)
	vga := screen.Palette()
	palette := make(color.Palette, paletteSize)
	for i := range palette {
		level := func(c int) uint8 { return uint8(int(vga[i*3+c]) * math.MaxUint8 / vgaMaxLevel) }
		palette[i] = color.RGBA{level(0), level(1), level(2), math.MaxUint8}
	}
	img := &image.Paletted{Pix: screen.Pixels(ega.Screen), Stride: ega.Width, Rect: image.Rect(0, 0, ega.Width, ega.Height), Palette: palette}
	f, err := os.Create(filepath.Join(dir, "COVER.PNG"))
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
