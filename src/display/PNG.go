package display

import (
	"image"
	"shazammini/src/utilities"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

type EPDPNG struct {
	path  string
	png   image.Image
	coord utilities.Coordonates
}

func (e *EPDPNG) LoadPNG(path string, scale float64, coord utilities.Coordonates) {
	e.path = path
	png_raw, err := gg.LoadPNG(path)
	if err != nil {
		panic(err)
	}

	e.png = resize.Resize(uint(float64(png_raw.Bounds().Dx())*scale), 0, png_raw, resize.Lanczos2)
	e.coord = coord
}
