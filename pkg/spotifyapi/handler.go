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

	loadPage(w, "home", []string{"text"},
		[]string{"Успех! Теперь можешь управлять спотифаем или поиграть в угадаечку."})
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		loadPage(w, "auth", []string{"auth_link"}, []string{auth.AuthURL(state)})
		return
	}

	var err error
	ps, err := client.PlayerState()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: player state: %v\n", err)
		return
	}
	text := "<br/> "
	if ps.Playing {
		ft, err := client.GetTrack(ps.Item.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: get track: %v\n", err)
		}
		text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name, ps.Item.Name)
	} else {
		text += "Музыка не играет."
	}

	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	switch action {
	case "auth":
		loadPage(w, action, []string{"auth_link"}, []string{auth.AuthURL(state)})
		return
	case "home":
		loadPage(w, action, []string{"text"}, []string{text})
		return
	case "settings":
		loadPage(w, action, []string{}, []string{})
		return
	case "help":
		loadPage(w, action, []string{}, []string{})
		return
	case "play":
		query := r.URL.Query()
		artist := query.Get("artist")
		album := query.Get("album")

		if artist == "" && album == "" {
			ps, err := client.PlayerState()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: get player state: %v\n", err)
				return
			}
			if ps.Playing {
				err = client.Pause()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					fmt.Printf("error: play: toggle to pause: %v\n", err)
					return
				}
			} else {
				err = client.Play()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					fmt.Printf("error: play: toggle to play: %v\n", err)
					return
				}
			}
		} else {
			sr, err := client.Search(fmt.Sprintf("%s %s", artist, album), spotify.SearchTypeAlbum)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: search: %v\n", err)
				return
			}
			stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: get album tracks: %v\n", err)
				return
			}
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: queue song: %v\n", err)
				return
			}
			err = client.Next()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				fmt.Printf("error: play: next: %v\n", err)
				return
			}
		}
		loadPage(w, "home", []string{"text"}, []string{text})
	case "pause":
		err = client.Pause()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: pause: pause: %v\n", err)
			return
		}
		loadPage(w, "home", []string{"text"}, []string{text})
	case "next":
		err = client.Next()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: next: next: %v\n", err)
			return
		}
		loadPage(w, "home", []string{"text"}, []string{text})
	case "previous":
		err = client.Previous()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: previous: previous: %v\n", err)
			return
		}
		loadPage(w, "home", []string{"text"}, []string{text})
	default:
		http.NotFound(w, r)
	}
}

func loadPage(w http.ResponseWriter, p string, old []string, new []string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("html/%s.html", p))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: read file: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	for i := 0; i < len(old); i++ {
		html = []byte(strings.Replace(string(html), "{{"+old[i]+"}}", new[i], -1))
	}
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: %v", err)
	}
}
