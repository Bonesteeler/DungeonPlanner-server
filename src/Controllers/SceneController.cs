using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;

namespace DungeonPlanner.Controllers
{
  [ApiController]
  [Route("scenes")]
  public class SceneController : ControllerBase
  {
    private readonly IServiceProvider _serviceProvider;

    public SceneController(IServiceProvider serviceProvider)
    {
      _serviceProvider = serviceProvider;
    }

    [HttpGet]
    public IActionResult GetScenes()
    {
      return Ok("Hello World!");
    }

    [HttpPut("add")]
    public IActionResult UpdateScene(Scene scene)
    {
      Console.WriteLine(scene.Name);
      Console.WriteLine(scene.Author);
      var context = _serviceProvider.GetService<SceneContext>();
      if (context != null)
      {
        var addedScene = context.Scenes.Add(scene);
        scene.Tiles.ForEach(tile =>
        {
          context.Tiles.Add(tile);
        });
        context.SaveChanges();
        Console.WriteLine(context.Scenes.Count());
        return Ok("Added");
      }
      else
      {
        return StatusCode(500, "Failed to add scene");
      }
    }
  }
}