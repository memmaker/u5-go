package ega

import "encoding/binary"

// Pictures is a picture file (the .16 and .BIT files): a count, then the offsets of the
// pictures. In a colour file the pictures come in pairs, the picture (4 bits a pixel)
// and its mask (1 bit a pixel, set where the picture is transparent; offset 0 for none).
// A picture starts with its width and height.
type Pictures []byte

func (p Pictures) word(at int) int { return int(binary.LittleEndian.Uint16(p[at:])) }

const (
	countBytes  = 2
	offsetBytes = 2
	sizeBytes   = 4
	pixelsAByte = 2
	bitsAByte   = 8
)

// Count is the number of entries (pictures and masks) in the file.
func (p Pictures) Count() int { return p.word(0) }

func (p Pictures) entry(index int) int { return p.word(countBytes + index*offsetBytes) }

// picture is a colour picture with its mask.
type picture struct {
	width, height int
	pixels        []byte
	stride        int
	mask          []byte
	maskStride    int
}

// picture returns the colour picture index (its pair of entries).
// C: IMAGE_GetImageView
func (p Pictures) picture(index int) (picture, bool) {
	if index >= p.Count() {
		return picture{}, false
	}
	at := p.entry(index * 2)
	maskAt := p.entry(index*2 + 1)
	pic := picture{width: p.word(at), height: p.word(at + 2)}
	pic.pixels = p[at+sizeBytes:]
	// the rows are padded to whole bytes of the 4 bit planes
	const planes = 4
	pic.stride = (pic.width + bitsAByte - 1) / bitsAByte * planes
	if maskAt != 0 {
		pic.mask = p[maskAt+sizeBytes:]
		pic.maskStride = (p.word(maskAt) + bitsAByte - 1) / bitsAByte
	}
	return pic, true
}

// Flags of PutPicture.
const (
	FlipVertical   = 1
	FlipHorizontal = 2
)

// PutPicture draws picture index of file at (x, y), flipped by flags.
// C: DRV_4b
func (d *Display) PutPicture(file Pictures, index, x, y, flags int) {
	pic, ok := file.picture(index)
	if !ok || pic.width <= 0 || pic.height <= 0 {
		return
	}
	shown := func(row []byte, px int) bool {
		return row == nil || row[px/bitsAByte]&(highBit>>(px%bitsAByte)) == 0
	}
	colour := func(packed byte, px int) byte {
		if px&1 == 0 {
			return packed >> 4
		}
		return packed & colourMask
	}
	byteWidth := (pic.width + 1) / pixelsAByte
	for yy := range pic.height {
		srcY, dstY := yy, y+yy
		if flags&FlipVertical != 0 {
			srcY, dstY = pic.height-yy-1, y+yy+1
		}
		line := pic.pixels[srcY*pic.stride:]
		var maskLine []byte
		if pic.mask != nil {
			maskLine = pic.mask[srcY*pic.maskStride:]
		}
		for bx := range byteWidth {
			packed := line[bx]
			left := bx * pixelsAByte
			dstX, step := x+left, 1
			if flags&FlipHorizontal != 0 {
				dstX, step = x+pic.width-pixelsAByte-left, -1
			}
			if shown(maskLine, left) {
				d.putClipped(d.Page, dstX, dstY, colour(packed, left))
			}
			if left+1 < pic.width && shown(maskLine, left+1) {
				d.putClipped(d.Page, dstX+step, dstY, colour(packed, left+1))
			}
		}
	}
}

// PutPictureMirrored draws a picture flipped left to right (and by flags).
// C: DRV_63
func (d *Display) PutPictureMirrored(file Pictures, index, x, y, flags int) {
	d.PutPicture(file, index, x, y, flags|FlipHorizontal)
}

// bitPicture is a 1-bit picture.
type bitPicture struct {
	width, height int
	bits          []byte
	stride        int
}

// C: IMAGE_GetBitImageView
func (p Pictures) bitPicture(index int) (bitPicture, bool) {
	if index >= p.Count() {
		return bitPicture{}, false
	}
	at := p.entry(index)
	pic := bitPicture{width: p.word(at), height: p.word(at + 2), bits: p[at+sizeBytes:]}
	pic.stride = (pic.width + bitsAByte - 1) / bitsAByte
	return pic, true
}

func (b bitPicture) set(x, y int) bool {
	return b.bits[y*b.stride+x/bitsAByte]&(highBit>>(x%bitsAByte)) != 0
}

// PutBitPicture draws the set pixels of 1-bit picture index in white.
// C: DRV_4e
func (d *Display) PutBitPicture(file Pictures, index, x, y int) {
	pic, ok := file.bitPicture(index)
	if !ok {
		return
	}
	for yy := range pic.height {
		for xx := range pic.width {
			if pic.set(xx, yy) {
				d.putClipped(d.Page, x+xx, y+yy, White)
			}
		}
	}
}
