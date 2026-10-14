package music

import (
	"encoding/binary"
	"errors"
	"sort"
)

// XMI (Miles' extended MIDI) is IFF: FORM XDIR { INFO } then CAT XMID { FORM XMID { TIMB? EVNT } ... },
// chunk lengths big-endian and padded to even. An EVNT chunk is a MIDI track with two twists:
// a delay is a run of bytes below 80h that add up (no variable-length quantity), and a note-on
// carries its duration as a variable-length quantity instead of a matching note-off. Time is in
// 120 Hz ticks; MIDIFORM already folded the tempo into the delays.

// tickHz is XMI's fixed tick rate.
const tickHz = 120

// event is a MIDI channel message due at tick; status 0 is a no-op.
type event struct {
	tick    int
	status  byte
	d1, d2  byte
	noteOff bool // sorts before note-ons of the same tick
}

// parseXMI returns the sequences of an XMI file, in order.
func parseXMI(b []byte) ([][]event, error) {
	var songs [][]event
	var walk func(b []byte) error
	walk = func(b []byte) error {
		for len(b) >= 8 {
			id, n := string(b[:4]), int(binary.BigEndian.Uint32(b[4:8]))
			if 8+n > len(b) {
				return errors.New("xmi: truncated chunk " + id)
			}
			body := b[8 : 8+n]
			switch id {
			case "FORM", "CAT ":
				if len(body) < 4 {
					return errors.New("xmi: short " + id)
				}
				if err := walk(body[4:]); err != nil {
					return err
				}
			case "EVNT":
				ev, err := parseEVNT(body)
				if err != nil {
					return err
				}
				songs = append(songs, ev)
			}
			b = b[8+n+n&1:]
		}
		return nil
	}
	if err := walk(b); err != nil {
		return nil, err
	}
	if len(songs) == 0 {
		return nil, errors.New("xmi: no sequences")
	}
	return songs, nil
}

// MIDI: a status byte has its top bit set; its high four bits are the message and its low
// four the channel. Numbers (an XMI note's length) are 7 bits a byte, the top bit set on all
// but the last.
const (
	statusBit           = 128
	midiNoteOff         = 8
	midiNoteOn          = 9
	midiKeyPressure     = 10
	midiController      = 11
	midiProgram         = 12
	midiChannelPressure = 13
	midiPitchBend       = 14
	midiMeta            = 255
	metaEndOfTrack      = 47
)

func midiMessage(status byte) byte { return status >> 4 }
func midiChannel(status byte) byte { return status % 16 }

func parseEVNT(d []byte) ([]event, error) {
	var out []event
	tick, p := 0, 0
	errShort := errors.New("xmi: truncated event")
	vlq := func() (int, bool) {
		v := 0
		for p < len(d) {
			c := d[p]
			p++
			v = v<<7 | int(c%statusBit)
			if c < statusBit {
				return v, true
			}
		}
		return 0, false
	}
	for p < len(d) {
		s := d[p]
		if s < statusBit { // a delay, not a status
			tick += int(s)
			p++
			continue
		}
		p++
		switch midiMessage(s) {
		case midiNoteOn:
			if p+2 > len(d) {
				return nil, errShort
			}
			key, vel := d[p], d[p+1]
			p += 2
			dur, ok := vlq()
			if !ok {
				return nil, errShort
			}
			out = append(out, event{tick: tick, status: s, d1: key, d2: vel},
				event{tick: tick + dur, status: midiNoteOff<<4 | midiChannel(s), d1: key, noteOff: true})
		case midiNoteOff, midiKeyPressure, midiController, midiPitchBend:
			if p+2 > len(d) {
				return nil, errShort
			}
			// Controllers 110-120 are AIL's own (channel locks, loops, callbacks).
			if midiMessage(s) != midiController || d[p] < 110 || d[p] > 120 {
				out = append(out, event{tick: tick, status: s, d1: d[p], d2: d[p+1]})
			}
			p += 2
		case midiProgram, midiChannelPressure:
			if p+1 > len(d) {
				return nil, errShort
			}
			out = append(out, event{tick: tick, status: s, d1: d[p]})
			p++
		default: // system exclusive, or a meta event
			if s == midiMeta {
				if p >= len(d) {
					return nil, errShort
				}
				typ := d[p]
				p++
				if typ == metaEndOfTrack { // end of track: a no-op event marks where the loop restarts
					out = append(out, event{tick: tick})
					p = len(d)
					continue
				}
			}
			n, ok := vlq()
			if !ok || p+n > len(d) {
				return nil, errShort
			}
			p += n
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].tick != out[j].tick {
			return out[i].tick < out[j].tick
		}
		return out[i].noteOff && !out[j].noteOff
	})
	return out, nil
}
