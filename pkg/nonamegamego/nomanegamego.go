package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/spotifyhelper"
	"strconv"

	"github.com/pkg/errors"
)

var defaultHints = []string{
	"Первая буква наименования артиста",
	"Первая буква наименования альбома",
	"Первая буква наименования трека",
	"Количество букв в наименовании артиста",
	"Количество букв в наименовании альбома",
	"Количество букв в наименовании трека",
}

type NonaMegaMego struct {
	web      *webhelper.WebHelper
	s        *spotifyhelper.SpotifyHelper
	stats    *Stats
	playlist spotifyhelper.Playlist
	settings settings
	round    round
}

type settings struct {
	playersCount int
}

type round struct {
	number int
	hints  []string
}

type handler func(params url.Values) error

var handlers map[string]handler

func New(web *webhelper.WebHelper, s *spotifyhelper.SpotifyHelper) *NonaMegaMego {
	n := &NonaMegaMego{
		web: web,
		s:   s,
		round: round{
			number: 1,
			hints:  defaultHints,
		},
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
	playlistParam := params.Get("playlist")
	var err error
	n.playlist, err = n.s.SearchPlaylist(playlistParam)
	if err != nil {
		return errors.Wrap(err, "get playlist")
	}

	n.web.LoadStartPage()

	return nil
}

func (n *NonaMegaMego) handleSetup(params url.Values) error {
	playersParam := params.Get("players")
	if playersParam != "" {
		var err error
		if n.settings.playersCount, err = strconv.Atoi(playersParam); err != nil {
			return errors.Wrap(err, "не удалось спарсить число участников")
		}
	}

	n.stats = NewStats(n.settings.playersCount)

	buttons := make(Buttons, len(n.round.hints))
	for i, v := range []string{"1", "2", "3", "4", "5"} {
		buttons[i] = Button{
			Link: "setup",
			Text: v,
			Params: url.Values{
				"players": {strconv.Itoa(i + 1)},
			},
		}
	}
	buttons = append(buttons, Button{
		Link: "main",
		Text: "Начать игру",
		Params: url.Values{
			"start": {"true"},
		},
	})

	n.web.LoadSetupPage(
		fmt.Sprintf("Выберите количество участников: (%d)", n.settings.playersCount),
		buttons.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleMain(params url.Values) error {
	startParam := params.Get("start")
	hintParam := params.Get("hint")
	if startParam == "true" {
		if err := n.s.StartRandomPlaylist(n.playlist.ID); err != nil {
			return errors.Wrap(err, "play random")
		}
	} else if hintParam != "" {
		hintNumber, err := strconv.Atoi(hintParam)
		if err != nil {
			return errors.Wrap(err, "parse hint")
		}
		n.round.hints[hintNumber] = "************"
		n.stats.ActivePlayer().AddScore(-10)
	} else {
		if err := n.s.StartRandomPlaylist(n.playlist.ID); err != nil {
			return errors.Wrap(err, "play random")
		}
		if n.stats.SetActiveNext() {
			n.round = round{
				number: n.round.number + 1,
				hints:  defaultHints,
			}
		}
	}

	buttons := make(Buttons, len(n.round.hints))
	for i, v := range n.round.hints {
		buttons[i] = Button{
			Link: "main",
			Text: v,
			Params: url.Values{
				"hint": {strconv.Itoa(i)},
			},
		}
	}

	n.web.LoadMainPage(
		fmt.Sprintf("Раунд: %d", n.round.number),
		n.stats.String(),
		fmt.Sprintf("Ходит <b>%s</b>", n.stats.ActivePlayer().name),
		buttons.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleAnswer(params url.Values) error {
	a, err := n.s.CurrentTrack()
	if err != nil {
		return err
	}

	n.web.LoadAnswerPage(
		fmt.Sprintf("Правильный ответ: %s", a),
	)

	return nil
}
