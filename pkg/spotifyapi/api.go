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
		fmt.Sprintf("http://%s/spotify/callback", os.Getenv("HOST")),
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserModifyPlaybackState,
	)
	ch    = make(chan *spotify.Client)
	state = "sal123ale456ser789"
)

func Start() {
	go func() {
		url := auth.AuthURL(state)

		// TODO auto auth
		//c := http.Client{}
		//req, err := http.NewRequest(http.MethodGet, url, nil)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//res, err := c.Do(req)
		//if err != nil {
		//	log.Println(err)
		//}
		//defer func() {
		//	err := res.Body.Close()
		//	if err != nil {
		//		log.Printf("res close: %v\n", err)
		//	}
		//}()
		//b, err := ioutil.ReadAll(res.Body)
		//if err != nil {
		//	log.Print(err)
		//}
		//fmt.Printf("status %q\nResponse body: %q", res.Status, string(b))

		fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

		// wait for auth to complete
		client = <-ch

		// use the client to make calls that require authorization
		user, err := client.CurrentUser()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("You are logged in as:", user.ID)

		playerState, err = client.PlayerState()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Found your %s (%s)\n", playerState.Device.Type, playerState.Device.Name)
	}()
}
