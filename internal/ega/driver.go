package ega

import (
	"encoding/binary"
	"errors"
)

// Where the driver's own tables lie in T1K.DRV (file offsets).
const (
	maskPatternsAt      = 611
	tileEffectSeedAt    = 5342
	revealMasksAt       = 6625
	originSequenceAt    = 4396
	clockHandRowsAt     = 7288
	clockHandColumnsAt  = 7312
	clockHandBitsAt     = 7336
	greenHighNibblesAt  = 7660
	greenLowNibblesAt   = 7676
	maskPatternCount    = 16
	revealMaskCount     = 15
	clockPositions      = 12
	originSequenceBytes = 264
	// originFinalPassAt is where the last pass starts in the sequence; an interrupted logo
	// jumps to it to show the whole logo.
	originFinalPassAt = 191
)

// driverTables are the tables read from the driver.
type driverTables struct {
	// maskPatterns turn a nibble into a word of four pixels, each all-on or all-off.
	maskPatterns   [maskPatternCount]uint16
	tileEffectSeed uint16
	// revealMasks are the feedback masks of the reveal's shift register, by its bit count
	// less 2.
	revealMasks [revealMaskCount]uint16
	// originSequence drives the Origin logo animation.
	originSequence []byte
	// the clock tile's hands: the row and column (bytes) of each of the 12 positions, and
	// the pixel of the byte
	clockHandRows    [clockPositions]int
	clockHandColumns [clockPositions]int
	clockHandBits    [clockPositions]byte
	// greenHigh and greenLow turn a pixel pair's colours to the ending's green ones.
	greenHigh, greenLow [colourCount]byte
}

var errDriver = errors.New("T1K.DRV is too short")

func loadDriverTables(drv []byte) (*driverTables, error) {
	if len(drv) < greenLowNibblesAt+colourCount {
		return nil, errDriver
	}
	word := func(at int) uint16 { return binary.LittleEndian.Uint16(drv[at:]) }
	t := &driverTables{}
	for i := range t.maskPatterns {
		t.maskPatterns[i] = word(maskPatternsAt + 2*i)
	}
	t.tileEffectSeed = word(tileEffectSeedAt)
	for i := range t.revealMasks {
		t.revealMasks[i] = word(revealMasksAt + 2*i)
	}
	t.originSequence = append([]byte(nil), drv[originSequenceAt:originSequenceAt+originSequenceBytes]...)
	for i := range clockPositions {
		t.clockHandRows[i] = int(word(clockHandRowsAt + 2*i))
		t.clockHandColumns[i] = int(word(clockHandColumnsAt + 2*i))
		t.clockHandBits[i] = drv[clockHandBitsAt+i]
	}
	copy(t.greenHigh[:], drv[greenHighNibblesAt:])
	copy(t.greenLow[:], drv[greenLowNibblesAt:])
	return t, nil
}
