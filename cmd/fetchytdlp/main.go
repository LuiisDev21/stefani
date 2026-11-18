// Command fetchytdlp downloads the standalone yt-dlp for this platform into a
// folder, so the installer can ship it and the app works offline from the start.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LuiisDev21/stefani/internal/music"
)

func main() {
	dir := "bin"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch yt-dlp:", err)
		os.Exit(1)
	}
	music.SetManagedDir(abs)

	client := music.DetectYtDlp()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	version, err := client.Update(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch yt-dlp:", err)
		os.Exit(1)
	}
	fmt.Printf("yt-dlp %s -> %s\n", version, abs)
}
