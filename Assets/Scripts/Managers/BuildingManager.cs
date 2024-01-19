using UnityEngine;
using System.Collections;
using System.Linq;
using System.Collections.Generic;
using System.IO;
using System;


/// <summary>
/// This is Building Manager Singleton class
/// manage all the room we have, load data, and save the data
/// </summary>
public class BuildingManager : MonoBehaviour
{
    //single instance convention
    public static BuildingManager _instance;
    private BuildingEntryList _building_entries;

    //this is buildings components
    public List<Building> buildings = new();
    public List<GameObject> Buildings = new();

    private void Awake()
    {
        _instance = this;
    }
    // Use this for initialization
    //loading data at @Start stage
    //Mention: before you starting code your loading data, you have to create init a file first
    void Start()
    {
        _building_entries = DataManager._instance.GetData<BuildingEntryList>(DataType.BuildingData);
        var data = DataMapper.BuildingJsonToData(_building_entries);
        Buildings = data.Buildings; //the actual GameObject holding building comp
        buildings = data.buildings;
        Logger.Log(LogType.INIT_DONE);
    }

    //TODO think about it, how we relate our json data to our actual gameobject?
    // Read: Find a Building by its ID

    public Building FindBuildingById(string id)
    {

        Building res = buildings.FirstOrDefault(item => item.Id == id);


        return res;
    }
    public BuildingEntry FindBuildingEntryById(string id)
    {
        return _building_entries.Buildings.FirstOrDefault(item => item.Id == id);
    }
    public GameObject FindBuildingObjectById(string id)
    {
        return Buildings.FirstOrDefault(item =>
        {
            var building = item.GetComponent<Building>();
            return building.Id.Equals(id);
        });
    }
    // Read: Find a Building by its name
    public GameObject FindBuildingObjectByName(string name)
    {
        return Buildings.FirstOrDefault(item =>
        {
            var building = item.GetComponent<Building>();
            return building.Id.Equals(name);
        });
    }
    public List<GameObject> GetBuildingObjects()
    {
        return this.Buildings;
    }
}


