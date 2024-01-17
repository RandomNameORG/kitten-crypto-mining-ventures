using System;
using System.Collections.Generic;
using UnityEngine;
using System.IO;


/// <summary>
/// The position where data store at
/// </summary>
[Serializable]
public enum DataType : int
{
    DataLoaderData = 1,
    PopLogData = 2,
    BuildingData = 3,
    GraphicCardData = 4,
    PlayerData = 5,
}
[Serializable]
public class DataTypeEntry
{
    public DataType Type;
    public string Filename;
}
[Serializable]
public class DataTypeList : GameJsonData
{
    public List<DataTypeEntry> DataTypes = new();
}
/// <summary>
/// This class is data loader put on manager layer
/// it will do some init when game start
/// it help project manage whole store part
/// </summary>

public class DataLoader : MonoBehaviour
{
    private static string SAFilepath = Application.streamingAssetsPath + "/";

    //data for position, money that data delete game wont be exist
    private static string PFilePath;
    private void Awake()
    {
        MaintainDataInit();
    }
    //check if init data in datatypes.json exist or not, if not init it
    private bool FileLoaded = true;
    void MaintainDataInit()
    {

        PFilePath = Application.persistentDataPath + "/";
        var filepath = SAFilepath + "datatypes.json";
        if (File.Exists(filepath) && FileLoaded)
        {
            using (StreamReader input = new(filepath))
            {
                var json = input.ReadToEnd();
                //check if datatypes.json str still exist
                if (!json.Contains("datatypes.json"))
                {
                    FileLoaded = false;
                    return;
                }

                var lists = JsonUtility.FromJson<DataTypeList>(json);

                lists.DataTypes.ForEach(e =>
                {
                    DataTypes.Add(e.Type, e);
                });
            }
        }
        else
        {
            DataTypeList list = new();
            DataTypeEntry entry = new();
            entry.Type = DataType.DataLoaderData;
            entry.Filename = "datatypes.json";
            list.DataTypes.Add(entry);
            var json = JsonUtility.ToJson(list, true);
            using (StreamWriter writer = new(SAFilepath + entry.Filename))
            {
                writer.Write(json);
                writer.Close();
            }

            FileLoaded = true;
        }
    }

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


    //useful dic
    private static readonly Dictionary<DataType, DataTypeEntry> DataTypes = new();

    //below is public static method

    public static T LoadData<T>(DataType dataType)
    {
        //TODO temp deal with large json file
        var entry = DataTypes[dataType];
        var filepath = GetPath(entry.Filename);
        using StreamReader sr = new(filepath);
        var json = sr.ReadToEnd();
        var res = JsonUtility.FromJson<T>(json);
        return res;
    }
    public static void SaveData<T>(DataType dataType, T data)
    {

        var entry = DataTypes[dataType];
        var filepath = GetPath(entry.Filename);
        using StreamWriter writer = new(filepath, false);
        var json = JsonUtility.ToJson(data, true);
        writer.Write(json);
        writer.Close();
    }

    



    //TODO setup load mod data here


}