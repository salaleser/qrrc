package nonamegamego

import "fmt"

type hint struct {
	id    int
	text  string
	value int
	f     hintHandler
}

type hintHandler func() string

func (h *hint) String() string {
	return fmt.Sprintf("%s > %d", h.text, h.value)
}

func updateHints() map[int]hint {
	return map[int]hint{
		1: {
			id:    1,
			text:  "Первая буква наименования артиста",
			value: 10,
			f: func() string {
				return "*"
			},
		},
		2: {
			id:    2,
			text:  "Первая буква наименования альбома",
			value: 10,
			f: func() string {
				return "*"
			},
		},
		3: {
			id:    3,
			text:  "Первая буква наименования трека",
			value: 10,
			f: func() string {
				return "*"
			},
		},
		4: {
			id:    4,
			text:  "Количество букв в наименовании артиста",
			value: 10,
			f: func() string {
				return "*"
			},
		},
		5: {
			id:    5,
			text:  "Количество букв в наименовании альбома",
			value: 10,
			f: func() string {
				return "*"
			},
		},
		6: {
			id:    6,
			text:  "Количество букв в наименовании трека",
			value: 10,
			f: func() string {
				return "*"
			},
		},
	}
}
