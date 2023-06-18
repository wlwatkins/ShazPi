package display

import "shazammini/src/utilities"

type Assets struct {
	WifiOn  EPDPNG
	WifiOff EPDPNG
}

func (a *Assets) LoadAssets() {

	a.WifiOff.LoadPNG("wifi_unconnected.png", 0.04, utilities.Coordonates{X: 240, Y: 10})
	a.WifiOn.LoadPNG("wifi_connected.png", 0.04, utilities.Coordonates{X: 240, Y: 10})

}
