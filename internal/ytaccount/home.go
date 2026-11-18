package ytaccount

import (
	"context"
	"strings"
)

// Shelf is one row of the home feed: a heading and its cards.
type Shelf struct {
	Title string
	Items []ShelfItem
}

// ShelfItem is one card. VideoID plays a song; ListID opens a playlist.
type ShelfItem struct {
	Title    string
	Subtitle string
	Cover    string
	VideoID  string
	ListID   string
	ArtistID string
}

// Home reads the personalised rows YouTube Music builds for the account:
// forgotten favourites, song mixes and what to listen to next, plus the new
// releases of the artists the user follows.
func (c *Client) Home(ctx context.Context) ([]Shelf, error) {
	var out []Shelf
	seen := map[string]bool{}
	var firstErr error
	for _, browseID := range []string{"FEmusic_home", "FEmusic_new_releases"} {
		res, err := c.call(ctx, "browse", map[string]any{"browseId": browseID})
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		var shelves []map[string]any
		collectShelves(res, &shelves)
		for _, raw := range shelves {
			title := shelfTitle(raw)
			items := shelfItems(raw)
			if title == "" || len(items) < 2 || seen[title] {
				continue
			}
			seen[title] = true
			out = append(out, Shelf{Title: title, Items: items})
		}
	}
	if len(out) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

func collectShelves(node any, out *[]map[string]any) {
	switch value := node.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "musicCarouselShelfRenderer" {
				if shelf, ok := child.(map[string]any); ok {
					*out = append(*out, shelf)
				}
			}
			collectShelves(child, out)
		}
	case []any:
		for _, child := range value {
			collectShelves(child, out)
		}
	}
}

func shelfTitle(shelf map[string]any) string {
	header, _ := dig(shelf, "header", "musicCarouselShelfBasicHeaderRenderer").(map[string]any)
	if header == nil {
		return ""
	}
	return strings.TrimSpace(runsText(header["title"]))
}

func shelfItems(shelf map[string]any) []ShelfItem {
	contents, _ := shelf["contents"].([]any)
	items := make([]ShelfItem, 0, len(contents))
	for _, raw := range contents {
		if row, ok := dig(raw, "musicTwoRowItemRenderer").(map[string]any); ok {
			item := ShelfItem{
				Title:    strings.TrimSpace(runsText(row["title"])),
				Subtitle: strings.TrimSpace(runsJoin(row["subtitle"])),
				Cover:    thumbnail(row),
			}
			item.VideoID, _ = dig(row, "navigationEndpoint", "watchEndpoint", "videoId").(string)
			if id, ok := dig(row, "navigationEndpoint", "watchEndpoint", "playlistId").(string); ok {
				// RD* is the auto-radio of a song, not a list the user can open.
				if trimmed := strings.TrimPrefix(id, "VL"); !strings.HasPrefix(trimmed, "RD") {
					item.ListID = trimmed
				}
			}
			// Albums come as a browse endpoint; PlaylistTracks reads those too.
			if id, ok := dig(row, "navigationEndpoint", "browseEndpoint", "browseId").(string); ok {
				if strings.HasPrefix(id, "MPRE") || strings.HasPrefix(id, "VL") {
					item.ListID = strings.TrimPrefix(id, "VL")
				}
			}
			// Cards without artwork look broken, and the "listen together"
			// stations are live streams the player cannot seek.
			item.ArtistID = artistFromRuns(row["subtitle"])
			// A card that leads nowhere (radio mixes) cannot be opened or played.
			if item.VideoID == "" && item.ListID == "" {
				continue
			}
			if item.Title == "" || item.Cover == "" || isLive(row) {
				continue
			}
			items = append(items, item)
			continue
		}
		row, ok := dig(raw, "musicResponsiveListItemRenderer").(map[string]any)
		if !ok {
			continue
		}
		item := ShelfItem{
			Title:    strings.TrimSpace(flexText(row, 0)),
			Subtitle: strings.TrimSpace(flexText(row, 1)),
			Cover:    thumbnail(row),
		}
		item.VideoID, _ = dig(row, "flexColumns", 0, "musicResponsiveListItemFlexColumnRenderer",
			"text", "runs", 0, "navigationEndpoint", "watchEndpoint", "videoId").(string)
		item.ArtistID = artistFromRuns(dig(row, "flexColumns", 0, "musicResponsiveListItemFlexColumnRenderer", "text"))
		if item.Title == "" || item.Cover == "" || isLive(row) {
			continue
		}
		items = append(items, item)
	}
	return items
}

// isLive reports whether a card is a live station rather than a song or a
// record; those streams come back unplayable through the media proxy.
func isLive(row map[string]any) bool {
	badges, _ := row["subtitleBadges"].([]any)
	for _, raw := range badges {
		if _, ok := dig(raw, "liveBadgeRenderer").(map[string]any); ok {
			return true
		}
	}
	return false
}

// flexText joins the runs of one flex column.
func flexText(row map[string]any, index int) string {
	text, _ := dig(row, "flexColumns", index, "musicResponsiveListItemFlexColumnRenderer", "text").(map[string]any)
	if text == nil {
		return ""
	}
	// The subtitle carries the artist and the play count; the card shows the
	// artist and drops the counts, which read as noise in a shelf.
	return strings.SplitN(runsJoin(text), " • ", 2)[0]
}

// artistFromRuns pulls the artist page id out of a card subtitle, so the card
// can send the user to the artist.
func artistFromRuns(text any) string {
	run, ok := firstRun(text)
	if !ok {
		return ""
	}
	pageType, _ := dig(run, "navigationEndpoint", "browseEndpoint",
		"browseEndpointContextSupportedConfigs", "browseEndpointContextMusicConfig", "pageType").(string)
	if !strings.Contains(pageType, "ARTIST") {
		return ""
	}
	id, _ := dig(run, "navigationEndpoint", "browseEndpoint", "browseId").(string)
	return id
}
