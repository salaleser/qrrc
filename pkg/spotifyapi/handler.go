package spotifyapi

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/zmb3/spotify"
)

const errorFormat = `Спотифай выключен! Попроси хозяина запустить его или потыкать
в уже включенный, чтобы удалось подцепить его.<br/><p class="error">Сообщение об
ошибке: %s</p>`

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
	if client == nil {
		loadPage(w, "error", []string{"text"}, []string{fmt.Sprintf("Сорян, "+
			"братиш (bro), аутентификацию в спотифае (Spotify) владелец"+
			" аккаунта еще не прошел. Если у тебя есть права, то пройди сам "+
			"(do it yourself) по <a href=%q>этой ссылке</a>. Или попроси "+
			"того, кто сможет.", auth.AuthURL(state))})
		return
	}

	var err error
	ps, err := client.PlayerState()
	if err != nil {
		loadPage(w, "error", []string{"text"},
			[]string{fmt.Sprintf(errorFormat, err.Error())})
		fmt.Printf("error: player state: %v\n", err)
		return
	}

	var togglePlay string
	var text string
	action := strings.TrimPrefix(r.URL.Path, "/spotify/")
	switch action {
	case "home":
		// __повтор кода__
		if ps.Playing {
			togglePlay = "Продолжить воспроизведение"
			text += "Музыка не играет."
		} else {
			togglePlay = "Остановить воспроизведение"
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: get track: %v\n", err)
				return
			}
			text += fmt.Sprintf("Сейчас играет: %s — %s", ft.Artists[0].Name,
				ps.Item.Name)
		} // ^^повтор кода^^
		loadPage(w, action, []string{"text", "toggle_play"}, []string{text,
			togglePlay})
		return
	case "game":
		loadPage(w, action, []string{"text"}, []string{"Правила пока просты " +
			"и меняются: жми кнопку и пытайся угадать."})
		return
	case "game/next":
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
		fmt.Printf("debug: line=%s", strings.Join(line, "-"))
		sr, err := client.Search(fmt.Sprintf("%s %s", line[1], line[0]),
			spotify.SearchTypeTrack)
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: search: %v\n", err)
			return
		}
		if sr.Tracks.Total == 0 {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					"Не найдено треков в спотифае")})
			fmt.Printf("error: game: next: %v\n", "sr.Tracks.Total == 0")
			return
		}
		err = client.QueueSong(sr.Tracks.Tracks[0].ID)
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: queue song: %v\n", err)
			return
		}
		err = client.Next()
		if err != nil {
			loadPage(w, "error", []string{"text"},
				[]string{fmt.Sprintf("<p class=\"error\">Ошибка: %s</p>",
					err.Error())})
			fmt.Printf("error: game: next: next: %v\n", err)
			return
		}
		loadPage(w, "game", []string{"text"},
			[]string{"Запущен трек, попытайтесь отгадать!"})
		return
	case "game/show":
		if ps.Playing {
			ft, err := client.GetTrack(ps.Item.ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: get track: %v\n", err)
				return
			}
			text += fmt.Sprintf("Это был: \"%s — %s\"", ft.Artists[0].Name,
				ps.Item.Name)
		} else {
			text += "Музыка не играет."
		}
		loadPage(w, "game", []string{"text"}, []string{text})
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
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: play: get player state: %v\n", err)
				return
			}
			if ps.Playing {
				err = client.Pause()
				if err != nil {
					loadPage(w, "error", []string{"text"},
						[]string{fmt.Sprintf(errorFormat, err.Error())})
					fmt.Printf("error: play: toggle to pause: %v\n", err)
					return
				}
			} else {
				err = client.Play()
				if err != nil {
					loadPage(w, "error", []string{"text"},
						[]string{fmt.Sprintf(errorFormat, err.Error())})
					fmt.Printf("error: play: toggle to play: %v\n", err)
					return
				}
			}
		} else {
			sr, err := client.Search(fmt.Sprintf("%s %s", artist, album),
				spotify.SearchTypeAlbum)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: play: search: %v\n", err)
				return
			}
			stp, err := client.GetAlbumTracks(sr.Albums.Albums[0].ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: play: get album tracks: %v\n", err)
				return
			}
			err = client.QueueSong(stp.Tracks[rand.Intn(stp.Total)].ID)
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: play: queue song: %v\n", err)
				return
			}
			err = client.Next()
			if err != nil {
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: play: next: %v\n", err)
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
				loadPage(w, "error", []string{"text"},
					[]string{fmt.Sprintf(errorFormat, err.Error())})
				fmt.Printf("error: get track: %v\n", err)
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
