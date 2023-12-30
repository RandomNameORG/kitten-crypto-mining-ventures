using UnityEngine;
using System.Collections;
using System.Collections.Generic;

public class Building : ScriptableObject
{
    public string Id;
    public long Capacity;
    public List<GeneralEvent> Events;
    public List<GraphicCardItem> GraphicCards;
    public double EventHappenProbs;
}

