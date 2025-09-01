using DungeonPlanner.Data;
using DungeonPlanner.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace DungeonPlanner.Pages
{
    public class IndexModel : PageModel
    {
        public int SceneCount { get; private set; }
        private readonly ILogger<IndexModel> _logger;
        private readonly SceneContext _context;

        public IndexModel(ILogger<IndexModel> logger, SceneContext context)
        {
            _logger = logger;
            _context = context;
        }

        public void OnGet()
        {
          this.SceneCount = _context.Scenes.Where(s => s.ModerationStatus == SceneModerationStatus.Pending).Count();
        }
    }
}