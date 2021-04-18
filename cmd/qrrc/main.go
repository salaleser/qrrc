package main

import (
	"log"
	"net/http"
	"os"
	"qrrc/pkg/spotifyapi"
)

func main() {
	http.HandleFunc("/spotify/", spotifyapi.DefaultHandler)
	http.HandleFunc("/spotify/callback", spotifyapi.CompleteAuthHandler)

	spotifyapi.Start()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatal("No port provided")
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
