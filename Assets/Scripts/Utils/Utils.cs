using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public static class Utils 
{
    public static T[] GetAllInstance<T>() where T : ScriptableObject
    {
        return Resources.LoadAll<T>("ScriptableObjects/GraphicCardItems");
    }
}
