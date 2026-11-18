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
	ArtistID string `json:"artistId"`
}

// ShelfDTO is one row of the home feed.
type ShelfDTO struct {
	Title string         `json:"title"`
	Items []ShelfItemDTO `json:"items"`
}

// CollectionDTO is an album, a playlist or an artist page.
type CollectionDTO struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Subtitle string     `json:"subtitle"`
	Artist   string     `json:"artist"`
	ArtistID string     `json:"artistId"`
	Cover    string     `json:"cover"`
	Kind     string     `json:"kind"`
	Meta     string     `json:"meta"`
	Tracks   []TrackDTO `json:"tracks"`
	Shelves  []ShelfDTO `json:"shelves"`
}

// HomeFeed reads the rows the account sees on the YouTube Music home.
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
		dto := ShelfDTO{Title: shelf.Title, Items: make([]ShelfItemDTO, 0, len(shelf.Items))}
		for _, item := range shelf.Items {
			dto.Items = append(dto.Items, ShelfItemDTO{
				Title:    item.Title,
				Subtitle: item.Subtitle,
				Cover:    item.Cover,
				VideoID:  item.VideoID,
				ListID:   item.ListID,
				ArtistID: item.ArtistID,
			})
		}
		out = append(out, dto)
	}
	return out, nil
}

// OpenCollection reads an album, a playlist or an artist and stores its tracks
// so every row can be streamed straight away.
func (a *App) OpenCollection(id string) (CollectionDTO, error) {
	id = strings.TrimSpace(id)
	if id == "" || a.account == nil {
		return CollectionDTO{}, errors.New("no hay nada que abrir")
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	col, err := a.account.Collection(ctx, id)
	if err != nil {
		return CollectionDTO{}, err
	}
	dto := CollectionDTO{
		ID:       col.ID,
		Title:    col.Title,
		Subtitle: col.Subtitle,
		Artist:   col.Artist,
		ArtistID: col.ArtistID,
		Cover:    col.Cover,
		Kind:     col.Kind,
		Meta:     col.Meta,
		Tracks:   make([]TrackDTO, 0, len(col.Tracks)),
		Shelves:  make([]ShelfDTO, 0, len(col.Shelves)),
	}
	for _, track := range col.Tracks {
		savedID, err := a.db.SaveTrack(ctx, track)
		if err != nil {
			continue
		}
		track.ID = savedID
		dto.Tracks = append(dto.Tracks, toTrackDTO(track, false, false, track.ProviderID != ""))
	}
	for _, shelf := range col.Shelves {
		row := ShelfDTO{Title: shelf.Title, Items: make([]ShelfItemDTO, 0, len(shelf.Items))}
		for _, item := range shelf.Items {
			row.Items = append(row.Items, ShelfItemDTO{
				Title:    item.Title,
				Subtitle: item.Subtitle,
				Cover:    item.Cover,
				VideoID:  item.VideoID,
				ListID:   item.ListID,
				ArtistID: item.ArtistID,
			})
		}
		dto.Shelves = append(dto.Shelves, row)
	}
	return dto, nil
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

// PlayCollection resolves a collection and plays it from the first song.
func (a *App) PlayCollection(collectionID string) ([]TrackDTO, error) {
	col, err := a.OpenCollection(collectionID)
	if err != nil {
		return nil, err
	}
	return col.Tracks, nil
}

// OpenArtist finds an artist by name and opens its page, which is how a song in
// the player leads to the rest of the catalogue.
func (a *App) OpenArtist(name string) (CollectionDTO, error) {
	name = strings.TrimSpace(name)
	if name == "" || a.account == nil {
		return CollectionDTO{}, errors.New("no hay artista que abrir")
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	id, err := a.account.FindArtist(ctx, name)
	if err != nil {
		return CollectionDTO{}, err
	}
	return a.OpenCollection(id)
}
