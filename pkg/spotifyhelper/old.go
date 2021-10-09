package spotifyhelper

import (
	"fmt"
	"math/rand"
	"net/http"
	"qrrc/internal/pkg/webhelper"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

func CompleteAuthHandler(w http.ResponseWriter, r *http.Request) {
	Instance.web = webhelper.New(w)

	if Instance.client == nil {
		token, err := Instance.auth.Token(state, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			fmt.Printf("error: get token: %v\n", err)
			return
		}
		if s := r.FormValue("state"); s != state {
			http.Error(w, "State mismatch", http.StatusForbidden)
			fmt.Printf("error: state mismatch: %s != %s\n", s, state)
			return
		}
		client := Instance.auth.NewClient(token)
		Instance.ch <- &client
	}

	Instance.web.LoadPage("home", []string{"text", "toggle_play"},
		[]string{"Успех!", "Toggle Play/Pause"})
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	Instance.web = webhelper.New(w)

	var err error
	var togglePlay string
	var togglePlayImage string
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	params := r.URL.Query()

	if Instance.client == nil {
		Instance.web.LoadErrorPage(
			fmt.Sprintf("<a href=%q>Пройти аутентификацию</a>",
				Instance.auth.AuthURL(state)),
			errors.New("not authenticated in spotify"))
		return
	}

	ps, err := Instance.client.PlayerState()
	if err != nil {
		Instance.web.LoadErrorPage("player state", err)
		return
	}

	initGamePlaylistsImagesCache()

	switch action {
	case "home":
		var text string
		if !ps.Playing {
			togglePlay = "Play"
			togglePlayImage = "https://i.imgur.com/DJbZhLN.png"
			text = "Музыка не играет"
		} else {
			togglePlay = "Pause"
			togglePlayImage = "https://i.imgur.com/doV24uc.png"
			t, err := Instance.GetCurrentTrack()
			if err != nil {
				Instance.web.LoadErrorPage("get current track", err)
				return
			}
			text = t.String()
		}
		Instance.web.LoadPage("home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%q>", togglePlay,
				togglePlayImage)})
	case "game":
		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			p = append(p, fmt.Sprintf("<a href=game/next?playlist=%q>%s</a>", k,
				fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
		}
		sort.Strings(p)
		Instance.web.LoadPage(action, []string{"text", "step", "playlists"},
			[]string{"Жми кнопку и пытайся угадать.", "0",
				strings.Join(p, " ")})
	case "game/next":
		playlistParam := params.Get("playlist")
		playlist, err := Instance.SearchPlaylist(playlistParam)
		if err != nil {
			Instance.web.LoadErrorPage("search playlist", err)
			return
		}

		err = Instance.PlayRandomTrack(playlist)
		if err != nil {
			Instance.web.LoadErrorPage("play random track", err)
			return
		}

		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			p = append(p, fmt.Sprintf("<a href=next?playlist=%q>%s</a>", k,
				fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
		}
		sort.Strings(p)
		Instance.web.LoadPage("game", []string{"text", "step", "playlists"},
			[]string{"Угадай", "0", strings.Join(p, " ")})
	case "game/hint":
		stepParam := params.Get("step")
		step, err := strconv.Atoi(stepParam)
		if err != nil {
			Instance.web.LoadErrorPage("parse step", err)
			return
		}

		var text string
		if ps.Playing {
			ft, err := Instance.client.GetTrack(ps.Item.ID)
			if err != nil {
				Instance.web.LoadErrorPage("get track", err)
				return
			}
			fa, err := Instance.client.GetArtist(ft.Artists[0].ID)
			if err != nil {
				Instance.web.LoadErrorPage("get artist", err)
				return
			}
			stp, err := Instance.client.GetAlbumTracks(ft.Album.ID)
			if err != nil {
				Instance.web.LoadErrorPage("get album tracks", err)
				return
			}
			trackList := "<ul>"
			for _, v := range stp.Tracks {
				trackList += fmt.Sprintf("<li>%v</li>", v.Name)
			}

			faName := strings.TrimLeft(fa.Name, "The ")
			_, faw := utf8.DecodeRuneInString(faName)
			_, ftw := utf8.DecodeRuneInString(ft.Name)
			hints := make([]string, 0)
			hints = append(hints,
				fmt.Sprintf("Первая буква имени исполнителя %q",
					faName[:faw]))
			hints = append(hints,
				fmt.Sprintf("Первая буква названия трека %q",
					ft.Name[:ftw]))
			hints = append(hints,
				fmt.Sprintf("Количество слов в названии трека %d",
					len(strings.Split(ft.Name, " "))))
			hints = append(hints, fmt.Sprintf("Жанры: %s",
				strings.Join(fa.Genres, ", ")))
			hints = append(hints,
				fmt.Sprintf("Фото исполнителя:<img src=%q>",
					fa.Images[0].URL))
			hints = append(hints,
				fmt.Sprintf("Исполнитель %q", fa.Name))
			hints = append(hints,
				fmt.Sprintf("Обложка альбома:<img src=%q>",
					ft.Album.Images[0].URL))
			hints = append(hints,
				fmt.Sprintf("Список треков с альбома:<br/>%s",
					trackList+"</ul>"))

			if step >= len(hints) {
				t, err := Instance.GetCurrentTrack()
				if err != nil {
					Instance.web.LoadErrorPage("", err)
					return
				}
				text = t.String()
				step = 0
			} else {
				text = fmt.Sprintf("Подсказка #%d/%d: %s",
					step+1, len(hints), hints[step])
				step++
			}
		} else {
			text = "Музыка не играет"
		}

		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			p = append(p, fmt.Sprintf("<a href=next?playlist=%q>%s</a>", k,
				fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
		}
		sort.Strings(p)
		Instance.web.LoadPage("game", []string{"text", "step", "playlists"},
			[]string{text, strconv.Itoa(step), strings.Join(p, " ")})
	case "settings":
		deviceIDParam := params.Get("device_id")
		d, err := Instance.GetDevices()
		if err != nil {
			Instance.web.LoadErrorPage("get devices", err)
			return
		}
		devices := make(Devices, len(*d))
		for i, v := range *d {
			if v.ID.String() == deviceIDParam {
				err := Instance.client.NextOpt(&spotify.PlayOptions{DeviceID: &v.ID})
				if err != nil {
					Instance.web.LoadErrorPage("next opt", err)
					return
				}
				err = Instance.client.PauseOpt(&spotify.PlayOptions{DeviceID: &v.ID})
				if err != nil {
					Instance.web.LoadErrorPage("pause opt", err)
					return
				}
				v.Active = true
			}
			devices[i] = v
		}

		Instance.web.LoadSettingsPage(devices.Join("<br>"))
		return
	case "help":
		Instance.web.LoadPage(action, []string{}, []string{})
	case "play":
		artistParam := params.Get("artist")
		albumParam := params.Get("album")
		if artistParam == "" && albumParam == "" {
			ps, err := Instance.client.PlayerState()
			if err != nil {
				Instance.web.LoadErrorPage("player state", err)
				return
			}
			if ps.Playing {
				err = Instance.client.Pause()
				if err != nil {
					Instance.web.LoadErrorPage("pause", err)
					return
				}
			} else {
				err = Instance.client.Play()
				if err != nil {
					Instance.web.LoadErrorPage("play", err)
					return
				}
			}
		} else {
			sr, err := Instance.client.Search(
				fmt.Sprintf("%s %s", artistParam, albumParam),
				spotify.SearchTypeAlbum)
			if err != nil {
				Instance.web.LoadErrorPage("search", err)
				return
			}
			stp, err := Instance.client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				Instance.web.LoadErrorPage("get album tracks", err)
				return
			}
			err = Instance.client.QueueSong(stp.Tracks[rand.Intn(stp.Total-1)].ID)
			if err != nil {
				Instance.web.LoadErrorPage("queue song", err)
				return
			}
			err = Instance.client.Next()
			if err != nil {
				Instance.web.LoadErrorPage("next", err)
				return
			}
		}
		var text string
		if ps.Playing {
			togglePlay = "Play"
			togglePlayImage = "https://i.imgur.com/DJbZhLN.png"
			text = "Музыка не играет"
		} else {
			togglePlay = "Pause"
			togglePlayImage = "https://i.imgur.com/doV24uc.png"
			t, err := Instance.GetCurrentTrack()
			if err != nil {
				Instance.web.LoadErrorPage("get track", err)
				return
			}
			text = t.String()
		}
		Instance.web.LoadPage("home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%s>", togglePlay,
				togglePlayImage)})
	default:
		http.NotFound(w, r)
	}
}
