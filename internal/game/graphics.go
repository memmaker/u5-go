package game

import "github.com/memmaker/u5-go/internal/ega"

// The program's own layer over the video driver (segment 0000): clipping, the pen
// position and the file loaders for pictures and tiles.

const (
	screenRight  = ega.Width - 1
	screenBottom = ega.Height - 1
	// noPen is the pen colour that draws nothing.
	noPen = -1
)

// videoState is the part of the driver parameter block (D_52ba) the program uses.
type videoState struct {
	// tileXOffset and tileYOffset place the map's tiles in the frame (D_52bc, D_52be).
	tileXOffset, tileYOffset int
	// clipping turns on the clip window for lines and points (D_52c4).
	clipping bool
	// penX, penY is where the last line ended (D_52cc, D_52ce).
	penX, penY int
	// the clip window (D_52d0..D_52d6)
	clipLeft, clipRight, clipTop, clipBottom int
	// penColour is the colour set last, noPen for none (D_52da).
	penColour int
	// lineStartX..lineEndY are the clipped line's ends (D_52fa..D_5300); stepX, stepY the
	// last bisection step (D_52f6, D_52f8).
	lineStartX, lineStartY, lineEndX, lineEndY int
	stepX, stepY                               int
}

const (
	defaultTileOffset = 8
)

var video = videoState{
	tileXOffset: defaultTileOffset, tileYOffset: defaultTileOffset,
	clipRight: screenRight, clipBottom: screenBottom,
}

// screen is the video driver.
var screen *ega.Display

// insideClipWindow says whether (x, y) lies in the clip window.
// C: ULTIMA_08ca_InsideClipWindow
func insideClipWindow(x, y int) bool {
	return video.clipLeft <= x && x <= video.clipRight && video.clipTop <= y && y <= video.clipBottom
}

// clipRect orders a rectangle's corners and clamps them to the screen.
// C: ULTIMA_08e6_ClipRectCoord
func clipRect(x1, y1, x2, y2 int) (int, int, int, int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	return clamp(x1, 0, screenRight), clamp(y1, 0, screenBottom), clamp(x2, 0, screenRight), clamp(y2, 0, screenBottom)
}

func clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

// Outcodes: where a point lies off the clip window.
const (
	outLeft   = 1
	outRight  = 2
	outBelow  = 4
	outAbove  = 8
	noOutcode = 0
)

// outcodes returns where the two ends of a line lie off the clip window.
// C: ULTIMA_0a22
func outcodes(x1, y1, x2, y2 int) (int, int) {
	first, second := noOutcode, noOutcode
	if x1 < video.clipLeft {
		first |= outLeft
	}
	if x2 < video.clipLeft {
		second |= outLeft
	}
	if x1 > video.clipRight {
		first |= outRight
	}
	if x2 > video.clipRight {
		second |= outRight
	}
	if y1 > video.clipBottom {
		first |= outBelow
	}
	if y2 > video.clipBottom {
		second |= outBelow
	}
	if y1 < video.clipTop {
		first |= outAbove
	}
	if y2 < video.clipTop {
		second |= outAbove
	}
	return first, second
}

// clipLine clips a line to the clip window; false when nothing of it is to be drawn.
// C: ULTIMA_0935_ClipLineCoord
func clipLine(x1, y1, x2, y2 int) (int, int, int, int, bool) {
	first, second := outcodes(x1, y1, x2, y2)
	if first == noOutcode || second == noOutcode {
		return x1, y1, x2, y2, true
	}
	if first != second {
		return x1, y1, x2, y2, false
	}
	video.lineStartX, video.lineStartY, video.lineEndX, video.lineEndY = x1, y1, x2, y2
	x1, y1, x2, y2 = x2, y2, x1, y1
	x1, y1, x2, y2 = bisectToWindow(x1, y1, x2, y2)
	video.lineStartX, video.lineStartY = x1, y1
	x2, y2 = video.lineEndX, video.lineEndY
	x1, y1, x2, y2 = bisectToWindow(x1, y1, x2, y2)
	video.lineEndX, video.lineEndY = x1, y1
	return video.lineStartX, video.lineStartY, video.lineEndX, video.lineEndY, true
}

// bisectToWindow halves the line from (x1, y1) towards (x2, y2) until its first end
// reaches the clip window's edge.
// C: ULTIMA_0991
func bisectToWindow(x1, y1, x2, y2 int) (int, int, int, int) {
	for {
		for {
			oldX, oldY := x1, y1
			x1 = (x1 + x2) / 2
			y1 = (y1 + y2) / 2
			video.stepX, video.stepY = oldX-x1, oldY-y1
			if first, _ := outcodes(x1, y1, x2, y2); first == noOutcode {
				break
			}
			x2, y2 = x1, y1
			x1 += video.stepX
			y1 += video.stepY
		}
		if (x1 == video.clipLeft || x1 == video.clipRight) && (y1 == video.clipTop || y1 == video.clipBottom) {
			return x1, y1, x2, y2
		}
		if max(abs(video.stepX), abs(video.stepY)) < 1 {
			return x1, y1, x2, y2
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// setPenColour sets the colour of lines, points and fills; noPen draws nothing.
// C: ULTIMA_0a70_GRAP_2d_SetPenColor
func setPenColour(colour int) {
	const colourMask = 15
	if colour != noPen {
		colour &= colourMask
	}
	video.penColour = colour
	if colour != noPen {
		screen.SetPenColor(colour)
	}
}

// fillRect fills a rectangle with the pen colour.
// C: ULTIMA_0aa6_GRAP_3f_FillRect
func fillRect(x1, y1, x2, y2 int) {
	x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
	screen.FillRect(x1, y1, x2, y2, false)
}

// xorFillRect mixes the pen colour into a rectangle.
// C: ULTIMA_0b86_GRAP_XorFillRect
func xorFillRect(x1, y1, x2, y2 int) {
	x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
	screen.FillRect(x1, y1, x2, y2, true)
}

// copyArea copies a rectangle from page from to page to, in place.
// C: ULTIMA_0ace_GRAP_18_TransferArea
func copyArea(from, to, x1, y1, x2, y2 int) {
	if from != to && from <= ega.BackBuffer && to <= ega.BackBuffer {
		x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
		screen.CopyArea(x1, y1, x2, y2, from != ega.Screen)
	}
}

// line draws a line in the pen colour.
// C: ULTIMA_0b10_GRAP_Line, ULTIMA_0b2d_GRAP_Line
func line(x1, y1, x2, y2 int) {
	video.penX, video.penY = x2, y2
	// the shape is decided by the ends given, the drawing done with the clipped ones
	ax, ay, bx, by := x1, y1, x2, y2
	if video.clipping {
		var ok bool
		if ax, ay, bx, by, ok = clipLine(x1, y1, x2, y2); !ok {
			return
		}
	}
	switch {
	case y1 == y2 && x1 == x2:
		screen.Plot(ax, ay)
	case y1 == y2:
		ax, ay, bx, _ = clipRect(ax, ay, bx, by)
		screen.HorizLine(ax, ay, bx)
	case x1 == x2:
		ax, ay, _, by = clipRect(ax, ay, bx, by)
		screen.VertLine(ax, ay, by)
	default:
		screen.Line(ax, ay, bx, by)
	}
}

// lineTo draws a line from where the last one ended.
// C: ULTIMA_0f90_GRAP_Pen
func lineTo(x, y int) { line(video.penX, video.penY, x, y) }

// selectPage picks the page drawn on.
// C: ULTIMA_0c22_GRAP_0f_SelectPage
func selectPage(page int) { screen.SetPage(page) }

// setClipWindow sets the clip window.
// C: ULTIMA_0c3c
func setClipWindow(left, top, right, bottom int) {
	video.clipLeft, video.clipTop, video.clipRight, video.clipBottom = left, top, right, bottom
}

// plot sets a pixel in the pen colour.
// C: ULTIMA_0c64_GRAP_30_Pset
func plot(x, y int) {
	video.penX, video.penY = x, y
	if video.penColour != noPen {
		if video.clipping && !insideClipWindow(x, y) {
			return
		}
		screen.Plot(x, y)
	}
}

// horizLine draws a horizontal line.
// C: ULTIMA_0c9c_GRAP_39_HorizLine
func horizLine(x1, y, x2 int) {
	video.penX = x2
	if video.clipping && !horizInClip(x1, x2) {
		return
	}
	screen.HorizLine(x1, y, x2)
}

// horizInClip: C: ULTIMA_0ccd
func horizInClip(x1, x2 int) bool {
	if x1 < x2 {
		x1, x2 = x2, x1
	}
	return x1 >= video.clipLeft && x2 <= video.clipRight && x1 >= 0 && x1 <= screenRight && x2 <= screenRight
}

// vertLine draws a vertical line.
// C: ULTIMA_0cf2_GRAP_3c_VertLine
func vertLine(x, y1, y2 int) {
	video.penY = y2
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	if !video.clipping || vertInClip(y1, y2) {
		screen.VertLine(x, y1, y2)
	}
}

// vertInClip: C: ULTIMA_0d2b
func vertInClip(y1, y2 int) bool {
	return video.clipTop <= y1 && y2 <= video.clipBottom && 0 <= y1 && y1 <= screenBottom && y2 <= screenBottom
}

// putPicture draws picture index of file at (x, y).
// C: ULTIMA_0d4c_GRAP_4b_PutImage
func putPicture(file ega.Pictures, index, x, y, flags int) {
	screen.PutPicture(file, index, x, y, flags)
}

// putPictureMirrored draws a picture flipped left to right.
// C: ULTIMA_0bfc_GRAP_63
func putPictureMirrored(file ega.Pictures, index, x, y, flags int) {
	screen.PutPictureMirrored(file, index, x, y, flags)
}

// putBitPicture draws a 1-bit picture in white.
// C: ULTIMA_1044_GRAP_4e_CopyBitImageIntoPage
func putBitPicture(file ega.Pictures, index, x, y int) {
	screen.PutBitPicture(file, index, x, y)
}

// reveal dissolves a rectangle of the back buffer onto the screen.
// C: ULTIMA_0f46_GRAP_66_Reveal
func reveal(x1, y1, x2, y2 int) {
	x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
	screen.Reveal(x1, y1, x2, y2)
}

// copyPage copies a whole page.
// C: ULTIMA_0f6e_GRAP_1b_TransferFullscreen
func copyPage(from, to int) {
	if from != to {
		screen.CopyPage(from, to)
	}
}

// loadPictures reads a picture file, unpacking it as needed.
// C: ULTIMA_0bae_LoadImageFile, ULTIMA_0fae_LoadResourceFile
func loadPictures(name string) ega.Pictures {
	contents, err := disk.LoadResource(name)
	if err != nil {
		diskError(err)
		return nil
	}
	return contents
}

// loadTiles reads a tile set and gives it to the driver.
// C: ULTIMA_0ff4_LoadTileset
func loadTiles(name string) bool {
	tiles, err := disk.LoadResource(name)
	if err != nil {
		diskError(err)
		return false
	}
	screen.LoadTiles(tiles)
	return true
}

// titleMap is the map number the title sequence runs with.
const titleMap = 66

// revealTileSteps is the number of steps of a tile's pixel-by-pixel reveal.
const revealTileSteps = 256

// revealTile draws tile at view cell (x, y) a pixel group at a time, updating the frame
// (or, in the title sequence, its animation) every eight steps. It returns what the title
// animation returned, 0 when the reveal finished.
// C: ULTIMA_1068
func revealTile(tile, x, y int) int {
	step := 0
	for {
		for {
			for {
				screen.PutTileRevealStep(x, y, tile, step, video.tileXOffset, video.tileYOffset)
				step++
				screen.PutTileRevealStep(x, y, tile, step, video.tileXOffset, video.tileYOffset)
				step++
				if step == revealTileSteps {
					return 0
				}
				if step%8 == 0 {
					break
				}
			}
			if mapID == titleMap {
				break
			}
			updateFrame()
		}
		if result := animateTitle(1); result != 0 {
			return result
		}
	}
}

// putTile draws tile at view cell (x, y).
// C: ULTIMA_10e0_GRAP_51_PutTile
func putTile(tile, x, y int) {
	screen.PutTile(x, y, tile, video.tileXOffset, video.tileYOffset)
}

// putMoongate draws a moongate risen visibleRows high at view cell (x, y).
// C: ULTIMA_1112_GRAP_60
func putMoongate(visibleRows, x, y int) {
	const brickFloorMap = 255
	screen.PutMoongate(x, y, visibleRows, mapID == brickFloorMap, video.tileXOffset, video.tileYOffset)
}
