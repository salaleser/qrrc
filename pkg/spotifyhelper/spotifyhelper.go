package spotifyhelper

import (
	"bufio"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"qrrc/internal/pkg/webhelper"
	"time"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

const ErrNoActiveDeviceFound = "Player command failed: No active device found"
const state = "pidor777"

var Instance *SpotifyHelper
var gamePlaylistsImages map[string]string

type SpotifyHelper struct {
	client *spotify.Client
	web    *webhelper.WebHelper
	auth   spotify.Authenticator
	ch     chan *spotify.Client
}

func Start() {
	s := &SpotifyHelper{
		auth: spotify.NewAuthenticator(
			fmt.Sprintf("https://%s/spotify/callback", os.Getenv("QRRC_REDIRECT_HOST")),
			spotify.ScopeUserReadCurrentlyPlaying,
			spotify.ScopeUserReadPlaybackState,
			spotify.ScopeUserModifyPlaybackState,
		),
		ch: make(chan *spotify.Client),
	}

	go s.start()

	Instance = s
}

func (s *SpotifyHelper) start() {
	s.client = <-s.ch

	user, err := s.client.CurrentUser()
	if err != nil {
		fmt.Printf("error: get current user: %v\n", err)
	}
	fmt.Println("Logged in as:", user.ID)

	ps, err := s.client.PlayerState()
	if err != nil {
		fmt.Printf("error: get player state: %v\n", err)
	}
	fmt.Printf("Found your %s (%s)\n", ps.Device.Type, ps.Device.Name)
}

type Track struct {
	ID       spotify.ID
	Title    string
	Duration int
	Album    Album
	Artist   Artist
}

type Album struct {
	ID          spotify.ID
	Title       string
	ReleaseDate string
	ImageURL    string
}

type Artist struct {
	ID    spotify.ID
	Title string
}

type Playlist struct {
	ID       spotify.ID
	Title    string
	ImageURL string
}

func (s *SpotifyHelper) SearchTrack(query string) (Track, error) {
	sr, err := Instance.client.Search(query, spotify.SearchTypeTrack)
	if err != nil {
		return Track{}, errors.Wrap(err, "search track")
	}
	fts := sr.Tracks.Tracks
	if len(fts) == 0 {
		return Track{}, errors.New("no such tracks")
	}
	ft := fts[0]

	return Track{
		ID:       ft.ID,
		Title:    ft.Name,
		Duration: ft.Duration,
		Album: Album{
			ID:          ft.Album.ID,
			Title:       ft.Album.Name,
			ReleaseDate: ft.Album.ReleaseDate,
			ImageURL:    ft.Album.Images[0].URL,
		},
		Artist: Artist{
			ID:    ft.Artists[0].ID,
			Title: ft.Artists[0].Name,
		},
	}, nil
}

func (s *SpotifyHelper) GetCurrentTrack() (Track, error) {
	ps, err := s.client.PlayerState()
	if err != nil {
		return Track{}, errors.Wrap(err, "player state")
	}

	i := ps.CurrentlyPlaying.Item

	return Track{
		ID:       i.ID,
		Title:    i.Name,
		Duration: i.Duration,
		Album: Album{
			ID:          i.Album.ID,
			Title:       i.Album.Name,
			ReleaseDate: i.Album.ReleaseDate,
			ImageURL:    i.Album.Images[0].URL,
		},
		Artist: Artist{
			ID:    i.Artists[0].ID,
			Title: i.Artists[0].Name,
		},
	}, nil
}

func (t *Track) String() string {
	return fmt.Sprintf("%s — %s | %q (%s)", t.Artist.Title,
		t.Title, t.Album.Title, t.Album.ReleaseDate)
}

func (s *SpotifyHelper) Pause() error {
	if err := Instance.client.Pause(); err != nil {
		return errors.Wrap(err, "pause")
	}

	return nil
}

func (s *SpotifyHelper) StartRandomPlaylist(id spotify.ID) error {
	tracks, err := Instance.GetPlaylistTracks(id)
	if err != nil {
		return errors.Wrapf(err, "get playlist tracks (%s)", id)
	}

	rand.Seed(time.Now().UnixNano())
	track := tracks[rand.Intn(len(tracks)-1)]

	if err := Instance.client.QueueSong(track.ID); err != nil {
		return errors.Wrapf(err, "queue song (%s)", track.ID)
	}

	if err := Instance.client.Next(); err != nil {
		return errors.Wrap(err, "next")
	}

	d := track.Duration
	if err := Instance.client.Seek(d/4 + rand.Intn(d/4)); err != nil {
		return errors.Wrapf(err, "seek (%d)", d)
	}

	return nil
}

func (s *SpotifyHelper) SearchPlaylist(playlistTitle string) (Playlist, error) {
	sr, err := Instance.client.Search(playlistTitle, spotify.SearchTypePlaylist)
	if err != nil {
		return Playlist{}, errors.Wrap(err, "search playlist")
	}
	sps := sr.Playlists.Playlists
	if len(sps) == 0 {
		return Playlist{}, errors.New("no such playlists")
	}
	sp := sps[0]

	var imageURL string
	if len(sp.Images) > 0 {
		imageURL = sp.Images[0].URL
	}

	return Playlist{
		ID:       sp.ID,
		Title:    sp.Name,
		ImageURL: imageURL,
	}, nil
}

func (s *SpotifyHelper) GetPlaylistTracks(id spotify.ID) ([]Track, error) {
	ptp, err := Instance.client.GetPlaylistTracks(id)
	if err != nil {
		return []Track{}, err
	}

	result := make([]Track, len(ptp.Tracks))
	for i, v := range ptp.Tracks {
		result[i] = Track{
			ID:       v.Track.ID,
			Title:    v.Track.Name,
			Duration: v.Track.Duration,
		}
	}
	return result, nil
}

func activateFirstDevice() (*spotify.PlayerDevice, error) {
	devices, err := Instance.client.PlayerDevices()
	if err != nil {
		return nil, errors.Wrap(err, "player devices")
	}

	var device *spotify.PlayerDevice
	for i, v := range devices {
		fmt.Printf("%d: %v\n", i, v)
		if v.Active {
			device = &v
			fmt.Printf("%q is activated\n", v.Name)
			return device, nil
		}
	}

	if device == nil {
		return nil, errors.Wrap(err, "no active devices found")
	}

	err = Instance.client.PlayOpt(&spotify.PlayOptions{DeviceID: &device.ID})
	if err != nil {
		return nil, errors.Wrap(err, "play with options")
	}

	return device, nil
}

func handleError(w http.ResponseWriter, err error, message string) {
	if err.Error() == ErrNoActiveDeviceFound {
		device, err := activateFirstDevice()
		if err != nil {
			Instance.web.LoadErrorPage("Спотифай выключен! Попроси хозяина "+
				"запустить его", err)
			return
		}

		if device == nil {
			Instance.web.LoadPage("home", []string{"text", "toggle_play"},
				[]string{"Не удалось автоматически активировать устройство, " +
					"выберите устройство вручную в настройках",
					"<img class=button alt=\"Toggle Play\">"})
			return
		}

		Instance.web.LoadPage("home", []string{"text", "toggle_play"},
			[]string{fmt.Sprintf("Устройство %q (%s) активировано.",
				device.Name, device.Type), "<img class=button alt=\"Toggle Play\">"})
	} else {
		Instance.web.LoadErrorPage("Спотифай выключен! Попроси хозяина "+
			"запустить его", err)
		return
	}
}

func initGamePlaylistsImagesCache() error {
	if gamePlaylistsImages != nil {
		return nil
	}

	playlistsFile, err := os.Open("playlists.txt")
	if err != nil {
		return errors.Wrap(err, "open playlists file")
	}
	defer playlistsFile.Close()
	scanner := bufio.NewScanner(playlistsFile)
	playlists := make([]string, 0)
	for scanner.Scan() {
		playlists = append(playlists, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return errors.Wrap(err, "scan")
	}

	gamePlaylistsImages = make(map[string]string)
	for _, v := range playlists {
		p, err := Instance.SearchPlaylist(v)
		if err != nil {
			fmt.Printf("error: search playlist (%s): %v\n", v, err)
			continue
		}
		gamePlaylistsImages[v] = p.ImageURL
	}

	return nil
}
