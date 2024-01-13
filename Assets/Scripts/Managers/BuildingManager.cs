using UnityEngine;
using System.Collections;
using System.Linq;
using System.Collections.Generic;
using System.IO;
using AlternatorProject;
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
        Logger.Log(LogType.INIT);
        //get json data
        _building_entries = DataLoader.LoadData<BuildingEntryList>(DataType.BuildingData);
        ;

        //TODO we need to make this process more easy to write, this way so stupid...
        //or pust some where else
        //convert to building object list
        _building_entries.Buildings.ForEach(e =>
        {

            var obj = new GameObject(e.Name);
            //create building comp
            obj.AddComponent<Building>();
            var building = obj.GetComponent<Building>();
            building.Id = e.Id.ToString(); // Assuming  convert the int Id to string
            building.Name = e.Name;
            building.Capacity = e.MaxCardNum;
            building.MaxVolt = e.MaxVolt; // Assuming Capacity is equivalent to MaxVolt
            building.Events = new List<GeneralEvent>(e.Events);
            building.Cards = e.CardSlots.Select(cs =>
            {
                return GraphicCardManager._instance.FindCardById(cs.Id);
            }).ToList();
            building.EventHappenProbs = e.ProbabilityOfBeingAttacked;
            building.MoneyPerSecond = e.MoneyPerSecond;
            building.Alts = new List<Alternator>(e.alts);
            building.VoltPerSecond = e.VoltPerSecond;
            buildings.Add(building);
            Buildings.Add(obj);

            //here we have to do something to building house
            //make it exist in the world
            obj.transform.SetPositionAndRotation(new Vector3(0, 0, 0), Quaternion.identity);
            obj.transform.localScale = new Vector3(1, 1, 1);
        });
        Logger.Log(LogType.INIT_DONE);
    }
    private void OnApplicationQuit()
    {
        _building_entries = new BuildingEntryList();
        //TODO optimaztion this process plz
        Buildings.ForEach(obj =>
        {
            var building = obj.GetComponent<Building>();
            BuildingEntry e = new BuildingEntry();
            e.Id = int.Parse(building.Id);
            e.Name = building.Name;
            e.MaxCardNum = building.Capacity; // Assuming MaxCardNum is equivalent to Capacity
            e.MaxVolt = building.MaxVolt; // Assuming this assignment logic remains the same
            // Assuming e has a CardSlots property that can be assigned from building.Cards
            e.CardSlots = building.Cards.Select(gc =>
            {
                return new GraphicCardReference { Id = gc.Id, Name = gc.Name }; // Assuming CardSlot has an Id property and you can create new instances like this
            }).ToList();
            e.ProbabilityOfBeingAttacked = building.EventHappenProbs;
            e.MoneyPerSecond = building.MoneyPerSecond;
            e.VoltPerSecond = building.VoltPerSecond;
            _building_entries.Buildings.Add(e);
        });
        DataLoader.SaveData<BuildingEntryList>(DataType.BuildingData, _building_entries);
    }


    //TODO think about it, how we relate our json data to our actual gameobject?
    // Read: Find a Building by its ID

    public Building FindBuildingById(string id)
    {
        return buildings.FirstOrDefault(item => item.Id == id);
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
}


