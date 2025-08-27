#nullable disable
namespace DungeonPlanner.Models
{
    public class Scene
    {
        public int ID { get; set; }
        public string Name { get; set; } = "default";
        public string Author { get; set; }
        public List<Tile> Tiles { get; set; }
        public List<string> UniqueTileIDs { get; set; }
    }
}