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
	playlist spotifyhelper.Playlist
	settings settings
	round    round
}

type settings struct {
	playersCount int
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

	options := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	buttons := make(Buttons, len(options))
	for i, v := range options {
		buttons[i] = Button{
			Link: "setup",
			Text: v,
			Params: url.Values{
				"players": {strconv.Itoa(i + 1)},
			},
		}
	}

	n.web.LoadSetupPage(
		fmt.Sprintf("Выберите количество участников: (%d)", n.settings.playersCount),
		buttons.Join(" | "),
	)

	return nil
}

func (n *NonaMegaMego) handleMain(params url.Values) error {
	startParam := params.Get("start")
	hintParam := params.Get("hint")
	if startParam == "true" {
		n.round = round{
			number: 1,
			turn: turn{
				hint:  []string{},
				hints: n.updateHints(),
			},
		}
		if err := n.s.StartRandomPlaylist(n.playlist.ID); err != nil {
			return errors.Wrap(err, "play random")
		}
	} else if hintParam != "" {
		hintID, err := strconv.Atoi(hintParam)
		if err != nil {
			return errors.Wrap(err, "parse hint")
		}

		hint, ok := n.round.turn.hints[hintID]
		if !ok {
			return errors.New("Нет подсказки с таким ID")
		}
		delete(n.round.turn.hints, hintID)
		n.round.turn.hint = append(n.round.turn.hint,
			fmt.Sprintf("%s: <b>%s</b>", hint.text, hint.f()))
		n.stats.ActivePlayer().AddScore(-hint.value)
	} else {
		if err := n.s.StartRandomPlaylist(n.playlist.ID); err != nil {
			return errors.Wrap(err, "play random")
		}
		if n.stats.SetActiveNext() {
			n.round.number = n.round.number + 1
		}
	}

	buttons := make(Buttons, 0)
	for _, v := range n.round.turn.hints {
		buttons = append(buttons, Button{
			Link: "main",
			Text: v.String(),
			Params: url.Values{
				"hint": {strconv.Itoa(v.id)},
			},
		})
	}

	n.web.LoadMainPage(
		fmt.Sprintf("Раунд: %d", n.round.number),
		n.stats.String(),
		strings.Join(n.round.turn.hint, "<br>"),
		fmt.Sprintf("Ходит <b>%s</b>", n.stats.ActivePlayer().name),
		buttons.Join("<br>"),
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

	n.web.LoadAnswerPage(
		fmt.Sprintf("Правильный ответ: <b>%s</b><br>"+
			"<img src=%q>", t.String(), t.Album.ImageURL),
	)

	return nil
}
