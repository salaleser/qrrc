package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/spotifyhelper"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type NonaMegaMego struct {
	web      *webhelper.WebHelper
	s        *spotifyhelper.SpotifyHelper
	stats    *Stats
	settings settings
	round    round
}

type settings struct {
	playlist *spotifyhelper.Playlist
}

func (s *settings) String() string {
	return s.playlist.String()
}

type round struct {
	number int
	turn   turn
}

type turn struct {
	// hint is the list of hints used by the current player in this turn
	hint []string
	// hints is the list of all unused available hints
	hints    map[int]hint
	started  time.Time
	response time.Duration
}

type handler func(params url.Values) error

var handlers map[string]handler

func New(web *webhelper.WebHelper, s *spotifyhelper.SpotifyHelper) *NonaMegaMego {
	n := &NonaMegaMego{
		web:      web,
		s:        s,
		settings: settings{},
	}

	handlers = map[string]handler{
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

func (n *NonaMegaMego) handleSetup(params url.Values) error {
	var err error
	playlistParam := params.Get("playlist")
	if playlistParam != "" {
		n.settings.playlist, err = n.s.SearchPlaylist(playlistParam)
		if err != nil {
			return errors.Wrap(err, "search playlist")
		}
	} else {
		n.settings.playlist = nil
	}

	playersCount := 5
	playerNames := make(Fields, playersCount)
	for i := 0; i < playersCount; i++ {
		playerNames[i] = Field{
			Text: func() string {
				if i < 2 {
					return fmt.Sprintf("Игрок %d", i+1)
				}
				return ""
			}(),
		}
	}

	n.web.LoadSetupPage(
		n.settings.String(),
		playerNames.Join("<br>"),
	)

	return nil
}

func (n *NonaMegaMego) handleMain(params url.Values) error {
	startParam := params.Get("start")
	hintParam := params.Get("hint")

	if startParam == "true" {
		playerNamesParam := params.Get("player_names")
		playerNames := strings.Split(playerNamesParam, ",")
		n.stats = NewStats(playerNames)
		n.round = round{
			number: 1,
			turn: turn{
				hint:    []string{},
				hints:   n.updateHints(),
				started: time.Now(),
			},
		}
		if n.settings.playlist != nil {
			err := n.s.PlayRandomTrack(n.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := n.s.PlayNextTrack()
			if err != nil {
				return errors.Wrap(err, "play next track")
			}
		}
	} else if hintParam != "" {
		hintID, err := strconv.Atoi(hintParam)
		if err != nil {
			return errors.Wrap(err, "parse hint")
		}

		if hint, ok := n.round.turn.hints[hintID]; ok {
			delete(n.round.turn.hints, hintID)
			n.round.turn.hint = append(n.round.turn.hint,
				fmt.Sprintf("%s: <b>%s</b>", hint.text, hint.f()))
			n.stats.ActivePlayer().AddScore(-float64(hint.value))
		}
	} else {
		if n.settings.playlist != nil {
			err := n.s.PlayRandomTrack(n.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := n.s.PlayNextTrack()
			if err != nil {
				return errors.Wrap(err, "play next track")
			}
		}

		correctParam := params.Get("correct")
		score, err := strconv.ParseFloat(correctParam, 64)
		if err != nil {
			return errors.Wrap(err, "parse correct")
		}
		multiplier := 1.0
		if n.round.turn.response < 2*time.Second {
			multiplier = 1.2
		} else if n.round.turn.response > 90*time.Second {
			multiplier = 0.8
		}

		n.stats.ActivePlayer().AddScore(multiplier * score)

		n.round.turn = turn{
			hint:    []string{},
			hints:   n.updateHints(),
			started: time.Now(),
		}

		if n.stats.SetActiveNext() {
			n.round.number = n.round.number + 1
		}
	}

	hints := make(Buttons, 0)
	for k, v := range n.round.turn.hints {
		hints = append(hints, Button{
			Link: "main",
			Text: v.String(),
			Params: url.Values{
				"hint": {strconv.Itoa(k)},
			},
		})
	}

	n.web.LoadMainPage(
		strconv.Itoa(n.round.number),
		n.stats.String(),
		strings.Join(n.round.turn.hint, "<br>"),
		n.stats.ActivePlayer().name,
		hints.Join(" "),
	)

	return nil
}

func (n *NonaMegaMego) handleAnswer(params url.Values) error {
	n.round.turn.response = time.Since(n.round.turn.started)

	err := n.s.Pause()
	if err != nil {
		return err
	}

	t, err := n.s.GetCurrentTrack()
	if err != nil {
		return err
	}

	n.web.LoadAnswerPage(
		fmt.Sprintf(`<p>%s</p><img src="%s" class="answer-album-cover">`,
			t.String(), t.Album.ImageURL),
		fmt.Sprintf("%.1f", n.round.turn.response.Seconds()),
	)

	return nil
}
