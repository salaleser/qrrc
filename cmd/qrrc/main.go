package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"qrrc/internal/pkg/router"
	"qrrc/pkg/spotifyhelper"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html, err := ioutil.ReadFile("template/root.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(html)
	})
	http.HandleFunc("/styles", func(w http.ResponseWriter, r *http.Request) {
		css, err := ioutil.ReadFile("static/styles/style.css")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write(css)
	})
	http.HandleFunc("/scripts", func(w http.ResponseWriter, r *http.Request) {
		css, err := ioutil.ReadFile("static/scripts/script.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write(css)
	})
	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		f, err := ioutil.ReadFile(fmt.Sprintf("static/%s", r.URL.Path))
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %v", r.URL.Path, err),
				http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(f)
	})
	http.HandleFunc("/spotify/nonamegamego/", router.NonaMegaMegoHandler)
	http.HandleFunc("/spotify/", spotifyhelper.DefaultHandler)
	http.HandleFunc("/spotify/callback", spotifyhelper.CompleteAuthHandler)

	spotifyhelper.Start()

	err := http.ListenAndServeTLS(":443",
		os.Getenv("QRRC_CERT_PATH"), os.Getenv("QRRC_KEY_PATH"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
