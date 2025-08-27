using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;

namespace DungeonPlanner.Controllers
{
  [ApiController]
  [Route("scenes")]
  public class SceneController : ControllerBase
  {
    private const int LIST_LIMIT = 20;
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

    [HttpGet("list")]
    public IActionResult ListScenes()
    {
      var context = _serviceProvider.GetService<SceneContext>();
      if (context == null)
      {
        return StatusCode(500, "Database context not available");
      }
      var scenesToGet = context.Scenes.OrderBy(s => s.ID).Take(LIST_LIMIT).ToList();
      return Ok(scenesToGet);
    }

    [HttpGet("{id}")]
    public IActionResult GetSceneById(int id)
    {
      Console.WriteLine($"Fetching scene with ID: {id}");
      var context = _serviceProvider.GetService<SceneContext>();
      if (context == null)
      {
        return StatusCode(500, "Database context not available");
      }
      var scene = context.Scenes.FirstOrDefault(s => s.ID == id);
      if (scene == null)
      {
        return NotFound();
      }
      scene.Tiles = [.. context.Tiles.Where(t => t.SceneID == scene.ID)];
      return Ok(scene);
    }

    [HttpPut("add")]
    public IActionResult UpdateScene(Scene scene)
    {
      Console.WriteLine(scene.Name);
      Console.WriteLine(scene.Author);
      var context = _serviceProvider.GetService<SceneContext>();
      if (context != null)
      {
        var uniqueIds = new HashSet<string>();
        scene.Tiles.ForEach(tile =>
        {
          context.Tiles.Add(tile);
          uniqueIds.Add(tile.TileID);
        });
        scene.UniqueTileIDs = [.. uniqueIds];
        context.Scenes.Add(scene);
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