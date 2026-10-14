package host

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Setting is one line of the settings window (F2): a name and the values it switches
// between.
type Setting struct {
	Name   string
	Values func() []string // the choices there are now
	Get    func() int      // the one chosen
	Set    func(int)
}

// Settings are the window's lines, registered by the program.
var Settings []Setting

// AltKeys are shortcuts held with Alt (Alt-V mutes the music, ...); they never reach the game.
var AltKeys = map[ebiten.Key]func(){}

var settingSel int

// pollAltKeys runs the Alt shortcuts pressed this tick.
func pollAltKeys() {
	if !AltDown {
		return
	}
	for k, f := range AltKeys {
		if inpututil.IsKeyJustPressed(k) {
			f()
		}
	}
}

// settingNameWidth is the width of the names' column.
const settingNameWidth = 15

// settingsLines are the window's lines and the selected one.
func settingsLines() ([]string, int) {
	var lines []string
	for _, s := range Settings {
		value := ""
		if v := s.Values(); s.Get() < len(v) {
			value = "< " + v[s.Get()] + " >"
		}
		lines = append(lines, pad(s.Name, settingNameWidth)+value)
	}
	return lines, settingSel
}

// settingsKeys handles the keys of the open settings window.
func settingsKeys() {
	n := len(Settings)
	switch {
	case repeating(ebiten.KeyArrowUp):
		settingSel = (settingSel + n - 1) % n
	case repeating(ebiten.KeyArrowDown):
		settingSel = (settingSel + 1) % n
	case repeating(ebiten.KeyArrowLeft):
		cycleSetting(settingSel, -1)
	case repeating(ebiten.KeyArrowRight), pressed(ebiten.KeySpace), pressed(ebiten.KeyEnter):
		cycleSetting(settingSel, 1)
	}
}

func cycleSetting(i, by int) {
	s := Settings[i%len(Settings)]
	if n := len(s.Values()); n > 0 {
		s.Set((s.Get() + by + n) % n)
	}
}

func pad(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
