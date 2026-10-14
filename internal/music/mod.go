package music

import (
	"encoding/binary"
	"errors"
	"math"
)

// A Protracker module ("M.K.", four channels, 31 samples): a 20-byte title, 31 sample
// headers of 30 bytes (name, length in words, finetune, volume, loop start and length in
// words; big-endian), the song length, a restart byte, 128 pattern numbers, the tag, then the
// patterns (64 rows of 4 channels of 4 bytes) and the 8-bit signed samples.

const (
	modSamples      = 31
	modChannels     = 4
	modRows         = 64
	modCellBytes    = 4
	modPatternBytes = modRows * modChannels * modCellBytes
	modHeaderBytes  = 1084
	modOrderAt      = 952
	modLengthAt     = 950
	modSampleAt     = 20
	modSampleBytes  = 30
	// paulaClock is the Amiga's (PAL) sample clock: a period p plays paulaClock/p samples a
	// second.
	paulaClock    = 3546895
	defaultSpeed  = 6
	defaultTempo  = 125
	maxVolume     = 64
	stereoSpread  = 0.75 // how far the Amiga's hard-panned channels are pulled apart
	modOutputGain = 0.5
)

// Effects of the cells.
const (
	fxArpeggio       = 0x0
	fxSlideUp        = 0x1
	fxSlideDown      = 0x2
	fxTonePortamento = 0x3
	fxVibrato        = 0x4
	fxPortaVolSlide  = 0x5
	fxVibVolSlide    = 0x6
	fxSampleOffset   = 0x9
	fxVolumeSlide    = 0xA
	fxPositionJump   = 0xB
	fxSetVolume      = 0xC
	fxPatternBreak   = 0xD
	fxExtended       = 0xE
	fxSetSpeed       = 0xF
)

// Extended effects (E x).
const (
	exFineSlideUp   = 0x1
	exFineSlideDown = 0x2
	exPatternLoop   = 0x6
	exRetrigger     = 0x9
	exFineVolUp     = 0xA
	exFineVolDown   = 0xB
	exNoteCut       = 0xC
	exNoteDelay     = 0xD
	exPatternDelay  = 0xE
)

type modSample struct {
	data       []int8
	volume     int
	finetune   int
	loopStart  int
	loopLength int
}

type modCell struct {
	sample int // 1-based, 0 none
	period int
	effect int
	param  int
}

type modChannel struct {
	sample     *modSample
	pos        float64 // in samples
	period     int
	volume     int
	portaTo    int
	portaSpeed int
	vibPos     int
	vibSpeed   int
	vibDepth   int
	loopRow    int
	loopCount  int
	basePeriod int
}

// modPlayer plays a module, looping it.
type modPlayer struct {
	samples  [modSamples]modSample
	order    []int
	patterns [][]modCell
	ch       [modChannels]modChannel

	speed, tempo int
	position     int
	row          int
	tick         int
	breakRow     int
	jump         int
	delayRows    int
	untilTick    float64 // samples left in this tick
	stopped      bool
}

var errNotModule = errors.New("music: not a Protracker module")

func newModPlayer(b []byte) (*modPlayer, error) {
	if len(b) < modHeaderBytes || string(b[1080:1084]) != "M.K." {
		return nil, errNotModule
	}
	p := &modPlayer{}
	length := int(b[modLengthAt])
	patterns := 0
	for i := range 128 {
		if n := int(b[modOrderAt+i]); n+1 > patterns {
			patterns = n + 1
		}
		if i < length {
			p.order = append(p.order, int(b[modOrderAt+i]))
		}
	}
	at := modHeaderBytes
	for i := range patterns {
		if at+modPatternBytes > len(b) {
			return nil, errNotModule
		}
		cells := make([]modCell, modRows*modChannels)
		for c := range cells {
			raw := b[at+c*modCellBytes:]
			cells[c] = modCell{
				sample: int(raw[0]&0xF0) | int(raw[2]>>4),
				period: int(raw[0]&0x0F)<<8 | int(raw[1]),
				effect: int(raw[2] & 0x0F),
				param:  int(raw[3]),
			}
		}
		p.patterns = append(p.patterns, cells)
		at += modPatternBytes
		_ = i
	}
	for i := range modSamples {
		h := b[modSampleAt+i*modSampleBytes:]
		length := int(binary.BigEndian.Uint16(h[22:])) * 2
		s := &p.samples[i]
		s.finetune = int(h[24] & 0x0F)
		if s.finetune > 7 {
			s.finetune -= 16
		}
		s.volume = min(int(h[25]), maxVolume)
		s.loopStart = int(binary.BigEndian.Uint16(h[26:])) * 2
		s.loopLength = int(binary.BigEndian.Uint16(h[28:])) * 2
		end := min(at+length, len(b))
		for _, v := range b[at:end] {
			s.data = append(s.data, int8(v))
		}
		at += length
		if s.loopStart+s.loopLength > len(s.data) {
			s.loopLength = max(len(s.data)-s.loopStart, 0)
		}
	}
	if len(p.order) == 0 {
		return nil, errNotModule
	}
	return p, nil
}

func (p *modPlayer) start(int) {
	if p.speed != 0 && !p.stopped {
		return // the Amiga plays its one tune on through every song change
	}
	p.speed, p.tempo = defaultSpeed, defaultTempo
	p.position, p.row, p.tick, p.breakRow, p.jump = 0, 0, 0, -1, -1
	p.ch = [modChannels]modChannel{}
	p.untilTick = 0
	p.stopped = false
}

func (p *modPlayer) samplesPerTick() float64 { return rate * 2.5 / float64(p.tempo) }

func (p *modPlayer) render(l, r []float32) {
	for i := range l {
		if p.untilTick <= 0 {
			p.step()
			p.untilTick += p.samplesPerTick()
		}
		p.untilTick--
		var left, right float32
		for c := range p.ch {
			v := p.ch[c].next()
			// Paula: channels 0 and 3 left, 1 and 2 right
			if c == 0 || c == 3 {
				left += v * (0.5 + stereoSpread/2)
				right += v * (0.5 - stereoSpread/2)
			} else {
				right += v * (0.5 + stereoSpread/2)
				left += v * (0.5 - stereoSpread/2)
			}
		}
		l[i] += left * modOutputGain
		r[i] += right * modOutputGain
	}
}

// next is the channel's next output sample.
func (c *modChannel) next() float32 {
	s := c.sample
	if s == nil || c.period == 0 || len(s.data) == 0 {
		return 0
	}
	at := int(c.pos)
	if at >= len(s.data) {
		if s.loopLength > 2 {
			c.pos = float64(s.loopStart) + math.Mod(c.pos-float64(s.loopStart), float64(s.loopLength))
			at = int(c.pos)
		} else {
			return 0
		}
	}
	v := float32(s.data[at]) / 128 * float32(c.volume) / maxVolume
	c.pos += paulaClock / float64(c.period) / rate
	if s.loopLength > 2 && c.pos >= float64(s.loopStart+s.loopLength) {
		c.pos -= float64(s.loopLength)
	}
	return v
}

// step runs one tick: a new row on tick 0, the running effects on the others.
func (p *modPlayer) step() {
	if p.tick == 0 {
		if p.delayRows > 0 {
			p.delayRows--
		} else {
			p.playRow()
		}
	} else {
		for c := range p.ch {
			p.tickEffect(c, p.cell(c))
		}
	}
	p.tick++
	if p.tick >= p.speed {
		p.tick = 0
		if p.delayRows > 0 {
			return
		}
		p.advance()
	}
}

func (p *modPlayer) cell(c int) modCell {
	pat := p.patterns[p.order[p.position]%len(p.patterns)]
	return pat[p.row*modChannels+c]
}

func (p *modPlayer) advance() {
	switch {
	case p.jump >= 0:
		p.position = p.jump
		p.row = max(p.breakRow, 0)
	case p.breakRow >= 0:
		p.position++
		p.row = p.breakRow
	default:
		p.row++
		if p.row >= modRows {
			p.row = 0
			p.position++
		}
	}
	p.jump, p.breakRow = -1, -1
	if p.position >= len(p.order) {
		p.position = 0 // loop the tune
	}
	p.row = min(p.row, modRows-1)
}

func (p *modPlayer) playRow() {
	for c := range p.ch {
		ch := &p.ch[c]
		cell := p.cell(c)
		if cell.sample > 0 && cell.sample <= modSamples {
			ch.sample = &p.samples[cell.sample-1]
			ch.volume = ch.sample.volume
		}
		if cell.period != 0 {
			if cell.effect == fxTonePortamento || cell.effect == fxPortaVolSlide {
				ch.portaTo = cell.period
			} else if !(cell.effect == fxExtended && cell.param>>4 == exNoteDelay) {
				ch.period, ch.basePeriod = cell.period, cell.period
				ch.pos = 0
				ch.vibPos = 0
			}
		}
		x, y := cell.param>>4, cell.param&0x0F
		switch cell.effect {
		case fxTonePortamento:
			if cell.param != 0 {
				ch.portaSpeed = cell.param
			}
		case fxVibrato:
			if x != 0 {
				ch.vibSpeed = x
			}
			if y != 0 {
				ch.vibDepth = y
			}
		case fxSampleOffset:
			if cell.period != 0 {
				ch.pos = float64(cell.param * 256)
			}
		case fxPositionJump:
			p.jump = cell.param
		case fxSetVolume:
			ch.volume = min(cell.param, maxVolume)
		case fxPatternBreak:
			p.breakRow = x*10 + y
		case fxSetSpeed:
			if cell.param == 0 {
				break
			}
			if cell.param < 32 {
				p.speed = cell.param
			} else {
				p.tempo = cell.param
			}
		case fxExtended:
			switch x {
			case exFineSlideUp:
				ch.period = max(ch.period-y, 113)
			case exFineSlideDown:
				ch.period = min(ch.period+y, 856)
			case exFineVolUp:
				ch.volume = min(ch.volume+y, maxVolume)
			case exFineVolDown:
				ch.volume = max(ch.volume-y, 0)
			case exPatternLoop:
				if y == 0 {
					ch.loopRow = p.row
				} else if ch.loopCount == 0 {
					ch.loopCount = y
					p.breakRow, p.jump = ch.loopRow, p.position
				} else if ch.loopCount--; ch.loopCount > 0 {
					p.breakRow, p.jump = ch.loopRow, p.position
				}
			case exPatternDelay:
				p.delayRows = y
			}
		}
	}
}

var vibratoTable = func() [64]int {
	var t [64]int
	for i := range t {
		t[i] = int(math.Round(255 * math.Sin(float64(i)*math.Pi/32)))
	}
	return t
}()

func (p *modPlayer) tickEffect(c int, cell modCell) {
	ch := &p.ch[c]
	x, y := cell.param>>4, cell.param&0x0F
	volumeSlide := func() {
		if x != 0 {
			ch.volume = min(ch.volume+x, maxVolume)
		} else {
			ch.volume = max(ch.volume-y, 0)
		}
	}
	porta := func() {
		if ch.portaTo == 0 {
			return
		}
		if ch.period < ch.portaTo {
			ch.period = min(ch.period+ch.portaSpeed, ch.portaTo)
		} else {
			ch.period = max(ch.period-ch.portaSpeed, ch.portaTo)
		}
		ch.basePeriod = ch.period
	}
	vibrato := func() {
		ch.period = ch.basePeriod + vibratoTable[ch.vibPos&63]*ch.vibDepth/128
		ch.vibPos += ch.vibSpeed
	}
	switch cell.effect {
	case fxArpeggio:
		if cell.param != 0 {
			semis := [3]int{0, x, y}[p.tick%3]
			ch.period = int(float64(ch.basePeriod) / math.Pow(2, float64(semis)/12))
		}
	case fxSlideUp:
		ch.period = max(ch.period-cell.param, 113)
		ch.basePeriod = ch.period
	case fxSlideDown:
		ch.period = min(ch.period+cell.param, 856)
		ch.basePeriod = ch.period
	case fxTonePortamento:
		porta()
	case fxVibrato:
		vibrato()
	case fxPortaVolSlide:
		porta()
		volumeSlide()
	case fxVibVolSlide:
		vibrato()
		volumeSlide()
	case fxVolumeSlide:
		volumeSlide()
	case fxExtended:
		switch x {
		case exRetrigger:
			if y != 0 && p.tick%y == 0 {
				ch.pos = 0
			}
		case exNoteCut:
			if p.tick == y {
				ch.volume = 0
			}
		case exNoteDelay:
			if p.tick == y && cell.period != 0 {
				ch.period, ch.basePeriod = cell.period, cell.period
				ch.pos = 0
			}
		}
	}
}
