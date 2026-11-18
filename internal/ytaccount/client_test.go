package ytaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := New("unused")
	c.origin = server.URL
	c.webURL = server.URL + "/"
	c.cookies = func() ([]cookie, error) {
		return []cookie{
			{name: "SAPISID", value: "youtube-value", host: ".youtube.com", path: "/"},
			{name: "SAPISID", value: "google-value", host: ".google.com", path: "/"},
			{name: "SID", value: "sid-value", host: ".google.com", path: "/"},
		}, nil
	}
	return c
}

func TestSignedInReadsTheAccountName(t *testing.T) {
	var gotAuth, gotCookie string
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`<script>{"INNERTUBE_CLIENT_VERSION":"1.20261115.14.00"}</script>`))
			return
		}
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		name, _ := body["context"].(map[string]any)["client"].(map[string]any)["clientVersion"].(string)
		if name != "1.20261115.14.00" {
			t.Errorf("unexpected client version %q", name)
		}
		_, _ = w.Write([]byte(`{"actions":[{"openPopupAction":{"popup":{"multiPageMenuRenderer":{"header":{"activeAccountHeaderRenderer":{"accountName":{"runs":[{"text":"Luis Leo"}]}}}}}}}]}`))
	}))

	account, err := client.SignedIn(context.Background())
	if err != nil {
		t.Fatalf("signed in: %v", err)
	}
	if account.Name != "Luis Leo" {
		t.Fatalf("unexpected name %q", account.Name)
	}
	if !strings.HasPrefix(gotAuth, "SAPISIDHASH ") {
		t.Fatalf("missing signature: %q", gotAuth)
	}
	if !strings.Contains(gotCookie, "SAPISID=youtube-value") || strings.Contains(gotCookie, "google-value") {
		t.Fatalf("wrong cookie order or duplicate: %q", gotCookie)
	}
	if !strings.Contains(gotCookie, "SID=sid-value") {
		t.Fatalf("google cookies missing: %q", gotCookie)
	}
}

func TestSignedInWithoutNameIsNotSignedIn(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`"INNERTUBE_CLIENT_VERSION":"1.0"`))
			return
		}
		_, _ = w.Write([]byte(`{"actions":[]}`))
	}))
	if _, err := client.SignedIn(context.Background()); err != ErrNotSignedIn {
		t.Fatalf("want ErrNotSignedIn, got %v", err)
	}
}

func TestPlaylistsReadsTheGrid(t *testing.T) {
	client := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`"INNERTUBE_CLIENT_VERSION":"1.0"`))
			return
		}
		_, _ = w.Write([]byte(`{"contents":{"singleColumnBrowseResultsRenderer":{"tabs":[{"tabRenderer":{"content":{"sectionListRenderer":{"contents":[{"gridRenderer":{"items":[
			{"musicTwoRowItemRenderer":{"title":{"runs":[{"text":"Favoritas"}]},"subtitle":{"runs":[{"text":"12 canciones"}]},"navigationEndpoint":{"browseEndpoint":{"browseId":"VLPLabc"}},"thumbnailRenderer":{"musicThumbnailRenderer":{"thumbnail":{"thumbnails":[{"url":"https://img/1"}]}}}}},
			{"musicTwoRowItemRenderer":{"title":{"runs":[{"text":"Para correr"}]},"navigationEndpoint":{"browseEndpoint":{"browseId":"VLPLdef"}}}}
		]}}]}}}}]}}}`))
	}))
	list, err := client.Playlists(context.Background())
	if err != nil {
		t.Fatalf("playlists: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 playlists, got %d", len(list))
	}
	if list[0].ID != "PLabc" || list[0].Title != "Favoritas" || list[0].Subtitle != "12 canciones" || list[0].Cover != "https://img/1" {
		t.Fatalf("unexpected first playlist: %+v", list[0])
	}
}
