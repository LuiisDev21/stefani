package music

import "testing"

func TestCleanSubtitleDropsResultKinds(t *testing.T) {
	// YouTube puts the result kind in the second column of many rows; showing it
	// as the artist ("Song", "Episode") is wrong and breaks lyric matching.
	for _, raw := range []string{"Song", "song", "Video", "Episode", "Album", "Playlist", "267K plays", "1.2M views"} {
		if got := cleanSubtitle(raw); got != "" {
			t.Fatalf("cleanSubtitle(%q) = %q, want empty", raw, got)
		}
	}
}

func TestCleanSubtitleKeepsArtistsAndSplitsBullets(t *testing.T) {
	cases := map[string]string{
		"Rick Astley": "Rick Astley",
		"Rick Astley • Whenever You Need Somebody": "Rick Astley",
		"Queen • A Night At The Opera":             "Queen",
		"  Stefani Studio  ":                       "Stefani Studio",
	}
	for raw, want := range cases {
		if got := cleanSubtitle(raw); got != want {
			t.Fatalf("cleanSubtitle(%q) = %q, want %q", raw, got, want)
		}
	}
}
