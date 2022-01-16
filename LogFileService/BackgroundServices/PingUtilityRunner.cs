using LogFileService.Controllers;
using PingUtilityLibrary;

namespace LogFileService.BackgroundServices
{
    public class PingUtilityRunner : BackgroundService
    {
        public string Target { get; }
        public PingUtilityRunner(string target)
        {
            Target = target;
        }

        protected override async Task ExecuteAsync(CancellationToken stoppingToken)
        {
            await Task.Run(() =>
            {
                using (File.Create(PingUtility.logFileName)) { };
                PingUtility.pingAndLog(Target,
                                       PingUtility.noRouter,
                                       PingUtility.defaultLocation,
                                       PingUtility.optionaLocationGenerator(LogFileController.LogFolder + "log"));
            });
        }
    }

}
