package storage

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a prefixed random identifier. IDs are opaque to the UI, so a
// random value is enough and avoids coordinating sequences across tables.
func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing means the OS entropy source is broken; there is
		// no safe fallback for identifiers, so panic rather than collide.
		panic("storage: cannot read random bytes: " + err.Error())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
