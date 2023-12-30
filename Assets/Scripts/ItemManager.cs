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
        Debug.Log("Item Manager init...");
        Instance = this;
        GraphicCardItems = Utils.GetAllInstance<GraphicCardItem>();
        GraphicCardItems = GraphicCardItems.OrderBy(item => item.Id).ToArray();
        foreach(GraphicCardItem item in GraphicCardItems)
        {
            Debug.Log(item);
        }

        Debug.Log("Item Manager Done!s");
    }
}
