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
		[]string{"Успех! Теперь можешь управлять спотифаем или поиграть в " +
			"угадаечку.", "Toggle Play/Pause"})
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	Instance.web = webhelper.New(w)

	var err error
	var togglePlay string
	var togglePlayImage string
	var text string
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	query := r.URL.Query()

	if Instance.client == nil {
		Instance.web.LoadPage("error", []string{"text"}, []string{fmt.Sprintf("Сорян, "+
			"братиш (bro), аутентификацию в спотифае (Spotify) владелец"+
			" аккаунта еще не прошел. Если у тебя есть права, то пройди сам "+
			"(do it yourself) по <a href=%q>этой ссылке</a>. Или попроси "+
			"того, кто сможет.", Instance.auth.AuthURL(state))})
		return
	}

	ps, err := Instance.client.PlayerState()
	if err != nil {
		handleError(w, err, "player state")
		return
	}

	initGamePlaylistsImagesCache()

	switch action {
	case "home":
		if !ps.Playing {
			togglePlay = "Play"
			togglePlayImage = "https://i.imgur.com/DJbZhLN.png"
			text += "Музыка не играет."
		} else {
			togglePlay = "Pause"
			togglePlayImage = "https://i.imgur.com/doV24uc.png"
			ft, err := Instance.client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
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
		q := query.Get("playlist")
		playlist, err := Instance.SearchPlaylist(q)
		if err != nil {
			Instance.web.LoadErrorPage(q, err)
			return
		}

		if err := Instance.StartRandomPlaylist(playlist.ID); err != nil {
			Instance.web.LoadErrorPage(q, err)
			return
		}

		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			p = append(p, fmt.Sprintf("<a href=next?playlist=%q>%s</a>", k,
				fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
		}
		sort.Strings(p)
		Instance.web.LoadPage("game", []string{"text", "step", "playlists"},
			[]string{"Запущен трек, попытайтесь отгадать!", "0",
				strings.Join(p, " ")})
	case "game/hint":
		step, err := strconv.Atoi(query.Get("step"))
		if err != nil {
			handleError(w, err, "game: hint: get parameter")
			return
		}

		var text string
		if ps.Playing {
			ft, err := Instance.client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "game: hint: get track")
				return
			}
			fa, err := Instance.client.GetArtist(ft.Artists[0].ID)
			if err != nil {
				handleError(w, err, "game: hint: get artist")
				return
			}
			stp, err := Instance.client.GetAlbumTracks(ft.Album.ID)
			if err != nil {
				handleError(w, err, "game: hint: get album tracks")
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
				text, err = Instance.CurrentTrack()
				if err != nil {
					Instance.web.LoadErrorPage("", err)
					return
				}
				step = 0
			} else {
				text = fmt.Sprintf("Подсказка #%d/%d: %s", step+1, len(hints),
					hints[step])
				step++
			}
		} else {
			text = "Музыка не играет."
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
		devices, err := Instance.client.PlayerDevices()
		if err != nil {
			handleError(w, err, "settings: player devices")
			return
		}
		deviceIDParameter := query.Get("deviceId")
		for _, v := range devices {
			if v.ID.String() == deviceIDParameter {
				err := Instance.client.PlayOpt(&spotify.PlayOptions{DeviceID: &v.ID})
				if err != nil {
					handleError(w, err, "settings: play with options")
					return
				}
				err = Instance.client.Pause()
				if err != nil {
					handleError(w, err, "settings: pause")
					return
				}
			}
		}
		var devicesList string
		for _, v := range devices {
			if v.Active {
				devicesList += fmt.Sprintf(
					"<p>%s [%s] %d%%</p>", v.Name, v.Type, v.Volume)
			} else {
				devicesList += fmt.Sprintf(
					"<a href=/spotify/settings?deviceId=%s>%s [%s] %d%%"+
						"</a><br/>", v.ID.String(), v.Name, v.Type, v.Volume)
			}
		}
		Instance.web.LoadPage(action, []string{"text", "option_1"},
			[]string{"Зарегистрированные устройства:", devicesList})
	case "help":
		Instance.web.LoadPage(action, []string{}, []string{})
	case "play":
		artist := query.Get("artist")
		album := query.Get("album")
		if artist == "" && album == "" {
			ps, err := Instance.client.PlayerState()
			if err != nil {
				handleError(w, err, "play: get player state")
				return
			}
			if ps.Playing {
				err = Instance.client.Pause()
				if err != nil {
					handleError(w, err, "play: toggle to pause")
					return
				}
			} else {
				err = Instance.client.Play()
				if err != nil {
					handleError(w, err, "play: toggle to play")
					return
				}
			}
		} else {
			sr, err := Instance.client.Search(fmt.Sprintf("%s %s", artist, album),
				spotify.SearchTypeAlbum)
			if err != nil {
				handleError(w, err, "play: search")
				return
			}
			stp, err := Instance.client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				handleError(w, err, "play: get album tracks")
				return
			}
			err = Instance.client.QueueSong(stp.Tracks[rand.Intn(stp.Total-1)].ID)
			if err != nil {
				handleError(w, err, "play: queue song")
				return
			}
			err = Instance.client.Next()
			if err != nil {
				handleError(w, err, "play: next")
				return
			}
		}
		if ps.Playing {
			togglePlay = "Play"
			togglePlayImage = "https://i.imgur.com/DJbZhLN.png"
			text += "Музыка не играет."
		} else {
			togglePlay = "Pause"
			togglePlayImage = "https://i.imgur.com/doV24uc.png"
			ft, err := Instance.client.GetTrack(ps.CurrentlyPlaying.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		Instance.web.LoadPage("home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%s>", togglePlay,
				togglePlayImage)})
	default:
		http.NotFound(w, r)
	}
}
