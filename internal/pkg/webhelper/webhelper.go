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

func (h *WebHelper) LoadMainPage(round, stats, hint, text, hints string) {
	h.LoadPage(
		"nonamegamego/main",
		[]string{"round", "stats", "hint", "text", "hints"},
		[]string{round, stats, hint, text, hints},
	)
}

func (h *WebHelper) LoadAnswerPage(
	text string,
	response string,
) {
	h.LoadPage(
		"nonamegamego/answer",
		[]string{
			"text",
			"response",
		},
		[]string{
			text,
			response,
		},
	)
}

func (h *WebHelper) LoadSetupPage(
	settingsPlaylist string,
	playerNames string,
) {
	h.LoadPage(
		"nonamegamego/setup",
		[]string{
			"settings-playlist",
			"player-names",
		},
		[]string{
			settingsPlaylist,
			playerNames,
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
