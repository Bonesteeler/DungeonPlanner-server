#nullable disable
namespace DungeonPlanner.Models
{
  public enum SceneModerationStatus
  {
    Pending,
    Approved,
    Rejected
  }

  public class Scene
  {
    public Guid ID { get; set; }
    public string Name { get; set; } = "default";
    public string Author { get; set; }
    public List<Tile> Tiles { get; set; }
    public List<string> UniqueTileIDs { get; set; }
    public SceneModerationStatus ModerationStatus { get; set; } = SceneModerationStatus.Pending;
  }
}