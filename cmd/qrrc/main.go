package main

import (
	"log"
	"net/http"
	"os"
	"qrrc/pkg/spotifyapi"
)

func main() {
	http.HandleFunc("/callback", spotifyapi.CompleteAuthHandler)
	http.HandleFunc("/spotify/", spotifyapi.DefaultHandler)

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
