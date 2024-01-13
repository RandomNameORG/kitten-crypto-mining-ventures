using UnityEngine;

public class GraphicCard : ScriptableObject
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
