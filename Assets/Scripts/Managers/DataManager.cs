using System;
using System.Collections.Generic;
using UnityEngine;


/// <summary>
/// Game json data
/// </summary>
public interface GameJsonData
{
}
/// <summary>
/// This class is use for manage the data load and save period in one place.
/// This class load all data when game start, and store in the dict.
/// </summary>
public class DataManager : MonoBehaviour
{

    public static DataManager _instance;
    private Dictionary<DataType, object> Map = new();

    private void Start()
    {
        //single
        _instance = this;
        //load all data
        //TODO if there is chance to load generic
        DataMapper.InitAllData();

    }
    /// <summary>
    /// get data
    /// </summary>
    /// <typeparam name="T"></typeparam>
    /// <param name="type"></param>
    /// <returns></returns>
    public T GetData<T>(DataType type) where T : GameJsonData
    {
        return (T)Map[type];
    }

    private void OnApplicationQuit()
    {

        DataMapper.OnApplicationQuit();

    }
}