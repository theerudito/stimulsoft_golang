using System.Text.Json;
using System.Text.Json.Serialization;
using System.Linq;
using System.Text;
using Stimulsoft.Base;
using Stimulsoft.Drawing;
using Stimulsoft.Report;
using Stimulsoft.Report.Dictionary;
using Stimulsoft.Report.Export;

internal sealed class Request
{
    [JsonPropertyName("license_key")]
    public string? LicenseKey { get; set; }
    [JsonPropertyName("mrt")]
    public string MRT { get; set; } = "";
    [JsonPropertyName("data")]
    public string Data { get; set; } = "null";
    [JsonPropertyName("data_source_name")]
    public string DataSourceName { get; set; } = "Data";
    [JsonPropertyName("format")]
    public string? Format { get; set; }
    [JsonPropertyName("variables")]
    public Dictionary<string, JsonElement>? Variables { get; set; }
    [JsonPropertyName("options")]
    public Dictionary<string, JsonElement>? Options { get; set; }
}

internal static class Program
{
    public static int Main()
    {
        try
        {
            var request = JsonSerializer.Deserialize<Request>(Console.In.ReadToEnd(), new JsonSerializerOptions { PropertyNameCaseInsensitive = true })
                ?? throw new InvalidDataException("request is empty");
            if (!string.IsNullOrWhiteSpace(request.LicenseKey)) StiLicense.Key = request.LicenseKey;
            Graphics.GraphicsEngine = GraphicsEngine.Gdi;
            if (string.IsNullOrWhiteSpace(request.MRT)) throw new InvalidDataException("mrt is required");
            var dataSourceName = string.IsNullOrWhiteSpace(request.DataSourceName) ? "Data" : request.DataSourceName;

            var template = request.MRT.TrimStart('\uFEFF', '\u200B', '\u0000', '\r', '\n', ' ', '\t');
            if (template.StartsWith("ï»¿", StringComparison.Ordinal)) template = template[3..];
            var workDirectory = Path.Combine(Path.GetTempPath(), "stimulsoft-runner", Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(workDirectory);
            try
            {
                var templatePath = Path.Combine(workDirectory, "report.mrt");
                var jsonPath = Path.Combine(workDirectory, "data.json");
                File.WriteAllText(templatePath, template, new UTF8Encoding(false));
                File.WriteAllText(jsonPath, request.Data, new UTF8Encoding(false));

                using var report = new StiReport();
                report.Load(templatePath);
                report.Dictionary.Databases.Clear();
                report.Dictionary.Databases.Add(new StiJsonDatabase(dataSourceName, jsonPath));
                report.Dictionary.Synchronize();
                InjectVariables(report, request.Variables);
                report.Render(false);
                using var output = new MemoryStream();
                Export(report, output, (request.Format ?? "pdf").ToLowerInvariant(), request.Options);
                Console.OpenStandardOutput().Write(output.GetBuffer(), 0, checked((int)output.Length));
                return 0;
            }
            finally
            {
                try { Directory.Delete(workDirectory, true); } catch { }
            }
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine(ex.ToString());
            return 1;
        }
    }

    private static void ConfigureJsonDatabase(StiReport report, string dataSourceName, string data)
    {
        var database = report.Dictionary.Databases
            .Cast<StiDatabase>()
            .FirstOrDefault(item => string.Equals(item.Name, dataSourceName, StringComparison.OrdinalIgnoreCase)
                || string.Equals(item.Alias, dataSourceName, StringComparison.OrdinalIgnoreCase));

        if (database is StiJsonDatabase jsonDatabase)
        {
            // Keep the MRT dictionary, relations, aliases, and datasource definitions intact.
            // Only replace the JSON payload supplied by the caller.
            jsonDatabase.PathData = data;
            return;
        }

        report.Dictionary.Databases.Add(new StiJsonDatabase(dataSourceName, data));
    }

    private static void InjectVariables(StiReport report, Dictionary<string, JsonElement>? values)
    {
        if (values is null) return;
        foreach (var pair in values)
        {
            // Variable definitions belong to the MRT dictionary. Reflection is isolated here because
            // ValueObject differs between Stimulsoft package generations; unknown names are ignored.
            var variable = report.Dictionary.Variables[pair.Key];
            if (variable is null) continue;
            var value = PrimitiveValue(pair.Value);
            var property = variable.GetType().GetProperty("ValueObject") ?? variable.GetType().GetProperty("Value");
            if (property?.CanWrite == true) property.SetValue(variable, value);
        }
    }

    private static object? PrimitiveValue(JsonElement value) => value.ValueKind switch
    {
        JsonValueKind.String => value.GetString(),
        JsonValueKind.True => true,
        JsonValueKind.False => false,
        JsonValueKind.Number when value.TryGetInt64(out var i) => i,
        JsonValueKind.Number when value.TryGetDecimal(out var d) => d,
        JsonValueKind.Null => null,
        _ => value.GetRawText()
    };

    private static void Export(StiReport report, Stream stream, string format, Dictionary<string, JsonElement>? options)
    {
        switch (format)
        {
            case "pdf": report.ExportDocument(StiExportFormat.Pdf, stream); break;
            case "excel": report.ExportDocument(StiExportFormat.Excel, stream); break;
            case "csv": new StiCsvExportService().ExportCsv(report, stream); break;
            case "html": new StiHtmlExportService().ExportHtml(report, stream); break;
            default: throw new ArgumentException($"unsupported format '{format}' (expected pdf, excel, csv, or html)");
        }
    }
}
