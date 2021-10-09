package nonamegamego

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type hint struct {
	text  string
	value int
	f     hintHandler
}

type hintHandler func() string

func (h *hint) String() string {
	return fmt.Sprintf("%s (%d)", h.text, h.value)
}

func (n *NonaMegaMego) updateHints() map[int]hint {
	return map[int]hint{
		1: {
			text:  "Первая буква наименования артиста",
			value: 6,
			f:     n.hintArtistTitleFirstLetter,
		},
		2: {
			text:  "Первая буква наименования альбома",
			value: 3,
			f:     n.hintAlbumTitleFirstLetter,
		},
		3: {
			text:  "Первая буква наименования трека",
			value: 6,
			f:     n.hintTrackTitleFirstLetter,
		},
		4: {
			text:  "Количество букв в наименовании артиста",
			value: 6,
			f:     n.hintArtistTitleLettersCount,
		},
		5: {
			text:  "Количество букв в наименовании альбома",
			value: 4,
			f:     n.hintAlbumTitleLettersCount,
		},
		6: {
			text:  "Количество букв в наименовании трека",
			value: 5,
			f:     n.hintTrackTitleLettersCount,
		},
		7: {
			text:  "Дата релиза",
			value: 2,
			f:     n.hintAlbumReleaseDate,
		},
		8: {
			text:  "Обложка альбома",
			value: 6,
			f:     n.hintAlbumImage,
		},
		9: {
			text:  "Последняя буква наименования артиста",
			value: 4,
			f:     n.hintArtistTitleLastLetter,
		},
		10: {
			text:  "Вторая буква наименования трека",
			value: 8,
			f:     n.hintTrackTitleSecondLetter,
		},
		11: {
			text:  "Последняя буква наименования трека",
			value: 3,
			f:     n.hintTrackTitleLastLetter,
		},
		12: {
			text:  "Структура наименования трека",
			value: 9,
			f:     n.hintTrackTitleStructure,
		},
		13: {
			text:  "Вторая буква наименования трека",
			value: 12,
			f:     n.hintTrackTitleThirdLetter,
		},
	}
}

func (n *NonaMegaMego) hintArtistTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Artist.Title)
	return t.Artist.Title[:c]
}

func (n *NonaMegaMego) hintArtistTitleLastLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}

	a := []rune{}
	b := []byte(t.Artist.Title)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		a = append(a, r)
		i += size
	}
	return string(a[len(a)-1])
}

func (n *NonaMegaMego) hintAlbumTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Album.Title)
	return t.Album.Title[:c]
}

func (n *NonaMegaMego) hintArtistTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Artist.Title)))
}

func (n *NonaMegaMego) hintAlbumTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Album.Title)))
}

func (n *NonaMegaMego) hintTrackTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Title)))
}

func (n *NonaMegaMego) hintAlbumReleaseDate() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return t.Album.ReleaseDate
}

func (n *NonaMegaMego) hintAlbumImage() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("<img src=%q>", t.Album.ImageURL)
}

func (n *NonaMegaMego) hintTrackTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Title)
	return t.Title[0*c : 1*c]
}

func (n *NonaMegaMego) hintTrackTitleSecondLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Title)
	return t.Title[1*c : 2*c]
}

func (n *NonaMegaMego) hintTrackTitleThirdLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Title)
	return t.Title[2*c : 3*c]
}

func (n *NonaMegaMego) hintTrackTitleLastLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}

	a := []rune{}
	b := []byte(t.Title)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		a = append(a, r)
		i += size
	}
	return string(a[len(a)-1])
}

func (n *NonaMegaMego) hintTrackTitleStructure() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}

	l := regexp.MustCompile(`[а-яА-ЯёЁa-zA-Z0-9]`)
	b := l.ReplaceAll([]byte(t.Title), []byte("?"))
	return string(b)
}
