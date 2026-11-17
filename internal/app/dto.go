package app

import (
	"net/url"
	"time"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// DTOs are the wire format between Go and the React app. They are plain structs
// with JSON tags: no domain type crosses the bridge directly, so the storage
// model can evolve without breaking the UI.

// InfoDTO is the About panel and window metadata.
type InfoDTO struct {
	Name       string `json:"name"`
	Tagline    string `json:"tagline"`
	Version    string `json:"version"`
	Identifier string `json:"identifier"`
	DBPath     string `json:"dbPath"`
	Schema     int    `json:"schema"`
}

// TrackDTO is a track as the UI sees it.
type TrackDTO struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	DurationMs  int64  `json:"durationMs"`
	ArtworkURL  string `json:"artworkUrl"`
	ExternalURL string `json:"externalUrl"`
	AudioURL    string `json:"audioUrl"`
	Playable    bool   `json:"playable"`
	InLibrary   bool   `json:"inLibrary"`
	Favorite    bool   `json:"favorite"`
	Note        string `json:"note"`
}

// PlaylistDTO is a playlist summary.
type PlaylistDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Artwork     string `json:"artwork"`
	Origin      string `json:"origin"`
	ExternalID  string `json:"externalId"`
	TrackCount  int    `json:"trackCount"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// WordTimingDTO is one word with its span.
type WordTimingDTO struct {
	Text    string `json:"text"`
	StartMs int64  `json:"startMs"`
	EndMs   int64  `json:"endMs"`
}

// NoteDTO is a pronunciation hint.
type NoteDTO struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnnotationDTO is study data for a line.
type AnnotationDTO struct {
	LineID        string    `json:"lineId"`
	Accent        string    `json:"accent"`
	IPA           string    `json:"ipa"`
	SpanishApprox string    `json:"spanishApprox"`
	Translation   string    `json:"translation"`
	Notes         []NoteDTO `json:"notes"`
	Model         string    `json:"model"`
}

// LineDTO is one lyric line with optional study data.
type LineDTO struct {
	ID         string          `json:"id"`
	Index      int             `json:"index"`
	Text       string          `json:"text"`
	StartMs    int64           `json:"startMs"`
	EndMs      int64           `json:"endMs"`
	Words      []WordTimingDTO `json:"words"`
	Annotation *AnnotationDTO  `json:"annotation"`
}

// LyricsDTO is a lyric document plus its availability state.
type LyricsDTO struct {
	TrackID   string    `json:"trackId"`
	Available bool      `json:"available"`
	Provider  string    `json:"provider"`
	Sync      string    `json:"sync"`
	Source    string    `json:"source"`
	License   string    `json:"license"`
	HasWords  bool      `json:"hasWords"`
	Lines     []LineDTO `json:"lines"`
	Notice    string    `json:"notice"`
}

// WordDTO is a vocabulary entry.
type WordDTO struct {
	ID           string `json:"id"`
	Word         string `json:"word"`
	Lemma        string `json:"lemma"`
	PartOfSpeech string `json:"partOfSpeech"`
	Meaning      string `json:"meaning"`
	IPA          string `json:"ipa"`
	Translation  string `json:"translation"`
	Example      string `json:"example"`
	Source       string `json:"source"`
	TrackID      string `json:"trackId"`
	CreatedAt    int64  `json:"createdAt"`
}

// ProgressDTO is per-track study progress.
type ProgressDTO struct {
	TrackID        string `json:"trackId"`
	Title          string `json:"title"`
	Artist         string `json:"artist"`
	LinesPracticed int    `json:"linesPracticed"`
	StudySeconds   int64  `json:"studySeconds"`
	LastStudiedAt  int64  `json:"lastStudiedAt"`
	Status         string `json:"status"`
	WordsSaved     int    `json:"wordsSaved"`
}

// HistoryDTO is one playback row.
type HistoryDTO struct {
	ID         int64  `json:"id"`
	TrackID    string `json:"trackId"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	PlayedAt   int64  `json:"playedAt"`
	ListenedMs int64  `json:"listenedMs"`
}

// LLMSettingsDTO is the editable AI configuration (never the key).
type LLMSettingsDTO struct {
	BaseURL        string  `json:"baseUrl"`
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	TimeoutSeconds int     `json:"timeoutSeconds"`
	MaxRetries     int     `json:"maxRetries"`
	JSONMode       bool    `json:"jsonMode"`
}

// FeaturesDTO gates experimental providers.
type FeaturesDTO struct {
	YouTubeMusic bool `json:"youtubeMusic"`
}

// SettingsDTO is everything the settings screen needs.
type SettingsDTO struct {
	LLM           LLMSettingsDTO `json:"llm"`
	Features      FeaturesDTO    `json:"features"`
	Accent        string         `json:"accent"`
	HasKey        bool           `json:"hasKey"`
	KeyAvailable  bool           `json:"keyStoreAvailable"`
	ConfigPath    string         `json:"configPath"`
	KeyStoreError string         `json:"keyStoreError"`
}

// ProviderStateDTO reports a provider's availability.
type ProviderStateDTO struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

// StoredCountsDTO summarizes local data.
type StoredCountsDTO struct {
	Tracks        int `json:"tracks"`
	LibraryTracks int `json:"libraryTracks"`
	Favorites     int `json:"favorites"`
	Playlists     int `json:"playlists"`
	Vocabulary    int `json:"vocabulary"`
	Annotated     int `json:"annotated"`
}

// BootstrapDTO is the first payload the UI loads.
type BootstrapDTO struct {
	App            InfoDTO            `json:"app"`
	DemoTracks     []TrackDTO         `json:"demoTracks"`
	RecentlyPlayed []TrackDTO         `json:"recentlyPlayed"`
	Library        []TrackDTO         `json:"library"`
	Favorites      []TrackDTO         `json:"favorites"`
	Playlists      []PlaylistDTO      `json:"playlists"`
	Progress       []ProgressDTO      `json:"progress"`
	Vocabulary     []WordDTO          `json:"vocabulary"`
	Settings       SettingsDTO        `json:"settings"`
	Providers      []ProviderStateDTO `json:"providers"`
	Counts         StoredCountsDTO    `json:"counts"`
}

// SearchDTO is a search result set with its provider states.
type SearchDTO struct {
	Query     string             `json:"query"`
	Items     []TrackDTO         `json:"items"`
	Providers []ProviderStateDTO `json:"providers"`
	Notice    string             `json:"notice"`
}

// ImportDTO reports an import.
type ImportDTO struct {
	Imported []TrackDTO   `json:"imported"`
	Skipped  []string     `json:"skipped"`
	Notice   string       `json:"notice"`
	Playlist *PlaylistDTO `json:"playlist"`
}

// AnnotateOutcome reports an annotation run.
type AnnotateOutcome struct {
	Generated  int      `json:"generated"`
	Cached     int      `json:"cached"`
	Failed     []string `json:"failed"`
	NeedsSetup bool     `json:"needsSetup"`
	Message    string   `json:"message"`
}

// TestOutcome reports an AI connectivity check.
type TestOutcome struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latencyMs"`
}

// WordRequest asks for a dictionary lookup.
type WordRequest struct {
	TrackID string `json:"trackId"`
	LineID  string `json:"lineId"`
	Word    string `json:"word"`
	Context string `json:"context"`
}

// WordInput is a vocabulary entry coming from the UI.
type WordInput struct {
	Word         string `json:"word"`
	Lemma        string `json:"lemma"`
	PartOfSpeech string `json:"partOfSpeech"`
	Meaning      string `json:"meaning"`
	IPA          string `json:"ipa"`
	Translation  string `json:"translation"`
	Example      string `json:"example"`
	TrackID      string `json:"trackId"`
}

// StudyInput records practice time.
type StudyInput struct {
	TrackID string `json:"trackId"`
	Seconds int64  `json:"seconds"`
	Lines   int    `json:"lines"`
}

// SettingsInput is a settings update from the UI.
type SettingsInput struct {
	LLM      LLMSettingsDTO `json:"llm"`
	Features FeaturesDTO    `json:"features"`
	Accent   string         `json:"accent"`
	APIKey   string         `json:"apiKey"`
}

// WindowStateInput is the geometry to persist.
type WindowStateInput struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Maximised bool `json:"maximised"`
}

// ---- converters ----

func toTrackDTO(t domain.Track, inLibrary, favorite bool) TrackDTO {
	dto := TrackDTO{
		ID:          t.ID,
		Provider:    string(t.Provider),
		Title:       t.Title,
		Artist:      t.Artist,
		Album:       t.Album,
		DurationMs:  t.DurationMs,
		ArtworkURL:  t.ArtworkURL,
		ExternalURL: t.ExternalURL,
		Playable:    t.Playable(),
		InLibrary:   inLibrary,
		Favorite:    favorite,
	}
	if t.Playable() {
		dto.AudioURL = "/media/" + url.PathEscape(t.ID)
	}
	return dto
}

func toPlaylistDTO(p domain.Playlist) PlaylistDTO {
	return PlaylistDTO{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Artwork:     p.Artwork,
		Origin:      p.Origin,
		ExternalID:  p.ExternalID,
		TrackCount:  p.TrackCount,
		UpdatedAt:   p.UpdatedAt.UnixMilli(),
	}
}

func toLineDTOs(lines []domain.Line, annotations map[string]domain.Annotation) []LineDTO {
	out := make([]LineDTO, 0, len(lines))
	for _, line := range lines {
		dto := LineDTO{
			ID:      line.ID,
			Index:   line.Index,
			Text:    line.Text,
			StartMs: line.StartMs,
			EndMs:   line.EndMs,
			Words:   make([]WordTimingDTO, 0, len(line.Words)),
		}
		for _, w := range line.Words {
			dto.Words = append(dto.Words, WordTimingDTO{Text: w.Text, StartMs: w.StartMs, EndMs: w.EndMs})
		}
		if a, ok := annotations[line.ID]; ok {
			notes := make([]NoteDTO, 0, len(a.Notes))
			for _, n := range a.Notes {
				notes = append(notes, NoteDTO{Type: n.Type, Text: n.Text})
			}
			dto.Annotation = &AnnotationDTO{
				LineID:        a.LineID,
				Accent:        a.Accent,
				IPA:           a.IPA,
				SpanishApprox: a.SpanishApprox,
				Translation:   a.Translation,
				Notes:         notes,
				Model:         a.Model,
			}
		}
		out = append(out, dto)
	}
	return out
}

func toWordDTO(w domain.WordEntry) WordDTO {
	return WordDTO{
		ID:           w.ID,
		Word:         w.Word,
		Lemma:        w.Lemma,
		PartOfSpeech: w.PartOfSpeech,
		Meaning:      w.Meaning,
		IPA:          w.IPA,
		Translation:  w.Translation,
		Example:      w.Example,
		Source:       w.Source,
		TrackID:      w.TrackID,
		CreatedAt:    w.CreatedAt.UnixMilli(),
	}
}

func toProgressDTO(p domain.Progress) ProgressDTO {
	dto := ProgressDTO{
		TrackID:        p.TrackID,
		LinesPracticed: p.LinesPracticed,
		StudySeconds:   p.StudySeconds,
		Status:         p.Status,
	}
	if !p.LastStudiedAt.IsZero() {
		dto.LastStudiedAt = p.LastStudiedAt.UnixMilli()
	}
	return dto
}

func toHistoryDTO(h domain.HistoryEntry) HistoryDTO {
	return HistoryDTO{
		ID:         h.ID,
		TrackID:    h.TrackID,
		Title:      h.TrackTitle,
		Artist:     h.TrackArtist,
		PlayedAt:   h.PlayedAt.UnixMilli(),
		ListenedMs: h.ListenedMs,
	}
}

func millis(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}
