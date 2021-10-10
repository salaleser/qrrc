package router

import (
	"net/http"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/nonamegamego"
	"qrrc/pkg/spotifyhelper"
	"strings"
)

var instance *nonamegamego.NonaMegaMego

func NonaMegaMegoHandler(w http.ResponseWriter, r *http.Request) {
	web := webhelper.New(w)

	if instance == nil {
		instance = nonamegamego.New(web, spotifyhelper.Instance)
	} else {
		instance.Update(web)
	}

	action := strings.TrimPrefix(r.URL.Path, "/spotify/nonamegamego/")
	params := r.URL.Query()

	err := instance.Route(action, params)
	if err != nil {
		web.LoadErrorPage(action, err)
		return
	}
}
