package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/memmaker/u5-go/internal/game"
	"github.com/memmaker/u5-go/internal/host"
	"github.com/memmaker/u5-go/internal/music"
	"github.com/memmaker/u5-go/internal/speaker"
)

// musicVolumeSteps are the volume settings, 0 to 10.
const musicVolumeSteps = 11

// setUpSettings registers the settings menu (F2) and the Alt shortcuts.
func setUpSettings() {
	host.Settings = []host.Setting{
		{
			Name: "Music",
			Values: func() []string {
				var names []string
				for _, s := range music.Available() {
					names = append(names, s.String())
				}
				return names
			},
			Get: func() int {
				for i, s := range music.Available() {
					if s == music.CurrentSource() {
						return i
					}
				}
				return 0
			},
			Set: func(i int) {
				if a := music.Available(); i < len(a) {
					music.SetSource(a[i])
				}
			},
		},
		{
			Name: "Music volume",
			Values: func() []string {
				v := make([]string, musicVolumeSteps)
				for i := range v {
					v[i] = fmt.Sprint(i)
				}
				return v
			},
			Get: music.Volume,
			Set: music.SetVolume,
		},
		{
			Name:   "Sound effects",
			Values: func() []string { return []string{"Off", "PC speaker"} },
			Get: func() int {
				if speaker.Enabled {
					return 1
				}
				return 0
			},
			Set: func(i int) { speaker.Enabled = i == 1 },
		},
	}
	host.Settings = append(host.Settings, host.Setting{
		Name:   "Layout",
		Values: func() []string { return []string{"Original", "Modern (menus, wide map)"} },
		Get: func() int {
			if host.Modern {
				return 1
			}
			return 0
		},
		Set: func(i int) { host.Modern = i == 1 },
	})
	host.Settings = append(host.Settings, host.Setting{
		Name:   "Steps",
		Values: func() []string { return []string{"Quiet", "Echoed (original)"} },
		Get: func() int {
			if game.EchoSteps {
				return 1
			}
			return 0
		},
		Set: func(i int) { game.EchoSteps = i == 1 },
	})
	if speaker.HarpLoaded() {
		host.Settings = append(host.Settings, host.Setting{
			Name:   "Instruments",
			Values: func() []string { return []string{"PC speaker", "Amiga harp"} },
			Get: func() int {
				if speaker.HarpEnabled {
					return 1
				}
				return 0
			},
			Set: func(i int) { speaker.HarpEnabled = i == 1 },
		})
	}
	host.AltKeys[ebiten.KeyV] = music.Toggle
	host.AltKeys[ebiten.KeyEqual] = music.Louder
	host.AltKeys[ebiten.KeyNumpadAdd] = music.Louder
	host.AltKeys[ebiten.KeyMinus] = music.Softer
	host.AltKeys[ebiten.KeyNumpadSubtract] = music.Softer
}
