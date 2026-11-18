package app

import (
	"errors"
	"time"

	"github.com/LuiisDev21/stefani/internal/ytaccount"
)

const (
	// After signing in Google lands on the account page, not on the YouTube
	// Music interface: the window is only for the login.
	loginURL  = "https://accounts.google.com/ServiceLogin?service=youtube&continue=https%3A%2F%2Fmyaccount.google.com%2F"
	logoutURL = "https://www.youtube.com/logout"
)

// AccountStatus reports the Google session that lives in the app's WebView
// profile. Locked means the login window is still open, so the cookies cannot
// be read yet.
func (a *App) AccountStatus() AccountDTO {
	if a.account == nil {
		return AccountDTO{}
	}
	ctx, cancel := a.opCtx()
	defer cancel()

	account, err := a.account.SignedIn(ctx)
	if errors.Is(err, ytaccount.ErrLocked) {
		return AccountDTO{Locked: true}
	}
	if err != nil {
		return AccountDTO{}
	}
	lists, err := a.account.Playlists(ctx)
	if err != nil {
		return AccountDTO{Connected: true, Name: account.Name, Email: account.Email, Avatar: account.Avatar}
	}
	return AccountDTO{
		Connected: true,
		Name:      account.Name,
		Email:     account.Email,
		Avatar:    account.Avatar,
		Playlists: len(lists),
	}
}

// ConnectAccount opens the Google login inside the app. It returns as soon as
// the window is up; FinishAccount is what imports the session.
func (a *App) ConnectAccount() error {
	if a.openLogin == nil {
		return errors.New("el inicio de sesión no está disponible en este entorno")
	}
	return a.openLogin(loginURL)
}

// SignOutAccount opens Google's logout page in the login window; FinishAccount
// then sees a profile without a session.
func (a *App) SignOutAccount() error {
	if a.openLogin == nil {
		return errors.New("el inicio de sesión no está disponible en este entorno")
	}
	return a.openLogin(logoutURL)
}

// FinishAccount closes the login window and waits for the resulting session.
func (a *App) FinishAccount() AccountDTO {
	if a.closeLogin != nil {
		_ = a.closeLogin()
	}
	return a.RefreshAccount()
}

// RefreshAccount retries until WebView2 releases the profile database, so a
// session that was already signed in is picked up instead of reported as
// locked.
func (a *App) RefreshAccount() AccountDTO {
	for attempt := 0; attempt < 24; attempt++ {
		status := a.AccountStatus()
		if !status.Locked {
			return status
		}
		time.Sleep(500 * time.Millisecond)
	}
	return a.AccountStatus()
}

// AccountPlaylists lists the playlists saved in the signed-in library.
func (a *App) AccountPlaylists() ([]AccountPlaylistDTO, error) {
	if a.account == nil {
		return nil, nil
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	lists, err := a.account.Playlists(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AccountPlaylistDTO, 0, len(lists))
	for _, item := range lists {
		out = append(out, AccountPlaylistDTO{
			ID:       item.ID,
			Title:    item.Title,
			Subtitle: item.Subtitle,
			Cover:    item.Cover,
		})
	}
	return out, nil
}

// ImportAccountPlaylist imports one of the account's playlists as a local list.
func (a *App) ImportAccountPlaylist(id, name string) (ImportDTO, error) {
	if !a.currentConfig().Features.YouTubeMusic {
		return ImportDTO{Notice: "El catalogo de YouTube Music esta desactivado en Ajustes"}, nil
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	return a.importPlaylist(ctx, id, name)
}
