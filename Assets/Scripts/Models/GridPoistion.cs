using System;
using UnityEngine;



/// <summary>
/// this model stand for speicific position in our grid
/// </summary>
[Serializable]
public class GridPosition
{
    public int X;
    public int Y;

    public Vector3 ToVector3()
    {
        return new Vector3(X, Y, -5);
    }
}