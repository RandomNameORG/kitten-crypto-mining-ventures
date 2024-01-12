using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;


[Serializable]
public class CardReference
{
    public int Id;
    public string Name;
    public GridPosition Pos;

}
[Serializable]
public class GraphicCardIDList
{
    public List<GraphicCard> Cards;
}

[Serializable]
public class GraphicCardList
{
    public List<GraphicCard> Cards;
}

[Serializable]
public class GraphicCard
{
    public string Name;
    public string Id;
    public bool IsLocked;
    public long PerSecondEarn;
    public long Price;
    public long PerSecondLoseVolt;
    public int Quantity;
    public string ResourcePath;
    //jsoningore
    public Sprite Icon;
}
