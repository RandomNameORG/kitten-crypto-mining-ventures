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
    public List<Skill> Skills = new List<Skill>(); //havent craete ref for this yet, WIP later with skill
    public Text MoneyText;
    public Text voltText;
    //the current building player at;

    private readonly double SECOND = 1.0f;
    private double Timer = 0;

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
        Debug.Log("player manager start init...");
        //load data
        var playerData = DataLoader.LoadData<PlayerEntry>(DataType.PlayerData);

        Debug.Log(playerData.ToString());
        Name = playerData.Name;
        TechPoint = playerData.TechPoint;
        Money = playerData.Money;
        TotalCardNum = playerData.TotalCardNum;
        var tempBuild = BuildingManager._instance.FindBuildingById(playerData.CurrBuildingAt.Id);
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
        DataLoader.SaveData<PlayerEntry>(DataType.PlayerData, data);
    }

    // Update is called once per frame
    void Update()
    {//update money pane
        Timer += Time.deltaTime;
        if (Timer >= SECOND)
        {
            PerSecondEarnMoney();
            DisplayVoltage();

            Timer -= SECOND;

            //TextMoney.text = StringUtils.ConvertMoneyNumToString(Money);
        }

    }

    private void PerSecondEarnMoney()
    {
        long totalMoney = 0;
        foreach (Building building in Buildings)
        {
            Debug.LogWarning("building:" + building);
            totalMoney += building.MoneyPerSecond;
        }
        long preMoney = Money;
        Money += totalMoney;
        DOTween.To(value => { MoneyText.text = Mathf.Floor(value).ToString(); }, startValue: preMoney, endValue: Money, duration: 0.1f);
        //animation.DelayFunc(Money, totalMoney);
    }

    private void DisplayVoltage()
    {
        DOTween.To(value => { voltText.text = Mathf.Floor(value).ToString() + "/" + currBuildingAt.MaxVolt; },
        startValue: currBuildingAt.VoltPerSecond, endValue: currBuildingAt.VoltPerSecond, duration: 0.1f);
    }



}
