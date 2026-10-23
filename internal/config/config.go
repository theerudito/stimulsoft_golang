package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	DefaultVersion    = "v0.1.5"
	DefaultOwner      = "example"
	DefaultRepo       = "stimulsoft-golang"
	DefaultReleaseURL = "https://github.com/theerudito/stimulsoft_golang/releases/download"
)

type Config struct {
	Owner, Repo, Version            string
	BaseURL, ReleaseURL             string
	RunnerPath, SHA256, ChecksumURL string
	RequireChecksum                 bool
	ReleaseURLComplete              bool
}

func Load() (Config, error) {
	c := Config{
		Owner:              envOr("STIMULSOFT_RUNNER_OWNER", DefaultOwner),
		Repo:               envOr("STIMULSOFT_RUNNER_REPO", DefaultRepo),
		Version:            envOr("STIMULSOFT_RUNNER_VERSION", DefaultVersion),
		BaseURL:            os.Getenv("STIMULSOFT_RUNNER_BASE_URL"),
		ReleaseURL:         envOr("STIMULSOFT_RUNNER_RELEASE_URL", DefaultReleaseURL),
		RunnerPath:         os.Getenv("STIMULSOFT_RUNNER_PATH"),
		SHA256:             "",
		ChecksumURL:        os.Getenv("STIMULSOFT_RUNNER_CHECKSUM_URL"),
		RequireChecksum:    os.Getenv("STIMULSOFT_RUNNER_REQUIRE_CHECKSUM") == "1",
		ReleaseURLComplete: os.Getenv("STIMULSOFT_RUNNER_RELEASE_URL") != "",
	}
	if values := strings.Fields(os.Getenv("STIMULSOFT_RUNNER_SHA256")); len(values) > 0 {
		c.SHA256 = values[0]
	}
	return c, c.Validate()
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Version) == "" {
		return fmt.Errorf("runner version is required")
	}
	if c.RunnerPath != "" {
		return nil
	}
	for name, value := range map[string]string{"release URL": c.ReleaseURL, "base URL": c.BaseURL} {
		if value != "" {
			if parsed, err := url.Parse(value); err != nil || parsed.Scheme == "" || parsed.Host == "" {
				return fmt.Errorf("invalid %s", name)
			}
		}
	}
	return nil
}

func (c Config) DownloadBase() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	base := strings.TrimRight(c.ReleaseURL, "/")
	if !c.ReleaseURLComplete {
		base += "/" + c.Version
	}
	return base
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
