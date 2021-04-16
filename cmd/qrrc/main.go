package main

import (
	"log"
	"net/http"
	"qrrc/pkg/spotifyapi"
)

func main() {
	http.HandleFunc("/callback", spotifyapi.CompleteAuthHandler)
	http.HandleFunc("/spotify/", spotifyapi.DefaultHandler)

	spotifyapi.Start()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
