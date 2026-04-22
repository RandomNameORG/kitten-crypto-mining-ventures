using UnityEngine;

public static class TimeUtils {

    public static double SECOND = 1.0;
    public static void PauseGame()
    {
        Time.timeScale = 0;
    }
    public static void ResumeGame()
    {
        Time.timeScale = 1;
    }
}