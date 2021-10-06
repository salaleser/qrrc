package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/spotifyhelper"
	"strconv"

	"github.com/pkg/errors"
)

type NonaMegaMego struct {
	web      *webhelper.WebHelper
	s        *spotifyhelper.SpotifyHelper
	stats    *Stats
	playlist string
	round    int
}

type handler func(params url.Values) error

var handlers map[string]handler

func New(web *webhelper.WebHelper, s *spotifyhelper.SpotifyHelper) *NonaMegaMego {
	n := &NonaMegaMego{
		web:   web,
		s:     s,
		round: 1,
	}

	handlers = map[string]handler{
		"start": n.handleStart,
		"main":  n.handleMain,
		"round": n.handleRound,
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
	n.playlist = "beatles"

	return nil
}

func (n *NonaMegaMego) handleMain(params url.Values) error {
	num, err := strconv.Atoi(params.Get("players"))
	if err != nil {
		return errors.Wrap(err, "не удалось спарсить число участников")
	}

	n.stats = NewStats(num)

	buttons := Buttons([]Button{
		{
			Link: "round",
			Text: "Крутить рулетку",
		},
	})

	n.web.LoadMainPage(
		fmt.Sprintf("Раунд: %d", n.round),
		n.stats.String(),
		fmt.Sprintf("Ходит %s", n.stats.ActivePlayer().name),
		buttons.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleRound(params url.Values) error {
	if n.stats.SetActiveNext() {
		n.round++
	}

	buttons := Buttons([]Button{
		{
			Link: "round",
			Text: "Крутить рулетку",
		},
	})

	n.web.LoadMainPage(
		fmt.Sprintf("Раунд: %d", n.round),
		n.stats.String(),
		fmt.Sprintf("Ходит %s", n.stats.ActivePlayer().name),
		buttons.Join("<br>"),
	)

	return nil
}
