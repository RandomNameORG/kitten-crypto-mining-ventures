using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using UnityEngine;
using UnityEngine.PlayerLoop;


/// <summary>
/// LogType ENUM, Explicit all Type please
/// </summary>
[Serializable]
public enum LogType : int
{
    NO_ENOUGH_MONEY = 0,
    TEST = 1
}

[Serializable]
public class PopLogEntry
{
    public LogType LogType;
    public string Message;//value
}

[Serializable]
public class PopLogList
{
    public List<PopLogEntry> Logs = new();
}

/// <summary>
/// Singleton Class LogManager
/// </summary>
public class PopLogManager : MonoBehaviour
{

    public static PopLogManager _instance;

    private Dictionary<LogType, PopLogEntry> PopLogMap = new Dictionary<LogType, PopLogEntry>();


    //the pane we generate log
    [SerializeField]
    private GameObject LogPane;
    private double Timer = 0.0;
    private bool LogNow = false;

    //load log data
    void LoadLog()
    {
        var loadLists = DataLoader.LoadData<PopLogList>(DataType.PopLogData);

        loadLists.Logs.ForEach(e => PopLogMap.Add(e.LogType, e));
    }
    private void Awake()
    {

    }
    private void OnApplicationQuit()
    {
        PopLogList list = new PopLogList();
        foreach (var entry in PopLogMap.Values)
        {
            list.Logs.Add(entry);
        }
        DataLoader.SaveData<PopLogList>(DataType.PopLogData, list);
        Debug.Log("log save correct");
    }
    private void Start()
    {
        _instance = this;
        LoadLog();

    }
    private void Update()
    {
        if (LogNow)
        {
            Timer += Time.deltaTime;
            if (Timer > TimeUtils.SECOND)
            {
                LogNow = false;
                LogPane.SetActive(false);
            }
        }
    }
    public void Show(LogType logType)
    {
        switch (logType)
        {
            case LogType.NO_ENOUGH_MONEY:
                LogPane.SetActive(true);
                Timer = 0.0;
                LogNow = true;
                break;
        }
    }

}