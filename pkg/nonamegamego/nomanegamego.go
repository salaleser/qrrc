package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/spotifyhelper"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type NonaMegaMego struct {
	web      *webhelper.WebHelper
	s        *spotifyhelper.SpotifyHelper
	stats    *Stats
	settings settings
	round    round
}

type settings struct {
	playersCount           int
	playlist               *spotifyhelper.Playlist
	answerCorrectAll       int
	answerCorrectPartially int
	answerIncorrect        int
}

func (s *settings) String() string {
	return fmt.Sprintf("Количество игроков: %d<br>Плейлист: %s",
		s.playersCount, s.playlist,
	)
}

type round struct {
	number int
	turn   turn
}

type turn struct {
	hint  []string
	hints map[int]hint
}

type handler func(params url.Values) error

var handlers map[string]handler

func New(web *webhelper.WebHelper, s *spotifyhelper.SpotifyHelper) *NonaMegaMego {
	n := &NonaMegaMego{
		web: web,
		s:   s,
		settings: settings{
			playersCount: 2,
		},
	}

	handlers = map[string]handler{
		"start":  n.handleStart,
		"setup":  n.handleSetup,
		"main":   n.handleMain,
		"answer": n.handleAnswer,
	}

	return n
}

func (n *NonaMegaMego) Update(web *webhelper.WebHelper) {
	n.web = web
}

func (n *NonaMegaMego) Route(action string, params url.Values) error {
	h, ok := handlers[action]
	if !ok {
		return errors.New("не удалось найти подходящий обработчик")
	}

	if err := h(params); err != nil {
		return errors.Wrap(err, "handle")
	}

	return nil
}

func (n *NonaMegaMego) handleStart(params url.Values) error {
	n.web.LoadStartPage()

	return nil
}

func (n *NonaMegaMego) handleSetup(params url.Values) error {
	var err error

	playersCountParam := params.Get("players_count")
	if playersCountParam != "" {
		n.settings.playersCount, err = strconv.Atoi(playersCountParam)
		if err != nil {
			return errors.Wrap(err, "не удалось спарсить число участников")
		}
	}

	playlistParam := params.Get("playlist")
	if playlistParam == "-" {
		n.settings.playlist = nil
	} else if playlistParam != "" {
		n.settings.playlist, err = n.s.SearchPlaylist(playlistParam)
		if err != nil {
			return errors.Wrap(err, "search playlist")
		}
	}

	n.settings.answerCorrectAll = 100
	n.settings.answerCorrectPartially = 50
	n.settings.answerIncorrect = -50

	playersCountOptions := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9",
	}
	playersCount := make(Buttons, len(playersCountOptions))
	for i, v := range playersCountOptions {
		playersCount[i] = Button{
			Link: "setup",
			Text: v,
			Params: url.Values{
				"players_count": {v},
			},
		}
	}

	playerNames := make(Fields, n.settings.playersCount)
	for i := 0; i < n.settings.playersCount; i++ {
		playerNames[i] = Field{
			Text: fmt.Sprintf("Игрок %d", i+1),
		}
	}

	playlistOptions := []struct {
		text string
		code string
	}{
		{"BALDEJ", "BALDEJ"},
		{"Русский рок", "русский+рок"},
		{"Russian Pop (1980)", "russian+pop+1980"},
		{"Best of Rock (1970)", "best+of+rock+1970"},
		{"(использовать текущий плейлист)", "-"},
	}
	playlists := make(Buttons, len(playlistOptions))
	for i, v := range playlistOptions {
		playlists[i] = Button{
			Link: "setup",
			Text: v.text,
			Params: url.Values{
				"playlist": {v.code},
			},
		}
	}

	n.web.LoadSetupPage(
		n.settings.String(),
		playersCount.Join(" "),
		playerNames.Join("<br>"),
		playlists.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleMain(params url.Values) error {
	startParam := params.Get("start")
	hintParam := params.Get("hint")

	if startParam == "true" {
		playerNamesParam := params.Get("player_names")
		playerNames := strings.Split(playerNamesParam, ",")
		n.settings.playersCount = len(playerNames)
		n.stats = NewStats(playerNames)
		n.round = round{
			number: 1,
			turn: turn{
				hint:  []string{},
				hints: n.updateHints(),
			},
		}
		if n.settings.playlist != nil {
			err := n.s.PlayRandomTrack(n.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := n.s.PlayNextTrack()
			if err != nil {
				return errors.Wrap(err, "play next track")
			}
		}
	} else if hintParam != "" {
		hintID, err := strconv.Atoi(hintParam)
		if err != nil {
			return errors.Wrap(err, "parse hint")
		}

		if hint, ok := n.round.turn.hints[hintID]; ok {
			delete(n.round.turn.hints, hintID)
			n.round.turn.hint = append(n.round.turn.hint,
				fmt.Sprintf("%s: <b>%s</b>", hint.text, hint.f()))
			n.stats.ActivePlayer().AddScore(-hint.value)
		}
	} else {
		if n.settings.playlist != nil {
			err := n.s.PlayRandomTrack(n.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := n.s.PlayNextTrack()
			if err != nil {
				return errors.Wrap(err, "play next track")
			}
		}

		correctParam := params.Get("correct")
		correct, err := strconv.Atoi(correctParam)
		if err != nil {
			return errors.Wrap(err, "parse correct")
		}
		n.stats.ActivePlayer().AddScore(correct)

		if n.stats.SetActiveNext() {
			n.round.number = n.round.number + 1
		}
	}

	hints := make(Buttons, 0)
	for k, v := range n.round.turn.hints {
		hints = append(hints, Button{
			Link: "main",
			Text: v.String(),
			Params: url.Values{
				"hint": {strconv.Itoa(k)},
			},
		})
	}

	n.web.LoadMainPage(
		strconv.Itoa(n.round.number),
		n.stats.String(),
		strings.Join(n.round.turn.hint, "<br>"),
		n.stats.ActivePlayer().name,
		hints.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleAnswer(params url.Values) error {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err
	}

	n.round.turn = turn{
		hint:  []string{},
		hints: n.updateHints(),
	}

	options := []struct {
		text  string
		value int
	}{
		{"Правильно все", n.settings.answerCorrectAll},
		{"Правильно частично", n.settings.answerCorrectPartially},
		{"Неправильно", n.settings.answerIncorrect},
	}
	buttons := make(Buttons, len(options))
	for i, v := range options {
		buttons[i] = Button{
			Link: "main",
			Text: v.text,
			Params: url.Values{
				"correct": {strconv.Itoa(v.value)},
			},
		}
	}

	n.web.LoadAnswerPage(
		fmt.Sprintf("%s<br><img src=%q>", t.String(), t.Album.ImageURL),
		buttons.Join("<br>"),
	)

	return nil
}
