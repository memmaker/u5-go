// Package music plays Ultima V's music from one of two optional sources the player supplies:
//
//   - the MIDI score of the Exodus Project's "Ultima V Upgrade" (Voyager Dragon): sixteen XMI
//     files, the Apple II / Commodore 128 arrangement, synthesized with meltysynth from a
//     SoundFont (GeneralUser GS, embedded in desktop builds when tools/fetch-soundfont.sh
//     fetched it; on js/wasm read from the file $GOPLAY_SF2 names);
//   - the Amiga version's Protracker module, the one tune it plays throughout the game.
//
// The original PC version has no music. Both play into one Ebiten audio stream, mixed by
// Ebiten with the PC speaker effects.
package music

//go:generate sh ../../tools/fetch-soundfont.sh

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/memmaker/u5-go/internal/meltysynth"
)

// The songs, numbered as the upgrade's MIDI driver numbers them (its MIDI_TBL).
const (
	UltimaTheme = iota
	BritannicLands
	Hornpipe
	EngagementAndMelee
	Stones
	GreysonsTale
	Fanfare
	MissingMonarch
	VillagerTarantella
	HallsOfDoom
	WorldsBelow
	LordBlackthorn
	DreamOfLadyNan
	JoyousReunion
	RuleBritannia
	AmigaTheme
	songCount
	// Stop is silence.
	Stop = -1
)

// SongFiles are the upgrade's file names of the songs.
var SongFiles = [songCount]string{
	"U5THEME.XMI", "BRITLAND.XMI", "HORNPIPE.XMI", "ENGGMNT.XMI", "STONES.XMI", "GREYSON.XMI",
	"FANFARE.XMI", "MONARCH.XMI", "TRNTLLA.XMI", "HALLS.XMI", "WRLDBLW.XMI", "BLCKTHRN.XMI",
	"LADYNAN.XMI", "REUNION.XMI", "RULEBRIT.XMI", "AMIGA.XMI",
}

// AmigaModule is where the Amiga extras keep the module, under the game directory.
const AmigaModule = "DISK2/U5MUSIC.MOD"

// Source is where the music comes from.
type Source int

const (
	Off Source = iota
	MIDI
	Amiga
)

func (s Source) String() string {
	switch s {
	case MIDI:
		return "MIDI (Exodus upgrade)"
	case Amiga:
		return "Amiga"
	}
	return "Off"
}

const rate = 44100

// renderer is a music source playing into stereo float buffers.
type renderer interface {
	start(song int)
	render(l, r []float32)
}

var (
	mu      sync.Mutex
	midi    *midiPlayer // nil without the XMI files or a SoundFont
	amiga   *modPlayer  // nil without the Amiga module
	source  = Off
	song    = Stop // what the game asked for
	next    = Stop // what follows song once it ended, Stop to loop song
	playing = Stop // what the current source plays
	current renderer
	muted   bool
	volume  = 8 // 0..10
	started bool
	ready   = make(chan struct{})
)

// Start loads what the game directory dir has (the XMI files, the Amiga module) and the
// SoundFont, in the background, and opens the stream, once. preferred is the source to start
// with when it is there; otherwise the first one there is. U5GO_SILENT keeps tests silent.
func Start(dir string, preferred Source) {
	if started || os.Getenv("U5GO_SILENT") != "" {
		return
	}
	started = true
	go func() {
		defer close(ready)
		if m, err := os.ReadFile(filepath.Join(dir, AmigaModule)); err == nil {
			if p, err := newModPlayer(m); err == nil {
				mu.Lock()
				amiga = p
				mu.Unlock()
			}
		}
		if p := loadMIDI(dir); p != nil {
			mu.Lock()
			midi = p
			mu.Unlock()
		}
		mu.Lock()
		switch {
		case preferred == MIDI && midi != nil, preferred == Amiga && amiga != nil, preferred == Off:
			source = preferred
		case midi != nil:
			source = MIDI
		case amiga != nil:
			source = Amiga
		}
		mu.Unlock()
		ctx := audio.CurrentContext()
		if ctx == nil {
			ctx = audio.NewContext(rate)
		}
		p, err := ctx.NewPlayerF32(stream{})
		if err != nil {
			return
		}
		p.SetBufferSize(100 * time.Millisecond)
		p.Play()
	}()
}

// loadMIDI reads the songs from dir and the SoundFont; nil when either is missing.
func loadMIDI(dir string) *midiPlayer {
	var songs [songCount][]event
	found := 0
	for i, name := range SongFiles {
		b, err := readAnyCase(dir, name)
		if err != nil {
			continue
		}
		seqs, err := parseXMI(b)
		if err != nil || len(seqs) == 0 {
			continue
		}
		songs[i] = seqs[0]
		found++
	}
	if found == 0 {
		return nil
	}
	sf2 := embeddedSF2()
	if p := os.Getenv("GOPLAY_SF2"); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			sf2 = b
		}
	}
	if sf2 == nil {
		return nil
	}
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sf2))
	if err != nil {
		return nil
	}
	syn, err := meltysynth.NewSynthesizer(sf, meltysynth.NewSynthesizerSettings(rate))
	if err != nil {
		return nil
	}
	return &midiPlayer{synth: syn, songs: songs}
}

// readAnyCase reads dir/name in the case it is stored in.
func readAnyCase(dir, name string) ([]byte, error) {
	if b, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
		return b, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if equalFold(e.Name(), name) {
			return os.ReadFile(filepath.Join(dir, e.Name()))
		}
	}
	return nil, os.ErrNotExist
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		x, y := a[i], b[i]
		if 'a' <= x && x <= 'z' {
			x -= 'a' - 'A'
		}
		if 'a' <= y && y <= 'z' {
			y -= 'a' - 'A'
		}
		if x != y {
			return false
		}
	}
	return true
}

// Available lists the sources there are (Off always), once loading finished.
func Available() []Source {
	<-readyOrNow()
	mu.Lock()
	defer mu.Unlock()
	s := []Source{Off}
	if midi != nil {
		s = append(s, MIDI)
	}
	if amiga != nil {
		s = append(s, Amiga)
	}
	return s
}

// readyOrNow is the loading's end, or an already closed channel when music never started.
func readyOrNow() <-chan struct{} {
	if !started {
		c := make(chan struct{})
		close(c)
		return c
	}
	return ready
}

// SetSource switches the music to s, if it is there.
func SetSource(s Source) {
	mu.Lock()
	defer mu.Unlock()
	if (s == MIDI && midi == nil) || (s == Amiga && amiga == nil) {
		return
	}
	source = s
	playing = Stop
	current = nil
}

// CurrentSource is the source playing.
func CurrentSource() Source {
	mu.Lock()
	defer mu.Unlock()
	return source
}

// Play asks for song (Stop for none); asking for the song already playing changes nothing.
// The Amiga source plays its one tune for every song.
func Play(n int) {
	mu.Lock()
	if n != song {
		next = Stop
	}
	song = n
	mu.Unlock()
}

// PlayThen plays first once, then then on and on.
func PlayThen(first, then int) {
	mu.Lock()
	song, next = first, then
	mu.Unlock()
}

// Current is the song last asked for.
func Current() int {
	mu.Lock()
	defer mu.Unlock()
	return song
}

// Toggle mutes or unmutes the music (Alt-V).
func Toggle() { mu.Lock(); muted = !muted; mu.Unlock() }

// Muted says whether the music is muted.
func Muted() bool { mu.Lock(); defer mu.Unlock(); return muted }

// Louder and Softer step the volume (Alt-'+' and Alt-'-').
func Louder() { mu.Lock(); volume = min(volume+1, 10); mu.Unlock() }
func Softer() { mu.Lock(); volume = max(volume-1, 0); mu.Unlock() }

// Volume is the music's volume, 0..10; SetVolume sets it.
func Volume() int     { mu.Lock(); defer mu.Unlock(); return volume }
func SetVolume(v int) { mu.Lock(); volume = min(max(v, 0), 10); mu.Unlock() }

type stream struct{}

var buf [2][]float32

func (stream) Read(b []byte) (int, error) {
	frames := len(b) / 8
	mu.Lock()
	defer mu.Unlock()
	if cap(buf[0]) < frames {
		buf[0], buf[1] = make([]float32, frames), make([]float32, frames)
	}
	l, r := buf[0][:frames], buf[1][:frames]
	clear(l)
	clear(r)
	want := current
	switch source {
	case MIDI:
		want = midi
	case Amiga:
		want = amiga
	default:
		want = nil
	}
	if want != current {
		current, playing = want, Stop
	}
	if current != nil {
		if song != playing {
			current.start(song)
			playing = song
		}
		current.render(l, r)
		if m, ok := current.(*midiPlayer); ok && m.ended && next != Stop {
			song, next = next, Stop
		}
	}
	g := float32(volume) / 10 * float32(fader("music"))
	if muted {
		g = 0
	}
	for i := range frames {
		putF32(b[i*8:], l[i]*g)
		putF32(b[i*8+4:], r[i]*g)
	}
	return frames * 8, nil
}

func putF32(b []byte, v float32) { binary.LittleEndian.PutUint32(b, math.Float32bits(v)) }

// midiPlayer renders one XMI song at a time through a synthesizer, looping it.
type midiPlayer struct {
	synth  *meltysynth.Synthesizer
	songs  [songCount][]event
	events []event
	pos    int  // next event
	at     int  // samples since the song started
	ended  bool // the song came to its end at least once
}

func (s *midiPlayer) start(n int) {
	s.synth.NoteOffAll(false)
	s.synth.ResetAllControllers()
	s.events, s.pos, s.at, s.ended = nil, 0, 0, false
	if n >= 0 && n < songCount {
		s.events = s.songs[n]
	}
}

func (s *midiPlayer) render(l, r []float32) {
	for len(l) > 0 {
		n := len(l)
		if len(s.events) > 0 {
			for s.pos < len(s.events) && s.events[s.pos].tick*rate/tickHz <= s.at {
				e := s.events[s.pos]
				if e.status != 0 {
					s.synth.ProcessMidiMessage(int32(midiChannel(e.status)), int32(midiMessage(e.status)<<4), int32(e.d1), int32(e.d2))
				}
				s.pos++
			}
			if s.pos == len(s.events) { // loop
				s.pos, s.at, s.ended = 0, 0, true
				continue
			}
			n = min(n, s.events[s.pos].tick*rate/tickHz-s.at)
		}
		s.synth.Render(l[:n], r[:n])
		s.at += n
		l, r = l[n:], r[n:]
	}
}
