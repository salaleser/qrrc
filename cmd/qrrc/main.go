package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/nonamegamego"
	"strings"
)

var game *nonamegamego.Game

func main() {
	game = nonamegamego.New()

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		html, err := ioutil.ReadFile("template/index.html")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "text/html")
		_, _ = rw.Write(html)
	})
	http.HandleFunc("/styles", func(rw http.ResponseWriter, r *http.Request) {
		css, err := ioutil.ReadFile("static/styles/style.css")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "text/css")
		_, _ = rw.Write(css)
	})
	http.HandleFunc("/scripts", func(rw http.ResponseWriter, r *http.Request) {
		css, err := ioutil.ReadFile("static/scripts/script.js")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "text/css")
		_, _ = rw.Write(css)
	})
	http.HandleFunc("/images/", func(rw http.ResponseWriter, r *http.Request) {
		f, err := ioutil.ReadFile(fmt.Sprintf("static/%s", r.URL.Path))
		if err != nil {
			http.Error(rw, fmt.Sprintf("%s: %v", r.URL.Path, err),
				http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "image/png")
		_, _ = rw.Write(f)
	})
	http.HandleFunc("/callback/", func(rw http.ResponseWriter, r *http.Request) {
		w := webhelper.New(rw, r)
		game.SetWebHelper(w)
		roomID := strings.TrimPrefix(r.URL.Path, "/callback/")
		room, ok := game.Rooms[nonamegamego.RoomID(roomID)]
		if !ok {
			w.LoadErrorPage(fmt.Sprintf("Комната %q не найдена", roomID),
				fmt.Errorf(""))
			return
		}

		err := room.CompleteAuth(r)
		if err != nil {
			w.LoadErrorPage("complete auth", err)
			return
		}

		err = game.Route("setup", url.Values{"room": {roomID}})
		if err != nil {
			w.LoadErrorPage("setup", err)
			return
		}
	})
	http.HandleFunc("/spotify/nonamegamego/", func(rw http.ResponseWriter, r *http.Request) {
		w := webhelper.New(rw, r)
		game.SetWebHelper(w)
		params := r.URL.Query()
		action := strings.TrimPrefix(r.URL.Path, "/spotify/nonamegamego/")
		err := game.Route(action, params)
		if err != nil {
			w.LoadErrorPage(action, err)
			return
		}
	})

	err := http.ListenAndServeTLS(":443",
		os.Getenv("QRRC_CERT_PATH"), os.Getenv("QRRC_KEY_PATH"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
