using UnityEngine;
using System.Diagnostics; // 用于StackTrace
using System.Linq;
using System.Collections.Generic;
using System; // 用于Linq


public enum LogType
{
    INIT,
    INIT_DONE,
}
/// <summary>
/// The logger system class help us prompt more infomation we need in unity monitor
/// it causing more memo and cpu to do reflections.
/// stop using it when we felt not confort
/// </summary>
public static class Logger
{
    private static readonly Dictionary<LogType, string> Messages = new Dictionary<LogType, string>
    {
        { LogType.INIT, "start init..." },
        { LogType.INIT_DONE, "finish init..." },
    };
    public static void Log(object obj)
    {
        StackTrace stackTrace = new StackTrace();
        var frame = stackTrace.GetFrames()?.FirstOrDefault(f => f.GetMethod().DeclaringType != typeof(Logger));

        if (frame != null)
        {
            string className = frame.GetMethod().DeclaringType.Name;
            UnityEngine.Debug.Log($"[{className}]: {obj}");
        }
        else
        {
            UnityEngine.Debug.Log(obj);
        }
    }
    public static void Log(LogType type)
    {
        Log(Messages[type]);
    }

    public static void LogError(object obj)
    {
        StackTrace stackTrace = new StackTrace();
        var frame = stackTrace.GetFrames()?.FirstOrDefault(f => f.GetMethod().DeclaringType != typeof(Logger));

        if (frame != null)
        {
            string className = frame.GetMethod().DeclaringType.Name;
            UnityEngine.Debug.Log($"[{className}]: {obj}");
        }
        else
        {
            UnityEngine.Debug.LogError(obj);
        }
    }

    public static void LogWarning(object obj)
    {
        StackTrace stackTrace = new StackTrace();
        var frame = stackTrace.GetFrames()?.FirstOrDefault(f => f.GetMethod().DeclaringType != typeof(Logger));

        if (frame != null)
        {
            string className = frame.GetMethod().DeclaringType.Name;
            UnityEngine.Debug.Log($"[{className}]: {obj}");
        }
        else
        {
            UnityEngine.Debug.LogWarning(obj);
        }
    }
}
