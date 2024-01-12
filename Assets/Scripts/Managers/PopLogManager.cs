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
    void LoadLogs()
    {
        var loadLists = DataLoader.LoadData<PopLogList>(DataType.PopLogData);

        loadLists.Logs.ForEach(e => PopLogMap.Add(e.LogType, e));
    }
    private void Awake()
    {

    }

    //we dont have to save it for now
    // private void OnApplicationQuit()
    // {
    //     PopLogList list = new PopLogList();
    //     foreach (var entry in PopLogMap.Values)
    //     {
    //         list.Logs.Add(entry);
    //     }
    //     DataLoader.SaveData<PopLogList>(DataType.PopLogData, list);
    //     Debug.Log("log save correct");
    // }

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
            Debug.LogError("Image component not found on " + obj.name);
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
    /// havent implement AnimationManager, put here first 
    /// </summary>
    /// <param name="obj"></param>
    /// <param name="fadeDuration"></param>
    /// <returns></returns>
    private IEnumerator ShowLogFadeSeq(LogType logType, float fadeDuration)
    {
        fadeDuration /= 2;
        var shadow = LogPane.transform.Find("Shadow").gameObject;
        var wood = LogPane.transform.Find("Wood").gameObject;
        var outline = LogPane.transform.Find("Outline").gameObject;
        var text = LogPane.transform.Find("Text (TMP)").gameObject;
        text.GetComponent<TextMeshProUGUI>().text = PopLogMap[logType].Message;

        StartCoroutine(ImageUtils.ImageFadeIn(wood, fadeDuration));
        StartCoroutine(ImageUtils.ImageFadeIn(shadow, fadeDuration));
        StartCoroutine(ImageUtils.ImageFadeIn(outline, fadeDuration));
        StartCoroutine(ImageUtils.TextFadeIn(text, fadeDuration));
        yield return new WaitForSeconds(fadeDuration * 2);
        StartCoroutine(ImageUtils.ImageFadeOut(wood, fadeDuration));
        StartCoroutine(ImageUtils.ImageFadeOut(shadow, fadeDuration));
        StartCoroutine(ImageUtils.ImageFadeOut(outline, fadeDuration));
        StartCoroutine(ImageUtils.TextFadeOut(text, fadeDuration));
    }



    public void Show(LogType logType, float fadeDuration = 1f)
    {
        switch (logType)
        {
            case LogType.NO_ENOUGH_MONEY:
                Debug.Log("debug bei calll lalfasldal");
                StartCoroutine(ShowLogFadeSeq(logType, fadeDuration));

                break;
        }
    }

}