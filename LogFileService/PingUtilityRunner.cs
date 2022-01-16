using LogFileService.Controllers;
using PingUtilityLibrary;

public static class PingUtilityRunner
{
    public static void Run(string target)
    {
        Task.Run(() => PingUtility.pingAndLog(target, PingUtility.noRouter, PingUtility.defaultLocation, PingUtility.optionaLocationGenerator(LogFileController.LogFolder + "log")));
    }

}
