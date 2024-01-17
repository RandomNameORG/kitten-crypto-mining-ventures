using System;
using System.Collections.Generic;
using UnityEngine;
using System.IO;


/// <summary>
/// The position where data store at
/// </summary>
[Serializable]
public enum PathType : int
{
    DataLoaderData = 1,
    PopLogData = 2,
    BuildingData = 3,
    GraphicCardData = 4,
    PlayerData = 5,
}

/// <summary>
/// This class is data loader put on manager layer
/// it will do some init when game start
/// it help project manage whole store part
/// </summary>

public class DataLoader
{
    private static string SAFilepath = Application.streamingAssetsPath + "/";
    //data for position, money that data delete game wont be exist
    private static string PFilePath = Application.persistentDataPath + "/";

    private static Dictionary<PathType, string> DataTypes = new() {
        {PathType.PlayerData, "player.json"},
        {PathType.BuildingData, "buildings.json"},
        {PathType.GraphicCardData, "graphiccards.json"},
        {PathType.PopLogData, "poplogs.json"},
    };

    private static string GetPath(string path, bool isSA = true)
    {
        if (isSA)
        {
            return SAFilepath + path;
        }
        else
        {
            return PFilePath + path;
        }
    }

    //below is public static method

    public static T LoadData<T>(PathType dataType)
    {
        //TODO temp deal with large json file
        var path = DataTypes[dataType];
        var filepath = GetPath(path);
        using StreamReader sr = new(filepath);
        var json = sr.ReadToEnd();
        var res = JsonUtility.FromJson<T>(json);
        return res;
    }

    public static void SaveData<T>(PathType dataType, T data)
    {
        var entry = DataTypes[dataType];
        var filepath = GetPath(entry);
        using StreamWriter writer = new(filepath, false);
        var json = JsonUtility.ToJson(data, true);
        writer.Write(json);
        writer.Close();
    }

}