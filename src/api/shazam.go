package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"shazammini/src/structs"
	"strings"
)

type ShazamResponse struct {
	Matches   []structs.Match `json:"matches"`
	Timestamp int64           `json:"timestamp"`
	Timezone  string          `json:"timezone"`
	TagID     string          `json:"tagid"`
	Track     structs.Track   `json:"track"`
}

type shazamAPI struct {
	url      string
	host     string
	key      string
	payload  *strings.Reader
	response ShazamResponse
}

func (s *shazamAPI) ReadFile() {
	file, err := os.Open("temp/output.wav")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
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
		log.Println(err)
		fmt.Println(string(body))
	}
}

func (s *shazamAPI) GetSong() {
	s.ReadFile()
	s.CallAPI()

}
