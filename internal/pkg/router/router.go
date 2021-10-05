package router

import (
	"fmt"
	"net/http"
	"qrrc/internal/pkg/webhelper"
	"qrrc/pkg/nonamegamego"
	"strings"
)

func NonaMegaMegoHandler(w http.ResponseWriter, r *http.Request) {
	web := webhelper.New(w)
	n := nonamegamego.New(web)

	action := strings.TrimPrefix(r.URL.Path, "/spotify/nonamegamego/")
	params := r.URL.Query()

	if err := n.Route(action, params); err != nil {
		web.LoadErrorPage(fmt.Sprintf("handle %q: %v", action, err))
		return
	}
}
