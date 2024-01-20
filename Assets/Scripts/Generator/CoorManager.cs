using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.EventSystems;

public class CoorManager : MonoBehaviour
{
    [SerializeField]
    private Camera SceneCamera;

    private Vector3 LastPosition;

    [SerializeField]
    private LayerMask PlacementLayermask;

    public event Action OnClicked, OnExit;

    private void Update()
    {
        if (Input.GetMouseButtonDown(0))
            OnClicked?.Invoke();
        if (Input.GetKeyDown(KeyCode.Escape))
            OnExit?.Invoke();
    }

    public bool IsPointerOverUI()
        => EventSystem.current.IsPointerOverGameObject();

    public Vector2 StartPosition()
    {
        Vector2 mousePos = Camera.main.ScreenToWorldPoint(Input.mousePosition);
        return mousePos;
    }
    public Vector2 GetSelectedMapPosition()
    {
        Vector2 mousePos = Camera.main.ScreenToWorldPoint(Input.mousePosition);
        RaycastHit2D hit = Physics2D.Raycast(mousePos, Vector2.zero, 0f, PlacementLayermask);
        if (hit.collider != null)
        {
            LastPosition = hit.point;
        }
        return LastPosition;
    }
}