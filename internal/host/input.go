package host

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Key is a key press as the BIOS keyboard buffer saw it (int 16h): a scan code and the
// character it typed (0 for keys that type nothing).
type Key struct {
	Scan byte
	Char byte
}

// Scan codes of the keys the game tests.
const (
	ScanEscape    = 1
	ScanBackspace = 14
	ScanTab       = 15
	ScanEnter     = 28
	ScanSpace     = 57
	ScanF1        = 59
	ScanF10       = 68
	ScanF11       = 87
	ScanF12       = 88
	ScanHome      = 71
	ScanUp        = 72
	ScanPageUp    = 73
	ScanLeft      = 75
	ScanRight     = 77
	ScanEnd       = 79
	ScanDown      = 80
	ScanPageDown  = 81
	ScanInsert    = 82
	ScanDelete    = 83
)

const (
	charBackspace  = 8
	charTab        = 9
	charEnter      = 13
	charEscape     = 27
	biosBufferKeys = 16
)

// KeyQueue holds key presses the game has not read yet.
var KeyQueue []Key

// Modifier state.
var ShiftDown, CtrlDown, AltDown bool

var specialKeys = map[ebiten.Key]Key{
	ebiten.KeyEscape: {ScanEscape, charEscape}, ebiten.KeyBackspace: {ScanBackspace, charBackspace},
	ebiten.KeyTab: {ScanTab, charTab}, ebiten.KeyEnter: {ScanEnter, charEnter},
	ebiten.KeyNumpadEnter: {ScanEnter, charEnter},
	ebiten.KeyHome:        {ScanHome, 0}, ebiten.KeyArrowUp: {ScanUp, 0}, ebiten.KeyPageUp: {ScanPageUp, 0},
	ebiten.KeyArrowLeft: {ScanLeft, 0}, ebiten.KeyArrowRight: {ScanRight, 0}, ebiten.KeyEnd: {ScanEnd, 0},
	ebiten.KeyArrowDown: {ScanDown, 0}, ebiten.KeyPageDown: {ScanPageDown, 0},
	ebiten.KeyInsert: {ScanInsert, 0}, ebiten.KeyDelete: {ScanDelete, 0},
	ebiten.KeyF11: {ScanF11, 0}, ebiten.KeyF12: {ScanF12, 0},
}

func init() {
	for i := 0; i < 10; i++ {
		specialKeys[ebiten.KeyF1+ebiten.Key(i)] = Key{ScanF1 + byte(i), 0}
	}
}

// scan codes of the printable keys, by unshifted character
var charScan = map[byte]byte{' ': ScanSpace}

func init() {
	rows := []struct {
		first byte
		keys  string
	}{{2, "1234567890-="}, {16, "qwertyuiop[]"}, {30, "asdfghjkl;'`"}, {43, "\\zxcvbnm,./"}}
	for _, r := range rows {
		for i := 0; i < len(r.keys); i++ {
			charScan[r.keys[i]] = r.first + byte(i)
		}
	}
	shifted := map[byte]byte{'!': '1', '@': '2', '#': '3', '$': '4', '%': '5', '^': '6', '&': '7',
		'*': '8', '(': '9', ')': '0', '_': '-', '+': '=', '{': '[', '}': ']', ':': ';', '"': '\'',
		'~': '`', '|': '\\', '<': ',', '>': '.', '?': '/'}
	for k, v := range shifted {
		charScan[k] = charScan[v]
	}
	for c := byte('A'); c <= 'Z'; c++ {
		charScan[c] = charScan[c-'A'+'a']
	}
}

// PushChar queues the key press that types c.
func PushChar(c byte) { KeyQueue = append(KeyQueue, Key{charScan[c], c}) }

var keysDown = map[ebiten.Key]bool{}

// readModifiers reads Shift, Ctrl and Alt.
func readModifiers() {
	ShiftDown = ebiten.IsKeyPressed(ebiten.KeyShift)
	CtrlDown = ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	AltDown = ebiten.IsKeyPressed(ebiten.KeyAlt)
}

// pollInput queues the keys pressed this tick, as the BIOS would have.
func pollInput() {
	if len(inpututil.AppendJustPressedKeys(nil)) > 0 {
		padLast = false // the keyboard is in use: free text is typed on it
	}
	for k, key := range specialKeys {
		if inpututil.IsKeyJustPressed(k) {
			KeyQueue = append(KeyQueue, key)
		}
	}
	if CtrlDown { // ctrl+letter types the control code
		for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
			if inpututil.IsKeyJustPressed(k) {
				c := byte('a' + k - ebiten.KeyA)
				KeyQueue = append(KeyQueue, Key{charScan[c], c - 'a' + 1})
			}
		}
	} else if AltDown { // Alt shortcuts are the host's
		ebiten.AppendInputChars(nil)
	} else {
		for _, r := range ebiten.AppendInputChars(nil) {
			if r >= ' ' && r <= '~' {
				PushChar(byte(r))
			}
		}
	}
	if len(KeyQueue) > biosBufferKeys {
		KeyQueue = KeyQueue[len(KeyQueue)-biosBufferKeys:]
	}
}
