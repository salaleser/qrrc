package spotifyhelper

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"qrrc/internal/pkg/webhelper"
	"strings"
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

	u, err := s.client.CurrentUser()
	if err != nil {
		fmt.Printf("error: get current user: %v\n", err)
	}
	fmt.Println("Logged in as:", u.ID)

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

func (p *Playlist) String() string {
	if p == nil {
		return "(текущий плейлист)"
	}
	return fmt.Sprintf("%q", p.Title)
}

type Device struct {
	ID     spotify.ID
	Title  string
	Active bool
	Type   string
}

type Devices []Device

func (d *Device) String() string {
	if d.Active {
		return fmt.Sprintf("%s (%s) *", d.Title, d.Type)
	}
	return fmt.Sprintf("<a href=/spotify/settings?device_id=%s>%s (%s)",
		d.ID, d.Title, d.Type)
}

func (d *Devices) Join(sep string) string {
	var builder strings.Builder
	for _, v := range *d {
		builder.WriteString(v.String())
		builder.WriteString(sep)
	}
	result := builder.String()
	return strings.TrimRight(result, sep)
}

func (s *SpotifyHelper) GetDevices() (*Devices, error) {
	d, err := Instance.client.PlayerDevices()
	if err != nil {
		return nil, errors.Wrap(err, "player devices")
	}

	devices := make(Devices, len(d))
	for i, v := range d {
		devices[i] = Device{
			ID:     v.ID,
			Title:  v.Name,
			Active: v.Active,
			Type:   v.Type,
		}
	}
	return &devices, nil
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
	return fmt.Sprintf("<b>%s — %s</b><br>%q (%s)", t.Artist.Title,
		t.Title, t.Album.Title, t.Album.ReleaseDate)
}

func (s *SpotifyHelper) Pause() error {
	if err := Instance.client.Pause(); err != nil {
		return errors.Wrap(err, "pause")
	}

	return nil
}

func (s *SpotifyHelper) PlayRandomTrack(p *Playlist) error {
	tracks, err := Instance.GetPlaylistTracks(p.ID)
	if err != nil {
		return errors.Wrapf(err, "get playlist tracks (%s)", p.ID)
	}

	rand.Seed(time.Now().UnixNano())
	track := tracks[rand.Intn(len(tracks)-1)]

	err = Instance.client.QueueSong(track.ID)
	if err != nil {
		return errors.Wrapf(err, "queue song (%s)", track.ID)
	}

	err = Instance.client.Next()
	if err != nil {
		return errors.Wrap(err, "next")
	}

	d := track.Duration
	err = Instance.client.Seek(d/4 + rand.Intn(d/4))
	if err != nil {
		return errors.Wrapf(err, "seek (%d)", d)
	}

	return nil
}

func (s *SpotifyHelper) PlayNextTrack() error {
	err := Instance.client.Next()
	if err != nil {
		return errors.Wrap(err, "next")
	}

	track, err := s.GetCurrentTrack()
	if err != nil {
		return errors.Wrap(err, "get current track")
	}

	d := track.Duration
	rand.Seed(time.Now().UnixNano())
	err = Instance.client.Seek(d/4 + rand.Intn(d/4))
	if err != nil {
		return errors.Wrapf(err, "seek (%d)", d)
	}

	return nil
}

func (s *SpotifyHelper) SearchPlaylist(playlistTitle string) (*Playlist, error) {
	sr, err := Instance.client.Search(playlistTitle, spotify.SearchTypePlaylist)
	if err != nil {
		return nil, errors.Wrap(err, "search playlist")
	}
	sps := sr.Playlists.Playlists
	if len(sps) == 0 {
		return nil, errors.New("no such playlists")
	}
	sp := sps[0]

	var imageURL string
	if len(sp.Images) > 0 {
		imageURL = sp.Images[0].URL
	}

	return &Playlist{
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
