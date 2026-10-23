using System.Text.Json;
using System.Text.Json.Serialization;
using System.Text;
using Stimulsoft.Base;
using Stimulsoft.Report;
using Stimulsoft.Report.Dictionary;
using Stimulsoft.Report.Export;

internal sealed class Request
{
    [JsonPropertyName("license_key")]
    public string? LicenseKey { get; set; }
    [JsonPropertyName("mrt_path")]
    public string? MRTPath { get; set; }
    [JsonPropertyName("data")]
    public string Data { get; set; } = "null";
    [JsonPropertyName("data_source_name")]
    public string DataSourceName { get; set; } = "Data";
    [JsonPropertyName("format")]
    public string? Format { get; set; }
    [JsonPropertyName("variables")]
    public Dictionary<string, JsonElement>? Variables { get; set; }
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
            // Report exporters may write trial notices to Console.Out. The process stdout is
            // the binary export channel, so text output must never share it.
            Console.SetOut(TextWriter.Null);
            if (!string.IsNullOrWhiteSpace(request.LicenseKey)) StiLicense.Key = request.LicenseKey;
            Stimulsoft.Drawing.Graphics.GraphicsEngine = OperatingSystem.IsWindows()
                ? Stimulsoft.Drawing.GraphicsEngine.Gdi
                : Stimulsoft.Drawing.GraphicsEngine.ImageSharp;
            if (string.IsNullOrWhiteSpace(request.MRTPath)) throw new InvalidDataException("mrt_path is required");
            var dataSourceName = string.IsNullOrWhiteSpace(request.DataSourceName) ? "Data" : request.DataSourceName;

            string? workDirectory = null;
            try
            {
                var templatePath = ValidateLocalFile(request.MRTPath!, "mrt_path");
                workDirectory = Path.Combine(Path.GetTempPath(), "stimulsoft-runner", Guid.NewGuid().ToString("N"));
                Directory.CreateDirectory(workDirectory);
                var jsonPath = Path.Combine(workDirectory, "data.json");
                File.WriteAllText(jsonPath, request.Data.TrimStart('\uFEFF'), new UTF8Encoding(true));

                using var report = new StiReport();
                report.Load(templatePath);
                report.Dictionary.Databases.Clear();
                report.Dictionary.Databases.Add(new StiJsonDatabase(dataSourceName, jsonPath));
                report.Dictionary.Synchronize();
                InjectVariables(report, request.Variables);
                report.Render(false);
                using var output = new MemoryStream();
                var format = (request.Format ?? "pdf").ToLowerInvariant();
                Export(report, output, format);
                WriteOutput(output, format);
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

    private static void Export(StiReport report, Stream stream, string format)
    {
        switch (format)
        {
            case "pdf": report.ExportDocument(StiExportFormat.Pdf, stream); break;
            case "excel": new StiExcel2007ExportService().ExportExcel(report, stream); break;
            default: throw new ArgumentException($"unsupported format '{format}' (expected pdf or excel)");
        }
    }

    private static void WriteOutput(MemoryStream output, string format)
    {
        var bytes = output.ToArray();
        if (format == "excel")
        {
            var zipStart = FindZipStart(bytes);
            if (zipStart < 0)
                throw new InvalidDataException("Excel export did not produce a ZIP archive");
            Console.OpenStandardOutput().Write(bytes, zipStart, bytes.Length - zipStart);
            return;
        }

        Console.OpenStandardOutput().Write(bytes);
    }

    private static int FindZipStart(byte[] bytes)
    {
        for (var i = 0; i <= bytes.Length - 4; i++)
        {
            if (bytes[i] == 0x50 && bytes[i + 1] == 0x4B && bytes[i + 2] == 0x03 && bytes[i + 3] == 0x04)
                return i;
        }

        return -1;
    }
}
