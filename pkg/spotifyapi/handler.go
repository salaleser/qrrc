package spotifyapi

import (
	"fmt"
	"github.com/zmb3/spotify"
	"log"
	"math/rand"
	"net/http"
	"strings"
)

func CompleteAuthHandler(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	w.Header().Set("Content-Type", "text/html")
	_, _ = fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}

func DefaultHandler(rw http.ResponseWriter, r *http.Request) {
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	var err error
	switch action {
	case "play":
		query := r.URL.Query()
		artist := query.Get("artist")
		album := query.Get("album")

		if artist == "" && album == "" {
			if playerState.Playing {
				err = client.Pause()
				if err != nil {
					fmt.Printf("pause: %v", err)
				}
				_, err = rw.Write([]byte("Paused"))
				if err != nil {
					fmt.Printf("pause write: %v", err)
				}
			} else {
				err = client.Play()
				if err != nil {
					fmt.Printf("play: %v", err)
				}
				_, err = rw.Write([]byte("Paused"))
				if err != nil {
					fmt.Printf("play write: %v", err)
				}
			}
			return
		}

		sr, err := client.Search(fmt.Sprintf("%s %s", artist, album), spotify.SearchTypeAlbum)
		if err != nil {
			log.Print(err)
			return
		}

		stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
		if err != nil {
			log.Print(err)
			return
		}
		err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
		if err != nil {
			log.Print(err)
			return
		}
		err = client.Next()
	case "pause":
		err = client.Pause()
	case "next":
		err = client.Next()
	case "previous":
		err = client.Previous()
	}
	if err != nil {
		log.Print(err)
	}
}
