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
        //TODO if there is chance to load gernetic
        Map[DataType.BuildingData] = DataLoader.LoadData<BuildingEntryList>(DataType.BuildingData);
        Map[DataType.GraphicCardData] = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        Map[DataType.PlayerData] = DataLoader.LoadData<PlayerEntry>(DataType.PlayerData);
        Map[DataType.PopLogData] = DataLoader.LoadData<PopLogList>(DataType.PopLogData);
        Logger.Log(LogType.INIT_DONE);

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

        //save data
        DataLoader.SaveData<BuildingEntryList>(DataType.BuildingData, (BuildingEntryList)Map[DataType.BuildingData]);
        DataLoader.SaveData<GraphicCardList>(DataType.GraphicCardData, (GraphicCardList)Map[DataType.GraphicCardData]);
        DataLoader.SaveData<PlayerEntry>(DataType.PlayerData, (PlayerEntry)Map[DataType.PlayerData]);
        DataLoader.SaveData<PopLogList>(DataType.PopLogData, (PopLogList)Map[DataType.PopLogData]);
        Logger.Log(LogType.QUIT_DONE);

    }
}