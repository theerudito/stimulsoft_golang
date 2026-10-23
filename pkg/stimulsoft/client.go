package stimulsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/theerudito/stimulsoft_golang/internal/config"
	"github.com/theerudito/stimulsoft_golang/internal/downloader"
	"github.com/theerudito/stimulsoft_golang/internal/runner"
)

func Export(ctx context.Context, request ExportRequest) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	path, err := downloader.New(cfg, http.DefaultClient).RunnerPath(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode export request: %w", err)
	}
	return runner.Execute(ctx, path, bytes.NewReader(payload))
}

// ExportPDF renders a local MRT path with in-memory JSON data as PDF.
func ExportPDF(ctx context.Context, mrtPath string, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportMixedFormat(ctx, mrtPath, data, "pdf", sourceName, licenseKey, optionalVariables(variables))
}

// ExportExcel renders a local MRT path with in-memory JSON data as Excel.
func ExportExcel(ctx context.Context, mrtPath string, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportMixedFormat(ctx, mrtPath, data, "excel", sourceName, licenseKey, optionalVariables(variables))
}

func exportMixedFormat(ctx context.Context, mrtPath string, data []byte, format, sourceName, licenseKey string, variables map[string]any) ([]byte, error) {
	if err := validateLocalFile(mrtPath, "mrt_path"); err != nil {
		return nil, err
	}
	if sourceName == "" {
		sourceName = "Data"
	}
	if len(data) == 0 {
		data = []byte("null")
	}
	if !json.Valid(data) {
		return nil, errors.New("data must be valid JSON")
	}
	return Export(ctx, ExportRequest{
		LicenseKey: licenseKey, MRTPath: mrtPath, Data: string(data),
		DataSourceName: sourceName, Format: format, Variables: variables,
	})
}

func optionalVariables(values []map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func validateLocalFile(path, field string) error {
	if path == "" {
		return fmt.Errorf("%s is required", field)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s must be a regular file", field)
	}
	return nil
}
