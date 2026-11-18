package ytaccount

import (
	"encoding/json"
	"testing"
)

// listPage is the shape YouTube Music returns for a playlist: one playable row
// and one row YouTube will not serve (no playlistItemData).
const listPage = `{
  "contents": {
    "singleColumnBrowseResultsRenderer": {
      "tabs": [{
        "tabRenderer": {
          "content": {
            "sectionListRenderer": {
              "contents": [{
                "musicPlaylistShelfRenderer": {
                  "contents": [
                    {
                      "musicResponsiveListItemRenderer": {
                        "playlistItemData": {"videoId": "abc123"},
                        "flexColumns": [
                          {"musicResponsiveListItemFlexColumnRenderer": {"text": {"runs": [{"text": "Wake Me Up"}]}}},
                          {"musicResponsiveListItemFlexColumnRenderer": {"text": {"runs": [
                            {"text": "Avicii"}, {"text": " • "}, {"text": "True"}, {"text": " • "}, {"text": "4:09"}
                          ]}}}
                        ],
                        "thumbnail": {"musicThumbnailRenderer": {"thumbnail": {"thumbnails": [
                          {"url": "https://img/1.jpg", "width": 60}, {"url": "https://img/2.jpg", "width": 226}
                        ]}}}
                      }
                    },
                    {
                      "musicResponsiveListItemRenderer": {
                        "flexColumns": [
                          {"musicResponsiveListItemFlexColumnRenderer": {"text": {"runs": [{"text": "Private Song"}]}}},
                          {"musicResponsiveListItemFlexColumnRenderer": {"text": {"runs": [{"text": "Someone"}]}}}
                        ]
                      }
                    }
                  ]
                }
              }]
            }
          }
        }
      }]
    }
  }
}`

func TestPlaylistTracksKeepsUnavailableRows(t *testing.T) {
	var doc map[string]any
	if err := json.Unmarshal([]byte(listPage), &doc); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	rows := collectPlaylistRows(doc)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	first, ok := playlistTrack(rows[0])
	if !ok {
		t.Fatal("first row rejected")
	}
	if first.ProviderID != "abc123" || first.Title != "Wake Me Up" {
		t.Errorf("first = %+v", first)
	}
	if first.Artist != "Avicii" || first.Album != "True" || first.DurationMs != 249000 {
		t.Errorf("first metadata = artist %q album %q duration %d", first.Artist, first.Album, first.DurationMs)
	}
	if first.ExternalURL != "https://music.youtube.com/watch?v=abc123" {
		t.Errorf("external = %q", first.ExternalURL)
	}
	if first.ArtworkURL != "https://img/2.jpg" {
		t.Errorf("artwork = %q", first.ArtworkURL)
	}

	second, ok := playlistTrack(rows[1])
	if !ok {
		t.Fatal("unavailable row rejected")
	}
	if second.ProviderID != "" {
		t.Errorf("unavailable row must have no video id, got %q", second.ProviderID)
	}
	if second.Title != "Private Song" || second.Artist != "Someone" {
		t.Errorf("second = %+v", second)
	}
	if second.ExternalURL != "" {
		t.Errorf("unavailable row must not link out, got %q", second.ExternalURL)
	}
}

func TestSplitSubtitle(t *testing.T) {
	cases := []struct {
		in       string
		artist   string
		album    string
		duration int64
	}{
		{"Avicii • True • 4:09", "Avicii", "True", 249000},
		{"Avicii • 1:02:03", "Avicii", "", 3723000},
		{"Avicii", "Avicii", "", 0},
		{"", "", "", 0},
	}
	for _, c := range cases {
		artist, album, duration := splitSubtitle(c.in)
		if artist != c.artist || album != c.album || duration != c.duration {
			t.Errorf("splitSubtitle(%q) = %q, %q, %d", c.in, artist, album, duration)
		}
	}
}
