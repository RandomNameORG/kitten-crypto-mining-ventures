using System;
using System.Collections.Generic;
using UnityEngine;


[Serializable]
public class GraphicCardList
{
    public List<GraphicCardEntry> GraphicCards = new();
}
[Serializable]
public class GraphicCardEntry
{
    public string Name;
    public string Id;
    public bool IsLocked;
    public long PerSecondEarn;
    public long Price;
    public long PerSecondLoseVolt;
    public int Quantity;

    public Resource ImageSource;

}
