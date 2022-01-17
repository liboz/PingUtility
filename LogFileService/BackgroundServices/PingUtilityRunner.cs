using LogFileService.Controllers;
using PingUtilityLibrary;

namespace LogFileService.BackgroundServices
{
    public class PingUtilityRunner : BackgroundService
    {
        public string Target { get; }
        public string Location { get; }

        public PingUtilityRunner(string target, string location)
        {
            Target = target;
            Location = location;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            await Task.Run(() =>
            {
                using (File.Create(PingUtility.logFileName)) { };
                PingUtility.pingAndLog(Target,
                                       PingUtility.noRouter,
                                       Location,
                                       PingUtility.optionaLocationGenerator(LogFileController.LogFolder + "log"));
            });
        }
    }

}
