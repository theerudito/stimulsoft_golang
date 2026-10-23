using System.Text.Json;
using System.Text.Json.Serialization;
using System.Linq;
using System.Text;
using Stimulsoft.Base;
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
	[JsonPropertyName("mrt_path")]
	public string? MRTPath { get; set; }
	[JsonPropertyName("data_path")]
	public string? DataPath { get; set; }
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
            Console.InputEncoding = new UTF8Encoding(false);
            var request = JsonSerializer.Deserialize<Request>(Console.In.ReadToEnd(), new JsonSerializerOptions { PropertyNameCaseInsensitive = true })
                ?? throw new InvalidDataException("request is empty");
            if (!string.IsNullOrWhiteSpace(request.LicenseKey)) StiLicense.Key = request.LicenseKey;
            Stimulsoft.Drawing.Graphics.GraphicsEngine = OperatingSystem.IsWindows()
                ? Stimulsoft.Drawing.GraphicsEngine.Gdi
                : Stimulsoft.Drawing.GraphicsEngine.ImageSharp;
				var hasMRTPath = !string.IsNullOrWhiteSpace(request.MRTPath);
				var hasDataPath = !string.IsNullOrWhiteSpace(request.DataPath);
				var hasInlineData = !string.IsNullOrWhiteSpace(request.Data);
				if (hasDataPath && !hasMRTPath) throw new InvalidDataException("data_path requires mrt_path");
				if (!hasMRTPath && string.IsNullOrWhiteSpace(request.MRT)) throw new InvalidDataException("mrt or mrt_path is required");
				if (hasMRTPath && !hasDataPath && !hasInlineData) throw new InvalidDataException("data or data_path is required with mrt_path");
			var dataSourceName = string.IsNullOrWhiteSpace(request.DataSourceName) ? "Data" : request.DataSourceName;

			string? workDirectory = null;
			try
			{
				string templatePath;
				string jsonPath;
					if (hasMRTPath)
					{
						templatePath = ValidateLocalFile(request.MRTPath!, "mrt_path");
						if (hasDataPath)
						{
							jsonPath = ValidateLocalFile(request.DataPath!, "data_path");
						}
						else
						{
							workDirectory = Path.Combine(Path.GetTempPath(), "stimulsoft-runner", Guid.NewGuid().ToString("N"));
							Directory.CreateDirectory(workDirectory);
							jsonPath = Path.Combine(workDirectory, "data.json");
							File.WriteAllText(jsonPath, request.Data.TrimStart('\uFEFF'), new UTF8Encoding(true));
						}
					}
				else
				{
					var template = request.MRT.TrimStart('\uFEFF');
					var data = request.Data.TrimStart('\uFEFF');
					workDirectory = Path.Combine(Path.GetTempPath(), "stimulsoft-runner", Guid.NewGuid().ToString("N"));
					Directory.CreateDirectory(workDirectory);
					templatePath = Path.Combine(workDirectory, "report.mrt");
					jsonPath = Path.Combine(workDirectory, "data.json");
					File.WriteAllText(templatePath, template, new UTF8Encoding(true));
					File.WriteAllText(jsonPath, data, new UTF8Encoding(true));
				}

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
                if (workDirectory is not null)
                {
                    try { Directory.Delete(workDirectory, true); } catch { }
                }
            }
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine(ex.ToString());
            return 1;
        }
    }

    private static string ValidateLocalFile(string path, string field)
    {
        if (Uri.TryCreate(path, UriKind.Absolute, out var uri) && !string.IsNullOrEmpty(uri.Scheme) && uri.Scheme != Uri.UriSchemeFile)
            throw new InvalidDataException($"{field} must be a local file path");
        var fullPath = Path.GetFullPath(path);
        if (!File.Exists(fullPath)) throw new FileNotFoundException($"{field} does not exist", fullPath);
        if ((File.GetAttributes(fullPath) & FileAttributes.Directory) != 0) throw new InvalidDataException($"{field} must be a file");
        return fullPath;
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
            case "excel": new StiExcel2007ExportService().ExportExcel(report, stream); break;
            case "csv": new StiCsvExportService().ExportCsv(report, stream); break;
            case "html": new StiHtmlExportService().ExportHtml(report, stream); break;
            default: throw new ArgumentException($"unsupported format '{format}' (expected pdf, excel, csv, or html)");
        }
    }
}
