using LogFileService.BackgroundServices;
using LogFileService.Controllers;

var builder = WebApplication.CreateBuilder(args);

// Setup Logs Folder
Directory.CreateDirectory(LogFileController.LogFolder);

// Add services to the container.

builder.Services.AddControllers();
builder.Services.AddHostedService<PingUtilityRunner>(s => new PingUtilityRunner("facebook.com"));

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseAuthorization();

app.MapControllers();

app.Run();
