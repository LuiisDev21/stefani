//go:build windows

package ytaccount

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	_ "modernc.org/sqlite"
)

// cookie is one row of the WebView2 cookie store, already decrypted.
type cookie struct {
	name   string
	value  string
	host   string
	path   string
	secure bool
}

func profileExists(profile string) bool {
	_, err := os.Stat(filepath.Join(profile, "Local State"))
	return err == nil
}

func unprotect(data []byte) ([]byte, error) {
	var in, out windows.DataBlob
	in.Data = &data[0]
	in.Size = uint32(len(data))
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, 0, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return unsafe.Slice(out.Data, out.Size), nil
}

// readCookies decrypts the profile cookies. WebView2 uses the same v10 scheme
// as Chromium (DPAPI-protected master key + AES-GCM) but not App-Bound
// Encryption, so this works where `--cookies-from-browser` fails.
func readCookies(profile string) ([]cookie, error) {
	raw, err := os.ReadFile(filepath.Join(profile, "Local State"))
	if err != nil {
		return nil, ErrNotSignedIn
	}
	var state struct {
		OSCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("perfil ilegible: %w", err)
	}
	blob, err := base64.StdEncoding.DecodeString(state.OSCrypt.EncryptedKey)
	if err != nil || len(blob) < 6 {
		return nil, fmt.Errorf("clave del perfil ilegible")
	}
	if string(blob[:5]) != "DPAPI" {
		return nil, fmt.Errorf("método de cifrado no soportado: %q", blob[:5])
	}
	key, err := unprotect(blob[5:])
	if err != nil {
		return nil, fmt.Errorf("descifrar la clave del perfil: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	dsn := "file:" + filepath.ToSlash(filepath.Join(profile, "Default", "Network", "Cookies")) + "?mode=ro"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		if isLocked(err) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("abrir las cookies: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("select host_key, name, encrypted_value, path, is_secure from cookies")
	if err != nil {
		if isLocked(err) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("leer las cookies: %w", err)
	}
	defer rows.Close()

	var out []cookie
	for rows.Next() {
		var host, name, path string
		var enc []byte
		var secure bool
		if err := rows.Scan(&host, &name, &enc, &path, &secure); err != nil {
			return nil, err
		}
		if len(enc) < 15 || string(enc[:3]) != "v10" {
			continue
		}
		plain, err := gcm.Open(nil, enc[3:15], enc[15:], nil)
		if err != nil || len(plain) == 0 {
			continue
		}
		value := string(plain)
		// Chromium prefixes the plaintext with SHA256(host_key).
		if len(plain) > 32 {
			sum := sha256.Sum256([]byte(host))
			if bytes.Equal(sum[:], plain[:32]) {
				value = string(plain[32:])
			}
		}
		out = append(out, cookie{name: name, value: value, host: host, path: path, secure: secure})
	}
	if err := rows.Err(); err != nil {
		if isLocked(err) {
			return nil, ErrLocked
		}
		return nil, err
	}
	return out, nil
}

func isLocked(err error) bool {
	if errors.Is(err, sql.ErrConnDone) {
		return true
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "locked") || strings.Contains(text, "busy")
}
