package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"qrrc/pkg/spotifyapi"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html, err := ioutil.ReadFile("template/root.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error: root handler: read file: %v", err)
		}
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write(html)
	})
	http.HandleFunc("/spotify/", spotifyapi.DefaultHandler)
	http.HandleFunc("/spotify/callback", spotifyapi.CompleteAuthHandler)

	spotifyapi.Start()

	err := http.ListenAndServeTLS(":443",
		os.Getenv("QRRC_CERT_PATH"), os.Getenv("QRRC_KEY_PATH"), nil)
	if err != nil {
		log.Fatal(err)
	}
}
