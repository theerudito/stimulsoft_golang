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

func ExportSimple(ctx context.Context, mrt, data, sourceName, licenseKey, format string, variables map[string]any) ([]byte, error) {
	return exportFormat(ctx, []byte(mrt), []byte(data), format, sourceName, licenseKey, variables)
}

// ExportPaths renders a report using the original local MRT and JSON files.
// Unlike the byte-based helpers, it does not read or copy either file in Go.
func ExportPaths(ctx context.Context, mrtPath, dataPath, sourceName, licenseKey, format string, variables map[string]any) ([]byte, error) {
	if err := validateLocalFile(mrtPath, "mrt_path"); err != nil {
		return nil, err
	}
	if err := validateLocalFile(dataPath, "data_path"); err != nil {
		return nil, err
	}
	if sourceName == "" {
		sourceName = "Data"
	}
	return Export(ctx, ExportRequest{
		LicenseKey: licenseKey, MRTPath: mrtPath, DataPath: dataPath,
		DataSourceName: sourceName, Format: format, Variables: variables,
	})
}

// ExportPDFPaths renders a local MRT and JSON file as PDF without copying them.
func ExportPDFPaths(ctx context.Context, mrtPath, dataPath, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return ExportPaths(ctx, mrtPath, dataPath, sourceName, licenseKey, "pdf", optionalVariables(variables))
}

// ExportExcelPaths renders a local MRT and JSON file as Excel without copying them.
func ExportExcelPaths(ctx context.Context, mrtPath, dataPath, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return ExportPaths(ctx, mrtPath, dataPath, sourceName, licenseKey, "excel", optionalVariables(variables))
}

// ExportPDF renders local MRT and JSON files as PDF without reading or copying them in Go.
func ExportPDF(ctx context.Context, mrtPath, dataPath, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return ExportPaths(ctx, mrtPath, dataPath, sourceName, licenseKey, "pdf", optionalVariables(variables))
}

// ExportExcel renders local MRT and JSON files as Excel without reading or copying them in Go.
func ExportExcel(ctx context.Context, mrtPath, dataPath, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return ExportPaths(ctx, mrtPath, dataPath, sourceName, licenseKey, "excel", optionalVariables(variables))
}

// ExportPDFBytes renders in-memory MRT and JSON content as PDF.
func ExportPDFBytes(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "pdf", sourceName, licenseKey, optionalVariables(variables))
}

// ExportExcelBytes renders in-memory MRT and JSON content as Excel.
func ExportExcelBytes(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "excel", sourceName, licenseKey, optionalVariables(variables))
}

// ExportCSVBytes renders in-memory MRT and JSON content as CSV.
func ExportCSVBytes(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "csv", sourceName, licenseKey, optionalVariables(variables))
}

// ExportHTMLBytes renders in-memory MRT and JSON content as HTML.
func ExportHTMLBytes(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "html", sourceName, licenseKey, optionalVariables(variables))
}

func exportFormat(ctx context.Context, mrt, data []byte, format, sourceName, licenseKey string, variables map[string]any) ([]byte, error) {
	if sourceName == "" {
		sourceName = "Data"
	}
	if len(data) == 0 {
		data = []byte("null")
	}
	if !json.Valid(data) {
		return nil, errors.New("data must be valid JSON")
	}
	return Export(ctx, ExportRequest{LicenseKey: licenseKey, MRT: string(mrt), Data: string(data), DataSourceName: sourceName, Format: format, Variables: variables})
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
