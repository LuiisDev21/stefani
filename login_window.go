package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// runLoginWindow is the helper process behind the account login. It owns its
// own WebView profile, so closing it releases the cookie database for the main
// process to read.
func runLoginWindow(profileRoot, url string) {
	loginApp := application.New(application.Options{
		Name: "STEFANI",
		Windows: application.WindowsOptions{
			WebviewUserDataPath: profileRoot,
		},
	})

	window := loginApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "account-login",
		Title:            "Conectar con YouTube Music",
		URL:              url,
		Width:            560,
		Height:           780,
		MinWidth:         420,
		MinHeight:        560,
		BackgroundColour: application.NewRGB(11, 14, 22),
	})
	window.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		loginApp.Quit()
	})

	_ = loginApp.Run()
}

// copyTree clones a WebView profile once, so a session that already existed in
// the main profile keeps working after the login moved to its own process.
func copyTree(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
