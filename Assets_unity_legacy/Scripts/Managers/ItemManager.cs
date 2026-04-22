using UnityEngine;
using System.Collections;
using static UnityEditor.Progress;
using System.Linq;
/// <summary>
/// singalton item manager
/// </summary>
public class ItemManager : MonoBehaviour
{
    public static ItemManager Instance;

    public GraphicCardItem[] GraphicCardItems;

    /// <summary>
    /// Awake is called when script instance is being loaded, which means
    /// it will init when game start if we put it on helper object
    /// </summary>
    private void Awake()
    {
        Instance = this;
        GraphicCardItems = Utils.GetAllInstance<GraphicCardItem>();
        GraphicCardItems = GraphicCardItems.OrderBy(item => item.Id).ToArray();
        foreach (GraphicCardItem item in GraphicCardItems)
        {
            Logger.Log(item);
        }

        Logger.Log("Item Manager Done!");
    }

    public GraphicCardItem FindGraphicCardItem(string id)
    {
        return GraphicCardItems.FirstOrDefault(item => item.Id.Equals(id));
    }
    // Read: Find a GraphicCardItem by its name
    public GraphicCardItem FindGraphicCardItemByName(string name)
    {
        return GraphicCardItems.FirstOrDefault(item => item.Name.Equals(name));
    }

}
