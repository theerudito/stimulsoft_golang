package stimulsoft

// ExportRequest describes one report export sent to the runner.
type ExportRequest struct {
	LicenseKey     string         `json:"license_key,omitempty"`
	MRT            string         `json:"mrt"`
	Data           string         `json:"data"`
	MRTPath        string         `json:"mrt_path,omitempty"`
	DataPath       string         `json:"data_path,omitempty"`
	DataSourceName string         `json:"data_source_name,omitempty"`
	Format         string         `json:"format,omitempty"`
	Variables      map[string]any `json:"variables,omitempty"`
	Options        map[string]any `json:"options,omitempty"`
}
