package spotifyapi

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"

	"github.com/zmb3/spotify"
)

func CompleteAuthHandler(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		token, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			fmt.Printf("error: get token: %v\n", err)
		}
		if s := r.FormValue("state"); s != state {
			http.Error(w, "State mismatch: %s != %s", http.StatusForbidden)
			fmt.Printf("State mismatch: %s != %s\n", s, state)
		}
		client := auth.NewClient(token)
		ch <- &client
		return
	}

	html, err := ioutil.ReadFile("html/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: auth read file home: %v\n", err)
	}
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(html)
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: get player state: %v", err)
			}
			if ps.Playing {
				err = client.Pause()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					fmt.Printf("error: play: toggle to pause: %v", err)
				}
			} else {
				err = client.Play()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					fmt.Printf("error: play: toggle to play: %v", err)
				}
			}
		} else {
			sr, err := client.Search(fmt.Sprintf("%s %s", artist, album), spotify.SearchTypeAlbum)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: search: %v", err)
				return
			}
			stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: get album tracks: %v", err)
				return
			}
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: queue song: %v", err)
				return
			}
			err = client.Next()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: next: %v", err)
			}
		}
	case "pause":
		err = client.Pause()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: pause: pause: %v", err)
		}
	case "next":
		err = client.Next()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: next: next: %v", err)
		}
	case "previous":
		err = client.Previous()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: previous: previous: %v", err)
		}
	case "settings":
		html, err := ioutil.ReadFile("html/settings.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: settings: read file: %v", err)
		}
		_, err = w.Write(html)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: settings: write: %v", err)
		}
		return
	case "help":
		html, err := ioutil.ReadFile("html/help.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: help: read file: %v", err)
		}
		_, err = w.Write(html)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: help: write: %v", err)
		}
		return
	default:
		http.NotFound(w, r)
	}
	html, err := ioutil.ReadFile("html/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: play: read file: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: write: %v", err)
	}
}
