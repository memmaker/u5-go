// Package speaker is the PC speaker as a plain synth: a sound is a list of notes (a square
// wave of a frequency for a duration, or silence), rendered into one Ebiten audio stream.
package speaker

import (
	"encoding/binary"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const rate = 44100

// Note is a square wave of Hz for Dur; Hz 0 is silence.
type Note struct {
	Hz  float64
	Dur time.Duration
}

// Half is one half period of a square wave lasting us microseconds: sweeps and noise are
// sequences of such halves.
func Half(us float64) Note {
	return Note{Hz: 1e6 / (2 * us), Dur: time.Duration(us * float64(time.Microsecond))}
}

// Beep is the error beep (sound_1 in both programs): 16 half periods of 202*3 calibrated
// delay-loop passes (6.875 us each), a 120 Hz buzz of 67 ms.
var Beep = []Note{{Hz: 1e6 / (2 * beepPasses * 6.875), Dur: 16 * time.Duration(beepPasses*6875) * time.Nanosecond}}

const beepPasses = 202 * 3

var (
	mu      sync.Mutex
	pcm     []int16 // queued samples
	started bool
	frac    float64 // sample time owed between notes
	phase   float64 // of the square wave, in periods
	lp      float64 // one-pole low-pass state (the speaker is no ideal square)
	held    float64 // Hz of the tone On left sounding, 0 for none
)

// Enabled switches the effects on and off (the settings menu).
var Enabled = true

// On leaves the speaker sounding hz once the queued notes are played, until Off.
func On(hz float64) { mu.Lock(); held = hz; mu.Unlock() }

// Off silences what On started.
func Off() { mu.Lock(); held = 0; mu.Unlock() }

// Start opens the audio stream, once. Scripted test runs (U5GO_SILENT) stay silent.
func Start() {
	if started || os.Getenv("U5GO_SILENT") != "" {
		return
	}
	started = true
	ctx := audio.CurrentContext()
	if ctx == nil {
		ctx = audio.NewContext(rate)
	}
	p, err := ctx.NewPlayer(stream{})
	if err != nil {
		return
	}
	p.SetBufferSize(50 * time.Millisecond)
	p.Play()
}

// NoteHook, when set, sees every note played (the replay test records them).
var NoteHook func(Note)

// Play queues notes and returns their total length.
func Play(notes []Note) time.Duration {
	if NoteHook != nil {
		for _, n := range notes {
			NoteHook(n)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	pcm = append(pcm, Render(notes)...)
	if len(pcm) > rate*4 { // never let the queue run away
		pcm = pcm[len(pcm)-rate*4:]
	}
	var total time.Duration
	for _, n := range notes {
		total += n.Dur
	}
	return total
}

// Render turns notes into mono 44.1 kHz samples, carrying the phase across calls.
func Render(notes []Note) []int16 {
	var out []int16
	for _, n := range notes {
		f := n.Dur.Seconds()*rate + frac
		k := int(f)
		frac = f - float64(k)
		for i := 0; i < k; i++ {
			target := 0.0
			if n.Hz > 0 {
				target = 6000
				if phase >= 0.5 {
					target = -6000
				}
				phase += n.Hz / rate
				phase -= float64(int(phase))
			}
			lp += (target - lp) * 0.5
			out = append(out, int16(lp))
		}
	}
	return out
}

type stream struct{}

func (stream) Read(b []byte) (int, error) {
	frames := len(b) / 4
	mu.Lock()
	defer mu.Unlock()
	g := volume("sfx") // the speaker plays the effects
	if !Enabled {
		g = 0
	}
	for f := 0; f < frames; f++ {
		v := int16(0)
		if len(pcm) > 0 {
			v = pcm[0]
			pcm = pcm[1:]
		} else if held > 0 {
			v = Render([]Note{{Hz: held, Dur: time.Second / rate}})[0]
		} else {
			lp *= 0.99 // relax to the middle when silent
			v = int16(lp)
		}
		v = int16(float64(v) * g)
		binary.LittleEndian.PutUint16(b[f*4:], uint16(v))
		binary.LittleEndian.PutUint16(b[f*4+2:], uint16(v))
	}
	return frames * 4, nil
}
