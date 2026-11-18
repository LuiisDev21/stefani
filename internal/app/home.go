package app

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
