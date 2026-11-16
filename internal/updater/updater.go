package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultLatestReleaseURL = "https://api.github.com/repos/eliel-elie/ai-commits/releases/latest"
	checkCacheFileName      = ".ai-commits-update-check"
	checkInterval           = 24 * time.Hour
	backgroundCheckTimeout  = 1500 * time.Millisecond
)

// ReleaseInfo represents the subset of GitHub Release API fields we need.
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type updateCache struct {
	LastChecked time.Time `json:"last_checked"`
	LatestTag   string    `json:"latest_tag"`
}

// ExpectedAssetName returns the release binary name for the current OS and architecture.
func ExpectedAssetName() string {
	return assetNameFor(runtime.GOOS, runtime.GOARCH)
}

func assetNameFor(goos, goarch string) string {
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("aic-%s-%s%s", goos, goarch, ext)
}

// CompareSemVer returns true if latest is strictly greater than current.
func CompareSemVer(current, latest string) (bool, error) {
	cMaj, cMin, cPat, err := parseSemVer(current)
	if err != nil {
		return false, err
	}
	lMaj, lMin, lPat, err := parseSemVer(latest)
	if err != nil {
		return false, err
	}

	if lMaj != cMaj {
		return lMaj > cMaj, nil
	}
	if lMin != cMin {
		return lMin > cMin, nil
	}
	return lPat > cPat, nil
}

func parseSemVer(v string) (int, int, int, error) {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver %q", v)
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	pat, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}
	return maj, min, pat, nil
}

// FetchLatestRelease queries the GitHub Releases API for the latest release.
func FetchLatestRelease(ctx context.Context, client *http.Client, apiURL string) (*ReleaseInfo, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ai-commits-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("github API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode release JSON: %w", err)
	}
	if info.TagName == "" {
		return nil, fmt.Errorf("release has empty tag_name")
	}
	return &info, nil
}

// CheckOnly checks whether a newer version is available without downloading or installing it.
func CheckOnly(ctx context.Context, currentVersion string) (latestTag string, hasUpdate bool, err error) {
	info, err := FetchLatestRelease(ctx, nil, DefaultLatestReleaseURL)
	if err != nil {
		return "", false, err
	}
	hasUpdate, err = CompareSemVer(currentVersion, info.TagName)
	if err != nil {
		return info.TagName, false, err
	}
	return info.TagName, hasUpdate, nil
}

// SelfUpdate downloads and replaces the running executable if a newer release is available.
func SelfUpdate(ctx context.Context, currentVersion string) (latestTag string, updated bool, err error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", false, fmt.Errorf("locate current executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	return selfUpdateTarget(ctx, currentVersion, DefaultLatestReleaseURL, exePath, ExpectedAssetName())
}

func selfUpdateTarget(ctx context.Context, currentVersion, apiURL, exePath, assetName string) (string, bool, error) {
	client := &http.Client{Timeout: 2 * time.Minute}
	info, err := FetchLatestRelease(ctx, client, apiURL)
	if err != nil {
		return "", false, err
	}

	hasUpdate, err := CompareSemVer(currentVersion, info.TagName)
	if err != nil {
		return info.TagName, false, err
	}
	if !hasUpdate {
		return info.TagName, false, nil
	}

	var downloadURL string
	for _, asset := range info.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return info.TagName, false, fmt.Errorf("no binary asset %q found in release %s", assetName, info.TagName)
	}

	dir := filepath.Dir(exePath)
	tmpFile, err := os.CreateTemp(dir, "aic-update-*.tmp")
	if err != nil {
		return info.TagName, false, fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		_ = tmpFile.Close()
		return info.TagName, false, err
	}
	req.Header.Set("User-Agent", "ai-commits-updater")

	resp, err := client.Do(req)
	if err != nil {
		_ = tmpFile.Close()
		return info.TagName, false, fmt.Errorf("download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = tmpFile.Close()
		return info.TagName, false, fmt.Errorf("download binary HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return info.TagName, false, fmt.Errorf("write temp binary: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return info.TagName, false, err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			return info.TagName, false, fmt.Errorf("chmod binary: %w", err)
		}
	}

	if err := replaceBinary(exePath, tmpPath); err != nil {
		return info.TagName, false, err
	}
	cleanupTmp = false
	return info.TagName, true, nil
}

func replaceBinary(exePath, tmpPath string) error {
	if runtime.GOOS == "windows" {
		oldPath := exePath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(exePath, oldPath); err != nil {
			return fmt.Errorf("move current binary aside: %w", err)
		}
		if err := os.Rename(tmpPath, exePath); err != nil {
			// Roll back if moving the new binary fails.
			_ = os.Rename(oldPath, exePath)
			return fmt.Errorf("install new binary: %w", err)
		}
		_ = os.Remove(oldPath)
		return nil
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

// CleanupOldBinary removes any leftover .old executable left behind by a Windows self-update.
func CleanupOldBinary() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}
	_ = os.Remove(exePath + ".old")
}

// StartBackgroundCheck performs a non-blocking 24h-debounced update check and returns the newer tag if found.
func StartBackgroundCheck(ctx context.Context, currentVersion string) <-chan string {
	ch := make(chan string, 1)
	if os.Getenv("AIC_NO_UPDATE_CHECK") != "" || currentVersion == "" || currentVersion == "dev" {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		cachePath := filepath.Join(home, checkCacheFileName)
		if tag := checkWithCache(ctx, currentVersion, DefaultLatestReleaseURL, cachePath, time.Now()); tag != "" {
			ch <- tag
		}
	}()

	return ch
}

func checkWithCache(ctx context.Context, currentVersion, apiURL, cachePath string, now time.Time) string {
	if raw, err := os.ReadFile(cachePath); err == nil {
		var cached updateCache
		if json.Unmarshal(raw, &cached) == nil && now.Sub(cached.LastChecked) < checkInterval {
			if hasUpdate, err := CompareSemVer(currentVersion, cached.LatestTag); err == nil && hasUpdate {
				return cached.LatestTag
			}
			return ""
		}
	}

	reqCtx, cancel := context.WithTimeout(ctx, backgroundCheckTimeout)
	defer cancel()

	client := &http.Client{Timeout: backgroundCheckTimeout}
	info, err := FetchLatestRelease(reqCtx, client, apiURL)
	if err != nil {
		return ""
	}

	encoded, err := json.Marshal(updateCache{
		LastChecked: now,
		LatestTag:   info.TagName,
	})
	if err == nil {
		_ = os.WriteFile(cachePath, encoded, 0o600)
	}

	if hasUpdate, err := CompareSemVer(currentVersion, info.TagName); err == nil && hasUpdate {
		return info.TagName
	}
	return ""
}
