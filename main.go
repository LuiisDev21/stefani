package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/LuiisDev21/stefani/internal/ai"
	"github.com/LuiisDev21/stefani/internal/app"
	"github.com/LuiisDev21/stefani/internal/brand"
	"github.com/LuiisDev21/stefani/internal/config"
	"github.com/LuiisDev21/stefani/internal/demo"
	"github.com/LuiisDev21/stefani/internal/lyrics"
	"github.com/LuiisDev21/stefani/internal/media"
	"github.com/LuiisDev21/stefani/internal/music"
	"github.com/LuiisDev21/stefani/internal/phonetics"
	"github.com/LuiisDev21/stefani/internal/secret"
	"github.com/LuiisDev21/stefani/internal/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

// assetFS lets the web build serve the frontend from disk instead of the copy
// embedded at compile time: `pnpm build` then shows up on reload, which is what
// makes a visual review loop fast. Only STEFANI_WEB_DIST turns it on.
func assetFS(embedded embed.FS) fs.FS {
	if dir := os.Getenv("STEFANI_WEB_DIST"); dir != "" {
		return os.DirFS(dir)
	}
	return embedded
}

// webHost is loopback unless the operator asks for another interface. Opening
// the app to the LAN has no authentication, so it must be explicit.
func webHost() string {
	if host := os.Getenv("STEFANI_WEB_HOST"); host != "" {
		return host
	}
	return "127.0.0.1"
}

// settingsState is the shared settings snapshot for services that need the live
// configuration outside the app service (the phonetics engine reads the model
// and key on every call).
type settingsState struct {
	mu      sync.RWMutex
	cfg     config.Config
	secrets secret.Store
}

func (s *settingsState) set(cfg config.Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

// llm implements phonetics.ConfigFunc.
func (s *settingsState) llm() (config.LLM, string, error) {
	s.mu.RLock()
	cfg := s.cfg
	s.mu.RUnlock()
	key, err := s.secrets.Get(secret.KeyLLMAPIKey)
	if err != nil {
		return cfg.LLM, "", err
	}
	return cfg.LLM, key, nil
}

func (s *settingsState) youtubeMusic() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Features.YouTubeMusic
}

// wailsPicker opens native file dialogs for the app service.
type wailsPicker struct{}

func (p wailsPicker) attach(d *application.OpenFileDialogStruct) *application.OpenFileDialogStruct {
	if window := application.Get().Window.Current(); window != nil {
		d.AttachToWindow(window)
	}
	return d
}

func (p wailsPicker) PickAudioFiles() ([]string, error) {
	d := p.attach(application.Get().Dialog.OpenFile().
		SetTitle("Elige archivos de audio").
		SetMessage("Se agregan como referencia local; los archivos no se copian.").
		AddFilter("Audio", "*.mp3;*.m4a;*.aac;*.wav;*.flac;*.ogg;*.opus"))
	return d.PromptForMultipleSelection()
}

func (p wailsPicker) PickLyricsFile() (string, error) {
	d := p.attach(application.Get().Dialog.OpenFile().
		SetTitle("Elige un archivo de letra").
		AddFilter("Letras", "*.lrc;*.txt"))
	return d.PromptForSingleSelection()
}

// wailsOpener sends links to the system browser.
type wailsOpener struct{}

func (wailsOpener) OpenURL(rawURL string) error {
	return application.Get().Browser.OpenURL(rawURL)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "stefani: %v\n", err)
		os.Exit(1)
	}
}

// keyringService returns the credential namespace. Profiles that must not share
// the user's real credentials (E2E runs, portable installs) override it with
// STEFANI_KEYRING_SERVICE.
func keyringService() string {
	if override := strings.TrimSpace(os.Getenv("STEFANI_KEYRING_SERVICE")); override != "" {
		return override
	}
	return brand.DataFolder
}

func run() error {
	var paths config.Paths
	var err error
	if override := strings.TrimSpace(os.Getenv("STEFANI_HOME")); override != "" {
		pathOpts, pathErr := config.ResolvePathsAt(override)
		paths, err = pathOpts, pathErr
	} else {
		paths, err = config.ResolvePaths(brand.DataFolder)
	}
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(filepath.Join(paths.Logs, "stefani.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()
	logger := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, logFile),
		&slog.HandlerOptions{Level: slog.LevelInfo}))

	db, err := storage.Open(paths.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	cfgStore := config.NewStore(paths.Config)
	cfg, err := cfgStore.Load()
	if err != nil {
		// A corrupt config must not brick the app: keep the file for the user
		// and start from defaults.
		logger.Warn("config could not be read, using defaults", "error", err)
		cfg = config.Default()
	}
	secrets := secret.New(keyringService())
	// CI and offline runs must not depend on YouTube: this forces the online
	// provider off without touching the user's saved settings.
	if os.Getenv("STEFANI_DISABLE_ONLINE") == "1" {
		cfg.Features.YouTubeMusic = false
	}
	state := &settingsState{cfg: cfg, secrets: secrets}

	// The generated catalog is a test fixture now: production starts empty and
	// the UI never presents those songs as content.
	if os.Getenv("STEFANI_SEED_DEMO") == "1" {
		demoLib := demo.New(paths.DemoDir)
		startupCtx, cancelStartup := context.WithTimeout(context.Background(), 90*time.Second)
		if err := demoLib.Ensure(startupCtx, db); err != nil {
			logger.Error("demo fixture could not be prepared", "error", err)
		}
		cancelStartup()
	}

	yt := music.NewYTMusic(12 * time.Second)
	streamer := music.DetectYtDlp()
	localPhon := phonetics.NewLocal()
	musicSvc := music.NewService(db, yt, logger, state.youtubeMusic)
	lyricSvc := lyrics.NewService(db, lyrics.NewLRCLIB(10*time.Second), logger)
	aiClient := ai.New(45 * time.Second)
	phonSvc := phonetics.NewService(db, aiClient, state.llm, logger)

	appSvc := app.New(app.Deps{
		DB:             db,
		Configs:        cfgStore,
		Config:         cfg,
		Secrets:        secrets,
		Music:          musicSvc,
		Lyrics:         lyricSvc,
		Phonetics:      phonSvc,
		AI:             aiClient,
		Streamer:       streamer,
		LocalPhon:      localPhon,
		Log:            logger,
		Opener:         wailsOpener{},
		Picker:         wailsPicker{},
		OnConfigChange: state.set,
	})

	mediaHandler := media.StreamHandler(func(r *http.Request, trackID string) (media.Source, error) {
		return appSvc.ResolveAudioSource(trackID)
	})
	assetMiddleware := application.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/media/") {
				mediaHandler.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	geometry := appSvc.WindowState()
	wailsApp := application.New(application.Options{
		Name:        brand.Name,
		Description: brand.Tagline,
		Services: []application.Service{
			application.NewService(appSvc),
		},
		Assets: application.AssetOptions{
			Handler:    application.AssetFileServerFS(assetFS(assets)),
			Middleware: assetMiddleware,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			// Opt-in only: a debugging port must never be open in normal use.
			AdditionalBrowserArgs: webviewDebugArgs(),
		},
		Server: application.ServerOptions{
			Host: webHost(),
			Port: 8080,
		},
	})

	window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            brand.Name,
		Width:            geometry.Width,
		Height:           geometry.Height,
		MinWidth:         960,
		MinHeight:        620,
		BackgroundColour: application.NewRGB(11, 14, 22),
		URL:              "/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 38,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})
	if geometry.X != 0 || geometry.Y != 0 {
		window.SetPosition(geometry.X, geometry.Y)
	}
	if geometry.Maximised {
		window.Maximise()
	}

	saveGeometry := func() {
		w, h := window.Size()
		x, y := window.Position()
		_ = appSvc.SaveWindowState(app.WindowStateInput{
			Width: w, Height: h, X: x, Y: y, Maximised: window.IsMaximised(),
		})
	}
	// WindowClosing fires while the window still has a valid size; OnShutdown is
	// the backup for quits that bypass the close button.
	window.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		saveGeometry()
	})
	wailsApp.OnShutdown(saveGeometry)

	logger.Info("starting", "name", brand.Name, "version", brand.Version, "db", db.Path(),
		"online", streamer.Detail())
	return wailsApp.Run()
}

// webviewDebugArgs exposes the WebView2 debugging port only when the operator
// asks for it, which is what lets an automated check inspect the renderer.
func webviewDebugArgs() []string {
	if os.Getenv("STEFANI_DEBUG_WEBVIEW") == "1" {
		return []string{"--remote-debugging-port=9222"}
	}
	return nil
}
