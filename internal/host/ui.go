package host

// The gamepad and the mouse. Like the menus, they only type the original's key words for
// the player: the d-pad the arrows, A Enter or the prompt's menu, B its "back" key, and free
// text comes from an on-screen keyboard that types it letter by letter.

import (
	"image"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type button int

const (
	bUp button = iota
	bDown
	bLeft
	bRight
	bA
	bB
	bX
	bY
	bStart
	bBack
	bLB
	bRB
	bLS
	nButtons
)

var padButtons = map[ebiten.StandardGamepadButton]button{
	ebiten.StandardGamepadButtonLeftTop: bUp, ebiten.StandardGamepadButtonLeftBottom: bDown,
	ebiten.StandardGamepadButtonLeftLeft: bLeft, ebiten.StandardGamepadButtonLeftRight: bRight,
	ebiten.StandardGamepadButtonRightBottom: bA, ebiten.StandardGamepadButtonRightRight: bB,
	ebiten.StandardGamepadButtonRightLeft: bX, ebiten.StandardGamepadButtonRightTop: bY,
	ebiten.StandardGamepadButtonCenterRight: bStart, ebiten.StandardGamepadButtonCenterLeft: bBack,
	ebiten.StandardGamepadButtonFrontTopLeft: bLB, ebiten.StandardGamepadButtonFrontTopRight: bRB,
	ebiten.StandardGamepadButtonLeftStick: bLS,
}

var (
	padHeld [nButtons]int // ticks each button has been held, any pad

	// the pointer this tick, in window pixels
	ptrX, ptrY        int
	ptrLeft, ptrRight bool
	ptrWheel          float64
)

// readPad reads every connected pad (with the standard layout); the left stick is the d-pad.
func readPad() {
	var down [nButtons]bool
	for _, id := range ebiten.AppendGamepadIDs(nil) {
		if !ebiten.IsStandardGamepadLayoutAvailable(id) {
			continue
		}
		for sb, b := range padButtons {
			down[b] = down[b] || ebiten.IsStandardGamepadButtonPressed(id, sb)
		}
		x := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
		y := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
		down[bLeft] = down[bLeft] || x < -0.5
		down[bRight] = down[bRight] || x > 0.5
		down[bUp] = down[bUp] || y < -0.5
		down[bDown] = down[bDown] || y > 0.5
	}
	for b, d := range down {
		if d {
			padLast = padLast || padHeld[b] == 0
			padHeld[b]++
		} else {
			padHeld[b] = 0
		}
	}
}

func readPointer() {
	ptrX, ptrY = ebiten.CursorPosition()
	ptrLeft = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	ptrRight = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	_, ptrWheel = ebiten.Wheel()
}

// padLast is set when the latest input came from a pad (the keyboard clears it): only
// then does free text come from the on-screen keyboard.
var padLast bool

// padMode says whether the on-screen keyboard is offered.
func padMode() bool {
	return padLast && len(ebiten.AppendGamepadIDs(nil)) > 0
}

func padPressed(b button) bool { return padHeld[b] == 1 }

// padRepeat is padPressed with the keyboard's typematic repeat while b is held.
func padRepeat(b button) bool {
	d := padHeld[b]
	return d == 1 || d >= TicksPerSecond/4 && (d-TicksPerSecond/4)%(TicksPerSecond/15) == 0
}

func push(keys ...Key) { KeyQueue = append(KeyQueue, keys...) }

var arrowKeys = map[button]Key{bUp: KeyUp, bDown: KeyDown, bLeft: KeyLeft, bRight: KeyRight}

func promptMenu() *Menu {
	if PromptMenu == nil {
		return nil
	}
	return PromptMenu(true)
}

// confirm is A (or a click that is not the game's): the prompt's menu, else Enter for a
// "press a key" wait (only while the game reads keys: one pressed during a cutscene would
// reach the next prompt).
func confirm() {
	if m := promptMenu(); m != nil {
		openMenu(m)
	} else if KeyPolled {
		push(KeyEnter)
	}
}

// uiInput handles the pad and the mouse, in every layout.
func uiInput(g *game) {
	moved := ptrX != g.lastMouseX || ptrY != g.lastMouseY
	g.lastMouseX, g.lastMouseY = ptrX, ptrY
	if padPressed(bBack) {
		Modern = !Modern
		menu, modal = nil, 0
		return
	}
	if padPressed(bLS) {
		modal = modalHelp
		return
	}
	switch {
	case osk.open:
		oskInput(g)
	case modal == modalSettings:
		i, on := g.menuRow(ptrX, ptrY)
		if on && moved {
			settingSel = i
		}
		switch {
		case ptrLeft && on:
			settingSel = i
			cycleSetting(i, 1)
		case ptrRight && on:
			cycleSetting(i, -1)
		case ptrLeft, ptrRight, padPressed(bB), padPressed(bStart):
			modal = 0
		case padRepeat(bUp):
			settingSel = (settingSel + len(Settings) - 1) % len(Settings)
		case padRepeat(bDown):
			settingSel = (settingSel + 1) % len(Settings)
		case padRepeat(bLeft):
			cycleSetting(settingSel, -1)
		case padRepeat(bRight), padPressed(bA):
			cycleSetting(settingSel, 1)
		}
	case modal != 0:
		switch {
		case padRepeat(bUp):
			histBack++
		case padRepeat(bDown):
			histBack--
		case padRepeat(bLB):
			histBack += histRows
		case padRepeat(bRB):
			histBack -= histRows
		case padPressed(bA), padPressed(bB), padPressed(bY), ptrLeft, ptrRight:
			modal = 0
		}
		histBack += int(ptrWheel * 3)
	case menu != nil:
		n := len(menu.Items)
		i, on := g.menuRow(ptrX, ptrY)
		if on && moved {
			menu.Sel = i
		}
		switch {
		case ptrLeft && on:
			choose(i)
		case ptrLeft, ptrRight, padPressed(bB), padPressed(bStart):
			menu = nil
		case padRepeat(bUp):
			menu.Sel = (menu.Sel + n - 1) % n
		case padRepeat(bDown):
			menu.Sel = (menu.Sel + 1) % n
		case padPressed(bLB):
			menu.Sel = 0
		case padPressed(bRB):
			menu.Sel = n - 1
		case padPressed(bA):
			choose(menu.Sel)
		}
	default:
		for b, k := range arrowKeys {
			if padRepeat(b) {
				push(k)
			}
		}
		switch {
		case padPressed(bA):
			confirm()
		case padPressed(bStart), ptrRight:
			openMenu(promptMenu())
		case padPressed(bB):
			if BackKey == nil {
				push(KeyEsc)
			} else if k, ok := BackKey(); ok {
				push(k)
			}
		case padPressed(bX):
			push(KeySpace)
		case padPressed(bY):
			modal, histBack = modalHistory, 0
		case padPressed(bLB):
			ShowStats = !ShowStats
		case padPressed(bRB):
			ShowText = !ShowText
		case ptrLeft:
			g.click(ptrX, ptrY)
		}
	}
}

// click is a left click with no menu or window open.
func (g *game) click(x, y int) {
	if modern() {
		if image.Pt(x, y).In(g.world) && MapClick != nil {
			t := float64(g.world.Dx()) / mapTiles
			dx := int(math.Floor(float64(x-g.world.Min.X)/t)) - mapTiles/2
			dy := int(math.Floor(float64(y-g.world.Min.Y)/t)) - mapTiles/2
			if keys, ok := MapClick(dx, dy); ok {
				push(keys...)
			} else if dx == 0 && dy == 0 {
				openMenu(promptMenu())
			}
			return
		}
	} else if sx, sy, ok := g.toScreen(x, y); ok && ScreenClick != nil {
		if keys, ok := ScreenClick(sx, sy); ok {
			push(keys...)
			return
		}
	}
	confirm()
}

// ---- the on-screen keyboard ----

var osk struct {
	open  bool
	row   int
	col   int
	lower bool
	typed string
}

var oskRows = []string{"ABCDEFGHIJ", "KLMNOPQRST", "UVWXYZ-.'?", "1234567890"}

// the last row's cells
var oskSpecial = []string{"Aa", "Spc", "Del", "Done"}

const oskSpecialW = 5 // its cells' width in characters (the others are 2)

func openKeyboard() {
	menu = nil
	if !padMode() {
		return
	}
	osk.open, osk.row, osk.col, osk.typed = true, 0, 0, ""
}

// oskLines are the keyboard's text lines: what was typed, a blank, then the keys.
func oskLines() []string {
	l := []string{"> " + osk.typed + "_", ""}
	for _, r := range oskRows {
		if osk.lower {
			r = strings.ToLower(r)
		}
		l = append(l, strings.Join(strings.Split(r, ""), " "), "")
	}
	var b strings.Builder
	for _, s := range oskSpecial {
		b.WriteString(s + strings.Repeat(" ", oskSpecialW-len(s)))
	}
	return append(l, strings.TrimRight(b.String(), " "))
}

// the key rows' lines in oskLines: what was typed and a blank line first, a blank line
// between rows
const oskTop = 2

func oskLine(row int) int { return oskTop + 2*row }

func cellW(row int) int {
	if row == len(oskRows) {
		return oskSpecialW
	}
	return 2
}

func rowLen(row int) int {
	if row == len(oskRows) {
		return len(oskSpecial)
	}
	return len(oskRows[row])
}

// oskInv says which characters of oskLines show the selected key.
func oskInv(x, y int) bool {
	if y != oskLine(osk.row) {
		return false
	}
	w := cellW(osk.row)
	return x >= osk.col*w && x < osk.col*w+w-1
}

// oskPress types key (row, col).
func oskPress(row, col int) {
	if row < len(oskRows) {
		c := oskRows[row][col]
		if osk.lower && c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		push(CharKey(c))
		osk.typed += string(c)
		return
	}
	switch oskSpecial[col] {
	case "Aa":
		osk.lower = !osk.lower
	case "Spc":
		push(KeySpace)
		osk.typed += " "
	case "Del":
		push(KeyBack)
		if osk.typed != "" {
			osk.typed = osk.typed[:len(osk.typed)-1]
		}
	case "Done":
		push(KeyEnter)
		osk.open = false
	}
}

func oskMove(drow, dcol int) {
	n := len(oskRows) + 1
	if drow != 0 {
		x := osk.col*cellW(osk.row) + cellW(osk.row)/2 // keep the column under the same place
		osk.row = (osk.row + drow + n) % n
		osk.col = min(x/cellW(osk.row), rowLen(osk.row)-1)
	}
	osk.col = (osk.col + dcol + rowLen(osk.row)) % rowLen(osk.row)
}

// oskCell is the key under window pixel (x, y).
func (g *game) oskCell(x, y int) (row, col int, ok bool) {
	if !image.Pt(x, y).In(g.oskRect) {
		return 0, 0, false
	}
	cx := int(float64(x-g.oskRect.Min.X)/(cell*g.oskK)) - 1
	ly := int(float64(y-g.oskRect.Min.Y)/(cell*g.oskK)) - 2 - oskTop // 2: the title rows
	cy := (ly + 1) / 2                                            // a blank line counts as the row below it
	if ly < -1 || cy > len(oskRows) || cx < 0 {
		return 0, 0, false
	}
	col = cx / cellW(cy)
	return cy, col, col < rowLen(cy)
}

func oskInput(g *game) {
	if r, c, ok := g.oskCell(ptrX, ptrY); ok && ptrLeft {
		osk.row, osk.col = r, c
		oskPress(r, c)
		return
	} else if ptrLeft || ptrRight || padPressed(bY) {
		osk.open = false
		return
	}
	switch {
	case padRepeat(bUp):
		oskMove(-1, 0)
	case padRepeat(bDown):
		oskMove(1, 0)
	case padRepeat(bLeft):
		oskMove(0, -1)
	case padRepeat(bRight):
		oskMove(0, 1)
	case padRepeat(bA):
		oskPress(osk.row, osk.col)
	case padRepeat(bB):
		oskPress(len(oskRows), 2)
	case padPressed(bX):
		oskPress(len(oskRows), 1)
	case padPressed(bStart):
		oskPress(len(oskRows), 3)
	}
}

// toScreen maps window pixel (x, y) to the 320x200 screen in the original layout.
func (g *game) toScreen(x, y int) (sx, sy int, ok bool) {
	if modern() || g.scaleX == 0 {
		return 0, 0, false
	}
	sx = int(math.Floor((float64(x) - g.offsetX) / g.scaleX))
	sy = int(math.Floor((float64(y) - g.offsetY) / g.scaleY))
	return sx, sy, true
}
