# Stimulsoft Go Client

Render Stimulsoft `.mrt` reports from Go through a self-contained .NET 8 runner.
The MRT stays on disk, while report data can come directly from a query, struct,
or pointer as JSON bytes.

## Quick start

### 1. Add the module

```bash
 go get github.com/theerudito/stimulsoft_golang@v0.1.25
```

### 2. Render PDF or Excel

```go
package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	stimulsoft "github.com/theerudito/stimulsoft_golang"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The MRT can be relative or absolute.
	mrtPath := "./document/RPT_Factura_Electronica_Rollo.mrt"
	licenseKey := os.Getenv("STIMULSOFT_LICENSE_KEY")

	// pedido may come from SQL, an API, a struct, or a pointer.
	pedido := map[string]any{"items": []any{map[string]any{"name": "Example", "quantity": 2}}}
	jsonBytes, err := json.MarshalIndent(pedido, "", "  ")
	if err != nil {
		panic(err)
	}

	pdf, err := stimulsoft.ExportPDF(ctx, mrtPath, jsonBytes, "informe", licenseKey)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("report.pdf", pdf, 0644); err != nil {
		panic(err)
	}

	excel, err := stimulsoft.ExportExcel(ctx, mrtPath, jsonBytes, "informe", licenseKey)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("report.xlsx", excel, 0644); err != nil {
		panic(err)
	}
}
```

The `informe` value must match the JSON datasource expected by the MRT. If the
report uses another datasource name, pass that name instead.

## API

The client intentionally exposes only two export helpers:

| Function | Input | Output |
|---|---|---|
| `ExportPDF` | MRT path + JSON `[]byte` | PDF bytes |
| `ExportExcel` | MRT path + JSON `[]byte` | XLSX bytes |

Both functions:

- Accept relative or absolute MRT paths.
- Validate the MRT before starting the runner.
- Accept JSON generated with `json.Marshal` or `json.MarshalIndent`.
- Leave the original MRT unchanged.
- Accept optional report variables as the final argument.

```go
variables := map[string]any{
	"CompanyName": "Example Company",
}

pdf, err := stimulsoft.ExportPDF(
	ctx,
	"./document/report.mrt",
	jsonBytes,
	"informe",
	licenseKey,
	variables,
)
```

## Runner

The runner is downloaded automatically for the current platform and cached
locally. To use a locally published runner instead:

```powershell
dotnet publish runner\Stimulsoft.Runner.csproj `
  -c Release `
  -r win-x64 `
  --self-contained true `
  -o "$env:TEMP\stimulsoft-runner\win-x64"

$env:STIMULSOFT_RUNNER_PATH = `
  "$env:TEMP\stimulsoft-runner\win-x64\Stimulsoft.Runner.exe"
```

The runner targets .NET 8 and uses Stimulsoft Reports Engine NetCore `2026.1.7`.
Windows uses GDI; Linux and macOS use ImageSharp.

## Configuration

| Environment variable | Purpose |
|---|---|
| `STIMULSOFT_LICENSE_KEY` | Stimulsoft license key |
| `STIMULSOFT_RUNNER_PATH` | Use a local runner instead of downloading one |
| `STIMULSOFT_RUNNER_VERSION` | Select a released runner version |
| `STIMULSOFT_RUNNER_OWNER` | Override the GitHub owner |
| `STIMULSOFT_RUNNER_REPO` | Override the GitHub repository |
| `STIMULSOFT_RUNNER_BASE_URL` | Override the runner download base URL |

An empty license key is allowed for evaluation or environments where the
Stimulsoft license is configured elsewhere.

## Development

```bash
go test ./...
go vet ./...
dotnet build runner/Stimulsoft.Runner.csproj
```

## Scope

Currently supported:

- PDF
- Excel (`.xlsx`)

CSV, HTML, and other formats are intentionally outside the current public API.
