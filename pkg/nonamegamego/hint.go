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

func (r *Room) updateHints() map[int]hint {
	return map[int]hint{
		11: {
			text:  "Первая буква исполнителя",
			image: "artist-first-letter",
			value: 6,
			f:     r.hintArtistTitleFirstLetter,
		},
		12: {
			text:  "Вторая буква исполнителя",
			image: "artist-second-letter",
			value: 8,
			f:     r.hintArtistTitleSecondLetter,
		},
		13: {
			text:  "Третья буква исполнителя",
			image: "artist-third-letter",
			value: 9,
			f:     r.hintArtistTitleThirdLetter,
		},
		14: {
			text:  "Последняя буква исполнителя",
			image: "artist-last-letter",
			value: 6,
			f:     r.hintArtistTitleLastLetter,
		},
		15: {
			text:  "Количество букв в исполнителе",
			image: "artist-letters-count",
			value: 6,
			f:     r.hintArtistTitleLettersCount,
		},
		21: {
			text:  "Первая буква трека",
			image: "track-first-letter",
			value: 6,
			f:     r.hintTrackTitleFirstLetter,
		},
		22: {
			text:  "Вторая буква трека",
			image: "track-second-letter",
			value: 8,
			f:     r.hintTrackTitleSecondLetter,
		},
		23: {
			text:  "Третья буква трека",
			image: "track-third-letter",
			value: 9,
			f:     r.hintTrackTitleThirdLetter,
		},
		24: {
			text:  "Последняя буква трека",
			image: "track-last-letter",
			value: 3,
			f:     r.hintTrackTitleLastLetter,
		},
		25: {
			text:  "Количество букв в треке",
			image: "track-letters-count",
			value: 5,
			f:     r.hintTrackTitleLettersCount,
		},
		26: {
			text:  "Структура наименования трека",
			image: "track-name-structure",
			value: 9,
			f:     r.hintTrackTitleStructure,
		},
		31: {
			text:  "Дата релиза",
			image: "release-date",
			value: 2,
			f:     r.hintAlbumReleaseDate,
		},
		32: {
			text:  "Обложка альбома",
			image: "album-cover",
			value: 6,
			f:     r.hintAlbumImage,
		},
		// 1: {
		// 	text:  "Буква исполнителя",
		// 	image: "selector",
		// 	value: 6,
		// 	f:     r.hintArtistTitleNextLetter,
		// },
		// 2: {
		// 	text:  "Буква трека",
		// 	image: "selector",
		// 	value: 6,
		// 	f:     r.hintTrackTitleNextLetter,
		// },
	}
}

func (r *Room) hintArtistTitleFirstLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 1)
}

func (r *Room) hintArtistTitleSecondLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 2)
}

func (r *Room) hintArtistTitleThirdLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, 3)
}

func (r *Room) hintArtistTitleLastLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Artist.Title, -1)
}

func (r *Room) hintArtistTitleLettersCount() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Artist.Title)))
}

func (r *Room) hintTrackTitleFirstLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 1)
}

func (r *Room) hintTrackTitleSecondLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 2)
}

func (r *Room) hintTrackTitleThirdLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return getChar(t.Title, 3)
}

func (r *Room) hintTrackTitleLastLetter() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}

	return getChar(t.Title, -1)
}

func (r *Room) hintTrackTitleLettersCount() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(utf8.RuneCount([]byte(t.Title)))
}

func (r *Room) hintTrackTitleStructure() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}

	l := regexp.MustCompile(`[а-яА-ЯёЁa-zA-Z0-9]`)
	b := l.ReplaceAll([]byte(t.Title), []byte("?"))
	return string(b)
}

func (r *Room) hintAlbumReleaseDate() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return t.Album.ReleaseDate
}

func (r *Room) hintAlbumImage() string {
	t, err := r.s.CurrentTrack()
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf(`<img src="%s" class="hint-album-cover">`,
		t.Album.ImageURL)
}

// func (r *Room) hintArtistTitleNextLetter() string {
// 	r.round.turn.hintArtistCursor++

// }

// func (r *Room) hintTrackTitleNextLetter() string {

// }

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
