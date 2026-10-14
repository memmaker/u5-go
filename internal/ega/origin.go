package ega

import "time"

// The Origin logo animation: the logo (drawn in the back buffer) folds out of a line, one
// band of rows at a time. The driver's sequence holds the pass count, then for each pass
// the source row, the screen row and the row count of the band, and a list of row numbers
// in groups ended by groupEnd: each group shows (or hides) more rows of the band.
const (
	groupEnd       = 255
	groupsAPass    = 6
	maxOriginRows  = 128
	originBlueMask = 9
	maxLogoRow     = Height
)

type originState struct {
	seq                []byte
	cursor             int
	listStart, listLen int
	sourceY, targetY   int
	rows               int
	visible            [maxOriginRows]bool
}

func (o *originState) read() byte {
	if o.cursor >= len(o.seq) {
		return 0
	}
	b := o.seq[o.cursor]
	o.cursor++
	return b
}

func (o *originState) readHeader() {
	o.sourceY = int(o.read())
	o.targetY = int(o.read())
	o.rows = int(o.read())
}

func (o *originState) readList() {
	n := int(o.read())
	o.listLen = n - 1
	o.listStart = o.cursor
	o.cursor += n
}

func (o *originState) listAt(i int) byte {
	if i < 0 || o.listStart+i >= len(o.seq) {
		return groupEnd
	}
	return o.seq[o.listStart+i]
}

// AnimateOrigin plays the logo animation; it reports whether a key press cut it short (the
// logo is then shown whole).
// C: GRAP_BUF_AnimateOriginLogo (DRV_6f)
func (d *Display) AnimateOrigin() bool {
	o := &originState{seq: d.driver.originSequence}
	passes := int(o.read())
	upward := true
	for passes != 0 {
		o.readHeader()
		o.readList()
		upward = !upward
		if d.originShowGroups(o, upward) {
			d.originFinish(o)
			return true
		}
		passes--
		if passes == 0 {
			break
		}
		if d.originHideGroups(o, upward) {
			d.originFinish(o)
			return true
		}
	}
	return false
}

func (d *Display) originClearRow(y int) {
	if y < maxLogoRow {
		clear(d.pages[Screen][y*Width : (y+1)*Width])
	}
}

func (d *Display) originCopyRow(srcY, dstY int) {
	if srcY < maxLogoRow && dstY < maxLogoRow {
		for x := range Width {
			d.pages[Screen][dstY*Width+x] = d.pages[BackBuffer][srcY*Width+x] & originBlueMask
		}
	}
}

// originDraw draws the band: its shown rows packed together in the middle, cleared rows
// above and below. Upward bands grow from their bottom.
func (d *Display) originDraw(o *originState, upward bool) bool {
	y, step := o.targetY, 1
	if upward {
		y, step = o.targetY+o.rows, -1
	}
	hidden, shown := 0, 0
	for i := range o.rows {
		if o.visible[i] {
			shown++
		} else {
			hidden++
		}
	}
	for range hidden / 2 {
		d.originClearRow(y)
		y += step
	}
	scan := 0
	for range shown {
		for scan < o.rows && !o.visible[scan] {
			scan++
		}
		if scan >= o.rows {
			break
		}
		d.originCopyRow(o.sourceY+scan, y)
		scan++
		y += step
	}
	hidden++
	for range hidden / 2 {
		d.originClearRow(y)
		y += step
	}
	d.present()
	if d.KeyHit() {
		return true
	}
	const waitPerRow, baseWait = 3, 15
	d.wait(time.Duration(2*(o.rows/waitPerRow+baseWait)) * time.Millisecond)
	return false
}

func (d *Display) originShowGroups(o *originState, upward bool) bool {
	clear(o.visible[:o.rows])
	pos := o.listLen - 1
	for count := groupsAPass; count >= 0; {
		v := o.listAt(pos)
		pos--
		if v == groupEnd {
			if d.originDraw(o, upward) {
				return true
			}
			count--
		} else if int(v) < o.rows {
			o.visible[v] = true
		}
	}
	return false
}

func (d *Display) originHideGroups(o *originState, upward bool) bool {
	pos := 1
	for count := groupsAPass; count > 0; {
		v := o.listAt(pos)
		pos++
		if v == groupEnd {
			if d.originDraw(o, upward) {
				return true
			}
			count--
		} else if int(v) < o.rows {
			o.visible[v] = false
		}
	}
	return false
}

func (d *Display) originFinish(o *originState) {
	o.cursor = originFinalPassAt
	o.readHeader()
	for i := range o.rows {
		o.visible[i] = true
	}
	d.originDraw(o, false)
}
