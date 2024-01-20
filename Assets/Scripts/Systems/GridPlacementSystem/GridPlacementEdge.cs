using System.Collections;
using System.Collections.Generic;
using UnityEngine;

/// <summary>
/// Check Grid placement if cursor out of boundary or not
/// </summary>
public class GridPlacementEdge : MonoBehaviour
{
    private Collider2D myCollider;
    public GameObject buildingSystem;

    private void Start()
    {
        myCollider = GetComponent<Collider2D>();
        if (myCollider == null)
        {
            Logger.LogError("Collider2D not found!");
        }
    }

    private void Update()
    {
        Vector2 mousePosition = Input.mousePosition;
        // Convert screen coordinates to world coordinates
        Vector2 worldMousePosition = Camera.main.ScreenToWorldPoint(mousePosition);

        // Check if mouse position collides with object
        bool isColliding = myCollider.OverlapPoint(worldMousePosition);
        if (!isColliding)
        {
            buildingSystem.SetActive(false);
        }
        else
        {
            buildingSystem.SetActive(true);
        }
    }
}
