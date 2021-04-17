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
			http.Error(w, err.Error(), http.StatusForbidden)
			fmt.Printf("error: get token: %v\n", err)
			return
		}
		if s := r.FormValue("state"); s != state {
			http.Error(w, "State mismatch", http.StatusForbidden)
			fmt.Printf("State mismatch: %s != %s\n", s, state)
			return
		}
		client := auth.NewClient(token)
		ch <- &client
	}

	ps, err := client.PlayerState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: player state: %v\n", err)
		return
	}
	text := "<br/>"
	if ps.Playing {
		ft, err := client.GetTrack(ps.Item.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: get track: %v\n", err)
		}
		text = fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name, ps.Item.Name)
	} else {
		text = "Музыка не играет."
	}
	loadPageReplace(w, "home", "text",
		"Успех! Теперь можешь управлять спотифаем ("+ps.Device.Name+
			") или поиграть в угадаечку."+text)
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		loadPageReplace(w, "auth", "auth_link", auth.AuthURL(state))
		return
	}

	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	var err error
	switch action {
	case "auth":
		loadPageReplace(w, action, "{{auth_link}}", auth.AuthURL(state))
		return
	case "home":
		loadPage(w, action)
		return
	case "settings":
		loadPage(w, action)
		return
	case "help":
		loadPage(w, action)
		return
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
		loadPage(w, "home")
	case "pause":
		err = client.Pause()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: pause: pause: %v", err)
		}
		loadPage(w, "home")
	case "next":
		err = client.Next()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: next: next: %v", err)
		}
		loadPage(w, "home")
	case "previous":
		err = client.Previous()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: previous: previous: %v", err)
		}
		loadPage(w, "home")
	default:
		http.NotFound(w, r)
	}
}

func loadPage(w http.ResponseWriter, p string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("html/%s.html", p))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page: read file: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page: %v", err)
	}
}

func loadPageReplace(w http.ResponseWriter, p string, old string, new string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("html/%s.html", p))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: read file: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	html = []byte(strings.Replace(string(html), "{{"+old+"}}", new, -1))
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: %v", err)
	}
}
