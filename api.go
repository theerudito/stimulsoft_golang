package stimulsoftgo

import (
	"context"

	legacy "github.com/theerudito/stimulsoft_golang/pkg/stimulsoft"
)

// ExportRequest describes one report export sent to the runner.
type ExportRequest = legacy.ExportRequest

// Export sends an export request to the Stimulsoft runner.
func Export(ctx context.Context, request ExportRequest) ([]byte, error) {
	return legacy.Export(ctx, request)
}

// ExportPDF renders a local MRT path with in-memory JSON data as PDF.
func ExportPDF(ctx context.Context, mrtPath string, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return legacy.ExportPDF(ctx, mrtPath, data, sourceName, licenseKey, variables...)
}

// ExportExcel renders a local MRT path with in-memory JSON data as Excel.
func ExportExcel(ctx context.Context, mrtPath string, data []byte, sourceName, licenseKey string, variables ...map[string]any) ([]byte, error) {
	return legacy.ExportExcel(ctx, mrtPath, data, sourceName, licenseKey, variables...)
}
