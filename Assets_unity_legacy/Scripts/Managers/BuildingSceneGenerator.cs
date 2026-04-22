using System.Collections;
using System.Collections.Generic;
using UnityEngine.SceneManagement;
using System.Linq;
using UnityEngine;
using Cinemachine;

public static class BuildingSceneGenerator
{

    //scene manager
    // void Start()
    // {
    //     _instance = this;
    //     BuildingData = Building.GetBuildingEntryList();
    //     Scene currentScene = SceneManager.GetActiveScene();
    //     // FirstGenerator();

    //     // unload controller scene
    //     SceneManager.UnloadSceneAsync(currentScene);
    // }
    // public Scene SceneGenerator(string Id)
    // {
    //     BuildingEntry Building = BuildingData.Buildings.FirstOrDefault(item => item.Id == Id);

    //     Scene newScene = SceneManager.CreateScene(Building.Name + Id + "", new CreateSceneParameters(LocalPhysicsMode.None));

    //     SceneManager.SetActiveScene(newScene);

    //     CopyMainCameraToNewScene();

    //     return newScene;
    // }

    public static Scene GenerateScene(BuildingEntry buildingEntry)
    {
        Scene newScene = SceneManager.CreateScene(buildingEntry.Name + buildingEntry.Id);
        SceneManager.SetActiveScene(newScene);
        var building = DataMapper.GenerateBuilding(buildingEntry);

        GameObject cameraObject = new GameObject("vm camera " + building.Id);
        cameraObject.AddComponent<CinemachineVirtualCamera>();
        var camera = cameraObject.GetComponent<CinemachineVirtualCamera>();
        camera.Follow = building.transform;
        camera.Priority = 0;
        return newScene;
    }
    // public Scene SceneGenerator(BuildingEntry Building)
    // {

    //     Scene newScene = SceneManager.CreateScene(Building.Name + Building.Id + "");
    //     SceneManager.SetActiveScene(newScene);
    //     //TODO do mapper here
    //     GameObject temp = new GameObject("temp"); //testing object
    //     GameObject cameraObject = new GameObject("vm camera " + Building.Id);
    //     cameraObject.AddComponent<CinemachineVirtualCamera>();
    //     var camera = cameraObject.GetComponent<CinemachineVirtualCamera>();
    //     camera.Follow = temp.transform;
    //     camera.Priority = 0;
    //     return newScene;
    // }

    private static void PlaceDecoration(BuildingEntry B)
    {
        List<Decoration> Ds = B.Decorations;
        foreach (Decoration D in Ds)
        {
            GameObject resourcePrefab = UnityEngine.Resources.Load<GameObject>(D.Resource.Path);
            if (resourcePrefab != null)
            {
                GameObject resourceObj = Object.Instantiate(resourcePrefab, D.Coordinates.ToVector3(), Quaternion.identity);
                resourceObj.transform.localScale = new Vector3(20, 20, 20);
            }

        }

    }

    //TODO Think about it lator
    // private static void ChangeRoomSize(BuildingEntry BuildingIfo)
    // {
    //     BuildingEntry BuildingIfo = Building.FindBuildingEntryById(ID);
    //     // generate Room at original point with scale
    //     Room.transform.localPosition = new Vector3(0, 0, 3);
    //     Room.transform.localScale = new Vector3(BuildingIfo.GridSize, BuildingIfo.GridSize, BuildingIfo.GridSize);

    //     // generate resctriciton at original point with scale
    //     Restriction.transform.localPosition = new Vector3(0, 0, -5);

    //     float size = (float)(0.002 * BuildingIfo.GridSize * BuildingIfo.GridSize - 1.7 * BuildingIfo.GridSize + 875);

    //     Restriction.transform.localScale = new Vector3(size, size, size);
    //     Debug.Log(BuildingIfo.GridSize);

    // }
}
