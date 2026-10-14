package game

import "github.com/memmaker/u5-go/internal/ega"

// Segment 7000's flickering firelight: the light of up to two fires in a settlement or
// the world shimmers by switching shapes of tiles lit and unlit in the light map.

// The fires whose light flickers (D_217e..D_2184); x -1 marks none.
var (
	fireX, fireY             int16 = -1, -1 // C: D_217e, D_2180
	secondFireX, secondFireY int16 = -1, -1 // C: D_2182, D_2184
	// flickerStep is the flicker shape shown, 255 before the first (D_2186).
	flickerStep byte = flickerStart // C: D_2186
)

const (
	flickerStart = 255
	noFire       = -1
	flickerLit   = 255
)

// markFlickerShape sets the points of flicker shape (lit or dark) around (x, y) in the
// light map.
// C: ULTIMA_7040
func markFlickerShape(shape int, lit bool, x, y int) {
	value := byte(0)
	if lit {
		value = flickerLit
	}
	shape %= flickerShapeCount
	for point := range flickerShapePoints {
		at := (shape*flickerShapePoints + point) * flickerPairBytes
		px := int(int8(flickerShapes[at]))
		py := int(int8(flickerShapes[at+1]))
		index := (py+y)*viewStride + px + x
		// fix: the original writes outside the light map for fires at its edge.
		if index >= 0 && index < len(combatMap) {
			combatMap[index] = value
		}
	}
}

// flickerLights moves the fires' light on a step, while it is dark.
// C: ULTIMA_70a6
func flickerLights() {
	if light >= dayLight || fireX == noFire {
		flickerStep = flickerStart
		return
	}
	viewChanged = 1
	if flickerStep == flickerStart {
		flickerStep++
		for shape := range 3 {
			markFlickerShape(shape, true, int(fireX), int(fireY))
		}
		if secondFireX != noFire {
			for shape := range 3 {
				markFlickerShape(shape, true, int(secondFireX), int(secondFireY))
			}
		}
	} else {
		markFlickerShape(int(flickerStep), false, int(fireX), int(fireY))
		if secondFireX != noFire {
			markFlickerShape(int(flickerStep), false, int(secondFireX), int(secondFireY))
		}
		flickerStep++
		markFlickerShape(int(flickerStep)+2, true, int(fireX), int(fireY))
		if secondFireX != noFire {
			markFlickerShape(int(flickerStep)+2, true, int(secondFireX), int(secondFireY))
		}
	}
	if flickerStep >= flickerShapeCount {
		flickerStep = 0
	}
}

// updateTimeTiles redraws the clock faces for the time.
// C: ULTIMA_71aa
func updateTimeTiles() { screen.UpdateTimeTiles(ega.TimeClock, int(hour), int(minute)) }

// changeTimeTiles has the driver change the tiles of mode (a ladder broken or mended,
// night, the green glow).
// C: ULTIMA_6fbc
func changeTimeTiles(mode int) { screen.UpdateTimeTiles(mode, 0, 0) }
