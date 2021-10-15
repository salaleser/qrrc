package nonamegamego

import (
	"net/url"

	"github.com/pkg/errors"
)

type handler func(params url.Values) error

var handlers map[string]handler

func (g *Game) Route(action string, params url.Values) error {
	h, ok := handlers[action]
	if !ok {
		return errors.New("не удалось найти подходящий обработчик")
	}

	err := h(params)
	if err != nil {
		return errors.Wrap(err, "handle")
	}

	return nil
}
