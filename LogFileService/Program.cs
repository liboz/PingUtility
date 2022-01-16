var builder = WebApplication.CreateBuilder(args);

// Start PingUtility
foreach (var item in args)
{
    Console.WriteLine(item);
}
Console.WriteLine(args.Length);
PingUtilityRunner.Run("facebook.com");

// Add services to the container.

builder.Services.AddControllers();

var app = builder.Build();

// Configure the HTTP request pipeline.

app.UseAuthorization();

app.MapControllers();

app.Run();
