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
			togglePlay = "Продолжить воспроизведение"
			text += "Музыка не играет."
		} else {
			togglePlay = "Остановить воспроизведение"
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		loadPage(w, action, []string{"text", "toggle_play"}, []string{text,
			togglePlay})
	case "game":
		cp, err := client.GetCategories()
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: search: %v\n", err)
			return
		}
		playlists := fmt.Sprintf("<img src=%s>", cp.Categories[0].Icons[0].URL)
		loadPage(w, action, []string{"text", "step", "playlists"},
			[]string{"Жми кнопку и пытайся угадать.", "0", playlists})
	case "game/next":
		playlist := query.Get("playlist")
		var sr *spotify.SearchResult
		var err error
		var searchQuery string
		if playlist == "top500" {
			file, err := os.Open("tracks.txt")
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>", err.Error())})
				fmt.Printf("error: game: next: open file: %v", err)
				return
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			var lines []string
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
		}

		sr, err = client.Search(searchQuery, spotify.SearchTypePlaylist)
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: search: %v\n", err)
			return
		}

		if sr == nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					"Поиск не вернул результат")})
			fmt.Printf("error: game: next: %v\n", "sr == nil")
			return
		}

		if sr.Tracks.Total == 0 {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s %q</p>",
					"Не найдено треков по запросу", searchQuery)})
			fmt.Printf("error: game: next: %s (%s)\n", "sr.Tracks.Total == 0",
				searchQuery)
			return
		}

		if playlist == "top500" {
			err = client.QueueSong(sr.Tracks.Tracks[0].ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
						err.Error())})
				fmt.Printf("error: game: next: queue song: %v\n", err)
				return
			}
		} else {
			//
		}

		err = client.Next()
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: next: %v\n", err)
			return
		}

		client.Seek(sr.Tracks.Tracks[0].Duration / 3)
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: next: %v\n", err)
			return
		}
		loadPage(w, "game", []string{"text", "step"},
			[]string{"Запущен трек, попытайтесь отгадать!", "0"})
	case "game/show":
		if ps.Playing {
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "game: show: get track")
				return
			}
			text += fmt.Sprintf("Это был: \"%s — %s\"", ft.Artists[0].Name,
				ps.Item.Name)
		} else {
			text += "Музыка не играет."
		}
		loadPage(w, "game", []string{"text", "step"}, []string{text, "0"})
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
				text += "Подсказок больше нет."
				step = 0
			} else {
				text += fmt.Sprintf("Подсказка #%d: %s", step+1, hints[step])
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
				client.PlayOpt(&spotify.PlayOptions{DeviceID: &v.ID})
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
			togglePlay = "Продолжить воспроизведение"
			text += "Музыка не играет."
		} else {
			togglePlay = "Остановить воспроизведение"
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				handleError(w, err, "get track")
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		}
		loadPage(w, "home", []string{"text", "toggle_play"}, []string{text,
			togglePlay})
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

	err = client.PlayOpt(&spotify.PlayOptions{DeviceID: &devices[0].ID})
	if err != nil {
		return errors.Wrap(err, "play with options")
	}

	loadPage(w, "home", []string{"text", "toggle_play"},
		[]string{fmt.Sprintf("Device %s (%s) activated",
			devices[0].Name, devices[0].Type), "Toggle Play"})
	fmt.Printf("device %q activated", devices[0].Name)

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
