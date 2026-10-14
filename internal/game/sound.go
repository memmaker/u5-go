package game

import (
	"math/rand/v2"
	"time"

	"github.com/memmaker/u5-go/internal/host"
	"github.com/memmaker/u5-go/internal/speaker"
)

// The PC speaker effects (segment 2000, written in assembly in the original). They play as
// square waves on the speaker synth and take the time the effect lasts by the
// decompilation's measurements.

// pitClock is the PC's timer clock: a timer divisor d sounds pitClock/d Hz.
const pitClock = 1193182

// Speaker frequencies beyond hearing are dropped.
const (
	lowestHz  = 20
	highestHz = 15000
)

// audible keeps hz in range, or 0 for silence.
func audible(hz float64) float64 {
	if hz < lowestHz || hz > highestHz {
		return 0
	}
	return hz
}

// shortEffect is the length (in the effect's units) up to which an effect plays without the
// game waiting for it, as the decompilation measures.
const shortEffect = 100

// sound plays notes and, for effects longer than shortEffect, waits for them, as the original
// busy-loops through its effects.
func sound(notes []speaker.Note, duration int) {
	if !soundOn { // Ctrl-S
		return
	}
	if d := speaker.Play(notes); d > 0 && duration > shortEffect {
		host.Wait(d)
	}
}

// effectLength is duration units of per each, at least a millisecond.
func effectLength(duration int, per time.Duration) time.Duration {
	return max(time.Duration(duration)*per, time.Millisecond)
}

// pulse plays a pulse-width tone: freq is the timer divisor of its pitch, the width falls or
// rises by widthStep, colouring it; the speaker synth has only the pitch.
// C: ULTIMA_2192_AudioPulse
func pulse(freq, delay, duration, width, widthStep int) {
	const unitsPerMillisecond = 20
	if freq <= 0 {
		return
	}
	sound([]speaker.Note{{Hz: audible(float64(pitClock) / float64(freq)),
		Dur: effectLength(duration, time.Millisecond/unitsPerMillisecond)}}, duration)
}

// instrumentNote plays a note of the bards' lutes or the harpsichord: a pulse, as the PC
// plays it, or the Amiga's harp when the player has it and chose it.
func instrumentNote(freq, duration, width, widthStep int) {
	const unitsPerMillisecond = 20
	if !speaker.HarpEnabled || !speaker.HarpLoaded() || freq <= 0 {
		pulse(freq, 1, duration, width, widthStep)
		return
	}
	if !soundOn {
		return
	}
	d := speaker.PlayHarp(audible(float64(pitClock)/float64(freq)), effectLength(duration, time.Millisecond/unitsPerMillisecond))
	if duration > shortEffect {
		host.Wait(d)
	}
}

// whiteNoise plays noise: random half periods up to limit microseconds.
// C: ULTIMA_223c_AudioWhiteNoise
func whiteNoise(rate, duration, limit int) {
	const unitsPerMillisecond = 10
	const shortestHalf = 20 // microseconds
	total := effectLength(duration, time.Millisecond/unitsPerMillisecond)
	var notes []speaker.Note
	for t := time.Duration(0); t < total; {
		n := speaker.Half(float64(shortestHalf + rand.IntN(max(limit/rate/10, 1)+shortestHalf)))
		notes = append(notes, n)
		t += n.Dur
	}
	sound(notes, duration)
}

// tone plays a tone of freq Hz.
// C: ULTIMA_22c0_AudioTone
func tone(freq, duration int) {
	const tenthsPerUnit = 8
	sound([]speaker.Note{{Hz: audible(float64(freq)), Dur: effectLength(duration, tenthsPerUnit*time.Millisecond/10)}}, duration)
}

// sweepTone plays a tone sliding between two frequencies in Hz, step by step.
// C: ULTIMA_43ae
func sweepTone(from, to, step, duration int) {
	const tenthsPerUnit = 15
	if step <= 0 || duration <= 0 {
		return
	}
	inc := (to - from) * step / duration
	var notes []speaker.Note
	for i, f := 0, from; i < duration; i, f = i+step, f+inc {
		notes = append(notes, speaker.Note{Hz: audible(float64(f)), Dur: time.Duration(step*tenthsPerUnit/10) * time.Millisecond})
	}
	sound(notes, duration)
}

// speakerOn starts the speaker at a frequency in Hz.
// C: ULTIMA_22e2_PcspkOn
func speakerOn(freq int) {
	if soundOn {
		speaker.On(audible(float64(freq)))
	}
}

// speakerOff silences the speaker.
// C: ULTIMA_230e_PcspkOff
func speakerOff() { speaker.Off() }
