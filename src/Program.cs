using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Text.Json;
using System.Text.RegularExpressions;

var builder = WebApplication.CreateBuilder(args);
var port = Environment.GetEnvironmentVariable("PORT");
if (!string.IsNullOrEmpty(port))
{
  Console.WriteLine($"Binding to port {port}");
  builder.WebHost.UseUrls($"http://*:{port}");
}
builder.Services.AddRazorPages();

builder.Services.AddDbContext<SceneContext>(options =>
{
  var databaseUrl = Environment.GetEnvironmentVariable("DATABASE_URL");
  if (string.IsNullOrEmpty(databaseUrl))
  {
    var passwordSecretPath = $"/run/secrets/db-password";
    string? sceneConnectionTemplate = builder.Configuration.GetConnectionString("SceneContext");
    if (File.Exists(passwordSecretPath) && sceneConnectionTemplate != null)
    {
      var passwordSecret = File.ReadAllText(passwordSecretPath);
      var sceneConnectionString = string.Format(sceneConnectionTemplate, passwordSecret);
      options.UseNpgsql(sceneConnectionString);
    }
    else
    {
      Console.WriteLine("Failed to find database configuration");
      Console.WriteLine($"Password secret exists: {File.Exists(passwordSecretPath)}");
      Console.WriteLine($"Scene connection string exists: {sceneConnectionTemplate != null}");
    }
  }
  else
  {
    var match = Regex.Match(Environment.GetEnvironmentVariable("DATABASE_URL") ?? "", @"postgres://(.*):(.*)@(.*):(.*)/(.*)");
    options.UseNpgsql($"Server={match.Groups[3]};Port={match.Groups[4]};User Id={match.Groups[1]};Password={match.Groups[2]};Database={match.Groups[5]};sslmode=Prefer;Trust Server Certificate=true");
  }
});

var app = builder.Build();
using (var scope = app.Services.CreateScope())
{
   var services = scope.ServiceProvider;
   try
   {
       var context = services.GetRequiredService<SceneContext>();
       var created = context.Database.EnsureCreated();
   }
   catch (Exception ex)
   {
       var logger = services.GetRequiredService<ILogger<Program>>();
       logger.LogError(ex, "An error occurred creating the DB.");
   }
}


// Configure the HTTP request pipeline.
if (!app.Environment.IsDevelopment())
{
    app.UseExceptionHandler("/Error");
}
app.UseStaticFiles();

app.UseRouting();

app.UseAuthorization();

app.MapRazorPages();

app.MapControllers();

app.Run();