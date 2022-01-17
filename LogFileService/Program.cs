using LogFileService.BackgroundServices;
using LogFileService.Controllers;
using PingUtilityLibrary;

var builder = WebApplication.CreateBuilder(args);

// Handle parameters
var location = PingUtility.defaultLocation;
var target = "facebook.com";
if (args.Length > 0)
{
    location = args[0];
}

if (args.Length > 1)
{
    target = args[1];
}

// Setup Logs Folder
Directory.CreateDirectory(LogFileController.LogFolder);

// Add services to the container.

builder.Services.AddControllers();
builder.Services.AddHostedService<PingUtilityRunner>(s => new PingUtilityRunner(target, location));

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseAuthorization();

app.MapControllers();

app.Run();
