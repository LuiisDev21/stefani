package app

import (
	"errors"
	"strings"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// ShelfItemDTO is one card of a home row.
type ShelfItemDTO struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Cover    string `json:"cover"`
	VideoID  string `json:"videoId"`
	ListID   string `json:"listId"`
}

// ShelfDTO is one personalised row of the home feed.
type ShelfDTO struct {
	Title string         `json:"title"`
	Items []ShelfItemDTO `json:"items"`
}

// HomeFeed returns the rows YouTube Music builds for the signed-in account:
// forgotten favourites, song mixes, what to listen to next and new releases of
// the artists the user follows. Empty when there is no session.
func (a *App) HomeFeed() ([]ShelfDTO, error) {
	if a.account == nil {
		return nil, nil
	}
	ctx, cancel := a.opCtx()
	defer cancel()

	shelves, err := a.account.Home(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ShelfDTO, 0, len(shelves))
	for _, shelf := range shelves {
		items := make([]ShelfItemDTO, 0, len(shelf.Items))
		for _, item := range shelf.Items {
			items = append(items, ShelfItemDTO{
				Title:    item.Title,
				Subtitle: item.Subtitle,
				Cover:    item.Cover,
				VideoID:  item.VideoID,
				ListID:   item.ListID,
			})
		}
		out = append(out, ShelfDTO{Title: shelf.Title, Items: items})
	}
	return out, nil
}

// PlayShelfItem stores one recommendation and returns its track record, so the
// player can stream a song that is not in the library yet.
func (a *App) PlayShelfItem(videoID, title, artist, cover string) (TrackDTO, error) {
	videoID = strings.TrimSpace(videoID)
	if videoID == "" {
		return TrackDTO{}, errors.New("la recomendación no tiene video")
	}
	ctx, cancel := a.opCtx()
	defer cancel()

	track := domain.Track{
		Provider:    domain.ProviderYTMusic,
		ProviderID:  videoID,
		Title:       strings.TrimSpace(title),
		Artist:      strings.TrimSpace(artist),
		ArtworkURL:  cover,
		ExternalURL: "https://www.youtube.com/watch?v=" + videoID,
	}
	id, err := a.db.SaveTrack(ctx, track)
	if err != nil {
		return TrackDTO{}, err
	}
	track.ID = id
	return toTrackDTO(track, false, false, true), nil
}

// PlayCollection resolves an album or a list from the account and stores its
// tracks, so the player can queue the whole thing.
func (a *App) PlayCollection(collectionID string) ([]TrackDTO, error) {
	collectionID = strings.TrimSpace(collectionID)
	if collectionID == "" || a.account == nil {
		return nil, errors.New("no hay una colección que abrir")
	}
	ctx, cancel := a.opCtx()
	defer cancel()

	tracks, err := a.account.PlaylistTracks(ctx, collectionID)
	if err != nil {
		return nil, err
	}
	out := make([]TrackDTO, 0, len(tracks))
	for _, track := range tracks {
		id, err := a.db.SaveTrack(ctx, track)
		if err != nil {
			return nil, err
		}
		track.ID = id
		out = append(out, toTrackDTO(track, false, false, true))
	}
	return out, nil
}
