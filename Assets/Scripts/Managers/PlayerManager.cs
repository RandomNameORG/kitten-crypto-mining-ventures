    using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using DG.Tweening;

public class PlayerManager : MonoBehaviour
{
    //single instance convention
    public static PlayerManager _instance;
    private PlayerEntry _player_entry;
    public Text MoneyText;
    public Text voltText;
    //the current building player at;

    private readonly double SECOND = 1.0f;
    private double Timer = 0;
    //this is buildings components
    public Player CurPlayer;
    public BuildingManager Building;

    private void Awake()
    {
        _instance = this;
    }
    // Use this for initialization
    //loading data at @Start stage
    //Mention: before you starting code your loading data, you have to create init a file first
    void Start()
    {
        
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
        Building b = new();
         foreach (Building building in CurPlayer.Buildings)
        {
            totalMoney += building.MoneyPerSecond;
        }
        long preMoney = CurPlayer.Money;
        CurPlayer.Money += totalMoney;
        DOTween.To(value => { MoneyText.text = Mathf.Floor(value).ToString(); }, startValue: preMoney, endValue: CurPlayer.Money, duration: 0.1f);
        //animation.DelayFunc(Money, totalMoney);
    }

    private void DisplayVoltage()
    {
        DOTween.To(value => { voltText.text = Mathf.Floor(value).ToString() + "/" + CurPlayer.currBuildingAt.MaxVolt; },
        startValue: CurPlayer.currBuildingAt.VoltPerSecond, endValue: CurPlayer.currBuildingAt.VoltPerSecond, duration: 0.1f);
    }

}
