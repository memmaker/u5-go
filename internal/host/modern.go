package host

// The modern layout (no part of the original): on the play screen the map window fills the
// window, the party box and the message scroll float over it as overlays that can be
// toggled, and menus offer the commands and answers the original wanted typed. A menu only
// types the original's keys for the player, so the game itself runs exactly as before.
// The windows drawn over the game (menus, help, history, settings) are drawn the same way in
// both layouts: the game's own font in a frame of the play screen's border pieces.

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MenuItem is one choice: the keys it types for the player.
type MenuItem struct {
	Label string
	Keys  []Key
	// Keyboard opens the on-screen keyboard instead (free text).
	Keyboard bool
	// Sub opens another menu instead.
	Sub *Menu
}

// Menu is a list of choices shown over the game.
type Menu struct {
	Title string
	Items []MenuItem
	Sel   int
	// Pause holds the game while the menu is open (a command prompt passes the turn when
	// left waiting).
	Pause bool
	id    int
}

var (
	// Modern selects the modern layout (F6 switches to the original one).
	Modern = true
	// InWorld is set by the game while the frame shows the play screen.
	InWorld bool
	// ShowStats and ShowText toggle the party box (F3) and message scroll (F4) overlays.
	ShowStats = true
	ShowText  = true

	// PromptMenu returns the menu for what the game waits for: asked is set when the
	// player asked for it (Enter, A or the right button); otherwise it is about to open by
	// itself. nil: none.
	PromptMenu func(asked bool) *Menu
	// MapClick returns the keys a click on the map types, the tile given relative to the
	// middle of the view; nil keys with handled false open the menu instead.
	MapClick func(dx, dy int) (keys []Key, handled bool)
	// WideView returns cols x rows tiles of map around the party (16x16 pixels each,
	// palette indexes) for the wider view, or nil when the map window shows something else.
	WideView func(cols, rows int) []byte
	// WideRows is the number of map rows the wider view shows (the original shows 11).
	WideRows = 15
	// Glyphs returns the game's font, 8x8 palette indexes per character.
	Glyphs func() [][64]byte
	// BackKey is the key the gamepad's B button (back) types for the current prompt; ok
	// false: none.
	BackKey func() (Key, bool)
	// ScreenClick returns the keys a left click at (x, y) of the 320x200 screen types in the
	// original layout; ok false: the click is not the game's.
	ScreenClick func(x, y int) (keys []Key, ok bool)
	// KeyPolled is set by the game whenever it looks at the keyboard; the host clears it
	// each tick.
	KeyPolled bool

	promptID, autoID int
	menu             *Menu
	modal            int // 0 none, or one of the modals
	histBack         int // lines scrolled back in the history
)

const (
	modalHistory = 1 + iota
	modalHelp
	modalSettings
)

// The keys menus type.
var (
	KeyEnter = Key{ScanEnter, charEnter}
	KeyEsc   = Key{ScanEscape, charEscape}
	KeySpace = Key{ScanSpace, ' '}
	KeyBack  = Key{ScanBackspace, charBackspace}
	KeyUp    = Key{Scan: ScanUp}
	KeyDown  = Key{Scan: ScanDown}
	KeyLeft  = Key{Scan: ScanLeft}
	KeyRight = Key{Scan: ScanRight}
)

// CharKey is the key press that types c.
func CharKey(c byte) Key { return Key{charScan[c], c} }

// NewPrompt tells the host the game now waits for something else: an open menu for the
// last prompt closes.
func NewPrompt() { promptID++ }

var helpText = []string{
	"Enter / right click   commands & choices",
	"Click the map         walk / pick a tile",
	"Click elsewhere       continue / choices",
	"Arrows, letters       as in the original",
	"",
	"Gamepad",
	"D-pad / left stick    move, pick",
	"A       confirm, commands & choices",
	"B       back / cancel",
	"X       pass (space)",
	"Y       message history",
	"Start   command menu",
	"Back    original layout (F6)",
	"LB, RB  party, message overlays;",
	"        page in the history",
	"L-stick click  this help",
	"Text (gamepad): A types, B deletes,",
	"        X space, Start done",
	"",
	"F1   this help",
	"F2   settings (music, sound, ...)",
	"F3   party overlay",
	"F4   message overlay",
	"F5   message history",
	"F6   original layout",
	"Alt-V  music on/off, Alt-+/- volume",
}

func modern() bool { return Modern && InWorld }

func openMenu(m *Menu) {
	if m == nil || len(m.Items) == 0 {
		return
	}
	if !padMode() { // free text is typed on the keyboard
		var items []MenuItem
		for _, it := range m.Items {
			if !it.Keyboard {
				items = append(items, it)
			}
		}
		if len(items) == 0 {
			return
		}
		m.Items = items
	}
	if len(m.Items) == 1 && m.Items[0].Keyboard {
		openKeyboard()
		return
	}
	m.id = promptID
	menu = m
}

func choose(i int) {
	m := menu
	menu = nil
	switch {
	case i < 0 || i >= len(m.Items):
	case m.Items[i].Keyboard:
		openKeyboard()
	case m.Items[i].Sub != nil:
		openMenu(m.Items[i].Sub)
	default:
		KeyQueue = append(KeyQueue, m.Items[i].Keys...)
	}
}

// updateModern runs before the game's tick: it opens and closes menus. It reports whether
// the game is held this tick.
func updateModern() bool {
	if menu != nil && menu.id != promptID {
		menu = nil
	}
	if menu == nil && modern() && len(KeyQueue) == 0 && autoID != promptID && PromptMenu != nil {
		autoID = promptID
		openMenu(PromptMenu(false))
	}
	return (menu != nil && menu.Pause) || modal != 0
}

// pressed reports a key that went down this tick.
func pressed(k ebiten.Key) bool { return inpututil.IsKeyJustPressed(k) }

// repeating is pressed with typematic repeat while k is held (after a quarter second, 15 a
// second).
func repeating(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || d >= TicksPerSecond/4 && (d-TicksPerSecond/4)%(TicksPerSecond/15) == 0
}

// modernInput handles the layout's keys and the keys of an open menu or window, in any
// layout. It reports whether it used the tick's keyboard input, which then does not reach
// the game.
func modernInput() bool {
	switch {
	case pressed(ebiten.KeyF6):
		Modern = !Modern
		menu, modal = nil, 0
		return true
	case pressed(ebiten.KeyF1):
		modal = toggle(modal, modalHelp)
		return true
	case pressed(ebiten.KeyF2) && len(Settings) > 0:
		modal = toggle(modal, modalSettings)
		return true
	case pressed(ebiten.KeyF5):
		modal, histBack = toggle(modal, modalHistory), 0
		return true
	}
	if osk.open {
		if pressed(ebiten.KeyEscape) {
			osk.open = false
			return true
		}
		return false // typing goes to the game as ever
	}
	if modal != 0 {
		modalKeys()
		ebiten.AppendInputChars(nil)
		return true
	}
	if menu != nil {
		n := len(menu.Items)
		switch {
		case repeating(ebiten.KeyArrowUp):
			menu.Sel = (menu.Sel + n - 1) % n
		case repeating(ebiten.KeyArrowDown):
			menu.Sel = (menu.Sel + 1) % n
		case pressed(ebiten.KeyPageUp), pressed(ebiten.KeyHome):
			menu.Sel = 0
		case pressed(ebiten.KeyPageDown), pressed(ebiten.KeyEnd):
			menu.Sel = n - 1
		case pressed(ebiten.KeyEnter), pressed(ebiten.KeyNumpadEnter):
			choose(menu.Sel)
		case pressed(ebiten.KeyEscape):
			menu = nil
		default:
			// typing answers the prompt as in the original: the menu steps aside
			if len(ebiten.AppendInputChars(nil)) > 0 {
				menu = nil
				return false
			}
		}
		return true
	}
	if !modern() {
		return false
	}
	switch {
	case pressed(ebiten.KeyF3):
		ShowStats = !ShowStats
		return true
	case pressed(ebiten.KeyF4):
		ShowText = !ShowText
		return true
	case pressed(ebiten.KeyEnter) || pressed(ebiten.KeyNumpadEnter):
		if PromptMenu != nil {
			if m := PromptMenu(true); m != nil {
				openMenu(m)
				return true
			}
		}
	}
	return false
}

func toggle(open, m int) int {
	if open == m {
		return 0
	}
	return m
}

// modalKeys handles the keys of the open help, history or settings.
func modalKeys() {
	switch {
	case pressed(ebiten.KeyEscape), modal != modalSettings && pressed(ebiten.KeyEnter):
		modal = 0
	case modal == modalSettings:
		settingsKeys()
	case repeating(ebiten.KeyArrowUp):
		histBack++
	case repeating(ebiten.KeyArrowDown):
		histBack--
	case repeating(ebiten.KeyPageUp):
		histBack += histRows - 1
	case repeating(ebiten.KeyPageDown):
		histBack -= histRows - 1
	case pressed(ebiten.KeyHome):
		histBack = len(History)
	case pressed(ebiten.KeyEnd):
		histBack = 0
	}
}

// The text drawn over the game: the font's cells.
const (
	cell     = 8
	tileSize = 16
	// mapCells is the original map window's size in cells (11 tiles).
	mapPixels = 176
	mapLeft   = 8
	mapTop    = 8
	mapTiles  = 11
)

// The play screen's parts the modern layout shows as overlays, with their own borders.
var (
	partStats  = image.Rect(184, 0, 320, 92)
	partScroll = image.Rect(184, 92, 320, 200)
	partMap    = image.Rect(mapLeft, mapTop, mapLeft+mapPixels, mapTop+mapPixels)
	partTop    = image.Rect(0, 0, 184, 8)
	partBottom = image.Rect(0, 184, 184, 200)
)

// The pieces of the play screen's border (8x8 cells) that frame every overlay, menu and
// window: its rounded corners and the pipes, their white edge inside.
var (
	cellCornerTL = image.Rect(0, 0, 8, 8)
	cellCornerTR = image.Rect(312, 0, 320, 8)
	cellCornerBL = image.Rect(0, 192, 8, 200)
	cellTop      = image.Rect(8, 0, 16, 8)
	cellLeft     = image.Rect(0, 40, 8, 48)
	cellRight    = image.Rect(312, 40, 320, 48)
)

// text paints lines in the game's font into an image, row sel highlighted.
func (g *game) text(lines []string, sel int, title string) *ebiten.Image {
	return g.textInv(lines, func(_, y int) bool { return y == sel }, title)
}

// The colours of the windows' text: white on black, the selection black on white.
var (
	textFg = [4]byte{255, 255, 255, 255}
	textBg = [4]byte{0, 0, 0, 255}
)

// textInv is text with the cells (x, y) for which inv holds highlighted.
func (g *game) textInv(lines []string, inv func(x, y int) bool, title string) *ebiten.Image {
	glyphs := Glyphs()
	w := len(title)
	for _, l := range lines {
		w = max(w, len(l))
	}
	w += 2
	top := 1
	if title != "" {
		top = 2
	}
	h := len(lines) + top + 1
	pix := make([]byte, w*cell*h*cell*4)
	put := func(cx, cy int, c byte, inv bool) {
		var gl [64]byte
		if int(c) < len(glyphs) {
			gl = glyphs[c]
		}
		for r := range cell {
			for k := range cell {
				on := gl[r*cell+k] != 0
				rgb := textBg
				if on != inv {
					rgb = textFg
				}
				o := ((cy*cell+r)*w*cell + cx*cell + k) * 4
				copy(pix[o:o+4], rgb[:])
			}
		}
	}
	for y := range h {
		for x := range w {
			put(x, y, ' ', false)
		}
	}
	for i := range len(title) {
		put(1+i, 0, title[i], false)
	}
	for j, l := range lines {
		for x := range w - 2 {
			c := byte(' ')
			if x < len(l) {
				c = l[x]
			}
			put(1+x, top+j, c, inv(x, j))
		}
	}
	img := ebiten.NewImage(w*cell, h*cell)
	img.WritePixels(pix)
	return img
}

// piece draws the border cell c stretched over w x h window pixels at (x, y), flipped
// upside down with flip.
func (g *game) piece(screen *ebiten.Image, c image.Rectangle, x, y, w, h float64, flip bool) {
	op := &ebiten.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, cell)
	}
	op.GeoM.Scale(w/cell, h/cell)
	op.GeoM.Translate(x, y)
	screen.DrawImage(g.image.SubImage(c).(*ebiten.Image), op)
}

// drawBox draws img scaled by k with its top left at (x, y), in the play screen's border
// (one cell, 8k, around it). It returns the content's place.
func (g *game) drawBox(screen, img *ebiten.Image, x, y, k float64) image.Rectangle {
	b := img.Bounds()
	w, h, e := float64(b.Dx())*k, float64(b.Dy())*k, cell*k
	r := image.Rect(int(x), int(y), int(x+w), int(y+h))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(x, y)
	if !InWorld || g.image == nil || g.width != ScreenWidth { // no play screen border: a plain frame
		screen.SubImage(r.Inset(-int(e))).(*ebiten.Image).Fill(color.RGBA{0, 0, 0xaa, 0xff})
		screen.SubImage(r.Inset(-int(e / 2))).(*ebiten.Image).Fill(color.White)
		screen.SubImage(r.Inset(-int(e/2) + 1)).(*ebiten.Image).Fill(color.Black)
		screen.DrawImage(img, op)
		return r
	}
	screen.SubImage(r).(*ebiten.Image).Fill(color.Black)
	screen.DrawImage(img, op)
	g.piece(screen, cellTop, x, y-e, w, e, false)
	g.piece(screen, cellTop, x, y+h, w, e, true)
	g.piece(screen, cellLeft, x-e, y, e, h, false)
	g.piece(screen, cellRight, x+w, y, e, h, false)
	g.piece(screen, cellCornerTL, x-e, y-e, e, e, false)
	g.piece(screen, cellCornerTR, x+w, y-e, e, e, false)
	g.piece(screen, cellCornerBL, x-e, y+h, e, e, false)
	g.piece(screen, cellCornerTR, x+w, y+h, e, e, true)
	return r
}

// drawPart draws a part of the play screen as it is.
func (g *game) drawPart(screen *ebiten.Image, part image.Rectangle, x, y, k float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(k, k)
	op.GeoM.Translate(x, y)
	screen.DrawImage(g.image.SubImage(part).(*ebiten.Image), op)
}

// drawModern draws the play screen in the modern layout.
func (g *game) drawModern(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	k := float64(min(sw, sh)) / mapPixels
	k = float64(int(k*tileSize)) / tileSize
	if WideView != nil {
		rows := WideRows
		wk := float64(sh) / float64(rows*tileSize)
		cols := int(float64(sw)/(tileSize*wk)) | 1
		cols += 2 // a partly shown column on each side
		if wide := WideView(cols, rows); wide != nil {
			k = wk
			img := ebiten.NewImage(cols*tileSize, rows*tileSize)
			pix := make([]byte, len(wide)*4)
			for i, c := range wide {
				copy(pix[i*4:i*4+4], g.lut[c][:])
			}
			img.WritePixels(pix)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(-float64(cols*tileSize)/2, -float64(rows*tileSize)/2)
			op.GeoM.Scale(k, k)
			op.GeoM.Translate(float64(sw)/2, float64(sh)/2)
			screen.DrawImage(img, op)
			img.Deallocate()
		}
	}
	size := int(mapPixels * k)
	wx, wy := (sw-size)/2, (sh-size)/2
	g.world = image.Rect(wx, wy, wx+size, wy+size)
	if d := dungeon(); d != nil {
		g.drawDungeon(screen, d)
	} else {
		g.drawPart(screen, partMap, float64(wx), float64(wy), k)
	}

	// overlays: about a third of the map's scale, at least 2; the parts of the play screen
	// that have their own border are shown with it
	ok := max(2, float64(int(k/1.6)))
	fw, fh := float64(sw), float64(sh)
	if ShowStats {
		g.drawPart(screen, partStats, fw-float64(partStats.Dx())*ok, 0, ok)
		g.drawPart(screen, partTop, 0, 0, ok)
		g.drawPart(screen, partBottom, 0, fh-float64(partBottom.Dy())*ok, ok)
	}
	if ShowText {
		scroll := partScroll
		scroll.Min.X += cell // inside the border pieces drawBox adds
		g.drawBox(screen, g.image.SubImage(scroll).(*ebiten.Image),
			fw-float64(scroll.Dx()+cell)*ok, fh-float64(scroll.Dy()+cell)*ok, ok)
	}
	g.drawUI(screen, ok, sw, sh)
}

// histRows is the history's height in lines.
const histRows = 20

// drawUI draws the open window, menu or on-screen keyboard (any layout), ok the overlays'
// scale.
func (g *game) drawUI(screen *ebiten.Image, ok float64, sw, sh int) {
	g.menuRect, g.oskRect = image.Rectangle{}, image.Rectangle{}
	if Glyphs == nil {
		return
	}
	fit := func(img *ebiten.Image) float64 {
		b := img.Bounds()
		return min(ok, float64(sh)*0.92/float64(b.Dy()+2*cell), float64(sw)*0.92/float64(b.Dx()+2*cell))
	}
	switch {
	case modal == modalHelp:
		img := g.text(helpText, -1, "Keys")
		g.centred(screen, img, fit(img), sw, sh)
	case modal == modalHistory:
		histBack = max(0, min(histBack, len(History)-histRows))
		end := len(History) - histBack
		lines := History[max(0, end-histRows):max(0, end)]
		img := g.text(lines, -1, "Messages (arrows, wheel; Esc)")
		g.centred(screen, img, fit(img), sw, sh)
	case modal == modalSettings:
		lines, sel := settingsLines()
		img := g.text(lines, sel, "Settings (arrows change; Esc)")
		g.menuK = fit(img)
		g.menuTop = 2
		g.menuRect = g.centred(screen, img, g.menuK, sw, sh)
	case osk.open:
		img := g.textInv(oskLines(), oskInv, "Type (Start: done)")
		g.oskK = fit(img)
		g.oskRect = g.centred(screen, img, g.oskK, sw, sh)
	case menu != nil:
		img := g.text(labels(menu), menu.Sel, menu.Title)
		g.menuK = fit(img)
		g.menuTop = 2
		if menu.Title == "" {
			g.menuTop = 1
		}
		g.menuRect = g.centred(screen, img, g.menuK, sw, sh)
	}
}

func labels(m *Menu) []string {
	s := make([]string, len(m.Items))
	for i, it := range m.Items {
		s[i] = it.Label
	}
	return s
}

func (g *game) centred(screen, img *ebiten.Image, k float64, sw, sh int) image.Rectangle {
	b := img.Bounds()
	x := (float64(sw) - float64(b.Dx())*k) / 2
	y := (float64(sh) - float64(b.Dy())*k) / 2
	return g.drawBox(screen, img, x, y, k)
}

// menuRow is the menu (or settings) row under window pixel (x, y).
func (g *game) menuRow(x, y int) (int, bool) {
	n := 0
	switch {
	case menu != nil:
		n = len(menu.Items)
	case modal == modalSettings:
		n = len(Settings)
	}
	if n == 0 || !image.Pt(x, y).In(g.menuRect) {
		return 0, false
	}
	i := int(float64(y-g.menuRect.Min.Y)/(cell*g.menuK)) - g.menuTop
	return i, i >= 0 && i < n
}

func dungeon() *Dungeon {
	if DungeonView == nil {
		return nil
	}
	return DungeonView()
}
