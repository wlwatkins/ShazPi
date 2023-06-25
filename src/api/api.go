package api

import (
	"fmt"
	"log"
	"os"
	"shazammini/src/structs"

	"github.com/pelletier/go-toml"
	"gobot.io/x/gobot"
)

type Config struct {
	Shazam struct {
		Key string `toml:"key"`
	} `toml:"Shazam"`
	Spotify struct {
		PlaylistID   string               `toml:"playlist_id"`
		ClientID     string               `toml:"clientID"`
		ClientSecret string               `toml:"clientSecret"`
		TokenLogin   SpotifyTokenResponse `toml:"TokenLogin"`
		TokenSearch  SpotifyTokenResponse `toml:"TokenSearch"`
	} `toml:"Spotify"`
}

func (cfg Config) loadToml() Config {
	tomlFile := "creds.toml"
	tomlData, err := os.ReadFile(tomlFile)
	if err != nil {
		log.Fatalf("Failed to read TOML file: %v", err)
	}

	err = toml.Unmarshal(tomlData, &cfg)
	if err != nil {
		log.Fatalf("Failed to parse TOML: %v", err)
	}

	return cfg
}
func (cfg Config) saveToml() error {
	// Create a TOML representation of the config struct
	tomlData, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal TOML: %v", err)
	}

	// Write the TOML data to a file
	tomlFile := "creds.toml"
	err = os.WriteFile(tomlFile, tomlData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write TOML file: %v", err)
	}

	return nil
}

func run(commChannels *structs.CommChannels) {

	cfg := Config{}.loadToml()

	shazam := shazamAPI{
		url:  "https://shazam.p.rapidapi.com/songs/v2/detect?timezone=Europe%2FParis&locale=fr-FR",
		host: "shazam.p.rapidapi.com",
		key:  cfg.Shazam.Key,
	}

	spotify := spotifyAPI{
		add_playlist_url: "https://api.spotify.com/v1/playlists/{playlist_id}/tracks?uris={track_ui}",
		playlist_id:      cfg.Spotify.PlaylistID,
		clientID:         cfg.Spotify.ClientID,
		clientSecret:     cfg.Spotify.ClientSecret,
		host:             "api.spotify.com",
		search_url:       "https://api.spotify.com/v1/search?q={uri}&type=track",
		data:             "{ 'uris': ['string'],'position': 0}",
		token_login:      cfg.Spotify.TokenLogin,
		token_search:     cfg.Spotify.TokenSearch,
	}

	_ = spotify

	for range commChannels.FetchAPI {
		shazam.GetSong()
		fmt.Print(shazam.response)
		track := spotify.AddSong(&shazam.response)
		commChannels.DisplayResult <- track
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
