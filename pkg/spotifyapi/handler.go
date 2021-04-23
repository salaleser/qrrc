package spotifyapi

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

const ErrNoActiveDeviceFound = "Player command failed: No active device found"

var gamePlaylists = []string{
	"metalcore",
	"hip-hop",
	"classic+punk",
	"emo+2007",
	"2000s+russian+pop",
	"top500",
	"best+of+rock+1970",
	"best+of+rock+1980",
	"best+of+rock+1990",
	"best+of+rock+2000",
	"best+of+rock+2007",
	"best+of+rock+2010",
}

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
			fmt.Printf("error: state mismatch: %s != %s\n", s, state)
			return
		}
		client := auth.NewClient(token)
		ch <- &client
	}

	loadPage(w, "home", []string{"text", "toggle_play"},
		[]string{"Успех! Теперь можешь управлять спотифаем или поиграть в " +
			"угадаечку.", "Toggle Play/Pause"})
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var togglePlay string
	var togglePlayImage string
	var text string
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	query := r.URL.Query()

	if client == nil {
		loadPage(w, "error", []string{"text"}, []string{fmt.Sprintf("Сорян, "+
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
		loadPage(w, "home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%s>", togglePlay,
				togglePlayImage)})
	case "game":
		playlists := ""
		for _, v := range gamePlaylists {
			if v == "top500" {
				playlists += fmt.Sprintf("<a href=game/next?playlist=top500>" +
					"<img class=playlist alt=\"Top 500\"></a>")
			} else {
				sr, err := client.Search(v, spotify.SearchTypePlaylist)
				if err != nil {
					handleError(w, err, "game search")
					return
				}
				p := sr.Playlists.Playlists[0]

				playlists += fmt.Sprintf(
					"<a href=game/next?playlist=%s>%s</a>", v,
					fmt.Sprintf("<img class=playlist src=%s alt=%s>",
						p.Images[0].URL, p.Name))
			}
		}
		loadPage(w, action, []string{"text", "step", "playlists"},
			[]string{"Жми кнопку и пытайся угадать.", "0", playlists})
	case "game/next":
		playlist := query.Get("playlist")
		var sr *spotify.SearchResult
		var err error
		if playlist == "top500" {
			file, err := os.Open("tracks.txt")
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: open file: %v", err)
				return
			}
			defer func() {
				err := file.Close()
				if err != nil {
					loadPage(w, "error", []string{"text"},
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
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: scan file: %v", err)
				return
			}
			line := strings.Split(lines[rand.Intn(len(lines))], "\t")
			searchQuery := fmt.Sprintf("%s %s", line[1], line[0])
			sr, err = client.Search(searchQuery, spotify.SearchTypeTrack)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: search: %v\n", err)
				return
			}
		} else {
			sr, err = client.Search(playlist, spotify.SearchTypePlaylist)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: search: %v\n", err)
				return
			}
		}

		if sr == nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					"Поиск не вернул результат")})
			fmt.Printf("error: game: next: %v\n", "sr == nil")
			return
		}

		var track spotify.FullTrack
		if playlist == "top500" {
			if sr.Tracks.Total == 0 {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков по запросу", playlist)})
				fmt.Printf("error: game: next: %s (%s)\n", "sr.Tracks.Total == 0",
					playlist)
				return
			}

			track = sr.Tracks.Tracks[0]
		} else {
			if sr.Playlists.Total == 0 {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков по запросу", playlist)})
				fmt.Printf("error: game: next: %s (%s)\n", "sr.Playlists.Total == 0",
					playlist)
				return
			}

			ptp, err := client.GetPlaylistTracks(sr.Playlists.Playlists[0].ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						err.Error(), playlist)})
				fmt.Printf("error: game: next: get playlist tracks: %v\n", err)
				return
			}

			if ptp.Total == 0 {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
						"Не найдено треков в плейлисте",
						sr.Playlists.Playlists[0].Name)})
				fmt.Printf("error: game: next: %s (%s)\n", "ptp.Total == 0",
					playlist)
				return
			}

			track = ptp.Tracks[rand.Intn(ptp.Total)].Track
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

		loadPage(w, "game", []string{"text", "step"},
			[]string{"Запущен трек, попытайтесь отгадать!", "0"})
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

			hints := make([]string, 0)
			hints = append(hints, fmt.Sprintf("Первая буква имени исполнителя %q",
				strings.TrimLeft(fa.Name, "The ")[0]))
			hints = append(hints, fmt.Sprintf("Первая буква названия трека %q",
				ft.Name[0]))
			hints = append(hints, fmt.Sprintf("Количество слов в названии трека %d",
				len(strings.Split(ft.Name, " "))))
			hints = append(hints, fmt.Sprintf("Жанр %q",
				strings.Join(fa.Genres, ", ")))
			if ft.Explicit {
				hints = append(hints, fmt.Sprintf("Содержит матюки"))
			} else {
				hints = append(hints, fmt.Sprintf("Без матюков"))
			}
			hints = append(hints, fmt.Sprintf("Фото исполнителя: <img src=\"%s\">",
				fa.Images[0].URL))
			hints = append(hints, fmt.Sprintf("Исполнитель %q", fa.Name))

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
		loadPage(w, "game", []string{"text", "step"}, []string{text,
			strconv.Itoa(step)})
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
		loadPage(w, action, []string{"text", "option_1"},
			[]string{"Зарегистрированные устройства:", devicesList})
	case "help":
		loadPage(w, action, []string{}, []string{})
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
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
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
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		loadPage(w, "home", []string{"text", "toggle_play"}, []string{text,
			fmt.Sprintf("<img class=button alt=%q src=%s>", togglePlay,
				togglePlayImage)})
	default:
		http.NotFound(w, r)
	}
}

func loadPage(w http.ResponseWriter, p string, old []string, new []string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("template/%s.html", p))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: read file: %v", err)
	}
	w.Header().Set("Content-Type", "text/html")
	for i := 0; i < len(old); i++ {
		html = []byte(strings.Replace(string(html), "{{"+old[i]+"}}", new[i],
			-1))
	}
	_, err = w.Write(html)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: %v", err)
	}
}

func activateFirstDevice(w http.ResponseWriter) error {
	devices, err := client.PlayerDevices()
	if err != nil {
		return errors.Wrap(err, "player devices")
	}

	if len(devices) == 0 {
		return errors.Wrap(err, "no devices found")
	}

	device := devices[0]
	err = client.PlayOpt(&spotify.PlayOptions{DeviceID: &device.ID})
	if err != nil {
		return errors.Wrap(err, "play with options")
	}
	err = client.Pause()
	if err != nil {
		handleError(w, err, "settings: pause")
		return errors.Wrap(err, "pause")
	}

	loadPage(w, "home", []string{"text", "toggle_play"},
		[]string{fmt.Sprintf("Устройство %q (%s) активировано.",
			device.Name, device.Type), "<img class=button alt=\"Toggle Play\">"})
	fmt.Printf("device %q activated\n", device.Name)

	return nil
}

func handleError(w http.ResponseWriter, err error, message string) {
	if err.Error() == ErrNoActiveDeviceFound {
		err := activateFirstDevice(w)
		if err != nil {
			handleError(w, err, message)
		}
	} else {
		loadPage(w, "error", []string{"text"},
			[]string{fmt.Sprintf("Спотифай выключен! Попроси хозяина "+
				"запустить его.<br/><p class=error>Сообщение об ошибке: %s"+
				"</p>", err.Error())})
		fmt.Printf("error: %s: %v\n", message, err)
		return
	}
}
