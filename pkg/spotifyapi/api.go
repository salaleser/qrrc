package spotifyapi

import (
	"fmt"
	"os"
	"qrrc/internal/pkg/webhelper"

	"github.com/zmb3/spotify"
)

const state = "pidor777"

var (
	web    *webhelper.WebHelper
	client *spotify.Client
	auth   = spotify.NewAuthenticator(
		fmt.Sprintf("https://%s/spotify/callback", os.Getenv("QRRC_REDIRECT_HOST")),
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
	)
	ch = make(chan *spotify.Client)
)

func Start() {
	go func() {
		client = <-ch

		user, err := client.CurrentUser()
		if err != nil {
			fmt.Printf("error: get current user: %v\n", err)
		}
		fmt.Println("Logged in as:", user.ID)

		ps, err := client.PlayerState()
		if err != nil {
			fmt.Printf("error: get player state: %v\n", err)
		}
		fmt.Printf("Found your %s (%s)\n", ps.Device.Type, ps.Device.Name)
	}()
}
