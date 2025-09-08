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

    private class ListResponse
    {
      public int PageSize { get; set; } = LIST_LIMIT;
      public int SceneCount { get; set; }
      public List<Scene> Scenes { get; set; } = [];
    }

    [HttpGet("list/{start}")]
    public IActionResult ListScenes(int start)
    {
      var context = _serviceProvider.GetService<SceneContext>();
      if (context == null)
      {
        return StatusCode(500, "Database context not available");
      }
      var sceneCount = context.Scenes.Where(s => s.ModerationStatus == SceneModerationStatus.Approved).Count();
      var startingScene = 0;
      if (start > 0 && start < sceneCount)
      {
        startingScene = start;
      }
      var scenesToGet =
        context.Scenes.OrderBy(s => s.ID)
          .Where(s => s.ModerationStatus == SceneModerationStatus.Approved)
          .Skip(startingScene)
          .Take(LIST_LIMIT)
          .ToList();
      var response = new ListResponse
      {
        SceneCount = sceneCount,
        Scenes = scenesToGet
      };
      return Ok(response);
    }

    [HttpGet("{id}")]
    public IActionResult GetSceneById(Guid id)
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
    public IActionResult UpdateScene(JsonScene scene)
    {
      var newScene = new Scene
      {
        Name = scene.Name,
        Author = scene.Author,
        ModerationStatus = SceneModerationStatus.Pending,
        Tiles = [.. scene.Tiles.Select(t => new Tile
        {
          TileID = t.TileID,
          Rotation = (int)t.Rotation,
          XPos = (int)t.XPos,
          YPos = (int)t.YPos
        })]
      };
      Console.WriteLine(newScene.Name);
      Console.WriteLine(newScene.Author);
      var context = _serviceProvider.GetService<SceneContext>();
      if (context != null)
      {
        var uniqueIds = new HashSet<string>();
        newScene.Tiles.ForEach(tile =>
        {
          context.Tiles.Add(tile);
          uniqueIds.Add(tile.TileID);
        });
        newScene.UniqueTileIDs = [.. uniqueIds];
        context.Scenes.Add(newScene);
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