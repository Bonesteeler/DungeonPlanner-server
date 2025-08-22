#nullable disable
namespace DungeonPlanner.Models
{
    public class Scene
    {
        public int ID { get; set; }
        public string Name { get; set; }
        public string Author { get; set; }
        public List<string> Tile { get; set; }
    }
}