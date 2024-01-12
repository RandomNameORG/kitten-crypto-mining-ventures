using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class BuildingEntry : MonoBehaviour
{
    public int GridSize = 7; // 7 grids
    public List<Decoration> Decorations;
    public Resorces RoomMaterial;

    public int VoltPerSecond;
    public int MaxVolt;

    public int MaxCardNum;


    
}

public class Decoration
{
    public int ID;
    public GameObject Prefab;
    public float[] Coordinates;
}

public class Resorces
{
    public string RightWallMaterial;
    public string LeftWallMaterial;

    public string RightFloorMaterial;
    public string LeftFloorMaterial;
}

