package structs

import (
	"time"
)

type CommChannels struct {
	PlayChannel     chan bool
	RecordChannel   chan time.Duration
	FetchAPI        chan bool
	DisplayResult   chan Track
	DisplayThinking chan bool
	DisplayRecord   chan bool
}

type Match struct {
	ID            string  `json:"id"`
	Offset        float64 `json:"offset"`
	TimeSkew      float64 `json:"timeskew"`
	FrequencySkew float64 `json:"frequencyskew"`
}

type Track struct {
	Layout   string      `json:"layout"`
	Type     string      `json:"type"`
	Key      string      `json:"key"`
	Title    string      `json:"title"`
	Subtitle string      `json:"subtitle"`
	Image    TrackImages `json:"image"`
	Share    Share       `json:"share"`
	Hub      Hub         `json:"hub"`
	Url      string      `json:"url"`
	Artists  []Artist    `json:"artist"`
	Isrc     string      `json:"isrc"`
	Genre    Genre       `json:"genre"`
	// Urlparams   Urlparams   `json:"urlparams"`
	MyShazam    MyShazam `json:"myshazam"`
	Albumadamid string   `json:"albumadamid"`
	// Sections    []Sections `json:"sections"`
}

type TrackImages struct {
	Background string `json:"background"`
	CoverArt   string `json:"coverart"`
	CoverArtHQ string `json:"coverarthq"`
	JoeColor   string `json:"joecolor"`
}

type Share struct {
	Subject  string `json:"subject"`
	Text     string `json:"text"`
	Href     string `json:"href"`
	Image    string `json:"image"`
	Twitter  string `json:"twitter"`
	HTML     string `json:"html"`
	Avatar   string `json:"avatar"`
	Snapchat string `json:"snapchat"`
}

type Hub struct {
	Type        string     `json:"type"`
	Image       string     `json:"image"`
	Actions     []Actions  `json:"actions"`
	Options     []Options  `json:"options"`
	Providers   []Provider `json:"providers"`
	Explicit    bool       `json:"explicit"`
	DisplayName string     `json:"displayname"`
}

type Actions struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Id   string `json:"id"`
	Uri  string `json:"uri"`
}

type Options struct {
	Caption             string     `json:"caption"`
	Actions             []Actions  `json:"actions"`
	Beacondata          Beacondata `json:"beacondata"`
	Image               string     `json:"image"`
	Type                string     `json:"type"`
	Listcaption         string     `json:"listcaption"`
	Overflowimage       string     `json:"overflowimage"`
	Colouroverflowimage bool       `json:"colouroverflowimage"`
	Providername        string     `json:"providername"`
}

type Beacondata struct {
	Type         string `json:"type"`
	Providername string `json:"providername"`
}

type Provider struct {
	Caption string `json:"caption"`
	Images  struct {
		Overflow string `json:"overflow"`
		Default  string `json:"default"`
	} `json:"images"`
	Actions []Actions `json:"actions"`
	Type    string    `json:"type"`
}

type Artist struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
	URL      string `json:"url"`
}

type Genre struct {
	Primary string `json:"primary"`
}

type MyShazam struct {
	Apple struct {
		Actions []struct {
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"actions"`
	} `json:"apple"`
}
