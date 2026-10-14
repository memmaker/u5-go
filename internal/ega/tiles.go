package ega

import "encoding/binary"

// A tile is 16x16 pixels, two pixels a byte, high nibble left: 128 bytes.
const (
	TileSize     = 16
	tileRowBytes = TileSize / 2
	tileBytes    = TileSize * tileRowBytes
	tileWords    = tileBytes / 2
)

// Tiles the driver's animations work on (by number in TILES.16).
const (
	tileWater1          = 1
	tileWater2          = 2
	tileWater3          = 3
	tileGrass           = 5
	tileFlagCastle      = 18
	tileFlagTown        = 20
	tileFlagKeep        = 21
	tileShipFlags       = 32
	tileWaterfronts     = 48
	tileFountains       = 52
	tileSign            = 62
	tileBrickFloor      = 68
	tileSwamp           = 76
	tileShoreline       = 96
	tileBridgeAcross    = 106
	tileBridgeAlong     = 107
	tileBridgeEnds      = 108
	tileShoreMasks      = 112
	tileBridgeAlongMask = 123
	tileLava            = 143
	tileWaterSpouts     = 176
	tileFireplaces      = 188
	tileWaterSpoutMasks = 192
	tileFirelight       = 194
	tileBlackLadder     = 200
	tileFireplaceMasks  = 204
	tileFountainMasks   = 208
	tileMoongate        = 220
	tileFire            = 222
	tileFountainWater   = 228
	tileClock           = 250
	tileLampFlash       = 256
	tileLamp            = 264
	// tileScratch is the magic circle, which the moongate's frames are built in.
	tileScratch     = 278
	tileLadderMan   = 279
	tileFireFlags   = 289
	tileTorchFlags  = 291
	tileCandleFlags = 301
	tileBrazierFlag = 303
	tileCandles     = 436
	// tileNoise are the four tiles of noise patterns made each frame.
	tileNoise      = 488
	tileNoiseFine  = 489
	tileNoiseSpots = 490
	tileNoiseFlame = 491
)

func tileAt(tile int) int { return tile * tileBytes }

// LoadTiles takes the tile set (TILES.16 unpacked).
// C: DRV_48_LoadTileset
func (d *Display) LoadTiles(tiles []byte) { d.tiles = tiles }

// UnloadTiles drops the tile set.
// C: DRV_5a_FreeTileset
func (d *Display) UnloadTiles() { d.tiles = nil }

// PutTile draws tile in tile cell (column, row), shifted by (xOffset, yOffset) pixels.
// C: DRV_51_PutTile
func (d *Display) PutTile(column, row, tile, xOffset, yOffset int) {
	src := d.tiles[tileAt(tile):]
	left, top := column*TileSize+xOffset, row*TileSize+yOffset
	for y := range TileSize {
		for x := 0; x < TileSize; x += 2 {
			packed := src[y*tileRowBytes+x/2]
			d.put(d.Page, left+x, top+y, packed>>4)
			d.put(d.Page, left+x+1, top+y, packed&colourMask)
		}
	}
}

// PutMoongate draws the moongate risen visibleRows pixels out of the floor (red brick in
// a dungeon when brickFloor, else grass), built in the scratch tile.
// C: GRAP_BUF_PutAnimatedMoongateTile
func (d *Display) PutMoongate(column, row, visibleRows int, brickFloor bool, xOffset, yOffset int) {
	scratch := d.tiles[tileAt(tileScratch) : tileAt(tileScratch)+tileBytes]
	var backup [tileBytes]byte
	copy(backup[:], scratch)
	floor := tileGrass
	if brickFloor {
		floor = tileBrickFloor
	}
	copy(scratch, d.tiles[tileAt(floor):])
	if visible := visibleRows * tileRowBytes; visible != 0 {
		copy(scratch[tileBytes-visible:], d.tiles[tileAt(tileMoongate):tileAt(tileMoongate)+visible])
	}
	d.PutTile(column&lowByte, row&lowByte, tileScratch, xOffset, yOffset)
	copy(scratch, backup[:])
}

const lowByte = 255

// tileRevealMask is the feedback mask of the 8-bit shift register that picks the pixels of
// a tile being revealed.
const tileRevealMask = 184

// PutTileRevealStep draws one more pixel of tile, picked by a shift register so that the
// tile appears scattered; step 0 draws the top-left pixel, step 1 restarts the sequence.
// C: GRAP_BUF_PutTileRevealStep
func (d *Display) PutTileRevealStep(column, row, tile, step, xOffset, yOffset int) {
	if step == 0 {
		d.putTilePixel(column, row, tile, 0, 0, xOffset, yOffset)
		return
	}
	seed := d.effects.tileRevealSeed
	if step == 1 || seed == 0 {
		seed = 1
	}
	d.putTilePixel(column, row, tile, int(seed>>4), int(seed&colourMask), xOffset, yOffset)
	d.effects.tileRevealSeed = stepShiftRegister(seed, tileRevealMask)
}

func (d *Display) putTilePixel(column, row, tile, x, y, xOffset, yOffset int) {
	packed := d.tiles[tileAt(tile)+y*tileRowBytes+x/2]
	colour := packed >> 4
	if x&1 != 0 {
		colour = packed & colourMask
	}
	d.put(d.Page, column*TileSize+xOffset+x, row*TileSize+yOffset+y, colour)
}

// stepShiftRegister advances a Galois shift register.
func stepShiftRegister(seed, mask uint16) uint16 {
	carry := seed & 1
	seed >>= 1
	if carry != 0 {
		seed ^= mask
	}
	return seed
}

// word access to the tile set (the driver's animations work a word at a time)
func (d *Display) word(at int) uint16       { return binary.LittleEndian.Uint16(d.tiles[at:]) }
func (d *Display) setWord(at int, v uint16) { binary.LittleEndian.PutUint16(d.tiles[at:], v) }

// shiftDown rolls a tile down a row (flowing water).
func (d *Display) shiftDown(tile int) {
	t := d.tiles[tileAt(tile) : tileAt(tile)+tileBytes]
	var last [tileRowBytes]byte
	copy(last[:], t[tileBytes-tileRowBytes:])
	copy(t[tileRowBytes:], t[:tileBytes-tileRowBytes])
	copy(t, last[:])
}

// swapPairs swaps two pixel words of row at with the ones two rows below (a flag waving).
func (d *Display) swapPairs(at int) {
	const twoRows = 2 * tileRowBytes
	for _, k := range []int{0, 2} {
		a, b := d.word(at+k), d.word(at+twoRows+k)
		d.setWord(at+k, b)
		d.setWord(at+twoRows+k, a)
	}
}

func (d *Display) swapPairsRight(at int) { d.swapPairs(at + halfRow) }

func (d *Display) swapPairsTwice(at int) {
	d.swapPairs(at)
	d.swapPairsRight(at)
}

const halfRow = tileRowBytes / 2

// andWords keeps the bits of mask in words words from at.
func (d *Display) andWords(at int, mask uint16, words int) {
	for i := range words {
		d.setWord(at+2*i, d.word(at+2*i)&mask)
	}
}

// cutOut clears in words words from at the bits set from mask.
func (d *Display) cutOut(at, mask, words int) {
	for i := range words {
		d.setWord(at+2*i, d.word(at+2*i)&^d.word(mask+2*i))
	}
}

// xorMasked flips in count bytes from at the bits of the noise the mask lets through.
func (d *Display) xorMasked(at, mask, noise, count int) {
	for i := range count {
		d.tiles[at+i] ^= d.tiles[noise+i] & d.tiles[mask+i]
	}
}

// xorMaskedTiles is xorMasked over whole tiles, the same noise tile for each.
func (d *Display) xorMaskedTiles(at, mask, noise, tiles int) {
	for t := range tiles {
		d.xorMasked(at+t*tileBytes, mask+t*tileBytes, noise, tileBytes)
	}
}

// mixIn sets in each of tiles tiles from at the bits of src the mask lets through.
func (d *Display) mixIn(at, mask, src, tiles int) {
	for t := range tiles {
		for off := 0; off < tileBytes; off += 2 {
			dst := at + t*tileBytes + off
			d.setWord(dst, d.word(dst)|d.word(mask+t*tileBytes+off)&d.word(src+off))
		}
	}
}

// storeOr sets each of tiles tiles from at to the tile at src (same place) or'ed with mask.
func (d *Display) storeOr(at, src, mask, tiles int) {
	for t := range tiles {
		for off := 0; off < tileBytes; off += 2 {
			d.setWord(at+t*tileBytes+off, d.word(src+t*tileBytes+off)|d.word(mask+off))
		}
	}
}

func (d *Display) invertWords(at, words int) {
	for i := range words {
		d.setWord(at+2*i, ^d.word(at+2*i))
	}
}

// mergeMasked puts into the tile at at the noise where the tile at src is set.
func (d *Display) mergeMasked(at, src, noise int) {
	d.invertWords(src, tileWords)
	for i := range tileBytes {
		d.tiles[at+i] &= d.tiles[src+i]
	}
	d.invertWords(src, tileWords)
	for i := range tileBytes {
		d.tiles[at+i] = d.tiles[noise+i]&d.tiles[src+i] | d.tiles[at+i]
	}
}

// RNG constants of the driver's noise (the game's own random generator).
const (
	noiseAdd    = 37448
	noiseRotate = 3
	noiseXor    = 37448
	noiseOffset = 17
	wordBits    = 16
)

// Pixel masks: which of the four pixels of a word (bits of each pixel's colour) the noise
// patterns keep.
const (
	maskEvenBits  = 43690
	maskNoBit1    = 56797
	maskHighBits  = 52428
	maskOuterBits = 39321
)

// makeNoise fills the four noise tiles with new random patterns.
// C: AnimateTile_GenerateMasks (T1K:156c)
func (d *Display) makeNoise() {
	at := tileAt(tileNoise)
	for range tileBytes / 4 {
		s := d.effects.tileEffectSeed + noiseAdd
		s = s>>noiseRotate | s<<(wordBits-noiseRotate)
		s = (s ^ noiseXor) + noiseOffset
		d.effects.tileEffectSeed = s
		random := byte(s)
		high := d.driver.maskPatterns[random>>4]
		low := d.driver.maskPatterns[random&colourMask]
		for _, w := range []uint16{low, high} {
			d.setWord(at, w&maskEvenBits)
			d.setWord(at+tileBytes, w&maskNoBit1)
			d.setWord(at+2*tileBytes, w&maskNoBit1&maskHighBits)
			d.setWord(at+3*tileBytes, w&maskOuterBits)
			at += 2
		}
	}
}

// maskNoLightBlue drops colour bit 0 and 3 (light blue) from the waves.
const maskNoLightBlue = 26214

// AnimateTiles advances the tile animations a frame: water flows, shores and bridges
// ripple, fountains, fires and lamps flicker, flags wave.
// C: AnimateTileset (DRV_60)
func (d *Display) AnimateTiles() {
	if d.tiles == nil {
		return
	}
	d.makeNoise()
	for _, t := range []int{tileWater1, tileWater2, tileWater3, tileLava} {
		d.shiftDown(t)
	}
	const shoreTiles, bridgeEndTiles, fountainTiles, spoutTiles = 10, 4, 4, 4
	const bridgeEdgeRows = 2
	d.andWords(tileAt(tileShoreline), maskNoLightBlue, shoreTiles*tileWords)
	clear(d.tiles[tileAt(tileBridgeAcross) : tileAt(tileBridgeAcross)+bridgeEdgeRows*tileRowBytes])
	clear(d.tiles[tileAt(tileBridgeAcross+1)-bridgeEdgeRows*tileRowBytes : tileAt(tileBridgeAcross+1)])
	d.cutOut(tileAt(tileBridgeAlong), tileAt(tileBridgeAlongMask), tileWords)
	d.andWords(tileAt(tileBridgeEnds), maskNoLightBlue, bridgeEndTiles*tileWords)
	const shoreAndBridgeTiles = 16
	d.mixIn(tileAt(tileShoreline), tileAt(tileShoreMasks), tileAt(tileWater3), shoreAndBridgeTiles)
	d.mergeMasked(tileAt(tileLamp), tileAt(tileLampFlash), tileAt(tileNoiseFlame))
	d.mergeMasked(tileAt(tileCandles), tileAt(tileLampFlash), tileAt(tileNoiseFine))
	d.cutOut(tileAt(tileFountains), tileAt(tileFountainMasks), fountainTiles*tileWords)
	d.mixIn(tileAt(tileFountains), tileAt(tileFountainMasks), tileAt(tileWater3), fountainTiles)
	d.storeOr(tileAt(tileFountainWater), tileAt(tileFountainMasks), tileAt(tileWater3), fountainTiles)
	d.xorMaskedTiles(tileAt(tileWaterSpouts), tileAt(tileWaterSpoutMasks), tileAt(tileNoiseSpots), spoutTiles)
	d.xorMaskedTiles(tileAt(tileFireplaces), tileAt(tileFireplaceMasks), tileAt(tileNoiseSpots), spoutTiles)
	d.xorMasked(tileAt(tileFire), tileAt(tileFirelight), tileAt(tileNoiseFlame), tileBytes)

	const flagRow = 2 * tileRowBytes // the flags' rows start a quarter in
	seed := d.effects.tileEffectSeed
	const (
		waveCastle = 1 << iota
		waveTown
		waveKeep
		waveSign
		waveFire
		waveTorch
		waveCandle
	)
	if seed&waveCastle != 0 {
		d.swapPairsTwice(tileAt(tileFlagCastle))
	}
	if seed&waveTown != 0 {
		d.swapPairs(tileAt(tileFlagTown) + flagRow)
	}
	if seed&waveKeep != 0 {
		d.swapPairsTwice(tileAt(tileFlagKeep))
	}
	if seed&waveSign != 0 {
		d.swapPairsTwice(tileAt(tileSign))
	}
	if seed&waveFire != 0 {
		d.swapPairsRight(tileAt(tileFireFlags) + tileRowBytes)
	}
	if seed&waveTorch != 0 {
		d.swapPairs(tileAt(tileTorchFlags) + tileRowBytes)
	}
	if seed&waveCandle != 0 {
		d.swapPairsRight(tileAt(tileCandleFlags) + tileRowBytes)
	}
	if seed&waveCastle != 0 {
		d.swapPairs(tileAt(tileBrazierFlag) + tileRowBytes)
	}
}

// Time-of-day modes of UpdateTimeTiles.
const (
	// TimeLadderBreak cuts the ladders' bottom rung, TimeLadderMend puts it back (moonlight
	// ladders in towns).
	TimeLadderBreak = 0
	TimeLadderMend  = 1
	// TimeClock sets the clock tiles' hands.
	TimeClock = 2
	// TimeNight swaps colour planes of the outdoor tiles.
	TimeNight = 3
	// TimeGreen turns the ending's tiles green.
	TimeGreen = 4
)

// UpdateTimeTiles changes the tiles that show the time (DRV_6c).
// C: AnimateTimeTileset
func (d *Display) UpdateTimeTiles(mode int, hour, minute int) {
	const lastRow = tileBytes - tileRowBytes
	switch mode {
	case TimeLadderBreak:
		for i, t := range []int{tileBlackLadder, tileLadderMan} {
			row := d.tiles[tileAt(t)+lastRow : tileAt(t)+tileBytes]
			copy(d.effects.ladderRows[i][:], row)
			copy(row, d.tiles[tileAt(t):])
		}
	case TimeLadderMend:
		for i, t := range []int{tileBlackLadder, tileLadderMan} {
			copy(d.tiles[tileAt(t)+lastRow:], d.effects.ladderRows[i][:])
		}
	case TimeClock:
		d.clearClockFace(tileAt(tileClock) + 2*tileRowBytes)
		d.clearClockFace(tileAt(tileClock+1) + 2*tileRowBytes)
		const hoursOnDial, minutesAHour = 12, 5
		if hour >= hoursOnDial {
			hour -= hoursOnDial
		}
		d.drawClockHand(hour)
		d.drawClockHand(minute / minutesAHour)
	case TimeGreen:
		for _, t := range greenTiles {
			for i := range tileBytes {
				v := &d.tiles[tileAt(t)+i]
				*v = d.driver.greenHigh[*v>>4] | d.driver.greenLow[*v&colourMask]
			}
		}
	default:
		for _, block := range nightTiles {
			d.swapPlanes(tileAt(block.first), block.count)
		}
	}
}

// greenTiles are the tiles the ending turns green.
var greenTiles = []int{68, 92, 93, 144, 146, 148, 150, 155, 171, 172, 175, 176, 177, 191, 220, 264, 270, 282, 312, 313, 314, 315}

// nightTiles are the runs of tiles whose colour planes swap with the time of day.
var nightTiles = []struct{ first, count int }{
	{tileGrass, 1}, {30, 1}, {31, 1}, {tileSwamp, 1}, {202, 1}, {tileShipFlags, 7}, {tileWaterfronts, 8}, {tileShoreline, 16},
}

// clockFaceBytes are the bytes of a clock tile's face (from its third row) that the hands
// cross, the low pixel of each set back to the face colour.
var clockFaceBytes = []int{3, 11, 12, 19}

const (
	clockFaceColour = 4
	clockFacePair   = clockFaceColour<<4 | clockFaceColour
	// clockCentreBytes are the face's centre bytes, both pixels face coloured.
	clockCentreTop    = 4
	clockCentreBottom = 20
	highNibble        = 240
)

func (d *Display) clearClockFace(at int) {
	for _, off := range clockFaceBytes {
		d.tiles[at+off] = d.tiles[at+off]&highNibble | clockFaceColour
	}
	d.tiles[at+clockCentreTop] = clockFacePair
	d.tiles[at+clockCentreBottom] = clockFacePair
}

func (d *Display) drawClockHand(position int) {
	position %= clockPositions
	at := tileAt(tileClock) + d.driver.clockHandRows[position] + d.driver.clockHandColumns[position]
	d.tiles[at] |= d.driver.clockHandBits[position]
	d.tiles[at+tileBytes] |= d.driver.clockHandBits[position]
}

// Colour planes swapped at night: bit 1 and bit 2 of each pixel trade places.
const (
	planeBit1 = 8738
	planeBit2 = 17476
	planeKeep = 39321
)

func (d *Display) swapPlanes(at, tiles int) {
	for i := 0; i < tiles*tileBytes; i += 2 {
		v := d.word(at + i)
		d.setWord(at+i, (v&planeBit1)<<1|(v&planeBit2)>>1|v&planeKeep)
	}
}

// TilePixel is the colour of pixel (x, y) of tile, 0 without a tile set.
func (d *Display) TilePixel(tile, x, y int) byte {
	at := tileAt(tile) + y*tileRowBytes + x/2
	if at >= len(d.tiles) {
		return 0
	}
	if x&1 != 0 {
		return d.tiles[at] & colourMask
	}
	return d.tiles[at] >> 4
}
