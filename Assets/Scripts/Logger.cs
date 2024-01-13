using UnityEngine;
using System.Diagnostics; // 用于StackTrace
using System.Linq;
using System.Collections.Generic; // 用于Linq


public enum LogType
{
    INIT,
    INIT_DONE,
}
public static class Logger
{
    private static readonly Dictionary<LogType, string> Messages = new Dictionary<LogType, string>
    {
        { LogType.INIT, "start init..." },
        { LogType.INIT_DONE, "finish init..." },
    };
    public static void Log(string text)
    {
        StackTrace stackTrace = new StackTrace();
        var frame = stackTrace.GetFrames()?.FirstOrDefault(f => f.GetMethod().DeclaringType != typeof(Logger));

        if (frame != null)
        {
            string className = frame.GetMethod().DeclaringType.Name;
            UnityEngine.Debug.Log($"[{className}]: {text}");
        }
        else
        {
            UnityEngine.Debug.Log(text);
        }
    }
    public static void Log(LogType type)
    {
        Log(Messages[type]);
    }
}
