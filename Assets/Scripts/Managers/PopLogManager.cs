using UnityEngine;
using UnityEngine.PlayerLoop;

/// <summary>
/// Singleton Class LogManager
/// </summary>


public enum LogType: int {
    NO_ENOUGH_MONEY,
}
public class PopLogManager : MonoBehaviour {


    public static PopLogManager _instance;
    //the pane we generate log
    [SerializeField] 
    private GameObject LogPane;
    private double Timer = 0.0;
    private bool LogNow = false;

    private void Start() {
        _instance = this;
    }
    private void Update() {
        if(LogNow) {
            Timer += Time.deltaTime;
            if(Timer > TimeUtils.SECOND) {
                LogNow = false;
                LogPane.SetActive(false);
            }
        }
    }
    public void Show(LogType logType){
        switch(logType) {
            case LogType.NO_ENOUGH_MONEY:
                LogPane.SetActive(true);
                Timer = 0.0;
                LogNow = true;
                break;
        }
    }

}