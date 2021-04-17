package spotifyapi

import (
	"fmt"
	"github.com/zmb3/spotify"
	"log"
	"os"
)

var client *spotify.Client
var playerState *spotify.PlayerState

var (
	auth = spotify.NewAuthenticator(
		fmt.Sprintf("https://%s/spotify/callback", os.Getenv("REDIRECT_HOST")),
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
	)
	ch    = make(chan *spotify.Client)
	state = "sal123ale456ser789"
)

func Start() {
	go func() {
		// log in to Spotify by visiting the following page in your browser
		fmt.Println(auth.AuthURL(state))

		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Logged in as:", user.ID)

		playerState, err = client.PlayerState()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Found your %s (%s)\n", playerState.Device.Type, playerState.Device.Name)
	}()
}
