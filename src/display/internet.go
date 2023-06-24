package display

import "net/http"

func Connected() (ok bool) {
	_, err := http.Get("http://clients3.google.com/generate_204")
	return err == nil
}
