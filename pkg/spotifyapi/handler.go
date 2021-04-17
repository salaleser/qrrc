package spotifyapi

import (
	"fmt"
	"github.com/zmb3/spotify"
	"io/ioutil"
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
			ps, err := client.PlayerState()
			if err != nil {
				fmt.Printf("play get player state: %v", err)
			}
			if ps.Playing {
				err = client.Pause()
				if err != nil {
					fmt.Printf("play toggle to pause: %v", err)
				}
			} else {
				err = client.Play()
				if err != nil {
					fmt.Printf("play toggle to play: %v", err)
				}
			}
		} else {
			sr, err := client.Search(fmt.Sprintf("%s %s", artist, album), spotify.SearchTypeAlbum)
			if err != nil {
				fmt.Printf("play search: %v", err)
				return
			}
			stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				fmt.Printf("play get album tracks: %v", err)
				return
			}
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
			if err != nil {
				fmt.Printf("play queue song: %v", err)
				return
			}
			err = client.Next()
			if err != nil {
				fmt.Printf("play next: %v", err)
			}
		}
	case "pause":
		err = client.Pause()
		if err != nil {
			fmt.Printf("pause pause: %v", err)
		}
	case "next":
		err = client.Next()
		if err != nil {
			fmt.Printf("next next: %v", err)
		}
	case "previous":
		err = client.Previous()
		if err != nil {
			fmt.Printf("previous previous: %v", err)
		}
	case "settings":
		html, err := ioutil.ReadFile("html/settings.html")
		if err != nil {
			fmt.Printf("settings read file: %v", err)
		}
		_, err = rw.Write(html)
		if err != nil {
			fmt.Printf("settings write: %v", err)
		}
		return
	}
	html, err := ioutil.ReadFile("html/index.html")
	if err != nil {
		fmt.Printf("play read file: %v", err)
	}
	rw.Header().Set("Content-Type", "text/html")
	_, err = rw.Write(html)
	if err != nil {
		fmt.Printf("play write: %v", err)
	}
}
