using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class PlacementSystem : MonoBehaviour
{
    [SerializeField]
    private GameObject mouseIndicator, cellIndicator;
    [SerializeField]
    private inputManager inputManager;
    [SerializeField]
    private Grid grid;

    [SerializeField]
    private ObjectDataBaseSO database;
    private int selectedObjectIndex = -1; 

    Transform canvasTransform;

    // [SerializeField]
    // private GameObject gridVisualization;

    public void Start(){
        // StopPlacement();
        // floorData = new();
        // furnitureData = new();
        // previewRender  =  cellIndicator.GetComponentInChildren<Render>();
    }
    public void startPlacement(int ID){ 
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        StopPlacement();
        selectedObjectIndex = database.Objects.FindIndex(data => data.ID == ID);    
        Debug.Log(ID);
        Debug.Log(selectedObjectIndex);
        if(selectedObjectIndex < 0){
            
            return;
        }
        //gridVisualization.SetActive(true);
        cellIndicator.SetActive(true);
        inputManager.OnClicked += PlaceStructure;
        inputManager.OnExit += StopPlacement;
       
    }
    private void StopPlacement(){
        selectedObjectIndex = -1;
        //gridVisualization.SetActive(false);
        //cellIndicator.SetActive(false);
        inputManager.OnClicked -= PlaceStructure;
        inputManager.OnExit -= StopPlacement;
    }
    private void PlaceStructure(){
        if(inputManager.IsPointerOverUI()){ 
            Debug.Log("return");
            return;
        }
        
        Vector2 mousePosition = inputManager.StartPosition();
        Vector3Int gridPosition = grid.WorldToCell(new Vector3(mousePosition.x, mousePosition.y, 0f));

        GameObject gameobject = Instantiate(database.Objects[selectedObjectIndex].Prefab);
        gameobject.transform.SetParent(null, true);
        gameobject.transform.position = grid.GetCellCenterWorld(gridPosition);
        // Set the Canvas as the parent of the instantiated object

    }
    public void Update(){
        
        if(selectedObjectIndex < 0){
            return;
        }
        Vector2 mousePosition = inputManager.StartPosition();
        Vector3Int gridPosition = grid.WorldToCell(new Vector3(mousePosition.x, mousePosition.y, 0f));
        mouseIndicator.transform.position = mousePosition;
        cellIndicator.transform.position = grid.GetCellCenterWorld(gridPosition);
    }
}