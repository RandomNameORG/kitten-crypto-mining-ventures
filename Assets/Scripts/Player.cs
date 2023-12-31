using System.Collections;
using System.Collections.Generic;
using UnityEngine;

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
    public List<Building> Buildings;
    public int TechPoint;
    public List<Skill> Skills;
    public long Money;
    public int TotalCardNum;
    /// <summary>
    /// TODO: Store player data later.
    /// </summary>
    private void Awake()
    {
        Buildings = new List<Building>();
        TechPoint = 0;
        Skills = new List<Skill>();
        Money = 0;
        TotalCardNum = 0;
    }
    // Start is called before the first frame update
    void Start()
    {
        
    }

    // Update is called once per frame
    void Update()
    {
        
    }
}
