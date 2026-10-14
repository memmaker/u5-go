package game

import (
	"github.com/memmaker/u5-go/internal/ega"
	"github.com/memmaker/u5-go/internal/host"
)

// Text output (segment 1000): four text windows on a 40x25 grid of 8x8 cells, printed in
// one of four fonts, with underline, inverse and centred modes, and word-wrapped strings.

const (
	textColumns = 40
	textRows    = 25
	cellSize    = 8
	windowCount = 4
	fontCount   = 4
	glyphBytes  = 8
)

// Control characters PrintChar understands.
const (
	charClearWindow     = 255
	charToggleUnderline = 254
	charToggleInverse   = 253
	charCentre          = 252
	charUncentre        = 251
	charNewline         = '\n'
	charReturn          = '\r'
	charSpace           = ' '
	asciiMask           = 127
	lastASCII           = 127
)

// Bits of a text window's flags.
const (
	windowUnderline = 1
	windowCentred   = 2
	windowInverse   = 4
)

// textWindow is one text window (D_535e): its cells (inclusive), the cursor in it, its
// colours (background high nibble, foreground low) and its modes.
type textWindow struct {
	left, top, right, bottom int
	x, y                     int
	colours                  byte
	flags                    byte
}

const defaultWindowColours = 15

var (
	textWindows [windowCount]textWindow
	// windowIndex and window are the window printed in (D_5386, D_539a).
	windowIndex int
	window      *textWindow
	// the current window's modes and colours, unpacked (D_53a4, D_53a6, D_53a8, D_53aa,
	// D_53ab)
	underline, centred, inverse bool
	textFg, textBg              byte
	// fonts are the loaded fonts (D_539c); font the one printed in (D_5398, D_5388).
	fonts     [fontCount][]byte
	font      []byte
	fontIndex int
	// glyphBackup keeps a glyph while the effects change it for printing (D_53ea).
	glyphBackup [glyphBytes]byte
	// advanceCursor makes PrintChar move the cursor on (D_538e).
	advanceCursor = true
	// the blinking cursor: its first glyph, glyph count and frame (D_5390, D_5392, D_540c)
	cursorGlyph      = 5
	cursorGlyphCount = 4
	cursorFrame      int
)

// initTextWindows sets every window to the whole screen, white on black.
// C: ULTIMA_1184_InitTextWindow, ULTIMA_1226_InitTextVars
func initTextWindows() {
	for i := range textWindows {
		textWindows[i] = textWindow{right: textColumns - 1, bottom: textRows - 1, colours: defaultWindowColours}
	}
	windowIndex = 0
	window = &textWindows[0]
	underline, centred, inverse = false, false, false
}

// windowPixels returns the window's rectangle in pixels.
// C: ULTIMA_1f77_ConvertCharCoordToPixel
func windowPixels(w *textWindow) (int, int, int, int) {
	return w.left * cellSize, w.top * cellSize, w.right*cellSize + cellSize - 1, w.bottom*cellSize + cellSize - 1
}

// printChar prints a character or obeys a control character in the current window.
// C: ULTIMA_16ba_PrintChar
func printChar(c int) {
	if pendingEcho != noEcho && windowIndex == windowScroll {
		flushEcho()
	}
	glyph := byte(c)
	w := window
	if c > lastASCII {
		switch c {
		case charClearWindow:
			w.x, w.y = 0, 0
			screen.SetPenColor(int(w.colours >> 4))
			x1, y1, x2, y2 := windowPixels(w)
			screen.FillRect(x1, y1, x2, y2, false)
			screen.SetPenColor(video.penColour)
			return
		case charToggleUnderline:
			underline = !underline
			w.flags ^= windowUnderline
			return
		case charToggleInverse:
			inverse = !inverse
			w.flags ^= windowInverse
			return
		case charCentre:
			centred = true
			w.flags |= windowCentred
			return
		case charUncentre:
			centred = false
			w.flags &^= windowCentred
			return
		}
		c &= asciiMask
	}
	if windowIndex == windowScroll {
		historyPut(c)
	}
	if c != charNewline {
		if c == charReturn {
			carriageReturn(w)
			return
		}
		offset := int(glyph) * glyphBytes
		applyGlyphEffects(offset)
		screen.PrintChar(font, offset, textFg, textBg, w.x+w.left, w.y+w.top)
		copy(font[offset:offset+glyphBytes], glyphBackup[:])
		w = window
		if !advanceCursor {
			return
		}
		w.x++
		if int(int8(w.x+w.left)) <= w.right {
			return
		}
	}
	w.y++
	carriageReturn(w)
}

// carriageReturn moves the cursor to the start of its row, scrolling the window up when
// the cursor went below it.
func carriageReturn(w *textWindow) {
	w.x = 0
	if w.bottom < w.y+w.top {
		x1, y1, x2, y2 := windowPixels(w)
		w.y--
		screen.ScrollWindow(x1, y1, x2, y2, -cellSize)
	}
}

// applyGlyphEffects saves the glyph at offset, then underlines or inverts it as the modes
// say; printChar puts it back after drawing.
// C: ULTIMA_17f4_ApplyCharEffects
func applyGlyphEffects(offset int) {
	glyph := font[offset : offset+glyphBytes]
	copy(glyphBackup[:], glyph)
	const allSet = 255
	if underline {
		glyph[glyphBytes-1] = allSet
	}
	if inverse {
		for i := range glyph {
			glyph[i] = ^glyph[i]
		}
	}
}

// byteAt returns s[i], or 0 past its end (C strings end in a zero).
func byteAt(s string, i int) byte {
	if i < 0 || i >= len(s) {
		return 0
	}
	return s[i]
}

// printString prints a string in the current window, wrapping it at word boundaries (and
// centring the lines in centred mode).
// C: ULTIMA_1850_PrintString
func printString(s string) {
	const lineBuffer = 40
	if len(s) == 0 {
		return
	}
	w := &textWindows[windowIndex]
	centring := w.flags&windowCentred != 0
	width := w.right - w.left
	pos, done, wrapped := 0, false, false
	var line [lineBuffer + 1]byte
	isBreak := func(c byte) bool { return c == charNewline || c == charReturn || c == 0 }
	for !done {
		if wrapped {
			printChar(charNewline)
		}
		wrapped = false
		skipSpaces, first := false, 0
		room := width - window.x
		n := 0
		for ; !isBreak(byteAt(s, pos)) && n <= room; pos, n = pos+1, n+1 {
			line[n] = byteAt(s, pos)
		}
		if n == 0 {
			if byteAt(s, pos) != 0 {
				printChar(int(byteAt(s, pos)))
				pos++
			} else {
				done = true
			}
			continue
		}
		if n > room {
			fullN, fullPos := n, pos
			for ; byteAt(s, pos) != ' ' && n != 0 && !isBreak(byteAt(s, pos)); n, pos = n-1, pos-1 {
			}
			if n == 0 {
				n, pos = fullN, fullPos
				if w.x != 0 {
					printChar(charNewline)
					skipSpaces = true
				}
			} else {
				for {
					n--
					if c := line[n]; c != ' ' && c != charReturn && c != charNewline {
						break
					}
				}
				if byteAt(s, pos) != 0 {
					pos++
				}
				if n < room {
					wrapped = true
				}
			}
			if n > room {
				n--
			}
		} else {
			n--
		}
		if centring {
			if n > width {
				printChar(charNewline)
			}
			setTextPosition((room-n)/2, textY())
		}
		if skipSpaces {
			for ; line[first] == ' '; first++ {
			}
		}
		for ; first <= n; first++ {
			printChar(int(line[first]))
		}
	}
}

// printNumber prints value at least minWidth characters wide, padded on the left with fill.
// C: ULTIMA_1a3e_PrintNumber
func printNumber(value, minWidth int, fill byte) {
	powersOfTen := [...]int{1, 10, 100, 1000, 10000}
	const maxWidth = textColumns - 1
	minWidth = min(minWidth, maxWidth)
	digits := 1
	negative := value < 0
	if negative {
		value = -value
	}
	for _, p := range powersOfTen[1:] {
		if p <= value {
			digits++
		}
	}
	if negative {
		digits++
	}
	var out []byte
	for range max(minWidth-digits, 0) {
		out = append(out, fill)
	}
	if negative {
		out = append(out, '-')
		digits--
	}
	for ; digits != 0; digits-- {
		p := powersOfTen[digits-1]
		digit := value / p
		out = append(out, byte(digit)+'0')
		value -= p * digit
	}
	printString(string(out))
}

// selectWindow makes window id the current one.
// C: ULTIMA_1b94_SelectTextWindow
func selectWindow(id int) {
	if id < windowCount {
		windowIndex = id
		window = &textWindows[id]
		textFg = window.colours & ega.White
		textBg = window.colours >> 4
		underline = window.flags&windowUnderline != 0
		centred = window.flags&windowCentred != 0
		inverse = window.flags&windowInverse != 0
	}
}

// setTextPosition moves the cursor of the current window.
// C: ULTIMA_1bf2_SetTextPosition
func setTextPosition(x, y int) {
	if int(byte(x))+window.left < textColumns && int(byte(y))+window.top < textRows {
		window.x, window.y = x, y
	}
}

// setWindowSize sets the cells of window id.
// C: ULTIMA_1c22_SetTextWindowSize, ULTIMA_1c5b_ClipTextWindow
func setWindowSize(id, x1, y1, x2, y2 int) {
	if id < windowCount {
		x1, y1 = clamp(x1, 0, textColumns-1), clamp(y1, 0, textRows-1)
		x2, y2 = clamp(x2, 0, textColumns-1), clamp(y2, 0, textRows-1)
		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}
		w := &textWindows[id]
		w.left, w.top, w.right, w.bottom = x1, y1, x2, y2
	}
}

// selectFont picks one of the loaded fonts.
// C: ULTIMA_1c9e_SelectCharset
func selectFont(id int) {
	if id < fontCount && fonts[id] != nil {
		font = fonts[id]
		fontIndex = id
	}
}

// setTextColour sets the current window's foreground colour.
// C: ULTIMA_1cca_SetTextForegroundColor
func setTextColour(colour int) {
	textFg = byte(colour) & ega.White
	window.colours = window.colours&^ega.White | textFg
}

// setTextBackground sets the current window's background colour.
// C: ULTIMA_1f26_SetTextBackgroundColor
func setTextBackground(colour int) {
	textBg = byte(colour) & ega.White
	window.colours = window.colours&ega.White | textBg<<4
}

// textX and textY are the cursor of the current window.
// C: ULTIMA_1f12_GetCurrentTextX
func textX() int { return window.x }

// C: ULTIMA_1cee_GetCurrentTextY
func textY() int { return window.y }

// loadFont reads a font file into slot id.
// C: ULTIMA_1d02_LoadCharset
func loadFont(name string, id int) {
	fonts[id] = loadPictures(name)
}

// scrollWindowUp scrolls a pixel rectangle up by amount (at most its height).
// C: ULTIMA_71ca_DRV_27_ScrollTextWindow1
func scrollWindowUp(x1, y1, x2, y2, amount int) {
	x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
	amount = min(amount, y2-y1)
	screen.ScrollWindow(x1, y1, x2, y2, -amount)
}

// scrollWindowDown scrolls a pixel rectangle down.
// C: ULTIMA_7200_DRV_27_ScrollTextWindow2
func scrollWindowDown(x1, y1, x2, y2, amount int) {
	x1, y1, x2, y2 = clipRect(x1, y1, x2, y2)
	amount = min(amount, y2-y1)
	screen.ScrollWindow(x1, y1, x2, y2, amount)
}

// backspace erases count characters before the cursor.
// C: ULTIMA_1fa0_Backspace
func backspace(count int) {
	if (textX() == 0 && textY() == 0) || count == 0 {
		return
	}
	saved := advanceCursor
	w := &textWindows[windowIndex]
	advanceCursor = false
	for range count {
		x := textX()
		printChar(charSpace)
		if x != 0 {
			setTextPosition(x-1, textY())
		} else {
			setTextPosition(w.right-w.left, textY()-1)
		}
	}
	advanceCursor = saved
}

var _ = host.Yield

// Quiet steps (no part of the original): a step's direction ("North", "Advance") is only
// printed when something else follows it before the next prompt, a "Blocked" or what the
// step found; a plain step leaves the prompt line as it was, so the message scroll keeps
// what happened rather than a column of directions.

// EchoSteps prints every step's direction, as the original does.
var EchoSteps = false

// noEcho is no pending direction.
const noEcho = -1

// pendingEcho is the message of the step's direction, not printed yet.
var pendingEcho = noEcho

// echoStep prints a step's direction message, or keeps it for when more text follows.
func echoStep(address int) {
	if EchoSteps {
		printMessage(address)
		return
	}
	pendingEcho = address
}

// flushEcho prints the direction kept back.
func flushEcho() {
	address := pendingEcho
	pendingEcho = noEcho
	printMessage(address)
}

// reusePrompt says whether the last step printed nothing, so the prompt still stands and
// is not printed again.
func reusePrompt() bool {
	if pendingEcho == noEcho {
		return false
	}
	pendingEcho = noEcho
	return true
}

// The message history (F5, no part of the original): every line the message scroll shows,
// as text, kept by the host.

// historyLine is the scroll's line being printed.
var historyLine []byte

// historyPut follows printChar on the message scroll.
func historyPut(c int) {
	switch {
	case c == charNewline:
		host.AddHistory(string(historyLine))
		historyLine = historyLine[:0]
	case c == charReturn:
	case c == glyphOpenCaption:
		historyLine = append(historyLine, '>')
	case c >= ' ' && c <= '~':
		historyLine = append(historyLine, byte(c))
		if columns := int(window.right) - int(window.left) + 1; len(historyLine) >= columns {
			host.AddHistory(string(historyLine)) // the scroll wraps here
			historyLine = historyLine[:0]
		}
	}
}
