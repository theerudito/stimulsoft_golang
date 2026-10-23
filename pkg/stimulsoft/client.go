package stimulsoft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/theerudito/stimulsoft_golang/internal/config"
	"github.com/theerudito/stimulsoft_golang/internal/downloader"
	"github.com/theerudito/stimulsoft_golang/internal/runner"
)

func Export(ctx context.Context, request ExportRequest) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	request.MRT = strings.TrimPrefix(request.MRT, "\uFEFF")
	request.Data = strings.TrimPrefix(request.Data, "\uFEFF")
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

// ExportPDF renders an MRT report as PDF. Variables are optional.
func ExportPDF(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "pdf", sourceName, licenseKey, optionalVariables(variables))
}

// ExportExcel renders an MRT report as Excel. Variables are optional.
func ExportExcel(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "excel", sourceName, licenseKey, optionalVariables(variables))
}

// ExportCSV renders an MRT report as CSV. Variables are optional.
func ExportCSV(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return exportFormat(ctx, mrt, data, "csv", sourceName, licenseKey, optionalVariables(variables))
}

// ExportHTML renders an MRT report as HTML. Variables are optional.
func ExportHTML(ctx context.Context, mrt, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
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
