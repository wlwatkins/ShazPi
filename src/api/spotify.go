package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

const redirectURI = "http://localhost:8080/callback"

var loginChan = make(chan bool)

type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

type SotifyResponse struct {
	Tracks Tracks `json:"tracks"`
}

type Tracks struct {
	Href   string  `json:"href"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
	Items  []Items `json:"items"`
}
type Items struct {
	Type string `json:"type"`
	Uri  string `json:"uri"`
}

type spotifyAPI struct {
	add_playlist_url string
	search_url       string
	host             string
	clientID         string
	clientSecret     string
	data             string
	playlist_id      string
	token_search     SpotifyTokenResponse
	token_login      SpotifyTokenResponse
	expireTime       int64
	state            string
}

func sendEmail(text string) {
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", text)

}

func generateRandomString(length int) string {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	// Available characters for random string generation
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate random string
	var sb strings.Builder
	for i := 0; i < length; i++ {
		index := rng.Intn(len(charset))
		sb.WriteByte(charset[index])
	}

	return sb.String()
}

func (s *spotifyAPI) completeAuth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	state := queryParams.Get("state")

	if state != s.state {
		log.Fatalf("States are different. This is a security risk... %s vs %s", state, s.state)
	}

	Authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(s.clientID+":"+s.clientSecret))

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	requestBody := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", requestBody)
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", Authorization)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	err = json.Unmarshal(body, &s.token_login)
	if err != nil {
		log.Fatalf("Could not decode response JSON: %s", err)
	}

	currentTime := time.Now().Unix()
	s.token_login.ExpiresAt = currentTime + int64(s.token_login.ExpiresIn)

	loginChan <- true

}

func (s *spotifyAPI) RefreshToken() {

	Authorization := "Basic " + base64.StdEncoding.EncodeToString([]byte(s.clientID+":"+s.clientSecret))

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", s.token_login.RefreshToken)
	requestBody := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", requestBody)
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", Authorization)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	err = json.Unmarshal(body, &s.token_login)
	if err != nil {
		log.Fatalf("Could not decode response JSON: %s", err)
	}

	currentTime := time.Now().Unix()
	s.token_login.ExpiresAt = currentTime + int64(s.token_login.ExpiresIn)

}

func (s *spotifyAPI) Login() {
	// first start an HTTP server
	http.HandleFunc("/callback", s.completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	s.state = generateRandomString(16)
	scope := "user-read-private  playlist-modify-private playlist-read-private playlist-read-collaborative playlist-modify-public"

	url := "https://accounts.spotify.com/authorize?" +
		url.Values{
			"response_type": {"code"},
			"client_id":     {s.clientID},
			"scope":         {scope},
			"redirect_uri":  {redirectURI},
			"state":         {s.state},
		}.Encode()
	sendEmail(url)
}

func (s *spotifyAPI) GetAccessToken() {

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)

	requestBody := strings.NewReader(data.Encode())

	url := "https://accounts.spotify.com/api/token"
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(body, &s.token_search)
	if err != nil {
		log.Fatalf("Could not decode response JSON: %s", err)
	}

	currentTime := time.Now().Unix()
	s.token_search.ExpiresAt = currentTime + int64(s.token_search.ExpiresIn)
	cfg := Config{}.loadToml()
	cfg.Spotify.TokenSearch = s.token_search
	cfg.saveToml()
}

func (s *spotifyAPI) SearchTrack(uri string) string {
	currentTime := time.Now().Unix()
	if currentTime >= s.expireTime {
		s.GetAccessToken()
	}

	url := strings.Replace(s.search_url, "{uri}", uri, 1)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.token_search.AccessToken)
	req.Header.Add("X-RapidAPI-Host", s.host)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", res.StatusCode)
	}
	var spotifyReponse SotifyResponse
	err = json.Unmarshal(body, &spotifyReponse)
	if err != nil {
		log.Fatalf("Could not decode response JSON: %s", err)
	}

	fmt.Println(spotifyReponse)

	return spotifyReponse.Tracks.Items[0].Uri
}

func (s *spotifyAPI) AddToPlaylist(trackUri string) {
	currentTime := time.Now().Unix()
	if currentTime >= s.expireTime {
		s.RefreshToken()
	}

	url := strings.Replace(s.add_playlist_url, "{playlist_id}", s.playlist_id, 1)
	url = strings.Replace(url, "{track_ui}", trackUri, 1)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(s.data)))
	if err != nil {
		log.Fatalf("Could not get res %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.token_login.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not get res %s", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))

	if res.StatusCode != http.StatusCreated {
		log.Fatalf("Unexpected status code: %d", res.StatusCode)
	}

}

func (s *spotifyAPI) LoadTokens() {

	tomlFile := "toeks.toml"
	file, err := os.Open(tomlFile)
	if err != nil {
		log.Fatalf("Failed to open TOML file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := toml.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Failed to parse TOML: %v", err)
	}
}

func (s *spotifyAPI) EstablishAcces() {
	if s.token_login.AccessToken == "" || s.token_login.RefreshToken == "" {
		s.Login()
		<-loginChan

		cfg := Config{}.loadToml()
		cfg.Spotify.TokenLogin = s.token_login
		cfg.saveToml()

	}

	currentTime := time.Now().Unix()
	if currentTime >= s.token_login.ExpiresAt {
		s.RefreshToken()
	}

	if s.token_search.AccessToken == "" {
		s.GetAccessToken()

		cfg := Config{}.loadToml()
		cfg.Spotify.TokenLogin = s.token_login
		cfg.saveToml()

	}
}

func (s *spotifyAPI) AddSong(song *ShazamResponse) {

	s.EstablishAcces()

	for _, provider := range song.Track.Hub.Providers {
		if provider.Type == "SPOTIFY" {
			uri := provider.Actions[0].Uri
			trackUri := s.SearchTrack(uri)
			s.AddToPlaylist(trackUri)
			return
		}
	}
	fmt.Println(song.Track.Hub.Providers)
	fmt.Println("Could not find track in spotify. Need to send email with info")

}
