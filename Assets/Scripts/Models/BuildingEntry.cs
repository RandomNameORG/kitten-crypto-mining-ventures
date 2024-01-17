using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

[Serializable]
public class BuildingEntryList : GameJsonData
{
    public List<BuildingEntry> Buildings = new List<BuildingEntry>();
}

[Serializable]
public class BuildingReference
{
    public string Id;
    public string Name;
}
/// <summary>
/// Building Json Data class
/// </summary>
[Serializable]
public class BuildingEntry
{
    public string Id;
    public string Name;
    public int GridSize;

    public long VoltPerSecond;
    public long MoneyPerSecond; //Money earn persecond
    public long MaxVolt;

    public long MaxCardNum;

    public double ProbabilityOfBeingAttacked;

    public int HeatDissipationLevel;

    public int LocationOfTheBuilding;
    public List<Decoration> Decorations = new List<Decoration>();

    public List<Cat> Cats = new();
    public Resources BuildingMaterial = new();


    //this is the card info we need to store
    public List<GraphicCardReference> CardSlots = new();

    //TODO We cannot store entity data for these, each only needs a little bit of data
    //For example, the following three data should not be stored in json
    //using NonSerialized like this for all of them. otherwise, it will apear on json file

    public List<GeneralEvent> Events;
    public List<GraphicCard> Cards;
    public List<Alternator> alts;
}



[Serializable]
public class Decoration
{
    public int ID;
    public Resource Resource = new Resource();
    public GridPosition Coordinates = new();
}

[Serializable]
public class Cat
{
    public int ID;
    public string Name;
    //TODO: skills and so on
}

[Serializable]
public class Resources
{

#nullable enable
    public string? RightWallMaterial;
    public string? LeftWallMaterial;

    public string? RightFloorMaterial;
    public string? LeftFloorMaterial;

}