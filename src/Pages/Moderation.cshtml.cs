using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace DungeonPlanner.Pages
{
  public class ModerationModel(SceneContext context, ILogger<ModerationModel> logger) : PageModel
  {
    public List<Scene> PendingScenes { get; private set; } = [];
    public int QueueSize { get; private set; }

    private readonly SceneContext _context = context;
    private readonly ILogger<ModerationModel> _logger = logger;

    public void OnGet()
    {
      QueueSize = _context.Scenes.Count(s => s.ModerationStatus == SceneModerationStatus.Pending);
      PendingScenes = [.. _context.Scenes.Where(s => s.ModerationStatus == SceneModerationStatus.Pending)];
    }

    public string UniqueTileIdsOfScene(Scene scene)
    {
      return string.Join(",", scene.UniqueTileIDs);
    }

    public async Task<IActionResult> OnPostApproveAsync(Guid sceneId)
    {
      var scene = _context.Scenes.Where(s => s.ID == sceneId).FirstOrDefault();
      if (scene == null)
      {
        _logger.LogWarning("Could not find scene with ID {sceneId} to approve", sceneId);
        return NotFound();
      }
      _logger.LogInformation("Approving scene with ID {sceneId}", sceneId);
      scene.ModerationStatus = SceneModerationStatus.Approved;
      await _context.SaveChangesAsync();
      return RedirectToPage();
    }

    public async Task<IActionResult> OnPostRejectAsync(Guid sceneId)
    {
      var scene = _context.Scenes.Where(s => s.ID == sceneId).FirstOrDefault();
      if (scene == null)
      {
        _logger.LogWarning("Could not find scene with ID {sceneId} to reject", sceneId);
        return NotFound();
      }
      _logger.LogInformation("Rejecting scene with ID {sceneId}", sceneId);
      scene.ModerationStatus = SceneModerationStatus.Rejected;
      await _context.SaveChangesAsync();
      return RedirectToPage();
    }
  }
}