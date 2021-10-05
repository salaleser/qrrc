package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/webhelper"
	"strconv"

	"github.com/pkg/errors"
)

type NonaMegaMego struct {
	web   *webhelper.WebHelper
	stats *Stats
	round int
}

type handler func(params url.Values) error

var handlers map[string]handler

func New(web *webhelper.WebHelper) *NonaMegaMego {
	n := &NonaMegaMego{
		web:   web,
		round: 1,
	}

	handlers = map[string]handler{
		"start": n.handleStart,
		"main":  n.handleMain,
		"round": n.handleRound,
	}

	return n
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
			Params: url.Values{
				"n": {strconv.Itoa(n.round)},
			},
		},
	})

	n.web.LoadMainPage(
		fmt.Sprintf("Раунд: %d", n.round),
		n.stats.String(),
		fmt.Sprintf("Ходит %s", n.stats.Active().name),
		buttons.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleRound(params url.Values) error {
	n.web.LoadRoundPage("test")
	return nil
}
