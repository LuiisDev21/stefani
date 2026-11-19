package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/LuiisDev21/stefani/internal/storage"
)

const (
	audioSettingsKey   = "audio.settings"
	audioSettingsLimit = 16 << 10
	audioSettingsWait  = 3 * time.Second
)

// AudioSettingsDTO carries the equaliser and the spatial settings as a JSON
// document. The shape lives in the frontend, which is also the only place that
// can apply it, so the backend keeps it as an opaque, validated blob instead
// of duplicating a schema that would drift.
type AudioSettingsDTO struct {
	JSON string `json:"json"`
}

// AudioSettings returns the saved document, or an empty string on first run.
func (a *App) AudioSettings() (AudioSettingsDTO, error) {
	if a.db == nil {
		return AudioSettingsDTO{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), audioSettingsWait)
	defer cancel()
	raw, err := a.db.Setting(ctx, audioSettingsKey)
	if errors.Is(err, storage.ErrNotFound) {
		return AudioSettingsDTO{}, nil
	}
	if err != nil {
		return AudioSettingsDTO{}, err
	}
	return AudioSettingsDTO{JSON: raw}, nil
}

// SaveAudioSettings stores the document after checking it is JSON we could read
// back. A corrupt value must never reach the player, so nothing else is written.
func (a *App) SaveAudioSettings(in AudioSettingsDTO) (AudioSettingsDTO, error) {
	if a.db == nil {
		return AudioSettingsDTO{}, nil
	}
	if len(in.JSON) > audioSettingsLimit {
		return AudioSettingsDTO{}, fmt.Errorf("los ajustes de audio son demasiado grandes")
	}
	if in.JSON != "" && !json.Valid([]byte(in.JSON)) {
		return AudioSettingsDTO{}, fmt.Errorf("los ajustes de audio no son válidos")
	}
	if in.JSON == "" {
		return AudioSettingsDTO{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), audioSettingsWait)
	defer cancel()
	if err := a.db.SetSetting(ctx, audioSettingsKey, in.JSON); err != nil {
		return AudioSettingsDTO{}, err
	}
	return in, nil
}
