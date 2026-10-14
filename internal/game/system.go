package game

import (
	"fmt"
	"time"

	"github.com/memmaker/u5-go/internal/host"
)

// The program's system layer (segments 1000 and 2000): the keyboard, the clock, the random
// numbers and the speaker.

// Key codes the game reads from pollKey besides plain characters.
const (
	keyLeft      = 1
	keyRight     = 2
	keyUp        = 3
	keyDown      = 4
	keyCtrlE     = 5
	keyBackspace = 8
	keyCtrlK     = 11
	keyEnter     = 13
	keyCtrlS     = 19
	keyCtrlV     = 22
	keyCtrlB     = 252
	keyEscape    = 27
	keySpace     = ' '
	keyHome      = 211
	keyEnd       = 212
	keyPageUp    = 213
	keyPageDown  = 214
	// keyF1 is F1's code; F2..F10 follow (the BIOS scan code plus functionKeyOffset).
	functionKeyOffset = 142
)

// navigationScans are the BIOS scan codes of the arrows and the four keys above them, and
// navigationKeys the codes the game sees for them (D_540e, D_5416).
var (
	navigationScans = [...]byte{host.ScanLeft, host.ScanRight, host.ScanUp, host.ScanDown, host.ScanHome, host.ScanEnd, host.ScanPageUp, host.ScanPageDown}
	navigationKeys  = [...]int{keyLeft, keyRight, keyUp, keyDown, keyHome, keyEnd, keyPageUp, keyPageDown}
)

// keyBuffering keeps the keys typed ahead; without it each key read drops the rest
// (D_538c, which holds the opposite: set when the buffer is off).
var keyBuffering bool // C: D_538c

// navigationKey is set when the last key read was one of the navigation keys (D_538a).
var navigationKey bool

// pollKey returns the next key press, 0 when there is none. It lets a frame pass, so loops
// polling the keyboard do not spin.
// C: ULTIMA_1d5e_PollKey
func pollKey() int {
	host.KeyPolled = true
	navigationKey = false
	host.Yield()
	if len(host.KeyQueue) == 0 {
		return 0
	}
	k := host.KeyQueue[0]
	host.KeyQueue = host.KeyQueue[1:]
	if !keyBuffering {
		clearKeyboard()
	}
	if k.Char != 0 {
		return int(k.Char)
	}
	if k.Scan >= host.ScanF1 && k.Scan <= host.ScanF10 {
		return int(k.Scan) + functionKeyOffset
	}
	for i, scan := range navigationScans {
		if k.Scan == scan {
			navigationKey = true
			return navigationKeys[i]
		}
	}
	return 0
}

// keyHit says whether a key is waiting, and takes it: the effects stop at one.
func keyHit() bool { return pollKey() != 0 }

// clearKeyboard drops the keys typed ahead.
// C: ULTIMA_1b16_ClearKbdBuffer
func clearKeyboard() { host.KeyQueue = host.KeyQueue[:0] }

// pollKeyWithCursor shows the cursor's next frame, polls the keyboard and waits a tick if
// no key came.
// C: ULTIMA_1b38_PollKeyWithCursor
func pollKeyWithCursor() int {
	saved := advanceCursor
	advanceCursor = false
	printChar(cursorGlyph + cursorFrame)
	cursorFrame++
	key := pollKey()
	if cursorFrame >= cursorGlyphCount {
		cursorFrame = 0
	}
	if key != 0 {
		printChar(charSpace)
	} else {
		waitTicks(1)
	}
	advanceCursor = saved
	return key
}

// waitForKey waits for a key with the cursor blinking; with limit > 0 it gives up after
// limit polls and returns 0.
// C: ULTIMA_1dda_WaitForKeystroke
func waitForKey(limit int) int {
	polls := -1
	if limit != 0 {
		polls++
	}
	for {
		key := pollKeyWithCursor()
		if limit != 0 {
			polls++
		}
		if polls >= limit || key != 0 {
			return key
		}
	}
}

// readLine reads up to maxLen printable characters until Enter, echoing them.
// C: ULTIMA_1e38_IntroGetString
func readLine(maxLen int) string {
	defer asking(promptText)()
	var typed []byte
	for {
		key := waitForKey(0)
		switch {
		case (key == keyBackspace || key == keyLeft) && len(typed) != 0:
			backspace(1)
			typed = typed[:len(typed)-1]
		case key == keyEscape && len(typed) != 0:
			backspace(len(typed))
			typed = typed[:0]
		case key >= ' ' && key <= lastASCII && len(typed) < maxLen:
			typed = append(typed, byte(key))
			printChar(key)
		}
		if key == keyEnter {
			return string(typed)
		}
	}
}

// toUpper upper-cases a letter.
// C: ULTIMA_2032_ToUpper
func toUpper(c int) int {
	if c >= 'a' && c <= 'z' {
		return c - ('a' - 'A')
	}
	return c
}

// randomSeed is the random generator's state (D_5420).
var randomSeed uint16

// seedRandom seeds the random generator.
// C: ULTIMA_207e_srand
func seedRandom(seed uint16) { randomSeed = seed }

// clockSeed is the seed the game takes from the time of day.
// C: ULTIMA_2056_GetTime
func clockSeed() uint16 { return uint16(time.Now().Unix()) }

// The random generator's constants.
const (
	randomAdd    = 37448
	randomRotate = 3
	randomXor    = 37448
	randomOffset = 17
	randomMask   = 32767
	wordBits     = 16
)

// randomRange returns a random number from low to high (inclusive).
// C: ULTIMA_2092_RandomRange
func randomRange(low, high int) int {
	v := randomSeed + randomAdd
	v = v>>randomRotate | v<<(wordBits-randomRotate)
	v = (v ^ randomXor) + randomOffset
	randomSeed = v
	return int(v&randomMask)%(high-low+1) + low
}

// tick is the PC timer's tick (18.2 a second).
const tick = 55 * time.Millisecond

// waitTicks waits n timer ticks.
// C: ULTIMA_20fa_WaitTicks
func waitTicks(n int) { host.Wait(time.Duration(n) * tick) }

// delayUnit is a step of the calibrated busy wait (its 20 steps make a 60 Hz frame).
const delayUnit = 800 * time.Microsecond

// busyWait waits steps calibrated delay steps.
// C: ULTIMA_20c8_SomeDelay
func busyWait(steps int) { host.Wait(time.Duration(steps) * delayUnit) }

// diskError reports a game file that could not be read: the original asked for the disk.
func diskError(err error) {
	panic(fmt.Errorf("game file: %w", err))
}

// tandyVideo is the video driver selection of the Tandy driver the port draws with.
const tandyVideo = 2

// getChar waits for a key, upper-cased, updating the frame while none comes on the maps
// that animate; the number keys of the keypad become directions.
// C: ULTIMA_266c_GetChar
func getChar() int {
	if videoDriver == tandyVideo {
		animateFrame = 255
	}
	var key int
	for {
		key = toUpper(pollKeyWithCursor())
		if key != 0 {
			break
		}
		if mapID < firstDungeonMap || mapID >= firstMapCombat {
			updateFrame()
		}
	}
	if navigationKey {
		if mapped, ok := keypadKeys[key]; ok {
			key = mapped
		}
	} else if key < keyCtrlE {
		key += functionKeyLowOffset
	}
	return key
}

// functionKeyLowOffset moves the codes below 5 out of the way when they came from a key
// that was not a navigation key.
const functionKeyLowOffset = 250

// keypadKeys are the directions of the keypad's digits.
var keypadKeys = map[int]int{
	'1': keyEnd, '2': keyDown, '3': keyPageDown, '4': keyLeft,
	'6': keyRight, '7': keyHome, '8': keyUp, '9': keyPageUp,
}

// videoDriver is the driver selected (D_52c8): the port always draws as the Tandy
// driver T1K.DRV.
var videoDriver = tandyVideo // C: D_52c8
