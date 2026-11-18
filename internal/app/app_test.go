package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/demo"
	"github.com/LuiisDev21/stefani/internal/domain"
	"github.com/LuiisDev21/stefani/internal/lyrics"
	"github.com/LuiisDev21/stefani/internal/media"
	"github.com/LuiisDev21/stefani/internal/music"
	"github.com/LuiisDev21/stefani/internal/phonetics"
	"github.com/LuiisDev21/stefani/internal/secret"
	"github.com/LuiisDev21/stefani/internal/storage"
)

type testEnv struct {
	app     *App
	db      *storage.DB
	secrets *secret.Memory
	configs *config.Store
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	// Tests stay offline: online playback has its own cases below.
	cfg := config.Default()
	cfg.Features.YouTubeMusic = false
	return newTestEnvFull(t, cfg, music.NewYtDlpWithRunner("", nil, nil, 0), music.NewYTMusic(time.Second))
}

// newTestEnvWith lets a test choose the feature flags and the stream resolver,
// so online playback is exercised without yt-dlp installed.
func newTestEnvWith(t *testing.T, cfg config.Config, streamer *music.YtDlp) testEnv {
	t.Helper()
	return newTestEnvFull(t, cfg, streamer, music.NewYTMusic(time.Second))
}

func newTestEnvFull(t *testing.T, cfg config.Config, streamer *music.YtDlp, catalog music.Catalog) testEnv {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(filepath.Join(dir, "app.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	logger := slog.New(slog.NewTextHandler(testWriter{t}, &slog.HandlerOptions{Level: slog.LevelError}))
	demoLib := demo.New(filepath.Join(dir, "demo"))
	if err := demoLib.Ensure(context.Background(), db); err != nil {
		t.Fatalf("demo ensure: %v", err)
	}

	cfgStore := config.NewStore(filepath.Join(dir, "config.json"))
	secrets := secret.NewMemory()

	phoneticsCfg := func() (config.LLM, string, error) {
		key, err := secrets.Get(secret.KeyLLMAPIKey)
		if err != nil {
			return cfg.LLM, "", err
		}
		return cfg.LLM, key, nil
	}

	svc := New(Deps{
		DB:        db,
		Configs:   cfgStore,
		Config:    cfg,
		Secrets:   secrets,
		Music:     music.NewService(db, catalog, logger, func() bool { return cfg.Features.YouTubeMusic }),
		Lyrics:    lyrics.NewService(db, nil, logger),
		Phonetics: phonetics.NewService(db, ai.New(time.Second), phoneticsCfg, logger),
		AI:        ai.New(time.Second),
		Streamer:  streamer,
		LocalPhon: phonetics.NewLocal(),
		Log:       logger,
	})
	return testEnv{app: svc, db: db, secrets: secrets, configs: cfgStore}
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(strings.TrimSpace(string(p)))
	return len(p), nil
}

func TestBootstrapLoadsSeededLibrary(t *testing.T) {
	env := newTestEnv(t)
	data, err := env.app.Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if data.Counts.Tracks != len(demo.Songs) {
		t.Fatalf("got %d tracks, want %d", data.Counts.Tracks, len(demo.Songs))
	}
	for _, track := range data.RecentlyPlayed {
		if !track.Playable || track.AudioURL == "" {
			t.Fatalf("demo track is not playable: %+v", track)
		}
	}
	if data.Counts.Tracks == 0 {
		t.Fatal("expected seeded annotations in the counts")
	}
}

func TestLyricsReportsMissingWithoutError(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	id, err := env.db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderLocal, ProviderID: "no-lyrics", Title: "Silence",
	})
	if err != nil {
		t.Fatalf("save track: %v", err)
	}
	dto, err := env.app.Lyrics(id, false)
	if err != nil {
		t.Fatalf("lyrics should not fail for a missing document: %v", err)
	}
	if dto.Available {
		t.Fatal("expected unavailable lyrics")
	}
	if dto.Notice == "" {
		t.Fatal("expected an explanatory notice")
	}
}

func TestDemoLyricsExposeWordTimingsAndAnnotations(t *testing.T) {
	env := newTestEnv(t)
	dto, err := env.app.Lyrics("demo:paper-boats", false)
	if err != nil {
		t.Fatalf("lyrics: %v", err)
	}
	if !dto.Available || !dto.HasWords {
		t.Fatalf("expected word-synced lyrics: %+v", dto)
	}
	if len(dto.Lines) == 0 || len(dto.Lines[0].Words) == 0 {
		t.Fatal("expected word timings")
	}
	if dto.Lines[0].Annotation == nil || dto.Lines[0].Annotation.IPA == "" {
		t.Fatal("expected a seeded annotation on the first line")
	}
}

func TestSaveSettingsValidation(t *testing.T) {
	env := newTestEnv(t)

	cases := []struct {
		name    string
		baseURL string
		wantErr string
	}{
		{"plain http to a remote host", "http://example.com/v1", "https"},
		{"not a url", "not a url", "no es valida"},
		{"empty", "", "obligatoria"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.app.SaveSettings(SettingsInput{
				LLM:      LLMSettingsDTO{BaseURL: tc.baseURL, Model: "m", Temperature: 0.2, TimeoutSeconds: 30, MaxRetries: 1},
				Features: FeaturesDTO{},
				Accent:   "en-US",
			})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want one containing %q", err, tc.wantErr)
			}
		})
	}

	settings, err := env.app.SaveSettings(SettingsInput{
		LLM: LLMSettingsDTO{
			BaseURL: "http://127.0.0.1:9099/v1", Model: "stub", Temperature: 0.1,
			TimeoutSeconds: 20, MaxRetries: 1, JSONMode: true,
		},
		Features: FeaturesDTO{YouTubeMusic: true},
		Accent:   "en-GB",
		APIKey:   "test-key",
	})
	if err != nil {
		t.Fatalf("save local endpoint: %v", err)
	}
	if !settings.HasKey {
		t.Fatal("expected the key to be reported as stored")
	}
	if settings.Accent != "en-GB" || !settings.Features.YouTubeMusic {
		t.Fatalf("settings not persisted: %+v", settings)
	}
	if stored, err := env.secrets.Get(secret.KeyLLMAPIKey); err != nil || stored != "test-key" {
		t.Fatalf("key not stored in the secret store: %q %v", stored, err)
	}
	// The key must never travel back to the frontend.
	if strings.Contains(settings.ConfigPath, "test-key") {
		t.Fatal("settings leaked the key")
	}
}

func TestAnnotateWithoutKeyAsksForSetup(t *testing.T) {
	env := newTestEnv(t)
	outcome, err := env.app.AnnotateTrack("demo:paper-boats")
	if err != nil {
		t.Fatalf("annotate should degrade, not fail: %v", err)
	}
	if !outcome.NeedsSetup {
		t.Fatalf("expected needsSetup, got %+v", outcome)
	}
	if outcome.Message == "" {
		t.Fatal("expected an explanatory message")
	}
}

func TestLibraryToggle(t *testing.T) {
	env := newTestEnv(t)

	saved, err := env.app.SetInLibrary("demo:slow-morning", true)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if !saved.InLibrary {
		t.Fatalf("expected the track to be in the library: %+v", saved)
	}
	list, err := env.app.Library()
	if err != nil {
		t.Fatalf("library: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID == "demo:slow-morning" {
			found = true
		}
	}
	if !found {
		t.Fatalf("saved track missing from the library: %+v", list)
	}

	removed, err := env.app.SetInLibrary("demo:slow-morning", false)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if removed.InLibrary {
		t.Fatal("expected the track to be removed")
	}
}

func TestSearchFindsDemoTracks(t *testing.T) {
	env := newTestEnv(t)
	result, err := env.app.Search("paper", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Title != "Paper Boats" {
		t.Fatalf("unexpected results: %+v", result.Items)
	}
	if result.Items[0].AudioURL == "" {
		t.Fatal("search results must carry the playable URL")
	}
}

func TestImportDroppedFilesSkipsNonAudio(t *testing.T) {
	env := newTestEnv(t)
	result, err := env.app.ImportDroppedFiles([]string{"/tmp/notes.txt"})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(result.Imported) != 0 || len(result.Skipped) != 1 {
		t.Fatalf("unexpected import result: %+v", result)
	}
}

func TestParsePlaylistURL(t *testing.T) {
	good := "https://music.youtube.com/playlist?list=OLAK5uy_abcdefghijk"
	id, err := parsePlaylistURL(good)
	if err != nil || id != "OLAK5uy_abcdefghijk" {
		t.Fatalf("parse %q = %q, %v", good, id, err)
	}
	if _, err := parsePlaylistURL("https://example.com/playlist?list=x"); err == nil {
		t.Fatal("expected a rejection for a non-YouTube host")
	}
	if _, err := parsePlaylistURL("https://music.youtube.com/watch?v=abc"); err == nil {
		t.Fatal("expected a rejection when the list parameter is missing")
	}
}

func TestWindowStateRoundTrip(t *testing.T) {
	env := newTestEnv(t)
	initial := env.app.WindowState()
	if initial.Width == 0 || initial.Height == 0 {
		t.Fatalf("expected default geometry, got %+v", initial)
	}
	if err := env.app.SaveWindowState(WindowStateInput{Width: 1280, Height: 800, X: 40, Y: 60}); err != nil {
		t.Fatalf("save window state: %v", err)
	}
	saved := env.app.WindowState()
	if saved.Width != 1280 || saved.Height != 800 || saved.X != 40 || saved.Y != 60 {
		t.Fatalf("window state not persisted: %+v", saved)
	}
}

func TestOpenExternalRejectsUnknownSchemes(t *testing.T) {
	env := newTestEnv(t)
	if err := env.app.OpenExternal("file:///etc/passwd"); err == nil {
		t.Fatal("expected file:// to be rejected")
	}
	if err := env.app.OpenExternal("javascript:alert(1)"); err == nil {
		t.Fatal("expected javascript: to be rejected")
	}
}

const testStreamURL = "https://rr1---sn-test.googlevideo.com/videoplayback?expire=1893456000&itag=251"

func fakeStreamer(t *testing.T) *music.YtDlp {
	t.Helper()
	return music.NewYtDlpWithRunner("yt-dlp", func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
		return []byte(testStreamURL + "\n"), nil, nil
	}, nil, time.Hour)
}

func onlineTrack(t *testing.T, env testEnv) string {
	t.Helper()
	id, err := env.db.SaveTrack(context.Background(), domain.Track{
		Provider:   domain.ProviderYTMusic,
		ProviderID: "dQw4w9WgXcQ",
		Title:      "Never Gonna Give You Up",
		Artist:     "Rick Astley",
		Album:      "Whenever You Need Somebody",
		DurationMs: 213000,
	})
	if err != nil {
		t.Fatalf("save online track: %v", err)
	}
	return id
}

func TestOnlinePlaybackRequiresYtDlp(t *testing.T) {
	cfg := config.Default()
	env := newTestEnvWith(t, cfg, music.NewYtDlpWithRunner("", nil, nil, 0))
	id := onlineTrack(t, env)

	dto, err := env.app.Track(id)
	if err != nil {
		t.Fatalf("track: %v", err)
	}
	if dto.Playable || dto.AudioURL != "" {
		t.Fatalf("online track must not be playable without yt-dlp: %+v", dto)
	}
	if _, err := env.app.ResolveAudioSource(id); !errors.Is(err, media.ErrNoAudio) {
		t.Fatalf("resolver error = %v, want ErrNoAudio", err)
	}
	if info := env.app.Info(); info.Online.Available {
		t.Fatalf("online status should be unavailable: %+v", info.Online)
	}
}

func TestOnlinePlaybackStreamsWithYtDlp(t *testing.T) {
	cfg := config.Default()
	env := newTestEnvWith(t, cfg, fakeStreamer(t))
	id := onlineTrack(t, env)

	dto, err := env.app.Track(id)
	if err != nil {
		t.Fatalf("track: %v", err)
	}
	if !dto.Playable || dto.AudioURL != "/media/"+id {
		t.Fatalf("expected a playable stream url: %+v", dto)
	}
	src, err := env.app.ResolveAudioSource(id)
	if err != nil {
		t.Fatalf("resolve source: %v", err)
	}
	if src.URL != testStreamURL || src.Path != "" {
		t.Fatalf("unexpected source: %+v", src)
	}
	info := env.app.Info()
	if !info.Online.Available || !info.Online.Enabled {
		t.Fatalf("online status = %+v", info.Online)
	}
}

func TestOnlinePlaybackHonoursTheFlag(t *testing.T) {
	cfg := config.Default()
	cfg.Features.YouTubeMusic = false
	env := newTestEnvWith(t, cfg, fakeStreamer(t))
	id := onlineTrack(t, env)

	dto, _ := env.app.Track(id)
	if dto.Playable {
		t.Fatalf("flag off must disable online playback: %+v", dto)
	}
	if _, err := env.app.ResolveAudioSource(id); !errors.Is(err, media.ErrNoAudio) {
		t.Fatalf("resolver error = %v, want ErrNoAudio", err)
	}
}

func TestLocalTracksKeepStreamingWithoutOnline(t *testing.T) {
	cfg := config.Default()
	cfg.Features.YouTubeMusic = false
	env := newTestEnvWith(t, cfg, music.NewYtDlpWithRunner("", nil, nil, 0))

	if _, err := env.app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	track, err := env.app.Track("demo:paper-boats")
	if err != nil {
		t.Fatalf("seeded track: %v", err)
	}
	src, err := env.app.ResolveAudioSource(track.ID)
	if err != nil {
		t.Fatalf("demo source: %v", err)
	}
	if src.Path == "" || src.URL != "" {
		t.Fatalf("demo tracks must resolve to a file: %+v", src)
	}
}

// fakeCatalog returns canned remote results and records the call.
type fakeCatalog struct {
	results []domain.SearchResult
	calls   int
	fail    bool
}

func (f *fakeCatalog) Name() string { return "fake" }

func (f *fakeCatalog) BrowsePlaylist(ctx context.Context, playlistID string, limit int) ([]domain.Track, error) {
	return nil, nil
}

func (f *fakeCatalog) Search(ctx context.Context, query string, limit int) ([]domain.SearchResult, error) {
	f.calls++
	if f.fail {
		return nil, errors.New("provider down")
	}
	return f.results, nil
}

func TestSearchPersistsOnlineResultsWithStableID(t *testing.T) {
	cfg := config.Default()
	catalog := &fakeCatalog{results: []domain.SearchResult{{
		Track: domain.Track{
			Provider:   domain.ProviderYTMusic,
			ProviderID: "dQw4w9WgXcQ",
			Title:      "Never Gonna Give You Up",
			Artist:     "Rick Astley",
			DurationMs: 213000,
		},
		ProviderName: "ytmusic",
		Playable:     false,
	}}}
	env := newTestEnvFull(t, cfg, fakeStreamer(t), catalog)

	first, err := env.app.Search("never gonna", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(first.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(first.Items))
	}
	item := first.Items[0]
	if item.ID == "" {
		t.Fatal("remote results must be persisted to get a stable id")
	}
	if !item.Playable || item.AudioURL != "/media/"+item.ID {
		t.Fatalf("online result should be playable with yt-dlp: %+v", item)
	}
	if item.Note != "" {
		t.Fatalf("playable results must not carry the catalogue note: %q", item.Note)
	}

	second, err := env.app.Search("never gonna", 5)
	if err != nil {
		t.Fatalf("second search: %v", err)
	}
	if len(second.Items) != 1 || second.Items[0].ID != item.ID {
		t.Fatalf("the id is not stable: %+v", second.Items)
	}
}

func TestSearchDegradesWhenTheProviderFails(t *testing.T) {
	cfg := config.Default()
	env := newTestEnvFull(t, cfg, fakeStreamer(t), &fakeCatalog{fail: true})
	data, err := env.app.Search("paper", 5)
	if err != nil {
		t.Fatalf("a failing provider must not fail the search: %v", err)
	}
	if len(data.Items) != 1 || data.Items[0].Provider != "demo" {
		t.Fatalf("local results should survive: %+v", data.Items)
	}
}

func TestLyricsFillLocalPronunciationWithoutAI(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	id, err := env.db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderLocal, ProviderID: "local-1", Title: "Slow study", Artist: "Test",
	})
	if err != nil {
		t.Fatalf("save track: %v", err)
	}
	if err := env.db.SaveLyrics(ctx, domain.Lyrics{
		TrackID: id, Provider: "lrclib", Language: "en", Sync: domain.SyncLine,
		Source: "test#1", Lines: []domain.Line{
			{Index: 0, Text: "Take it slowly", StartMs: 0, EndMs: 2000},
			{Index: 1, Text: "zzzqx qqqzz wwwwk", StartMs: 2000, EndMs: 4000},
		},
	}); err != nil {
		t.Fatalf("save lyrics: %v", err)
	}

	dto, err := env.app.Lyrics(id, false)
	if err != nil {
		t.Fatalf("lyrics: %v", err)
	}
	if !dto.Available || len(dto.Lines) != 2 {
		t.Fatalf("unexpected document: %+v", dto)
	}
	first := dto.Lines[0].Annotation
	if first == nil {
		t.Fatal("a known line must get a local pronunciation without any AI key")
	}
	if first.Source != "local" || first.IPA != "/teɪk ɪt ˈsloʊli/" || first.SpanishApprox != "téik it slóuli" {
		t.Fatalf("unexpected local annotation: %+v", first)
	}
	if first.Translation != "" {
		t.Fatalf("local pronunciation must not invent a translation: %+v", first)
	}
	if dto.Lines[1].Annotation != nil {
		t.Fatalf("an unknown line must not get fake pronunciation: %+v", dto.Lines[1].Annotation)
	}
}

func TestGeneratedAnnotationsWinOverLocal(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	id, err := env.db.SaveTrack(ctx, domain.Track{
		Provider: domain.ProviderLocal, ProviderID: "local-2", Title: "Already annotated",
	})
	if err != nil {
		t.Fatalf("save track: %v", err)
	}
	if err := env.db.SaveLyrics(ctx, domain.Lyrics{
		TrackID: id, Provider: "lrclib", Language: "en", Sync: domain.SyncLine,
		Lines: []domain.Line{{Index: 0, Text: "Take it slowly", StartMs: 0, EndMs: 2000}},
	}); err != nil {
		t.Fatalf("save lyrics: %v", err)
	}
	saved, err := env.db.Lyrics(ctx, id)
	if err != nil || len(saved.Lines) != 1 {
		t.Fatalf("reload lyrics: %v", err)
	}
	if err := env.db.SaveAnnotation(ctx, domain.Annotation{
		LineID: saved.Lines[0].ID, Accent: "en-US", IPA: "/custom/", SpanishApprox: "custom",
		Translation: "Traducido", Model: "test-model",
	}); err != nil {
		t.Fatalf("save annotation: %v", err)
	}

	dto, err := env.app.Lyrics(id, false)
	if err != nil {
		t.Fatalf("lyrics: %v", err)
	}
	ann := dto.Lines[0].Annotation
	if ann == nil || ann.Source != "ai" || ann.IPA != "/custom/" || ann.Translation != "Traducido" {
		t.Fatalf("generated annotation should win: %+v", ann)
	}
}

func TestSearchRefreshesStaleOnlineMetadata(t *testing.T) {
	// A row written before the Songs filter existed has no artist and no
	// duration. The catalog now returns both, and the result must use the fresh
	// values instead of the stale row.
	cfg := config.Default()
	catalog := &fakeCatalog{results: []domain.SearchResult{{
		Track: domain.Track{
			Provider:    domain.ProviderYTMusic,
			ProviderID:  "dQw4w9WgXcQ",
			Title:       "Never Gonna Give You Up",
			Artist:      "Rick Astley",
			DurationMs:  214000,
			ExternalURL: "https://music.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		ProviderName: "YouTube Music",
	}}}
	env := newTestEnvFull(t, cfg, nil, catalog)

	stale := domain.Track{
		Provider: domain.ProviderYTMusic, ProviderID: "dQw4w9WgXcQ",
		Title: "Never Gonna Give You Up",
	}
	if _, err := env.db.SaveTrack(context.Background(), stale); err != nil {
		t.Fatalf("save stale: %v", err)
	}

	dto, err := env.app.Search("never gonna", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(dto.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(dto.Items))
	}
	if dto.Items[0].Artist != "Rick Astley" || dto.Items[0].DurationMs != 214000 {
		t.Fatalf("item = %+v, want the fresh catalog metadata", dto.Items[0])
	}
}

func TestTestLLMListsProviderModels(t *testing.T) {
	var modelsHit int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			w.Write([]byte(`{"choices":[{"message":{"content":"{\"ok\":true}"}}]}`))
		case "/models":
			modelsHit++
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf("authorization on /models = %q", got)
			}
			w.Write([]byte(`{"data":[{"id":"model-b"},{"id":"model-a"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	env := newTestEnv(t)
	env.app.mu.Lock()
	env.app.cfg.LLM.BaseURL = srv.URL
	env.app.cfg.LLM.Model = "model-a"
	env.app.cfg.LLM.Provider = "custom"
	env.app.cfg.LLM.TimeoutSeconds = 3
	env.app.mu.Unlock()
	if err := env.secrets.Set(secret.KeyLLMAPIKey, "test-key"); err != nil {
		t.Fatalf("secret: %v", err)
	}

	outcome, err := env.app.TestLLM()
	if err != nil {
		t.Fatalf("test llm: %v", err)
	}
	if !outcome.OK {
		t.Fatalf("outcome = %+v, want a successful connection", outcome)
	}
	if modelsHit != 1 {
		t.Fatalf("/models hit %d times, want 1", modelsHit)
	}
	if len(outcome.Models) != 2 || outcome.Models[0] != "model-a" || outcome.Models[1] != "model-b" {
		t.Fatalf("models = %v, want the ids the API returned", outcome.Models)
	}
}

func TestSaveSettingsKeepsOpenCodeSessionWiring(t *testing.T) {
	var gotSession string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSession = r.Header.Get("x-opencode-session")
		switch r.URL.Path {
		case "/chat/completions":
			w.Write([]byte(`{"choices":[{"message":{"content":"{\"ok\":true}"}}]}`))
		case "/models":
			w.Write([]byte(`{"data":[{"id":"m"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	env := newTestEnv(t)
	// A client id normally comes from Store.Load; set it explicitly here.
	env.app.mu.Lock()
	env.app.cfg.LLM.ClientID = "client-abc"
	env.app.mu.Unlock()
	if err := env.secrets.Set(secret.KeyLLMAPIKey, "test-key"); err != nil {
		t.Fatalf("secret: %v", err)
	}

	if _, err := env.app.SaveSettings(SettingsInput{
		LLM: LLMSettingsDTO{
			Provider:       "opencode-go",
			BaseURL:        srv.URL,
			Model:          "deepseek-v4.1-flash",
			TimeoutSeconds: 5,
			MaxRetries:     0,
		},
	}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	if provider := env.app.currentConfig().LLM.Provider; provider != "opencode-go" {
		t.Fatalf("provider = %q after save, want it persisted", provider)
	}
	if clientID := env.app.currentConfig().LLM.ClientID; clientID != "client-abc" {
		t.Fatalf("client id = %q, want it preserved across saves", clientID)
	}

	outcome, err := env.app.TestLLM()
	if err != nil || !outcome.OK {
		t.Fatalf("test llm: %+v %v", outcome, err)
	}
	if gotSession != "client-abc-ping" {
		t.Fatalf("x-opencode-session = %q, want the installation id plus the scope", gotSession)
	}
}
