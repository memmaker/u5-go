// Package ega is the video driver the game loads (EGA.DRV and its siblings, here after
// u5d's port of T1K.DRV): two 320x200 pages of 16-colour pixels, the drawing primitives,
// the 16x16 tile set with its animations, the pictures of the .16 files, and the special
// effects of the intro (the dissolving reveal, the Origin logo, the "Warriors of Destiny"
// banner). Page 0 is the screen; page 1 is the back buffer.
//
// A few tables of the driver itself (the effects' sequences, the clock hands, the colour
// maps) are read from T1K.DRV at start, see LoadDriver.
package ega

import (
	"time"

	"github.com/memmaker/u5-go/internal/host"
)

const (
	Width     = 320
	Height    = 200
	pageCount = 2
	// Screen and BackBuffer are the page numbers.
	Screen     = 0
	BackBuffer = 1

	colourCount = 16
	colourMask  = colourCount - 1
	// White is the colour 1-bit pictures are drawn in.
	White = 15
)

// egaPalette is the EGA's 16 colours as 6-bit VGA DAC levels.
var egaPalette = [colourCount][3]byte{
	{0, 0, 0}, {0, 0, 42}, {0, 42, 0}, {0, 42, 42}, {42, 0, 0}, {42, 0, 42}, {42, 21, 0}, {42, 42, 42},
	{21, 21, 21}, {21, 21, 63}, {21, 63, 21}, {21, 63, 63}, {63, 21, 21}, {63, 21, 63}, {63, 63, 21}, {63, 63, 63},
}

// Display is the driver's state.
type Display struct {
	pages [pageCount][]byte
	// Page is the page the primitives draw on (D_52d8).
	Page int
	pen  byte
	// tiles is the loaded tile set: tileBytes bytes per tile, two pixels a byte.
	tiles   []byte
	driver  *driverTables
	effects effectState
	palette [768]byte
	frame   host.Frame
	// KeyHit reports (and takes) a key press; the long effects stop at one. Set by the game.
	KeyHit func() bool
}

// New makes the display; driver is the contents of T1K.DRV.
func New(driver []byte) (*Display, error) {
	tables, err := loadDriverTables(driver)
	if err != nil {
		return nil, err
	}
	d := &Display{driver: tables, KeyHit: func() bool { return false }}
	for i := range d.pages {
		d.pages[i] = make([]byte, Width*Height)
	}
	for c, rgb := range egaPalette {
		copy(d.palette[c*3:], rgb[:])
	}
	d.effects.tileRevealSeed = 0
	d.effects.tileEffectSeed = tables.tileEffectSeed
	d.frame = host.Frame{Pixels: d.pages[Screen], Palette: &d.palette}
	return d, nil
}

// Frame is what the host shows: page 0.
func (d *Display) Frame() *host.Frame { return &d.frame }

// Pixels returns page's pixels (row by row, one colour a byte), for tests and screenshots.
func (d *Display) Pixels(page int) []byte { return d.pages[page] }

// Palette returns the EGA colours as 6-bit levels.
func (d *Display) Palette() *[768]byte { return &d.palette }

// SetPenColor sets the colour of lines and fills.
// C: DRV_2d
func (d *Display) SetPenColor(colour int) { d.pen = byte(colour) }

// SetPage picks the page drawn on.
// C: DRV_0f
func (d *Display) SetPage(page int) { d.Page = page }

func (d *Display) put(page, x, y int, colour byte) { d.pages[page][y*Width+x] = colour }

func (d *Display) putClipped(page, x, y int, colour byte) {
	if x >= 0 && x < Width && y >= 0 && y < Height {
		d.put(page, x, y, colour)
	}
}

// PrintChar draws the 8x8 character at offset of font (8 bytes, a bit a pixel, high bit
// left) at text cell (column, row) in the colours fg and bg.
// C: DRV_5d
func (d *Display) PrintChar(font []byte, offset int, fg, bg byte, column, row int) {
	const cell = 8
	for y := range cell {
		bits := font[offset+y]
		for x := range cell {
			colour := bg
			if bits&(highBit>>x) != 0 {
				colour = fg
			}
			d.put(d.Page, column*cell+x, row*cell+y, colour)
		}
	}
}

const highBit = 128

// ScrollWindow moves the rectangle of page 0 up (amount < 0) or down by |amount| pixels,
// clearing the rows uncovered.
// C: DRV_27
func (d *Display) ScrollWindow(left, top, right, bottom, amount int) {
	screen := d.pages[Screen]
	row := func(y int) []byte { return screen[y*Width+left : y*Width+right+1] }
	if amount < 0 {
		amount = -amount
		for y := top; y <= bottom-amount; y++ {
			copy(row(y), row(y+amount))
		}
		for y := bottom - amount + 1; y <= bottom; y++ {
			clear(row(y))
		}
		return
	}
	for y := bottom; y >= top+amount; y-- {
		copy(row(y), row(y-amount))
	}
	for y := top; y < top+amount; y++ {
		clear(row(y))
	}
}

// FillRect fills the rectangle (inclusive) with the pen colour, or with xor mixes the pen
// colour into it.
// C: DRV_3f
func (d *Display) FillRect(x1, y1, x2, y2 int, xor bool) {
	if y1 > y2 || x1 > x2 {
		return
	}
	page := d.pages[d.Page]
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			if xor {
				page[y*Width+x] = (page[y*Width+x] ^ d.pen) & colourMask
			} else {
				page[y*Width+x] = d.pen
			}
		}
	}
}

// Line draws a line in the pen colour (Bresenham).
// C: DRV_33
func (d *Display) Line(x1, y1, x2, y2 int) {
	dx, sx := abs(x2-x1), sign(x2-x1)
	dy, sy := -abs(y2-y1), sign(y2-y1)
	diff := dx + dy
	for {
		d.put(d.Page, x1, y1, d.pen)
		twice := 2 * diff
		if twice >= dy {
			if x1 == x2 {
				break
			}
			diff += dy
			x1 += sx
		}
		if twice <= dx {
			if y1 == y2 {
				break
			}
			diff += dx
			y1 += sy
		}
	}
}

// HorizLine and VertLine are the driver's straight-line entries.
// C: DRV_39
func (d *Display) HorizLine(x1, y, x2 int) { d.Line(x1, y, x2, y) }

// C: DRV_3c
func (d *Display) VertLine(x, y1, y2 int) { d.Line(x, y1, x, y2) }

// Plot sets one pixel in the pen colour.
// C: DRV_30
func (d *Display) Plot(x, y int) { d.putClipped(d.Page, x, y, d.pen) }

// Copy copies the rectangle (x1,y1)-(x2,y2) of page from to (toX, toY) of page to.
// C: GRAP_BUF_TransferPage
func (d *Display) Copy(from, to, x1, y1, x2, y2, toX, toY int) {
	for y := 0; y <= y2-y1; y++ {
		copy(d.pages[to][(y+toY)*Width+toX:(y+toY)*Width+toX+x2-x1+1], d.pages[from][(y+y1)*Width+x1:])
	}
}

// CopyArea copies a rectangle between the pages at the same place: from the back buffer
// to the screen when toScreen, else the other way.
// C: DRV_18
func (d *Display) CopyArea(x1, y1, x2, y2 int, toScreen bool) {
	if toScreen {
		d.Copy(BackBuffer, Screen, x1, y1, x2, y2, x1, y1)
	} else {
		d.Copy(Screen, BackBuffer, x1, y1, x2, y2, x1, y1)
	}
}

// CopyPage copies a whole page.
// C: DRV_1b
func (d *Display) CopyPage(from, to int) {
	if from == Screen && to == BackBuffer {
		d.Copy(Screen, BackBuffer, 0, 0, Width-1, Height-1, 0, 0)
	} else {
		d.Copy(BackBuffer, Screen, 0, 0, Width-1, Height-1, 0, 0)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	if v > 0 {
		return 1
	}
	return -1
}

// present shows page 0 now: the effects draw step by step.
func (d *Display) present() { host.Yield() }

func waitHost(t time.Duration) { host.Wait(t) }
