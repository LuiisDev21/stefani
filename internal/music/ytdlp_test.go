package music

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const sampleStreamURL = "https://rr1---sn-uxab05-hb5k.googlevideo.com/videoplayback?expire=1893456000&itag=251&id=o-abc"

func fakeRunner(url string, calls *atomic.Int32, stderr string, err error) Runner {
	return func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
		if calls != nil {
			calls.Add(1)
		}
		if err != nil {
			return nil, []byte(stderr), err
		}
		return []byte(url + "\n"), nil, nil
	}
}

func TestResolveRejectsUnavailableBinary(t *testing.T) {
	resolver := NewYtDlpWithRunner("", fakeRunner(sampleStreamURL, nil, "", nil), nil, 0)
	if resolver.Available() {
		t.Fatal("empty path must not be available")
	}
	_, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ")
	if !errors.Is(err, ErrYtDlpMissing) {
		t.Fatalf("error = %v, want ErrYtDlpMissing", err)
	}
}

func TestResolveReturnsStreamURLAndCachesIt(t *testing.T) {
	var calls atomic.Int32
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner(sampleStreamURL, &calls, "", nil), nil, time.Hour)

	first, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if first.URL != sampleStreamURL {
		t.Fatalf("url = %q", first.URL)
	}
	if first.ExpiresAt.Before(time.Now().Add(23 * time.Hour)) {
		t.Fatalf("expiry should come from the expire parameter: %v", first.ExpiresAt)
	}

	second, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if second.URL != first.URL {
		t.Fatalf("cache returned a different url")
	}
	if calls.Load() != 1 {
		t.Fatalf("runner calls = %d, want 1 (cached)", calls.Load())
	}
}

func TestResolveExpiresCachedEntry(t *testing.T) {
	var calls atomic.Int32
	now := time.Now()
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner(sampleStreamURL, &calls, "", nil), func() time.Time { return now }, time.Minute)
	// The URL's own expire is far in the future, so force a short TTL instead.
	resolver.run = fakeRunner("https://x.googlevideo.com/a?itag=251", &calls, "", nil)

	if _, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ"); err != nil {
		t.Fatalf("first: %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ"); err != nil {
		t.Fatalf("second: %v", err)
	}
	// The fallback TTL only applies when the URL has no expire parameter, so the
	// cache must have been bypassed for the second call.
	if calls.Load() != 2 {
		t.Fatalf("runner calls = %d, want 2", calls.Load())
	}
}

func TestResolveValidatesVideoID(t *testing.T) {
	var calls atomic.Int32
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner(sampleStreamURL, &calls, "", nil), nil, 0)
	for _, id := range []string{"", "short", "https://evil.example/x", "aaaaaaaaaaaaaaaaaaaa"} {
		if _, err := resolver.Resolve(context.Background(), id); !errors.Is(err, ErrStreamResolve) {
			t.Fatalf("id %q: error = %v", id, err)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("runner must not run for invalid ids: %d calls", calls.Load())
	}
}

func TestResolveRejectsNonGoogleHost(t *testing.T) {
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner("https://evil.example/audio.m4a", nil, "", nil), nil, 0)
	_, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ")
	if !errors.Is(err, ErrStreamResolve) {
		t.Fatalf("error = %v, want ErrStreamResolve", err)
	}
	if !strings.Contains(err.Error(), "host no permitido") {
		t.Fatalf("error should explain the host rule: %v", err)
	}
}

func TestResolveRejectsPlainHTTP(t *testing.T) {
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner("http://x.googlevideo.com/a", nil, "", nil), nil, 0)
	if _, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ"); !errors.Is(err, ErrStreamResolve) {
		t.Fatalf("error = %v, want ErrStreamResolve", err)
	}
}

func TestResolveReportsFailureWithoutLeakingURLs(t *testing.T) {
	boom := errors.New("exit status 1")
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner("", nil, "ERROR: https://secret.example/token failed", boom), nil, 0)
	_, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ")
	if !errors.Is(err, ErrStreamResolve) {
		t.Fatalf("error = %v, want ErrStreamResolve", err)
	}
	if strings.Contains(err.Error(), "https://") {
		t.Fatalf("error leaks a URL: %v", err)
	}
	if !strings.Contains(err.Error(), "hxxps://") {
		t.Fatalf("error should keep the sanitized detail: %v", err)
	}
}

func TestResolveEmptyOutputIsAnError(t *testing.T) {
	resolver := NewYtDlpWithRunner("yt-dlp", fakeRunner("", nil, "", nil), nil, 0)
	if _, err := resolver.Resolve(context.Background(), "dQw4w9WgXcQ"); !errors.Is(err, ErrStreamResolve) {
		t.Fatalf("error = %v", err)
	}
}

func TestIsAllowedStreamHost(t *testing.T) {
	allowed := []string{"googlevideo.com", "rr1---sn-x.googlevideo.com", "RR1.GOOGLEVIDEO.COM"}
	for _, host := range allowed {
		if !IsAllowedStreamHost(host) {
			t.Fatalf("%q should be allowed", host)
		}
	}
	for _, host := range []string{"", "googlevideo.com.evil.example", "evil.example", "notgooglevideo.com"} {
		if IsAllowedStreamHost(host) {
			t.Fatalf("%q should be rejected", host)
		}
	}
}

func TestDetailExplainsMissingBinary(t *testing.T) {
	if got := NewYtDlpWithRunner("", nil, nil, 0).Detail(); got != "yt-dlp no encontrado" {
		t.Fatalf("detail = %q", got)
	}
	if got := NewYtDlpWithRunner("yt-dlp", fakeRunner("", nil, "", nil), nil, 0).Detail(); !strings.Contains(got, "yt-dlp") {
		t.Fatalf("detail = %q", got)
	}
}
