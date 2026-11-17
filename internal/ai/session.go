package ai

import "strings"

// SessionID builds the conversation identifier used by providers that need one
// (OpenCode Go). It stays stable for the same scope, which is what prompt
// caching keys on, and never carries personal data.
func SessionID(clientID, scope string) string {
	client := strings.TrimSpace(clientID)
	if client == "" {
		client = "stefani"
	}
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "main"
	}
	return client + "-" + scope
}
