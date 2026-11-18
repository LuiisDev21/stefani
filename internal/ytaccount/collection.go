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
		// The discography brings every release; the rows on the artist page
		// only carry the first handful of each kind.
		full := c.discography(ctx, id)
		covered := map[string]bool{}
		for _, shelf := range full {
			covered[shelf.Title] = true
			out.Shelves = append(out.Shelves, shelf)
		}
		var shelves []map[string]any
		collectShelves(res, &shelves)
		for _, raw := range shelves {
			title := shelfTitle(raw)
			if title == "" || covered[title] {
				continue
			}
			items := shelfItems(raw)
			if len(items) < 2 {
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

// FindArtist resolves an artist name to its page id, so any song can lead to
// the artist even when it was imported without one.
func (c *Client) FindArtist(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("empty artist name")
	}
	res, err := c.call(ctx, "search", map[string]any{"query": name})
	if err != nil {
		return "", err
	}
	type candidate struct {
		name string
		id   string
	}
	var found []candidate
	// Search mixes cards, shelves and rows; any node pointing at an artist page
	// counts, whatever renderer wraps it.
	var walk func(n any)
	walk = func(n any) {
		switch value := n.(type) {
		case map[string]any:
			if id, ok := dig(value, "navigationEndpoint", "browseEndpoint", "browseId").(string); ok &&
				strings.HasPrefix(id, "UC") {
				pageType, _ := dig(value, "navigationEndpoint", "browseEndpoint",
					"browseEndpointContextSupportedConfigs", "browseEndpointContextMusicConfig", "pageType").(string)
				if strings.Contains(pageType, "ARTIST") {
					label := strings.TrimSpace(runsText(value["title"]))
					if label == "" {
						label = strings.TrimSpace(flexText(value, 0))
					}
					found = append(found, candidate{name: label, id: id})
				}
			}
			for _, child := range value {
				walk(child)
			}
		case []any:
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(res)
	if len(found) == 0 {
		return "", errors.New("no encontramos ese artista en YouTube Music")
	}
	lower := strings.ToLower(name)
	seen := map[string]bool{}
	var best string
	for _, item := range found {
		if seen[item.id] {
			continue
		}
		seen[item.id] = true
		// An exact name match wins over whatever appeared first.
		if strings.ToLower(item.name) == lower {
			return item.id, nil
		}
		if best == "" {
			best = item.id
		}
	}
	if best == "" {
		return "", errors.New("no encontramos ese artista en YouTube Music")
	}
	return best, nil
}
