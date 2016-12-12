package pilosaic

import (
	"math/rand"
	"testing"
)

func TestImageDrawer(t *testing.T) {
	// basic functionality test
	img1 := NewImageDrawer(100, 100, 100, 100)
	img1.SetMaxBit(0, 0)
	img1.WriteToFile("test-1.png")

	// scaling test
	img2 := NewImageDrawer(100, 100, 1000, 1000)
	for n := 0; n < 100000; n++ {
		img2.SetMeanBit(rand.Intn(1000), rand.Intn(1000))
	}
	img2.WriteToFile("test-2.png")

}
