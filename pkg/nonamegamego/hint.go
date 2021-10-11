package nonamegamego

import (
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type hint struct {
	text  string
	image string
	value int
	f     hintHandler
}

type hintHandler func() string

func (h *hint) String() string {
	return fmt.Sprintf(`<img class="hints" src="/images/%s.png" alt="%s">%d`,
		h.image, h.text, h.value)
}

func (n *NonaMegaMego) updateHints() map[int]hint {
	return map[int]hint{
		11: {
			text:  "Первая буква исполнителя",
			image: "artist-first-letter",
			value: 6,
			f:     n.hintArtistTitleFirstLetter,
		},
		12: {
			text:  "Вторая буква исполнителя",
			image: "artist-second-letter",
			value: 8,
			f:     n.hintArtistTitleSecondLetter,
		},
		13: {
			text:  "Третья буква исполнителя",
			image: "artist-third-letter",
			value: 9,
			f:     n.hintArtistTitleThirdLetter,
		},
		14: {
			text:  "Последняя буква исполнителя",
			image: "artist-last-letter",
			value: 6,
			f:     n.hintArtistTitleLastLetter,
		},
		15: {
			text:  "Количество букв в исполнителе",
			image: "artist-letters-count",
			value: 6,
			f:     n.hintArtistTitleLettersCount,
		},
		21: {
			text:  "Первая буква трека",
			image: "track-first-letter",
			value: 6,
			f:     n.hintTrackTitleFirstLetter,
		},
		22: {
			text:  "Вторая буква трека",
			image: "track-second-letter",
			value: 8,
			f:     n.hintTrackTitleSecondLetter,
		},
		23: {
			text:  "Третья буква трека",
			image: "track-third-letter",
			value: 9,
			f:     n.hintTrackTitleThirdLetter,
		},
		24: {
			text:  "Последняя буква трека",
			image: "track-last-letter",
			value: 3,
			f:     n.hintTrackTitleLastLetter,
		},
		25: {
			text:  "Количество букв в треке",
			image: "track-letters-count",
			value: 5,
			f:     n.hintTrackTitleLettersCount,
		},
		26: {
			text:  "Структура наименования трека",
			image: "track-name-structure",
			value: 9,
			f:     n.hintTrackTitleStructure,
		},
		31: {
			text:  "Дата релиза",
			image: "release-date",
			value: 2,
			f:     n.hintAlbumReleaseDate,
		},
		32: {
			text:  "Обложка альбома",
			image: "album-cover",
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
