using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class SizeManager : MonoBehaviour
{

    private BuildingEntry BuildingIfo;
    public BuildingManager _instance = BuildingManager._instance;
    public GameObject Room;
    public GameObject Restriction;
    public Player Player; //delete this
    private string ID;

    public void Start()
    {
        ID = PlayerManager._instance.CurPlayer.CurrBuildingAt.Id;
    }
    public void Update()
    {
        RoomSize(ID);
    }
    public void RoomSize(string ID)
    {
        BuildingIfo = _instance.FindBuildingEntryById(ID);
        // generate Room at original point with scale
        Room.transform.localPosition = new Vector3(0, 0, 0);
        Room.transform.localScale = new Vector3(BuildingIfo.GridSize, BuildingIfo.GridSize, BuildingIfo.GridSize);

        // generate resctriciton at original point with scale
        Restriction.transform.localPosition = new Vector3(0, 0, -5);
        
        float size = (float)(0.002 * BuildingIfo.GridSize * BuildingIfo.GridSize - 2.2 * BuildingIfo.GridSize + 875);

        Restriction.transform.localScale = new Vector3(size, size, size);
    }
}
