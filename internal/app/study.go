package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/lyrics"
	"github.com/LuiisDev21/stefani/internal/phonetics"
	"github.com/LuiisDev21/stefani/internal/secret"
)

// llmConfig adapts the settings plus the keychain to the phonetics service.
func (a *App) llmConfig() (config.LLM, string, error) {
	cfg := a.currentConfig()
	key, err := a.secrets.Get(secret.KeyLLMAPIKey)
	if err != nil {
		return cfg.LLM, "", err
	}
	return cfg.LLM, key, nil
}

func (a *App) annotationsMap(ctx context.Context, trackID string) map[string]domain.Annotation {
	items, err := a.db.AnnotationsByTrack(ctx, trackID)
	if err != nil {
		return map[string]domain.Annotation{}
	}
	out := make(map[string]domain.Annotation, len(items))
	for _, item := range items {
		out[item.LineID] = item
	}
	return out
}

// Lyrics returns the lyric document for a track, fetching it when needed.
//
// A panic inside a provider must never reach the bridge as an opaque error: it
// is logged with its stack and reported as a readable state instead.
func (a *App) Lyrics(trackID string, refresh bool) (dto LyricsDTO, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			a.log.Error("lyrics panic", "panic", recovered, "stack", string(debug.Stack()))
			dto = LyricsDTO{TrackID: trackID, Notice: "No pude cargar la letra de esta cancion."}
			err = nil
		}
	}()

	ctx, cancel := a.opCtx()
	defer cancel()

	track, err := a.db.Track(ctx, trackID)
	if err != nil {
		a.log.Warn("lyrics: track not found", "track", trackID, "error", err)
		return LyricsDTO{}, fmt.Errorf("no encontré esa canción")
	}
	doc, err := a.lyrics.ForTrack(ctx, track, refresh)
	if err != nil {
		if errors.Is(err, lyrics.ErrNoLyrics) {
			return LyricsDTO{
				TrackID: trackID,
				Notice:  noLyricsNotice(track),
			}, nil
		}
		return LyricsDTO{TrackID: trackID, Notice: err.Error()}, nil
	}
	return LyricsDTO{
		TrackID:   trackID,
		Available: true,
		Provider:  doc.Provider,
		Sync:      string(doc.Sync),
		Source:    doc.Source,
		License:   doc.License,
		HasWords:  doc.HasWordTimings(),
		Lines:     a.toLineDTOs(doc.Lines, a.annotationsMap(ctx, trackID)),
	}, nil
}

func noLyricsNotice(track domain.Track) string {
	if track.Provider == domain.ProviderYTMusic {
		return "No encontré letra para esta canción en LRCLIB. Puedes importar un archivo .lrc para estudiarla."
	}
	return "No hay letra disponible para esta cancion. Puedes importar un archivo .lrc."
}

// ImportLyricsFile lets the user attach a local .lrc to a track.
func (a *App) ImportLyricsFile(trackID string) (LyricsDTO, error) {
	if a.picker == nil {
		return LyricsDTO{}, fmt.Errorf("el selector de archivos no esta disponible en este entorno")
	}
	path, err := a.picker.PickLyricsFile()
	if err != nil {
		return LyricsDTO{}, fmt.Errorf("no pude abrir el selector de archivos: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return a.Lyrics(trackID, false)
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	track, err := a.db.Track(ctx, trackID)
	if err != nil {
		return LyricsDTO{}, fmt.Errorf("no encontré esa canción")
	}
	if _, err := a.lyrics.ImportFile(ctx, track, path); err != nil {
		return LyricsDTO{}, err
	}
	return a.Lyrics(trackID, false)
}

// Annotations returns stored study data for a track.
func (a *App) Annotations(trackID string) ([]AnnotationDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	items, err := a.db.AnnotationsByTrack(ctx, trackID)
	if err != nil {
		return nil, err
	}
	out := make([]AnnotationDTO, 0, len(items))
	for _, item := range items {
		notes := make([]NoteDTO, 0, len(item.Notes))
		for _, n := range item.Notes {
			notes = append(notes, NoteDTO{Type: n.Type, Text: n.Text})
		}
		out = append(out, AnnotationDTO{
			LineID:        item.LineID,
			Accent:        item.Accent,
			IPA:           item.IPA,
			SpanishApprox: item.SpanishApprox,
			Translation:   item.Translation,
			Notes:         notes,
			Model:         item.Model,
		})
	}
	return out, nil
}

// AnnotateTrack generates IPA, Spanish approximation and translation for the
// track lyrics, emitting progress events as it goes.
func (a *App) AnnotateTrack(trackID string) (AnnotateOutcome, error) {
	ctx, cancel := a.opCtx()
	defer cancel()

	cfg := a.currentConfig()
	accent := cfg.Accent
	if accent == "" {
		accent = config.DefaultAccent
	}
	outcome := AnnotateOutcome{}
	result, err := a.phon.AnnotateTrack(ctx, trackID, accent, func(p phonetics.Progress) {
		a.emit("annotate:progress", map[string]any{
			"trackId": trackID,
			"done":    p.Done,
			"total":   p.Total,
		})
	})
	if err != nil {
		outcome.Message = annotateMessage(err)
		outcome.NeedsSetup = needsSetup(err)
		if outcome.NeedsSetup {
			return outcome, nil
		}
		return outcome, err
	}
	outcome.Generated = result.Generated
	outcome.Cached = result.Cached
	outcome.Failed = result.Failed
	if len(result.Failed) > 0 {
		outcome.Message = fmt.Sprintf("%d verso(s) no pasaron la validacion y no se guardaron", len(result.Failed))
	}
	return outcome, nil
}

func annotateMessage(err error) string {
	switch {
	case errors.Is(err, ai.ErrNoEndpoint):
		return "Configura un proveedor de IA compatible y su clave en Ajustes."
	case errors.Is(err, secret.ErrNotFound):
		return "Falta la clave de la API de IA. Agregala en Ajustes."
	case errors.Is(err, secret.ErrUnavailable):
		return "El almacen de credenciales del sistema no esta disponible; no puedo leer la clave."
	case errors.Is(err, ai.ErrAuth):
		return "El proveedor de IA rechazo la clave. Revisa la configuracion."
	case errors.Is(err, ai.ErrRateLimited):
		return "El proveedor de IA limito las peticiones. Espera un momento y reintenta."
	case errors.Is(err, context.DeadlineExceeded):
		return "La generacion tardo demasiado. Revisa la conexion y reintenta."
	default:
		return err.Error()
	}
}

func needsSetup(err error) bool {
	return errors.Is(err, ai.ErrNoEndpoint) || errors.Is(err, secret.ErrNotFound) ||
		errors.Is(err, secret.ErrUnavailable)
}

// ResetAnnotations drops generated study data for a track.
func (a *App) ResetAnnotations(trackID string) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.phon.ResetCache(ctx, trackID)
}

// LookupWord asks the AI provider for dictionary data about a word.
func (a *App) LookupWord(req WordRequest) (WordDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	context := strings.TrimSpace(req.Context)
	if context == "" && req.LineID != "" {
		if doc, err := a.db.Lyrics(ctx, req.TrackID); err == nil {
			for _, line := range doc.Lines {
				if line.ID == req.LineID {
					context = line.Text
					break
				}
			}
		}
	}
	entry, err := a.phon.LookupWord(ctx, req.Word, context)
	if err != nil {
		if needsSetup(err) {
			return WordDTO{}, fmt.Errorf("%s", annotateMessage(err))
		}
		return WordDTO{}, err
	}
	entry.TrackID = req.TrackID
	return toWordDTO(entry), nil
}

// Vocabulary lists saved words.
func (a *App) Vocabulary() ([]WordDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	items, err := a.db.Vocabulary(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]WordDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toWordDTO(item))
	}
	return out, nil
}

// SaveWord stores a vocabulary entry.
func (a *App) SaveWord(in WordInput) (WordDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	if strings.TrimSpace(in.Word) == "" {
		return WordDTO{}, fmt.Errorf("la palabra esta vacia")
	}
	entry, err := a.db.SaveVocabulary(ctx, domain.WordEntry{
		Word:         strings.TrimSpace(in.Word),
		Lemma:        strings.TrimSpace(in.Lemma),
		PartOfSpeech: strings.TrimSpace(in.PartOfSpeech),
		Meaning:      strings.TrimSpace(in.Meaning),
		IPA:          strings.TrimSpace(in.IPA),
		Translation:  strings.TrimSpace(in.Translation),
		Example:      strings.TrimSpace(in.Example),
		Source:       "ai",
		TrackID:      in.TrackID,
	})
	if err != nil {
		return WordDTO{}, err
	}
	return toWordDTO(entry), nil
}

// DeleteWord removes a saved word.
func (a *App) DeleteWord(id string) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.db.DeleteVocabulary(ctx, id)
}

// RecordStudy adds practice time and practiced lines.
func (a *App) RecordStudy(in StudyInput) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if in.TrackID == "" {
		return fmt.Errorf("falta la cancion")
	}
	if in.Seconds > 12*3600 {
		in.Seconds = 12 * 3600
	}
	if in.Lines < 0 || in.Lines > 5000 {
		return fmt.Errorf("numero de versos invalido")
	}
	return a.db.RecordStudy(ctx, in.TrackID, in.Seconds, in.Lines)
}

// Progress lists study progress with track metadata.
func (a *App) Progress() ([]ProgressDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	items, err := a.db.AllProgress(ctx)
	if err != nil {
		return nil, err
	}
	words, err := a.db.Vocabulary(ctx)
	if err != nil {
		return nil, err
	}
	wordsByTrack := make(map[string]int, len(words))
	for _, w := range words {
		if w.TrackID != "" {
			wordsByTrack[w.TrackID]++
		}
	}
	out := make([]ProgressDTO, 0, len(items))
	for _, item := range items {
		dto := toProgressDTO(item)
		if track, err := a.db.Track(ctx, item.TrackID); err == nil {
			dto.Title = track.Title
			dto.Artist = track.Artist
		}
		dto.WordsSaved = wordsByTrack[item.TrackID]
		out = append(out, dto)
	}
	return out, nil
}

// History lists recent playback.
func (a *App) History(limit int) ([]HistoryDTO, error) {
	ctx, cancel := a.opCtx()
	defer cancel()
	items, err := a.db.History(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]HistoryDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toHistoryDTO(item))
	}
	return out, nil
}

// ReportPlayback records a playback session.
func (a *App) ReportPlayback(trackID string, listenedMs int64) error {
	ctx, cancel := a.opCtx()
	defer cancel()
	if listenedMs < 0 {
		listenedMs = 0
	}
	if listenedMs > 6*3600*1000 {
		listenedMs = 6 * 3600 * 1000
	}
	return a.db.RecordPlayback(ctx, trackID, listenedMs)
}

// ---- settings ----

// Settings returns the current configuration without secrets.
func (a *App) Settings() (SettingsDTO, error) {
	cfg := a.currentConfig()
	dto := SettingsDTO{
		LLM: LLMSettingsDTO{
			BaseURL:        cfg.LLM.BaseURL,
			Model:          cfg.LLM.Model,
			Temperature:    cfg.LLM.Temperature,
			TimeoutSeconds: cfg.LLM.TimeoutSeconds,
			MaxRetries:     cfg.LLM.MaxRetries,
			JSONMode:       cfg.LLM.JSONMode,
		},
		Features:     FeaturesDTO{YouTubeMusic: cfg.Features.YouTubeMusic},
		Accent:       cfg.Accent,
		ConfigPath:   a.configs.Path(),
		KeyAvailable: a.secrets.Available(),
	}
	if _, err := a.secrets.Get(secret.KeyLLMAPIKey); err == nil {
		dto.HasKey = true
	} else if errors.Is(err, secret.ErrUnavailable) {
		dto.KeyStoreError = "El almacen de credenciales del sistema no responde"
	}
	return dto, nil
}

// SaveSettings validates and persists settings, storing the key in the keychain.
func (a *App) SaveSettings(in SettingsInput) (SettingsDTO, error) {
	if err := validateLLM(in.LLM); err != nil {
		return SettingsDTO{}, err
	}
	cfg := a.currentConfig()
	cfg.LLM = config.LLM{
		BaseURL:        strings.TrimRight(strings.TrimSpace(in.LLM.BaseURL), "/"),
		Model:          strings.TrimSpace(in.LLM.Model),
		Temperature:    in.LLM.Temperature,
		TimeoutSeconds: in.LLM.TimeoutSeconds,
		MaxRetries:     in.LLM.MaxRetries,
		JSONMode:       in.LLM.JSONMode,
	}
	cfg.Features.YouTubeMusic = in.Features.YouTubeMusic
	if accent := strings.TrimSpace(in.Accent); accent != "" {
		cfg.Accent = accent
	}
	if err := a.configs.Save(cfg); err != nil {
		return SettingsDTO{}, err
	}
	a.mu.Lock()
	a.cfg = cfg
	onChange := a.onConfigChange
	a.mu.Unlock()
	if onChange != nil {
		onChange(cfg)
	}
	if key := strings.TrimSpace(in.APIKey); key != "" {
		if err := a.secrets.Set(secret.KeyLLMAPIKey, key); err != nil {
			if errors.Is(err, secret.ErrUnavailable) {
				return a.Settings()
			}
			return SettingsDTO{}, err
		}
	}
	a.emit("settings:changed")
	return a.Settings()
}

func validateLLM(in LLMSettingsDTO) error {
	base := strings.TrimRight(strings.TrimSpace(in.BaseURL), "/")
	if base == "" {
		return fmt.Errorf("la URL base es obligatoria")
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("la URL base no es valida")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
	case "http":
		host := strings.ToLower(parsed.Hostname())
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return fmt.Errorf("usa https; http solo se permite para endpoints locales")
		}
	default:
		return fmt.Errorf("la URL base debe usar https")
	}
	if strings.TrimSpace(in.Model) == "" {
		return fmt.Errorf("el modelo es obligatorio")
	}
	if in.Temperature < 0 || in.Temperature > 2 {
		return fmt.Errorf("la temperatura debe estar entre 0 y 2")
	}
	if in.TimeoutSeconds < 5 || in.TimeoutSeconds > 300 {
		return fmt.Errorf("el timeout debe estar entre 5 y 300 segundos")
	}
	if in.MaxRetries < 0 || in.MaxRetries > 5 {
		return fmt.Errorf("los reintentos deben estar entre 0 y 5")
	}
	return nil
}

// TestLLM checks connectivity with the configured provider.
func (a *App) TestLLM() (TestOutcome, error) {
	cfg := a.currentConfig()
	key, err := a.secrets.Get(secret.KeyLLMAPIKey)
	if err != nil {
		outcome := TestOutcome{OK: false, Message: annotateMessage(err)}
		return outcome, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(cfg.LLM.TimeoutSeconds)*time.Second)
	defer cancel()
	start := time.Now()
	if err := a.ai.Ping(ctx, ai.Config{
		BaseURL:     cfg.LLM.BaseURL,
		Model:       cfg.LLM.Model,
		APIKey:      key,
		Temperature: 0,
		Timeout:     time.Duration(cfg.LLM.TimeoutSeconds) * time.Second,
		MaxRetries:  0,
		JSONMode:    cfg.LLM.JSONMode,
	}); err != nil {
		return TestOutcome{OK: false, Message: annotateMessage(err), LatencyMs: time.Since(start).Milliseconds()}, nil
	}
	return TestOutcome{OK: true, Message: "Conexion correcta", LatencyMs: time.Since(start).Milliseconds()}, nil
}

// ClearLLMKey deletes the stored API key.
func (a *App) ClearLLMKey() (SettingsDTO, error) {
	if err := a.secrets.Delete(secret.KeyLLMAPIKey); err != nil {
		return SettingsDTO{}, err
	}
	return a.Settings()
}

// SaveWindowState persists the window geometry.
func (a *App) SaveWindowState(in WindowStateInput) error {
	if in.Width < 400 || in.Height < 300 || in.Width > 10000 || in.Height > 10000 {
		return nil
	}
	cfg := a.currentConfig()
	cfg.Window = config.Window{
		Width: in.Width, Height: in.Height, X: in.X, Y: in.Y,
		Maximised: in.Maximised, Valid: true,
	}
	// A geometry update is best-effort: never block quitting on it.
	if err := a.configs.Save(cfg); err != nil {
		a.log.Warn("cannot persist window state", "error", err)
		return nil
	}
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
	return nil
}

// WindowState reports the stored geometry so the window can restore itself.
func (a *App) WindowState() WindowStateInput {
	cfg := a.currentConfig()
	if !cfg.Window.Valid {
		return WindowStateInput{Width: 1280, Height: 820}
	}
	return WindowStateInput{
		Width: cfg.Window.Width, Height: cfg.Window.Height,
		X: cfg.Window.X, Y: cfg.Window.Y, Maximised: cfg.Window.Maximised,
	}
}
