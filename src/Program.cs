using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Text.Json;

var builder = WebApplication.CreateBuilder(args);
builder.Services.AddRazorPages();

string? sceneConnectionTemplate = builder.Configuration.GetConnectionString("SceneContext");
var passwordSecretName = "db-password";
var passwordSecretPath = $"/etc/secrets/{passwordSecretName}";
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
if (File.Exists(passwordSecretPath))
{
  Console.WriteLine($"Found secret {passwordSecretName} at {passwordSecretPath}");
  var passwordSecret = File.ReadAllText(passwordSecretPath);
  var sceneConnectionString = string.Format(sceneConnectionTemplate!, passwordSecret);
  builder.Services.AddDbContext<SceneContext>(options =>
      options.UseNpgsql(sceneConnectionString));
}
else
{
  Console.WriteLine($"Secret {passwordSecretName} not found at {passwordSecretPath}");
}

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