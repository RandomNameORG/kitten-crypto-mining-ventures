using System.Collections.Generic;
using Cinemachine;
using Unity.VisualScripting;
using UnityEngine;
using UnityEngine.InputSystem;

public class CameraController : MonoBehaviour
{

    public static CameraController _instance;
    public Dictionary<GameObject, CinemachineVirtualCamera> cameras = new();
    public CinemachineVirtualCamera activeCamera = null;
    private void Awake()
    {
        _instance = this;
    }

    private void Update()
    {

    }
    public void RegisterCamera(GameObject obj)
    {
        var virtualCameraObj = new GameObject(obj + "vc");
        virtualCameraObj.AddComponent<CinemachineVirtualCamera>();
        var camera = virtualCameraObj.GetComponent<CinemachineVirtualCamera>();
        camera.Priority = 0;
        camera.Follow = obj.transform;
        cameras[obj] = camera;
    }
    public void UnregisterCamera(GameObject obj)
    {
        //Todo: mention that I haven't delete the camera gameobject
        var camera = cameras[obj];
        Destroy(camera.gameObject);
    }
    public void SetCamera(GameObject obj)
    {
        activeCamera.Priority = 0;
        cameras[obj].Priority = 10;
    }
}