// Package ytaccount reads the YouTube Music session that lives in the app's
// WebView2 profile and uses it to read the signed-in account.
//
// The session is the WebView's own: the user signs in once in the login window,
// the profile keeps the cookies, and this package decrypts them with DPAPI to
// sign InnerTube requests. Nothing is copied into SQLite, config files or logs.
package ytaccount

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Errors the UI can tell apart.
var (
	ErrNotSignedIn = errors.New("no hay una sesión de YouTube Music")
	ErrLocked      = errors.New("el perfil está en uso por la ventana de inicio de sesión")
	ErrUnavailable = errors.New("YouTube Music no respondió")
)

const (
	origin    = "https://music.youtube.com"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/157.0.0.0 Safari/537.36 Edg/157.0.0.0"
)

// Account is what the app shows about the signed-in user.
type Account struct {
	Name      string
	Email     string
	Playlists int
}

// Playlist is one entry of the account's library.
type Playlist struct {
	ID       string
	Title    string
	Subtitle string
	Cover    string
}

// Client reads the profile and talks to InnerTube with its cookies.
type Client struct {
	profile string
	http    *http.Client
	origin  string
	webURL  string
	cookies func() ([]cookie, error)

	mu        sync.Mutex
	version   string
	versionAt time.Time
}

// New points the client at a WebView2 profile root (the directory that holds
// Local State and the EBWebView tree).
func New(profile string) *Client {
	c := &Client{
		profile: profile,
		http:    &http.Client{Timeout: 30 * time.Second},
		origin:  origin,
		webURL:  origin + "/",
	}
	c.cookies = func() ([]cookie, error) { return readCookies(profile) }
	return c
}

// Available reports whether a profile exists at all, without touching the DB.
func (c *Client) Available() bool {
	return profileExists(c.profile)
}

var clientVersionRe = regexp.MustCompile(`"INNERTUBE_CLIENT_VERSION":"([^"]+)"`)

// clientVersion must be current: an old value makes InnerTube answer as an
// anonymous client even with valid cookies. The signed-in page does not inline
// ytcfg, so this asks anonymously.
func (c *Client) clientVersion(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.version != "" && time.Since(c.versionAt) < 6*time.Hour {
		version := c.version
		c.mu.Unlock()
		return version, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.webURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	res, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("%w: leer la página: %v", ErrUnavailable, err)
	}
	match := clientVersionRe.FindSubmatch(body)
	if match == nil {
		return "", fmt.Errorf("%w: no encontré la versión del cliente", ErrUnavailable)
	}
	version := string(match[1])
	c.mu.Lock()
	c.version = version
	c.versionAt = time.Now()
	c.mu.Unlock()
	return version, nil
}

func (c *Client) call(ctx context.Context, path string, payload map[string]any) (map[string]any, error) {
	jar, err := c.cookies()
	if err != nil {
		return nil, err
	}
	sapisid := findOn(jar, "SAPISID", "youtube.com")
	if sapisid == "" {
		sapisid = findOn(jar, "__Secure-3PAPISID", "youtube.com")
	}
	if sapisid == "" {
		return nil, ErrNotSignedIn
	}
	version, err := c.clientVersion(ctx)
	if err != nil {
		return nil, err
	}

	ts := time.Now().Unix()
	sum := sha1.Sum([]byte(fmt.Sprintf("%d %s %s", ts, sapisid, c.origin)))

	body, err := json.Marshal(map[string]any{
		"context": map[string]any{
			"client": map[string]any{
				"clientName":    "WEB_REMIX",
				"clientVersion": version,
				"hl":            "es",
				"gl":            "US",
			},
		},
		"payload": payload,
	})
	if err != nil {
		return nil, err
	}
	// Flatten the payload into the envelope.
	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	delete(envelope, "payload")
	for key, value := range payload {
		envelope[key] = value
	}
	body, err = json.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.origin+"/youtubei/v1/"+path+"?prettyPrint=false", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-AuthUser", "0")
	req.Header.Set("X-Origin", c.origin)
	req.Header.Set("Cookie", cookieHeader(jar))
	req.Header.Set("Authorization", fmt.Sprintf("SAPISIDHASH %d_%x", ts, sum))
	req.Header.Set("User-Agent", userAgent)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: leer la respuesta: %v", ErrUnavailable, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrUnavailable, res.StatusCode)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: respuesta ilegible", ErrUnavailable)
	}
	return out, nil
}

// SignedIn asks InnerTube for the account menu; the name is the proof that the
// cookies are accepted.
func (c *Client) SignedIn(ctx context.Context) (Account, error) {
	res, err := c.call(ctx, "account/account_menu", map[string]any{})
	if err != nil {
		return Account{}, err
	}
	header := dig(res, "actions", 0, "openPopupAction", "popup", "multiPageMenuRenderer", "header", "activeAccountHeaderRenderer")
	name := runsText(dig(header, "accountName"))
	if name == "" {
		return Account{}, ErrNotSignedIn
	}
	email := runsText(dig(header, "email"))
	if email == "" {
		email = runsText(dig(header, "channelHandle"))
	}
	return Account{Name: name, Email: email}, nil
}

// Playlists reads the playlists saved in the account's library.
func (c *Client) Playlists(ctx context.Context) ([]Playlist, error) {
	res, err := c.call(ctx, "browse", map[string]any{"browseId": "FEmusic_liked_playlists"})
	if err != nil {
		return nil, err
	}
	items, _ := dig(res, "contents", "singleColumnBrowseResultsRenderer", "tabs", 0,
		"tabRenderer", "content", "sectionListRenderer", "contents", 0, "gridRenderer", "items").([]any)

	out := make([]Playlist, 0, len(items))
	for _, item := range items {
		row, _ := dig(item, "musicTwoRowItemRenderer").(map[string]any)
		if row == nil {
			continue
		}
		id := strings.TrimPrefix(browseID(row), "VL")
		// LM (liked music) and SE (saved episodes) are auto-generated shelves,
		// not playlists the user can import as a list.
		if id == "" || id == "LM" || id == "SE" {
			continue
		}
		out = append(out, Playlist{
			ID:       id,
			Title:    runsText(row["title"]),
			Subtitle: runsJoin(row["subtitle"]),
			Cover:    thumbnail(row),
		})
	}
	return out, nil
}

func browseID(row map[string]any) string {
	if id, ok := dig(row, "navigationEndpoint", "browseEndpoint", "browseId").(string); ok {
		return id
	}
	if id, ok := dig(row, "navigationEndpoint", "watchEndpoint", "playlistId").(string); ok {
		return id
	}
	return ""
}

func thumbnail(row map[string]any) string {
	list, _ := dig(row, "thumbnailRenderer", "musicThumbnailRenderer", "thumbnail", "thumbnails").([]any)
	if len(list) == 0 {
		return ""
	}
	last, _ := list[len(list)-1].(map[string]any)
	url, _ := last["url"].(string)
	return url
}

func dig(value any, path ...any) any {
	cur := value
	for _, step := range path {
		switch key := step.(type) {
		case string:
			m, ok := cur.(map[string]any)
			if !ok {
				return nil
			}
			cur = m[key]
		case int:
			s, ok := cur.([]any)
			if !ok || key >= len(s) {
				return nil
			}
			cur = s[key]
		}
	}
	return cur
}

func runsText(value any) string {
	return runsJoin(value)
}

func runsJoin(value any) string {
	m, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	runs, ok := m["runs"].([]any)
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(runs))
	for _, run := range runs {
		if text, ok := run.(map[string]any)["text"].(string); ok {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}
