package spotifyapi

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"qrrc/internal/pkg/webhelper"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

const ErrNoActiveDeviceFound = "Player command failed: No active device found"

var (
	gamePlaylistsImages map[string]string
)

func CompleteAuthHandler(w http.ResponseWriter, r *http.Request) {
	web = webhelper.New(w)

	if client == nil {
		token, err := auth.Token(state, r)
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
		client := auth.NewClient(token)
		ch <- &client
	}

	web.LoadPage("home", []string{"text", "toggle_play"},
		[]string{"Успех! Теперь можешь управлять спотифаем или поиграть в " +
			"угадаечку.", "Toggle Play/Pause"})
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	web = webhelper.New(w)

	var err error
	var togglePlay string
	var togglePlayImage string
	var text string
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	query := r.URL.Query()

	if client == nil {
		web.LoadPage("error", []string{"text"}, []string{fmt.Sprintf("Сорян, "+
			"братиш (bro), аутентификацию в спотифае (Spotify) владелец"+
			" аккаунта еще не прошел. Если у тебя есть права, то пройди сам "+
			"(do it yourself) по <a href=%q>этой ссылке</a>. Или попроси "+
			"того, кто сможет.", auth.AuthURL(state))})
		return
	}

	ps, err := client.PlayerState()
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
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		web.LoadPage("home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%q>", togglePlay,
				togglePlayImage)})
	case "game":
		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			if k == "top500" {
				p = append(p, fmt.Sprintf("<a href=game/next?playlist=top500>%s</a>",
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", "-", k)))
			} else {
				p = append(p, fmt.Sprintf("<a href=game/next?playlist=%q>%s</a>", k,
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
			}
		}
		sort.Strings(p)
		web.LoadPage(action, []string{"text", "step", "playlists"},
			[]string{"Жми кнопку и пытайся угадать.", "0",
				strings.Join(p, " ")})
	case "game/next":
		playlist := query.Get("playlist")
		var sr *spotify.SearchResult
		var err error
		if playlist == "top500" {
			file, err := os.Open("tracks.txt")
			if err != nil {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: open file: %v", err)
				return
			}
			defer func() {
				err := file.Close()
				if err != nil {
					web.LoadPage("error", []string{"text"},
						[]string{fmt.Sprintf(
							"<p class=\"error\">Ошибка: %s</p>", err.Error())})
					fmt.Printf("error: game: next: close file: %v", err)
					return
				}
			}()
			scanner := bufio.NewScanner(file)
			lines := make([]string, 0)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			err = scanner.Err()
			if err != nil {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: scan file: %v", err)
				return
			}
			line := strings.Split(lines[rand.Intn(len(lines)-1)], "\t")
			searchQuery := fmt.Sprintf("%s %s", line[1], line[0])
			sr, err = client.Search(searchQuery, spotify.SearchTypeTrack)
			if err != nil {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: search: %v\n", err)
				return
			}
		} else {
			sr, err = client.Search(playlist, spotify.SearchTypePlaylist)
			if err != nil {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %v</p>",
						err)})
				fmt.Printf("error: game: next: search: %v\n", err)
				return
			}
		}

		if sr == nil {
			web.LoadPage("error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					"Поиск не вернул результат")})
			fmt.Print("error: game: next: sr == nil\n")
			return
		}

		var track spotify.FullTrack
		if playlist == "top500" {
			if sr.Tracks.Total == 0 {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков по запросу", playlist)})
				fmt.Printf("error: game: next: sr.Tracks.Total == 0 (%s)\n",
					playlist)
				return
			}
			track = sr.Tracks.Tracks[0]
		} else {
			if sr.Playlists.Total == 0 {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков по запросу", playlist)})
				fmt.Printf("error: game: next: %s (%s)\n", "sr.Playlists.Total == 0",
					playlist)
				return
			}

			ptp, err := client.GetPlaylistTracks(sr.Playlists.Playlists[0].ID)
			if err != nil {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %v %q</p>",
						err, playlist)})
				fmt.Printf("error: game: next: get playlist tracks: %v\n", err)
				return
			}

			if ptp.Total == 0 {
				web.LoadPage("error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков в плейлисте",
						sr.Playlists.Playlists[0].Name)})
				fmt.Printf("error: game: next: ptp.Total == 0 (%s)\n",
					playlist)
				return
			}

			track = ptp.Tracks[rand.Intn(len(ptp.Tracks)-1)].Track
		}

		err = client.QueueSong(track.ID)
		if err != nil {
			handleError(w, err, "game: game: next: queue song")
			return
		}

		err = client.Next()
		if err != nil {
			handleError(w, err, "game: game: next: next")
			return
		}

		err = client.Seek(track.Duration/4 + rand.Intn(track.Duration/4))
		if err != nil {
			handleError(w, err, "game: game: next: seek")
			return
		}

		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			if k == "top500" {
				p = append(p, fmt.Sprintf("<a href=next?playlist=top500>%s</a>",
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", "-", k)))
			} else {
				p = append(p, fmt.Sprintf("<a href=next?playlist=%q>%s</a>", k,
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
			}
		}
		sort.Strings(p)
		web.LoadPage("game", []string{"text", "step", "playlists"},
			[]string{"Запущен трек, попытайтесь отгадать!", "0",
				strings.Join(p, " ")})
	case "game/hint":
		step, err := strconv.Atoi(query.Get("step"))
		if err != nil {
			handleError(w, err, "game: hint: get parameter")
			return
		}

		if ps.Playing {
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "game: hint: get track")
				return
			}
			fa, err := client.GetArtist(ft.Artists[0].ID)
			if err != nil {
				handleError(w, err, "game: hint: get artist")
				return
			}
			stp, err := client.GetAlbumTracks(ft.Album.ID)
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
				text += fmt.Sprintf("Это был: \"%s — %s\"", ft.Artists[0].Name,
					ps.Item.Name)
				if err := client.Pause(); err != nil {
					handleError(w, err, "game: hint: pause")
					return
				}
				step = 0
			} else {
				text += fmt.Sprintf("Подсказка #%d/%d: %s", step+1, len(hints),
					hints[step])
				step++
			}
		} else {
			text += "Музыка не играет."
		}

		p := make([]string, 0)
		for k, v := range gamePlaylistsImages {
			if k == "top500" {
				p = append(p, fmt.Sprintf("<a href=next?playlist=top500>%s</a>",
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", "-", k)))
			} else {
				p = append(p, fmt.Sprintf("<a href=next?playlist=%q>%s</a>", k,
					fmt.Sprintf("<img class=playlist src=%q alt=%q>", v, k)))
			}
		}
		sort.Strings(p)
		web.LoadPage("game", []string{"text", "step", "playlists"},
			[]string{text, strconv.Itoa(step), strings.Join(p, " ")})
	case "settings":
		devices, err := client.PlayerDevices()
		if err != nil {
			handleError(w, err, "settings: player devices")
			return
		}
		deviceIDParameter := query.Get("deviceId")
		for _, v := range devices {
			if v.ID.String() == deviceIDParameter {
				err := client.PlayOpt(&spotify.PlayOptions{DeviceID: &v.ID})
				if err != nil {
					handleError(w, err, "settings: play with options")
					return
				}
				err = client.Pause()
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
		web.LoadPage(action, []string{"text", "option_1"},
			[]string{"Зарегистрированные устройства:", devicesList})
	case "help":
		web.LoadPage(action, []string{}, []string{})
	case "play":
		artist := query.Get("artist")
		album := query.Get("album")
		if artist == "" && album == "" {
			ps, err := client.PlayerState()
			if err != nil {
				handleError(w, err, "play: get player state")
				return
			}
			if ps.Playing {
				err = client.Pause()
				if err != nil {
					handleError(w, err, "play: toggle to pause")
					return
				}
			} else {
				err = client.Play()
				if err != nil {
					handleError(w, err, "play: toggle to play")
					return
				}
			}
		} else {
			sr, err := client.Search(fmt.Sprintf("%s %s", artist, album),
				spotify.SearchTypeAlbum)
			if err != nil {
				handleError(w, err, "play: search")
				return
			}
			stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				handleError(w, err, "play: get album tracks")
				return
			}
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total-1)].ID)
			if err != nil {
				handleError(w, err, "play: queue song")
				return
			}
			err = client.Next()
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
			ft, err := client.GetTrack(ps.CurrentlyPlaying.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		web.LoadPage("home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%s>", togglePlay,
				togglePlayImage)})
	default:
		http.NotFound(w, r)
	}
}

func activateFirstDevice() (*spotify.PlayerDevice, error) {
	devices, err := client.PlayerDevices()
	if err != nil {
		return nil, errors.Wrap(err, "player devices")
	}

	var device *spotify.PlayerDevice
	for i, v := range devices {
		fmt.Printf("%d: %v\n", i, v)
		if v.Active {
			device = &v
			fmt.Printf("%q is activated\n", v.Name)
			return device, nil
		}
	}

	if device == nil {
		return nil, errors.Wrap(err, "no active devices found")
	}

	err = client.PlayOpt(&spotify.PlayOptions{DeviceID: &device.ID})
	if err != nil {
		return nil, errors.Wrap(err, "play with options")
	}

	return device, nil
}

func handleError(w http.ResponseWriter, err error, message string) {
	if err.Error() == ErrNoActiveDeviceFound {
		device, err := activateFirstDevice()
		if err != nil {
			// FIXME
			web.LoadPage("error", []string{"text"},
				[]string{fmt.Sprintf("Спотифай выключен! Попроси хозяина "+
					"запустить его.<br/><p class=error>Сообщение об ошибке: %v"+
					"</p>", err)})
			fmt.Printf("error: %s: %v\n", message, err)
			return
		}

		if device == nil {
			web.LoadPage("home", []string{"text", "toggle_play"},
				[]string{"Не удалось автоматически активировать устройство, " +
					"выберите устройство вручную в настройках",
					"<img class=button alt=\"Toggle Play\">"})
			return
		}

		web.LoadPage("home", []string{"text", "toggle_play"},
			[]string{fmt.Sprintf("Устройство %q (%s) активировано.",
				device.Name, device.Type), "<img class=button alt=\"Toggle Play\">"})
		fmt.Printf("device %q activated\n", device.Name)
	} else {
		web.LoadPage("error", []string{"text"},
			[]string{fmt.Sprintf("Спотифай выключен! Попроси хозяина "+
				"запустить его.<br/><p class=error>Сообщение об ошибке: %s"+
				"</p>", err.Error())})
		fmt.Printf("error: %s: %v\n", message, err)
	}
}

func initGamePlaylistsImagesCache() {
	if gamePlaylistsImages != nil {
		return
	}

	playlistsFile, err := os.Open("playlists.txt")
	if err != nil {
		fmt.Printf("error: init: load playlists: %v", err)
		return
	}
	defer playlistsFile.Close()
	scanner := bufio.NewScanner(playlistsFile)
	playlists := make([]string, 0)
	for scanner.Scan() {
		playlists = append(playlists, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		fmt.Printf("error: init: scan playlists file: %v", err)
		return
	}

	gamePlaylistsImages = make(map[string]string)
	for _, v := range playlists {
		sr, err := client.Search(v, spotify.SearchTypePlaylist)
		if err != nil {
			fmt.Printf("error: init game playlists images cache: search %q", v)
			continue
		}

		if len(sr.Playlists.Playlists) > 0 &&
			len(sr.Playlists.Playlists[0].Images) > 0 {
			gamePlaylistsImages[v] = sr.Playlists.Playlists[0].Images[0].URL
		}
	}
}
