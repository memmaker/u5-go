//go:build js

package music

// The browser build stays small: the page puts the SoundFont in a file named by $GOPLAY_SF2.
func embeddedSF2() []byte { return nil }
