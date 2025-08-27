using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using System.Text.Json;

var builder = WebApplication.CreateBuilder(args);
builder.Services.AddRazorPages();

string? sceneConnectionTemplate = builder.Configuration.GetConnectionString("SceneContext");
var passwordSecretName = "db-password";
var passwordSecretPath = $"/run/secrets/{passwordSecretName}";
if (File.Exists(passwordSecretPath))
{
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