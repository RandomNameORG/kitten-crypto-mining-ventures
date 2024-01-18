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

    public static Player _instance; //delete this, using player manager access player
    public string Name;
    public int TechPoint;
    public long Money = 0;
    public int TotalCardNum = 0;
    public Building CurrBuildingAt;

    //all building we own right now
    public List<Building> Buildings = new();
    //public List<Skill> Skills = new List<Skill>(); //havent craete ref for this yet, WIP later with skill
}
