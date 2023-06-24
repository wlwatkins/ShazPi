package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"shazammini/src/structs"
	"strings"

	"gobot.io/x/gobot"
)

type Match struct {
	ID            string  `json:"id"`
	Offset        float64 `json:"offset"`
	TimeSkew      float64 `json:"timeskew"`
	FrequencySkew float64 `json:"frequencyskew"`
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

type Provider struct {
	Caption string `json:"caption"`
	Images  struct {
		Overflow string `json:"overflow"`
		Default  string `json:"default"`
	} `json:"images"`
	Actions []struct {
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"actions"`
	Type               string `json:"type"`
	ListCaption        string `json:"listcaption"`
	OverflowImage      string `json:"overflowimage"`
	ColorOverflowImage bool   `json:"colouroverflowimage"`
	ProviderName       string `json:"providername"`
}

type Genre struct {
	Primary string `json:"primary"`
}

type Metadata struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type Section struct {
	Type      string     `json:"type"`
	MetaPages []MetaPage `json:"metapages"`
	TabName   string     `json:"tabname"`
	Metadata  []Metadata `json:"metadata"`
}

type MetaPage struct {
	Image   string `json:"image"`
	Caption string `json:"caption"`
}

type YoutubeURL struct {
	Caption string `json:"caption"`
	Image   struct {
		Dimensions struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"dimensions"`
		URL string `json:"url"`
	} `json:"image"`
	Actions []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Share struct {
			Subject  string `json:"subject"`
			Text     string `json:"text"`
			Href     string `json:"href"`
			Image    string `json:"image"`
			Twitter  string `json:"twitter"`
			HTML     string `json:"html"`
			Avatar   string `json:"avatar"`
			Snapchat string `json:"snapchat"`
		} `json:"share"`
		URI string `json:"uri"`
	} `json:"actions"`
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

type Hub struct {
	Type        string     `json:"type"`
	Image       string     `json:"image"`
	Actions     []Provider `json:"actions"`
	Options     []Provider `json:"options"`
	Providers   []Provider `json:"providers"`
	Explicit    bool       `json:"explicit"`
	DisplayName string     `json:"displayname"`
}

type Track struct {
	ID            string      `json:"id"`
	Offset        float64     `json:"offset"`
	TimeSkew      float64     `json:"timeskew"`
	FrequencySkew float64     `json:"frequencyskew"`
	Images        TrackImages `json:"images"`
	Share         Share       `json:"share"`
	Provider      Provider    `json:"provider"`
	Genre         Genre       `json:"genre"`
	Sections      []Section   `json:"sections"`
	YoutubeURL    YoutubeURL  `json:"youtubeurl"`
	MyShazam      MyShazam    `json:"myshazam"`
}

type Result struct {
	Match      Match      `json:"match"`
	Track      Track      `json:"track"`
	Artists    []Artist   `json:"artists"`
	Sections   []Section  `json:"sections"`
	YoutubeURL YoutubeURL `json:"youtubeurl"`
	MyShazam   MyShazam   `json:"myshazam"`
	Hub        Hub        `json:"hub"`
}

type Artist struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
	URL      string `json:"url"`
}

type ShazamResponse struct {
	Matches   []Match `json:"matches"`
	Timestamp int64   `json:"timestamp"`
	Timezone  string  `json:"timezone"`
	TagID     string  `json:"tagid"`
	Track     Track   `json:"track"`
}

type shazamAPI struct {
	url      string
	host     string
	key      string
	payload  *strings.Reader
	response ShazamResponse
}

func (s *shazamAPI) ReadFile() {
	data, err := ioutil.ReadFile("output.wav")
	if err != nil {
		panic(err)
	}

	encodedString := base64.StdEncoding.EncodeToString(data)

	s.payload = strings.NewReader(encodedString)
}

func (s *shazamAPI) CallAPI() {
	req, err := http.NewRequest("POST", s.url, s.payload)
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}

	req.Header.Add("content-type", "text/plain")
	req.Header.Add("X-RapidAPI-Key", s.key)
	req.Header.Add("X-RapidAPI-Host", s.host)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	s.response = ShazamResponse{}

	err = json.Unmarshal(body, &s.response)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *shazamAPI) GetSong() {
	s.ReadFile()
	s.CallAPI()

}

func run(commChannels *structs.CommChannels) {

	shazam := shazamAPI{
		url:  "https://shazam.p.rapidapi.com/songs/v2/detect?timezone=Europe%2FParis&locale=fr-FR",
		host: "shazam.p.rapidapi.com",
		key:  "68870327c5msh33d760331b46bd5p144895jsn48821e689b9b",
	}

	for range commChannels.FetchAPI {
		shazam.GetSong()
		fmt.Print(shazam.response)

	}

}

func Api(commChannels *structs.CommChannels) *gobot.Robot {
	work := func() {
		run(commChannels)
	}

	robot := gobot.NewRobot("api",
		work,
	)

	return robot

}
