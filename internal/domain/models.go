// Package domain holds the core types shared by every layer of STEFANI.
// Nothing here imports storage, network or UI code: these are plain data
// structures plus the small rules that operate on them.
package domain

import "time"

// Provider identifies where a track or lyric comes from. Providers are
// interchangeable; the educational engine never depends on a specific one.
type Provider string

const (
	// ProviderDemo backs the catalog used by the test fixtures only: the app no
	// longer seeds it, but the Go and end-to-end suites need stable songs with
	// word timings.
	ProviderDemo    Provider = "demo"
	ProviderLocal   Provider = "local"
	ProviderYTMusic Provider = "ytmusic"
)

// SyncLevel describes how precisely a lyric document is timed. The renderer
// must never present a lower level as if it were a higher one.
type SyncLevel string

const (
	SyncUnsynced SyncLevel = "UNSYNCED"
	SyncLine     SyncLevel = "LINE"
	SyncWord     SyncLevel = "WORD"
	SyncSyllable SyncLevel = "SYLLABLE"
)

// Rank orders sync levels so callers can compare them.
func (s SyncLevel) Rank() int {
	switch s {
	case SyncUnsynced:
		return 0
	case SyncLine:
		return 1
	case SyncWord:
		return 2
	case SyncSyllable:
		return 3
	}
	return 0
}

// Track is a song known to the application, from any provider.
type Track struct {
	ID          string
	Provider    Provider
	ProviderID  string
	Title       string
	Artist      string
	Album       string
	DurationMs  int64
	ArtworkURL  string
	ExternalURL string
	// AudioPath is the local file backing playback, when the provider has one.
	// Empty for catalog-only entries: those can be displayed but not played.
	AudioPath string
	CreatedAt time.Time
}

// Playable reports whether the app can actually produce sound for this track.
func (t Track) Playable() bool { return t.AudioPath != "" }

// Word is a single timed token inside a line.
type Word struct {
	Text      string
	StartMs   int64
	EndMs     int64
	TimingSrc string // provider | estimated | generated
}

// Line is one lyric verse, optionally word-timed.
type Line struct {
	ID      string
	Index   int
	Text    string
	StartMs int64
	EndMs   int64
	Words   []Word
}

// Lyrics is a normalized lyric document, independent of its source format.
type Lyrics struct {
	ID       string
	TrackID  string
	Provider string
	Language string
	Sync     SyncLevel
	Source   string
	License  string
	Lines    []Line
}

// HasWordTimings reports whether every non-empty line carries word timings,
// which is what the karaoke renderer needs.
func (l Lyrics) HasWordTimings() bool {
	if l.Sync.Rank() < SyncWord.Rank() || len(l.Lines) == 0 {
		return false
	}
	for _, line := range l.Lines {
		if line.Text != "" && len(line.Words) == 0 {
			return false
		}
	}
	return true
}

// Note is a short pronunciation explanation attached to a line.
type Note struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Annotation is the AI-generated study data for one line.
type Annotation struct {
	LineID        string
	Accent        string
	IPA           string
	SpanishApprox string
	Translation   string
	Notes         []Note
	Model         string
	Version       string
	CacheKey      string
	CreatedAt     time.Time
}

// Playlist is a local (or imported copy of an external) collection.
type Playlist struct {
	ID          string
	Name        string
	Description string
	Artwork     string
	Origin      string // local | ytmusic | file
	ExternalID  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TrackCount  int
}

// PlaylistItem is one ordered entry of a playlist.
type PlaylistItem struct {
	Position int
	Track    Track
}

// HistoryEntry is one playback record.
type HistoryEntry struct {
	ID          int64
	TrackID     string
	PlayedAt    time.Time
	ListenedMs  int64
	TrackTitle  string
	TrackArtist string
}

// SearchResult is one provider hit.
type SearchResult struct {
	Track        Track
	ProviderName string
	Playable     bool
	Note         string
}

// ImportedPlaylist is the outcome of importing an external playlist.
type ImportedPlaylist struct {
	Playlist    Playlist
	Imported    int
	Unavailable []string
	Source      string
}
