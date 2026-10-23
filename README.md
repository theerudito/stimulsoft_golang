 # Stimulsoft Go client

Source-only Go client for a self-contained .NET 8 Stimulsoft runner. The Go module contains no binaries or embedded assets. The runner is downloaded lazily or selected with `STIMULSOFT_RUNNER_PATH`.

## Export API

Every export receives MRT text, raw JSON text, a data source name, and a license key. The data source name defaults to `Data`; use another name when the MRT expects it. Pass an empty license key to run without a license.

Pass an empty source name to select the default explicitly:

```go
licenseKey := os.Getenv("STIMULSOFT_LICENSE_KEY")
pdf, err := stimulsoft.ExportPDF(ctx, mrt, jsonData, "Data", licenseKey)
```

```go
licenseKey := os.Getenv("STIMULSOFT_LICENSE_KEY")
pdf, err := stimulsoft.ExportPDF(ctx, mrt, jsonData, "Data", licenseKey)
excel, err := stimulsoft.ExportExcel(ctx, mrt, jsonData, "Data", licenseKey)
```

Variables are optional:

```go
pdf, err := stimulsoft.ExportPDF(ctx, mrt, jsonData, "Data", licenseKey, variables)
```

If JSON comes from SQL, do not create a temporary file:

```go
jsonFromSQL := []byte(`{"items":[{"name":"From SQL","quantity":2}]}`)
pdf, err := stimulsoft.ExportPDF(ctx, mrt, jsonFromSQL, "Data", licenseKey)
```

The compatibility method `ExportSimple` has the explicit source and format parameters:

```go
out, err := stimulsoft.ExportSimple(ctx, string(mrt), string(jsonData), "Data", licenseKey, "pdf", nil)
```

## Example

From `cmd/`:

```powershell
go run . ..\document\RPT_Comanda.mrt ..\document\RPT_Comanda.json
```

## Local runner

```powershell
dotnet publish runner\Stimulsoft.Runner.csproj -c Release -r win-x64 --self-contained true -o "$env:TEMP\stimulsoft-runner\win-x64"
$env:STIMULSOFT_RUNNER_PATH="$env:TEMP\stimulsoft-runner\win-x64\Stimulsoft.Runner.exe"
go run .
```

The runner supports PDF, Excel, CSV, and HTML through stdin/stdout and never writes report temporary files. The official Stimulsoft package and license terms apply. Trimming is disabled by default because Stimulsoft uses reflection; test an aggressive trimming profile before enabling it.
