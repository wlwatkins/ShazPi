package display

type Assets struct {
	WifiOn  EPDPNG
	WifiOff EPDPNG
}

func (a *Assets) LoadAssets(d *Display) {

	a.WifiOff.LoadPNG("static/wifi_unconnected.png", 0.04, Coordonates{X: d.width - 10, Y: 10})
	a.WifiOn.LoadPNG("static/wifi_connected.png", 0.04, Coordonates{X: d.width - 10, Y: 10})

}
