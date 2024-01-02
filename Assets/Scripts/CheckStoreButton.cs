using UnityEngine.SceneManagement;
using UnityEngine;
using UnityEngine.UI;

public class CheckStoreButton : MonoBehaviour
{
    public static bool isFinish = false;
    // Start is called before the first frame update
    void Start()
    {
        GetComponent<Button>().onClick.AddListener(finishbuy);
    }
    public void finishbuy(){
        isFinish = true;
        Debug.Log("finish");
    }

    
}
