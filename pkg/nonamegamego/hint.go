package nonamegamego

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

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

func (n *NonaMegaMego) updateHints() map[int]hint {
	return map[int]hint{
		1: {
			id:    1,
			text:  "Первая буква наименования артиста",
			value: 9,
			f:     n.hintArtistTitleFirstLetter,
		},
		2: {
			id:    2,
			text:  "Первая буква наименования альбома",
			value: 3,
			f:     n.hintAlbumTitleFirstLetter,
		},
		3: {
			id:    3,
			text:  "Первая буква наименования трека",
			value: 5,
			f:     n.hintTrackTitleFirstLetter,
		},
		4: {
			id:    4,
			text:  "Количество букв в наименовании артиста",
			value: 8,
			f:     n.hintArtistTitleLettersCount,
		},
		5: {
			id:    5,
			text:  "Количество букв в наименовании альбома",
			value: 4,
			f:     n.hintAlbumTitleLettersCount,
		},
		6: {
			id:    6,
			text:  "Количество букв в наименовании трека",
			value: 6,
			f:     n.hintTrackTitleLettersCount,
		},
		7: {
			id:    7,
			text:  "Дата релиза",
			value: 5,
			f:     n.hintAlbumReleaseDate,
		},
		8: {
			id:    8,
			text:  "Обложка альбома",
			value: 24,
			f:     n.hintAlbumImage,
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

func (n *NonaMegaMego) hintAlbumTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Album.Title)
	return t.Album.Title[:c]
}

func (n *NonaMegaMego) hintTrackTitleFirstLetter() string {
	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err.Error()
	}
	_, c := utf8.DecodeRuneInString(t.Title)
	return t.Title[:c]
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
