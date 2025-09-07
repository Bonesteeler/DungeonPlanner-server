using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Text.Json;
using System.Text.RegularExpressions;

var builder = WebApplication.CreateBuilder(args);
builder.Services.AddRazorPages();

// string? sceneConnectionTemplate = builder.Configuration.GetConnectionString("SceneContext");
var passwordSecretName = "db-password";
var passwordSecretPath = $"/etc/secrets/{passwordSecretName}";
Console.WriteLine("Sanity check");
if (Directory.Exists("/etc/secrets"))
{
  Directory.GetFiles("/etc/secrets").ToList().ForEach(file =>
      Console.WriteLine($"Found secret file at {file}"));
  Console.WriteLine($"Found secrets directory at /etc/secrets");
}
else
{
  Console.WriteLine($"Secrets directory not found at /etc/secrets");
}
if (Directory.Exists("/run/secrets"))
{
  Console.WriteLine($"Found secrets directory at /run/secrets");
}
else
{
  Console.WriteLine($"Secrets directory not found at /run/secrets");
}
builder.Services.AddDbContext<SceneContext>(options =>
{
    var match = Regex.Match(Environment.GetEnvironmentVariable("DATABASE_URL") ?? "", @"postgres://(.*):(.*)@(.*):(.*)/(.*)");
    options.UseNpgsql($"Server={match.Groups[3]};Port={match.Groups[4]};User Id={match.Groups[1]};Password={match.Groups[2]};Database={match.Groups[5]};sslmode=Prefer;Trust Server Certificate=true");
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