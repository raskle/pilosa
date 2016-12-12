package pilosaic

// pilosaic (pilosa+mosaic) is a tool for creating database visualization images.
// It handles downsampling from database dimensions to image dimensions,
// using either mean() or max(). It can also write an output .png image.
// Usage example:
//
// drawer = pilosaic.NewImageDrawer(100, 100, 10000, 10000)
// for n := 0; n < iterations; n++ {
//   ...
//   drawer.SetMeanBit(bitmapID, profileID)
// }
// drawer.WriteToFile("mosaic.png")
import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

type ImageDrawer struct {
	pixelWidth  int
	pixelHeight int
	pointWidth  int
	pointHeight int
	buffer      [][]int
}

func NewImageDrawer(pixelWidth, pixelHeight, dbWidth, dbHeight int) *ImageDrawer {
	buffer := make([][]int, 0, pixelWidth)
	for x := 0; x < pixelWidth; x++ {
		col := make([]int, pixelHeight)
		buffer = append(buffer, col)
	}
	return &ImageDrawer{pixelWidth, pixelHeight, dbWidth, dbHeight, buffer}
}

func (d *ImageDrawer) pointToPixel(x, y int) (int, int) {
	px := int(float64(d.pixelWidth) / float64(d.pointWidth) * float64(x))
	py := int(float64(d.pixelHeight) / float64(d.pointHeight) * float64(y))
	return px, py
}

func (d *ImageDrawer) SetMeanBit(bitmapID, profileID int) {
	x, y := d.pointToPixel(bitmapID, profileID)
	d.buffer[x][y] += 1
}

func (d *ImageDrawer) SetMaxBit(bitmapID, profileID int) {
	x, y := d.pointToPixel(bitmapID, profileID)
	d.buffer[x][y] = 1
}

func (d *ImageDrawer) getBufferMax() int {
	fmt.Println("getBufferMax")
	max := 0
	for x := 0; x < d.pixelWidth; x++ {
		for y := 0; y < d.pixelWidth; y++ {
			if d.buffer[x][y] > max {
				max = d.buffer[x][y]
			}
		}
	}
	return max
}

func (d *ImageDrawer) scaleBuffer() {
	fmt.Println("scaleBuffer")
	max := float64(d.getBufferMax())
	for x := 0; x < d.pixelWidth; x++ {
		for y := 0; y < d.pixelWidth; y++ {
			d.buffer[x][y] = int(float64(d.buffer[x][y]) / max * 255)
		}
	}
}

func (d *ImageDrawer) WriteToFile(filename string) {
	d.scaleBuffer()
	imgRect := image.Rect(0, 0, d.pixelWidth, d.pixelHeight)
	img := image.NewGray(imgRect)

	for x := 0; x < d.pixelWidth; x++ {
		for y := 0; y < d.pixelWidth; y++ {
			img.Set(x, y, color.Gray{uint8(d.buffer[x][y])})
		}
	}

	out, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = png.Encode(out, img)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
