package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCompareSemVer(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"1.0.0", "v1.0.1", true},
		{"1.0.0", "v1.1.0", true},
		{"1.0.0", "v2.0.0", true},
		{"1.0.0", "v1.0.0", false},
		{"1.2.0", "v1.1.9", false},
		{"v1.0.0", "1.0.1", true},
	}

	for _, tc := range tests {
		got, err := CompareSemVer(tc.current, tc.latest)
		if err != nil {
			t.Fatalf("CompareSemVer(%q, %q) unexpected error: %v", tc.current, tc.latest, err)
		}
		if got != tc.want {
			t.Errorf("CompareSemVer(%q, %q) = %v, want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestSelfUpdateTarget(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "aic.exe")
	if err := os.WriteFile(exePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	const newBinaryContent = "new-binary-v1.0.1"
	const assetName = "aic-windows-amd64.exe"

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag_name": "v1.0.1",
				"assets": []map[string]string{
					{
						"name":                 assetName,
						"browser_download_url": srv.URL + "/download",
					},
				},
			})
		case "/download":
			_, _ = w.Write([]byte(newBinaryContent))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tag, updated, err := selfUpdateTarget(context.Background(), "1.0.0", srv.URL+"/latest", exePath, assetName)
	if err != nil {
		t.Fatalf("selfUpdateTarget failed: %v", err)
	}
	if !updated || tag != "v1.0.1" {
		t.Fatalf("expected updated=true tag=v1.0.1, got updated=%v tag=%q", updated, tag)
	}

	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != newBinaryContent {
		t.Fatalf("expected binary %q, got %q", newBinaryContent, string(got))
	}
}

func TestCheckWithCacheDebounce(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, ".ai-commits-update-check")

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name": "v1.2.0",
		})
	}))
	defer srv.Close()

	now := time.Now()
	if tag := checkWithCache(context.Background(), "1.0.0", srv.URL, cachePath, now); tag != "v1.2.0" {
		t.Fatalf("expected v1.2.0 on first call, got %q", tag)
	}
	if calls != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", calls)
	}

	// Within 24h: should use cache without calling HTTP server again.
	if tag := checkWithCache(context.Background(), "1.0.0", srv.URL, cachePath, now.Add(2*time.Hour)); tag != "v1.2.0" {
		t.Fatalf("expected cached v1.2.0, got %q", tag)
	}
	if calls != 1 {
		t.Fatalf("expected cache hit (still 1 HTTP call), got %d", calls)
	}
}
