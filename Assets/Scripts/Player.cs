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
    public Building CurrBuildingAt;

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
        PlayerEntry PayerData = DataLoader.LoadData<PlayerEntry>(DataType.PlayerData);
        
    }
    

    


}
