#nullable disable
namespace DungeonPlanner.Models
{
  public class Tile
  {
    public Guid ID { get; set; }
    public string TileID { get; set; }
    public int Rotation { get; set; }
    public int XPos { get; set; }
    public int YPos { get; set; }
    public Guid SceneID { get; set; }
  }
}