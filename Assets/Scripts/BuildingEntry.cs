using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

[Serializable]
public class BuildingList{
    public List<BuildingEntry> Buildings;
}

[Serializable]
public class BuildingEntry
{
    public int GridSize;

    public List<Decoration> Decorations;

    public List<Cat> Cats;

    public Resorces BuildingMaterial;

    public int VoltPerSecond;

    public int MaxVolt;

    public int MaxCardNum;

    public int ProbabilityOfBeingAttacked;

    public int HeatDissipationLevel;

    public int LocationOfTheBuilding;
}

public class Decoration
{
    public int ID;
    public GameObject Prefab;
    public float[] Coordinates;
}

public class Cat
{
    public int ID;
    public string Name;
    //TODO: skills and so on
}

public class Resorces
{
    public string RightWallMaterial;
    public string LeftWallMaterial;

    public string RightFloorMaterial;
    public string LeftFloorMaterial;
}

