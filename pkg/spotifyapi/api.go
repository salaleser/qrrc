package spotifyapi

import (
	"fmt"
	"os"

	"github.com/zmb3/spotify"
)

var client *spotify.Client
var state string

var (
	auth = spotify.NewAuthenticator(
		fmt.Sprintf("https://%s/spotify/callback", os.Getenv("REDIRECT_HOST")),
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
	)
	ch = make(chan *spotify.Client)
)

func Start() {
	state = os.Getenv("SPOTIFY_STATE")
	print(-10)
	go func() {
		client = <-ch
		print(-9)

		user, err := client.CurrentUser()
		print(-8)
		if err != nil {
			fmt.Printf("error: get current user: %v\n", err)
		}
		print(-7)
		fmt.Println("Logged in as:", user.ID)

		print(-6)
		ps, err := client.PlayerState()
		print(-5)
		if err != nil {
			fmt.Printf("error: get player state: %v\n", err)
		}
		print(-4)
		fmt.Printf("Found your %s (%s)\n", ps.Device.Type, ps.Device.Name)
	}()
}
