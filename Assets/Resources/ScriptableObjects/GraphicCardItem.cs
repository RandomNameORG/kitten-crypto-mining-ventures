using System.Collections;
using System.Collections.Generic;
using UnityEngine;

/// <summary>
/// Class <c>GraphicCardItem</c> basic item in our game
/// </summary>
[CreateAssetMenu(fileName = "NewGraphicCard", menuName ="ScriptableObjects/GraphicCardItem")]
public class GraphicCardItem : ScriptableObject
{
    public string Name;
    public string Id;
    public Sprite Icon;
    public bool IsLocked;
    public double PerSecondEarn;
    public long Price;
    public double PerSecondLoseVolt;
    public int Quantity;
}
