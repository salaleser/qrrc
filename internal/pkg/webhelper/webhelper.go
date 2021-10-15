package webhelper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type WebHelper struct {
	rw http.ResponseWriter
	r  *http.Request
}

func New(rw http.ResponseWriter, r *http.Request) *WebHelper {
	return &WebHelper{
		rw: rw,
		r:  r,
	}
}

func (w *WebHelper) Redirect(url string) {
	http.Redirect(w.rw, w.r, url, http.StatusFound)
}

func (w *WebHelper) LoadErrorPage(
	text string,
	err error,
) {
	w.LoadPage(
		"error",
		[]string{
			"text",
			"error",
		},
		[]string{
			text,
			err.Error(),
		},
	)
}

func (w *WebHelper) LoadLobbyPage(
	rooms string,
) {
	w.LoadPage(
		"nonamegamego/lobby",
		[]string{
			"rooms",
		},
		[]string{
			rooms,
		},
	)
}

func (w *WebHelper) LoadRoomPage(
	roomID string,
	round string,
	stats string,
	hint string,
	text string,
	hints string,
) {
	w.LoadPage(
		"nonamegamego/room",
		[]string{
			"room-id",
			"round",
			"stats",
			"hint",
			"text",
			"hints",
		},
		[]string{
			roomID,
			round,
			stats,
			hint,
			text,
			hints,
		},
	)
}

func (w *WebHelper) LoadAnswerPage(
	roomID string,
	text string,
	response string,
) {
	w.LoadPage(
		"nonamegamego/answer",
		[]string{
			"room-id",
			"text",
			"response",
		},
		[]string{
			roomID,
			text,
			response,
		},
	)
}

func (w *WebHelper) LoadSetupPage(
	roomID string,
	status string,
	settingsPlaylist string,
	playerNames string,
) {
	w.LoadPage(
		"nonamegamego/setup",
		[]string{
			"room-id",
			"status",
			"settings-playlist",
			"player-names",
		},
		[]string{
			roomID,
			status,
			settingsPlaylist,
			playerNames,
		},
	)
}

func (w *WebHelper) LoadPage(p string, old []string, new []string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("template/%s.html", p))
	if err != nil {
		http.Error(w.rw, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: read file: %v", err)
	}
	w.rw.Header().Set("Content-Type", "text/html")
	for i := 0; i < len(old); i++ {
		html = []byte(strings.Replace(
			string(html), "{{"+old[i]+"}}", new[i], -1,
		))
	}
	_, err = w.rw.Write(html)
	if err != nil {
		http.Error(w.rw, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: %v", err)
	}
}
