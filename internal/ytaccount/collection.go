package ytaccount

import (
	"context"
	"errors"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// Collection is an album, a playlist or an artist page opened from a card.
type Collection struct {
	ID       string
	Title    string
	Subtitle string
	Artist   string
	ArtistID string
	Cover    string
	Kind     string
	Meta     string
	Tracks   []domain.Track
	Shelves  []Shelf
}

// Collection reads one page of the catalogue. Albums and playlists bring their
// songs; an artist brings its popular songs plus the rows of albums, singles
// and playlists the account can keep exploring.
func (c *Client) Collection(ctx context.Context, rawID string) (Collection, error) {
	id := strings.TrimSpace(rawID)
	if id == "" {
		return Collection{}, errors.New("empty collection id")
	}
	kind := "playlist"
	browse := "VL" + strings.TrimPrefix(id, "VL")
	switch {
	case strings.HasPrefix(id, "UC"):
		kind, browse = "artist", id
	case strings.HasPrefix(id, "MPRE"):
		kind, browse = "album", id
	}
	res, err := c.call(ctx, "browse", map[string]any{"browseId": browse})
	if err != nil {
		return Collection{}, err
	}
	out := Collection{ID: id, Kind: kind}

	if header := firstRenderer(res, "musicImmersiveHeaderRenderer"); header != nil {
		out.Title = strings.TrimSpace(runsText(header["title"]))
		out.Cover = thumbnail(header)
		out.Subtitle = "Artista"
	} else if header := firstRenderer(res, "musicResponsiveHeaderRenderer", "musicDetailHeaderRenderer"); header != nil {
		out.Title = strings.TrimSpace(runsText(header["title"]))
		out.Subtitle = strings.TrimSpace(runsJoin(header["subtitle"]))
		out.Meta = strings.TrimSpace(runsJoin(header["secondSubtitle"]))
		out.Cover = thumbnail(header)
		if run, ok := firstRun(header["straplineTextOne"]); ok {
			out.Artist, _ = run["text"].(string)
			out.Artist = strings.TrimSpace(out.Artist)
			out.ArtistID, _ = dig(run, "navigationEndpoint", "browseEndpoint", "browseId").(string)
		}
	}

	seen := map[string]bool{}
	for _, row := range collectPlaylistRows(res) {
		track, ok := playlistTrack(row)
		if !ok || track.ProviderID == "" || seen[track.ProviderID] {
			continue
		}
		if track.Artist == "" {
			track.Artist = out.Artist
		}
		if track.ArtworkURL == "" {
			track.ArtworkURL = out.Cover
		}
		seen[track.ProviderID] = true
		out.Tracks = append(out.Tracks, track)
	}

	if kind == "artist" {
		var shelves []map[string]any
		collectShelves(res, &shelves)
		for _, raw := range shelves {
			title := shelfTitle(raw)
			items := shelfItems(raw)
			if title == "" || len(items) < 2 {
				continue
			}
			out.Shelves = append(out.Shelves, Shelf{Title: title, Items: items})
		}
	}
	return out, nil
}

// firstRenderer returns the first renderer with any of the given names.
func firstRenderer(node any, names ...string) map[string]any {
	var found map[string]any
	var walk func(n any) bool
	walk = func(n any) bool {
		switch value := n.(type) {
		case map[string]any:
			for _, name := range names {
				if row, ok := value[name].(map[string]any); ok {
					found = row
					return true
				}
			}
			for _, child := range value {
				if walk(child) {
					return true
				}
			}
		case []any:
			for _, child := range value {
				if walk(child) {
					return true
				}
			}
		}
		return false
	}
	walk(node)
	return found
}

// firstRun returns the first run of a text renderer.
func firstRun(text any) (map[string]any, bool) {
	runs, _ := dig(text, "runs").([]any)
	if len(runs) == 0 {
		return nil, false
	}
	run, ok := runs[0].(map[string]any)
	return run, ok
}
