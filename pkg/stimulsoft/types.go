package stimulsoft

// ExportRequest describes one report export sent to the runner.
type ExportRequest struct {
	LicenseKey     string         `json:"license_key,omitempty"`
	MRTPath        string         `json:"mrt_path,omitempty"`
	Data           string         `json:"data"`
	DataSourceName string         `json:"data_source_name,omitempty"`
	Format         string         `json:"format,omitempty"`
	Variables      map[string]any `json:"variables,omitempty"`
}
