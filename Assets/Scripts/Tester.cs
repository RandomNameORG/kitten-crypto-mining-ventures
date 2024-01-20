using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.EventSystems;


public class SceneSwitcher : MonoBehaviour
{
    
    public Camera newMainCamera;
    void Start()
    {
        // 创建新场景
        Scene newScene = SceneManager.CreateScene("NewScene");
         Scene currentScene = SceneManager.GetActiveScene();
         
         
         // 设置新场景为活动场景
        SceneManager.SetActiveScene(newScene);
       
        // 复制当前场景的主相机到新场景
        
        CopyMainCameraToNewScene();
         
        SceneManager.UnloadScene(currentScene);
    }

    void CopyMainCameraToNewScene()
    {
        // 获取当前场景的主相机
        Camera currentMainCamera = Camera.main;

        if (currentMainCamera != null)
        {
            // 创建新场景的主相机
            GameObject newMainCameraObject = new GameObject("Main Camera");
            Camera newMainCamera = newMainCameraObject.AddComponent<Camera>();

            // 复制当前相机的属性到新相机
            newMainCamera.CopyFrom(currentMainCamera);

            // 设置新相机的位置等属性（根据需求进行调整）
            newMainCameraObject.transform.position = new Vector3(0f, 0f, -10f);

            // 设置新相机为标签为 "MainCamera"
            newMainCameraObject.tag = "MainCamera";
        }
    }
}
