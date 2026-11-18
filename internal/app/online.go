package app

import "github.com/LuiisDev21/stefani/internal/music"

// UpdateYtDlp downloads the latest yt-dlp release into the app's bin folder and
// refreshes the detected binary, so the new version is used without a restart.
func (a *App) UpdateYtDlp() (OnlineDTO, error) {
	if a.streamer == nil {
		a.streamer = music.DetectYtDlp()
	}
	ctx, cancel := a.opCtx()
	defer cancel()
	if _, err := a.streamer.Update(ctx); err != nil {
		return OnlineDTO{}, err
	}
	a.streamer = music.DetectYtDlp()
	return OnlineDTO{
		Enabled:   a.currentConfig().Features.YouTubeMusic,
		Available: a.streamer.Available(),
		Version:   a.streamer.Version(),
		Detail:    a.streamer.Detail(),
	}, nil
}
