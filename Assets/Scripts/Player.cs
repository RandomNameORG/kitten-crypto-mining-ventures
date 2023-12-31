using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;
using DG.Tweening;

/// <summary>
/// for testing temporary
/// </summary>
public enum Skill:short
{
    HAHHASH,
    ASHDHASD
}
public class Player : MonoBehaviour
{
    public static Player Instance;
    public List<Building> Buildings;
    public int TechPoint;
    public List<Skill> Skills;
    public long Money;
    public Text text;
    public TextMeshProUGUI TextMoney;
    public int TotalCardNum;
    //the current building player at;
    public Building currBuildingAt;
    private readonly double SECOND = 1.0f;
    private double Timer = 0;
    private AnimationManager animation = new AnimationManager();

    private void Awake()
    {
        Instance = this;

    }

    /// <summary>
    /// TODO: Store player data later.
    /// </summary>
    // Start is called before the first frame update
    void Start()
    {
    }

    // Update is called once per frame
    void Update()
    {
        PerSecondEarnMoney();
    }
    private void PerSecondEarnMoney()
    {
        //update money pane
        Timer += Time.deltaTime;
        if (Timer >= SECOND)
        {
            
            long totalMoney = 0;
            foreach(Building building in Buildings)
            {
                totalMoney += building.MoneyPerSecond;
            }
            long preMoney = Money;
            Money += totalMoney;
            DOTween.To(value => { text.text = Mathf.Floor(value).ToString(); }, startValue: preMoney, endValue: Money, duration: 0.1f);
            //animation.DelayFunc(Money, totalMoney);

            Timer -= SECOND;
            
            //TextMoney.text = StringUtils.ConvertMoneyNumToString(Money);
        }
    }
}
