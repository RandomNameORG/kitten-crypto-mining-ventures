using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class GridSystem : MonoBehaviour
{
    [SerializeField]
    private InputManager InputManager;
    [SerializeField]
    private Grid grid;

    [SerializeField]
    private ObjectSO Database;
    private int SelectedObjectIndex = -1;

    Transform canvasTransform;

    // [SerializeField]
    // private GameObject gridVisualization;

    public void Start()
    {
        // StopPlacement();
        // floorData = new();
        // furnitureData = new();
        // previewRender  =  cellIndicator.GetComponentInChildren<Render>();
    }

     private void StopPlacement()
    {
        SelectedObjectIndex = -1;
        //gridVisualization.SetActive(false);
        //cellIndicator.SetActive(false);
        InputManager.OnClicked -= PlaceStructure;
        InputManager.OnExit -= StopPlacement;
    }
    public void startPlacement(int ID)
    {
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        SelectedObjectIndex = Database.Objects.FindIndex(data => data.ID == ID);
        if (SelectedObjectIndex < 0)
        {

            return;
        }
        InputManager.OnClicked += PlaceStructure;
        InputManager.OnExit += StopPlacement;
    }
    
    private void PlaceStructure()
    {
        if (InputManager.IsPointerOverUI())
        {
            Logger.Log("return");
            return;
        }

        Vector2 MousePosition = InputManager.StartPosition();
        Vector3Int GridPosition = grid.WorldToCell(new Vector3(MousePosition.x, MousePosition.y, 0f));

        GameObject InstantiateObject = Instantiate(Database.Objects[SelectedObjectIndex].Prefab);
        InstantiateObject.transform.SetParent(null, true);
        InstantiateObject.transform.position = grid.GetCellCenterWorld(GridPosition);
        // Set the Canvas as the parent of the instantiated object

    }
    
}