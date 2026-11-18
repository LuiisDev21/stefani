package app

import (
	"testing"

	"github.com/LuiisDev21/stefani/internal/domain"
)

// A row imported without a video id must reach the UI as unavailable: greyed
// out, not playable, and without an audio URL to hand to the player.
func TestToTrackDTOMarksUnavailable(t *testing.T) {
	dto := toTrackDTO(domain.Track{
		Provider:   domain.ProviderYTMusic,
		ProviderID: "na:0123456789abcdef",
		Title:      "Private Song",
	}, false, false, true)

	if !dto.Unavailable {
		t.Fatal("track with an na: id must be unavailable")
	}
	if dto.Playable {
		t.Error("unavailable track must not be playable")
	}
	if dto.AudioURL != "" {
		t.Errorf("unavailable track must have no audio url, got %q", dto.AudioURL)
	}
}

func TestToTrackDTOKeepsPlayableRows(t *testing.T) {
	dto := toTrackDTO(domain.Track{
		Provider:   domain.ProviderYTMusic,
		ProviderID: "abc123",
		Title:      "Wake Me Up",
	}, false, false, true)

	if dto.Unavailable {
		t.Error("row with a video id must stay available")
	}
	if !dto.Playable || dto.AudioURL == "" {
		t.Errorf("playable row lost its audio url: %+v", dto)
	}
}
