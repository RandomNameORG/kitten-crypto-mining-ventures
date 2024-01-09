using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public static class Utils 
{
    public static T[] GetAllInstance<T>() where T : ScriptableObject
    {
        return Resources.LoadAll<T>("ScriptableObjects/" + typeof(T).Name + "s")
;
    }

    public static void PauseGame()
    {
        Time.timeScale = 0;
    }
    public static void ResumeGame()
    {
        Time.timeScale = 1;
    }

}
