using Microsoft.AspNetCore.Mvc;

namespace LogFileService.Controllers
{
    [ApiController]
    [Route("[controller]")]
    public class LogFileController : ControllerBase
    {
        private readonly ILogger<LogFileController> _logger;

        public static readonly string LogFolder = "/app/logs/";

        public LogFileController(ILogger<LogFileController> logger)
        {
            _logger = logger;
        }

        [HttpGet]
        public IEnumerable<string> Get()
        {
            return Directory.GetFiles(LogFolder).Select(p => Path.GetFileName(p));
        }

        [HttpGet("{filename}")]
        public PhysicalFileResult GetFile(string filename)
        {
            return PhysicalFile(LogFolder + filename.Replace("..", ""), "text/plain");
        }

        [HttpDelete("{filename}")]
        public void DeleteFile(string filename)
        {
            System.IO.File.Delete(LogFolder + filename.Replace("..", ""));
            _logger.LogInformation("Deleted {path}", filename);
        }
    }
}