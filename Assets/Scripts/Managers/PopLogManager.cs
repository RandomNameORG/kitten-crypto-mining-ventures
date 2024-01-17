using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using TMPro;
using UnityEngine;
using UnityEngine.PlayerLoop;
using UnityEngine.UI;


/// <summary>
/// LogType ENUM, Explicit all Type please
/// </summary>
[Serializable]
public enum PaneLogType : int
{
    NO_ENOUGH_MONEY = 0,
    TEST = 1
}

[Serializable]
public class PopLogEntry
{
    public PaneLogType LogType;
    public string Message;//value
}

[Serializable]
public class PopLogList : GameJsonData
{
    public List<PopLogEntry> Logs = new();
}

/// <summary>
/// Singleton Class LogManager
/// </summary>
public class PopLogManager : MonoBehaviour
{

    public static PopLogManager _instance;

    private Dictionary<PaneLogType, PopLogEntry> PopLogMap = new Dictionary<PaneLogType, PopLogEntry>();


    //the pane we generate log
    [SerializeField]
    private GameObject LogPane;
    private double Timer = 0.0;
    private bool LogNow = false;

    //load log data
    void LoadLogs()
    {
        var loadLists = DataLoader.LoadData<PopLogList>(DataType.PopLogData);
        loadLists.Logs.ForEach(e => PopLogMap.Add(e.LogType, e));
    }
    private void Awake()
    {

    }

    private void SetTransparency(GameObject obj, float alpha)
    {
        Image image = obj.GetComponent<Image>();
        if (image != null)
        {
            Color color = image.color;
            color.a = alpha;
            image.color = color;
        }
        else
        {
            Logger.LogError("Image component not found on " + obj.name);
        }
    }
    private void InitLogPane()
    {
        LogPane.SetActive(true);
        var shadow = LogPane.transform.Find("Shadow").gameObject;
        var wood = LogPane.transform.Find("Wood").gameObject;
        var outline = LogPane.transform.Find("Outline").gameObject;
        var text = LogPane.transform.Find("Text (TMP)").gameObject;
        ImageUtils.SetTransparency(shadow, 0f);
        ImageUtils.SetTransparency(wood, 0f);
        ImageUtils.SetTransparency(outline, 0f);
        ImageUtils.SetTextTransparency(text, 0f);
    }
    private void Start()
    {
        _instance = this;

        LoadLogs();
        InitLogPane();
        Logger.Log(LogType.INIT_DONE);
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

    /// <summary>
    /// FadeAnimation
    /// </summary>
    /// <param name="obj"></param>
    /// <param name="fadeDuration"></param>
    /// <returns></returns>
    private void ShowLogFadeSeq(PaneLogType logType, float fadeDuration)
    {
        var animation = AnimationManager._instance;
        StartCoroutine(animation.FadeSequence(LogPane, fadeDuration / 2, fadeDuration));
    }



    public void Show(PaneLogType logType, float fadeDuration = 1f)
    {
        switch (logType)
        {
            case PaneLogType.NO_ENOUGH_MONEY:
                ShowLogFadeSeq(PaneLogType.NO_ENOUGH_MONEY, fadeDuration);
                break;
        }
    }

}