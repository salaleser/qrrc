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
			text:  "Первая буква исполнителя",
			value: 6,
			f:     n.hintArtistTitleFirstLetter,
		},
		2: {
			text:  "Вторая буква исполнителя",
			value: 8,
			f:     n.hintArtistTitleSecondLetter,
		},
		3: {
			text:  "Третья буква исполнителя",
			value: 10,
			f:     n.hintArtistTitleThirdLetter,
		},
		4: {
			text:  "Последняя буква исполнителя",
			value: 6,
			f:     n.hintArtistTitleLastLetter,
		},
		5: {
			text:  "Количество букв в исполнителе",
			value: 6,
			f:     n.hintArtistTitleLettersCount,
		},
		6: {
			text:  "Первая буква трека",
			value: 6,
			f:     n.hintTrackTitleFirstLetter,
		},
		7: {
			text:  "Вторая буква трека",
			value: 8,
			f:     n.hintTrackTitleSecondLetter,
		},
		8: {
			text:  "Третья буква трека",
			value: 12,
			f:     n.hintTrackTitleThirdLetter,
		},
		9: {
			text:  "Последняя буква трека",
			value: 3,
			f:     n.hintTrackTitleLastLetter,
		},
		10: {
			text:  "Количество букв в треке",
			value: 5,
			f:     n.hintTrackTitleLettersCount,
		},
		11: {
			text:  "Структура наименования трека",
			value: 9,
			f:     n.hintTrackTitleStructure,
		},
		12: {
			text:  "Первая буква альбома",
			value: 3,
			f:     n.hintAlbumTitleFirstLetter,
		},
		13: {
			text:  "Количество букв в альбоме",
			value: 4,
			f:     n.hintAlbumTitleLettersCount,
		},
		14: {
			text:  "Дата релиза",
			value: 2,
			f:     n.hintAlbumReleaseDate,
		},
		15: {
			text:  "Обложка альбома",
			value: 6,
			f:     n.hintAlbumImage,
		},
	}
}

func (n *NonaMegaMego) hintArtistTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 1)
}

func (n *NonaMegaMego) hintArtistTitleSecondLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 2)
}

func (n *NonaMegaMego) hintArtistTitleThirdLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 3)
}

func (n *NonaMegaMego) hintArtistTitleLastLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, -1)
}

func (n *NonaMegaMego) hintArtistTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Artist.Title)))
}

func (n *NonaMegaMego) hintTrackTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 1)
}

func (n *NonaMegaMego) hintTrackTitleSecondLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 2)
}

func (n *NonaMegaMego) hintTrackTitleThirdLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 3)
}

func (n *NonaMegaMego) hintTrackTitleLastLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}

	return getChar(t.Title, -1)
}

func (n *NonaMegaMego) hintTrackTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Title)))
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

func (n *NonaMegaMego) hintAlbumTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Album.Title, 1)
}

func (n *NonaMegaMego) hintAlbumTitleLettersCount() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Album.Title)))
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
	return fmt.Sprintf(`<img src="%s" class="hint-album-cover">`,
		t.Album.ImageURL)
}

func getChar(s string, n int) string {
	a := []rune{}
	b := []byte(s)
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		a = append(a, r)
		i += size
	}
	if n < 0 {
		return string(a[len(a)-1])
	}
	return string(a[n-1])
}
