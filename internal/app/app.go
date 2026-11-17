// Package app is the Wails service: the single surface the React app talks to.
// It orchestrates the services, applies user-facing rules and converts domain
// types into the DTOs the UI understands.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/brand"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/lyrics"
	"github.com/LuiisDev21/stefani/internal/media"
	"github.com/LuiisDev21/stefani/internal/music"
	"github.com/LuiisDev21/stefani/internal/phonetics"
	"github.com/LuiisDev21/stefani/internal/secret"
	"github.com/LuiisDev21/stefani/internal/storage"
)

// Emitter publishes events to the frontend.
type Emitter interface {
	Emit(name string, data ...any) bool
}

// FilePicker opens native dialogs. It is an interface so the app layer can be
// tested without a window.
type FilePicker interface {
	PickAudioFiles() ([]string, error)
	PickLyricsFile() (string, error)
}

// Opener opens links in the system browser.
type Opener interface {
	OpenURL(url string) error
}

// Deps are the collaborators the service needs. Everything is injected so the
// service stays testable.
type Deps struct {
	DB        *storage.DB
	Configs   *config.Store
	Config    config.Config
	Secrets   secret.Store
	Music     *music.Service
	Lyrics    *lyrics.Service
	Phonetics *phonetics.Service
	AI        *ai.Client
	Log       *slog.Logger
	Picker    FilePicker
	Emitter   Emitter
	Opener    Opener
	// OnConfigChange runs after settings are saved, so services holding their
	// own snapshot can refresh.
	OnConfigChange func(config.Config)
}

// App is the bound service.
type App struct {
	db      *storage.DB
	configs *config.Store
	secrets secret.Store
	music   *music.Service
	lyrics  *lyrics.Service
	phon    *phonetics.Service
	ai      *ai.Client
	log     *slog.Logger
	picker  FilePicker
	emitter Emitter
	opener  Opener

	onConfigChange func(config.Config)

	mu  sync.RWMutex
	cfg config.Config
}

// New builds the service.
func New(d Deps) *App {
	log := d.Log
	if log == nil {
		log = slog.Default()
	}
	return &App{
		db:      d.DB,
		configs: d.Configs,
		secrets: d.Secrets,
		music:   d.Music,
		lyrics:  d.Lyrics,
		phon:    d.Phonetics,
		ai:      d.AI,
		log:     log,
		picker:  d.Picker,
		emitter: d.Emitter,
		opener:  d.Opener,
		cfg:     d.Config,

		onConfigChange: d.OnConfigChange,
	}
}

func (a *App) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 45*time.Second)
}

func (a *App) currentConfig() config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

func (a *App) emit(name string, data ...any) {
	if a.emitter != nil {
		a.emitter.Emit(name, data...)
	}
}

// Info returns application metadata for the About panel.
func (a *App) Info() InfoDTO {
	ctx, cancel := a.opCtx()
	defer cancel()
	schema, _ := a.db.SchemaVersion(ctx)
	return InfoDTO{
		Name:       brand.Name,
		Tagline:    brand.Tagline,
		Version:    brand.Version,
		Identifier: brand.Identifier,
		DBPath:     a.db.Path(),
		Schema:     schema,
	}
}

// Bootstrap loads everything the shell needs on start.
func (a *App) Bootstrap() (BootstrapDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()

	demoTracks, err := a.tracksOfProvider(ctx, domain.ProviderDemo)
	if err != nil {
		return BootstrapDTO{}, err
	}
	library, err := a.library(ctx)
	if err != nil {
		return BootstrapDTO{}, err
	}
	favorites, err := a.favorites(ctx)
	if err != nil {
		return BootstrapDTO{}, err
	}
	recentIDs, err := a.db.RecentlyPlayed(ctx, 8)
	if err != nil {
		return BootstrapDTO{}, err
	}
	recent, err := a.tracksByIDs(ctx, recentIDs)
	if err != nil {
		return BootstrapDTO{}, err
	}
	playlists, err := a.playlists(ctx)
	if err != nil {
		return BootstrapDTO{}, err
	}
	progress, err := a.Progress()
	if err != nil {
		return BootstrapDTO{}, err
	}
	vocabulary, err := a.Vocabulary()
	if err != nil {
		return BootstrapDTO{}, err
	}
	settings, err := a.Settings()
	if err != nil {
		return BootstrapDTO{}, err
	}
	counts, err := a.counts(ctx)
	if err != nil {
		return BootstrapDTO{}, err
	}

	return BootstrapDTO{
		App:            a.Info(),
		DemoTracks:     demoTracks,
		RecentlyPlayed: recent,
		Library:        library,
		Favorites:      favorites,
		Playlists:      playlists,
		Progress:       progress,
		Vocabulary:     vocabulary,
		Settings:       settings,
		Providers:      a.providerStates(ctx),
		Counts:         counts,
	}, nil
}

func (a *App) counts(ctx context.Context) (StoredCountsDTO, error) {
	lib, err := a.db.CountLibrary(ctx)
	if err != nil {
		return StoredCountsDTO{}, err
	}
	fav, err := a.db.CountFavorites(ctx)
	if err != nil {
		return StoredCountsDTO{}, err
	}
	all, err := a.db.AllTracks(ctx)
	if err != nil {
		return StoredCountsDTO{}, err
	}
	playlists, err := a.db.Playlists(ctx)
	if err != nil {
		return StoredCountsDTO{}, err
	}
	words, err := a.db.Vocabulary(ctx)
	if err != nil {
		return StoredCountsDTO{}, err
	}
	annotated := 0
	for _, track := range all {
		items, err := a.db.AnnotationsByTrack(ctx, track.ID)
		if err != nil {
			continue
		}
		annotated += len(items)
	}
	return StoredCountsDTO{
		Tracks:        len(all),
		LibraryTracks: lib,
		Favorites:     fav,
		Playlists:     len(playlists),
		Vocabulary:    len(words),
		Annotated:     annotated,
	}, nil
}

func (a *App) providerStates(ctx context.Context) []ProviderStateDTO {
	cfg := a.currentConfig()
	ytState := "disabled"
	ytDetail := "Activa el catalogo experimental en Ajustes"
	if cfg.Features.YouTubeMusic {
		ytState = "enabled"
		ytDetail = "Busqueda de catalogo activa. La reproduccion queda en la app oficial."
	}
	return []ProviderStateDTO{
		{Name: "Biblioteca local", State: "ready", Detail: "Archivos de audio importados por ti"},
		{Name: "Demo", State: "ready", Detail: "Canciones originales sintetizadas con tiempos por palabra"},
		{Name: "LRCLIB", State: "ready", Detail: "Letras sincronizadas de la base comunitaria"},
		{Name: "YouTube Music", State: ytState, Detail: ytDetail},
	}
}

// ---- tracks ----

func (a *App) tracksOfProvider(ctx context.Context, provider domain.Provider) ([]TrackDTO, error) {
	tracks, err := a.db.TracksByProvider(ctx, provider)
	if err != nil {
		return nil, err
	}
	return a.toTrackList(ctx, tracks), nil
}

func (a *App) tracksByIDs(ctx context.Context, ids []string) ([]TrackDTO, error) {
	tracks, err := a.db.TracksByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return a.toTrackList(ctx, tracks), nil
}

func (a *App) toTrackList(ctx context.Context, tracks []domain.Track) []TrackDTO {
	out := make([]TrackDTO, 0, len(tracks))
	for _, t := range tracks {
		fav, inLibrary := a.db.IsFavorite(ctx, t.ID)
		out = append(out, toTrackDTO(t, inLibrary, fav))
	}
	return out
}

// Track returns one track by id.
func (a *App) Track(id string) (TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	track, err := a.db.Track(ctx, id)
	if err != nil {
		return TrackDTO{}, fmt.Errorf("no encontre esa cancion")
	}
	fav, inLibrary := a.db.IsFavorite(ctx, id)
	return toTrackDTO(track, inLibrary, fav), nil
}

// Search queries the configured providers.
func (a *App) Search(query string, limit int) (SearchDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	query = strings.TrimSpace(query)
	if query == "" {
		return SearchDTO{Query: query, Providers: a.providerStates(ctx)}, nil
	}
	results, err := a.music.Search(ctx, query, limit)
	if err != nil {
		return SearchDTO{}, err
	}
	items := make([]TrackDTO, 0, len(results))
	for _, r := range results {
		fav, inLibrary := false, false
		if existing, err := a.db.TrackByProviderID(ctx, r.Track.Provider, r.Track.ProviderID); err == nil {
			fav, inLibrary = a.db.IsFavorite(ctx, existing.ID)
			r.Track = existing
		}
		dto := toTrackDTO(r.Track, inLibrary, fav)
		dto.Playable = r.Playable
		dto.Note = r.Note
		items = append(items, dto)
	}
	cfg := a.currentConfig()
	notice := ""
	if !cfg.Features.YouTubeMusic {
		notice = "Buscando solo en tu biblioteca. Activa YouTube Music en Ajustes para buscar en el catalogo."
	}
	return SearchDTO{Query: query, Items: items, Providers: a.providerStates(ctx), Notice: notice}, nil
}

// ImportLocalTracks opens a native picker and imports the chosen files.
func (a *App) ImportLocalTracks() (ImportDTO, error) {
	if a.picker == nil {
		return ImportDTO{}, fmt.Errorf("el selector de archivos no esta disponible en este entorno")
	}
	paths, err := a.picker.PickAudioFiles()
	if err != nil {
		return ImportDTO{}, fmt.Errorf("no pude abrir el selector de archivos: %w", err)
	}
	return a.importPaths(paths)
}

// ImportDroppedFiles imports files dropped onto the window.
func (a *App) ImportDroppedFiles(paths []string) (ImportDTO, error) {
	return a.importPaths(paths)
}

func (a *App) importPaths(paths []string) (ImportDTO, error) {
	if len(paths) == 0 {
		return ImportDTO{Notice: "No seleccionaste archivos"}, nil
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	tracks, skipped, err := a.music.ImportFiles(ctx, paths)
	if err != nil {
		return ImportDTO{}, err
	}
	out := ImportDTO{
		Imported: a.toTrackList(ctx, tracks),
		Skipped:  skipped,
	}
	if len(tracks) > 0 {
		out.Notice = fmt.Sprintf("%d archivo(s) importados", len(tracks))
	}
	if len(skipped) > 0 {
		out.Notice += fmt.Sprintf(", %d omitidos (formato no soportado o ilegible)", len(skipped))
	}
	return out, nil
}

// SetTrackDuration stores the duration reported by the audio element.
func (a *App) SetTrackDuration(trackID string, durationMs int64) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.music.UpdateDuration(ctx, trackID, durationMs)
}

// ---- library ----

func (a *App) library(ctx context.Context) ([]TrackDTO, error) {
	ids, err := a.db.LibraryTrackIDs(ctx)
	if err != nil {
		return nil, err
	}
	return a.tracksByIDs(ctx, ids)
}

func (a *App) favorites(ctx context.Context) ([]TrackDTO, error) {
	ids, err := a.db.Favorites(ctx)
	if err != nil {
		return nil, err
	}
	return a.tracksByIDs(ctx, ids)
}

// Library lists saved tracks.
func (a *App) Library() ([]TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.library(ctx)
}

// Favorites lists favourite tracks.
func (a *App) Favorites() ([]TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.favorites(ctx)
}

// SetInLibrary saves or removes a track from the library.
func (a *App) SetInLibrary(trackID string, save bool) (TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	if save {
		if err := a.db.SaveToLibrary(ctx, trackID); err != nil {
			return TrackDTO{}, err
		}
	} else {
		if err := a.db.RemoveFromLibrary(ctx, trackID); err != nil {
			return TrackDTO{}, err
		}
	}
	track, err := a.db.Track(ctx, trackID)
	if err != nil {
		return TrackDTO{}, fmt.Errorf("no encontre esa cancion")
	}
	fav, inLibrary := a.db.IsFavorite(ctx, trackID)
	return toTrackDTO(track, inLibrary, fav), nil
}

// SetFavorite flags or unflags a track.
func (a *App) SetFavorite(trackID string, favorite bool) (TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	if err := a.db.SetFavorite(ctx, trackID, favorite); err != nil {
		return TrackDTO{}, err
	}
	track, err := a.db.Track(ctx, trackID)
	if err != nil {
		return TrackDTO{}, fmt.Errorf("no encontre esa cancion")
	}
	fav, inLibrary := a.db.IsFavorite(ctx, trackID)
	return toTrackDTO(track, inLibrary, fav), nil
}

// ---- playlists ----

func (a *App) playlists(ctx context.Context) ([]PlaylistDTO, error) {
	list, err := a.db.Playlists(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PlaylistDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toPlaylistDTO(p))
	}
	return out, nil
}

// Playlists lists every playlist.
func (a *App) Playlists() ([]PlaylistDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.playlists(ctx)
}

// Playlist returns one playlist.
func (a *App) Playlist(id string) (PlaylistDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	p, err := a.db.Playlist(ctx, id)
	if err != nil {
		return PlaylistDTO{}, fmt.Errorf("no encontre esa lista")
	}
	return toPlaylistDTO(p), nil
}

// PlaylistTracks lists a playlist's tracks in order.
func (a *App) PlaylistTracks(id string) ([]TrackDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	items, err := a.db.PlaylistItems(ctx, id)
	if err != nil {
		return nil, err
	}
	tracks := make([]domain.Track, 0, len(items))
	for _, item := range items {
		tracks = append(tracks, item.Track)
	}
	return a.toTrackList(ctx, tracks), nil
}

// CreatePlaylist creates an empty playlist.
func (a *App) CreatePlaylist(name string) (PlaylistDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	name = strings.TrimSpace(name)
	if name == "" {
		return PlaylistDTO{}, fmt.Errorf("la lista necesita un nombre")
	}
	p, err := a.db.CreatePlaylist(ctx, domain.Playlist{Name: name, Origin: "local"})
	if err != nil {
		return PlaylistDTO{}, err
	}
	return toPlaylistDTO(p), nil
}

// UpdatePlaylist renames or re-describes a playlist.
func (a *App) UpdatePlaylist(id, name, description string) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("la lista necesita un nombre")
	}
	return a.db.UpdatePlaylist(ctx, id, strings.TrimSpace(name), strings.TrimSpace(description))
}

// DeletePlaylist removes a playlist.
func (a *App) DeletePlaylist(id string) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.db.DeletePlaylist(ctx, id)
}

// AddToPlaylist appends a track to a playlist.
func (a *App) AddToPlaylist(playlistID, trackID string) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if err := a.db.AddToPlaylist(ctx, playlistID, trackID); err != nil {
		return err
	}
	a.emit("playlist:changed", playlistID)
	return nil
}

// RemoveFromPlaylist removes the item at a position.
func (a *App) RemoveFromPlaylist(playlistID string, position int) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if err := a.db.RemovePlaylistItem(ctx, playlistID, position); err != nil {
		return err
	}
	a.emit("playlist:changed", playlistID)
	return nil
}

// MovePlaylistItem reorders a playlist.
func (a *App) MovePlaylistItem(playlistID string, from, to int) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if err := a.db.MovePlaylistItem(ctx, playlistID, from, to); err != nil {
		return err
	}
	a.emit("playlist:changed", playlistID)
	return nil
}

var playlistIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{10,}$`)

// ImportPlaylist imports the metadata of a public YouTube Music playlist.
func (a *App) ImportPlaylist(rawURL string) (ImportDTO, error) {
	cfg := a.currentConfig()
	if !cfg.Features.YouTubeMusic {
		return ImportDTO{Notice: "El catalogo de YouTube Music esta desactivado en Ajustes"}, nil
	}
	playlistID, err := parsePlaylistURL(rawURL)
	if err != nil {
		return ImportDTO{}, err
	}
	ctx, cancel := a.opCtx()
	defer cancel()

	tracks, err := a.music.ImportPlaylist(ctx, playlistID)
	if err != nil {
		return ImportDTO{}, err
	}
	if len(tracks) == 0 {
		return ImportDTO{Notice: "La lista no devolvio canciones (puede ser privada o regional)"}, nil
	}

	saved := make([]domain.Track, 0, len(tracks))
	for _, track := range tracks {
		id, err := a.db.SaveTrack(ctx, track)
		if err != nil {
			return ImportDTO{}, err
		}
		track.ID = id
		saved = append(saved, track)
	}
	name := "Importada de YouTube Music"
	playlist, err := a.db.CreatePlaylist(ctx, domain.Playlist{
		Name:       name,
		Origin:     "ytmusic",
		ExternalID: playlistID,
	})
	if err != nil {
		return ImportDTO{}, err
	}
	for _, track := range saved {
		if err := a.db.AddToPlaylist(ctx, playlist.ID, track.ID); err != nil {
			return ImportDTO{}, err
		}
	}
	playlist, _ = a.db.Playlist(ctx, playlist.ID)
	dto := toPlaylistDTO(playlist)
	return ImportDTO{
		Imported: a.toTrackList(ctx, saved),
		Notice: fmt.Sprintf("Importadas %d pistas como metadatos. La reproduccion sigue en la app oficial.",
			len(saved)),
		Playlist: &dto,
	}, nil
}

func parsePlaylistURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if playlistIDRe.MatchString(raw) {
		return raw, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("esa URL no es valida")
	}
	host := strings.ToLower(parsed.Hostname())
	if !strings.Contains(host, "youtube.com") && !strings.Contains(host, "youtu.be") {
		return "", fmt.Errorf("pega una URL de YouTube Music")
	}
	id := parsed.Query().Get("list")
	if id == "" {
		return "", fmt.Errorf("la URL no contiene el identificador de una lista (parametro list)")
	}
	return id, nil
}

// OpenExternal opens a link in the system browser after validating the scheme.
func (a *App) OpenExternal(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("esa URL no es valida")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto":
	default:
		return fmt.Errorf("solo abro enlaces http, https o mailto")
	}
	if a.opener == nil {
		return fmt.Errorf("no puedo abrir enlaces en este entorno")
	}
	return a.opener.OpenURL(parsed.String())
}

// ResolveAudioPath maps a media URL id to a file, for the streaming handler.
func (a *App) ResolveAudioPath(trackID string) (string, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	track, err := a.db.Track(ctx, trackID)
	if err != nil {
		return "", media.ErrNoAudio
	}
	if track.AudioPath == "" {
		return "", media.ErrNoAudio
	}
	return track.AudioPath, nil
}
