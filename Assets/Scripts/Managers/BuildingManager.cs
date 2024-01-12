using UnityEngine;
using System.Collections;
using System.Linq;
using System.Collections.Generic;

/// <summary>
/// This is Building Manager Singleton class
/// manage all the room we have, load data, and save the data
/// </summary>
public class BuildingManager : MonoBehaviour
{
    //single instance convention
    public static BuildingManager _instance;
    public List<BuildingEntry> Buildings;
    // Use this for initialization
    //loading data at @Start stage
    //Mention: before you starting code your loading data, you have to create init a file first
    void Start()
    {
        _instance = this;
        //I do init data for you, so you dont have to build test data by yourself
        //data under /StreamingAssets/buildings.json
        //using that data for testing your load data
        //remember to encapsulation method make all code meaningful

    }

    //TODO think about it, how we relate our json data to our actual gameobject?
    // Read: Find a Building by its ID
    public Building FindBuildingById(string id)
    {
        return null;
        // return Buildings.FirstOrDefault(building => building.Id.Equals(id));
    }
    // Read: Find a Building by its name
    public Building FindBuildingByName(string name)
    {
        return null;
        // return Buildings.FirstOrDefault(building => building.Name.Equals(name));
    }
}


