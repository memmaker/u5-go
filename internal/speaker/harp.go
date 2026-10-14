package speaker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

// The Amiga version plays its instruments (the bards' lutes, the harpsichord) on a sampled
// harp, an IFF 8SVX sound: chunks of a tag and a big-endian length, VHDR with the sample
// rate, BODY with the 8-bit signed samples.

const (
	iffHeaderBytes = 12
	chunkHeader    = 8
	vhdrRateAt     = 12
	// harpPitch is the harp sample's pitch, in Hz, measured.
	harpPitch = 302.5
	// harpLevel scales the sample to the square waves' loudness.
	harpLevel = 6000.0 / math.MaxInt8
)

var errNot8SVX = errors.New("speaker: not an 8SVX sound")

var (
	harp     []float64 // the samples, -1..1
	harpRate float64
	// HarpEnabled plays the instruments on the harp, when it is loaded.
	HarpEnabled bool
)

// LoadHarp takes the Amiga harp sample.
func LoadHarp(b []byte) error {
	if len(b) < iffHeaderBytes || string(b[:4]) != "FORM" || string(b[8:12]) != "8SVX" {
		return errNot8SVX
	}
	var body []byte
	for at := iffHeaderBytes; at+chunkHeader <= len(b); {
		n := int(binary.BigEndian.Uint32(b[at+4:]))
		end := min(at+chunkHeader+n, len(b))
		chunk := b[at+chunkHeader : end]
		switch string(b[at : at+4]) {
		case "VHDR":
			if len(chunk) >= vhdrRateAt+2 {
				harpRate = float64(binary.BigEndian.Uint16(chunk[vhdrRateAt:]))
			}
		case "BODY":
			body = chunk
		}
		at = end + n&1 // chunks are padded to even lengths
	}
	if body == nil || harpRate == 0 {
		return errNot8SVX
	}
	harp = make([]float64, len(body))
	for i, v := range bytes.Clone(body) {
		harp[i] = float64(int8(v)) / math.MaxInt8
	}
	return nil
}

// HarpLoaded says whether there is a harp to play.
func HarpLoaded() bool { return harp != nil }

// PlayHarp queues a harp note of hz lasting dur (the sample rings out to its end if it is
// shorter) and returns dur.
func PlayHarp(hz float64, dur time.Duration) time.Duration {
	if hz <= 0 || harp == nil {
		return Play([]Note{{Dur: dur}})
	}
	step := hz / harpPitch * harpRate / rate
	n := int(dur.Seconds() * rate)
	out := make([]int16, 0, n)
	for pos := 0.0; int(pos) < len(harp)-1 && len(out) < n; pos += step {
		i := int(pos)
		f := pos - float64(i)
		out = append(out, int16((harp[i]*(1-f)+harp[i+1]*f)*harpLevel*math.MaxInt8))
	}
	for len(out) < n {
		out = append(out, 0)
	}
	mu.Lock()
	pcm = append(pcm, out...)
	lp = 0
	mu.Unlock()
	return dur
}
