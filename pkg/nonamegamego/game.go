package nonamegamego

import (
	"fmt"
	"net/url"
	"qrrc/internal/pkg/spotifyhelper"
	"qrrc/internal/pkg/webhelper"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Game struct {
	Rooms map[RoomID]*Room
	w     *webhelper.WebHelper
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
	hints            map[int]hint
	started          time.Time
	response         time.Duration
	hintTrackCursor  int
	hintArtistCursor int
}

func New() *Game {
	g := &Game{
		Rooms: make(map[RoomID]*Room),
	}

	handlers = map[string]handler{
		"setup":  g.setupHandler,
		"answer": g.answerHandler,
		"lobby":  g.lobbyHandler,
		"room":   g.roomHandler,
	}

	return g
}

func (g *Game) RoomsList() []string {
	rooms := make([]string, len(g.Rooms))
	i := 0
	for _, v := range g.Rooms {
		rooms[i] = v.String()
		i++
	}
	sort.Strings(rooms)
	return rooms
}

func (g *Game) SetWebHelper(w *webhelper.WebHelper) {
	g.w = w
}

func (g *Game) lobbyHandler(params url.Values) error {
	buttons := make(Buttons, len(g.Rooms))
	for i, v := range g.RoomsList() {
		buttons[i] = Button{
			Link: "room",
			Params: url.Values{
				"room":   {v},
				"status": {"continue"},
			},
			Text:  v,
			Class: "room",
		}
	}

	g.w.LoadLobbyPage(
		buttons.Join("<br>"),
	)

	return nil
}

func (g *Game) setupHandler(params url.Values) error {
	roomID := params.Get("room")
	clientID := params.Get("client_id")
	clientSecret := params.Get("client_secret")
	playerNamesParam := params.Get("player_names")
	playerNamesParamArray := strings.Split(playerNamesParam, ",")

	if clientID != "" && clientSecret != "" {
		s := spotifyhelper.New(roomID, clientID, clientSecret)
		go s.Auth()
		room := &Room{id: RoomID(roomID), s: s, settings: settings{}}
		g.Rooms[RoomID(roomID)] = room
		g.w.Redirect(room.s.AuthURL())
		return nil
	}

	room, ok := g.Rooms[RoomID(roomID)]
	if !ok {
		return fmt.Errorf("комната %q не найдена", roomID)
	}

	var err error
	playlistParam := params.Get("playlist")
	if playlistParam != "" {
		room.settings.playlist, err = room.s.SearchPlaylist(playlistParam)
		if err != nil {
			return errors.Wrap(err, "search playlist")
		}
	} else {
		room.settings.playlist = nil
	}

	playersCount := 5
	playerNames := make(Fields, playersCount)
	for i := 0; i < playersCount; i++ {
		playerNames[i] = Field{
			Text: func() string {
				if len(playerNamesParamArray) > i {
					return playerNamesParamArray[i]
				}
				return ""
			}(),
		}
	}

	g.w.LoadSetupPage(
		roomID,
		func() string {
			if room.s.HasClient() {
				return `<p class="success">Пройдена</p>`
			} else {
				return `<p class="fail">Не пройдена</p>`
			}
		}(),
		room.settings.String(),
		playerNames.Join("<br>"),
	)

	return nil
}

func (g *Game) roomHandler(params url.Values) error {
	roomID := params.Get("room")
	room, ok := g.Rooms[RoomID(roomID)]
	if !ok {
		return fmt.Errorf("комната %q не найдена", roomID)
	}

	statusParam := params.Get("status")
	hintParam := params.Get("hint")

	var err error
	var t spotifyhelper.Track
	if statusParam == "start" {
		playerNamesParam := params.Get("player_names")
		playerNames := strings.Split(playerNamesParam, ",")
		room.stats = NewStats(playerNames)
		room.round = round{
			number: 1,
			turn: turn{
				hint:    []string{},
				hints:   room.updateHints(),
				started: time.Now(),
			},
		}
		if room.settings.playlist != nil {
			err := room.s.PlayRandomTrack(room.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := room.s.PlayNextTrack()
			if err != nil {
				return errors.Wrap(err, "play next track")
			}
		}
	} else if statusParam == "continue" {
		if room.stats == nil {
			delete(g.Rooms, room.id)
			g.w.Redirect("/spotify/nonamegamego/lobby")
			return nil
		}
	} else if statusParam == "delete" {
		delete(g.Rooms, room.id)
		g.w.Redirect("/spotify/nonamegamego/lobby")
		return nil
	} else if hintParam != "" {
		hintID, err := strconv.Atoi(hintParam)
		if err != nil {
			return errors.Wrap(err, "parse hint")
		}

		if hint, ok := room.round.turn.hints[hintID]; ok {
			delete(room.round.turn.hints, hintID)
			room.round.turn.hint = append(room.round.turn.hint,
				fmt.Sprintf("%s: <b>%s</b>", hint.text, hint.f()))
			room.stats.ActivePlayer().AddScore(-float64(hint.value))
		}
	} else {
		if room.settings.playlist != nil {
			err := room.s.PlayRandomTrack(room.settings.playlist)
			if err != nil {
				return errors.Wrap(err, "start random playlist")
			}
		} else {
			err := room.s.PlayNextTrack()
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
		if score > 0 {
			if room.round.turn.response < 5*time.Second {
				multiplier = 1.2
			} else if room.round.turn.response > 90*time.Second {
				multiplier = 0.8
			}
		}

		room.stats.ActivePlayer().AddScore(multiplier * score)

		room.round.turn = turn{
			hint:    []string{},
			hints:   room.updateHints(),
			started: time.Now(),
		}

		if room.stats.SetActiveNext() {
			room.round.number = room.round.number + 1
		}
	}

	hints := make(Buttons, 0)
	for k, v := range room.round.turn.hints {
		hints = append(hints, Button{
			Link: "room",
			Text: v.String(),
			Params: url.Values{
				"room": {roomID},
				"hint": {strconv.Itoa(k)},
			},
		})
	}

	t, err = room.s.CurrentTrack()
	if err != nil {
		return errors.Wrap(err, "current track")
	}

	track := ""

	artist := ""

	g.w.LoadRoomPage(
		roomID,
		strconv.Itoa(room.round.number),
		strconv.Itoa(t.Duration-t.Progress),
		room.stats.String(),
		strings.Join(room.round.turn.hint, "<br>"),
		track,
		artist,
		room.stats.ActivePlayer().name,
		hints.Join(" "),
	)

	return nil
}

func (g *Game) answerHandler(params url.Values) error {
	roomID := params.Get("room")
	room, ok := g.Rooms[RoomID(roomID)]
	if !ok {
		return fmt.Errorf("комната %q не найдена", roomID)
	}

	room.round.turn.response = time.Since(room.round.turn.started)

	err := room.s.Pause()
	if err != nil {
		return err
	}

	t, err := room.s.CurrentTrack()
	if err != nil {
		return errors.Wrap(err, "current track")
	}

	g.w.LoadAnswerPage(
		roomID,
		fmt.Sprintf(`<p>%s</p><img src="%s" class="answer-album-cover">`,
			t.String(), t.Album.ImageURL),
		fmt.Sprintf("%.1f", room.round.turn.response.Seconds()),
	)

	return nil
}
