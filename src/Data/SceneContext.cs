using Microsoft.EntityFrameworkCore;
namespace DungeonPlanner.Data
{
    public class SceneContext : DbContext
    {
        public SceneContext(DbContextOptions<SceneContext> options) : base(options) { }
        public DbSet<Models.Scene> Scenes { get; set; }
    }
}