#nullable disable
namespace DungeonPlanner.Models
{
  public class JsonScene
  {
    public string Name { get; set; } = "default";
    public string Author { get; set; }
    public List<JsonTile> Tiles { get; set; }
  }
}