using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class PlacementSystem : MonoBehaviour
{
    [SerializeField]
    private GameObject MouseIndicator, cellIndicator;
    [SerializeField]
    private InputManager InputManager;
    [SerializeField]
    private Grid grid;

    [SerializeField]
    private ObjectDataBaseSO Database;
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
    public void startPlacement(int ID)
    {
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        StopPlacement();
        SelectedObjectIndex = Database.Objects.FindIndex(data => data.ID == ID);
        Logger.Log(ID);
        Logger.Log(SelectedObjectIndex);
        if (SelectedObjectIndex < 0)
        {

            return;
        }
        //gridVisualization.SetActive(true);
        cellIndicator.SetActive(true);
        InputManager.OnClicked += PlaceStructure;
        InputManager.OnExit += StopPlacement;

    }
    private void StopPlacement()
    {
        SelectedObjectIndex = -1;
        //gridVisualization.SetActive(false);
        //cellIndicator.SetActive(false);
        InputManager.OnClicked -= PlaceStructure;
        InputManager.OnExit -= StopPlacement;
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
    public void Update()
    {

        if (SelectedObjectIndex < 0)
        {
            return;
        }
        Vector2 MousePosition = InputManager.StartPosition();
        Vector3Int GridPosition = grid.WorldToCell(new Vector3(MousePosition.x, MousePosition.y, 0f));
        MouseIndicator.transform.position = MousePosition;
        cellIndicator.transform.position = grid.GetCellCenterWorld(GridPosition);
    }
}