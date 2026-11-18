//go:build !windows

package ytaccount

import "errors"

// The account session lives in the WebView2 profile, which only exists on
// Windows. Other platforms report the feature as unavailable.

type cookie struct {
	name   string
	value  string
	host   string
	path   string
	secure bool
}

func profileExists(string) bool { return false }

func readCookies(string) ([]cookie, error) {
	return nil, errors.New("la cuenta de YouTube Music solo está disponible en Windows")
}

func cookieHeader([]cookie) string { return "" }

func findOn([]cookie, string, string) string { return "" }
