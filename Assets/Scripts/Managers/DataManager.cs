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
    private Dictionary<PathType, object> Map = new();

    private void Start()
    {
        //single
        _instance = this;
        //load all data
        //TODO if there is chance to load gernetic
        Map[PathType.BuildingData] = DataLoader.LoadData<BuildingEntryList>(PathType.BuildingData);
        Map[PathType.GraphicCardData] = DataLoader.LoadData<GraphicCardList>(PathType.GraphicCardData);
        Map[PathType.PlayerData] = DataLoader.LoadData<PlayerEntry>(PathType.PlayerData);
        Map[PathType.PopLogData] = DataLoader.LoadData<PopLogList>(PathType.PopLogData);
        Logger.Log(LogType.INIT_DONE);

    }
    /// <summary>
    /// get data
    /// </summary>
    /// <typeparam name="T"></typeparam>
    /// <param name="type"></param>
    /// <returns></returns>
    public T GetData<T>(PathType type) where T : GameJsonData
    {
        return (T)Map[type];
    }

    private void OnApplicationQuit()
    {

        //save data
        DataLoader.SaveData<BuildingEntryList>(PathType.BuildingData, (BuildingEntryList)Map[PathType.BuildingData]);
        DataLoader.SaveData<GraphicCardList>(PathType.GraphicCardData, (GraphicCardList)Map[PathType.GraphicCardData]);
        DataLoader.SaveData<PlayerEntry>(PathType.PlayerData, (PlayerEntry)Map[PathType.PlayerData]);
        DataLoader.SaveData<PopLogList>(PathType.PopLogData, (PopLogList)Map[PathType.PopLogData]);
        Logger.Log(LogType.QUIT_DONE);

    }
}