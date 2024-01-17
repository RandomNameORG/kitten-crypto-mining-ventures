using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using DG.Tweening;
using System;

/// <summary>
/// for testing temporary
/// </summary>
public enum Skill : short
{
    HAHHASH,
    ASHDHASD
}

/// <summary>
/// Player and Player manager temp highly coupled, decoupling later
/// </summary>
public class Player : MonoBehaviour
{

    public static Player _instance;
    public string Name;
    public int TechPoint;
    public long Money = 0;
    public int TotalCardNum = 0;
    public Building currBuildingAt;

    //all building we own right now
    public List<Building> Buildings = new();
    //public List<Skill> Skills = new List<Skill>(); //havent craete ref for this yet, WIP later with skill

    


    /// <summary>
    /// TODO: Store player data later.
    /// </summary>
    // Start is called before the first frame update
    private void Awake()
    {
        _instance = this;

    }
    void Start()
    {
        Logger.Log("start init...");
        //load data
        var playerData = DataLoader.LoadData<PlayerEntry>(PathType.PlayerData);

        Logger.Log(playerData.ToString());
        Name = playerData.Name;
        TechPoint = playerData.TechPoint;
        Money = playerData.Money;
        TotalCardNum = playerData.TotalCardNum;
        var tempBuild = BuildingManager._instance.FindBuildingById(playerData.CurrBuildingAt.Id);
        Logger.LogWarning("tempbuild init here: " + playerData.CurrBuildingAt.Id);
        currBuildingAt = tempBuild;
        Buildings = BuildingManager._instance.buildings;
    }
    private void OnApplicationQuit()
    {
        List<BuildingReference> buildingRefs = new();
        Buildings.ForEach(item =>
        {
            buildingRefs.Add(new BuildingReference
            {
                Id = item.Id,
                Name = item.Name
            });
        });
        PlayerEntry data = new PlayerEntry
        {
            Name = Name,
            TechPoint = TechPoint,
            Money = Money,
            TotalCardNum = TotalCardNum,
            CurrBuildingAt = new BuildingReference
            {
                Id = currBuildingAt.Id,
                Name = currBuildingAt.Name
            },
            BuildingsRef = buildingRefs
        };
        DataLoader.SaveData<PlayerEntry>(PathType.PlayerData, data);
    }

    


}
