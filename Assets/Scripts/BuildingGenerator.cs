using System.Collections;
using System.Collections.Generic;
using UnityEngine.SceneManagement;
using System.Linq;
using UnityEngine;

public class BuildingGenerator : MonoBehaviour
{
    private BuildingEntryList BuildingData;
    public BuildingManager Building;
    public GameObject Room;    
    public GameObject Restriction;
    public Camera newMainCamera;

    void Start(){
        BuildingData = Building.GetBuildingEntryList();
        Scene currentScene = SceneManager.GetActiveScene();
        FirstGenerator();
        
        // unload controller scene
        SceneManager.UnloadScene(currentScene);
    }
    public void FirstGenerator(){
        
        foreach(BuildingEntry b in BuildingData.Buildings)
        {
            //place room and decoration
            Scene newScene = SceneGenerator(b);
            ChangeRoomSize(b.Id);
            PlaceDecoration(b);

        }
        
    }
    public Scene SceneGenerator(string Id){
        BuildingEntry Building = BuildingData.Buildings.FirstOrDefault(item => item.Id == Id);

        Scene newScene = SceneManager.CreateScene(Building.Name + Id + "", new CreateSceneParameters(LocalPhysicsMode.None));

        SceneManager.SetActiveScene(newScene);

        CopyMainCameraToNewScene();

        return newScene;
    }

    public Scene SceneGenerator(BuildingEntry Building){

        Scene newScene = SceneManager.CreateScene(Building.Name + Building.Id + "");

        SceneManager.SetActiveScene(newScene);
        Scene currentScene = SceneManager.GetActiveScene();
        CopyMainCameraToNewScene();

        return newScene;

    }

    void CopyMainCameraToNewScene()
    {
        // Get main camera in the scene
        Camera currentMainCamera = Camera.main;

        if (currentMainCamera != null)
        {
            // Create new camera
            GameObject newMainCameraObject = new GameObject("Main Camera");
            Camera newMainCamera = newMainCameraObject.AddComponent<Camera>();

            // Copy properties
            newMainCamera.CopyFrom(currentMainCamera);
            newMainCameraObject.transform.position = new Vector3(0f, 0f, -10f);

            newMainCameraObject.tag = "MainCamera";
        }
    }

    void PlaceDecoration(BuildingEntry B){
        List<Decoration> Ds = B.Decorations;
        foreach(Decoration D in Ds){
            
           
             GameObject resourcePrefab = UnityEngine.Resources.Load<GameObject>(D.Resource.Path);
             if( resourcePrefab  != null){
            GameObject resourceObj = Object.Instantiate(resourcePrefab, D.Coordinates.ToVector3(), Quaternion.identity);
            resourceObj.transform.localScale = new Vector3(20, 20, 20);
           }
            
        }
        
    }

    public void ChangeRoomSize(string ID)
    {
        BuildingEntry BuildingIfo = Building.FindBuildingEntryById(ID);
        // generate Room at original point with scale
        Room.transform.localPosition = new Vector3(0,0, 3);
        Room.transform.localScale = new Vector3(BuildingIfo.GridSize, BuildingIfo.GridSize, BuildingIfo.GridSize);

        // generate resctriciton at original point with scale
        Restriction.transform.localPosition = new Vector3(0, 0, -5);
        
        float size = (float)(0.002 * BuildingIfo.GridSize * BuildingIfo.GridSize - 1.7 * BuildingIfo.GridSize + 875);

        Restriction.transform.localScale = new Vector3(size, size, size);
            Debug.Log(BuildingIfo.GridSize);
        
    }
}
