package ega

import "time"

// effectState is the state the effects keep between calls.
type effectState struct {
	tileRevealSeed uint16
	tileEffectSeed uint16
	ladderRows     [2][tileRowBytes]byte
	bannerFrame    int
}

// tick is the PC timer's tick, the unit of the original's waits.
const tick = 55 * time.Millisecond

// revealState walks the pixels of a rectangle in a scattered order: a maximal-length shift
// register, as wide as needed to number every pixel, counts through them all once.
type revealState struct {
	state, mask     uint16
	maxState, total int
	produced        int
}

const minRevealBits = 2

func (d *Display) newReveal(width, height int, seed uint16) *revealState {
	total := width * height
	bits := minRevealBits
	maxState := 1<<bits - 1
	for maxState < total {
		bits++
		maxState = 1<<bits - 1
	}
	return &revealState{
		state:    seed % uint16(maxState),
		mask:     d.driver.revealMasks[bits-minRevealBits],
		maxState: maxState,
		total:    total,
	}
}

// next returns the next pixel's index, false when all are done.
func (r *revealState) next() (int, bool) {
	if r.produced >= r.total {
		return 0, false
	}
	for guard := 0; guard <= r.maxState; guard++ {
		index := int(r.state) - 1
		r.state = stepShiftRegister(r.state, r.mask)
		if index < r.total {
			r.produced++
			return index, true
		}
	}
	return 0, false
}

// revealBatch is how many pixels the reveal sets between frames.
const revealBatch = 512

// Reveal copies the rectangle of the back buffer to the screen pixel by pixel in a
// scattered order; a key press finishes it at once.
// C: GRAP_BUF_TransferPage_Reveal (DRV_66)
func (d *Display) Reveal(x1, y1, x2, y2 int) {
	width, height := x2-x1+1, y2-y1+1
	r := d.newReveal(width, height, 1)
	for {
		index, ok := r.next()
		if !ok {
			break
		}
		x, y := x1+index%width, y1+index/width
		d.pages[Screen][y*Width+x] = d.pages[BackBuffer][y*Width+x]
		if r.produced%revealBatch == revealBatch-1 {
			d.present()
			if d.KeyHit() {
				d.Copy(BackBuffer, Screen, x1, y1, x2, y2, x1, y1)
				return
			}
			d.wait(tick)
		}
	}
	d.present()
}

// wait lets time pass, showing the screen.
func (d *Display) wait(t time.Duration) { waitHost(t) }

// The "Warriors of Destiny" banner: four frames of it lie in the back buffer one under
// another; the screen shows them in turn.
const (
	bannerLeft    = 16
	bannerWidth   = 288
	bannerHeight  = 49
	bannerSpacing = 50
	bannerFrames  = 4
	bannerScreenY = 65
	bannerMask    = 13568
	bannerSteps   = 32768
	// bannerFrameEvery is how many shift register steps pass between banner frames.
	bannerFrameEvery = 256
	bannerBatch      = 64
)

// ShowNextBannerFrame shows the banner's next frame.
// C: GRAP_BUF_ShowNextWDFrame (DRV_69)
func (d *Display) ShowNextBannerFrame() {
	srcY := d.effects.bannerFrame * bannerSpacing
	for row := range bannerHeight {
		copy(d.pages[Screen][(bannerScreenY+row)*Width:(bannerScreenY+row+1)*Width], d.pages[BackBuffer][(srcY+row)*Width:])
	}
	d.effects.bannerFrame = (d.effects.bannerFrame + 1) % bannerFrames
}

// AnimateBanner makes the banner appear pixel by pixel: first the pixels outside the
// shape of mask (picture 0 of file), then the ones inside, all the while cycling the
// frames. A key press finishes it at once.
// C: GRAP_BUF_AnimateWD (DRV_69)
func (d *Display) AnimateBanner(file Pictures) {
	mask, ok := file.bitPicture(0)
	if !ok {
		return
	}
	back := d.pages[BackBuffer]
	backup := append([]byte(nil), back...)
	clear(back)
	aborted := d.bannerPass(mask, backup, false)
	if !aborted {
		aborted = d.bannerPass(mask, backup, true)
	}
	if !aborted {
		d.copyBannerPixel(backup, 0, 0)
	}
	if aborted {
		copy(back, backup)
	}
	d.ShowNextBannerFrame()
	d.present()
	copy(back, backup)
}

func (d *Display) copyBannerPixel(backup []byte, x, y int) {
	x += bannerLeft
	for frame := range bannerFrames {
		at := (frame*bannerSpacing+y)*Width + x
		d.pages[BackBuffer][at] = backup[at]
	}
}

// Waits between batches of the banner's pixels.
const (
	bannerOutsideWait = 33 * time.Millisecond
	bannerInsideWait  = 120 * time.Millisecond
)

func (d *Display) bannerPass(mask bitPicture, backup []byte, inside bool) bool {
	seed := uint16(1)
	countdown := bannerFrameEvery
	generated := 0
	for range bannerSteps {
		y, x := int(seed)/bannerWidth, int(seed)%bannerWidth
		if y < bannerHeight && x < mask.width && y < mask.height && mask.set(x, y) == inside {
			d.copyBannerPixel(backup, x, y)
			if generated%bannerBatch == bannerBatch-1 {
				d.present()
				if d.KeyHit() {
					return true
				}
				if inside {
					d.wait(bannerInsideWait)
				} else {
					d.wait(bannerOutsideWait)
				}
			}
			generated++
		}
		countdown--
		if countdown == 0 {
			d.ShowNextBannerFrame()
			d.present()
			countdown = bannerFrameEvery
		}
		seed = stepShiftRegister(seed, bannerMask)
		if seed == 1 {
			break
		}
	}
	return false
}
