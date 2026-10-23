package checksum

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ParseManifest(data []byte, asset string) string {
	var values map[string]string
	if json.Unmarshal(data, &values) == nil {
		return strings.TrimSpace(values[asset])
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && filepath.Base(fields[1]) == asset {
			return fields[0]
		}
	}
	return ""
}

func Fetch(ctx context.Context, client *http.Client, url, asset string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create checksum request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch checksum manifest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch checksum manifest: HTTP %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read checksum manifest: %w", err)
	}
	return ParseManifest(data, asset), nil
}

func ValidFile(path, expected string) bool {
	if expected == "" {
		_, err := os.Stat(path)
		return err == nil
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	actual, err := SHA256(f)
	return err == nil && strings.EqualFold(actual, expected)
}
