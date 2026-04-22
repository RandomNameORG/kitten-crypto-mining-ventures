using UnityEngine;


/// <summary>
/// GraphicCard Comp, need to attach to specific card gameobject.
/// </summary>
public class GraphicCard : MonoBehaviour
{
    public string Name;
    public string Id;
    public bool IsLocked;
    public long PerSecondEarn;
    public long Price;
    public long PerSecondLoseVolt;
    public int Quantity;
    public Sprite Icon;
}
