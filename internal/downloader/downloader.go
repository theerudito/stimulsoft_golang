package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/theerudito/stimulsoft_golang/internal/checksum"
	"github.com/theerudito/stimulsoft_golang/internal/config"
	"github.com/theerudito/stimulsoft_golang/internal/platform"
)

var downloadMu sync.Mutex

type Downloader struct {
	Config                 config.Config
	Client                 *http.Client
	CacheDir, GOOS, GOARCH string
}

func New(cfg config.Config, client *http.Client) *Downloader {
	return &Downloader{Config: cfg, Client: client, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
}

func (d *Downloader) RunnerPath(ctx context.Context) (string, error) {
	if d.Config.RunnerPath != "" {
		return d.Config.RunnerPath, nil
	}
	asset, err := platform.Asset(d.GOOS, d.GOARCH)
	if err != nil {
		return "", err
	}
	cache := d.CacheDir
	if cache == "" {
		cache, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("user cache directory: %w", err)
		}
	}
	dir := filepath.Join(cache, "stimulsoft-runner", d.Config.Version)
	destination := filepath.Join(dir, asset)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create runner cache: %w", err)
	}
	downloadMu.Lock()
	defer downloadMu.Unlock()
	expected := d.Config.SHA256
	if expected == "" && d.Config.ChecksumURL != "" {
		expected, _ = checksum.Fetch(ctx, d.client(), d.Config.ChecksumURL, asset)
	}
	if checksum.ValidFile(destination, expected) && (!d.Config.RequireChecksum || expected != "") {
		return destination, nil
	}
	if err := d.download(ctx, destination, asset, expected); err != nil {
		return "", err
	}
	return destination, nil
}

func (d *Downloader) download(ctx context.Context, destination, asset, expected string) error {
	url := d.Config.DownloadBase() + "/" + asset
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create runner download request: %w", err)
	}
	resp, err := d.client().Do(req)
	if err != nil {
		return fmt.Errorf("download runner: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("runner release asset not found: %s; publish the release or set STIMULSOFT_RUNNER_PATH to a local self-contained runner", url)
		}
		return fmt.Errorf("download runner: HTTP %s", resp.Status)
	}
	tmp, err := os.CreateTemp(filepath.Dir(destination), ".runner-*")
	if err != nil {
		return fmt.Errorf("create runner temporary file: %w", err)
	}
	name := tmp.Name()
	defer os.Remove(name)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		tmp.Close()
		return fmt.Errorf("read runner: %w", err)
	}
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write runner: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close runner: %w", err)
	}
	actual, err := checksum.SHA256(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("hash runner: %w", err)
	}
	if d.Config.RequireChecksum && expected == "" {
		return fmt.Errorf("runner checksum is required but was not provided")
	}
	if expected != "" && !equalHash(expected, actual) {
		return fmt.Errorf("runner checksum mismatch")
	}
	if err = platform.MakeExecutable(name, d.GOOS == "windows"); err != nil {
		return fmt.Errorf("make runner executable: %w", err)
	}
	if err = os.Rename(name, destination); err != nil {
		return fmt.Errorf("install runner atomically: %w", err)
	}
	_ = platform.RemoveQuarantine(destination, d.GOOS == "darwin")
	return nil
}

func (d *Downloader) client() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return http.DefaultClient
}
func equalHash(a, b string) bool { return len(a) == len(b) && strings.EqualFold(a, b) }
