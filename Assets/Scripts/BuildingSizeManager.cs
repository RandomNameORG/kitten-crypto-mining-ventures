using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class SizeManager : MonoBehaviour
{
    
    private BuildingEntry BuildingIfo;
    public BuildingManager _instance = BuildingManager._instance;
    public GameObject Room;
    public GameObject Restriction;
    public Player Player;
    private string ID;

    public void Start(){
        ID = Player.CurrBuildingAt.Id;
    }
    public void Update(){
        RoomSize(ID);
    }
    public void RoomSize(string ID){
        BuildingIfo = _instance.FindBuildingEntryById(ID);
        // generate Room at original point with scale
        Room.transform.localPosition = new Vector3(0,0,0);
        Room.transform.localScale = new Vector3(BuildingIfo.GridSize,BuildingIfo.GridSize,BuildingIfo.GridSize);

        // generate resctriciton at original point with scale
        Restriction.transform.localPosition = new Vector3(0,0,-30);
        int size = BuildingIfo.GridSize - 200;
        Restriction.transform.localScale = new Vector3(size,size,size);
    }
}
