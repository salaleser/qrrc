package webhelper

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type WebHelper struct {
	w http.ResponseWriter
}

func New(w http.ResponseWriter) *WebHelper {
	return &WebHelper{
		w: w,
	}
}

func (h *WebHelper) LoadErrorPage(text string, err error) {
	h.LoadPage(
		"error",
		[]string{"text", "error"},
		[]string{text, err.Error()},
	)
}

func (h *WebHelper) LoadSettingsPage(
	devices string,
) {
	h.LoadPage(
		"settings",
		[]string{
			"devices",
		},
		[]string{
			devices,
		},
	)
}

func (h *WebHelper) LoadMainPage(round, stats, hint, text, buttons string) {
	h.LoadPage(
		"nonamegamego/main",
		[]string{"round", "stats", "hint", "text", "buttons"},
		[]string{round, stats, hint, text, buttons},
	)
}

func (h *WebHelper) LoadStartPage() {
	h.LoadPage(
		"nonamegamego/start",
		[]string{},
		[]string{},
	)
}

func (h *WebHelper) LoadAnswerPage(text, buttons string) {
	h.LoadPage(
		"nonamegamego/answer",
		[]string{"text", "buttons"},
		[]string{text, buttons},
	)
}

func (h *WebHelper) LoadSetupPage(
	text string,
	playersCount string,
	playerNames string,
	playlists string,
) {
	h.LoadPage(
		"nonamegamego/setup",
		[]string{
			"text",
			"players-count",
			"player-names",
			"playlists",
		},
		[]string{
			text,
			playersCount,
			playerNames,
			playlists,
		},
	)
}

func (web *WebHelper) LoadPage(p string, old []string, new []string) {
	html, err := ioutil.ReadFile(fmt.Sprintf("template/%s.html", p))
	if err != nil {
		http.Error(web.w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: read file: %v", err)
	}
	web.w.Header().Set("Content-Type", "text/html")
	for i := 0; i < len(old); i++ {
		html = []byte(strings.Replace(
			string(html), "{{"+old[i]+"}}", new[i], -1,
		))
	}
	_, err = web.w.Write(html)
	if err != nil {
		http.Error(web.w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error: load page replace: %v", err)
	}
}
