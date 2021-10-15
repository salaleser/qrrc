package nonamegamego

import (
	"net/http"
	"qrrc/internal/pkg/spotifyhelper"

	"github.com/pkg/errors"
)

type RoomID string

type Room struct {
	id       RoomID
	s        *spotifyhelper.SpotifyHelper
	stats    *Stats
	settings settings
	round    round
}

func (r *Room) String() string {
	return string(r.id)
}

func (r *Room) CompleteAuth(req *http.Request) error {
	err := r.s.CompleteAuth(req)
	if err != nil {
		return errors.Wrap(err, "complete auth")
	}
	return nil
}
